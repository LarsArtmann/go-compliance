package cli

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/larsartmann/go-compliance"
	"github.com/larsartmann/go-compliance/hub/internal/discovery"
	"github.com/larsartmann/go-compliance/hub/internal/events"
)

const (
	// refreshDebounce coalesces a burst of triggers (a multi-repo report
	// burst, a discovery storm) into one re-query.
	refreshDebounce = 500 * time.Millisecond
	// watchReconnectBackoff bounds the re-subscribe loop when the
	// discovery daemon is down or the stream drops.
	watchReconnectBackoff = 30 * time.Second
	// refreshSafetyNet re-derives on a slow tick even without triggers:
	// external processes appending to the same SQLite file get picked up
	// eventually (the write hook only sees this process's writes), and
	// staleness itself is time-based state.
	refreshSafetyNet = 5 * time.Minute
)

// refresher keeps the served rows live: every trigger (store write,
// discovery event) re-queries the source and publishes a row refresh.
// Triggers are coalesced through a buffered channel and debounced, so a
// burst costs one re-derive. The latest evidence snapshot is kept for the
// JSON API surface.
type refresher struct {
	hub        *events.Hub
	source     compliance.Source
	staleAfter time.Duration
	worstSteps int
	stepFilter string
	window     time.Duration
	debounce   time.Duration

	// trigger has capacity 1: sends never block and a send while a
	// refresh is already pending is a no-op (that IS the coalescing).
	trigger chan struct{}

	evidence atomic.Pointer[[]compliance.RepoEvidence]

	// discovered is the latest discovery project set (nil = discovery
	// never answered); guarded atomically because the watch loop and the
	// refresh loop touch it.
	discovered atomic.Pointer[[]compliance.RepoName]
}

// newRefresher builds a live refresher over source. window is the
// evidence window length (since = now - window at every refresh).
func newRefresher(
	hub *events.Hub, source compliance.Source,
	window, staleAfter time.Duration, worstSteps int, stepFilter string,
) *refresher {
	return &refresher{
		hub:        hub,
		source:     source,
		staleAfter: staleAfter,
		worstSteps: worstSteps,
		stepFilter: stepFilter,
		window:     window,
		debounce:   refreshDebounce,
		trigger:    make(chan struct{}, 1),
	}
}

// Kick schedules a refresh (coalescing: never blocks, never queues twice).
func (r *refresher) Kick() {
	select {
	case r.trigger <- struct{}{}:
	default:
	}
}

// Evidence returns the latest derived evidence snapshot (nil before the
// first successful refresh).
func (r *refresher) Evidence() []compliance.RepoEvidence {
	if evidence := r.evidence.Load(); evidence != nil {
		return *evidence
	}

	return nil
}

// Refresh derives and publishes once, synchronously; the initial publish
// at boot uses it directly.
func (r *refresher) Refresh(ctx context.Context) error {
	now := time.Now()

	evidence, err := r.source.QueryEvidence(ctx, now.Add(-r.window), r.worstSteps, r.stepFilter)
	if err != nil {
		return err
	}

	if discovered := r.discovered.Load(); discovered != nil {
		evidence = compliance.CompleteWithDiscovery(evidence, *discovered)
	}

	rows := events.RowsFromEvidence(evidence, now, r.staleAfter)

	if _, err := r.hub.PublishRefresh(rows); err != nil {
		return err
	}

	r.evidence.Store(&evidence)

	return nil
}

// Run serves the trigger loop until ctx cancels. Refresh failures are
// logged and retried on the next trigger: a transient source error must
// not take the dashboard down (the health probe's evidence check keeps
// reporting the last good publish).
//
// art-dupl:accept nested selects are the canonical trigger-loop shape; the two ctx.Done exits share no logic
func (r *refresher) Run(ctx context.Context) {
	safetyNet := time.NewTicker(refreshSafetyNet)
	defer safetyNet.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-r.trigger:
		case <-safetyNet.C:
		}

		// Debounce: absorb the burst, then derive once.
		timer := time.NewTimer(r.debounce)

		waiting := true
		for waiting {
			select {
			case <-ctx.Done():
				timer.Stop()

				return
			case <-r.trigger: // stay in the window; more may come
			case <-timer.C:
				waiting = false
			}
		}

		if err := r.Refresh(ctx); err != nil {
			slog.Warn("compliance-hub: refresh failed", "error", err)
		}
	}
}

// runDiscoveryWatch subscribes to the discovery daemon's /v1/watch stream
// and kicks the refresher on every fleet change. The daemon being down is
// normal (best-effort completeness): errors back off and retry, and each
// successful (re)subscription also refreshes once so a late-starting
// daemon still completes the never-covered set.
func runDiscoveryWatch(ctx context.Context, client *discovery.Client, r *refresher) {
	for {
		if ctx.Err() != nil {
			return
		}

		eventCh, err := client.Watch(ctx)
		if err != nil {
			slog.Warn("compliance-hub: discovery watch unavailable, retrying",
				"socket", client.SocketPath(), "error", err, "backoff", watchReconnectBackoff)

			if !sleepCtx(ctx, watchReconnectBackoff) {
				return
			}

			continue
		}

		slog.Info("compliance-hub: discovery watch connected", "socket", client.SocketPath())

		if err := r.refreshDiscovered(ctx, client); err != nil {
			slog.Warn("compliance-hub: discovery snapshot failed", "error", err)
		}

		for event := range eventCh {
			if event.Project != nil {
				r.Kick()
			}
		}

		// Stream ended (daemon restart): re-subscribe after a beat.
		if !sleepCtx(ctx, time.Second) {
			return
		}
	}
}

// refreshDiscovered pulls one discovery snapshot and stores it as the
// never-covered completion set.
func (r *refresher) refreshDiscovered(ctx context.Context, client *discovery.Client) error {
	projects, err := client.Discover(ctx, nil)
	if err != nil {
		return err
	}

	names := discovery.RepoNames(projects)
	r.discovered.Store(&names)
	r.Kick()

	return nil
}

// sleepCtx sleeps for d or until ctx cancels; false means ctx cancelled.
func sleepCtx(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

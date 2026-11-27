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
	// refreshTimeout bounds one QueryEvidence: a hung source wedges a
	// single refresh (logged, retried on the next trigger), never the
	// loop and never the safety-net ticker.
	refreshTimeout = 30 * time.Second
	// refreshFreshnessWarn is the age of the last successful refresh at
	// which the refresh-freshness health check degrades to warn. Verdict
	// g1 (2026-11-18): warn, /readyz stays 200 — a stale view is degraded
	// truth, not service loss. 10 min = two missed safety-net ticks, i.e.
	// the derive loop is wedged or the source is dead.
	refreshFreshnessWarn = 10 * time.Minute
	// discoveryStaleWarn is the age of the last successful discovery
	// contact (watch connect or snapshot) at which the non-critical
	// discovery check degrades to warn — same warn-not-fail verdict as
	// refreshFreshnessWarn: discovery is best-effort completeness, the
	// hub stays useful without it.
	discoveryStaleWarn = 10 * time.Minute
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

	// discoverPaths are the discovery search roots (empty = disabled).
	discoverPaths []string

	// trigger has capacity 1: sends never block and a send while a
	// refresh is already pending is a no-op (that IS the coalescing).
	trigger chan struct{}

	evidence atomic.Pointer[[]compliance.RepoEvidence]

	// discovered is the latest discovery project set (nil = discovery
	// never answered); guarded atomically because the watch loop and the
	// refresh loop touch it.
	discovered atomic.Pointer[[]compliance.RepoName]

	// lastRefreshSuccess is the unix-nano stamp of the last fully
	// successful Refresh (query AND publish); 0 = never succeeded. The
	// refresh-freshness health check reads it via LastRefreshSuccess.
	lastRefreshSuccess atomic.Int64

	// lastDiscoverySuccess is the unix-nano stamp of the last successful
	// discovery contact (watch connect, snapshot, or stream event); 0 =
	// never contacted. The discovery health check reads it via
	// LastDiscoverySuccess and consults discoveryState first.
	lastDiscoverySuccess atomic.Int64

	// discoveryState tells the health check whether the stamp matters.
	discoveryState atomic.Uint32

	// queryTimeout defaults to refreshTimeout; tests shrink it.
	queryTimeout time.Duration
}

// newRefresher builds a live refresher over source. window is the
// evidence window length (since = now - window at every refresh).
func newRefresher(
	hub *events.Hub, source compliance.Source,
	window, staleAfter time.Duration, worstSteps int, stepFilter string,
) *refresher {
	return &refresher{
		hub:          hub,
		source:       source,
		staleAfter:   staleAfter,
		worstSteps:   worstSteps,
		stepFilter:   stepFilter,
		window:       window,
		debounce:     refreshDebounce,
		trigger:      make(chan struct{}, 1),
		queryTimeout: refreshTimeout,
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

// LastRefreshSuccess returns when the last fully successful Refresh ran;
// zero before the first success.
func (r *refresher) LastRefreshSuccess() time.Time {
	if nanos := r.lastRefreshSuccess.Load(); nanos != 0 {
		return time.Unix(0, nanos)
	}

	return time.Time{}
}

// discoveryState values: a LIVE watch stream is continuous contact even
// when the fleet is quiet, OFF means the operator disabled discovery, and
// IDLE falls back to the age of the last contact.
const (
	discoveryIdle uint32 = iota
	discoveryLive
	discoveryOff
)

// markDiscoverySuccess stamps a successful discovery contact.
func (r *refresher) markDiscoverySuccess() {
	r.lastDiscoverySuccess.Store(time.Now().UnixNano())
}

// DisableDiscovery records that no roots were configured: the watch loop
// is not started and the health check stays green instead of reporting a
// daemon that was never meant to be contacted.
func (r *refresher) DisableDiscovery() {
	r.discoveryState.Store(discoveryOff)
}

// DiscoveryLive reports whether discovery needs no staleness verdict: a
// connected stream is contact by itself, and a disabled feature cannot
// be stale.
func (r *refresher) DiscoveryLive() bool {
	return r.discoveryState.Load() != discoveryIdle
}

// LastDiscoverySuccess returns when the daemon last answered (watch
// connect, snapshot, or stream event); zero before the first contact.
func (r *refresher) LastDiscoverySuccess() time.Time {
	if nanos := r.lastDiscoverySuccess.Load(); nanos != 0 {
		return time.Unix(0, nanos)
	}

	return time.Time{}
}

// Refresh derives and publishes once, synchronously; the initial publish
// at boot uses it directly.
func (r *refresher) Refresh(ctx context.Context) error {
	now := time.Now()

	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	evidence, err := r.source.QueryEvidence(queryCtx, now.Add(-r.window), r.worstSteps, r.stepFilter)
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
	r.lastRefreshSuccess.Store(time.Now().UnixNano())

	return nil
}

// Run serves the trigger loop until ctx cancels. Refresh failures are
// logged and retried on the next trigger: a transient source error must
// not take the dashboard down (the health probe's evidence check keeps
// reporting the last good publish).
//
// The debounce is a fixed window from the FIRST trigger of a burst: later
// triggers inside the window are absorbed and never extend it, so a steady
// stream of kicks yields one refresh per window and cannot starve the loop.
func (r *refresher) Run(ctx context.Context) {
	safetyNet := time.NewTicker(refreshSafetyNet)
	defer safetyNet.Stop()

	for {
		select {
		case <-ctx.Done():
			// art-dupl:accept nested selects are the canonical trigger-loop shape; the two ctx.Done exits share no logic
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

// startDiscovery wires the refresher to the discovery daemon on
// socketPath. Without search roots the daemon would reject every snapshot,
// so discovery is off instead: never-covered stays an empty set by design,
// the discovery health check passes, and one log line says so once.
func startDiscovery(
	ctx context.Context, refresh *refresher, discoverPaths []string, socketPath string,
) {
	refresh.discoverPaths = discoverPaths

	if len(discoverPaths) == 0 {
		refresh.DisableDiscovery()
		slog.Info("compliance-hub: discovery disabled (--discover-paths is empty); never-covered is not computed")

		return
	}

	go runDiscoveryWatch(ctx, discovery.New(socketPath), refresh)
}

// runDiscoveryWatch subscribes to the discovery daemon's /v1/watch stream
// and kicks the refresher on every fleet change. The daemon being down is
// normal (best-effort completeness): errors back off and retry, and each
// successful (re)subscription also refreshes once so a late-starting
// daemon still completes the never-covered set.
func runDiscoveryWatch(ctx context.Context, client *discovery.Client, refresh *refresher) {
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

		refresh.markDiscoverySuccess()
		refresh.discoveryState.Store(discoveryLive)

		slog.Info("compliance-hub: discovery watch connected", "socket", client.SocketPath())

		if err := refresh.refreshDiscovered(ctx, client, refresh.discoverPaths); err != nil {
			slog.Warn("compliance-hub: discovery snapshot failed", "error", err)
		} else {
			refresh.markDiscoverySuccess()
		}

		for event := range eventCh {
			// Every frame proves the daemon is answering; only project
			// changes re-derive the rows.
			refresh.markDiscoverySuccess()

			if event.Project != nil {
				refresh.Kick()
			}
		}

		refresh.discoveryState.Store(discoveryIdle)

		// Stream ended (daemon restart): re-subscribe after a beat.
		if !sleepCtx(ctx, time.Second) {
			return
		}
	}
}

// refreshDiscovered pulls one discovery snapshot for the search paths and
// stores the fleet-eligible projects as the never-covered completion set.
func (r *refresher) refreshDiscovered(ctx context.Context, client *discovery.Client, searchPaths []string) error {
	projects, err := client.Discover(ctx, searchPaths)
	if err != nil {
		return err
	}

	names := discovery.RepoNames(discovery.FleetEligible(projects))
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

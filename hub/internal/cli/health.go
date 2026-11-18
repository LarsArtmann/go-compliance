package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/larsartmann/go-health"
	"github.com/larsartmann/go-health/checks"

	"github.com/larsartmann/go-compliance/hub/internal/events"
)

// evidenceCheckName is the one critical service: the hub has folded
// evidence and published state. Until it passes, /readyz and /startupz
// report not-ready — an evidence failure must not restart the process
// (liveness stays pass regardless).
const evidenceCheckName = "evidence"

// refreshFreshnessCheckName is non-critical on purpose (verdict g1,
// 2026-11-18): a stalled derive loop degrades the served view to stale
// truth but the process still serves — warn (200), never fail (503).
const refreshFreshnessCheckName = "refresh-freshness"

// discoveryCheckName is non-critical for the same verdict: discovery is
// best-effort completeness of the never-covered set, not serving.
const discoveryCheckName = "discovery"

// sseBackpressureCheckName is non-critical and delta-based: it warns only
// on drops NEW since the last evaluation, so a past burst self-clears
// instead of standing warn forever (the drop counter is cumulative).
const sseBackpressureCheckName = "sse-backpressure"

var (
	// ErrEvidenceNotPublished is the evidence check's failure: the hub booted
	// but has not folded evidence yet (source error or first query pending).
	ErrEvidenceNotPublished = errors.New("evidence not published yet")

	// ErrRefreshNeverSucceeded is the freshness check's pre-first-success
	// state; production serve exits before the probe starts when the boot
	// refresh fails, so it mostly matters for embedders.
	ErrRefreshNeverSucceeded = errors.New("no refresh has succeeded yet")

	// ErrRefreshStale wraps every past-threshold freshness failure with the
	// concrete age.
	ErrRefreshStale = errors.New("last successful refresh too old")

	// ErrDiscoveryNeverContacted is the discovery check's never-answered
	// state (daemon down since boot, or discovery not wired).
	ErrDiscoveryNeverContacted = errors.New("discovery daemon never contacted")

	// ErrDiscoverySilent wraps every past-threshold discovery silence with
	// the concrete age.
	ErrDiscoverySilent = errors.New("discovery daemon silent too long")

	// ErrSSEDropped wraps every backpressure warning with the new, total,
	// and subscriber counts.
	ErrSSEDropped = errors.New("sse patches dropped since the last health evaluation")
)

// buildHealthProbe wires the hub's health probe: liveness is always pass;
// readiness/startup gate on published evidence; refresh-freshness and
// discovery warn when their loops stop making progress; disk headroom
// keeps the SQLite cache honest.
func buildHealthProbe(
	ctx context.Context, version string, hub *events.Hub, refresher *refresher, cacheDir string,
) (*health.Probe, error) {
	checksByName := map[string]health.CheckFunc{
		evidenceCheckName: func(_ context.Context) error {
			if hub.Rows() == nil {
				return ErrEvidenceNotPublished
			}

			return nil
		},
		refreshFreshnessCheckName: func(_ context.Context) error {
			last := refresher.LastRefreshSuccess()
			if last.IsZero() {
				return ErrRefreshNeverSucceeded
			}

			if age := time.Since(last); age > refreshFreshnessWarn {
				return fmt.Errorf("%w: %s ago (warn threshold %s)",
					ErrRefreshStale, age.Round(time.Second), refreshFreshnessWarn)
			}

			return nil
		},
		discoveryCheckName: func(_ context.Context) error {
			last := refresher.LastDiscoverySuccess()
			if last.IsZero() {
				return ErrDiscoveryNeverContacted
			}

			if age := time.Since(last); age > discoveryStaleWarn {
				return fmt.Errorf("%w: %s ago (warn threshold %s)",
					ErrDiscoverySilent, age.Round(time.Second), discoveryStaleWarn)
			}

			return nil
		},
	}

	var (
		lastDropped int64
		dropMu      sync.Mutex
	)

	checksByName[sseBackpressureCheckName] = func(_ context.Context) error {
		dropMu.Lock()
		defer dropMu.Unlock()

		dropped := hub.Dropped()
		delta := dropped - lastDropped
		lastDropped = dropped

		if delta > 0 {
			return fmt.Errorf("%w: %d new (total %d, subscribers %d)",
				ErrSSEDropped, delta, dropped, hub.SubscriberCount())
		}

		return nil
	}

	if cacheDir != "" {
		const minFreeBytes = 128 << 20 // 128 MiB is plenty to append runs

		checksByName["cache-disk"] = checks.Disk(cacheDir, minFreeBytes)
	}

	probe := health.NewChecks(checksByName,
		health.WithVersion(version),
		health.WithCriticalServices(evidenceCheckName),
		health.WithEvaluationHook(checkTransitions()),
	)

	if err := probe.Start(ctx); err != nil {
		return nil, err
	}

	return probe, nil
}

// checkTransitions returns a WithEvaluationHook callback that slogs every
// check whose status flipped since the previous evaluation (fails log at
// warn, everything else at info; the first batch logs each check's boot
// baseline). warn↔pass flapping becomes visible without scraping /readyz.
func checkTransitions() func(health.Response) {
	var (
		mu   sync.Mutex
		last = map[string]health.Status{}
	)

	return func(resp health.Response) {
		mu.Lock()
		defer mu.Unlock()

		names := make([]string, 0, len(resp.Checks))
		for name := range resp.Checks {
			names = append(names, name)
		}

		sort.Strings(names)

		for _, name := range names {
			check := resp.Checks[name]

			previous, seen := last[name]
			last[name] = check.Status

			if seen && previous == check.Status {
				continue
			}

			log := slog.Info
			if check.Status == health.StatusFail {
				log = slog.Warn
			}

			log(
				"compliance-hub: health check status changed",
				"check",
				name,
				"status",
				check.Status,
				"error",
				check.Error,
			)
		}
	}
}

// cacheDirForDB returns the directory to watch disk headroom on, or ""
// when no cache database is configured.
func cacheDirForDB(dbPath string) string {
	if dbPath == "" {
		return ""
	}

	return filepath.Dir(dbPath)
}

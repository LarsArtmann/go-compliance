package cli

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
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
)

// buildHealthProbe wires the hub's health probe: liveness is always pass;
// readiness/startup gate on published evidence; refresh-freshness warns
// when the derive loop stops making progress; disk headroom keeps the
// SQLite cache honest.
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
	}

	if cacheDir != "" {
		const minFreeBytes = 128 << 20 // 128 MiB is plenty to append runs

		checksByName["cache-disk"] = checks.Disk(cacheDir, minFreeBytes)
	}

	probe := health.NewChecks(checksByName,
		health.WithVersion(version),
		health.WithCriticalServices(evidenceCheckName),
	)

	if err := probe.Start(ctx); err != nil {
		return nil, err
	}

	return probe, nil
}

// cacheDirForDB returns the directory to watch disk headroom on, or ""
// when no cache database is configured.
func cacheDirForDB(dbPath string) string {
	if dbPath == "" {
		return ""
	}

	return filepath.Dir(dbPath)
}

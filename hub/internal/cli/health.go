package cli

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/larsartmann/go-health"
	"github.com/larsartmann/go-health/checks"

	"github.com/larsartmann/go-compliance/hub/internal/events"
)

// evidenceCheckName is the one critical service: the hub has folded
// evidence and published state. Until it passes, /readyz and /startupz
// report not-ready — an evidence failure must not restart the process
// (liveness stays pass regardless).
const evidenceCheckName = "evidence"

// ErrEvidenceNotPublished is the evidence check's failure: the hub booted
// but has not folded evidence yet (source error or first query pending).
var ErrEvidenceNotPublished = errors.New("evidence not published yet")

// buildHealthProbe wires the hub's health probe: liveness is always pass;
// readiness/startup gate on published evidence; disk headroom keeps the
// SQLite cache honest.
func buildHealthProbe(version string, hub *events.Hub, cacheDir string) (*health.Probe, error) {
	checksByName := map[string]health.CheckFunc{
		evidenceCheckName: func(_ context.Context) error {
			if hub.Rows() == nil {
				return ErrEvidenceNotPublished
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

	if err := probe.Start(context.Background()); err != nil {
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

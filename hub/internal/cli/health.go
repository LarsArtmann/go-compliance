package cli

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/LarsArtmann/go-daemon"
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

// serveListener binds, notifies systemd, and serves until ctx cancels.
// Listener-first ordering means NotifyReady fires only once the socket
// accepts traffic; under a systemd WatchdogSec the watchdog loop keeps the
// unit from being killed mid-serve.
func serveListener(
	ctx context.Context, server *http.Server, address string, logger *slog.Logger,
) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	daemon.NotifyReady(logger)

	if interval, ok := daemon.WatchdogInterval(logger); ok {
		go daemon.Watchdog(ctx, logger, interval)
	}

	serveError := make(chan error, 1)
	go func() { serveError <- server.Serve(listener) }()

	select {
	case <-ctx.Done():
		daemon.NotifyStopping(logger)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = server.Shutdown(shutdownCtx)

		return nil
	case err := <-serveError:
		return err
	}
}

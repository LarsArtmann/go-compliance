package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"github.com/larsartmann/go-health"

	"github.com/larsartmann/go-compliance"
	"github.com/larsartmann/go-compliance/hub/internal/dashboard"
	"github.com/larsartmann/go-compliance/hub/internal/events"
)

// runServe runs the hub HTTP server: the realtime dashboard (page, DataStar
// patch stream, SDK bundle), the evidence JSON API, and the health probe
// surface (/healthz /readyz /startupz /version).
func runServe(out, errW io.Writer) func(context.Context, *Config, ServeFlags) error {
	return func(ctx context.Context, _ *Config, flags ServeFlags) error {
		evidence, since, staleAfter, err := queryEvidence(ctx, flags.QueryFlags)
		if err != nil {
			return err
		}

		hub := events.NewHub()
		if _, err := hub.PublishSnapshot(events.RowsFromEvidence(evidence, time.Now(), staleAfter)); err != nil {
			return err
		}

		cacheDir := ""
		if flags.DB != "" {
			cacheDir = filepath.Dir(flags.DB)
		}

		probe, err := buildHealthProbe(Version, hub, cacheDir)
		if err != nil {
			return err
		}

		board := dashboard.New(hub)

		mux := http.NewServeMux()
		board.RegisterRoutesAt(dashboard.Routes{
			Page:   "/",
			Stream: "/events",
			Script: "/datastar.js",
		}, mux)

		probe.RegisterRoutes(mux, health.DefaultRoutes())

		mux.HandleFunc("GET /v1/compliance", func(w http.ResponseWriter, _ *http.Request) {
			report := compliance.BuildV1Report(evidence, since, staleAfter, Version, time.Now())

			encoded, err := compliance.MarshalV1(report)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)

				return
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(encoded)
		})

		server := &http.Server{
			Addr:              flags.Addr,
			Handler:           mux,
			ReadHeaderTimeout: 10 * time.Second,
		}

		fmt.Fprintf(out, "compliance-hub listening on http://%s (localhost-only default; see docs/decisions)\n", flags.Addr)

		serveError := serveListener(ctx, server, flags.Addr, slog.Default())

		probe.MarkShuttingDown()
		probe.Shutdown()
		_ = hub.Shutdown(context.Background())

		if serveError != nil {
			fmt.Fprintf(errW, "serve: %v\n", serveError)
		}

		return serveError
	}
}

package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/larsartmann/go-health"

	"github.com/larsartmann/go-compliance"
	"github.com/larsartmann/go-compliance/hub/internal/cache"
	"github.com/larsartmann/go-compliance/hub/internal/dashboard"
	"github.com/larsartmann/go-compliance/hub/internal/discovery"
	"github.com/larsartmann/go-compliance/hub/internal/events"
)

const (
	// readHeaderTimeout bounds slowloris-style header stalls.
	readHeaderTimeout = 10 * time.Second
)

// securityHeaders is the browser-hardening floor on EVERY response:
// nosniff keeps user-controlled content (repo names in JSON, evidence
// bodies) from being reinterpreted as executable MIME types.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

// runServe runs the hub HTTP server: the realtime dashboard (page, DataStar
// patch stream, SDK bundle), the evidence JSON API, and the health probe
// surface (/healthz /readyz /startupz /version).
func runServe(out, errW io.Writer) func(context.Context, *Config, ServeFlags) error {
	return func(ctx context.Context, _ *Config, flags ServeFlags) error {
		window, err := parseSinceWindow(flags.Since)
		if err != nil {
			return err
		}

		staleAfter, err := parseStaleAfter(flags.StaleAfter)
		if err != nil {
			return err
		}

		source, closer, err := resolveSource(ctx, flags.QueryFlags)
		if err != nil {
			return err
		}

		if closer != nil {
			defer func() { _ = closer.Close() }()
		}

		hub := events.NewHub()

		// Live rows: the source stays open and every trigger (cache write,
		// discovery change, safety-net tick) re-derives and publishes.
		refresher := newRefresher(hub, source, window, staleAfter, flags.WorstSteps, flags.Step)
		if err := refresher.Refresh(ctx); err != nil {
			return err
		}

		if store, ok := source.(*cache.Store); ok {
			store.OnWrite(func(compliance.RepoName) { refresher.Kick() })
		}

		// Filesystem triggers: cross-process evidence writes (WAL sidecars,
		// atomic renames) kick the refresher within ~1-2s; the safety net
		// stays as the backstop.
		watcher := watchEvidence(ctx, flags.QueryFlags, refresher)

		// These loops start before the remaining fallible wiring (probe,
		// mux) on purpose: every trigger they fire only Kicks the
		// refresher, whose publishes land in the hub before any SSE
		// subscriber exists — late subscribers still get the snapshot.
		// If a later setup step fails and runServe returns early, ctx
		// cancellation (or process exit) reaps both goroutines.
		go refresher.Run(ctx)
		go runDiscoveryWatch(ctx, discovery.New(discovery.ResolveSocketPath()), refresher)

		probe, err := buildHealthProbe(ctx, Version, hub, refresher, watcher, cacheDirForDB(flags.DB))
		if err != nil {
			return err
		}

		slog.Info("compliance-hub: health probe started (startup latch set)", "version", Version)

		board := dashboard.New(hub)

		mux := http.NewServeMux()
		board.RegisterRoutesAt(dashboard.Routes{
			Page:   "/",
			Stream: "/events",
			Script: "/datastar.js",
			CSS:    "/app.css",
		}, mux)

		// Catch-all: unknown paths get the styled 404 (the dashboard page
		// owns the root exactly, so /anything lands here, not on the fleet).
		mux.Handle("/", board.NotFoundHandler())

		probe.RegisterRoutes(mux, health.DefaultRoutes())

		mux.HandleFunc("GET /v1/compliance", func(w http.ResponseWriter, _ *http.Request) {
			evidence := refresher.Evidence()
			if evidence == nil {
				http.Error(w, "evidence not published yet", http.StatusServiceUnavailable)

				return
			}

			report := compliance.BuildV1Report(evidence, time.Now().Add(-window), staleAfter, Version, time.Now())

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
			Handler:           securityHeaders(mux),
			ReadHeaderTimeout: readHeaderTimeout,
		}

		if !isLoopbackBind(flags.Addr) {
			fmt.Fprintf(
				errW,
				"warning: binding non-localhost %s exposes fleet evidence; see docs/decisions/bind-posture.md\n",
				flags.Addr,
			)
		}

		fmt.Fprintf(out, "compliance-hub listening on %s\n", flags.Addr)

		serveError := serveListener(ctx, server, flags.Addr, slog.Default())

		probe.MarkShuttingDown()
		probe.Shutdown()
		// Background, not ctx: ctx is cancelled at this point, and the
		// subscriber drain needs its own live deadline.
		_ = hub.Shutdown(context.Background()) //nolint:contextcheck

		if serveError != nil {
			fmt.Fprintf(errW, "serve: %v\n", serveError)
		}

		return serveError
	}
}

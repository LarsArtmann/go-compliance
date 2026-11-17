package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/larsartmann/go-compliance"
)

// runServe runs the hub HTTP server. The M7 surface is the evidence JSON
// API; the realtime stack (SSE, patches, UI) replaces the internals in
// the hub milestone.
func runServe(out, errW io.Writer) func(context.Context, *Config, ServeFlags) error {
	return func(ctx context.Context, _ *Config, flags ServeFlags) error {
		evidence, _, staleAfter, err := queryEvidence(ctx, flags.QueryFlags)
		if err != nil {
			return err
		}

		mux := http.NewServeMux()

		mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{"status":"ok"}`)
		})

		mux.HandleFunc("GET /v1/compliance", func(w http.ResponseWriter, r *http.Request) {
			since, err := parseSinceDuration(flags.Since)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

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

		go func() {
			<-ctx.Done()

			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_ = server.Shutdown(shutdownCtx)
		}()

		fmt.Fprintf(out, "compliance-hub listening on http://%s (localhost-only default; see docs/decisions)\n", flags.Addr)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(errW, "serve: %v\n", err)

			return err
		}

		return nil
	}
}

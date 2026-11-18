package dashboard

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/larsartmann/go-compliance/hub/internal/events"
)

// Routes names the HTTP surface RegisterRoutesAt wires. Empty fields fall
// back to the defaults.
type Routes struct {
	// Page serves the HTML dashboard (and, with Accept: application/json,
	// the current state as JSON).
	Page string
	// Stream serves the DataStar patch stream.
	Stream string
	// Script serves the pinned DataStar bundle.
	Script string
}

// DefaultRoutes returns the canonical single-service layout.
func DefaultRoutes() Routes {
	return Routes{Page: "/", Stream: "/events", Script: "/datastar.js"}
}

func (r Routes) withDefaults() Routes {
	defaults := DefaultRoutes()
	if r.Page == "" {
		r.Page = defaults.Page
	}

	if r.Stream == "" {
		r.Stream = defaults.Stream
	}

	if r.Script == "" {
		r.Script = defaults.Script
	}

	return r
}

// RegisterRoutes wires the dashboard at the default routes: one call embeds
// the whole surface into any service's mux.
func (d *Dashboard) RegisterRoutes(mux *http.ServeMux) {
	d.RegisterRoutesAt(DefaultRoutes(), mux)
}

// RegisterRoutesAt wires the dashboard at explicit routes (e.g. behind a
// base path).
func (d *Dashboard) RegisterRoutesAt(routes Routes, mux *http.ServeMux) {
	routes = routes.withDefaults()

	mux.HandleFunc("GET "+routes.Script, d.ScriptHandler().ServeHTTP)
	mux.HandleFunc("GET "+routes.Stream, d.StreamHandler().ServeHTTP)
	mux.Handle("GET "+routes.Page, d.pageHandler(routes))
}

// pageHandler renders the HTML page; a client whose Accept header prefers
// application/json gets the current state as JSON instead — one URL, both
// audiences.
func (d *Dashboard) pageHandler(routes Routes) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rows := d.hub.Rows()

		if preferJSON(r.Header.Get("Accept")) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(eventsSnapshotJSON(rows)))
			return
		}

		stateFilter := r.URL.Query().Get("state")
		if stateFilter != "" && !validStateFilter(stateFilter) {
			stateFilter = ""
		}

		model := ViewModel{
			Title:       d.title,
			StreamURL:   routes.Stream,
			ScriptSrc:   routes.Script,
			Rows:        filterRowsByState(rows, stateFilter),
			StateFilter: stateFilterOrAll(stateFilter),
			RenderedAt:  time.Now(),
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = View(model).Render(r.Context(), w)
	})
}

// preferJSON reports whether the Accept header prefers application/json
// over text/html by q-value (defaulting to HTML for browsers and wildcard
// accepts).
func preferJSON(accept string) bool {
	if accept == "" {
		return false
	}

	htmlQuality, jsonQuality := -1.0, -1.0

	for part := range strings.SplitSeq(accept, ",") {
		mediaType := part
		if parameters := strings.Index(mediaType, ";"); parameters >= 0 {
			mediaType = mediaType[:parameters]
		}

		quality := 1.0
		for parameter := range strings.SplitSeq(part, ";") {
			parameter = strings.TrimSpace(parameter)
			if !strings.HasPrefix(parameter, "q=") {
				continue
			}

			if parsed, err := strconv.ParseFloat(strings.TrimPrefix(parameter, "q="), 64); err == nil {
				quality = parsed
			}
		}

		switch strings.TrimSpace(mediaType) {
		case "text/html":
			htmlQuality = quality
		case "application/json":
			jsonQuality = quality
		}
	}

	return jsonQuality > htmlQuality && jsonQuality >= 0
}

// validStateFilter accepts the canonical state names plus "all".
func validStateFilter(filter string) bool {
	switch filter {
	case "all", "compliant", "failing", "stale", "never-covered", "unknown":
		return true
	default:
		return false
	}
}

func stateFilterOrAll(filter string) string {
	if filter == "" {
		return "all"
	}

	return filter
}

// filterRowsByState keeps rows matching the filter; empty keeps everything.
func filterRowsByState(rows []events.RepoRow, filter string) []events.RepoRow {
	if filter == "" || filter == "all" {
		return rows
	}

	filtered := make([]events.RepoRow, 0, len(rows))
	for _, row := range rows {
		if row.State.String() == filter {
			filtered = append(filtered, row)
		}
	}

	return filtered
}

// eventsSnapshotJSON renders the hub's current state in the same shape the
// JSON /events snapshot events carry.
func eventsSnapshotJSON(rows []events.RepoRow) string {
	return events.SnapshotJSON(rows)
}

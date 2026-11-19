package dashboard

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/larsartmann/go-datastar"

	"github.com/larsartmann/go-compliance"
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
	// CSS serves the compiled Tailwind artifact.
	CSS string
}

// DefaultRoutes returns the canonical single-service layout.
func DefaultRoutes() Routes {
	return Routes{Page: "/", Stream: "/events", Script: "/datastar.js", CSS: cssRoute}
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

	if r.CSS == "" {
		r.CSS = defaults.CSS
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

	// "/" alone is Go-mux-speak for EVERY path; the dashboard must own the
	// root exactly so unknown paths fall through to the host's 404 instead
	// of silently rendering the fleet at /anything.
	page := routes.Page
	if page == "/" {
		page = "/{$}"
	}

	mux.HandleFunc("GET "+routes.CSS, d.CSSHandler().ServeHTTP)
	mux.HandleFunc("GET "+routes.Script, d.ScriptHandler().ServeHTTP)
	mux.HandleFunc("GET "+routes.Stream, d.StreamHandler().ServeHTTP)
	mux.Handle("GET "+page, d.pageHandler(routes))
}

// NotFoundHandler renders the styled 404 page; serve wires it as the
// mux catch-all so unknown paths fail visibly instead of falling into a
// bare-text default.
func (d *Dashboard) NotFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Robots-Tag", "noindex")
		w.WriteHeader(http.StatusNotFound)
		_ = NotFound(d.title+" — page not found", versionedCSSHref(cssRoute)).Render(r.Context(), w)
	})
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
			ScriptSrc:   versionedScriptSrc(routes.Script),
			CSSSrc:      versionedCSSHref(routes.CSS),
			Rows:        filterRowsByState(rows, stateFilter),
			StateFilter: stateFilterOrAll(stateFilter),
			RenderedAt:  time.Now(),
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = View(model).Render(r.Context(), w)
	})
}

// versionedScriptSrc pins the embedded DataStar client version as a cache
// buster on the self-hosted bundle URL (ScriptTagVersioned semantics; the
// SDKScript component renders the URL we hand it). ScriptHandler caches the
// bundle for 24h, so an unversioned URL would serve yesterday's client to
// returning browsers after an upgrade.
func versionedScriptSrc(path string) string {
	if path == "" {
		return ""
	}

	return path + "?v=" + datastar.Version()
}

// versionedCSSHref pins the compiled stylesheet's content hash as a cache
// buster: the artifact is served immutable for a day, so an upgrade MUST
// change the URL or returning browsers keep yesterday's styles.
func versionedCSSHref(path string) string {
	if path == "" {
		return ""
	}

	return path + "?v=" + appCSSVersion
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

// validStateFilter accepts the canonical state names plus the
// dashboard-only "all" and "unknown". The four states delegate to the
// core enum so a newly added state is filterable without touching this
// check.
func validStateFilter(filter string) bool {
	if filter == "all" || filter == "unknown" {
		return true
	}

	return compliance.State(filter).IsValid()
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

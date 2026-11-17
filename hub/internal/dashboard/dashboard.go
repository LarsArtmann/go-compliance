// Package dashboard projects the hub's event stream onto DataStar SSE
// patches: fleet table rows as element patches, state counts as signal
// patches. The JSON /events stream stays machine-facing; this package is
// the browser-facing twin of the same Hub.
package dashboard

import (
	"encoding/json/v2"
	"fmt"
	"net/http"
	"time"

	"github.com/larsartmann/go-datastar"
	dstarstatic "github.com/larsartmann/go-datastar/static"
	"github.com/larsartmann/go-sse"

	"github.com/larsartmann/go-compliance/hub/internal/events"
)

// DefaultRowsSelector is the CSS selector of the fleet rows region every
// snapshot patch targets.
const DefaultRowsSelector = "#fleet-rows"

const (
	heartbeatInterval = 15 * time.Second

	scriptContentType = "text/javascript; charset=utf-8"
	scriptCacheControl = "public, max-age=86400"
)

// stateSignals is the signals payload patched on every state change:
// stat cards and filter chips read it reactively.
type stateSignals struct {
	GeneratedAt string         `json:"generatedAt"`
	Total       int            `json:"total"`
	Counts      map[string]int `json:"counts"`
}

// Dashboard serves the DataStar side of the hub: a patch stream and the
// DataStar JS bundle.
type Dashboard struct {
	hub          *events.Hub
	rowsSelector string
}

// Option configures a Dashboard at construction.
type Option func(*Dashboard)

// WithRowsSelector overrides the fleet rows region selector.
func WithRowsSelector(selector string) Option {
	return func(d *Dashboard) {
		if selector != "" {
			d.rowsSelector = selector
		}
	}
}

// New returns a Dashboard projecting the hub's events as DataStar patches.
func New(hub *events.Hub, opts ...Option) *Dashboard {
	dashboard := &Dashboard{hub: hub, rowsSelector: DefaultRowsSelector}
	for _, opt := range opts {
		opt(dashboard)
	}

	return dashboard
}

// StreamHandler serves the SSE patch stream: the current state first (so
// fresh clients render immediately), then live patches with a comment
// heartbeat.
func (d *Dashboard) StreamHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stream := sse.NewStream(w, r)
		defer func() { _ = stream.Close() }()

		if !d.sendState(stream, d.hub.Rows()) {
			return
		}

		feed, cancel := d.hub.SubscribeFeed()
		defer cancel()

		go stream.Heartbeat(stream.Context(), heartbeatInterval)

		for {
			select {
			case <-stream.Context().Done():
				return
			case published, ok := <-feed:
				if !ok || !d.sendPublished(stream, published) {
					return
				}
			}
		}
	})
}

// ScriptHandler serves the embedded DataStar JS bundle with day-long
// caching: the bundle is version-pinned by the go-datastar/static module.
func (d *Dashboard) ScriptHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", scriptContentType)
		w.Header().Set("Cache-Control", scriptCacheControl)
		_, _ = w.Write(dstarstatic.Bytes())
	})
}

// sendPublished projects one hub event; false stops the stream (client
// gone).
func (d *Dashboard) sendPublished(stream *sse.Stream, published events.Published) bool {
	switch {
	case published.Snapshot != nil:
		return d.sendState(stream, published.Snapshot.Rows)
	case published.Update != nil:
		return d.sendUpdate(stream, published.Update)
	default:
		return true
	}
}

// sendState writes the signals patch (counts) then the full rows patch.
func (d *Dashboard) sendState(stream *sse.Stream, rows []events.RepoRow) bool {
	if !d.sendSignals(stream, rows != nil, events.CountStates(rows), len(rows), time.Now()) {
		return false
	}

	if rows == nil {
		return true
	}

	patch, err := datastar.ElementsFromTempl(FleetRows(rows),
		datastar.WithSelector(d.rowsSelector),
		datastar.WithMode(datastar.ElementPatchModeInner),
	)
	if err != nil {
		return true // a render hiccup never kills the stream
	}

	return stream.Send(patch.Event()) == nil
}

// sendUpdate writes the refreshed signals then the smallest element
// patches: appends for added repos, replaces for changed rows, removes for
// disappeared repos.
func (d *Dashboard) sendUpdate(stream *sse.Stream, update *events.Update) bool {
	if !d.sendSignals(stream, false, update.Counts, update.Total, update.GeneratedAt) {
		return false
	}

	if len(update.Added) > 0 {
		patch, err := datastar.ElementsFromTempl(FleetRows(update.Added),
			datastar.WithSelector(d.rowsSelector),
			datastar.WithMode(datastar.ElementPatchModeAppend),
		)
		if err == nil && stream.Send(patch.Event()) != nil {
			return false
		}
	}

	for _, row := range update.Changed {
		patch, err := datastar.ElementsFromTempl(FleetRow(row),
			datastar.WithSelector(rowSelector(row.Repo)),
			datastar.WithMode(datastar.ElementPatchModeReplace),
		)
		if err == nil && stream.Send(patch.Event()) != nil {
			return false
		}
	}

	for _, repo := range update.Removed {
		patch := datastar.NewElementsPatch("",
			datastar.WithSelector(rowSelector(repo)),
			datastar.WithMode(datastar.ElementPatchModeRemove),
		)
		if stream.Send(patch.Event()) != nil {
			return false
		}
	}

	return true
}

func (d *Dashboard) sendSignals(stream *sse.Stream, includeRows bool, counts map[string]int, total int, generatedAt time.Time) bool {
	signals := stateSignals{
		GeneratedAt: generatedAt.UTC().Format(time.RFC3339),
		Total:       total,
		Counts:      counts,
	}
	if signals.Counts == nil {
		signals.Counts = map[string]int{}
	}

	encoded, err := json.Marshal(signals, json.Deterministic(true))
	if err != nil {
		return true
	}

	patch := datastar.SignalsPatch{Signals: encoded}

	return stream.Send(patch.Event()) == nil
}

// rowSelector targets one fleet row by its repo attribute. Go's %q quoting
// doubles as a CSS attribute-selector value.
func rowSelector(repo string) string {
	return fmt.Sprintf("tr[data-repo=%q]", repo)
}

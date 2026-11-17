package dashboard

import "github.com/larsartmann/go-compliance/hub/internal/events"

// ViewModel carries everything the page render needs; the templ components
// stay dumb projections of it.
type ViewModel struct {
	Title     string
	StreamURL string
	// ScriptSrc self-hosts the DataStar bundle; empty falls back to the
	// jsdelivr CDN.
	ScriptSrc string
	// Rows is the initial fleet state, rendered server-side so the page
	// paints before the stream connects.
	Rows []events.RepoRow
}

// Element identifiers the patch stream targets. The rows selector doubles
// as the TableProps.BodyID.
const (
	// FleetRegionID wraps the live table region.
	FleetRegionID = "fleet-region"
	// StatCardsID wraps the stat card strip; state refreshes re-render it.
	StatCardsID = "stat-cards"
)

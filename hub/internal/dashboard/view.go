package dashboard

import (
	"time"

	"github.com/larsartmann/go-compliance/hub/internal/events"
)

// ViewModel carries everything the page render needs; the templ components
// stay dumb projections of it.
type ViewModel struct {
	Title     string
	StreamURL string
	// ScriptSrc self-hosts the DataStar bundle; empty falls back to the
	// jsdelivr CDN.
	ScriptSrc string
	// CSSSrc links the compiled Tailwind artifact (content-hash versioned
	// so upgrades are fresh cache keys); empty suppresses the stylesheet
	// (embedders that own styling).
	CSSSrc string
	// Rows is the initial fleet state, rendered server-side so the page
	// paints before the stream connects.
	Rows []events.RepoRow
	// StateFilter is the initially-selected filter chip ("all" when
	// empty); it seeds the stateFilter signal and filters the rows.
	StateFilter string
	// RenderedAt is the page render instant, shown as a relative
	// "rendered X ago" freshness line. Its RelativeTime is the ONE
	// AutoRefresh carrier on the page: the singleton script it ships
	// ticks every row's data-tc-relative cell (rows opt into the
	// attribute without the script). A frozen "rendered 2 hours ago"
	// next to fresh rows also makes a dead SSE stream visible.
	RenderedAt time.Time
}

// Element identifiers the patch stream targets. The rows selector doubles
// as the TableProps.BodyID.
const (
	// FleetRegionID wraps the live table region.
	FleetRegionID = "fleet-region"
	// FleetSummaryID wraps the headline, state bar, and filter chips; state
	// refreshes re-render it.
	FleetSummaryID = "fleet-summary"
)

const (
	// fleetTableClass is the stylesheet hook that stacks the fleet rows once
	// the table gets narrower than its columns.
	fleetTableClass = "fleet-table"
)

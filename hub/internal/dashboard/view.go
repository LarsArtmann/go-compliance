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
	// paints before the stream connects. StateFilter may have narrowed it.
	Rows []events.RepoRow
	// Counts and Total describe the WHOLE fleet, whatever StateFilter
	// narrowed Rows to: the headline and the chips always speak for the
	// fleet.
	Counts map[string]int
	Total  int
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
	// fleetTableClass is the stylesheet hook on the <table>: it stacks the
	// fleet rows once the table gets narrower than its columns, and it marks
	// the table whose scroll box (the library wrapper around it) gets a
	// height in wide layouts.
	fleetTableClass = "fleet-table"

	// renderedStatusID marks the "rendered X ago" freshness line in the
	// page header; the browser harness reads its contrast from it.
	renderedStatusID = "rendered-status"

	// pageBodyClass paints the page canvas. layout.Base applies no body
	// class of its own, so without it a dark page falls back to the browser's
	// default canvas, which differs per browser. The dashboard names its own:
	// white, and in dark mode the darkest gray, one step below the table's
	// raised surfaces.
	pageBodyClass = "bg-white text-gray-900 antialiased dark:bg-gray-950 dark:text-gray-100"

	// pageThemeColor and pageDarkThemeColor tint the browser's own UI (the
	// address bar on a phone) through the theme-color meta tags, one per color
	// scheme: the hub's accent-600, and the dark canvas of pageBodyClass
	// (gray-950). The library's defaults are indigo, which the hub palette
	// does not contain.
	pageThemeColor     = "#1464c6"
	pageDarkThemeColor = "#030712"
)

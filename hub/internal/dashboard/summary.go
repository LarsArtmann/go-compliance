package dashboard

import (
	"encoding/json/v2"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize/english"

	"github.com/larsartmann/go-compliance"
)

// fleetStory is the headline above the table, split into the words around
// the one figure the eye should land on first.
type fleetStory struct {
	Lead string
	// Figure is the number that carries the story.
	Figure string
	// Alert colors the figure as failing.
	Alert bool
	Tail  string
}

// newFleetStory states the failing ratio of a fleet in one sentence.
func newFleetStory(counts map[string]int, total int) fleetStory {
	switch {
	case total == 0:
		return fleetStory{Lead: "No fleet evidence yet"}
	case counts[compliance.StateFailing.String()] == 0:
		return fleetStory{
			Lead:   "All",
			Figure: strconv.Itoa(total),
			Tail: english.PluralWord(total, "repo", "repos") + " " +
				english.PluralWord(total, "is", "are") + " compliant",
		}
	default:
		failing := counts[compliance.StateFailing.String()]

		return fleetStory{
			Figure: strconv.Itoa(failing),
			Alert:  true,
			Tail: "of " + strconv.Itoa(total) + " " + english.PluralWord(total, "repo", "repos") +
				" " + english.PluralWord(failing, "is", "are") + " failing",
		}
	}
}

// figureClass colors the headline figure: failing reads as a warning, a
// clean fleet reads as plain text.
func (s fleetStory) figureClass() string {
	if s.Alert {
		return "text-red-700 dark:text-red-400"
	}

	return "text-inherit"
}

// stateSegment is one colored share of the state bar.
type stateSegment struct {
	// Label names the state in the bar's accessible description.
	Label string
	Count int
	Class string
}

// stateBarOrder ranks the states worst first and names each one's color.
//
//nolint:gochecknoglobals // immutable rank ladder (keys carry the meaning)
var stateBarOrder = []struct {
	state compliance.State
	label string
	class string
}{
	{compliance.StateFailing, "failing", "bg-red-500 dark:bg-red-400"},
	{compliance.StateStale, "stale", "bg-amber-500 dark:bg-amber-400"},
	{compliance.StateNeverCovered, "never covered", "bg-gray-400 dark:bg-gray-500"},
	{compliance.StateCompliant, "compliant", "bg-green-500 dark:bg-green-400"},
}

// stateSegments lists the states that hold at least one repo, worst first.
func stateSegments(counts map[string]int) []stateSegment {
	segments := make([]stateSegment, 0, len(stateBarOrder))

	for _, slot := range stateBarOrder {
		if count := counts[slot.state.String()]; count > 0 {
			segments = append(segments, stateSegment{Label: slot.label, Count: count, Class: slot.class})
		}
	}

	return segments
}

// stateBarLabel describes the bar in exact counts for assistive technology.
func stateBarLabel(segments []stateSegment, total int) string {
	parts := make([]string, 0, len(segments))

	for _, segment := range segments {
		parts = append(parts, strconv.Itoa(segment.Count)+" "+segment.Label)
	}

	return strings.Join(parts, ", ") + " of " + strconv.Itoa(total) + " " +
		english.PluralWord(total, "repo", "repos")
}

// chipCount is the number a filter chip selects: every repo for "all", the
// repos in that state otherwise.
func chipCount(state string, counts map[string]int, total int) int {
	if state == "all" {
		return total
	}

	return counts[state]
}

// filterSignals is the page's initial signal seed. It carries the counts
// as well as the selection, so no reactive expression reads a missing
// signal before the stream connects.
type filterSignals struct {
	Selected string         `json:"stateFilter"`
	Total    int            `json:"total"`
	Counts   map[string]int `json:"counts"`
}

// filterSignalsJSON encodes the initial signal seed, or "" if encoding
// failed (the page then starts without a seed and the stream supplies it).
func filterSignalsJSON(counts map[string]int, total int, selected string) string {
	if counts == nil {
		counts = map[string]int{}
	}

	encoded, err := json.Marshal(
		filterSignals{Selected: selected, Total: total, Counts: counts},
		json.Deterministic(true),
	)
	if err != nil {
		return ""
	}

	return string(encoded)
}

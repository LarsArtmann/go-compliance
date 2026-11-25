package dashboard

import (
	"sort"
	"strconv"

	"github.com/dustin/go-humanize/english"
)

// findingsLabel pluralizes the gate-findings count for the fallback cell.
func findingsLabel(count int) string {
	return strconv.Itoa(count) + " " + english.PluralWord(count, "finding", "findings")
}

// severityLabel upper-cases a severity for the compact ERR/WARN/INFO cell
// display, keeping unknown severities readable as-is.
func severityLabel(severity string) string {
	switch severity {
	case "error", "critical":
		return "ERR"
	case "warning":
		return "WARN"
	case "info":
		return "INFO"
	default:
		return severity
	}
}

// severityTone picks the text tone per severity: errors red, warnings
// amber, everything else muted. Every pair clears 4.5:1 on its surface in
// both themes.
func severityTone(severity string) string {
	switch severity {
	case "error", "critical":
		return "text-red-700 dark:text-red-400"
	case "warning":
		return "text-amber-800 dark:text-amber-300"
	default:
		return "text-gray-600 dark:text-gray-300"
	}
}

// severityOrder ranks known severities for deterministic rendering; unknown
// severities sort after the known ones, alphabetically.
//
//nolint:mnd,gochecknoglobals // rank ladder (keys carry the meaning); intentional immutable lookup
var severityOrder = map[string]int{
	"error":    0,
	"critical": 1,
	"warning":  2,
	"info":     3,
	"note":     4,
	"unknown":  5,
}

// sortedSeverities returns the map's keys in canonical severity order so
// renders (and therefore patch diffs) are deterministic.
func sortedSeverities(severities map[string]int) []string {
	keys := make([]string, 0, len(severities))
	for severity := range severities {
		keys = append(keys, severity)
	}

	sort.Slice(keys, func(i, j int) bool {
		left, right := keys[i], keys[j]
		leftRank, leftKnown := severityOrder[left]
		rightRank, rightKnown := severityOrder[right]

		switch {
		case leftKnown && rightKnown:
			if leftRank != rightRank {
				return leftRank < rightRank
			}
		case leftKnown:
			return true
		case rightKnown:
			return false
		}

		return left < right
	})

	return keys
}

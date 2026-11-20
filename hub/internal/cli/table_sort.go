package cli

import (
	"sort"

	"github.com/larsartmann/go-compliance"
)

// sortBy is a parsed table --sort-by value: the tiebreak INSIDE the
// state rank, mirroring the producer's flag semantics (minus the
// producer-only importance key, which the hub has no metadata for).
type sortBy string

const (
	sortByFindings sortBy = "findings"
	sortByFailures sortBy = "failures"
	sortByName     sortBy = "name"
)

// parseSortBy validates a --sort-by value. cmdguard's values tag guards
// the CLI; this guards embedders calling runTable directly.
func parseSortBy(raw string) (sortBy, error) {
	switch sortBy(raw) {
	case sortByFindings, sortByFailures, sortByName:
		return sortBy(raw), nil
	default:
		return "", compliance.NewValidationError(
			"invalid --sort-by value %q (supported: findings, failures, name)", raw)
	}
}

// sortReposBy re-orders the report rows for the table's --sort-by flag:
// the STATE RANK STAYS PRIMARY (the triage invariant the producer
// shares), the chosen key breaks ties inside a rank, and equal keys keep
// BuildV1Report's canonical order (failures desc, then name) because the
// sort is stable over already-canonicalized rows.
func sortReposBy(repos []compliance.V1Repo, order sortBy, findingsByRepo map[string]int) {
	sort.SliceStable(repos, func(i, j int) bool {
		if rankI, rankJ := compliance.State(repos[i].State).
			Rank(),
			compliance.State(repos[j].State).
				Rank(); rankI != rankJ {
			return rankI < rankJ
		}

		switch order {
		case sortByFindings:
			if fi, fj := findingsByRepo[repos[i].Repo], findingsByRepo[repos[j].Repo]; fi != fj {
				return fi > fj
			}
		case sortByName:
			return repos[i].Repo < repos[j].Repo
		case sortByFailures:
			// the canonical order already sorts failures desc, then name
		}

		return false
	})
}

// findingTotalsByRepo maps every repo to its latest run's total finding
// count so the --sort-by findings tiebreak sorts by the run's finding
// count. The gate findings are the wire-portable rollup (the v1 contract
// cannot carry severities, so file-sourced evidence has none); the cache
// and buildflow sources carry the same totals in both shapes.
func findingTotalsByRepo(evidence []compliance.RepoEvidence) map[string]int {
	totals := make(map[string]int, len(evidence))
	for _, record := range evidence {
		total := 0
		for _, n := range record.GateFindings {
			total += n
		}

		totals[record.Repo.Get()] = total
	}

	return totals
}

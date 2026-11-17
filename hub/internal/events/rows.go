// Package events is the hub's realtime backbone: it derives dashboard row
// state from folded evidence, diffs consecutive states, and fans the results
// out to SSE subscribers with Last-Event-ID replay support.
package events

import (
	"time"

	"github.com/larsartmann/go-compliance"
)

// RepoRow is the per-repo live state one dashboard row renders. It is a
// display projection of compliance.RepoEvidence, deliberately narrow: rows
// carry what the table shows, not the full evidence.
type RepoRow struct {
	Repo        string             `json:"repo"`
	State       compliance.State   `json:"state"`
	LastRunAt   time.Time          `json:"lastRunAt"`
	Runs        int64              `json:"runs"`
	Failures    int64              `json:"failures"`
	Findings    int                `json:"findings"`
	Severities  map[string]int     `json:"severities"`
	WorstSteps  []string           `json:"worstSteps"`
	LastVerdict compliance.Verdict `json:"lastVerdict"`
}

// RowsFromEvidence projects folded evidence into dashboard rows, deriving
// each row's state the same way the report does.
func RowsFromEvidence(evidence []compliance.RepoEvidence, now time.Time, staleAfter time.Duration) []RepoRow {
	rows := make([]RepoRow, 0, len(evidence))
	for _, repo := range evidence {
		findings := 0
		for _, count := range repo.GateFindings {
			findings += count
		}

		steps := make([]string, 0, len(repo.WorstSteps))
		for _, step := range repo.WorstSteps {
			steps = append(steps, step.Step)
		}

		rows = append(rows, RepoRow{
			Repo:        repo.Repo.Get(),
			State:       compliance.DeriveState(repo, now, staleAfter),
			LastRunAt:   repo.LastRunAt,
			Runs:        repo.Runs,
			Failures:    repo.Failures,
			Findings:    findings,
			Severities:  repo.FindingSeverities,
			WorstSteps:  steps,
			LastVerdict: repo.LastRunVerdict,
		})
	}

	return rows
}

// DiffRows compares consecutive row states: a row is changed when it is new
// or when any displayed field moved; removed lists repos that disappeared.
// Unchanged rows are omitted so update payloads stay small.
func DiffRows(previous, next []RepoRow) (changed []RepoRow, removed []string) {
	previousByKey := make(map[string]RepoRow, len(previous))
	for _, row := range previous {
		previousByKey[row.Repo] = row
	}

	nextByKey := make(map[string]struct{}, len(next))
	for _, row := range next {
		nextByKey[row.Repo] = struct{}{}

		before, existed := previousByKey[row.Repo]
		if existed && rowsEqual(before, row) {
			continue
		}

		changed = append(changed, row)
	}

	for _, row := range previous {
		if _, stillPresent := nextByKey[row.Repo]; !stillPresent {
			removed = append(removed, row.Repo)
		}
	}

	return changed, removed
}

func rowsEqual(a, b RepoRow) bool {
	if a.Repo != b.Repo || a.State != b.State || a.LastRunAt != b.LastRunAt {
		return false
	}

	if a.Runs != b.Runs || a.Failures != b.Failures || a.Findings != b.Findings {
		return false
	}

	if a.LastVerdict != b.LastVerdict {
		return false
	}

	if len(a.Severities) != len(b.Severities) {
		return false
	}

	for severity, count := range a.Severities {
		if b.Severities[severity] != count {
			return false
		}
	}

	if len(a.WorstSteps) != len(b.WorstSteps) {
		return false
	}

	for i, step := range a.WorstSteps {
		if b.WorstSteps[i] != step {
			return false
		}
	}

	return true
}

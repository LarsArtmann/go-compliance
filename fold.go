package compliance

import (
	"sort"
	"time"
)

// WindowHealth is one repository's window-bounded aggregate: runs, step
// rows, and failures inside the evidence window.
type WindowHealth struct {
	Repo     RepoName
	Runs     int64
	StepRows int64
	Failures int64
	FailRate float64
}

// LatestRun is a repository's most recent run over ALL history: the join
// key that lets a repository which stopped running still report its true
// last verdict and be classified stale.
type LatestRun struct {
	Repo     RepoName
	RepoHash string
	// LastRunAt is the newest recording of the run; the zero value means
	// no rows exist.
	LastRunAt  time.Time
	AnyFailed  bool
	HasVerdict bool
	Version    string
}

// Verdict maps the run's outcome onto the contract vocabulary. Unknown
// when the repository has no rows at all (the consumer renders an
// unknown verdict rather than guessing).
func (r LatestRun) Verdict() Verdict {
	if !r.HasVerdict {
		return VerdictUnknown
	}

	if r.AnyFailed {
		return VerdictFailed
	}

	return VerdictPassed
}

// RepoSteps delivers one repository's failing-step aggregates, worst-first
// as ordered by the source; the fold caps the list.
type RepoSteps struct {
	Repo  RepoName
	Steps []StepEvidence
}

// RepoFindings delivers one repository's latest-run gate findings
// ({step: count}).
type RepoFindings struct {
	Repo     RepoName
	Findings map[string]int
}

// RepoSeverities delivers one repository's latest-run severity counts
// ({severity: count}).
type RepoSeverities struct {
	Repo       RepoName
	Severities map[string]int
}

// RepoVersions delivers one repository's producer versions seen in the
// window. ATTACH-ONLY in the fold: a repository known only through its
// versions is never materialized (version skew is a property of repos
// with runs, not a signal that a repo exists).
type RepoVersions struct {
	Repo     RepoName
	Versions []string
}

// FoldInput carries the source-agnostic evidence parts the fold unions.
// Every part is optional; a source that cannot answer one leaves it nil.
type FoldInput struct {
	// WindowHealth overwrites the window-bounded counters per repo.
	WindowHealth []WindowHealth
	// LatestRuns overwrites the all-history latest-run fields per repo.
	LatestRuns []LatestRun
	// WorstSteps caps at WorstStepLimit (<= 0 → DefaultWorstSteps) and
	// overwrites the per-repo worst-steps list.
	WorstSteps []RepoSteps
	// GateFindings and FindingSeverities attach only when non-empty.
	GateFindings []RepoFindings
	// FindingSeverities attach only when non-empty.
	FindingSeverities []RepoSeverities
	// Versions attach only to already-materialized repositories.
	Versions []RepoVersions
	// WorstStepLimit caps the per-repo worst-steps list.
	WorstStepLimit int
}

// FoldEvidence folds the evidence parts into per-repo evidence.
//
// Repositories are never dropped for being quiet: the union of window
// repos and all-history repos is returned, so the consumer can render
// stale (or never-covered, via discovery) rather than unknown. The output
// is sorted by repository name - deterministic regardless of input order.
func FoldEvidence(input FoldInput) []RepoEvidence {
	limit := input.WorstStepLimit
	if limit <= 0 {
		limit = DefaultWorstSteps
	}

	repos := make(map[RepoName]*RepoEvidence)

	ensure := func(repo RepoName) *RepoEvidence {
		record, ok := repos[repo]
		if !ok {
			record = &RepoEvidence{Repo: repo}
			repos[repo] = record
		}

		return record
	}

	for _, health := range input.WindowHealth {
		record := ensure(health.Repo)
		record.Runs = health.Runs
		record.StepRows = health.StepRows
		record.Failures = health.Failures
		record.FailRate = health.FailRate
	}

	for _, run := range input.LatestRuns {
		record := ensure(run.Repo)
		record.RepoHash = run.RepoHash
		record.LastRunAt = run.LastRunAt
		record.LastRunVerdict = run.Verdict()
		record.LastRunVersion = run.Version
	}

	for _, repoSteps := range input.WorstSteps {
		steps := repoSteps.Steps
		if len(steps) > limit {
			steps = steps[:limit]
		}

		ensure(repoSteps.Repo).WorstSteps = steps
	}

	for _, findings := range input.GateFindings {
		if len(findings.Findings) > 0 {
			ensure(findings.Repo).GateFindings = findings.Findings
		}
	}

	for _, severities := range input.FindingSeverities {
		if len(severities.Severities) > 0 {
			ensure(severities.Repo).FindingSeverities = severities.Severities
		}
	}

	// Versions are ATTACH-ONLY: every repository carrying window versions
	// also carries window health rows from a truthful source, so the loop
	// must never mint a fresh evidence record - a repository that never
	// ran would otherwise materialize out of its unfiltered version rows
	// alone.
	for _, repoVersions := range input.Versions {
		if record, ok := repos[repoVersions.Repo]; ok && len(repoVersions.Versions) > 0 {
			record.Versions = repoVersions.Versions
		}
	}

	names := make([]RepoName, 0, len(repos))
	for repo := range repos {
		names = append(names, repo)
	}

	sort.Slice(names, func(i, j int) bool { return names[i].Get() < names[j].Get() })

	result := make([]RepoEvidence, 0, len(repos))
	for _, name := range names {
		result = append(result, *repos[name])
	}

	return result
}

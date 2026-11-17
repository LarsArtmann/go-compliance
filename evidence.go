package compliance

import "time"

// StepEvidence is one repo+step failure aggregate in the evidence window:
// the worst-steps feed of the compliance contract.
type StepEvidence struct {
	Step      string
	Failures  int64
	TotalRuns int64
	FailRate  float64
}

// RepoEvidence is one repository's complete compliance evidence.
//
// Window fields (Runs/StepRows/Failures/WorstSteps/Versions) are bounded by
// the evidence window; LastRunAt/LastRunVerdict/GateFindings come from the
// repo's ALL-HISTORY latest run, so a repository that stopped running still
// reports its true last verdict and can be classified stale. Under a step
// filter the same fields describe THAT step's rows only (Runs = runs
// containing the step, latest run = the step's latest run).
//
// The zero value means "never measured": DeriveState maps it to
// never-covered. RepoHash/LastRunVersion/FindingSeverities feed display
// surfaces only and are deliberately not part of the v1 wire contract.
type RepoEvidence struct {
	Repo RepoName
	// RepoHash is the latest run's repository path hash: an opaque join
	// key for host-local display data (BuildFlow uses it to mark
	// missing-on-disk checkouts). Not part of the v1 wire contract.
	RepoHash string
	Runs     int64
	StepRows int64
	Failures int64
	FailRate float64
	// LastRunAt is the newest recording of the repo's latest run; the
	// zero value means the repo has no rows at all.
	LastRunAt time.Time
	// LastRunVerdict is the latest run's verdict: passed, failed, or
	// unknown (empty) when no rows exist.
	LastRunVerdict Verdict
	// LastRunVersion is the producer version of the latest run's newest
	// row; empty when no rows. Not part of the v1 wire contract.
	LastRunVersion string
	// GateFindings is the latest run's per-step finding counts
	// ({step: count}).
	GateFindings map[string]int
	// FindingSeverities aggregates the latest run's severity counts
	// ({severity: count}); display surfaces render it as ERR/WARN/INFO
	// columns. Not part of the v1 wire contract.
	FindingSeverities map[string]int
	Versions          []string
	WorstSteps        []StepEvidence
}

// HasEvidence reports whether any evidence at all exists for the
// repository: runs, step rows, or a last-run recording. False means the
// zero RepoEvidence (or an equivalent never-measured row), which derives
// as never-covered.
func (e RepoEvidence) HasEvidence() bool {
	return e.Runs > 0 || e.StepRows > 0 || !e.LastRunAt.IsZero()
}

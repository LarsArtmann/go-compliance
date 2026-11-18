package compliance

import "time"

// State is one repository's compliance state.
//
// The v1 wire contract (schema_version 1) carries exactly compliant,
// failing, and stale; overview mirrors those verbatim and re-derives
// nothing. never-covered is a go-compliance extension for repositories a
// discovery source knows exist but for which no evidence exists at all:
// it is derived locally (never emitted into a v1 document unless the
// producer explicitly builds extended rows).
type State string

const (
	// StateCompliant means the window carries evidence, no failures, and
	// the last run is fresh enough.
	StateCompliant State = "compliant"
	// StateFailing means any failed step row exists in the window (honest
	// fleet semantics - retries each add a row).
	StateFailing State = "failing"
	// StateStale means the last run is older than the staleness
	// threshold: no fresh evidence is itself evidence of neglect.
	StateStale State = "stale"
	// StateNeverCovered means no evidence exists at all: the repository
	// is known (via discovery) but has never been measured.
	StateNeverCovered State = "never-covered"
)

// String returns the wire representation of the state.
func (s State) String() string { return string(s) }

// IsValid reports whether the state is one of the four known states.
//
// art-dupl:accept the case list IS the canonical member set; a shared membership helper would obscure it
func (s State) IsValid() bool {
	switch s {
	case StateCompliant, StateFailing, StateStale, StateNeverCovered:
		return true
	default:
		return false
	}
}

// Rank returns the triage order of the state: failing sorts first, then
// stale, then never-covered, then compliant. Unknown states sort last,
// after compliant. The hub table uses this as the always-primary sort key.
//
//nolint:mnd // rank ladder; the case labels carry the meaning
func (s State) Rank() int {
	switch s {
	case StateFailing:
		return 0
	case StateStale:
		return 1
	case StateNeverCovered:
		return 2
	case StateCompliant:
		return 3
	default:
		return 4
	}
}

// Verdict is the outcome of a repository's latest run: "passed" or
// "failed", or the empty value when the repository has no rows at all
// (the consumer renders an unknown verdict rather than guessing).
type Verdict string

const (
	// VerdictPassed means the latest run completed without failed steps.
	VerdictPassed Verdict = "passed"
	// VerdictFailed means the latest run had at least one failed step.
	VerdictFailed Verdict = "failed"
	// VerdictUnknown means no rows exist: no verdict can be derived.
	VerdictUnknown Verdict = ""
)

// String returns the wire representation of the verdict.
func (v Verdict) String() string { return string(v) }

// ParseState converts a raw wire state string into a State. Unknown values
// are a contract violation, not a silent default: the consumer side
// verifies the enum instead of re-deriving it.
func ParseState(raw string) (State, error) {
	state := State(raw)
	if !state.IsValid() {
		return "", NewValidationError("unknown compliance state %q", raw)
	}

	return state, nil
}

// DeriveState implements the compliance state model.
//
// Any failed step row in the window is failing; a last run older than
// staleAfter is stale; otherwise compliant. staleAfter <= 0 disables the
// staleness arm. A repository with NO evidence at all (no runs, no step
// rows, no last run - the zero RepoEvidence, minted for repositories a
// discovery source knows but evidence does not) is never-covered.
//
// Compatibility: BuildFlow's fold never emits a zero-evidence row (every
// row it produces carries runs or a last-run timestamp), so for every
// input BuildFlow can produce this function returns exactly the v1
// contract states. never-covered is reachable only for discovery-minted
// rows, which the v1 producer never builds.
//
// Step-scoped evidence flows through UNCHANGED: the caller hands over
// evidence whose Failures and LastRunAt already describe the one step
// only, so "failing" means "that step failed in the window" and "stale"
// means "that step hasn't run within staleAfter" - no special casing here.
func DeriveState(evidence RepoEvidence, now time.Time, staleAfter time.Duration) State {
	if !evidence.HasEvidence() {
		return StateNeverCovered
	}

	if evidence.Failures > 0 {
		return StateFailing
	}

	if staleAfter > 0 && !evidence.LastRunAt.IsZero() {
		if now.Sub(evidence.LastRunAt) > staleAfter {
			return StateStale
		}
	}

	return StateCompliant
}

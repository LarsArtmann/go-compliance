package compliance

import (
	"bytes"
	"encoding/json/v2"
	"sort"
	"time"
)

// SchemaVersionV1 is the wire contract version of the compliance JSON
// document. The overview consumer rejects any other value: a changed
// contract needs a conscious bump on both sides, never a silent misparse.
// v1 stays emittable forever.
const SchemaVersionV1 = 1

// Default window parameters, mirroring the fleet sweep cadence: 7 days of
// evidence window, 14 days until a repository counts as stale, 3 worst
// steps per repository.
const (
	DefaultWindow     = 7 * 24 * time.Hour
	DefaultStaleAfter = 14 * 24 * time.Hour
	DefaultWorstSteps = 3
)

// V1Report is the frozen v1 compliance wire document (schema_version 1).
//
// The shape is byte-faithful to BuildFlow's `compliance --json` output,
// which the overview dashboard consumes: field names, JSON keys, key order
// (struct declaration order under encoding/json/v2), and the nil-map `{}` /
// nil-slice `[]` normalization all mirror the producer. MarshalV1
// additionally canonicalizes map key order (deterministic JSON).
type V1Report struct {
	SchemaVersion int    `json:"schema_version"`
	GeneratedAt   string `json:"generated_at"`
	Since         string `json:"since"`
	StaleAfter    string `json:"stale_after"`
	// BuildFlowVersion is the producing system's version; the JSON key is
	// frozen by the v1 contract. Non-BuildFlow producers fill their own
	// version here.
	BuildFlowVersion string   `json:"buildflow_version"`
	Repos            []V1Repo `json:"repos"`
}

// V1Repo is one repository's row in a V1Report.
type V1Repo struct {
	Repo           string         `json:"repo"`
	State          string         `json:"state"`
	Runs           int64          `json:"runs"`
	StepRows       int64          `json:"step_rows"`
	Failures       int64          `json:"failures"`
	FailRate       float64        `json:"fail_rate"`
	LastRunAt      string         `json:"last_run_at"`
	LastRunVerdict string         `json:"last_run_verdict"`
	GateFindings   map[string]int `json:"gate_findings"`
	// BuildFlowVersions lists the producer versions seen in the window;
	// the JSON key is frozen by the v1 contract.
	BuildFlowVersions []string `json:"buildflow_versions"`
	WorstSteps        []V1Step `json:"worst_steps"`
}

// V1Step is one repo+step failure aggregate (the worst-steps feed).
type V1Step struct {
	Step      string  `json:"step"`
	Failures  int64   `json:"failures"`
	TotalRuns int64   `json:"total_runs"`
	FailRate  float64 `json:"fail_rate"`
}

// marshalOpts ensures deterministic map key ordering in all JSON output:
// encoding/json/v2 serializes map keys in unspecified order otherwise.
//
//nolint:gochecknoglobals // immutable encoding/json/v2 option set, built once
var marshalOpts = json.Deterministic(true)

// MarshalV1 encodes a report deterministically (sorted map keys, struct
// field order). BuildFlow's producer does not sort map keys; sorted keys
// are a canonical subset of valid v1 and decode identically everywhere.
func MarshalV1(report V1Report) ([]byte, error) {
	encoded, err := json.Marshal(report, marshalOpts)
	if err != nil {
		return nil, NewInternalError(err, "failed to encode compliance report")
	}

	return encoded, nil
}

// DecodeV1Report parses and validates a v1 wire document. A
// schema_version other than 1 is rejected (mirroring the overview
// consumer); repository states are carried as-is - validate them with
// ParseState if the consumer needs the enum verified.
func DecodeV1Report(data []byte) (V1Report, error) {
	var report V1Report

	if len(bytes.TrimSpace(data)) == 0 {
		return V1Report{}, NewParseError(nil, "empty compliance document")
	}

	if err := json.Unmarshal(data, &report); err != nil {
		return V1Report{}, NewParseError(err, "parse compliance document")
	}

	if report.SchemaVersion != SchemaVersionV1 {
		return V1Report{}, NewUnsupportedSchemaError(report.SchemaVersion, SchemaVersionV1)
	}

	return report, nil
}

// BuildV1Report derives states and shapes the wire payload. Pure: the
// golden test pins its exact schema. Repos are sorted worst-first
// (failing before stale before never-covered before compliant, then most
// failures, then name) - both the table (triage view) and the JSON
// (stable, consumer-keyed) carry this order. Nil maps and slices are
// normalized to {} and [] so every byte of the contract is explicit.
func BuildV1Report(
	evidence []RepoEvidence,
	since time.Time,
	staleAfter time.Duration,
	producerVersion string,
	now time.Time,
) V1Report {
	repos := make([]V1Repo, 0, len(evidence))

	for _, record := range evidence {
		row := V1Repo{
			Repo:              record.Repo.Get(),
			State:             DeriveState(record, now, staleAfter).String(),
			Runs:              record.Runs,
			StepRows:          record.StepRows,
			Failures:          record.Failures,
			FailRate:          record.FailRate,
			LastRunAt:         FormatWireTime(record.LastRunAt),
			LastRunVerdict:    record.LastRunVerdict.String(),
			GateFindings:      record.GateFindings,
			BuildFlowVersions: record.Versions,
			WorstSteps:        make([]V1Step, 0, len(record.WorstSteps)),
		}

		if row.GateFindings == nil {
			row.GateFindings = map[string]int{}
		}

		if row.BuildFlowVersions == nil {
			row.BuildFlowVersions = []string{}
		}

		for _, step := range record.WorstSteps {
			row.WorstSteps = append(row.WorstSteps, V1Step(step))
		}

		repos = append(repos, row)
	}

	if repos == nil {
		repos = []V1Repo{}
	}

	sortV1ReposWorstFirst(repos)

	return V1Report{
		SchemaVersion:    SchemaVersionV1,
		GeneratedAt:      now.UTC().Format(time.RFC3339Nano),
		Since:            since.UTC().Format(time.RFC3339Nano),
		StaleAfter:       staleAfter.String(),
		BuildFlowVersion: producerVersion,
		Repos:            repos,
	}
}

// sortV1ReposWorstFirst orders the report rows: the STATE RANK IS ALWAYS
// PRIMARY (the triage-view invariant), then most failures, then name.
func sortV1ReposWorstFirst(repos []V1Repo) {
	sort.SliceStable(repos, func(i, j int) bool {
		if rankI, rankJ := State(repos[i].State).Rank(), State(repos[j].State).Rank(); rankI != rankJ {
			return rankI < rankJ
		}

		if repos[i].Failures != repos[j].Failures {
			return repos[i].Failures > repos[j].Failures
		}

		return repos[i].Repo < repos[j].Repo
	})
}

// FormatWireTime renders a time as the wire timestamp format (RFC 3339,
// nanosecond precision, UTC); the zero time renders as the empty string.
func FormatWireTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.UTC().Format(time.RFC3339Nano)
}

// ParseWireTime parses a wire timestamp, accepting both RFC 3339
// nanosecond and second precision; empty and malformed values return the
// zero time.
func ParseWireTime(raw string) time.Time {
	if raw == "" {
		return time.Time{}
	}

	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return parsed
		}
	}

	return time.Time{}
}

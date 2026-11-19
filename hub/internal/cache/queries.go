package cache

// Window queries: the SQL half of the Store. Every query funnels through
// queryRows + wrapIterateErr (the single IO-error wrap home).

import (
	"context"
	"database/sql"
	"encoding/json/v2"

	"github.com/larsartmann/go-compliance"
)

// art-dupl:accept the rows lifecycle is database/sql idiom; a collector taking scan callbacks costs more ceremony than the shared lines
// queryWindowHealth aggregates the window-bounded counters per repo.
func (s *Store) queryWindowHealth(ctx context.Context, from string) ([]compliance.WindowHealth, error) {
	rows, err := s.queryRows(ctx, "query window health", `
SELECT repo,
       COALESCE(SUM(runs), 0)      AS runs,
       COALESCE(SUM(step_rows), 0) AS step_rows,
       COALESCE(SUM(failures), 0)  AS failures
FROM evidence_runs
WHERE recorded_at >= ?
GROUP BY repo
ORDER BY repo`, from)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var health []compliance.WindowHealth

	for rows.Next() {
		var row compliance.WindowHealth
		if err := rows.Scan(&row.Repo, &row.Runs, &row.StepRows, &row.Failures); err != nil {
			return nil, compliance.NewIOError(err, "scan window health")
		}

		if row.StepRows > 0 {
			row.FailRate = float64(row.Failures) / float64(row.StepRows)
		} else if row.Runs > 0 {
			row.FailRate = float64(row.Failures) / float64(row.Runs)
		}

		health = append(health, row)
	}

	return health, wrapIterateErr(rows.Err(), "iterate window health")
}

// latestRecords is the triple queryLatestRecords scans out of one row
// set: each repo's newest run plus that record's gate findings and
// severity counts.
type latestRecords struct {
	runs       []compliance.LatestRun
	findings   []compliance.RepoFindings
	severities []compliance.RepoSeverities
}

// queryLatestRecords returns each repo's all-history newest record in ONE
// query: the latest-run fields plus that record's gate findings and
// severity counts (the per-repo follow-up loop was an N+1; a correlated
// id-subquery selects the same rows the latest-runs query picked).
func (s *Store) queryLatestRecords(
	ctx context.Context,
) (latestRecords, error) {
	rows, err := s.queryRows(ctx, "query latest records", `
SELECT r.repo, r.recorded_at, r.verdict, r.producer_version, r.gate_findings, r.severity_counts
FROM evidence_runs r
WHERE r.id = (
    SELECT id FROM evidence_runs newer
    WHERE newer.repo = r.repo
    ORDER BY newer.recorded_at DESC, newer.id DESC
    LIMIT 1
)
ORDER BY r.repo`)
	if err != nil {
		return latestRecords{}, err
	}
	defer func() { _ = rows.Close() }()

	records := latestRecords{
		runs:       make([]compliance.LatestRun, 0, latestRecordsCapacity),
		findings:   make([]compliance.RepoFindings, 0, latestRecordsCapacity),
		severities: make([]compliance.RepoSeverities, 0, latestRecordsCapacity),
	}

	for rows.Next() {
		var repo, recordedAt, verdict, version, gateJSON, severityJSON string

		if err := rows.Scan(&repo, &recordedAt, &verdict, &version, &gateJSON, &severityJSON); err != nil {
			return latestRecords{}, compliance.NewIOError(err, "scan latest record")
		}

		name := compliance.NewRepoName(repo)

		records.runs = append(records.runs, compliance.LatestRun{
			Repo:       name,
			LastRunAt:  compliance.ParseWireTime(recordedAt),
			HasVerdict: true,
			AnyFailed:  verdict == compliance.VerdictFailed.String(),
			Version:    version,
		})

		gate, err := unmarshalCounts(gateJSON)
		if err != nil {
			return latestRecords{}, compliance.NewParseError(err, "decode gate findings for %s", repo)
		}

		sev, err := unmarshalCounts(severityJSON)
		if err != nil {
			return latestRecords{}, compliance.NewParseError(err, "decode severity counts for %s", repo)
		}

		records.findings = append(records.findings, compliance.RepoFindings{Repo: name, Findings: gate})
		records.severities = append(records.severities, compliance.RepoSeverities{Repo: name, Severities: sev})
	}

	if err := wrapIterateErr(rows.Err(), "iterate latest records"); err != nil {
		return latestRecords{}, err
	}

	return records, nil
}

// art-dupl:accept the rows lifecycle is database/sql idiom; a collector taking scan callbacks costs more ceremony than the shared lines
// queryWindowVersions returns each repo's distinct producer versions in
// the window.
func (s *Store) queryWindowVersions(ctx context.Context, from string) ([]compliance.RepoVersions, error) {
	rows, err := s.queryRows(ctx, "query window versions", `
SELECT repo, producer_version
FROM evidence_runs
WHERE recorded_at >= ? AND producer_version != ''
GROUP BY repo, producer_version
ORDER BY repo, producer_version`, from)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	byRepo := map[compliance.RepoName][]string{}

	for rows.Next() {
		var repo, version string
		if err := rows.Scan(&repo, &version); err != nil {
			return nil, compliance.NewIOError(err, "scan window version")
		}

		name := compliance.NewRepoName(repo)
		byRepo[name] = append(byRepo[name], version)
	}

	if err := wrapIterateErr(rows.Err(), "iterate window versions"); err != nil {
		return nil, err
	}

	out := make([]compliance.RepoVersions, 0, len(byRepo))
	for repo, versions := range byRepo {
		out = append(out, compliance.RepoVersions{Repo: repo, Versions: versions})
	}

	return out, nil
}

// marshalCounts encodes a count map column.
//
//nolint:wrapcheck // callers wrap with family constructors
func marshalCounts(counts map[string]int) (string, error) {
	if len(counts) == 0 {
		return "{}", nil
	}

	encoded, err := json.Marshal(counts)
	if err != nil {
		return "", err
	}

	return string(encoded), nil
}

// unmarshalCounts decodes a count map column; empty and null yield nil.
//
//nolint:nilnil,wrapcheck // (nil, nil) is the documented empty contract; callers wrap
func unmarshalCounts(raw string) (map[string]int, error) {
	if raw == "" || raw == "{}" {
		return nil, nil
	}

	var counts map[string]int
	if err := json.Unmarshal([]byte(raw), &counts); err != nil {
		return nil, err
	}

	if len(counts) == 0 {
		return nil, nil
	}

	return counts, nil
}

// queryRows issues a read query and family-types failures as IO errors.
// The caller owns scanning and closing the returned rows.
func (s *Store) queryRows(ctx context.Context, what, query string, args ...any) (*sql.Rows, error) {
	rows, err := s.db.DB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, compliance.NewIOError(err, "%s", what)
	}

	return rows, nil
}

// wrapIterateErr keeps rows.Err() propagation family-typed.
func wrapIterateErr(err error, what string) error {
	if err == nil {
		return nil
	}

	return compliance.NewIOError(err, "%s", what)
}

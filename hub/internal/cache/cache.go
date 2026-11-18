// Package cache is the hub's own evidence store: an append-only SQLite
// log of recorded runs (the reporter SDK's durable sink) plus the window
// queries that fold it back into RepoEvidence for the hub's report and
// UI. BuildFlow's evidence stays in BuildFlow; this cache is for every
// other producer.
package cache

import (
	"context"
	"database/sql"
	"encoding/json/v2"
	"sync"
	"time"

	"github.com/larsartmann/go-sqlitestore"

	"github.com/larsartmann/go-compliance"
)

// latestRecordsCapacity pre-sizes the per-query result slices (repos in a
// typical fleet; growth beyond it is fine, this only avoids the first
// reallocations).
const latestRecordsCapacity = 16

// migrations owns the cache schema. Append-only history: never modify,
// reorder, or remove applied migrations (go-sqlitestore membership
// semantics).
//
//nolint:gochecknoglobals // append-only schema history, intentionally package-level
var migrations = []sqlitestore.Migration{{
	Version:     1,
	Description: "evidence_runs: append-only per-repo run records",
	Up: func(ctx context.Context, db *sql.DB) error {
		_, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS evidence_runs (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    repo             TEXT    NOT NULL,
    recorded_at      TEXT    NOT NULL,
    verdict          TEXT    NOT NULL,
    runs             INTEGER NOT NULL DEFAULT 1,
    step_rows        INTEGER NOT NULL DEFAULT 0,
    failures         INTEGER NOT NULL DEFAULT 0,
    gate_findings    TEXT    NOT NULL DEFAULT '{}',
    severity_counts  TEXT    NOT NULL DEFAULT '{}',
    producer_version TEXT    NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_evidence_runs_repo_time ON evidence_runs (repo, recorded_at);
`)
		if err != nil {
			return compliance.NewIOError(err, "create evidence_runs table")
		}

		return nil
	},
}}

// Store is the evidence cache.
type Store struct {
	db *sqlitestore.DB

	onWriteMu sync.RWMutex
	onWrite   func(repo compliance.RepoName)
}

// Open opens (creating) the cache database at path and applies migrations.
func Open(ctx context.Context, path string) (*Store, error) {
	db, err := sqlitestore.Open(path)
	if err != nil {
		return nil, compliance.NewIOError(err, "open evidence cache %s", path)
	}

	if err := sqlitestore.RunMigrations(ctx, db.DB(), migrations); err != nil {
		_ = db.Close()

		return nil, compliance.NewIOError(err, "migrate evidence cache %s", path)
	}

	return &Store{db: db}, nil
}

// Close closes the underlying database.
func (s *Store) Close() error { return s.db.Close() }

// Name identifies the cache as a compliance Source.
func (s *Store) Name() compliance.SourceName { return compliance.NewSourceName("cache") }

// OnWrite registers a hook invoked synchronously on the writer's
// goroutine after every successful evidence append (the hub uses it to
// kick a debounced re-derive). Register once before serving; the hook
// must not call back into the Store.
func (s *Store) OnWrite(hook func(repo compliance.RepoName)) {
	s.onWriteMu.Lock()
	defer s.onWriteMu.Unlock()

	s.onWrite = hook
}

// WriteEvidence implements compliance.EvidenceWriter: one append-only run
// record. Queries derive the newest state, so out-of-order records cannot
// corrupt the view.
func (s *Store) WriteEvidence(ctx context.Context, repo compliance.RepoName, result compliance.RunResult) error {
	gateFindings, err := marshalCounts(result.GateFindings)
	if err != nil {
		return compliance.NewInternalError(err, "encode gate findings")
	}

	severities, err := marshalCounts(result.FindingSeverities)
	if err != nil {
		return compliance.NewInternalError(err, "encode severity counts")
	}

	_, err = s.db.ExecWrite(ctx, `
INSERT INTO evidence_runs
    (repo, recorded_at, verdict, runs, step_rows, failures, gate_findings, severity_counts, producer_version)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		repo.Get(), compliance.FormatWireTime(result.ResolvedAt()), result.Verdict.String(),
		result.ResolvedRuns(), result.StepRows, result.Failures,
		gateFindings, severities, result.ProducerVersion)
	if err != nil {
		return compliance.NewIOError(err, "record evidence for %s", repo.Get())
	}

	s.onWriteMu.RLock()
	hook := s.onWrite
	s.onWriteMu.RUnlock()

	if hook != nil {
		hook(repo)
	}

	return nil
}

// QueryEvidence implements compliance.Source, folding the cache into
// per-repo evidence for the [since, now) window. Repositories are never
// dropped for being quiet: the all-history latest run is included so a
// repository that stopped reporting still derives stale. Per-step failure
// aggregates are not recorded per step here, so worst-steps stay empty
// and stepFilter cannot scope (same contract as the frozen v1 document).
func (s *Store) QueryEvidence(
	ctx context.Context, since time.Time, worstStepLimit int, _ string,
) ([]compliance.RepoEvidence, error) {
	windowFrom := compliance.FormatWireTime(since)

	health, err := s.queryWindowHealth(ctx, windowFrom)
	if err != nil {
		return nil, err
	}

	records, err := s.queryLatestRecords(ctx)
	if err != nil {
		return nil, err
	}

	versions, err := s.queryWindowVersions(ctx, windowFrom)
	if err != nil {
		return nil, err
	}

	return compliance.FoldEvidence(compliance.FoldInput{
		WindowHealth:      health,
		LatestRuns:        records.runs,
		GateFindings:      records.findings,
		FindingSeverities: records.severities,
		Versions:          versions,
		WorstStepLimit:    worstStepLimit,
	}), nil
}

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

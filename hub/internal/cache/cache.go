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
	"time"

	"github.com/larsartmann/go-sqlitestore"

	"github.com/larsartmann/go-compliance"
)

// migrations owns the cache schema. Append-only history: never modify,
// reorder, or remove applied migrations (go-sqlitestore membership
// semantics).
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
			return err
		}

		return nil
	},
}}

// Store is the evidence cache.
type Store struct {
	db *sqlitestore.DB
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

	latest, err := s.queryLatestRuns(ctx)
	if err != nil {
		return nil, err
	}

	findings, severities, err := s.queryLatestRunCounts(ctx, latest)
	if err != nil {
		return nil, err
	}

	versions, err := s.queryWindowVersions(ctx, windowFrom)
	if err != nil {
		return nil, err
	}

	return compliance.FoldEvidence(compliance.FoldInput{
		WindowHealth:      health,
		LatestRuns:        latest,
		GateFindings:      findings,
		FindingSeverities: severities,
		Versions:          versions,
		WorstStepLimit:    worstStepLimit,
	}), nil
}

// queryWindowHealth aggregates the window-bounded counters per repo.
func (s *Store) queryWindowHealth(ctx context.Context, from string) ([]compliance.WindowHealth, error) {
	rows, err := s.db.DB().QueryContext(ctx, `
SELECT repo,
       COALESCE(SUM(runs), 0)      AS runs,
       COALESCE(SUM(step_rows), 0) AS step_rows,
       COALESCE(SUM(failures), 0)  AS failures
FROM evidence_runs
WHERE recorded_at >= ?
GROUP BY repo
ORDER BY repo`, from)
	if err != nil {
		return nil, compliance.NewIOError(err, "query window health")
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

	return health, rows.Err()
}

// queryLatestRuns returns each repo's all-history newest record.
func (s *Store) queryLatestRuns(ctx context.Context) ([]compliance.LatestRun, error) {
	rows, err := s.db.DB().QueryContext(ctx, `
SELECT repo, recorded_at, verdict, producer_version
FROM evidence_runs r
WHERE id = (
    SELECT id FROM evidence_runs newer
    WHERE newer.repo = r.repo
    ORDER BY newer.recorded_at DESC, newer.id DESC
    LIMIT 1
)
ORDER BY repo`)
	if err != nil {
		return nil, compliance.NewIOError(err, "query latest runs")
	}
	defer func() { _ = rows.Close() }()

	var runs []compliance.LatestRun

	for rows.Next() {
		var repo, recordedAt, verdict, version string

		if err := rows.Scan(&repo, &recordedAt, &verdict, &version); err != nil {
			return nil, compliance.NewIOError(err, "scan latest run")
		}

		runs = append(runs, compliance.LatestRun{
			Repo:       compliance.NewRepoName(repo),
			LastRunAt:  compliance.ParseWireTime(recordedAt),
			HasVerdict: true,
			AnyFailed:  verdict == compliance.VerdictFailed.String(),
			Version:    version,
		})
	}

	return runs, rows.Err()
}

// queryLatestRunCounts loads each repo's latest gate findings and severity
// counts.
func (s *Store) queryLatestRunCounts(
	ctx context.Context, latest []compliance.LatestRun,
) ([]compliance.RepoFindings, []compliance.RepoSeverities, error) {
	findings := make([]compliance.RepoFindings, 0, len(latest))
	severities := make([]compliance.RepoSeverities, 0, len(latest))

	for _, run := range latest {
		var gateJSON, severityJSON string

		err := s.db.DB().QueryRowContext(ctx, `
SELECT gate_findings, severity_counts
FROM evidence_runs
WHERE repo = ? AND recorded_at = ?
ORDER BY id DESC LIMIT 1`, run.Repo.Get(), compliance.FormatWireTime(run.LastRunAt)).
			Scan(&gateJSON, &severityJSON)
		if err != nil {
			return nil, nil, compliance.NewIOError(err, "query latest run counts for %s", run.Repo.Get())
		}

		gate, err := unmarshalCounts(gateJSON)
		if err != nil {
			return nil, nil, compliance.NewParseError(err, "decode gate findings for %s", run.Repo.Get())
		}

		sev, err := unmarshalCounts(severityJSON)
		if err != nil {
			return nil, nil, compliance.NewParseError(err, "decode severity counts for %s", run.Repo.Get())
		}

		findings = append(findings, compliance.RepoFindings{Repo: run.Repo, Findings: gate})
		severities = append(severities, compliance.RepoSeverities{Repo: run.Repo, Severities: sev})
	}

	return findings, severities, nil
}

// queryWindowVersions returns each repo's distinct producer versions in
// the window.
func (s *Store) queryWindowVersions(ctx context.Context, from string) ([]compliance.RepoVersions, error) {
	rows, err := s.db.DB().QueryContext(ctx, `
SELECT repo, producer_version
FROM evidence_runs
WHERE recorded_at >= ? AND producer_version != ''
GROUP BY repo, producer_version
ORDER BY repo, producer_version`, from)
	if err != nil {
		return nil, compliance.NewIOError(err, "query window versions")
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

	if err := rows.Err(); err != nil {
		return nil, compliance.NewIOError(err, "iterate window versions")
	}

	out := make([]compliance.RepoVersions, 0, len(byRepo))
	for repo, versions := range byRepo {
		out = append(out, compliance.RepoVersions{Repo: repo, Versions: versions})
	}

	return out, nil
}

// marshalCounts encodes a count map column.
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

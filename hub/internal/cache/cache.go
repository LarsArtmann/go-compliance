// Package cache is the hub's own evidence store: an append-only SQLite
// log of recorded runs (the reporter SDK's durable sink) plus the window
// queries that fold it back into RepoEvidence for the hub's report and
// UI. BuildFlow's evidence stays in BuildFlow; this cache is for every
// other producer.
package cache

import (
	"context"
	"database/sql"
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

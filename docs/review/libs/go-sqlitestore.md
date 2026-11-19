# go-sqlitestore

| Field            | Value                                                                                          |
| ---------------- | ---------------------------------------------------------------------------------------------- |
| Reviewed         | 2026-11-19 (T15 completion report — closes the audit's dep-graph coverage gap)                 |
| Group            | T15 (post-audit completion; sole-author verification, same contract as the 5-group fleet run)  |
| Consumer modules | hub — **direct** (`hub/go.mod`); core — absent (core keeps zero private deps, zero engines)    |
| Library version  | `v0.1.1` (pinned; equals latest tag — current)                                                 |
| Recent tags      | `v0.1.1`, `v0.1.0`                                                                             |
| Modules          | single module `github.com/larsartmann/go-sqlitestore` (engine core: Open/DB, DSN, WAL, write   |
|                  | mutex + SQLITE_BUSY retry, migrations)                                                          |

## What the consumer uses (verified)

- **`Open(path)` with pure defaults** (`cache.go:66`): parent-dir creation,
  the DSN with busy timeout, `PRAGMA journal_mode=WAL` + optimization
  PRAGMAs, and the auto-sized connection pool all arrive in one call. WAL
  is not cosmetic — it is why the go-filewatcher adoption works at all
  (cross-process commits are `-wal` sidecar writes that directory-watch
  fires on; a rollback-journal DB would be near-silent, see
  `go-filewatcher.md`).
- **`ExecWrite` for every write** (`cache.go:111`): the engine's Go-level
  write mutex + SQLITE_BUSY retry policy owns the hub's hardest
  concurrency — multiple reporter processes appending to one cache DB
  while the hub reads (the exact scenario `watcher_internal_test.go`
  exercises cross-process under `-race`).
- **`RunMigrations` + the `Migration` type** (`cache.go:29,71`): one
  explicit v1 migration; the membership comment documents the no-reorder
  contract.
- **`.DB()`** for reads (`queries.go`): reads go straight to `database/sql`
  under the WAL pool — the engine wisely does not wrap reads.

## What the consumer does not use (and why)

- `WithBusyTimeout` / `WithMaxOpenConns`: defaults (busy timeout, pool
  max(4, min(NumCPU, 8))) have never produced a busy failure in the fleet
  fixture or tests; knobs without an observed problem stay unwound.
- `WithWriteLock` / `WriteLock`: every hub write is a single statement —
  `ExecWrite` already holds the mutex across it; no multi-statement
  transaction exists yet.
- `BuildDSN` / `BuildDSNWithBusyTimeout` / `EnableWAL` /
  `ApplyOptimizations` / `OpenDB` / `New`: `Open` composes exactly these;
  calling the parts would re-implement the composition.
- `TableExists` / `ColumnExists` / `AddColumnIfNotExists` /
  `IsDuplicateColumnError`: additive-migration helpers for schemas that
  ship column adds; the cache has one table, one version.
- `TruncateSQL`, `SetDebugSQL`: no dynamic SQL truncation or statement
  debugging needs to date.

## Missed opportunities (recorded, not blocking)

- **`BusyRetryCount()` is free cross-process-write telemetry.** The
  `evidence-watch` health check (cli/health.go) reports watcher liveness;
  a sustained busy-retry climb would distinguish "healthy but contended"
  from "healthy and quiet". Route: add to the cache-disk check's detail
  line, one field, when cross-process contention is ever suspected.
- **`HealthCheck(ctx)`** (`engine.go:218`) could power the hub's
  `cache-disk` check instead of a hand-rolled ping — only worth doing when
  that check next changes.

## Verdict — **ADOPTED**

Utilization score **85** (4.25 of 5 weighted rows: the engine's whole
 raison-d'être — open-hardened SQLite + serialized writes + busy retry — is
 load-bearing in the hub's core loop; the unused surface is either composed
 away by `Open` or genuinely premature). This is what the
 "evidence store: engine core vs hand-rolled DSN" STACK row promised, held
 honest by the cross-process watcher test.

**Flip triggers:** a second table or a column add (the migration helpers
become load-bearing); multi-statement write transactions (WithWriteLock);
observed SQLITE_BUSY under fleet load (WithBusyTimeout + BusyRetryCount
telemetry); the cache-disk check's next rewrite (HealthCheck).

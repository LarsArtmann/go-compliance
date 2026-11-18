# Domain Language

> The ubiquitous language of this project. Every term used in code,
> conversations, and docs should be defined here unambiguously.

## Core terms

| Term             | Definition                                                                         | Where used in code                                           |
| ---------------- | ---------------------------------------------------------------------------------- | ------------------------------------------------------------ |
| Compliance State | A repo's posture: `compliant`, `failing`, `stale`, or `never-covered`              | `state.go`, `State`                                          |
| Verdict          | The pass/fail outcome of a single run                                              | `VerdictPassed` / `VerdictFailed`                            |
| Evidence         | Per-repo observations (runs, failures, findings, worst steps) used to derive state | `evidence.go`, `RepoEvidence`                                |
| Fold             | Unioning evidence rows per repo within a window, capped by worstStepLimit          | `fold.go`, `FoldEvidence`                                    |
| Window           | The evidence lookback period (`since`)                                             | `Source.QueryEvidence` ctx + since                           |
| Stale            | A repo whose newest evidence is older than `staleAfter`                            | `state.go` DeriveState                                       |
| Never-covered    | A repo the discovery daemon knows but which has no evidence in the window          | `hub/internal/discovery`                                     |
| Source           | An adapter producing evidence for a window (file, buildflow, cache)                | `source.go`, `Source`                                        |
| EvidenceWriter   | The port reporters write evidence through                                          | `reporter.go`, `EvidenceWriter`                              |
| Record           | The reporter one-liner appending one run's evidence                                | `reporter.go`, `Record`                                      |
| Wire contract v1 | The frozen JSON shape (`schema_version: 1`) byte-faithful to BuildFlow output      | `v1.go`, `MarshalV1` / `DecodeV1Report`                      |
| Hub              | The in-memory owner of current rows; fans out snapshot/update events               | `hub/internal/events`, `Hub`                                 |
| Row              | One repo's dashboard projection derived from evidence                              | `hub/internal/events/rows.go`, `RepoRow`                     |
| Row diff         | The added/changed/removed split between two row sets; drives patches               | `rows.go`, `DiffRows`                                        |
| Patch stream     | DataStar element/signals patches over SSE keeping the table live                   | `hub/internal/dashboard`                                     |
| Snapshot         | A full-fleet event; the first thing a new subscriber receives                      | `Hub.PublishSnapshot`                                        |
| Update           | An incremental event carrying the changed rows                                     | `Hub.PublishUpdate`                                          |
| Refresh          | Re-derivation against new evidence; diffs against the hub's current rows           | `Hub.PublishRefresh`                                         |
| Discovery        | The daemon telling the hub which repos exist (`/v1/discover`)                      | `hub/internal/discovery`, `Client`                           |
| Cache            | The hub's own SQLite evidence history (runs, findings, schema versions)            | `hub/internal/cache`, `Store`                                |
| Probe            | The health surface: liveness, readiness (evidence published), startup              | `hub/internal/cli/health.go`                                 |
| Bind posture     | Policy: localhost-only default, unix-socket mode, non-localhost warning            | `hub/internal/cli/bind.go`, `docs/decisions/bind-posture.md` |
| Fleet table      | The go-output terminal view of all repos (severity cells, worst-first)             | `hub/internal/cli/table.go`                                  |

## Bounded contexts

| Term     | Core module (meaning)                            | Hub module (meaning)                                  |
| -------- | ------------------------------------------------ | ----------------------------------------------------- |
| Evidence | A value object: aggregated per-repo observations | Live state: rows derived from a Source's evidence now |
| Report   | The v1 wire document (`BuildV1Report`)           | The `report` CLI command emitting that document       |
| Hub      | (unused)                                         | The event-fanout component owning current rows        |

## Deprecated terms

| Old term | Current term    | Why changed                                             |
| -------- | --------------- | ------------------------------------------------------- |
| overview | hub / dashboard | Extracted from BuildFlow's overview; new home, new name |

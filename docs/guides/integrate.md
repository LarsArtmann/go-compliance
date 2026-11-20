# Integrating a project in one line

Every LarsArtmann project can report fleet compliance with a single call.
This guide shows the three integration shapes, from zero-work (BuildFlow
projects) to the one-line SDK call (everything else).

## The state model

A repository lands in exactly one state:

| State           | Meaning                                                     |
| --------------- | ----------------------------------------------------------- |
| `compliant`     | Evidence in the window, no failures, last run fresh         |
| `failing`       | Any failed step row in the window (honest accounting)       |
| `stale`         | Evidence exists but the last run is older than `staleAfter` |
| `never-covered` | Discovered by the fleet but no evidence at all              |

The honesty rule matters: one failing step anywhere in the evidence window
keeps the repo `failing`, even when the latest run's verdict is `passed`.
A green dashboard means the whole window was green.

## Zero-work: BuildFlow projects

If a project already runs BuildFlow, its evidence is produced by
`buildflow compliance --json` (the daemon pipes it to the fleet document).
Nothing to do.

## One line: everything else

A runner's last step records the run:

```go
err := compliance.RecordTo(ctx, compliance.NewFileStore(path),
    compliance.NewRepoName("my-repo"),
    compliance.RunResult{
        Verdict:      compliance.VerdictPassed,
        Runs:         3,
        StepRows:     42,
        GateFindings: map[string]int{"lint": 2},
    })
```

Or with the default fleet document
(`/var/lib/overview/compliance.json`, overridable via
`GO_COMPLIANCE_EVIDENCE_FILE`):

```go
err := compliance.Record(ctx, compliance.NewRepoName("my-repo"),
    compliance.RunResult{Verdict: compliance.VerdictPassed})
```

Zero values are filled: `At` defaults to the record time, `Runs` to 1,
`FailRate` to `Failures/StepRows`, maps to their empty wire shapes.
Severities (`FindingSeverities`) are dropped by file-backed writers — the
frozen v1 wire contract has no field for them; the hub SQLite cache keeps
them.

Rows are replaced only when the new result is at least as new as the stored
one: parallel reporters cannot regress the document.

A runnable version of the above lives in
[`examples/reporter`](../../examples/reporter/main.go):

```bash
go run ./examples/reporter compliance-example.json
```

## Viewing

Feed the document to the hub for the terminal table or the live dashboard:

```bash
compliance-hub table --source file --file compliance.json
compliance-hub serve --source file --file compliance.json
```

## Health checks (serve)

`serve` exposes `/healthz` `/readyz` `/startupz` `/version` for systemd,
Kubernetes, or a scraping dashboard, backed by five named checks:

| Check               | Criticality | Meaning                                                        |
| ------------------- | ----------- | -------------------------------------------------------------- |
| `evidence`          | critical    | The readiness gate: 503 until the first evidence publish       |
| `cache-disk`        | optional    | 128 MiB free-space floor; only present with `--db`             |
| `refresh-freshness` | warn        | The re-derive loop stalled past 10 minutes (view may be stale) |
| `discovery`         | warn        | The fleet daemon connection stalled past 10 minutes            |
| `sse-backpressure`  | warn        | New dashboard patch drops (self-clearing, delta-based)         |

A warn degrades `/readyz` to 200-with-warnings, never 503: a stale view is
degraded truth, not service loss. Only a failed `evidence` check (or
startup) gates readiness. If you scrape for fleet monitoring, key on the
check names — they are contract-tested and stable for federation consumers.

## Rules of thumb

- Record from ONE place in the runner (the end), not per step — unless you
  want per-step evidence, in which case record once per step with that
  step's `Failures`/`StepRows` scope (step-scoped evidence flows through
  unchanged).
- Do not retry-record old runs on backfill; older results are ignored by
  design.
- Treat `failing` as "the window contains a failure", not "the build broke
  yesterday" — fix the finding or wait it out of the window.

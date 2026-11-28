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

Or with the default fleet document (see [Where the evidence
lives](#where-the-evidence-lives) for its location):

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

## Where the evidence lives

One fleet document holds every repository's rows. Where it sits:

| Deployment             | Fleet document                             |
| ---------------------- | ------------------------------------------ |
| System (the default)   | `/var/lib/overview/compliance.json`        |
| User (`systemctl --user`) | `~/.local/state/overview/compliance.json` |
| Override               | `GO_COMPLIANCE_EVIDENCE_FILE`              |

The system path is the core's `DefaultEvidenceFile` constant
(`reporter.go`) and the `Record` helper's fallback; user deployments pass
the XDG state file explicitly, which is what both deploy scripts
(`scripts/deploy-hub.sh`, `scripts/deploy-fleet-evidence.sh`) do.
`GO_COMPLIANCE_EVIDENCE_FILE` overrides the `Record` default for every
caller of the SDK; the commands that take a path directly override it per
invocation (`--file` on the hub, `--evidence-file` on `library-policy`).

## Who writes the evidence

A producer is any process that records rows into the fleet document, and
there is more than one:

- Every `library-policy` run records its own repository's evidence unless
  `--no-record` is passed: a plain scan records each git repository it
  reaches, and `fleet` records every repository under its root.
- The daily producer timer (`fleet-evidence.timer`, 05:10, installed by
  `scripts/deploy-fleet-evidence.sh`) runs `library-policy fleet` from a
  nix build of library-policy's committed tree, so its rows carry the
  commit as the producer version.

A hand-built `library-policy` binary carries no stamp and therefore writes
a `dev` version; the next timed scan records the same repository again,
newer, and replaces the row. Mixed version stamps in the fleet document
are expected, not a defect.

Rows are replaced only when the new result is at least as new as the
stored one (the core's write rule), so parallel writers cannot regress the
document.

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

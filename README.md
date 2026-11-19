# go-compliance

Fleet compliance evidence for every LarsArtmann project: one state model
(`compliant` / `failing` / `stale` / `never-covered`), one frozen wire
contract, a realtime WebUI, and adapters instead of per-project plumbing.

> **Status: built, pre-release.** Core module and hub are built and tested: state
> model, evidence fold, frozen v1 contract, reporter SDK, the hub service
> with realtime WebUI, SQLite cache, and discovery adapter. See
> [FEATURES.md](FEATURES.md) and [STACK.md](STACK.md).

![Fleet compliance table](docs/assets/terminal-table.svg)

![Realtime fleet dashboard](docs/assets/dashboard.png)

## Architecture

- **Core module** (`github.com/larsartmann/go-compliance`) — domain types,
  state derivation over evidence windows, the frozen v1 contract, a `Source`
  adapter interface, and a one-liner reporter SDK. Two dependencies
  (go-branded-id, go-error-family).
- **Hub module** (`github.com/larsartmann/go-compliance/hub`) — HTTP service
  with a realtime WebUI (Datastar over SSE), the CLI (`compliance
  serve|report|table`), a SQLite evidence cache, and adapters: BuildFlow
  evidence (v1 JSON), project-discovery-daemon (fleet list → never-covered).

Integration goal: BuildFlow-covered projects are compliant by default (zero
changes); everything else reports with one CLI line or one SDK call.

## Embedding the dashboard

Any Go service can mount the realtime fleet dashboard in one call:

```go
hub := events.NewHub()                       // publish snapshots/refreshes here
board := dashboard.New(hub)

mux := http.NewServeMux()
board.RegisterRoutes(mux)                    // / (page), /events (patches), /datastar.js
```

Feed it evidence and everything downstream updates live:

```go
rows := events.RowsFromEvidence(evidence, time.Now(), staleAfter)
published, err := hub.PublishRefresh(rows)   // snapshot first, diffs after
```

`RegisterRoutesAt(dashboard.Routes{...}, mux)` relocates the surface (e.g.
behind `/compliance/`); the page route serves JSON to `Accept:
application/json` clients.

Or run the packaged service: `compliance-hub serve --source file --file
compliance.json` (also `--source buildflow`, `--source cache --db path`).

### Health surface (packaged service)

`serve` exposes the standard probe routes for systemd, Kubernetes, or a
scraping dashboard:

| Route       | Meaning                                                                    |
| ----------- | -------------------------------------------------------------------------- |
| `/healthz`  | Liveness, always passes (an evidence failure must not restart the process) |
| `/readyz`   | 503 until evidence is published; degraded-but-serving stays 200            |
| `/startupz` | Startup latch, passes once the critical `evidence` check first passes      |
| `/version`  | Build version stamp                                                        |

Checks: `evidence` (critical gate), `cache-disk` (128 MiB floor, only with
`--db`), `refresh-freshness` and `discovery` (warn when their loop stalls
past 10 minutes), `sse-backpressure` (warn on new patch drops). Shutdown
drains honestly: readiness flips to 503 before the listeners close.

### Exit codes

Failures exit with [go-error-family](https://github.com/LarsArtmann/go-error-family)
BSD sysexits codes, not generic 1 — `systemd` restart policy and shell
muscle memory can rely on them:

| Code | Meaning                                                              |
| ---- | -------------------------------------------------------------------- |
| 0    | Success                                                             |
| 1    | Rejection: invalid flags, bad window, malformed input — do not retry |
| 69   | Infrastructure: internal wiring failure                             |
| 75   | Transient: cache/IO unreachable — retry (e.g. `Restart=on-failure`)  |

Verify live: `compliance-hub report --db /nonexistent/x.db; echo $?` → `75`.

## CI

[`.github/workflows/ci.yml`](.github/workflows/ci.yml) runs per-module
build+test and lint (golangci-lint pinned to the devShell version), dprint,
actionlint, stack-drift-check, docs-gates, and `nix flake check`.

The hub module and the flake build need one private dependency
(`github.com/LarsArtmann/go-daemon`). CI reads it via a repository secret:

- **`FLAKE_PAT`** — fine-grained PAT with read-only **Contents** on
  `LarsArtmann/go-daemon`. Without it, the hub/flake jobs skip (not fail);
  an owner must create the secret on each runner host (GitHub **and**
  Forgejo) to activate them.

## Documentation

- [Integrating a project in one line](docs/guides/integrate.md)
- [Library utilization reports](docs/review/libs/index.html) — 20 audited
  dependencies, verdicts and flip triggers (citations mechanically
  verified by `scripts/check-citations.sh`)
- [STACK.md](STACK.md) — foundation decision record (enforced by
  stack-drift-check)

# STACK — go-compliance

> The deliberate foundation of this project: every significant dependency was
> chosen, not accumulated. Generated 2026-11-17 by the stack-selection skill
> (profile: multi-module Go library + realtime fleet-compliance hub with WebUI).
> This is a LIVING document — update the row when the reason changes, delete
> the row when the dependency goes.
> Enforced by `stack-selection/scripts/stack-drift-check.sh <project-dir>`:
> every direct github.com/larsartmann/* require must have a Chosen foundation
> row. Exit 0 = consistent; exit 1 = undeclared or contradicted dep; exit 2 =
> this file is missing.

## Profile

| Axis         | Answer                                                                                                                                                                                                                         |
| ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Shape        | Two modules: dep-light core library + hub (HTTP service with realtime WebUI + cmdguard CLI)                                                                                                                                    |
| Domain       | Fleet compliance evidence: per-project gate/health states over time windows (compliant / failing / stale / never-covered)                                                                                                      |
| Persistence  | SQLite (hub evidence cache; BuildFlow state DB stays source of truth, consumed via adapter)                                                                                                                                    |
| Interfaces   | terminal tables, HTTP + SSE WebUI (Datastar), JSON API (frozen v1 overview-compat + hub v2), unix-socket HTTP (discovery)                                                                                                      |
| Scale target | Single operator, fleet of ~280 local repos, one single-binary hub service                                                                                                                                                      |
| Constraints  | nix flake build, Go 1.27 floor, systemd via sd_notify; hosting: Forgejo (evo-x2) primary + GitHub mirror, both public. Core module is public-clean; the hub builds against one PRIVATE dep (go-daemon) — see errata 2026-11-18 |

## Module layout (mirrors the go-health / go-health-dashboard split rationale)

1. `github.com/larsartmann/go-compliance` — core: branded domain types, state
   model (generalized `DeriveComplianceState` ported from BuildFlow), evidence
   aggregation, frozen v1 wire contract, `Source` adapter interface, one-liner
   reporter SDK. Deps: go-branded-id + go-error-family only.
2. `github.com/larsartmann/go-compliance/hub` — server + WebUI + realtime +
   CLI (`cmd/`): carries templ, templ-components, go-datastar, go-sse,
   go-sqlitestore, cmdguard, go-output. Integration surface for embedders:
   one `RegisterRoutes(mux)` call (go-health-dashboard pattern).

Rationale: integrators that only report evidence or consume JSON pay zero UI
dependency cost; the UI-heavy stack lives where it is used.

## Chosen foundation

| Category             | Choice (first-party = ★)                                               | Why this, not that                                                                                                                                                                                                                          | Owning skill / docs                                             | Revisit when                                                                               |
| -------------------- | ---------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------- | ------------------------------------------------------------------------------------------ |
| Domain IDs & types   | ★ go-branded-id                                                        | branded RepoName/SourceName/etc. vs primitive strings (how-to-golang policy; 114 consumers)                                                                                                                                                 | how-to-golang                                                   | —                                                                                          |
| Error protocol       | ★ go-error-family                                                      | fleet classification/exit-code standard (135 consumers; BuildFlow precedent) vs bare stdlib wraps                                                                                                                                           | README + pkg.go.dev                                             | —                                                                                          |
| Evidence store       | ★ go-sqlitestore                                                       | fleet SQLite engine core (DSN+PRAGMAs, WAL, busy-retry; 23 consumers) vs hand-rolled modernc DSN                                                                                                                                            | README                                                          | Postgres-scale needs                                                                       |
| Realtime transport   | ★ go-sse                                                               | SSE wire format + fan-out + reconnection replay (59 consumers) vs go-aichat ssehub (correct primitive, AI-chat SDK drag)                                                                                                                    | README                                                          | —                                                                                          |
| WebUI patch protocol | ★ go-datastar (+ /static)                                              | patches-as-values producing `sse.Event`, broadcaster + Last-Event-ID replay (26 consumers) vs starfederation/datastar-go (connection-bound)                                                                                                 | README                                                          | —                                                                                          |
| WebUI components     | ★ templ-components (+datastar/icons/utils)                             | fleet server-rendered component library (40 consumers) vs hand-rolled HTML/CSS                                                                                                                                                              | `templ-components` skill                                        | —                                                                                          |
| CLI framework        | ★ cmdguard/v5                                                          | typed struct-tag per-command flags + DI + config files (36 consumers) vs raw cobra + package-level flag vars (BuildFlow compliance's current wart)                                                                                          | `cmdguard` skill                                                | —                                                                                          |
| Terminal output      | ★ go-output (+ /table, /delimited, /markdown, /markup, /serialization) | ANSI-display-width-aware tables (BuildFlow gotcha #298) + JSON, CSV, HTML and the hub's other `--format` renderers (51 consumers) vs fmt Printf width verbs                                                                                 | README + pkg.go.dev                                             | output needs exceed its formats                                                            |
| DI container         | samber/do v2                                                           | cmdguard's container; go-health built for it vs manual wiring                                                                                                                                                                               | `samber-do-best-practices` skill                                | —                                                                                          |
| Fleet discovery      | project-discovery-daemon via plain HTTP                                | zero-config fleet list + realtime watch (`POST /v1/discover`, SSE `GET /v1/watch`) over its unix socket vs importing project-discovery-sdk (private module → would force this repo private; the daemon exists to amortize the SDK pipeline) | daemon README                                                   | typed SDK events wanted → private submodule behind the Source interface                    |
| Hub runtime          | ★ go-daemon                                                            | unix-socket + sd_notify/systemd building blocks (project-discovery-daemon precedent) vs bare main.go                                                                                                                                        | README                                                          | —                                                                                          |
| Evidence liveness    | ★ go-filewatcher/v2                                                    | dir-watch of the cache DB + WAL sidecars closes the 5-min cross-process staleness window to ~1-2s vs fsnotify hand-roll or polling (report `docs/review/libs/go-filewatcher.md`: NOT-ADOPTED-but-should, 0/6)                               | `docs/review/libs/go-filewatcher.md` + AGENTS.md                | —                                                                                          |
| Hub's own health     | ★ go-health                                                            | the hub is a service; fleet probe standard; federation precedent (22 consumers)                                                                                                                                                             | README                                                          | —                                                                                          |
| Logging              | slog (stdlib)                                                          | how-to-golang policy (logrus/zerolog banned); the hub needs nothing beyond slog's text and JSON handlers (no logging library is required)                                                                                                   | how-to-golang                                                   | —                                                                                          |
| Terminal color       | lipgloss/v2 + colorprofile (Charm v2)                                  | styles the fleet-table cells and detects the terminal's color profile under `--color auto` (one authority decides colored or plain, so chrome and cells never disagree on a pipe) vs raw ANSI escape strings                                | how-to-golang (required libraries, TUI/Styling)                 | —                                                                                          |
| Human-readable units | go-humanize                                                            | plural words in the summary and the findings cells (`english.PluralWord`) and relative times in the health and table output (`humanize.Time`) vs hand-rolled plurals and duration formatting                                                | how-to-golang; owner-decisions digest row "go-humanize adopt"   | the three `humanize.Time` call sites each carry a standing library-policy finding (digest) |
| JSON                 | encoding/json/v2 (stdlib)                                              | GA in Go 1.27; all third-party JSON libs banned (policy)                                                                                                                                                                                    | how-to-golang                                                   | —                                                                                          |
| Testing              | ginkgo/v2 + gomega                                                     | policy: testify banned; BDD specs for state model + contract goldens                                                                                                                                                                        | `bdd-testing` skill                                             | —                                                                                          |
| Wire contract        | BuildFlow compliance v1 (frozen), ported                               | overview rejects `schema_version != 1` today; byte-compat golden tests on both sides vs inventing a v2 first                                                                                                                                | BuildFlow AGENTS.md + golden tests                              | v2 lives at the hub API layer only, v1 stays emittable forever                             |
| Nix packaging        | go-nix-helpers (flake-parts)                                           | fleet packaging standard                                                                                                                                                                                                                    | `nix-private-go-repos` skill (only if a private dep ever lands) | —                                                                                          |

Rules for this table:

- One row per concern actually needed by the profile — not per category in
  the capability map. Absent concern = absent row.
- "Why this, not that" must name the rejected alternative. A choice without
  a rejected alternative was not a decision.
- First-party libs are marked ★ and MUST have been verified to exist
  (README + go.mod + tagged version) before this table was written.

## Deliberately excluded

| Candidate                                         | Reason                                                                                                                                                                                       |
| ------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| go-cqrs-lite (104 modules) / cqrs-htmx (userless) | evidence ingest is adapter writes with no command-side invariants; window views are plain SQL; realtime is SSE fan-out; hub is single-operator/userless (YAGNI). Revisit: audit-grade replay |
| go-composable-business-types (96)                 | DataPoint[T] bitemporal audit trail overkill: window aggregates + last-run evidence, not point-in-time business facts                                                                        |
| go-aichat (ssehub)                                | right primitive (SSE fan-out with replay), wrong module (AI-chat SDK); go-sse owns the transport directly                                                                                    |
| project-discovery-sdk (87)                        | private module + expensive pipeline; the daemon already amortizes it and speaks HTTP; keeps this repo public-capable                                                                         |
| go-health-dashboard (17)                          | same architecture, different domain (live service probes vs fleet evidence windows); mirror the pattern, do not import it                                                                    |
| go-appkit (63)                                    | few endpoints; `net/http` mux + `RegisterRoutes` is less ceremony. Revisit: middleware/CORS needs grow                                                                                       |
| universal-workflow (92)                           | no DAG orchestration in the hub; adapters poll/subscribe                                                                                                                                     |
| overview                                          | stays the downstream consumer; go-compliance emits the exact contract overview already parses                                                                                                |

Every high-importance candidate the profile plausibly touched belongs here
with a reason — silence about a known candidate is an unmade decision, not a
decision.

## Wiring order

1. **root**: `go mod init github.com/larsartmann/go-compliance`; `go get
   github.com/larsartmann/go-branded-id@latest
   github.com/larsartmann/go-error-family@latest`. Port from BuildFlow:
   `DeriveComplianceState` → generalized state model (add `never-covered`),
   evidence fold, v1 wire types. Golden tests must pin byte-compat against
   real `buildflow compliance --json` output AND overview's decoder
   expectations (schema_version == 1). `Source` adapter interface +
   one-liner reporter SDK + BuildFlow adapter.
2. **hub**: requires the released sibling github.com/larsartmann/go-compliance
   (a tagged version from the module proxy; `go.work` resolves the local
   tree for development); cmdguard CLI (`compliance serve|report|table`) →
   go-output tables; go-sqlitestore evidence cache; discovery daemon HTTP
   client (fleet list → never-covered); realtime go-sse `Broadcaster` +
   go-datastar patches; templ-components WebUI mirroring
   go-health-dashboard (one `RegisterRoutes(mux)`, severity-grouped repo
   tables, filter chips, live badges); go-health self-probe; go-daemon
   sd_notify (SSE framing + probe-first socket checks) + go-filewatcher
   (cross-process evidence dir-watch); go-error-family exit-code contract
   in main; systemd unit.
3. flake.nix via go-nix-helpers; `buildflow` quality gate once covered.
4. Re-run `stack-drift-check.sh` after every new larsartmann/* require.

## Decision log

- 2026-11-17 — Initial selection (stack-selection skill; born from the
  BuildFlow `compliance` extraction evaluation — the v1 contract, state model
  and evidence fold port over; the BuildFlow-schema SQL stays in BuildFlow
  behind the adapter).
- 2026-11-17 — Hosting decided (owner): Forgejo (`forgejo@evo-x2.home.lan:lars/go-compliance.git`)
  is primary (`origin`), GitHub (`git@github.com:LarsArtmann/go-compliance.git`)
  is the mirror remote (`github`), both public — the fleet convention observed
  in quotaboard/pim-sync. Answers status-report question g3.
- 2026-11-18 — **Errata: templ generator is version-pinned.** The flake builds
  `templ-gen` from the `a-h/templ` tag matching the hub go.mod tool directive
  (v0.3.1100); nixpkgs' `templ` lags behind and its older parser can mangle
  generated output (templ-components precedent). Never swap the pin for
  `ps.templ` without checking parity with the tool directive.
- 2026-11-18 — **Errata: the "no private deps" constraint is superseded for
  the hub.** `hub/internal/cli/bind.go` imports `github.com/LarsArtmann/go-daemon`
  (private): unix-socket + sd_notify building blocks. Consequence: the CORE
  module stays proxy-fetchable by anyone; building the HUB requires a PAT
  with read on go-daemon (CI gates those jobs on `FLAKE_PAT`; local dev
  resolves it via nix.conf/git insteadOf). The discovery daemon is still
  spoken to over plain HTTP as decided.
  Escape hatch if the hub must become publicly buildable: publish go-daemon,
  or extract the sd_notify surface behind a tiny public shim.
- 2026-11-19 — **go-daemon surface widened, deliberately minimal** (post-audit
  T1+T2): `discovery/client.go` now uses `daemon.ParseSSEData` (WHATWG framing)
  and `bind.go` uses `daemon.IsSocketAlive` for probe-first stale-socket cleanup
  (a live sibling's socket is refused, never stolen). Full `Server.Start`
  adoption was considered and declined: serve.go's hand-rolled lifecycle keeps
  the exact sd_notify/watchdog ordering pinned by the M27 tests and the minimal
  `http.Server` posture (docs/decisions verdict 2026-11-18); revisit only if a
  lifecycle ordering bug class reopens (flip trigger in
  docs/review/libs/go-daemon.md).
- 2026-11-19 — **go-filewatcher/v2 v2.5.0 adopted** (post-audit T3, plan
  `docs/planning/2026-11-19_07-55_*`): the utilization audit's one
  NOT-ADOPTED-but-should verdict. `hub/internal/cli/watcher.go` dir-watches
  the cache DB plus its `-wal`/`-shm` siblings (and the evidence file's
  directory for `--source file`, catching atomic renames) and kicks the
  refresher — cross-process SQLite writes go from a 5-minute safety-net
  wait to ~1-2s, keeping the realtime promise honest. Public module (no
  FLAKE_PAT impact). The 5-min safety net stays as the backstop; the
  watcher also feeds a non-critical `evidence-watch` health check
  (`watcher.Healthy`, warn-not-fail like discovery). Pinned by
  `watcher_internal_test.go` (cross-process write → refresh ≤ 2s).
- 2026-11-19 — **Health-surface verdicts g1/g2/g3 + boot contract recorded**
  (asked by the 20-02 samber-do review, settled same day in the integrity-first
  execution pass; previously lived only in AGENTS.md): g1 — refresh
  staleness degrades `/readyz` to **warn, stays 200**, threshold 10 min
  (2× the safety net; a stale view is degraded truth, not service loss);
  g2 — an **empty published fleet IS ready** (nil rows before the first
  publish vs non-nil empty fleet is the load-bearing `Hub.Rows()` contract);
  g3 — lint posture is **real fixes + nolint-with-reason, NO baseline file**
  (both modules hold 0 findings). Boot contract: `serve` **exits** when the
  boot refresh fails; systemd `Restart=on-failure` owns the retry —
  tolerant-boot declined (a broken evidence source pages the operator, it
  does not boot a green 503 process).
- 2026-11-20 — **The hub consumes the released core, not a `replace`.** The
  `replace github.com/larsartmann/go-compliance => ../` directive was
  dropped at the v0.2.0 tag: `hub/go.mod` requires the tagged version, so a
  consumer of the tag and the hub build against the same code. Local
  development stays on the working tree through `go.work` (`use .` plus
  `use ./hub`; verified 2026-11-28: with the workspace on, `go list -m`
  resolves the core to this directory). Every per-module command therefore
  runs with `GOWORK=off`, which is what CI and a consumer see. (Entry
  recorded 2026-11-28; it was asked for by the v0.2.0 release run.)
- 2026-11-28 — **Three direct hub dependencies had no row since the first
  commit:** `go-humanize`, `charm.land/lipgloss/v2`, and
  `github.com/charmbracelet/colorprofile`. Rows added above. The Logging
  row named `charm.land/log/v2`, which the hub never required (it logs with
  stdlib `slog`); corrected. `stack-drift-check.sh` only audits
  first-party requires, so third-party drift is caught by reading, not by
  the gate.

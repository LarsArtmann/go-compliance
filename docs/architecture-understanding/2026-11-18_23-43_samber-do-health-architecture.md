# samber/do v2 + Health Checks: Architecture Deep Review, Post-Fix Re-Verification

**Date:** 2026-11-18 23:43
**Scope:** `hub` module DI posture, health-check architecture, service orientation, composability, resilience, self-health. Core module reviewed only where it defines the ports the hub implements.
**Method:** every wiring/liveness file re-read end-to-end against the current tree (`serve.go`, `bind.go`, `app.go`, `flags.go`, `health.go`, `refresher.go`, `events.go`, `routes.go`), `go mod why`/`go mod graph` re-run, go-health v0.10.0 semantics re-verified from library source (`probe.go`, `lifecycle.go`, `options.go`, `checks.go`), the full gate battery re-executed with redirect-captured exit codes, and every finding of the morning review (`2026-11-18_19-39_samber-do-health-architecture.md`) tracked to its resolution.
**Positioning:** this is the re-verification pass. The 19:39 review found two defects (both fixed same session) and ranked four open gaps. The same-day execution pass (WP-01..WP-16, see `docs/status/2026-11-18_23-09_execution-pass-19wp-status.md` and `docs/status/2026-11-18_20-02_samber-do-health-review-status.md`) claims to have closed them all. Every claim was re-verified against code and tests today; nothing below is taken from a status report on faith.

---

## 1. Verdict

> **Q1: "Are we using samber/do v2 in combination with Health Checks superbly?!"**
>
> **We use samber/do zero times, on purpose, and that remains the superb answer.**
> `go mod why github.com/samber/do/v2` resolves only through `cmdguard/v5` (re-run today);
> zero `.go` files in either module import it. samber/do v2.1.0 exists in the graph as a
> transitive dependency of cmdguard's internal CLI container and of go-health's injector
> path, and neither our code nor our runtime needs it. The "combination" the question
> imagines (health checks aggregated from a do container) is go-health's _other_ half:
> `health.New(do.Injector, ...)`. The hub deliberately uses the injector-free half,
> `health.NewChecks` (checks.go:44 in go-health), and uses it superbly: six checks with
> correct criticality classification, panic-safe transition logging, a 1s background
> cache that makes scraping safe, and the library's own two-phase drain recipe. The
> recorded decision (STACK.md:50, "vs manual wiring") still holds; its revisit triggers
> (§6) still do not fire.

> **Q2: "Do we have a PROPER Service Oriented, composable, resilient, and self-health architecture?"**
>
> **Yes. Scorecard average 4.6/5 (Excellent band), and the self-health property is now
> complete: every loop the hub owns is observable by the health surface.** The one real
> hole the 19:39 review found (a hung evidence query wedging the refresh loop while all
> checks stayed green) is closed, pinned by a hanging-source test, and re-verified. The
> four previously-invisible failure modes (stale refresh, dead discovery daemon, SSE
> drops, check flapping) now surface as warn checks and log lines. Residual gaps are
> minor and listed in §7; the one actionable gap found in this pass (health surface
> undocumented in README) was fixed on sight.

---

## 2. The samber/do reality (re-verified)

### 2.1 Dependency graph (re-run today)

```
hub ──> cmdguard/v5 v5.1.0 ────────> samber/do v2.1.0        (container INSIDE cmdguard)
hub ──> cmdguard/v5 v5.1.0 ────────> samber-do-auditlog      (cmdguard's observability)
hub ──> go-health v0.10.0 ─────────> samber/do v2.1.0        (go-health's injector path, unused)
hub ──> (direct calls to do.*: NONE - grep + go mod why, re-run today)
```

`go mod why github.com/samber/do/v2` → `hub/cmd/compliance-hub → cmdguard/v5 → samber/do/v2`.
The only samber/do container that ever executes is the one cmdguard builds for the CLI
flag layer itself; the hub's six long-lived services never enter a container.

### 2.2 Why this stays right

The service graph is unchanged since the morning review: six long-lived components wired
in exactly one function (`runServe`, hub/internal/cli/serve.go:29-134), each with an
explicit lifetime owner (source closer serve.go:46, refresher ctx, discovery watch ctx,
probe Shutdown serve.go:123, hub drain serve.go:126, HTTP drain bind.go:118-121).
A container would add named-service indirection and lazy-singleton semantics to a graph
that fits on one screen, and would introduce DO-6's cross-service shutdown hazard as a
NEW risk. The revisit triggers from the morning review all remain unfired: service
count still ~6, still one composition site, Source still has 3 flag-selected flavors
that never coexist, no request/tenant scopes, no DI-observability requirement.

---

## 3. DO-1 → DO-6 equivalence audit (re-verified against the current tree)

Each samber/do anti-pattern rule, checked against its intent in manual-wiring code:

| Rule | Smell                                        | Verdict      | Evidence (current tree)                                                                                                                                                                                                                                     |
| ---- | -------------------------------------------- | ------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| DO-1 | `Must*` panic in runtime paths               | ✅ N/A-clean | No `do.*` calls exist. All constructor and flag errors propagate as errors (serve.go:31-68, flags.go:41-66, app.go:27-64). The only panic surface is cmdguard-internal, which owns its own registration-time panics by design.                              |
| DO-2 | Container without `Shutdown`                 | ✅ N/A-clean | No container. Every resource has a teardown: source closer (serve.go:46), HTTP drain (bind.go:118-121), probe two-phase drain (serve.go:122-123), hub SSE drain (serve.go:126).                                                                             |
| DO-3 | `Override*` in production                    | ✅ N/A-clean | No overrides anywhere. Test seams are injectable writers (app.go:27) and interface fakes; cmdguard flags are declarative struct tags, not overrides.                                                                                                        |
| DO-4 | Global injector/service locator              | ✅ Clean     | No package-level resolver. Dependencies are struct fields (refresher.go:49-81 holds peers; events.Hub holds the broadcaster). The only globals left are lint-suppressed constants/ldflags targets (Version, app.go:19).                                     |
| DO-5 | Resolution inside loops                      | ✅ Clean     | Nothing is resolved per iteration; the trigger loop (refresher.go:178-211) and discovery loop (refresher.go:218-257) only touch already-built values.                                                                                                       |
| DO-6 | `Shutdown` reaching into other services      | ✅ Clean     | `Hub.Shutdown` drains only the broadcaster (events.go:282); `Store.Close` closes only its DB; `probe.Shutdown` owns only the probe loop (one-way, ADR-009 in go-health). Cross-service ORDER lives in `runServe`, the single lifetime owner, as prescribed. |
| 4.7  | Service-locator smell (injector passed deep) | ✅ Clean     | Nothing receives a resolver; mux closures capture already-built values (serve.go:81-102).                                                                                                                                                                   |
| 4.8  | Service storing the container                | ✅ Clean     | Services hold peers, never a locator.                                                                                                                                                                                                                       |

**Structural conclusion (unchanged, now twice-verified):** the codebase achieves every
rule's intent by construction rather than by lint. There is nothing for a do-analyzer to
find, because there is no do.

---

## 4. Health architecture deep-dive: the self-health property is now complete

### 4.1 The six checks (health.go:73-134)

| Check                                  | Criticality                    | Semantics                                                                                                                                                  | Pin (test)                                                         |
| -------------------------------------- | ------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------ |
| `evidence`                             | critical (gates ready+startup) | Fails with `ErrEvidenceNotPublished` while `hub.Rows() == nil`; empty published fleet IS ready (verdict g2)                                                | health_internal_test.go:211; events_test.go nil/non-nil/empty spec |
| `refresh-freshness`                    | warn, stays 200 (verdict g1)   | `ErrRefreshNeverSucceeded` before first success; `ErrRefreshStale` past 10 min (= 2 missed safety-net ticks)                                               | health_internal_test.go:234-263                                    |
| `discovery`                            | warn, stays 200                | `ErrDiscoveryNeverContacted` before first contact; `ErrDiscoverySilent` past 10 min; stamp = watch connect or snapshot success, never regresses on failure | health_internal_test.go:183-208; refresher_internal_test.go:133    |
| `sse-backpressure`                     | warn, delta-based              | Warns only on drops NEW since last evaluation (self-clearing; the cumulative counter would stand-warn forever), counts in error text                       | health_internal_test.go:150-180 (flood-and-clear spec)             |
| `cache-disk`                           | critical                       | go-health `checks.Disk`, 128 MiB floor; intentionally ABSENT when no `--db` (an intentional absence, not a misconfiguration)                               | health.go:130-134                                                  |
| `checkTransitions` (hook, not a check) | observability                  | Slogs every status flip (fail at warn, rest at info; boot baseline once), so flapping is visible without scraping /readyz                                  | health_internal_test.go:38                                         |

### 4.2 Probe mechanics, verified from go-health v0.10.0 source

- `NewChecks` is the injector-free path by contract ("no samber/do injector involved",
  checks.go:37-44); the hub never touches `health.New(do.Injector, ...)`.
- `WithCriticalServices(evidenceCheckName)` names are validated at `Probe.Start`
  (lifecycle.go:32-35 calls `Validate()`), so a typo'd critical name fails at boot
  instead of silently hanging the startup latch.
- `Start` arms a **1s background refresh cache** by default (probe.go:45,173), so
  `/readyz` scrapes serve cached evaluations and cannot amplify into per-request check
  batches (options.go:69-75 documents the live-mode DoS amplifier the hub avoids by
  using Start). This also explains why the HTTP readiness contract test polls with
  `Eventually` (health_internal_test.go:91, 1s cache noted in its comment).
- `WithEvaluationHook` is panic-safe inside go-health (options.go:63-67 wraps hook
  errors as a non-critical warn), so the hub's logging hook cannot kill the probe.
- Two-phase drain exactly matches the library recipe (lifecycle.go:206-212): the hub
  calls `probe.MarkShuttingDown()` (readiness → 503 from cache, loop still fresh) then
  `probe.Shutdown()` (serve.go:122-123), followed by the SSE drain under its own live
  deadline (serve.go:126, background ctx on purpose).

### 4.3 The wedge hole is closed (morning review §5.3 → today)

| Layer                  | Mechanism                                                                                                                                             | Location                      |
| ---------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------- |
| Bound the hang         | `refreshTimeout` (30s) wraps every `QueryEvidence`; a hung source fails ONE refresh (logged, retried), never the loop and never the safety-net ticker | refresher.go:29, 150-151      |
| Stamp success honestly | `lastRefreshSuccess` stored only after query AND publish both succeed                                                                                 | refresher.go:164-169          |
| Make staleness visible | `refresh-freshness` non-critical warn at 10 min (verdict g1: warn stays 200; stale view is degraded truth, not service loss)                          | health.go:81-93               |
| Pin it                 | Hanging-source spec: blocking `SourceFunc` → `DeadlineExceeded`, loop survives, next refresh works                                                    | refresher_internal_test.go:89 |

The other formerly-invisible failure modes are closed the same way: discovery silence
(§4.1 row 3), SSE drops (§4.1 row 4), and check flapping (§4.1 row 6).

### 4.4 Boot, liveness, and systemd posture (unchanged contracts, re-verified)

- **Boot contract (decided WP-08, kept):** `serve` performs the initial refresh
  synchronously and EXITS on failure (serve.go:54-56). A broken evidence source pages
  the operator via systemd `Restart=on-failure`; it does not boot a green process
  advertising a 503 window. Consequence, understood and accepted: `/startupz` and the
  503 branch of `/v1/compliance` (serve.go:85-89) are contracts for embedders and a
  future tolerant-boot mode, unreachable through the packaged serve path today.
- **Liveness always passes** (health.go:19-23): an evidence failure must not restart
  the process.
- **Listener-first sd_notify** (bind.go:98-107): READY=1 only once the socket accepts
  traffic; watchdog loop arms only under `WatchdogSec`; `NotifyStopping` precedes the
  drain. Because the boot refresh precedes binding, READY=1 additionally implies
  "evidence published".
- **Slowloris bound** (`ReadHeaderTimeout` 10s, serve.go:22,107) and unix-socket
  hygiene (0o660, stale-remove, cleanup on shutdown, bind.go:41-61,123-125) unchanged.

---

## 5. Architecture scorecard (rubric: architecture-review assessment-rubric)

| Dimension            | Score | Evidence                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| -------------------- | ----- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Coupling             | 5     | Ports in the public core (`compliance.Source`, `EvidenceWriter`), three adapters behind one composition switch (flags.go:41-66); refresher holds the interface, never a store (refresher.go:51); the OnWrite kick is wired at the composition root by type assertion so the refresher stays store-agnostic (serve.go:58-60); no global mutable state outside ldflags `Version`.                                                                                                        |
| Cohesion             | 5     | Every package passes the one-sentence test: `events` (fan-out owner of current rows), `cache` (the hub's own SQLite evidence history), `discovery` (owns fleet repos, not evidence), `dashboard` (browser projection), `cli` (report/table/serve), `testsupport` (cross-package helpers). No utils/grab-bags.                                                                                                                                                                          |
| Modularity           | 5     | Two modules with an enforced dependency direction (core: zero private deps, AGENTS.md rule, honored); `internal/` walls block external import; `go.work` coordinates, per-module work runs `GOWORK=off`; module boundaries align with domain boundaries.                                                                                                                                                                                                                               |
| Composability        | 4     | New health checks are map entries (health.go:73-134); new sources are switch cases; new event consumers use `SubscribeFeed`; functional options where variation is real (`Hub.WithReplayCapacity` etc.). Held at 4 by two real adaptation points: `Store.OnWrite` is single-slot (cache.go, documented) and the write-hook wiring at serve.go:58 works only for `*cache.Store` (a second writable source means editing `runServe`).                                                    |
| Scalability          | 4     | Feature scaling is excellent (new core `State` flows through filters via `State.IsValid` delegation, routes.go:145). Scale thresholds are known and bounded: single-process hub, SQLite append-only evidence, in-process SSE fan-out. Horizontal scaling is not built in (single writer, single listener) and is not a current requirement.                                                                                                                                            |
| Service Orientation  | 4     | Bounded contexts are honest and data ownership is explicit (cache owns its SQLite file; discovery owns repo identity; Hub owns rows; evidence producers stay in BuildFlow). Contracts are explicit at every boundary (`compliance.Source`, `/v1/discover`, `/v1/watch`, `/v1/compliance`, SSE wire, `/readyz` payload). Held at 4 rather than 5 because everything except the discovery daemon shares one deployable; extraction-ready, not extracted (and correctly so at this size). |
| Dependency Direction | 5     | The domain core defines ports and holds zero private deps; hub adapters implement them; cmdguard/go-health/go-daemon are confined to the CLI/bind/health edge layers; no framework weaves through the domain. Third-party deps isolated behind core interfaces.                                                                                                                                                                                                                        |

**Average: 4.6 → Excellent band (4.5-5.0): "Maintain current architecture; document
patterns for replication."** The patterns ARE documented: STACK.md (decision record,
drift-checked), docs/decisions/bind-posture.md, docs/notes/federation-interop.md,
DOMAIN_LANGUAGE.md, and the AGENTS.md health-surface contract.

---

## 6. Should the hub adopt samber/do? - No (decision reaffirmed, second pass)

Same table as the morning review, same answer: six services, one composition site, no
coexisting port implementations, no scopes, explicit tested shutdown ordering, no DI
observability requirement. Adopting do would buy indirection and pay in shutdown-ordering
risk. The `NewChecks` pairing resolves identically: using the injector-free half of a
do-aware library _is_ the correct use for this wiring style.

Revisit triggers (any two fire a stack-selection re-run): service count > ~10; two
coexisting implementations of one port; embedder-side provider swapping at composition
time; DI-lifecycle observability (samber-do-auditlog) becomes a compliance requirement.

---

## 7. Residual gaps (all minor; nothing P1 remains)

1. **README health-surface documentation - FIXED in this review.** The 20-02 session's
   item f21 ("document the health surface for embedders if not already there") was still
   open; README had zero mention of `/healthz` `/readyz` `/startupz` `/version`. Added
   an embedder-facing subsection under "Embedding the dashboard" (README, this pass);
   dprint gate green.
2. **Goroutine window on probe-construction failure (observation, no action).**
   `go refresher.Run(ctx)` and `go runDiscoveryWatch(...)` start (serve.go:62-63) before
   `buildHealthProbe` (serve.go:65). If `probe.Start` failed, those goroutines live until
   the process exits, which the returned error triggers immediately. Correct in practice;
   only an embedder hosting the probe inside a longer-lived process would care.
3. **`Store.OnWrite` single-slot** (last registration wins): documented in code; a
   fan-out lands the day a second consumer appears. Conditional, not actionable.
4. **go-health federation/aggregate adoption**: YAGNI until a second probe-owning
   service exists; the posture is already recorded in ROADMAP (federation options B-D).
5. **Unreachable-through-serve contracts kept deliberately**: `/v1/compliance` 503
   branch and `/startupz` are the embedder/future-tolerant-boot contracts; their
   truthful siblings (evidence gate, readiness payload) are now real and HTTP-pinned
   (health_internal_test.go:91: fail is 503, warn stays 200, payload shape decoded).

---

## 8. Verification log (all commands run this pass, exit codes redirect-captured)

| Check                     | Command                                                                                                                                           | Result                                                                                     |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------ |
| samber/do direct usage    | `grep -rln 'github.com/samber/do' --include='*.go' .`                                                                                             | none (both modules)                                                                        |
| samber/do resolution path | `cd hub && GOWORK=off go mod why github.com/samber/do/v2`                                                                                         | cmdguard/v5 only                                                                           |
| Hub build + vet           | `cd hub && GOWORK=off go build ./... && go vet ./...`                                                                                             | ok                                                                                         |
| Hub suite                 | `cd hub && GOWORK=off go test -race -count=1 ./...` (sequential per policy)                                                                       | all 5 test packages ok (cache 1.1s, cli 8.0s, dashboard 3.1s, discovery 1.0s, events 1.2s) |
| Core build + vet          | `GOWORK=off go build ./... && go vet ./...`                                                                                                       | ok                                                                                         |
| Core suite                | `GOWORK=off go test -count=1 ./...`                                                                                                               | ok (examples/scripts: no test files)                                                       |
| Core lint                 | `golangci-lint run ./...`                                                                                                                         | 0 issues, exit 0                                                                           |
| Hub lint                  | `cd hub && golangci-lint run -c ../.golangci.yml ./...`                                                                                           | 0 issues, exit 0                                                                           |
| art-dupl CI gate          | `art-dupl -t 1 --type-aware check --baseline-path .art-dupl-baseline.json .`                                                                      | exit 0, no new clones (3 baseline groups)                                                  |
| README formatting         | `dprint check README.md`                                                                                                                          | exit 0                                                                                     |
| go-health semantics cited | read from `/home/lars/projects/go-health` (v0.10.0 sibling checkout; the 20:02 session proved `git diff v0.10.0..HEAD` empty on every cited file) | probe.go:45,173 (1s cache); lifecycle.go:32-43,206-212; options.go:63-75; checks.go:37-44  |
| Health check pins         | read `hub/internal/cli/health_internal_test.go`, `refresher_internal_test.go`, `events_test.go` structure                                         | all six checks + HTTP payload contract pinned                                              |

Discipline note: the first lint run piped through `tail` before capturing `$?`, which is
exactly the exit-code-swallowing trap the 20-02 report documented (its §d1). Caught on
review of the command, re-run with output redirection for the numbers above.

---

## 9. Files changed by this review

| File                                                                                | Change                                                                                                                    |
| ----------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| `README.md`                                                                         | + "Health surface (packaged service)" subsection under Embedding (routes, checks, drain semantics; closes 20-02 item f21) |
| `docs/architecture-understanding/2026-11-18_23-43_samber-do-health-architecture.md` | this report                                                                                                               |

No Go code changed in this pass; every health/DI behavior claim above is pinned by tests
that were already committed by the same-day execution pass and re-run green here.

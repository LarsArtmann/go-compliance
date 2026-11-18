# samber/do v2 + Health Checks: Architecture Deep Review

**Date:** 2026-11-18 19:39
**Scope:** `hub` module DI posture, health-check architecture, service orientation, composability, resilience, self-health. Core module reviewed only where it defines the ports the hub implements.
**Method:** every file that participates in wiring or liveness read end-to-end (`main.go`, `app.go`, `serve.go`, `bind.go`, `flags.go`, `health.go`, `refresher.go`, `events/`, `cache/`, `discovery/`, `dashboard/`), the `go-health` v0.10.0 library internals (`probe.go`, `lifecycle.go`, `checks.go`, `types.go`, `FEATURES.md`), the `samber/do` best-practices skill (DO-1 → DO-6), `go mod graph`, and the recorded decisions in `STACK.md` / `docs/decisions/bind-posture.md`.
**Outcome:** two defects fixed on sight (inert readiness gate, dead code), regression tests added, full hub suite green under `-race`; remaining gaps ranked into TODO_LIST.

---

## 1. Verdict

> **Q1: "Are we using samber/do v2 in combination with Health Checks superbly?!"**
>
> **We are not using samber/do at all — and that is the recorded, correct decision.**
> Zero `.go` file in this repo calls `do.*`. samber/do v2.1.0 enters the graph strictly as a
> transitive dependency (§2.1). The hub deliberately uses go-health's _injector-free_
> `NewChecks` API instead of the container path, which matches `STACK.md` line 50
> ("samber/do v2 — cmdguard's container; go-health built for it **vs manual wiring**").
> The question as posed is a category error for this codebase; the honest rephrasing —
> "is the DI posture right, and is go-health used superbly?" — answers **yes, with two
> defects found and fixed** (§5.1, §5.2) and one real resilience gap remaining (§5.3).

> **Q2: "Do we have a PROPER Service Oriented, composable, resilient, and self-health architecture?"**
>
> **Yes — genuinely, not cosmetically — with one hole that matters.** Scorecard: §4.
> The wiring is constructor-injected, context-propagated, and shutdown-ordered; the
> failure paths (source errors, daemon down, stream drops) degrade instead of crash.
> The one hole: a hung evidence query wedges the refresh loop silently while every
> health endpoint stays green (§5.3) — the exact blind spot a "self-health"
> architecture exists to catch.

---

## 2. The samber/do reality

### 2.1 Dependency graph (verified via `go mod graph`)

```
hub ──> cmdguard/v5 ──────────────> samber/do v2.1.0        (container INSIDE cmdguard)
hub ──> cmdguard/v5 ──────────────> samber-do-auditlog      (cmdguard's observability)
hub ──> go-health v0.10.0 ────────> samber/do v2.1.0        (go-health's injector path)
hub ──> go-health v0.10.0 ─────────> go-branded-id
hub ──> (direct calls to do.*: NONE)
```

`go mod why github.com/samber/do/v2` resolves through `cmdguard/v5`. At runtime the only
samber/do container that executes is the one cmdguard builds for the CLI itself; the
hub's own services never live in a container.

### 2.2 Why this is right (and when it would stop being right)

`STACK.md:50` records the decision with the rejected alternative named ("vs manual
wiring" — a decision with no rejected alternative is not a decision; this one has it).
The hub's service graph is **six long-lived components wired in exactly one function**
(`runServe`, `hub/internal/cli/serve.go:29-131`):

| Service                    | Constructed at                | Lifetime owner                     |
| -------------------------- | ----------------------------- | ---------------------------------- |
| `compliance.Source`+closer | `resolveSource` (flags.go:39) | `defer closer.Close()` serve.go:46 |
| `events.Hub`               | serve.go:49                   | `hub.Shutdown` serve.go:124        |
| `refresher`                | serve.go:53                   | ctx (serve ctx)                    |
| discovery watch            | serve.go:63                   | ctx                                |
| `health.Probe`             | health.go:27                  | `probe.Shutdown()` serve.go:121    |
| `http.Server`              | serve.go:102                  | `server.Shutdown` bind.go:121      |

For a graph this small and this stable, a container adds indirection (named-service
strings, lazy-singleton semantics, hook machinery) and removes readability. samber/do's
own best-practices canon (the `samber-do-best-practices` skill) rates code by DO-1 → DO-6
_smells_ — none of which require a container to avoid. §3.2 shows this codebase passes
every rule's intent with plain constructors.

**Revisit triggers** (any two ⇒ re-run the stack-selection decision): service count
grows past ~10; two interchangeable implementations of one port appear (the day a second
`compliance.Source` flavor must coexist, `do.ProvideNamed` beats flag-switches in
`resolveSource`); embedders need to swap providers at composition time; DI-lifecycle
observability (`samber-do-auditlog`) becomes a compliance requirement. None hold today.

### 2.3 Where go-health _does_ pair with samber/do — and what the hub uses instead

go-health v0.10.0 is dual-natured by design:

- **Injector path:** `health.New(do.Injector, ...)` aggregates every service registered
  in a samber/do container that implements `do.Healthchecker`/`HealthcheckerWithContext`
  (probe.go:161), with `WithCriticalService[T]()` typing the critical set to do's
  service-name registry (service_name.go). This is the "samber/do + health checks"
  combination — **not used here**.
- **Injector-free path:** `health.NewChecks(map[string]CheckFunc, opts...)` (checks.go:44)
  — "no samber/do injector involved", concurrent per-check execution, per-check duration,
  panic recovery, batch-deadline abandonment, fail-closed nil checks. **This is what the
  hub uses** (health.go:44-47).

The hub picks the right half of the library for its wiring style, and uses it well:
`WithCriticalServices(evidenceCheckName)` gives the evidence check restart-relevant
classification; `checks.Disk` (the battery) covers the SQLite headroom; `WithVersion`
stamps `/version`. Nothing is cargo-culted in.

---

## 3. DO-1 → DO-6 equivalence audit (manual wiring, no container)

Each samber/do anti-pattern rule, checked against its _intent_ in this codebase:

| Rule | Smell                                        | Verdict      | Evidence                                                                                                                                                                                                                                                                                                                          |
| ---- | -------------------------------------------- | ------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| DO-1 | `Must*` panic in runtime paths               | ✅ N/A-clean | No `do.*` calls exist; all constructor errors propagate as errors (serve.go:31-67, app.go:26-61). The only panic-adjacent surface is cmdguard-internal.                                                                                                                                                                           |
| DO-2 | Container without `Shutdown`                 | ✅ N/A-clean | No container. Every resource has an explicit teardown: closer (serve.go:46), server (bind.go:121), probe (serve.go:121), hub drain (serve.go:124).                                                                                                                                                                                |
| DO-3 | `Override*` in production                    | ✅ N/A-clean | No overrides; test seams are injectable writers (app.go:25) and real fakes via interfaces.                                                                                                                                                                                                                                        |
| DO-4 | Global injector/service locator              | ✅ Clean     | No package-level container or resolver; dependencies are struct fields (`refresher` fields refresher.go:33-52; `dashboard.Dashboard` holds `*events.Hub`).                                                                                                                                                                        |
| DO-5 | Resolution inside loops                      | ✅ Clean     | No per-iteration resolution anywhere; all deps bound at construction.                                                                                                                                                                                                                                                             |
| DO-6 | `Shutdown` reaching into other services      | ✅ Clean     | `Hub.Shutdown` is self-contained (broadcaster drain, events.go:278); `Store.Close` closes only its DB (cache.go:80); `probe.Shutdown` owns only the probe loop. Shutdown ORDER is coordinated by the one place that owns all lifetimes (`runServe`), which is exactly the "parent lifecycle manager" pattern the rule prescribes. |
| 4.7  | Service-locator smell (injector passed deep) | ✅ Clean     | Nothing receives a resolver; the mux closures capture already-built values (serve.go:81-100).                                                                                                                                                                                                                                     |
| 4.8  | Service storing the container                | ✅ Clean     | Services hold peers, never a locator.                                                                                                                                                                                                                                                                                             |

**Structural conclusion:** the codebase achieves everything the container rules try to
force, by construction rather than by lint. This is the strongest possible position on
the DO ladder: there is nothing for `doanalyzerv2` to find.

---

## 4. Architecture scorecard (Q2)

### 4.1 Service oriented — strong

- **Ports in the public core, adapters in the hub:** `compliance.Source` /
  `EvidenceWriter` (core `source.go`) implemented by file/buildflow/cache adapters;
  `resolveSource` (flags.go:39-64) is the single composition switch. The domain core
  keeps zero private deps (AGENTS.md module rule, honored).
- **Ownership boundaries are honest:** discovery "owns fleet repos, not evidence"
  (discovery/client.go:1-4); the cache is "the hub's own evidence store … BuildFlow's
  evidence stays in BuildFlow" (cache.go:1-5); the Hub "owns the latest published rows
  so every surface reads one source of truth" (events.go:64-67).
- **One integration surface:** `dashboard.RegisterRoutesAt` mirrors the go-health
  `RegisterRoutes(mux, ...)` embedding pattern (STACK.md wiring order 5-6).

### 4.2 Composable — strong

- Constructor injection everywhere; functional options where variation is real
  (`Hub.WithReplayCapacity/WithSubscriberBuffer/WithHeartbeat`, events.go:84-109) and
  _not_ where it would be ceremony.
- The CLI layer is composed through cmdguard v5's own DI (which is where samber/do
  actually runs in this binary) — typed per-command flag structs (`QueryFlags`,
  `ServeFlags`), injectable stdio for tests (app.go:25).
- Consumers depend on the interface the port declares: the refresher takes
  `compliance.Source`, never a concrete store; the OnWrite kick is wired at the
  composition root via type assertion (serve.go:58-60) so the refresher stays
  store-agnostic.

### 4.3 Resilient — strong (one hole, §5.3)

Verified mechanisms, with the failure they answer:

| Mechanism                                                            | Location                      | Answers                                                                    |
| -------------------------------------------------------------------- | ----------------------------- | -------------------------------------------------------------------------- |
| Trigger coalescing (cap-1 channel, non-blocking send)                | refresher.go:44, 73-78        | report bursts cost one re-derive                                           |
| 500 ms debounce absorbing the burst                                  | refresher.go:132-146          | discovery storms, multi-repo writes                                        |
| 5-min safety-net ticker                                              | refresher.go:21-25, 120       | cross-process SQLite writes the hook can't see; time-based staleness drift |
| Refresh failure = log + serve last good state                        | refresher.go:115-118, 148-150 | transient source errors must not take the dashboard down                   |
| Discovery watch backoff 30 s + resubscribe + snapshot-on-(re)connect | refresher.go:159-194          | daemon down/restart is a _normal_ state                                    |
| SSE replay ring (Last-Event-ID) + snapshot-on-connect                | events.go:211-221, store.go   | client reconnects lose nothing                                             |
| Subscriber drop counting                                             | events.go:123, 272            | backpressure visible instead of silent                                     |
| Slowloris bound (`ReadHeaderTimeout`)                                | serve.go:22, 105              | header-stall resource drain                                                |
| Graceful HTTP drain (5 s, background ctx) + unix socket removal      | bind.go:24-25, 112-127        | clean stop under load; no stale socket file                                |
| Append-only evidence (newest-record-wins queries)                    | cache.go:96-98                | out-of-order records cannot corrupt the view                               |

### 4.4 Self-health — good and now honest (was: one lying check)

The three-probe contract is textbook (health.go:14-53, go-health handlers):

- **Liveness always passes** — an evidence failure must not restart the process
  (health.go:15-17 comment pins the reasoning).
- **Readiness/startup gate on exactly one critical service: `evidence`** —
  `WithCriticalServices(evidenceCheckName)`; `/startupz` latches once it passes.
- **`cache-disk` battery** (128 MiB floor) keeps the SQLite append path honest, and is
  correctly _absent_ when no cache DB is configured (health.go:38-42 — an intentional
  absence, not a misconfigured check).
- **systemd posture:** listener-first ordering means `NotifyReady` fires only once the
  socket accepts traffic (bind.go:91-107), and — because `runServe` performs the initial
  synchronous refresh _before_ binding (serve.go:53-56 → 118) — READY=1 additionally
  implies "evidence published". Watchdog loop arms only when `WatchdogSec` is set;
  `NotifyStopping` precedes the drain.
- **Two-phase drain:** `probe.MarkShuttingDown()` (readiness → 503 from cache, loop
  still fresh) _then_ `probe.Shutdown()`, matching go-health's documented two-phase
  recipe; `Hub.Shutdown` then drains SSE subscribers under its own live deadline
  (serve.go:120-124 — the background-ctx comments explain _why_).

**Defect fixed today:** the `evidence` check tested `hub.Rows() == nil`, but `Rows()`
returned `make([]RepoRow, len(h.rows))` — **never nil** (proven: `make([]T, 0) != nil`).
`ErrEvidenceNotPublished` could not fire, ever; the readiness gate was decorative, and
the SSE/stream "send snapshot only if rows exist" gates were equally inert. Today's
serve ordering masked it (initial refresh precedes probe construction; boot failure =
process exit), but any future refactor toward tolerant async boot would have shipped a
`/readyz` that says READY with an empty dashboard and a dead source. Fixed at the root —
see §5.1.

---

## 5. Findings (ranked; fixed items verified)

### 5.1 FIXED — inert readiness gate (`Hub.Rows()` nil-contract violation)

- **Where:** `hub/internal/events/events.go` `Rows()`; consumers `health.go:30`,
  `events.go:216`.
- **Root cause:** documented contract "nil before the first snapshot" was not
  implemented; `make` never returns nil.
- **Fix:** `Rows()` (events.go:134) now returns nil while nothing has been published; an _empty_
  published fleet is deliberately non-nil (empty is a valid ready state — the doc
  comment now says the split is load-bearing).
- **Pinned by:** new spec "returns nil rows only until the first publish; an empty
  fleet is published" (events_test.go) and a new gate spec through the real probe —
  `buildHealthProbe` → `Evaluate` → `StatusFail`/`ErrEvidenceNotPublished` before
  publish, `StatusPass` after (`hub/internal/cli/health_internal_test.go`).
- **Blast radius checked:** every `Rows()` caller audited (health.go:30, events.go:176,
  189, 222; dashboard.go:89, 168; routes.go:68) — all len-based or nil-gated, none
  breaks. Full suite green under `-race`.

### 5.2 FIXED — dead duplicate nil-check in `queryEvidence`

- `app.go` carried an unreachable `if err != nil` after the error had already been
  returned 8 lines above (gopls `nilness`: "impossible condition"). Removed. The gopls
  warning on the package is gone.

### 5.3 OPEN (P1) — refresh loop can wedge with all health green

- **Scenario:** `refresher.Refresh` calls `source.QueryEvidence(ctx, …)` with the _serve_
  context — no per-refresh deadline (refresher.go:95). A hung SQLite read (lock wait on
  a busy WAL, pathological I/O) blocks the loop _inside_ Refresh; the 5-minute
  safety-net ticker fires into a select nobody is listening to; every health check
  stays green (evidence = "rows exist" — they do, stale; disk = fine) while the
  dashboard silently freezes forever. Only process restart recovers.
- **Why the current checks can't see it:** `evidence` gates on _existence_, never
  _freshness_; go-health's batch deadline bounds the _check_, not the monitored loop.
- **Recommended design (owner decision on numbers):**
  1. Per-refresh timeout: `ctx, cancel := context.WithTimeout(ctx, 30*time.Second)`
     around `QueryEvidence` in `Refresh` — converts a wedge into the existing
     log-and-retry path.
  2. Freshness signal: `lastRefreshSuccess atomic.Int64` (unix nano) stored on success
     in `Refresh`.
  3. Non-critical check `refresh-freshness`: fail when `now - lastRefreshSuccess >
     2×refreshSafetyNet` (10 min). Non-critical ⇒ classification `warn`, `/readyz`
     stays 200 (degraded-but-serving is exactly the current posture), but the
     degradation becomes _visible in the health payload_ instead of only in logs.
- Tracked in TODO_LIST (High impact).

### 5.4 OPEN (P2) — discovery daemon health is invisible

Best-effort by design (never-covered completeness), correctly non-critical; but the
only signals are slog lines. A non-critical `discovery` check (age of last successful
snapshot/watch) would turn "daemon down for 3 days" from a log line into a standing
`warn`. Alternatively accept and document — but then the OnWrite-hook limitation plus
this makes the never-covered set quietly untrustworthy.

### 5.5 OPEN (P2) — SSE hub self-observability exists but surfaces nowhere

`Hub.Dropped()` / `SubscriberCount()` (events.go:272-275) feed nothing: no metric, no
check, no payload. Cheapest honest step: include both in the readiness JSON as
non-check fields, or a warn check on sustained drop growth. Today a pathological
subscriber can drop events invisibly (replay ring softens, but does not eliminate, the
gap).

### 5.6 OPEN (P2, pre-existing) — hub module lint is red

`golangci-lint run -c ../.golangci.yml ./...` exits 1 with 16 findings
(bodyclose ×1 in `discovery/client.go:165` — the closure-close shape, already a
TODO_LIST row; gochecknoglobals ×9; makezero ×3; ireturn ×1; varnamelen ×2). **None are
attributable to this review's diff** (verified: all flagged lines pre-date it; the
events.go line numbers in the linter output are the pre-edit positions). Decide
as an owner: fix the class (nolint-with-reason for the idiomatic globals) or configure
the linters — a permanently red lint is a gate that guards nothing.

### 5.7 Observations (no action required)

- `/v1/compliance`'s 503 branch (serve.go:82-87) is unreachable through `serve` today
  (initial refresh precedes listening; its failure aborts boot). Keep it: it is the
  correct contract for embedders and for a future tolerant-boot mode, and it now has a
  truthful sibling in the fixed readiness gate.
- `Store.OnWrite` is single-slot (last registration wins, cache.go:89-94) — documented,
  fine for the single-consumer hub; embedders composing multiple listeners would need
  a fan-out. Not a defect today.
- `refresher.Evidence()` (for `/v1/compliance`) and `hub.Rows()` (for the UI) are two
  projections of one derivation — Rows are derived from the evidence in the same
  `Refresh` call (refresher.go:104-110), so the "one source of truth" claim holds by
  construction.
- go-health `Probe.Start` validates that every `WithCriticalServices` name actually ran
  in the first batch (lifecycle.go:120-145) — so a future typo'd critical name fails at
  boot instead of hanging the startup latch. The hub benefits automatically.

---

## 6. Should the hub adopt samber/do? — No (decision reaffirmed)

| Criterion                    | Today                                            | Container payoff?                                                   |
| ---------------------------- | ------------------------------------------------ | ------------------------------------------------------------------- |
| Long-lived services          | 6                                                | No — one screen of wiring                                           |
| Composition sites            | 1 (`runServe`)                                   | No — nothing to share                                               |
| Multiple impls per port      | 3 Source flavors behind a flag, never coexistent | No — `ProvideNamed` would solve a problem that doesn't exist        |
| Request/tenant scopes        | none (SSE is not scoped)                         | No                                                                  |
| Lifecycle coordination       | explicit, ordered, tested                        | No — and DO-6's cross-service-shutdown hazard would be a _new_ risk |
| DI observability requirement | none                                             | No                                                                  |

Adopting do here would buy indirection and pay in shutdown-ordering risk. The recorded
decision stands. The go-health pairing question resolves the same way: `NewChecks` is
the deliberately injector-free half of a do-aware library — using it _is_ the superb
use for this wiring style.

---

## 7. Verification log

| Check                                     | Command                                                                                                   | Result                                                            |
| ----------------------------------------- | --------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| samber/do usage                           | `grep -rln "samber/do" --include='*.go'` → none; `go mod why`, `go mod graph`                             | transitive via cmdguard/v5 + go-health only                       |
| Build                                     | `GOWORK=off go build ./...` (hub)                                                                         | ok                                                                |
| Vet                                       | `GOWORK=off go vet ./...` (hub)                                                                           | ok                                                                |
| Events suite (fix 5.1)                    | `GOWORK=off go test -race -count=1 ./internal/events/`                                                    | ok 1.197s                                                         |
| CLI suite (fixes 5.1+5.2 + new gate spec) | `GOWORK=off go test -race -count=1 ./internal/cli/`                                                       | ok 7.094s                                                         |
| Full hub suite (sequential per AGENTS.md) | `GOWORK=off go test -race -count=1 ./...`                                                                 | all packages ok                                                   |
| Lint                                      | `golangci-lint run -c ../.golangci.yml ./...`                                                             | exit 1, 16 findings, all pre-existing (§5.6), zero from this diff |
| go-health semantics cited                 | read from `/home/lars/projects/go-health` v0.10.0 source (lifecycle.go, checks.go, types.go, FEATURES.md) | verified, not assumed                                             |
| make-slice nil semantics                  | executable proof (`make([]row,0) == nil` → false)                                                         | confirmed 5.1's root cause                                        |

---

## 8. Files changed by this review

| File                                       | Change                                                                                 |
| ------------------------------------------ | -------------------------------------------------------------------------------------- |
| `hub/internal/events/events.go`            | `Rows()` honors nil-before-first-publish contract (+ load-bearing doc)                 |
| `hub/internal/events/events_test.go`       | regression spec for the nil/non-nil/empty split                                        |
| `hub/internal/cli/health_internal_test.go` | NEW — pins readiness gate fail→pass through the real probe                             |
| `hub/internal/cli/app.go`                  | removed unreachable duplicate error check (gopls nilness)                              |
| `TODO_LIST.md`                             | + P1 refresh-freshness, P2 discovery check, P2 SSE observability, P2 lint-red decision |

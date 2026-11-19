# go-cqrs-lite

- Reviewed: 2026-11-19 (group E framework audit)
- Consumer modules: core (root `github.com/larsartmann/go-compliance`) and hub (`hub/`). **go-cqrs-lite: absent from both** (`rg 'go-cqrs-lite' go.mod hub/go.mod` → nothing; only mention is the STACK.md:73 exclusion row). Note for future readers: `hub/go.mod:86` carries `github.com/larsartmann/templ-components/htmx v1.27.0 // indirect` — that is templ-components' htmx subpackage, NOT cqrs-htmx, and no cqrs-lite module appears anywhere in either require graph.
- Library reviewed at: `git -C /home/lars/projects/go-cqrs-lite describe --tags --abbrev=0` → `cmd/cqrs-lint/v4.22.2`; `git log -1` → `7c7724a4f 2026-11-18 chore(deps): bump go-etag server v0.6.1 -> v0.6.2 (taskmanager example)`. Remote tags fetched (`git fetch --tags` no-op — local == origin).
- Tagging model: **per-module version trains, not lockstep tags** (the repo's own cqrs-lint V006 suppressions in dependents' go.mods say so). Root module `github.com/larsartmann/go-cqrs-lite` has bare tags only up to `v4.0.0` (stale facade); the living releases are `system/v4.19.0` (2026-11-15), `metaengine/v4.26.0` + `metaengine/sqliteengine/v4.9.1` (2026-11-15), `catalog/v4.15.0`, newest overall `cmd/cqrs-lint/v4.22.2` (2026-11-17).
- **104 `go.mod` sub-modules** (count includes `cmd/cqrs-lint/v4/testdata/*` fixtures and 8 `example/*` demo apps). A `core/v5` directory already exists at HEAD — the v5 cut (ADR-0123: low-level modules `event/command/decider/storage/stack/kv` legacy→removed, constructor renames, `cmd/cqrs-upgrade` for mechanical migration) is in preparation.
- Skill contract honored before auditing: `~/.config/crush/skills/go-cqrs-lite/SKILL.md` — new apps MUST build on `system.New(ctx, DomainConfig, DeploymentConfig)` + `metaengine` `Store`/`Query[Q,R]` (cost-based planner across memory/sqlite/turso/pebble/bbolt/badger/duckdb/postgres…); hand-wiring low-level modules is forbidden for new code; ADR-0142 I/O rule routes ALL durable state through metaengine engines.

## Verdict: NOT-ADOPTED-justified

The hub is a **read-model projection service, not a command-sourced domain**, and the mismatch is in the data model, not the feature list. The honest framing question — which cqrs-lite capabilities are isomorphic to what the hub hand-rolls? — has a sharp answer: **almost none, because the hub's derived state is `f(evidence_rows, now)`, while every cqrs-lite read model is `f(events)`**. Evidence arrives as pre-aggregated rows (`RunResult`) appended to a shared SQLite table by *external* processes through the reporter SDK (`hub/internal/cache/cache.go:101` `WriteEvidence`; schema `cache.go:30-55`); there are no commands to refuse, no invariants to enforce, no aggregate identities, no event stream to fold — the append-only `evidence_runs` table already IS the log. The hub's fold is a **sliding-window SQL aggregation** (`cache.go:172-323`, `since = now − window` per refresh at `hub/internal/cli/refresher.go:148-153`) plus a **wall-clock state derivation** (`DeriveState(evidence, now, staleAfter)`, `state.go:112`): a repo goes `stale` with *zero* writes happening. That last property is why the 5-minute safety-net ticker exists (`refresher.go:21-25`) — no incremental event fold can express a state transition that fires without an event. Routing the hub through `metaengine.Query[Q,R]` would require (1) inventing an event journal to duplicate the log SQLite already keeps, (2) an in-process ingest command API to replace direct file writes from other processes, and (3) keeping the ticker + full re-derive anyway for staleness and window slide. The 104 modules are a superpower for the fleet's command-sourced domains (38 direct `system` consumers per the 2026-10-24 census) and a liability here: a `go.sum` train of per-module version pins, a v5 rename window on the horizon, and a composition root whose headline features (bus, deciders, snapshots, timers, DLQ, catch-up subscribers) map to machinery the hub pointedly does not have.

**Flip trigger:** the hub grows a **command side it owns** — the moment writes carry invariants (operator acknowledge/snooze/mute of repos, annotations that must be consistent under concurrency, retention policy changes), the "adapter writes with no command-side invariants" premise of STACK.md:73 dies and `system.New` becomes the correct spine (hub only; core stays pure). Secondary trigger: audit-grade replay becomes a *hard* requirement with schema-versioned fold code (see gap analysis — today even that is mostly covered by the existing append-only log). If either fires, adopt per the skill: `system.New` + `metaengine.Store`/`Query[Q,R]`, never the low-level modules.

## Derived score

Counterfactual rubric (NOT-adopted framework): Fully (1) when the consumer's existing solution is equivalent-or-better for THIS consumer's needs, Partial (0.5) when adoption would add real marginal value today, Missed (0) when the consumer lacks a capability it genuinely needs. N/A rows carry a one-line reason.

- Capability inventory rows: **N = 12**
- N/A rows: 7 (event store, command dispatch, snapshots, engine swap, idempotency/dedup, tooling, scenario framework) → **applicable A = 5**
- Fully = 4 × 1, Partial = 1 × 0.5, Missed = 0 → **Σ = 4.5**
- **Score = round(100 × 4.5 / 5) = 90 — very high band (≥85). Publish "4.5 of 5".**

## Capability inventory

| Capability | Status | Library evidence (path) | Consumer evidence (path:line) |
| --- | --- | --- | ---|
| Event store / journal (append-only, seekable, versioned streams) | N/A — no events exist; writers are external processes appending aggregated rows directly to shared SQLite; a journal would duplicate `evidence_runs` | `system/v4/README.md` (EventStore wiring); skill SSE matrix (`event.SeekableJournal`) | `hub/internal/cache/cache.go:30-55` (append-only `evidence_runs` + index) |
| Command dispatch / deciders (load→fold→decide→save→publish) | N/A — "evidence ingest is adapter writes with no command-side invariants" (the recorded exclusion, STACK.md:73) | `system/v4/README.md` (DomainConfig quickstart) | `reporter.go:49-53` (`EvidenceWriter` — fire-and-forget appends, nothing to decide) |
| Projections / read models (`metaengine.Store` + `Query[Q,R]` `OnRecord` folds, cost-planned engines; `projectionhost`) | Fully (solved correctly differently) — hub read models are windowed SQL + pure folds; state depends on `now`, which incremental event folds cannot express | `metaengine/v4/README.md` (Three Roles, fold-return-type-as-ADT); skill read-model matrix | `cache.go:172-323` (three window queries), `fold.go` (`FoldEvidence`), `hub/internal/events/rows.go:29` (`RowsFromEvidence`), `state.go:112` (`DeriveState(…, now, staleAfter)`) |
| Replay (seekable journal, catch-up subscribers, crash-restart checkpoints; `watermill.CatchUpSubscriber`) | Partial — the one real future upside: schema-versioned, byte-exact "what did we believe at T" reconstruction. Today the append-only log + `recorded_at` already answers historical-window queries via SQL; what it cannot do is pin fold-code versions | skill "SSE: Which implementation?" + "Dead-letter" matrices; `metaengine/v4/sse_replay.go` | `cache.go:35-47` (`recorded_at` on every row); `events.go:88-94` (SSE reconnect replay ring, capacity 256) |
| Snapshots (aggregate state checkpointing) | N/A — no aggregates; a full "fleet snapshot" is one query over ~280 repos | `snapshot/v4` module (`git ls-files '*go.mod'` list) | `events.go:157-168` (`PublishSnapshot` — one struct, rebuilt per refresh) |
| Engine interchange (memory, sqlite, turso, pebble, bbolt, badger, duckdb, postgres, mysql, bigtable, dgraph, iroh…) | N/A — one SQLite file, single operator, ~280 repos; go-sqlitestore chosen with "Postgres-scale needs" as its own revisit trigger (STACK.md:44) | skill REQUIRED section (engine list); `metaengine/*engine/v4` modules | `STACK.md:44` |
| Timers / scheduling (system-owned `Domain.Timers`, saga timers, scheduling engine) | Fully (solved correctly differently) — liveness is a 40-line trigger/debounce/safety-net loop; saga-timer machinery would own nothing the hub does | skill core loop; `scheduling/v4`, `system` timers | `refresher.go:14-42` (500ms debounce, 5-min safety net, 30s timeout, 30s discovery backoff) + `refresher.go:178-210` (Run loop) |
| SSE of read models (`metaengine.ServeSSE[V]` over `Watcher[V]`, ring replay, Last-Event-ID, heartbeat) | Fully (solved differently, domain-richer) — the hub's stream carries typed snapshot/update/diff payloads feeding DataStar patches; `ServeSSE` streams materialized values generically | `metaengine/v4/sse_replay.go:9-27`; skill SSE matrix row 2 | `hub/internal/events/events.go:18-56` (event names + payloads), `:213-244` (snapshot-on-connect + heartbeat); `hub/internal/dashboard/dashboard.go:1-15` (DataStar projection) |
| Health & lifecycle integration (system-wired health checks, lifecycle, `systemscenario`) | Fully — hub wires go-health directly with domain checks; the boot contract (exit on failed first refresh, AGENTS.md "Live Architecture") is deliberately simpler than a system lifecycle | skill REQUIRED section (health/lifecycle auto-wiring) | `hub/internal/cli/health.go:74-78` (critical `evidence` gate), `:138`; `AGENTS.md:108-114` (boot contract) |
| Idempotency / dedup (`metaengine.DedupStore`, `middleware.*Idempotency`) | N/A — append-only log with fold-time resolution (last-write-wins); no dispatch layer to dedup | skill "Dedup: Which store?" matrix | `cache.go:101-138` (append; older results ignored at fold, `reporter.go:59-61`) |
| Tooling (`cqrs-lint`, `cqrs-bench`, `cqrs-upgrade`, api-stability checks) | N/A — no cqrs code in the consumer to lint/bench/migrate | `cmd/cqrs-lint/v4` (newest tag v4.22.2) | — |
| System-level BDD testing (`systemscenario`: Given/When/Then over the production composition root) | N/A — ginkgo/gomega + the M27 E2E contract test already pin the boot path end-to-end | skill "Testing such apps" paragraph | `hub/internal/cli/refresher_internal_test.go` (store write→OnWrite→Kick→debounce→refresh→patch, AGENTS.md:95-97) |

## Gap analysis

### (a) Is the hub's fold → derive → rows → SSE pipeline a metaengine projection? No — three structural reasons

1. **No event stream.** `metaengine.Query[Q,R]` folds are `OnRecord` handlers applied to journaled event records (`metaengine/v4/README.md` "Quick Example": `store.Apply(ctx, "UserCreated", …)`). The hub's inputs are rows written by *other processes* directly into `evidence_runs` — the hub's own `OnWrite` hook only sees this process's writes (AGENTS.md M27 note), which is exactly why the safety net exists. Feeding metaengine means interposing an event-journal writer between reporters and SQLite: a new ingest API, a new failure mode, and a second copy of the log — for zero capability gained.
2. **State is time-dependent.** `DeriveState(evidence, now, staleAfter)` (`state.go:112`) transitions a repo to `stale` as the clock passes `staleAfter` with no write at all. Every metaengine read model updates only on `Apply`. An event-sourced purist would emit synthetic "time passed" events; the hub's honest answer is a 5-minute ticker + full re-derive (`refresher.go:21-25, 179-189`) — 20 lines that make the wall-clock dependency *visible* instead of hiding it in a scheduler.
3. **The window slides.** `since = now − window` per refresh (`refresher.go:148-153`); the windowed aggregates (`cache.go:172-323`) are queries over a moving predicate, not keyed documents that can be maintained incrementally. SQL is the native tool; a maintained projection keyed by repo would need invalidation on every window slide — recompute dressed as a projection.

What the hub's `Hub`/`Dashboard` pair *is* isomorphic to: `metaengine.ServeSSE[V]` + `Watcher[V]` (materialized values → SSE with ring replay). The hub built that shape directly on go-sse (`events.go:116-132`) with domain payloads and DataStar patches — already audited as the transport layer's job (STACK.md:45-46).

### (b) Would core ever take a go-cqrs-lite dep? No

Core's contract is two public deps (go-branded-id, go-error-family — STACK.md:29) and its fold (`fold.go`, ~350 lines of pure functions over `WindowHealth`/`LatestRun`/`RepoSteps`… inputs) **is the product**: a source-agnostic evidence algebra any producer can implement against (`source.go:12-26`). A cqrs-lite dependency would drag `event/journal/record` types into every downstream import graph, break the public-clean posture, and buy nothing — the fold's inputs are aggregated window facts, not events, by design (BuildFlow's v1 contract is already aggregated on the wire).

### (c) GAIN vs LOSE if adopted anyway

GAIN: schema-versioned replay with pinned fold code (the only Partial row); engine swap (unused at this scale); saga timers/DLQ/checkpointed catch-up (no producers or workers to use them). LOSE: the 500ms-debounced trigger loop's simplicity and testability (M27 E2E pin, AGENTS.md:95-97); the exit-on-failed-boot contract (a system lifecycle would absorb and soften it); the frozen byte-exact v1 posture's auditability; a `go.sum` spanning system+metaengine+sqliteengine plus their per-module version trains; and a v5 rename window (skill: "the v5 cut renames some constructors") landing mid-fleet-migration. For THIS consumer the exchange rate is strictly negative today.

## Version currency

- Very active: ~1028 commits since 2026-10-20; HEAD 2026-11-18; newest tags 2026-11-15…17. `system` promoted out of EXPERIMENTAL 2026-11-03; `metaengine` per-engine modules promoted Stable 2026-11-15 (per skill stability note: 42 direct + 39 indirect fleet consumers for `metaengine/v4`).
- Stability caveat: v4 surface is stable, but v5 is a declared breaking cut (ADR-0123: `stack/*` presets and v1 read-model tiers removed; `On`/`OnTyped`/`Delta` deprecated→removed; `transport/http.SSEBroker` already deleted 2026-10-17). A consumer adopting today inherits a migration obligation the hub has no offsetting need for.
- Per-module tagging means "which version is go-cqrs-lite?" has no single answer — the answer is a pin set (`system/v4.19.0`, `metaengine/v4.26.0`, …), which is itself fleet friction when untaken.

## Fleet leverage

go-cqrs-lite is the fleet's command-sourcing backbone and it is earning its keep elsewhere (38 direct `system` consumers; the hub's own sibling BuildFlow-style domains). The two-way honesty: fleet consolidation argues for every *event-shaped* service to converge on it — and against *non-event-shaped* services (this hub, project-discovery-daemon's read side) being dragged onto it, because a wrong framework adoption costs the fleet a permanent divergence between the framework's model and the service's truth. The correct leverage relationship already exists: go-compliance core is itself a fleet-consumable library (evidence fold + state algebra on the two-primitive dep diet), i.e. the hub's *peer*, not a layer over cqrs-lite. Shared substrates the two do and should share — go-sse, go-health, go-branded-id, go-error-family — are already shared. The hub sits outside go-cqrs-lite's target shape by architecture, not by oversight; STACK.md:73 recorded exactly that, and this audit confirms it against the library's current `system`+`metaengine` contract rather than the older 70-module snapshot the row cites (the row's module count is stale — worth a one-word refresh when STACK.md is next edited).

## Ranked opportunities (impact × ease)

1. **None actionable in code.** Update STACK.md:73's parenthetical module count "(70)" → current shape ("104 sub-modules; `system` composition root + `metaengine` planner") so the row keeps citing reality — trivial ease, marginal impact, fold into the next STACK.md edit.
2. Standing watch: if a hub command side ever appears (acknowledge/mute/annotate), re-run this audit *before* hand-rolling invariants — that is the moment `system.New` is cheaper than a bespoke write path (the flip trigger above).

## How to verify

Executed 2026-11-19, all green unless noted:

```bash
# 1. Absence in both consumer modules and the whole tree (docs excluded — this report names the lib)
rg -l 'go-cqrs-lite' --glob '!vendor/**' --glob '!docs/**' /home/lars/projects/go-compliance
# → /home/lars/projects/go-compliance/STACK.md only (the "Deliberately excluded" row), exit 0

# 2. Library version metadata
git -C /home/lars/projects/go-cqrs-lite describe --tags --abbrev=0   # cmd/cqrs-lint/v4.22.2
git -C /home/lars/projects/go-cqrs-lite log -1 --format='%h %cs %s'  # 7c7724a4f 2026-11-18 …
git -C /home/lars/projects/go-cqrs-lite ls-files '*go.mod' | wc -l   # 104
git -C /home/lars/projects/go-cqrs-lite tag --sort=-creatordate | head -8
git -C /home/lars/projects/go-cqrs-lite log -1 --format='%h %cs %s' system/v4.19.0      # 2026-11-15
git -C /home/lars/projects/go-cqrs-lite log -1 --format='%h %cs %s' metaengine/v4.26.0  # 2026-11-15
git -C /home/lars/projects/go-cqrs-lite tag | grep -v '/' | tail -3                    # … v4.0.0 (root facade stale)

# 3. The three structural arguments, at their sources
sed -n '30,55p' /home/lars/projects/go-compliance/hub/internal/cache/cache.go   # append-only evidence_runs schema
sed -n '101,110p' /home/lars/projects/go-compliance/hub/internal/cache/cache.go # WriteEvidence (external writers)
rg -n 'func DeriveState' /home/lars/projects/go-compliance/state.go             # state.go:112 — (evidence, now, staleAfter)
sed -n '14,42p' /home/lars/projects/go-compliance/hub/internal/cli/refresher.go # debounce/safety-net/timeout constants

# 4. metaengine's model really is fold-over-applied-events
sed -n '25,45p' /home/lars/projects/go-cqrs-lite/metaengine/v4/README.md         # store.Apply(...) → ExecuteTyped
rg -n 'ServeSSE|Last-Event-ID' /home/lars/projects/go-cqrs-lite/metaengine/v4/sse_replay.go | head -4

# 5. Consumer's SSE-of-read-model equivalent
sed -n '116,132p' /home/lars/projects/go-compliance/hub/internal/events/events.go  # NewHub: broadcaster + ring store + heartbeat

# 6. Recorded decision this audit defends
rg -n 'go-cqrs-lite' /home/lars/projects/go-compliance/STACK.md                 # :73 exclusion row
```

Spot-checks confirmed: `cache.go:30-55`/`:101` as cited; `state.go:112` signature; `refresher.go:14-42` constants; `metaengine/v4/README.md` Quick Example applies discrete events; STACK.md:73 row text.

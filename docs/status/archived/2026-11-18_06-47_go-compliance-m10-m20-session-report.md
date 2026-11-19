# Session report: M10–M20 push (realtime stack built end to end)

Date: 2026-11-18 06:47 (session ran 2026-11-17 evening → 11-18 morning)
Scope: this session only — resuming from the M1–M10 progress report at
`2026-11-17_20-44_go-compliance-m1-m10-progress.md`. Plan of record:
`docs/planning/2026-11-17_20-01_go-compliance-pareto-execution-plan.md`.

**Tree state at report time: GREEN.** All five hub suites pass
(`go test ./internal/...`), core module untouched this session (last
verified green at resume). Working tree clean besides this report (daemon
auto-commits everything else).

---

## a) FULLY DONE (verified this session)

### M10 — discovery client compile fix (F047–F049 remainder)

- `client.go`: `json.NewDecoder(resp.Body).Decode` → `io.ReadAll` +
  `json.Unmarshal` (json/v2 has no NewDecoder); kept NewIOError/NewParseError
  wrapping.
- `client_test.go`: deleted the bogus `jsonNewDecoder` helper; request body
  now decoded inline; imports fixed.
- **Missing piece found:** no ginkgo bootstrap existed → "no tests to run"
  was masking a pass. Added `discovery_suite_test.go`. Suite green.
- M9 re-verified green (`internal/cache`) and flipped complete.

### M11 — events backbone (F050–F053)

`hub/internal/events`:

- `rows.go`: `RepoRow` projection (state/last-run/runs/failings/findings/
  severities/worst-steps/verdict), `RowsFromEvidence` (DeriveState per row),
  `DiffRows` → **added / changed / removed** three-way split.
- `store.go`: in-memory ring `EventStore` (numeric-ID replay, capacity
  eviction, zero/unknown-ID semantics).
- `events.go`: `Hub` — go-sse broadcaster + ring replay + monotonic IDs +
  OnDrop counter + 15s heartbeat; **Hub owns current rows** (single source
  of truth: `Rows()` deep-copies); `PublishSnapshot` / `PublishUpdate` /
  `PublishRefresh(next)` (diff-vs-internal, no-op when nothing changed);
  `Handler()` JSON SSE with Last-Event-ID replay; `SubscribeFeed()` typed
  feed; `Snapshot`/`Update` payloads carry `counts` + `total`;
  `SnapshotJSON`, `CountStates`.
- Tests: rows derivation table (all four states), diff semantics,
  snapshot-then-live ordering, replay-after-ID, refresh gating, row-copy
  isolation, counts on wire, monotonic IDs; internal ring + drop-counter
  tests (drain + timeout-ctx before Shutdown).

### M12 — DataStar bridge (F054–F057)

`hub/internal/dashboard` + deps `go-datastar`, `go-datastar/static`,
`go-datastar/datastartest`:

- `StreamHandler()`: SSE of patches — signals patch (counts/total) +
  elements patches; updates project as **append** (added rows) /
  **replace** (`tr[data-repo=X]`) / **remove** (disappeared) — the minimal
  patch per refresh.
- `ScriptHandler()`: serves pinned `dstarstatic.Bytes()` with day caching.
- templ toolchain added (`go get -tool github.com/a-h/templ/cmd/templ
  @v0.3.1100`); minimal `rows.templ` (full UI came with M13).
- datastartest E2E specs asserting real patch events (modes, selectors,
  signals via `UnmarshalSignals`).

### M13 — fleet table UI (F058–F062)

templ-components v1.27.0 (skill loaded before authoring):

- `view.templ`: `layout.Base` (no HTMX — DataStar owns realtime) +
  `Container` + `PageHeader` + `datastar.LiveRegion` (RetryAlways — same
  clean-EOF rationale as go-health-dashboard) wrapping the table; Play-CDN
  Tailwind until compiled CSS exists; `SDKScript` self-hosted via ScriptSrc.
- `statcards.templ`: Grid of StatCards (Total/Failing/Stale/Never covered)
  with icons (Server/ExclamationTriangle/Clock/Search) and semantic tones.
- `rows.templ` (enriched): state Badge (error/warning/success/neutral),
  severity cells (ERR/WARN/INFO, canonical order — `sortedSeverities`
  kills map-order nondeterminism), RelativeTime with "never" empty-text,
  worst-steps native `<details>` expansion per row.
- `view.go`: ViewModel; patch-stream also re-renders `#stat-cards`.
- **Generated `*_templ.go` COMMITTED** — verified fleet precedent
  (go-health-dashboard .gitignore explicitly says commit).
- F062 render tests: full-page assertions + empty state.

### M14 — routes, negotiation, chips (F063–F067)

- `routes.go`: `Routes`/`DefaultRoutes`/`RegisterRoutes(mux)` (one-call
  embed)/`RegisterRoutesAt` (base-path support); page route content
  negotiation (Accept q-values, JSON→`SnapshotJSON`, wildcard→HTML).
- Filter chips: signal-based buttons (`data-on-click` sets `stateFilter`,
  `data-class` active styling, `aria-pressed`) + rows carry `data-show`
  expressions + `?state=` server-side initial filter (no-JS path) with
  validation.
- CLI: `--source cache --db path` (resolveSource now returns a Closer for
  the SQLite store); `serve` mounts the dashboard at `/`, `/events`,
  `/datastar.js`.
- README: "Embedding the dashboard" quickstart (F067).
- 11 dashboard specs green incl. negotiation table, chips wiring, prefix
  routes.

### M15 — health + systemd (F068–F071)

- `health.go`: go-health `NewChecks` probe — critical `evidence` check
  (hub has published state), `cache-disk` headroom check when `--db` set
  (`cacheDirForDB` guards the empty case); `WithVersion`; probe.Start.
- `bind.go` (born here, extended in M20): listener-first `serveListener`
  with go-daemon `NotifyReady` / `NotifyStopping` / `Watchdog` loop
  (WatchdogSec-aware).
- serve wires `probe.RegisterRoutes(mux, health.DefaultRoutes())` →
  `/healthz /readyz /startupz /version`; graceful order:
  MarkShuttingDown → probe.Shutdown → hub.Shutdown.
- `deploy/compliance-hub.service`: hardened unit (DynamicUser,
  ProtectSystem=strict, empty CapabilityBoundingSet, AF restriction,
  WatchdogSec=30s, StateDirectory) modeled on go-health-dashboard's NixOS
  module; header documents install.
- Smoke (F071): serve answers 200 on all six endpoints + readyz JSON
  carries the evidence check with status pass.
- Deps pinned: go-health v0.10.0; go-daemon at the notify-capable commit
  `e8e1a1f` (v0.4.0 tag lacks notify.go — pinned pseudo-version,
  documented here).

## b) PARTIALLY DONE

### M20 — bind posture (F090–F093)

- **Done:** decision doc `docs/decisions/bind-posture.md` (localhost
  default, loud non-loopback warning, unix-socket first-class, no auth
  until g1); `bind.go` unix-socket listen mode (`unix:` prefix: mkdir,
  stale-socket removal, 0o660 perms, shutdown cleanup); `isLoopbackBind`;
  non-localhost warning wired into serve output.
- **Not done:** F093 test (non-localhost warning fires / unix-socket
  serve round-trip) — interruption landed here. Code compiles and suites
  are green; the new paths are untested.

### Housekeeping debt from this session (small, deliberate)

- No `go test -race` pass yet on the new concurrent code (hub, stream).
- `waitForSubscriber` poll helper duplicated in two test packages.
- templ-commit note not yet added to `.gitignore` (fleet convention
  comment).

## c) NOT STARTED

- M17 docs set + dprint (F077–F081)
- M18 flake.nix (F082–F085)
- M19 CI workflows (F086–F089)
- M21 examples + guide + screenshots (F094–F096)
- M22 benchmarks + baseline (F097–F098)
- M23 BuildFlow TODO_LIST pointer + STACK errata (F099–F100)
- M24 v2 design doc (F101–F102)
- M25 federation research (F103–F104)
- M26 overview endgame (F105–F106) — **blocked on owner question g2**
- M27 re-derive on writes + debounce + E2E (F107–F109; N+1
  `queryLatestRunCounts` fix prerequisite recorded)
- Final pass (whole-repo build+test, plan-vs-tree audit)

## d) TOTALLY FUCKED UP (and what it cost)

1. **Invented library APIs instead of reading the pinned source first —
   the session's dominant failure mode.** Six distinct instances:
   `json.MarshalOptions{Deterministic: true}` (real API:
   `json.Deterministic(true)` variadic), `GinkgoT()` as `testing.TB`
   (real: `GinkgoTB()`), a fantasized `DataLinesSignalsGeneratedAt()`,
   `RequireSignalsContain` treated as substring match (it matches JSON
   keys), wrong-arity `RequireElementsContains`, and templ attribute
   escaping assumptions (`'`→`&#39;`, `"`→`&#34;`). Each cost a red
   cycle; all were avoidable by `grep`-ing the module cache BEFORE
   writing the call. The cheat-sheet helped but encoded signatures, not
   semantics.
2. **Wrote a spec that deadlocked the suite for 601s** (drop-counter
   test calling `Shutdown` with an undrained subscriber channel —
   go-sse's Shutdown waits for drain). Should have read
   `fanout.Shutdown` before writing the test; the timeout panic ate a
   full debug cycle.
3. **Edit tool raced the auto-commit daemon 4×** (known gotcha, and I
   still reached for `edit` first each time before falling back to
   python heredocs). Discipline for this repo: python-patch or
   full-file write for anything the daemon's gofmt touches.
4. **Tree left RED at the interruption point** (health.go split
   half-done: stale imports, missing `cacheDirForDB`). Repaired before
   this report — suites green — but the report-time fix should have been
   the commit-time fix.
5. **Masked pass:** the discovery suite had no bootstrap, so "ok … no
   tests to run" looked green while testing nothing. Caught only because
   I read the full output line.

## e) WHAT WE SHOULD IMPROVE

- **Read-the-source-first rule:** for any fleet library call not used in
  the last hour, grep the module cache for the exact signature AND its
  doc comment semantics before writing code (assertion helpers
  especially).
- **Encode this session's API facts** (GinkgoTB, RequireSignalsContain
  key-semantics, templ escaping, go-sse Shutdown drain) into the repo
  AGENTS.md when M17 writes docs — they are exactly the
  hard-to-discover-from-code facts AGENTS.md exists for.
- **Concurrency tests deserve a race pass the moment they're written**,
  not "in the final pass".
- Patch purity: rows patched via `FleetRow` currently embed
  `RelativeTime(AutoRefresh: true)`. The skill says the script is
  singleton-guarded (idempotent across re-renders), but go-health-dashboard
  deliberately renders a bare `<time data-tc-relative>` in patches —
  worth aligning (verify no script tag actually lands in patch frames).
- `CollectN` + poll-subscriber-count is a mild race pattern; a
  hub-level `AwaitSubscriber` test hook would remove the polling
  duplication AND the flake surface.
- Fixed port 18477 in the serve smoke test will collide under parallel
  CI jobs — allocate a listener port in the test instead.

## f) NEXT (ordered, ~40 items)

**M20 finish (1)**

~~1. F093: test non-localhost warning + unix-socket serve round-trip.~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit

**M17 docs (5)**
~~2. dprint.json + format check.~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit
~~3. TODO_LIST.md seeded from the plan (annotate plan, never rewrite).~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit
4. ~~FEATURES.md honest inventory (core + hub).~~ done — FEATURES.md
5. ~~DOMAIN_LANGUAGE.md core terms.~~ done — docs/DOMAIN_LANGUAGE.md
~~6. CHANGELOG.md Unreleased section.~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit

**M18 nix (4)**
~~7. flake.nix skeleton (both modules).~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit
~~8. devShell (go, templ, golangci, dprint).~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit
~~9. packages + vendorHash bootstrap loop.~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit
~~10. `nix flake check` green.~~ done at 2026-11-19: green at both wave boundaries and twice re-verified after the go-filewatcher and errorpage go.mod windows

**M19 CI (4)**
~~11. build+test per module (Forgejo+GitHub compatible).~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit
~~12. lint job pinned to devShell golangci version.~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit
~~13. drift-check + docs-gates jobs.~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit
~~14. workflow files validated on both runner flavors.~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit

**M21 examples (3)**
15. ~~examples/reporter/main.go.~~ done — examples/reporter/main.go
~~16. One-line integration guide.~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit
~~17. README screenshots (terminal table + WebUI).~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit

**M22 benchmarks (2)**
~~18. Benchmark fold+derive at 300 repos.~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit
~~19. Baseline under benchmarks/.~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit

**M23 pointers (2)**
~~20. BuildFlow TODO_LIST pointer row + link.~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit
~~21. STACK.md errata check.~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit

**M24 v2 sketch (2)**
~~22. v2 API design doc (diff vs v1).~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit
~~23. Compat policy: v1 emittable forever.~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit

**M25 federation (2)**
~~24. Research go-health-dashboard federation API.~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit
~~25. Interop options note in docs.~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit

**M26 (blocked on g2) (2)**
26. Endgame decision note. **→ TODO_LIST Blocked row — M26 overview endgame, owner question g2**
27. Migration sketch if hub replaces overview views. **→ TODO_LIST Blocked row — M26 overview endgame, owner question g2**

**M27 realtime loop (4)**
~~28. Fix `queryLatestRunCounts` N+1 (single query with window join).~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit
~~29. Re-derive on store writes (hook → hub.PublishRefresh).~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit
~~30. discovery.Watch → refresh trigger with debounce/coalesce.~~ done at 2026-11-19: runDiscoveryWatch kicks the refresher — cap-1 trigger channel (coalescing) + 500ms debounce + 5-min safety net (hub/internal/cli/refresher.go, serve.go)
~~31. E2E: evidence change → browser event.~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit

**Quality debts from this session (8)**
~~32. `go test -race` across hub (events/dashboard/cli concurrency).~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit
~~33. golangci-lint config + run for hub module (per-module parity).~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit
~~34. Shared `waitForSubscriber`/AwaitSubscriber test helper.~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit
~~35. Dynamic port in serve smoke test.~~ done at 2026-11-19: serve smoke binds a free dynamic port (hub/internal/cli/app_test.go:100 freePort())
~~36. Patch purity check for RelativeTime (align with dashboard pattern).~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit
37. ~~.gitignore templ-commit note.~~ done — .gitignore templ-commit note
38. ~~AGENTS.md: hub conventions learned this session.~~ done — AGENTS.md
~~39. gofmt/format gate for .templ sources in CI (templ fmt check).~~ done at 2026-11-19: `templ fmt -fail` CI gate job (templ-fmt; AGENTS.md row)

**Release hygiene (6)**
~~40. go mod tidy + GOWORK=off builds for both modules.~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit
41. Drop hub `replace` directive when core tags v0.1.0. **→ TODO_LIST Low row — hub replace-drop executes at the next core tag (v0.1.0 predates the Wave core splits)**
~~42. Fixture 341-repo pin relaxation before next refresh.~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit
~~43. Tag train plan (core first, hub second) per fleet release procedure.~~ done — core v0.1.0 tagged (77881b3); the hub tag rides the replace-drop TODO row
~~44. Update STACK.md hosting/deploy row with systemd unit.~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit
~~45. Final pass: whole-repo build+test + plan-vs-tree audit + this~~ done — shipped in the M-series build; tree-verified by the 15-00 final-pass audit
report's items harvested into TODO_LIST.

## g) QUESTIONS (cannot resolve myself)

~~1. **g1 — hub auth/bind endgame:** is socket/localhost the PERMANENT~~ done — bind posture decided (docs/decisions/bind-posture.md); auth parked as a v2-era design question
posture (no auth ever, tunnel for remote), or should an auth layer be
planned? M20's warning strictness and M26 both lean on this.
2. **g2 — overview endgame:** replace or coexist? Blocks M26 (migration **→ TODO_LIST Blocked row — M26 overview endgame, owner question g2**
sketch vs interop note emphasis).
3. **Forgejo visibility:** `lars/go-compliance` still needs your manual **→ TODO_LIST Blocked row — Forgejo visibility, owner-only UI click**
UI click to go public (no API token exists). Say the word and I'll
assume it stays private for now.

---

_Assisted-by: Crush:glm-5.3_

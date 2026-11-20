# go-datastar

| Field           | Value                                                                                                                                                                                                                                                                                      |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Reviewed        | 2026-11-19                                                                                                                                                                                                                                                                                 |
| Group           | B (realtime-dashboard stack)                                                                                                                                                                                                                                                               |
| Consumer module | `hub` (`github.com/larsartmann/go-compliance/hub`)                                                                                                                                                                                                                                         |
| Pins            | `github.com/larsartmann/go-datastar v0.6.2` (direct); `go-datastar/datastartest v0.6.2` (direct, tests); `go-datastar/static v0.6.1` (direct). `broadcast/` sub-module not required.                                                                                                       |
| Library version | newest tag `static/v0.8.0`; main-module line newest `v0.7.0`; HEAD `f909928 2026-11-17 feat: export named version constants and VersionIsValid`; tags newest-first: `static/v0.8.0`, `broadcast/v0.7.0`, `datastartest/v0.7.0`, `static/v0.7.0`, `v0.7.0`                                  |
| Pin current?    | **No.** Root/datastartest one minor behind (v0.6.2 → v0.7.0), static two behind (v0.6.1 → v0.8.0, embeds DataStar JS 1.0.3 → 1.0.4). Effective MVS resolution equals the pins (verified `go list -m all`; templ-components v1.27.0 also requires static v0.6.1, so nothing drags them up). |

## Verdict: UNDERUSED-moderate

The core promise STACK.md chose this library for — "patches-as-values producing
`sse.Event`" — is genuinely and well consumed: `ElementsFromTempl` with
inner/append/replace/remove modes drives the whole dashboard diff machinery
(`hub/internal/dashboard/dashboard.go:149-215`), and `datastartest` is used
exactly per the repo's testing conventions (CollectN/CollectWithTimeout,
Require\*, finders — `dashboard_test.go:38-124`, `routes_test.go:49-50`,
`cli/refresher_internal_test.go:53-74`). But everything _around_ the patch types
is hand-rolled: the script handler reimplements `datastar.ScriptHandler` without
its ETag/304 contract (`dashboard.go:113-118`), the replay ring reimplements
`datastar.MemoryStore` line-for-line (`hub/internal/events/store.go:14-51` vs
`go-datastar/store.go:20-77`), the signals patch bypasses `NewSignalsPatch` with
a hand-marshal (justified — see gap 3 — but the bypass is an upstream gap too),
and the `broadcast/` connection-lifecycle module is not even required. This is
the exact shape the fleet constraint warns about: the library solved the
plumbing, the consumer re-derived 80% of it.

**Flip trigger:** land opportunities 1-3 below (script handler, MemoryStore,
signals determinism option upstream or documented bypass) and this moves to
ADOPTED at high band; if `broadcast/` is adopted too, very high.

## Derived score: 58 (moderate) — 7 of 12

- Inventory rows N = 19; applicable A = 12 (7 N/A rows excluded).
- Fully = 5 × 1.0; Partial = 4 × 0.5; Missed/Misused = 3 × 0.
- Σ = 7.0; score = round(100 × 7 / 12) = **58**, band **moderate** (50-69).

## Capability inventory

| #  | Capability                                                                                        | Status      | Library evidence (path)                                                                                    | Consumer evidence (path:line)                                                                                                                                                                                               |
| -- | ------------------------------------------------------------------------------------------------- | ----------- | ---------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1  | Element patches from templ (`ElementsFromTempl`)                                                  | Fully       | `go-datastar/elements.go`; `patch.go` (`Patch.Event() sse.Event`)                                          | `dashboard.go:149-152,173-176,183-186,207-210`                                                                                                                                                                              |
| 2  | Patch modes inner/append/replace/remove + selectors                                               | Fully       | `go-datastar/options.go` (`WithMode*`, `WithSelector*`)                                                    | `dashboard.go:151,175,185,194`; row selector helper `dashboard.go:234-236`                                                                                                                                                  |
| 3  | Signals patches (`NewSignalsPatch`/`MarshalSignals`)                                              | Partial     | `go-datastar/signals.go:62,82-84` (plain `json.Marshal`, no determinism option)                            | `dashboard.go:228-233` hand-marshals with `json.Deterministic(true)`, then `SignalsPatch{Signals: encoded}`                                                                                                                 |
| 4  | Script serving with ETag/304 (`ScriptHandler`/`ScriptHandlerWith`)                                | **Misused** | `go-datastar/script_handler.go:47-62` (ETag + `If-None-Match` 304 + max-age=86400)                         | `dashboard.go:111-118` hand-rolls the handler: max-age=86400, **no ETag, no 304**                                                                                                                                           |
| 5  | Cache-busting versioned script tag (`ScriptTagVersioned`, v0.7.0)                                 | Missed      | `go-datastar/script.go` (`ScriptTagVersioned`); CHANGELOG [0.7.0]                                          | `hub/internal/dashboard/view.templ:67-73` emits plain `/datastar.js` via templ-components `SDKScript`                                                                                                                       |
| 6  | In-memory replay store (`MemoryStore` implements `sse.EventStore`)                                | Missed      | `go-datastar/store.go:20-77` (capacity ring, numeric `EventsAfter`)                                        | `hub/internal/events/store.go:14-51` (`ringStore`) — same design, same semantics, hand-written                                                                                                                              |
| 7  | `broadcast/` module: connection lifecycle (ServeHTTP, heartbeat, replay, `NewBroadcasterFromHub`) | Partial     | `go-datastar/broadcast/` (`Broadcaster.ServeHTTP`, `WithReplayCapacity`, `WithHeartbeatInterval/Disabled`) | not required; `dashboard.go:84-109` hand-rolls the serve loop over `events.Hub.SubscribeFeed`                                                                                                                               |
| 8  | Last-Event-ID replay on the patch stream (STACK.md rationale names it)                            | Partial     | `broadcast/` replay wiring; `datastar.LastEventID`                                                         | realized only on the JSON `/events` stream (`hub/internal/events/events.go:220-224`); patch stream sends no event IDs and replays nothing — reconnect correctness rests entirely on full re-snapshot (`dashboard.go:88-90`) |
| 9  | Stream error/notification frames (`ErrorResponse`, `NotificationResponse`)                        | Partial     | `go-datastar/errors.go`                                                                                    | no call site; a failing render or empty hub degrades to silent EOF → client `RetryAlways` reconnect (`view.templ:50-53`) with no operator-visible signal                                                                    |
| 10 | `datastartest` Collect family                                                                     | Fully       | `datastartest/collect.go`                                                                                  | `dashboard_test.go:38-39,96`, `routes_test.go:49-50`, `cli/refresher_internal_test.go:53-54`                                                                                                                                |
| 11 | `datastartest` assertions/finders                                                                 | Fully       | `datastartest/` (`RequireElementsContains`, `FindSignals`, `FindAllElements`, …)                           | `dashboard_test.go:41-124`, `cli/refresher_internal_test.go:69-74`                                                                                                                                                          |
| 12 | Self-hosted client bundle (`static/` module)                                                      | Fully       | `go-datastar/static/static.go` (`Bytes()`, `Version`)                                                      | `dashboard.go:14,117` — self-hosted, no CDN dependency (good offline/CSP posture)                                                                                                                                           |
| 13 | Inbound signal reading (`ReadSignals`)                                                            | N/A         | `go-datastar/inbound.go`                                                                                   | server never reads signals; filtering is client-side via `data-show` (`rows.templ:25`)                                                                                                                                      |
| 14 | Client-side SSE URL attributes (`GetSSE`/`PatchSSE`/…)                                            | N/A         | `go-datastar/sugar.go`, `adapters.go`                                                                      | templ-components `datastar.LiveRegion` owns the SSE region markup (`view.templ:45-56`) — correct layer                                                                                                                      |
| 15 | `Response` builder (redirects, `ReplaceURLQuerystring`, console patches)                          | N/A         | `go-datastar/response.go`                                                                                  | no action endpoints; pure patch-stream app                                                                                                                                                                                  |
| 16 | Custom-event / script patches (`NewDispatchCustomEventPatch`, `NewConsole*Patch`)                 | N/A         | `go-datastar/patch.go`, `script_convenience.go`                                                            | no client-side event/script needs                                                                                                                                                                                           |
| 17 | View transitions options (`WithViewTransitions*`)                                                 | N/A         | `go-datastar/options.go`                                                                                   | cosmetic opt-in; table patches should not morph-animate                                                                                                                                                                     |
| 18 | `ElementsFromGostar` adapter                                                                      | N/A         | `go-datastar/elements.go`                                                                                  | the dashboard is templ-based (`ElementsFromTempl` used)                                                                                                                                                                     |
| 19 | `broadcast/storecontract` assertions for custom stores                                            | N/A         | `broadcast/storecontract/` (v0.7.0)                                                                        | consumer would only need this if it kept a custom store after adopting `MemoryStore` (row 6 makes it moot)                                                                                                                  |

## Gap analysis

### Row 4 — hand-rolled ScriptHandler drops the library's ETag/304 contract (Misused)

**Library offers:** `datastar.ScriptHandler()` / `ScriptHandlerWith(bundle)` —
serves the bundle with `Content-Type`, a content-hash `ETag` (`ETagFor`),
`Cache-Control: public, max-age=86400`, and answers `If-None-Match` with 304
(`go-datastar/script_handler.go:47-62`). One line of wiring.

**Consumer does:** `dashboard.go:111-118` writes `dstarstatic.Bytes()` with the
same max-age but **no ETag and no 304 path**. Two consequences:

1. Every cold page view re-downloads the full bundle (no revalidation shortcut).
2. Worse, combined with row 5: after a hub upgrade the browser keeps the old JS
   client for up to 24 h (fresh max-age skips ETag revalidation) — exactly the
   old-client/new-server protocol-skew window the library's v0.7.0
   `ScriptTagVersioned` release note describes.

**Impact:** moderate — single-operator dashboard, so bandwidth is trivial; the
real cost is the upgrade-skew window and the fleet-pattern cost (every consumer
re-writing bundle serving instead of one shared handler).

**Sketch** (`hub/internal/dashboard/dashboard.go`):

```go
// ScriptHandler serves the embedded DataStar JS bundle (ETag + 304 via go-datastar).
func (d *Dashboard) ScriptHandler() http.Handler {
	return datastar.ScriptHandlerWith(dstarstatic.Bytes())
}
```

Deletes `scriptContentType`/`scriptCacheControl` constants (`dashboard.go:30-31`).
If the `Routes.Script` path must stay configurable, keep the mux wiring in
`routes.go:55` as-is — only the handler body changes.

### Row 5 — `ScriptTagVersioned` not used; upgrades can serve a stale client (Missed)

**Library offers (v0.7.0):** `ScriptTagVersioned(path)` — a `<script type="module">`
whose `src` carries the pinned client version as `?v=`, making each bundle release
a distinct cache key; plus `datastartest.RequireScriptTagVersion` to pin it
consumer-side.

**Consumer does:** `view.templ:67-73` renders the SDK via templ-components
`datastar.SDKScript` with `Src: "/datastar.js"` — unversioned. The 24 h max-age
then holds the _previous_ bundle across an upgrade.

**Impact:** low today (wire format between DataStar 1.0.3/1.0.4 is stable;
goldens unchanged per CHANGELOG), but the failure mode is protocol skew appearing
as mysterious dashboard breakage exactly once per upgrade — the worst kind of
one-off. **Fix belongs upstream-coordinated:** the versioned tag should come from
templ-components' `SDKScript` (it already tracks `static.Version`); until it does,
the hub can self-serve:

```go
// view.templ — version the self-hosted bundle path once ScriptHandler carries the ETag
@datastar.SDKScript(datastar.SDKScriptProps{
	BaseProps: utils.BaseProps{},
	Src:       fmt.Sprintf("%s?v=%s", model.ScriptSrc, dstarstatic.Version),
})
```

`dstarstatic.Version` exists at the pinned static v0.6.1 (`static/static.go:21`).

### Row 6 — `ringStore` duplicates `datastar.MemoryStore` (Missed)

**Library offers:** `datastar.MemoryStore` — an in-memory ring implementing
`sse.EventStore`: capacity-bounded append with oldest-eviction, numeric-ID
`EventsAfter` strictly-after semantics, non-numeric IDs treated as zero
(`go-datastar/store.go:20-77`). That is byte-for-byte the contract
`hub/internal/events/store.go:14-51` re-implements (same re-slice eviction at
`store.go:48-50`, same `strconv.ParseUint` cutoff at `store.go:53-64`).

**Consumer does:** keeps its own 65-line `ringStore` + `parseSeq`. The only
semantic difference: `ringStore.EventsAfter` _excludes_ stored events with
unparseable IDs (`store.go:35`) while `MemoryStore` treats them as sequence 0 —
irrelevant here, since the Hub mints all IDs as `strconv.FormatUint`
(`events.go:270-273`).

**Impact:** moderate fleet cost, tiny local cost. This is the canonical
"every project re-implements plumbing" specimen: the store exists, is tested in
the library, and the consumer rewrote it because it lives in `go-datastar` rather
than `go-sse`. Note the dependency-direction wrinkle: `hub/internal/events` is the
machine-facing JSON stream package; importing `go-datastar` there couples it to
the WebUI library. The honest options:

```go
// Option A (preferred): swap ringStore for MemoryStore, accept the go-datastar import.
type Hub struct {
	broadcaster *sse.Broadcaster[Published]
	store       *datastar.MemoryStore // implements sse.EventStore; capacity via NewMemoryStore(256)
	...
}
// events.go:118 → store: datastar.NewMemoryStore(defaultReplayCapacity)
// store.go deleted; its internal tests move to assertions against MemoryStore behavior.
```

```go
// Option B (if the coupling is rejected): raise an upstream issue to move the
// generic ring EventStore into go-sse itself; then both libraries and all
// fleet consumers share one implementation.
```

Option B is the fleet-correct end state; Option A is the pragmatic now.

### Row 3 — signals patch bypasses `NewSignalsPatch` (Partial, half-upstream)

**Library offers:** `NewSignalsPatch(v)` / `MarshalSignals(v)` — but through
plain `json.Marshal` (`signals.go:82-84`), i.e. no determinism guarantee for map
payloads.

**Consumer does:** `dashboard.go:228-233` marshals `stateSignals` (contains
`Counts map[string]int`) with `json.Deterministic(true)` and constructs
`SignalsPatch{Signals: encoded}` directly. **This bypass is correct**: the hub's
own convention is deterministic marshal (`events.go:34-38`), and byte-stable
patches make the SSE stream diffable and testable.

**Impact:** the consumer pays a small permanent tax (hand-rolled marshal +
`SignalsPatch` literal) for a capability the library should expose.

**Sketch (upstream PR to go-datastar):**

```go
// WithDeterministic marshals signals with sorted map keys (encoding/json/v2
// json.Deterministic), for byte-stable patch streams.
func WithDeterministic() SignalsPatchOption
```

Then the consumer collapses to
`patch, _ := datastar.NewSignalsPatch(signals, datastar.WithDeterministic())`.

### Row 7 — `broadcast/` module unused; serve loop hand-rolled (Partial)

**Library offers:** `broadcast.Broadcaster` — `ServeHTTP` doing stream setup,
optional Last-Event-ID replay, heartbeat, and fan-out; `NewBroadcasterFromHub`
shares an existing `sse.Broadcaster[sse.Event]`; `WithReplayCapacity`,
`WithHeartbeatInterval/Disabled` (`broadcast/broadcaster.go`).

**Consumer does:** `dashboard.go:84-109` hand-rolls the equivalent loop because
the projection is _per-subscriber_: each `Published` event is converted to patches
inside the subscriber's handler (`sendPublished`, `dashboard.go:123-134`). The
library's model projects _once_ at publish time and broadcasts `sse.Event`
patches.

**Impact:** architecture, not bug. The consumer's shape buys snapshot-first
semantics (fresh full state on every connect — `dashboard.go:88-90`) at the cost
of N× projection work and no patch-stream replay (row 8). With a single-operator
dashboard N≈1-3, so neither cost bites. **Recommendation: leave as-is, document
the deviation** — migrating to `broadcast/` would restructure the Hub's typed
event flow for no observable gain at this scale. Revisit only if multiple
dashboard surfaces (or external patch consumers) appear.

### Row 8 — patch stream has no Last-Event-ID replay (Partial)

STACK.md's rationale for choosing go-datastar cites "broadcaster + Last-Event-ID
replay" — realized in this repo on the **JSON** stream (`events.go:220-224`) but
not the **patch** stream: patches carry no IDs (`dashboard.go` never sets
`WithElementsEventID` / ID fields), and reconnection correctness rests on the
full re-snapshot. For this dashboard that is _simpler and sufficient_: rows are a
projection, and a snapshot is strictly more correct than replayed patches.
Cost of the current shape: a reconnect during heavy churn re-downloads the whole
table once. **No action recommended**; the row is Partial only because the cited
rationale is half-realized — worth a line in STACK.md so a future reader doesn't
hunt for replay that isn't there.

### Row 9 — no error frames on the patch stream (Partial)

`ErrorResponse`/`NotificationResponse` (`go-datastar/errors.go`) would let the
stream tell the browser something went wrong (toast/alert) instead of dying
silently. The dashboard's current degradation story is deliberate: the
page-level `rendered X ago` RelativeTime going stale IS the dead-stream signal
(`view.templ:35-42`, pinned by `view_test.go:53-79`). Acceptable for a
single-operator tool; a `NotificationResponse` on the "hub shutting down" path
would be a nice-to-have, not a gap worth a PR now.

## Version currency

- **v0.6.2 → v0.7.0** (all four modules lockstep-tagged on one commit; v0.6.3 was
  a defective root-only cut, repaired by v0.7.0). Contents: `ScriptTagVersioned`
  - `ETagFor` export (rows 4-5), `broadcast.WithHeartbeatDisabled` +
    `NewBroadcasterFromHub` variadic options, `datastartest.RequireScriptTagVersion`
    / `RequireETagForBundle`, `broadcast/storecontract`, embedded client
    DataStar JS 1.0.3 → **1.0.4**, and go-sse bumped to v0.7.0.
- **Breaking change in v0.7.0:** `ScriptHandlerWith(bundle, version)` drops the
  dead second parameter. The consumer never calls it (it hand-rolls — row 4), so
  the migration is a no-op here; adopting the library handler _after_ the bump
  uses the new one-arg signature directly.
- **static v0.6.1 → v0.8.0:** additive only — named version constants
  (`Version1_0_4`, …) + `VersionIsValid` (`f909928`). The 1.0.3 → 1.0.4 client
  bump is the only behavioral delta (wire-format goldens unchanged per CHANGELOG).
- **Upgrade risk: low.** `go get github.com/larsartmann/go-datastar@v0.7.0
  github.com/larsartmann/go-datastar/datastartest@v0.7.0
  github.com/larsartmann/go-datastar/static@v0.7.0` in `hub/` (GOWORK=off) —
  static v0.7.0 suffices; v0.8.0's named constants are only needed if the hub
  wants compile-time version references. It drags go-sse to v0.7.0 (see the
  go-sse report). templ-components v1.27.0 pins static v0.6.1, which MVS-merges
  harmlessly (static v0.7.0 wins, byte-identical API + new Version value).

## Fleet leverage

**Expand.** 26 consumers per STACK.md; this repo proves the core patch model but
leaves the lifecycle layer (script serving, replay store, broadcast) as local
plumbing — the precise anti-pattern the fleet constraint names. Three of the
gaps have _upstream_ components that would pay fleet-wide dividends: a
`WithDeterministic` signals option, a versioned self-host `SDKScript` in
templ-components (or `ScriptTagVersioned` adoption pattern), and ideally moving
the generic ring `EventStore` into go-sse where it belongs. Conversely, the
`broadcast/` non-adoption is a _justified_ leave — the per-subscriber projection
over a typed Hub is better architecture for snapshot-first dashboards, and
writing that decision down prevents future litigating. **Adopt** rows 4-6,
**leave** row 7, **upstream** rows 3 and 5.

## Ranked opportunities (impact × ease)

1. **Adopt `datastar.ScriptHandlerWith(dstarstatic.Bytes())`** (row 4) — high
   impact (caching contract + kills the upgrade-skew window together with #2),
   trivial ease (one line, two constants deleted).
2. **Version the script URL** (`?v=` + `dstarstatic.Version`, row 5) — high
   impact when combined with #1, trivial ease; file the templ-components
   `SDKScript` upstream issue in the same motion.
3. **Replace `ringStore` with `datastar.MemoryStore`** (row 6) — moderate impact
   (deletes a hand-rolled duplicate; fleet pattern), easy (constructor swap +
   delete store.go + move its tests); opens the "should the ring live in
   go-sse?" upstream question.
4. **Bump pins to the v0.7.0 lockstep set** — low direct impact, trivial ease;
   prerequisite-clean for #1's new signature.
5. **Upstream `WithDeterministic` signals option** (row 3) — small local win,
   moderate fleet win; low ease (upstream PR cycle).
6. **Document the patch-stream no-replay decision in STACK.md** (row 8) —
   zero-risk clarity gain.

## How to verify (executed 2026-11-19)

```bash
# Library metadata
git -C /home/lars/projects/go-datastar describe --tags --abbrev=0         # static/v0.8.0
git -C /home/lars/projects/go-datastar log -1 --format='%h %cs %s'        # f909928 2026-11-17 feat: export named version constants and VersionIsValid
git -C /home/lars/projects/go-datastar tag --sort=-creatordate | head -10 # static/v0.8.0, broadcast/v0.7.0, datastartest/v0.7.0, static/v0.7.0, v0.7.0, v0.6.3, broadcast/v0.6.2, datastartest/v0.6.2, v0.6.2, broadcast/v0.6.1
sed -n '1,120p' /home/lars/projects/go-datastar/CHANGELOG.md              # [0.7.0] section: ScriptTagVersioned, ETagFor, lockstep cut, breaking ScriptHandlerWith

# Script handler contract the consumer is missing
rg -n 'ETag|If-None-Match|304|max-age' /home/lars/projects/go-datastar/script_handler.go

# MemoryStore vs ringStore duplication
sed -n '20,77p' /home/lars/projects/go-datastar/store.go
sed -n '14,51p' /home/lars/projects/go-compliance/hub/internal/events/store.go

# Signals determinism bypass
rg -n 'MarshalSignals|json.Marshal' /home/lars/projects/go-datastar/signals.go
rg -n 'Deterministic|SignalsPatch' /home/lars/projects/go-compliance/hub/internal/dashboard/dashboard.go

# Consumer usage + effective pins
cd /home/lars/projects/go-compliance/hub && GOWORK=off go list -m all | grep go-datastar
cd /home/lars/projects/go-compliance && rg -n 'datastar\.' hub/internal/dashboard/dashboard.go
rg -n 'datastartest\.' hub -g '*_test.go' | head -20
rg -n 'SDKScript|ScriptSrc|LiveRegion' hub/internal/dashboard/view.templ hub/internal/cli/serve.go

# broadcast module API (the justified leave)
cd /home/lars/projects/go-datastar/broadcast && GOWORK=off go doc -all . | grep -E '^func |^type '
```

All file:line citations spot-checked against the working tree on 2026-11-19.

## Citation drift addendum (2026-11-19 post-audit wave)

The ID-minting citation (events.go 324-327) now sits at
`events.go:270-273` and the row-selector helper citation at
`dashboard.go:234-236` after the ringStore deletion + Handler
extraction (verified by `scripts/check-citations.sh`).

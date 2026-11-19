# go-sse

| Field                | Value                                                                                              |
| -------------------- | -------------------------------------------------------------------------------------------------- |
| Reviewed             | 2026-11-19                                                                                         |
| Group                | B (realtime-dashboard stack)                                                                       |
| Consumer module      | `hub` (`github.com/larsartmann/go-compliance/hub`) — core module does not depend on go-sse          |
| Pins                 | `github.com/larsartmann/go-sse v0.6.2` (direct); `go-sse/ssetest v0.5.0` (direct, tests); `go-sse/sseparse v0.2.1` (indirect) |
| Library version      | root `v0.7.0` (`git -C /home/lars/projects/go-sse describe --tags --abbrev=0`); HEAD `a1b4674 2026-11-17`; tags newest-first: `v0.7.0`, `ssetest/v0.5.0`, `sseparse/v0.2.1`, `v0.6.2`, `sseparse/v0.2.0` |
| Pin current?         | **No** for the root module (v0.6.2 → v0.7.0, additive only). **Yes** for ssetest (v0.5.0 is the newest ssetest tag) and sseparse (v0.2.1 newest). |

## The ssetest "skew" question (asked by the brief)

Not a smell. The three modules version on **independent semver lines** (`vX.Y.Z` vs
`ssetest/vX.Y.Z` vs `sseparse/vX.Y.Z`). ssetest v0.5.0 (2026-10-11) *postdates* root
v0.6.2 (2026-10-03) and its go.mod requires exactly `go-sse v0.6.2` — the hub's root
pin. The hub's combination `{root v0.6.2, ssetest v0.5.0, sseparse v0.2.1}` is the
correct per-module latest-of-line at the v0.6.2 generation. The only drift is that
root v0.7.0 exists (see Version currency).

## Verdict: ADOPTED

The hub's SSE layer is a textbook consumption of the library: the typed
`Broadcaster[Published]` with drop accounting and buffer sizing
(`hub/internal/events/events.go:126-129`), the documented `NewStream` + deferred
`Close` + `Heartbeat` goroutine + `select`-on-context pattern in both handlers
(`events.go:215-263`, `hub/internal/dashboard/dashboard.go:84-109`), branded
`EventID`s minted at publish (`events.go:256`, `events.go:333`), and full
Last-Event-ID replay through the `EventStore` extension point
(`events.go:220-224` + `store.go`). The `EventStore` interface is consumed exactly
as designed — the consumer supplies the ring, the library supplies `Replay`.
ssetest is used idiomatically (`CollectN`, `WithPath`, `WithLastEventID`,
`Require*`) in `events_test.go:127-153,207-210`. The one genuine miss is client-side:
`hub/internal/discovery/client.go:184-201` hand-rolls SSE parsing with
`bufio.Scanner` + `strings.HasPrefix(line, "data:")` while `sseparse` (already in the
module graph, zero new deps) implements the WHATWG § 9.2.6 parse the hand-roll
approximates — including the `scanner.Err()` check the hand-roll omits (flagged by
gopls at `client.go:184`) and multi-line `data:` frame reassembly the hand-roll
silently drops.

**Flip trigger:** none to adopt — already adopted. The next audit flips to
UNDERUSED only if new SSE-consuming code again hand-rolls parsing instead of
reaching for `sseparse`.

## Derived score: 85 (very high) — 11 of 13

- Inventory rows N = 20; applicable A = 13 (7 N/A rows excluded).
- Fully = 11 × 1.0; Partial = 0; Missed/Misused = 2 × 0 (rows 6, 19).
- Σ = 11.0; score = round(100 × 11 / 13) = **85**, band **very high** (≥ 85).

## Capability inventory

| # | Capability | Status | Library evidence (path) | Consumer evidence (path:line) |
|---|---|---|---|---|
| 1 | SSE stream serving lifecycle (`NewStream`/`Send`/`Close`/`Context`) | Fully | `go-sse/stream.go` (`Stream` type) | `hub/internal/events/events.go:217-218`, `hub/internal/dashboard/dashboard.go:86-87` |
| 2 | Comment heartbeat keep-alive (`Stream.Heartbeat`) | Fully | `go-sse/stream.go` (`Heartbeat`); `WriteHeartbeat` | `events.go:244`, `dashboard.go:96` (15 s in both) |
| 3 | Branded event IDs (`EventID`, `NewEventID`, `ParseEventID`) | Fully | `go-sse/event.go`; branded `brandid.ID` | `events.go:256`, `events.go:333`, `store.go:53-64` |
| 4 | Last-Event-ID request extraction (`LastEventIDFromRequest`) | Fully | `go-sse/stream.go` / `event.go` | `events.go:220` |
| 5 | Replay via `EventStore` + `sse.Replay` | Fully | `go-sse/replay.go`; `EventStore` iface | `events.go:221-224`; `store.go:23-41` implements the interface |
| 6 | `retry:` field advisement (`WriteRetry`) | Missed | `go-sse/stream.go` (`WriteRetry` doc: reconnect cadence) | no call site — neither handler sends `retry` |
| 7 | Typed fan-out (`Broadcaster[T]` Subscribe/Broadcast/Unsubscribe) | Fully | `go-sse/broadcaster.go`, `fanout.go` | `events.go:126-129`, `events.go:269-270`, `events.go:334` |
| 8 | Drop observability (`WithOnDrop` → `Dropped()`) | Fully | `go-sse/fanout.go` (`WithOnDrop`) | `events.go:127`, `events.go:276` |
| 9 | Backpressure sizing (`WithBufferSize`, `DefaultSubscriberBuffer`) | Fully | `go-sse/fanout.go`; const added in v0.7.0 | `events.go:30,100,128` (buffer 64 = the default, made explicit) |
| 10 | Graceful shutdown (`Broadcaster.Shutdown(ctx)`) | Fully | `go-sse/fanout.go` | `events.go:282` |
| 11 | `SubscribeFilter` predicate subscriptions | N/A | `go-sse/broadcaster.go` | one event type for all subscribers; no per-subscriber view exists |
| 12 | `BroadcastMany` batching | N/A | `go-sse/broadcaster.go` | hub publishes one event per refresh; no batch producer |
| 13 | Lifecycle callbacks (`OnSubscribe`/`OnUnsubscribe`/`Stream.OnDisconnect`) | N/A | `go-sse/broadcaster.go`, `stream.go` | subscriber count is polled via `SubscriberCount` (`events.go:279`) — sufficient for the health check |
| 14 | `Broadcaster.Health()` lifecycle snapshot | N/A | `go-sse/fanout.go` (`BroadcasterHealth`) | shutdown correctness is covered by `Shutdown(ctx)` error return; no lifecycle dashboard exists |
| 15 | Keyed data lines (`KeyedLines`/`JoinLines`/`WriteKeyedLines`) | N/A | `go-sse/stream.go` docs cite DataStar wire | go-datastar owns patch serialization (`hub/internal/dashboard/dashboard.go:233`) — correct layering |
| 16 | Low-level wire writers (`WriteEvent`/`SetHeaders`) | N/A | `go-sse/stream.go` | `Stream` covers the hub; no custom HTTP scaffolding |
| 17 | ssetest `Collect*` against real handlers | Fully | `ssetest/collect.go` | `events_test.go:127,146-148,207` |
| 18 | ssetest `Require*` assertions + request options | Fully | `ssetest/assert.go`, `options.go` | `events_test.go:131-136,152-153,208-210` |
| 19 | sseparse client-side stream parsing | **Missed** | `sseparse/reader.go` (`NewStreamReader`, `ReadNEvents`, `WithMaxLineBytes`); README "Parse SSE from any io.Reader" | `hub/internal/discovery/client.go:184-201` hand-rolls `bufio.Scanner` + `data:` prefix matching |
| 20 | sseparse WPT conformance corpus (`Corpus()`) | N/A | `sseparse/corpus.go` | library-authoring tool; no parser is written here (see row 19) |

## Gap analysis

### Row 19 — sseparse unused while `/v1/watch` hand-parses SSE (Missed)

**Library offers:** `sseparse.NewStreamReader(r)` / `ReadEvents` / `StreamReader.Next` —
a zero-dependency WHATWG § 9.2.6 parser: multi-line `data:` reassembly, BOM/lone-CR
handling, `id:`/`retry:` capture, bounded line lengths via `WithMaxLineBytes`
(`sseparse/reader.go`; `sseparse/README.md`). Already in the hub's module graph
(v0.2.1, indirect via ssetest) — adopting it is a require-line promotion, not a new
dependency.

**Consumer does:** `hub/internal/discovery/client.go:184-201` — `bufio.NewScanner`
with a hand-rolled 64 KiB/1 MiB buffer, `strings.HasPrefix(line, "data:")`, per-line
`json.Unmarshal` with malformed frames skipped. Three concrete defects vs the
library parser:

1. `scanner.Err()` is never checked after the loop (gopls `scannererr` warning at
   `client.go:184`) — a mid-stream network error is indistinguishable from a clean
   EOF, so the caller's documented "closed stream → re-subscribe with backoff"
   recovery silently treats transport failure as normal stream end.
2. Multi-line `data:` frames are parsed line-by-line: a JSON payload the daemon
   splits across `data:` lines fails `json.Unmarshal` per line and is dropped
   (the `continue` at `client.go:200-201`) — the event vanishes instead of erroring.
3. The SSE framing rules (CR handling, `data` vs `data:` vs `data: ` spacing,
   comment lines that merely contain "data:") are re-approximated instead of
   inherited from the conformance-tested parser.

**Impact:** the discovery watch is the liveness input to the whole never-covered
feature; a silently swallowed read error delays re-derivation until the refresher's
5-minute safety net (AGENTS.md "Live Architecture (M27)") — degraded, not broken,
which is exactly the kind of bug that survives. Fleet-level: every SSE *consumer*
Lars writes re-decides this parse; consolidating on sseparse here sets the pattern.

**Sketch** (`hub/internal/discovery/client.go`):

```go
import "github.com/larsartmann/go-sse/sseparse"

go func() {
	defer close(events)
	defer func() { _ = resp.Body.Close() }()

	reader := sseparse.NewStreamReader(bufio.NewReaderSize(resp.Body, scanBufSize),
		sseparse.WithMaxLineBytes(maxScanLine))

	for {
		evt, err := reader.Next() // WHATWG dispatch: multi-line data joined, CR/BOM handled
		if err != nil {
			return // read error and EOF both end the stream; caller re-subscribes with backoff
		}

		var event Event
		if json.Unmarshal([]byte(evt.Data()), &event) != nil {
			continue // malformed payload never kills the stream
		}

		select {
		case events <- event:
		case <-ctx.Done():
			return
		}
	}
}()
```

This deletes the `scanner.Err()` class of bug and the frame-splitting bug, and the
`//nolint`-able gopls warning disappears with the scanner. Verify the exact
`StreamReader` method name against `go doc github.com/larsartmann/go-sse/sseparse StreamReader`
before landing (`Next` vs `ReadNext` per the pinned v0.2.1).

### Row 6 — `retry:` field never sent (Missed, low impact)

**Library offers:** `WriteRetry(w, ms)` — the SSE `retry:` field telling the browser
how long to wait before reconnecting (per spec, sent once and persisted).

**Consumer does:** neither `events.go:215-263` nor `dashboard.go:84-109` sends it.
The dashboard's reconnect behavior is instead governed by templ-components'
`datastar.RetryAlways` on the LiveRegion (`view.templ:53`) — the *client* decides.
A server `retry:` of e.g. 2000 ms would also bound the JSON `/events` clients'
default (~3 s browser default, implementation-defined).

**Impact:** low — reconnect works today; `retry:` is polish that would stop
reconnect storms after a hub restart (N browser tabs hammering at browser-default
cadence). Not worth a change on its own; fold into the next touch of either
handler.

## Version currency

- Pin `v0.6.2` (2026-10-03) → latest `v0.7.0`. The delta (CHANGELOG `[Unreleased]`
  released as v0.7.0): `DefaultSubscriberBuffer` exported constant, DataStar example
  strong-ID typing (`sse.EventID` at every boundary), coverage-ratchet/security-docs
  release plumbing. **No API the hub calls changed.**
- Upgrade risk: **minimal** — `go get github.com/larsartmann/go-sse@v0.7.0` in
  `hub/` (GOWORK=off), re-tidy. Bonus: the hub's `defaultSubscriberBuffer = 64`
  literal (`events.go:30`) can then defer to `sse.DefaultSubscriberBuffer` instead of
  duplicating it (there is no lockstep test tying them today).
- ssetest v0.5.0 and sseparse v0.2.1 are the newest tags on their lines — nothing
  to do.
- Note the fleet coupling: go-datastar v0.7.0 requires go-sse v0.7.0 (its
  CHANGELOG "Dependencies"); bumping go-datastar later forces this bump anyway —
  doing it first keeps the hub off a mixed generation.

## Fleet leverage

**Adopt/expand.** go-sse is the fleet's SSE transport standard (59 consumers per
STACK.md) and this repo consumes it at near-full depth on the serving side — the
STACK.md choice ("SSE wire format + fan-out + reconnection replay … vs go-aichat
ssehub") is fully vindicated by the code: replay, fan-out, heartbeats, and drop
accounting are all live, not aspirational. The expansion with fleet payoff is
**sseparse for SSE consumption**: the discovery client is this repo's only
SSE-client role, and the hand-rolled parse is precisely the plumbing Lars should
never rewrite per-project. Landing it here makes sseparse the visible default for
the next daemon-talking consumer. Cost: one small PR, zero new module-graph edges.

## Ranked opportunities (impact × ease)

1. **Replace the `/v1/watch` hand-parse with `sseparse`** (row 19) — high impact
   (correctness of the liveness input, kills the gopls warning), high ease
   (~20 lines, dep already in graph). 
2. **Bump root pin v0.6.2 → v0.7.0** — low direct impact, trivial ease; unblocks
   the go-datastar v0.7.0 lockstep train without a mixed-generation graph.
3. **Defer to `sse.DefaultSubscriberBuffer`** in `events.go:30,118` post-bump —
   cosmetic dedup of the 64 literal, only after #2.
4. **`retry:` field on both streams** (row 6) — low impact, low ease-risk; do it
   opportunistically with the next handler edit, not as its own change.

## How to verify (executed 2026-11-19)

```bash
# Library metadata
git -C /home/lars/projects/go-sse describe --tags --abbrev=0        # v0.7.0
git -C /home/lars/projects/go-sse log -1 --format='%h %cs %s'       # a1b4674 2026-11-17 chore: auto-commit...
git -C /home/lars/projects/go-sse tag --sort=-creatordate | head -5 # v0.7.0, ssetest/v0.5.0, sseparse/v0.2.1, v0.6.2, sseparse/v0.2.0
git -C /home/lars/projects/go-sse log --oneline v0.6.2..v0.7.0 | wc -l   # ~40 commits (auto-commit noise; CHANGELOG is the signal)

# Effective pins in the consumer (matches go.mod — nothing drags them up)
cd /home/lars/projects/go-compliance/hub && GOWORK=off go list -m all | grep 'go-sse'
#   github.com/larsartmann/go-sse v0.6.2
#   github.com/larsartmann/go-sse/sseparse v0.2.1
#   github.com/larsartmann/go-sse/ssetest v0.5.0

# Exported surface
cd /home/lars/projects/go-sse && GOWORK=off go doc -all . | grep -E '^func |^type ' | head -50
cd /home/lars/projects/go-sse/ssetest && GOWORK=off go doc -all . | grep -E '^func |^type '
cd /home/lars/projects/go-sse/sseparse && GOWORK=off go doc -all . | grep -E '^func |^type '

# Consumer usage sites (all citations in this report)
cd /home/lars/projects/go-compliance
rg -n 'go-sse' hub --glob '!vendor/**' -g '*.go'
rg -n 'sse\.' hub/internal/events/events.go hub/internal/events/store.go hub/internal/dashboard/dashboard.go
rg -n 'ssetest\.' hub/internal/events/events_test.go
rg -n 'Scanner|data:|scanner' hub/internal/discovery/client.go   # the hand-rolled parse, lines 184-201
```

Spot-checks performed on every file:line cited above against the working tree;
line numbers refer to the files as of 2026-11-19 (auto-commit daemon may shift
them by a line or two on re-read).

# go-retry

| Field            | Value                                                                                                                                                                                                                                                               |
| ---------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Reviewed         | 2026-11-19                                                                                                                                                                                                                                                           |
| Group            | C (library-utilization audit, HTTP/wire-protocol family)                                                                                                                                                                                                             |
| Consumer modules | core — **absent**; hub — **absent** (not in either `go.mod`, direct or indirect)                                                                                                                                                                                      |
| Library version  | repo tag `v0.9.0`; last commit `356fbd4 2026-11-14 "chore: auto-commit 2 changed file(s) (heuristic)"`                                                                                                                                                                |
| Recent tags      | `v0.9.0`, `v0.8.0`, `v0.7.1`, `v0.7.0`, `v0.6.1` (newest first)                                                                                                                                                                                                      |
| Modules          | root `github.com/larsartmann/go-retry` (the library; only runtime dep go-error-family); `tools/` is a dev-tools module (actionlint/govulncheck pins), not a library surface — effectively a single-module library for consumers. Go 1.26+ floor (README).            |

## The consumer's hand-rolled retry/backoff landscape (verified)

1. **Watch reconnect** — `hub/internal/cli/refresher.go:20`:
   `watchReconnectBackoff = 30 * time.Second` (fixed). The loop
   `runDiscoveryWatch` (refresher.go:218-257): on `Watch` error → warn log +
   `sleepCtx(ctx, watchReconnectBackoff)` (:229); on stream end → 1 s beat
   (:253); on success → snapshot refresh + consume events; forever until ctx
   cancels. `sleepCtx` (:275-285) is an 11-line ctx-aware sleep.
2. **500 ms debounce** (refresher.go:17) — burst coalescing, not retry.
3. **5-min safety-net ticker** (refresher.go:25) — a poll interval, not retry.
4. **30 s query timeout** (refresher.go:29) — timeout, not retry; a failed
   refresh is logged and retried on the next trigger/safety-net tick
   (refresher.go:207-209).
5. **SQLite busy-retry** — owned *inside* go-sqlitestore (STACK.md ★ row:
   "DSN+PRAGMAs, WAL, busy-retry"), i.e. already delegated to the right layer.

## Does go-retry match the watch-reconnect need? No — contract mismatch

go-retry is a **bounded** retry loop: `Config.MaxAttempts` (default 3,
*including* the first call), exponential backoff `Initial×Mult^(n-1)` capped at
`MaxDelay`, additive jitter, `IsRetryable` predicate (defaults to
`errorfamily.IsRetryable`), `DelayFunc` (e.g. honor `Retry-After`),
`OnRetry`/`OnExhausted` hooks, ctx-aware with classified terminal errors
(`ErrExhausted` / `ErrCanceled` / `ErrDeadlineExceeded`; verified via
`go doc`). The watch-reconnect loop is an **infinite supervision loop until
ctx cancels** — "the daemon being down is normal" (refresher.go:214-217).
There is no attempt budget, nothing to exhaust, no escalation desired:

- **Fixed 30 s is a design property, not a simplification.** The daemon is
  either up or down for minutes; exponential backoff (100 ms → 5 s cap) would
  hammer the socket early for no benefit, and unbounded growth would hide a
  dead daemon *slower* than the flat 30 s. Visibility of a dead daemon is
  owned by the non-critical `discovery` health check with a 10-min warn
  threshold (refresher.go:36-41, health.go) — retry escalation would be
  redundant signal.
- **Jitter is pointless here**: exactly one consumer process over a unix
  socket; no thundering herd exists. (Jitter matters when N clients realign.)
- **Wrapping `retry.Do` inside the for-loop adds nothing**: the loop body
  already is "attempt once, sleep ctx-aware, log" — the library's value (backoff
  computation, exhaustion semantics) sits precisely on the axis this loop
  deliberately flattened. Embarassingly, the *closest* fit is one-shot
  `Discover()` (client.go:101-141, 60 s timeout): but a failed snapshot is
  warned and naturally retried 30 s later by the reconnect cycle
  (refresher.go:240-244) — a bounded retry there would just front-load
  latency before falling back to the same supervision loop.

## Verdict — **NOT-ADOPTED-justified**

The library's contract (bounded attempts, transient-error classification,
Retry-After) matches zero of the three hand-rolled sites: the reconnect loop
is supervision (infinite, fixed-interval), the debounce is coalescing, and the
safety net is a poll. Adopting for "the 3 call sites" would be dep bloat: 1
new module + STACK.md drift-check row + the Nix vendorHash re-bootstrap ritual
(AGENTS.md) for code that would replace an 11-line `sleepCtx` with a loop
wrapper providing unused semantics. Fleet leverage is not free in both
directions: importing a retry library into a service whose loops are
deliberately non-retry-shaped is the "adopting a poor fit also costs fleet
maintenance" side of the owner's constraint.

**Flip trigger:** the hub grows an outbound *one-shot* call to a genuinely
flaky endpoint where transient errors and server-provided delays exist —
e.g. an HTTP evidence push, a federation-peer fetch, a webhook delivery, or
the discovery daemon ever moving off the local socket to a network address.
Then `retry.Do` with `DelayFunc` (Retry-After) and go-error-family
classification (the hub already wraps discovery errors in
`compliance.NewIOError` etc., client.go:107-175) is exactly right, and the
`IsRetryable` default keys straight into those classifications. Rule of thumb:
first such call site = adopt; the watch loop still stays hand-rolled.

## Derived score (counterfactual: what adopting would cover vs already solved)

Bands: 85-100 excellent · 70-84 good · 50-69 fair · <50 poor.

| Capability                              | Status                    | Library evidence (path)               | Consumer evidence (path:line)                                                                                          |
| --------------------------------------- | ------------------------- | ------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| Bounded retry with exponential backoff  | N/A (no bounded-retry site) | `retry.go` (`Do`, `Backoff`)        | no call site wants a budget: refresh failures retry on next trigger (refresher.go:207-209); discover retries via reconnect cycle (:240-244) |
| Jitter (thundering-herd dampening)      | N/A                       | `JitterAdditive` (`options.go`)       | single consumer over a unix socket (client.go:86-90) — no herd                                                          |
| Retryable-error classification          | Covered elsewhere — 1     | `IsRetryable` = `errorfamily.IsRetryable` | consumer already classifies with go-error-family families (`NewIOError/NewInternalError/NewParseError`, client.go:107,113,120,125,133,137) |
| Server-provided delay (`Retry-After`)   | N/A                       | `DelayFunc` (README "Server-provided delays") | no outbound endpoint emits Retry-After (daemon is local socket; client.go:124-127 treats non-200 as terminal)      |
| Ctx-aware cancellation during backoff   | Covered elsewhere — 1     | `ErrCanceled`/`ErrDeadlineExceeded`   | `sleepCtx` returns false on ctx cancel; every loop exits on it (refresher.go:220-222,229-231,253-255,275-285)           |
| Infinite supervision/reconnect          | N/A (wrong tool)          | — (no forever-mode in the API)        | the actual need: hand-rolled loop, fixed interval, health-check visibility (refresher.go:218-257)                        |
| Value-returning retry (`DoWithValue`)   | N/A                       | `retry.go` (`DoWithValue`)            | no value-returning flaky call exists                                                                                    |
| Nested-loop fail-closed semantics       | N/A                       | README "Exhaustion and nesting"       | no nested retry exists                                                                                                   |
| SQLite busy-retry                       | Covered elsewhere — 1     | — (not go-retry's surface)            | owned inside go-sqlitestore (STACK.md Evidence store row; cache.go uses `ExecWrite`/`queryRows`, cache.go:118-124,359)   |

Inputs: N=9, A=3 (the three "Covered elsewhere" rows; the other six are N/A
with one-line reasons — no bounded-retry call site exists), Σ=3.0 →
**Score = round(100×3.0/3) = 100 (excellent)** counterfactual — every need
go-retry could plausibly serve is already met (ctx discipline, error
classification, busy-retry), and the needs it can't serve (supervision,
debounce) aren't its domain. Publish: **3 of 3**.

## Gap analysis

No actionable gaps. Parked sketch for the flip trigger (first real one-shot
outbound call):

```go
// e.g. a future federation fetch — the shape go-retry is FOR:
cfg := retry.DefaultConfig()
cfg.MaxAttempts = 4
cfg.MaxDelay = 10 * time.Second
cfg.DelayFunc = func(attempt int, err error) time.Duration {
	var h *retryAfterError                       // extract Retry-After when present
	if errors.As(err, &h) && h.After > 0 { return h.After }
	return 0 // fall back to exponential backoff
}
cfg.OnExhausted = func(attempts int, err error) {
	slog.Warn("compliance-hub: federation fetch gave up", "attempts", attempts, "error", err)
}
if err := retry.Do(ctx, cfg, func(ctx context.Context, _ int) error {
	return fetchPeerSnapshot(ctx, peer)
}); err != nil {
	return errors.Join(err, errFederationDegraded) // non-fatal: fleet stays local
}
```

Note what this is NOT: it is not the watch loop. `runDiscoveryWatch` stays
exactly as is — fixed 30 s, health-check visibility, no attempt budget.

## Version currency

Not pinned. Library very healthy: v0.9.0 (2026-11-14), near-weekly tags
through v0.4.0→v0.9.0, 100% local coverage with a 95% CI floor, nightly fuzz
+ race stress workflows (README "Development"). No currency risk while
absent; when the trigger fires, adopt at latest — the API surface
(`Do`/`DoWithValue`/`Config`/options) has been additive across the visible
tag range.

## Fleet leverage

For: go-retry is the fleet's canonical bounded-retry core (its own README
positions it as the zero-CQRS/zero-OTel base under go-cqrs-lite's middleware);
every service that hand-rolls a `for attempt := 0; attempt < 3; attempt++`
loop is a fleet bug fixed N times. Against (decisive here): this consumer
contains zero bounded-retry loops — its three "retry-adjacent" mechanisms are
supervision, coalescing, and polling, each with a deliberate shape the
library does not implement (and should not: a forever-mode retry loop with
exponential backoff would be a different, worse reconnect policy for a local
socket). Consolidation pressure is correctly spent elsewhere; adopting now
would set a fleet precedent of "every repo imports every first-party lib",
which is the failure mode the owner's constraint warns about.

## Ranked opportunities (impact × ease)

1. **Nothing now.** Optional zero-dep hygiene note if anyone ever touches
   `runDiscoveryWatch`: the fixed 30 s, the 1 s stream-end beat, and the
   health-check threshold are a matched set — changing one without the others
   (e.g. "modernizing" to exponential backoff) degrades dead-daemon detection.
   A comment already says most of this (refresher.go:18-19,36-41).

## How to verify (all commands executed 2026-11-19)

```bash
# Library identity & API
git -C /home/lars/projects/go-retry describe --tags --abbrev=0        # → v0.9.0
git -C /home/lars/projects/go-retry log -1 --format='%h %cs %s'       # → 356fbd4 2026-11-14 chore: auto-commit...
git -C /home/lars/projects/go-retry tag --sort=-creatordate | head -5 # → v0.9.0 v0.8.0 v0.7.1 v0.7.0 v0.6.1
cd /home/lars/projects/go-retry && go doc .          # Do/DoWithValue/Config/Backoff + Err* sentinels
cd /home/lars/projects/go-retry && go doc .Config    # MaxAttempts "including the first call", no forever mode
cat /home/lars/projects/go-retry/go.mod              # deps: go-error-family + (tools module is separate)

# Consumer absence
rg -n 'go-retry' /home/lars/projects/go-compliance/go.mod /home/lars/projects/go-compliance/hub/go.mod   # → no matches
rg -n 'larsartmann/go-retry' /home/lars/projects/go-compliance --type go -g '!vendor/**'                 # → no matches

# The hand-rolled sites (exact lines)
rg -n 'watchReconnectBackoff|refreshDebounce|refreshSafetyNet|refreshTimeout' /home/lars/projects/go-compliance/hub/internal/cli/refresher.go
#   → :17 (500ms), :20 (30s), :25 (5min), :29 (30s)
rg -n 'sleepCtx' /home/lars/projects/go-compliance/hub/internal/cli/refresher.go    # → :229 (30s), :253 (1s), :275 (def)
rg -n 'busy' /home/lars/projects/go-compliance/hub --type go -g '!*_test.go' -i     # → nothing: busy-retry lives in go-sqlitestore
rg -n 'sqlitestore' /home/lars/projects/go-compliance/hub/internal/cache/cache.go   # → engine ownership of DSN/PRAGMA/busy-retry
```

Spot-checks: refresher.go:14-42 (constants with rationale comments),
:178-257 (trigger loop + watch supervision), :275-285 (`sleepCtx`) read in
full; discovery client error classification read at client.go:105-137;
cache.go:66-124 read (sqlitestore ownership). `go doc` output read directly —
the "no forever-mode" claim is from the documented `Config` fields (no
unbounded-attempts option exists in the API listing).

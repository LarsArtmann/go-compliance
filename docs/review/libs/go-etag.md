# go-etag

| Field            | Value                                                                                                                                                                                                                                                                                                                                                      |
| ---------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Reviewed         | 2026-11-19                                                                                                                                                                                                                                                                                                                                                 |
| Group            | C (library-utilization audit, HTTP/wire-protocol family)                                                                                                                                                                                                                                                                                                   |
| Consumer modules | core — **absent**; hub — **absent** (not in either `go.mod`, direct or indirect; note: `httputil` _depends on_ `go-etag/server`, but httputil itself is absent too)                                                                                                                                                                                        |
| Library version  | repo tag `v0.7.1`; last commit `39f9f37 2026-11-18 "chore: auto-commit 2 changed file(s) (heuristic)"`                                                                                                                                                                                                                                                     |
| Recent tags      | `v0.7.1`, `metrics/v0.6.2`, `client/v0.6.2`, `server/v0.6.2`, `entitytag/v0.6.2`, `v0.7.0`, …                                                                                                                                                                                                                                                              |
| Modules          | root is a **deprecated alias shim** for the server package (removed in v1.0.0 per README); real modules: `go-etag/server` (middleware, FNV-64a, 304s), `go-etag/client` (conditional-GET RoundTripper cache), `go-etag/entitytag` (RFC 7232 §2.3 domain type), `go-etag/metrics` (atomic counters for hooks). Sub-module lines are at v0.6.2; root v0.7.1. |

## What the consumer sets today (verified)

- `GET /v1/compliance` (`hub/internal/cli/serve.go:83-102`): sets **only**
  `Content-Type: application/json` (serve.go:100). **No `Cache-Control`, no
  `ETag`, no `Last-Modified`.** The response body is rebuilt on every request:
  `compliance.BuildV1Report(evidence, …, time.Now())` (serve.go:91).
- Dashboard page (`hub/internal/dashboard/routes.go:66-94`): `Content-Type`
  only (text/html at :91 or JSON at :71).
- The only cache header in the whole hub is the DataStar bundle:
  `Cache-Control: public, max-age=86400` (`hub/internal/dashboard/dashboard.go:31,113-119`)
  — correct as-is: the bundle is version-pinned by its module, so expiry
  caching beats revalidation; an ETag there would add nothing.
- SSE replay is a _different mechanism_: `/events` and the JSON stream honor
  `Last-Event-ID` via `sse.LastEventIDFromRequest` + `sse.Replay` over the
  in-memory ring store (`hub/internal/events/events.go:213-222`,
  `hub/internal/events/store.go`).

## The decisive finding: naive adoption produces zero 304s on /v1/compliance

`BuildV1Report` stamps `GeneratedAt: now.UTC().Format(time.RFC3339Nano)`
(`v1.go:162`), and the handler passes `time.Now()` per request (serve.go:91).
The wire body therefore differs **on every request** at nanosecond resolution.
go-etag's default mode hashes the buffered response body (FNV-64a) — the tag
changes every request, `If-None-Match` never matches, and the middleware is
pure overhead: ~170 KB buffered + hashed per poll (fixture size: 170,792
bytes / 341 repos, `testdata/buildflow-v1-fixture.json`) with a 0% hit rate.
The same applies to the dashboard HTML (render time in the page).

For ETags to pay off here, the validator must be derived from **snapshot
identity, not body bytes** — e.g. the refresher's `lastRefreshSuccess` stamp
(`hub/internal/cli/refresher.go:69-72,169`), set by the handler with go-etag's
`SkipIfPresent: true` mode. At that point the middleware's remaining
contribution is the RFC 7232 `If-None-Match` weak-comparison + compliant-304
mechanics (~6 lines of stdlib to hand-roll) plus its observability hooks.

## Why SSE's Last-Event-ID does not make this moot (and vice versa)

Last-Event-ID replay (`events.go:220`) solves **stream resume**: a reconnecting
subscriber asks "give me the events after N" and the ring store answers. It
only exists because the stream is unbounded and stateful — there is no body to
revalidate. ETag revalidation solves the **stateless poll**: "is the snapshot
I already have still current?" — one round trip, zero body transfer. The
consumer's JSON surface (`/v1/compliance`, and the page's JSON Accept variant)
is a poll surface served to downstream consumers (overview) that do NOT hold
an SSE connection. Neither mechanism substitutes for the other; today the
poll surface simply has no conditional-request support at all.

## Verdict — **NOT-ADOPTED-justified**

The theory in the audit brief ("snapshot only changes on refresh, 5-min safety
net at most, so 304s could turn polling into no-ops") is right about the
snapshot but wrong about the bytes: as served today, the document is never
byte-identical twice, so the library's default mode cannot help, and nobody
has complained — there is exactly one known poller class (overview, local),
payload is 170 KB over loopback or a unix socket, and the v2 design doc
(`docs/decisions/v2-wire-design.md`, pain point 2: "No snapshot identity — a
consumer cannot tell two documents apart") already owns this problem at the
_contract_ level, where it should be solved first. Adding go-etag now would be
a wire-level patch for a contract-level gap.

**Flip trigger (any one):** (a) v2 lands and adds snapshot identity — then
wire it as `ETag` via `SkipIfPresent` (the derived-validator sketch below) and
go-etag's 304 mechanics + `metrics.Attach` hit-ratio counters become the cheap
correct tool; (b) a remote/WAN federation consumer starts polling at a fixed
interval and bandwidth or marshal CPU shows up (then also add
`Cache-Control: no-cache` so caches revalidate rather than serve stale);
(c) the discovery daemon ever serves a big cacheable GET snapshot to the hub
(not today — it is POST discover + SSE watch, `hub/internal/discovery/client.go:101,157`).

When a trigger fires, adoption is two lines (`etag.New` with `SkipIfPresent`) —
but see the sketch: the handler-side validator is required either way.

## Derived score (counterfactual: what adopting would cover vs already solved)

Bands: 85-100 excellent · 70-84 good · 50-69 fair · <50 poor.

| Capability                                      | Status                 | Library evidence (path)                         | Consumer evidence (path:line)                                                                                                   |
| ----------------------------------------------- | ---------------------- | ----------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| Server ETag + If-None-Match → 304               | Missed — 0 (justified) | `server/` README "Server Quick Start"           | no validator set (`hub/internal/cli/serve.go:83-102`); body non-repeatable per request (`v1.go:162`) — default mode useless     |
| Handler-validator passthrough                   | Missed — 0 (justified) | `ETagConfig.SkipIfPresent` (go doc)             | no handler sets an ETag; `lastRefreshSuccess` exists and would serve (`hub/internal/cli/refresher.go:69-72`)                    |
| Streaming/SSE safety (Skip, Hijack/Flush-aware) | Covered elsewhere — 1  | `ETagConfig.Skip`, README "Streams gracefully"  | SSE streams use go-sse directly (`hub/internal/events/events.go:215`), never wrapped in body-buffering middleware               |
| HEAD correctness (Content-Length, no body)      | N/A                    | README "RFC 7230 compliant HEAD"                | nobody issues conditional HEADs to the hub; ServeMux handles bare HEAD via GET handler                                          |
| If-Match lost-update prevention                 | N/A                    | `MatchesIfMatch`                                | zero write/PUT endpoints — nothing to lose updates on (`hub/internal/cli/serve.go:75-83`, all GET)                              |
| Client conditional-GET cache                    | N/A                    | `client/` README "Client Quick Start"           | hub's only outbound GET is the SSE watch (a stream, not a cacheable resource) — `hub/internal/discovery/client.go:158`          |
| Entity-tag domain type (parse/compare)          | N/A                    | `entitytag/`                                    | no entity-tag handling anywhere in the consumer                                                                                 |
| Hit-ratio counters                              | N/A                    | `metrics/` (`Attach`, `HitRatio`)               | no ETag surface to count on yet                                                                                                 |
| Cache-Control policy                            | Covered elsewhere — 1  | README "Pair ETags with explicit Cache-Control" | the one resource that needs it (immutable JS bundle) already has `max-age=86400` (`hub/internal/dashboard/dashboard.go:31,116`) |

Inputs: N=9, A=4 (1+0+0+1), Σ=2.0 → **score = round(100×2.0/4) = 50 (fair)**.
Publish: **2 of 4**. The two misses are real gaps in the counterfactual but are
deliberately unpurchased: the body-hash mode is technically inoperative here
(the decisive finding), and the derived-validator mode's value is blocked on
the v2 snapshot-identity decision.

## Gap analysis

### G1 — The actual fix, when wanted: derived validator, not body hash

Impact: moderate _iff_ polling grows; ease: high. stdlib-only version:

```go
// hub/internal/cli/serve.go — inside the GET /v1/compliance handler
mux.HandleFunc("GET /v1/compliance", func(w http.ResponseWriter, r *http.Request) {
	evidence := refresher.Evidence()
	if evidence == nil {
		http.Error(w, "evidence not published yet", http.StatusServiceUnavailable)
		return
	}

	// Snapshot identity: the refresh stamp, not the bytes (generated_at moves every request).
	validator := fmt.Sprintf(`"v1-%d"`, refresher.LastRefreshSuccess().UnixNano())

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ETag", validator)
	w.Header().Set("Cache-Control", "no-cache") // revalidate every poll; never serve stale unilaterally

	if r.Header.Get("If-None-Match") == validator {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	report := compliance.BuildV1Report(evidence, time.Now().Add(-window), staleAfter, Version, time.Now())
	// ... marshal + write as today
})
```

go-etag version of the same (when a flip trigger fires) — replaces the manual
string compare with RFC 7232 weak comparison and buys the hooks:

```go
cfg := etag.DefaultETagConfig()
cfg.SkipIfPresent = true // handler's snapshot-identity ETag wins over body hash
cfg.Skip = func(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/events") // never buffer the SSE stream
}
handler := etag.New(cfg)(mux)
```

Caveat to record if adopted: a 304 means the client keeps its previous body —
including its previous `generated_at`. That is correct for a snapshot whose
identity did not change, and is precisely the "snapshot identity" semantics v2
wants to make explicit.

### G2 — Do not ETag the SSE streams

`/events` (DataStar patches) and the JSON stream are unbounded responses;
body-buffering middleware is the wrong shape entirely. The library's own
`Skip` predicate and Hijack/Flush degradation exist for this, but the consumer
needs no code: it just must not wrap those handlers. Recorded here so a future
middleware sweep (see httputil.md G2) doesn't do it by accident — wrap the
`/v1/compliance` + page routes only, or set `Skip` for the stream paths.

## Version currency

Not pinned anywhere (absent from both consumer modules). Library is stable and
active: root v0.7.1 (2026-11-18), sub-modules at v0.6.2 with a consistent
four-module release train (client/entitytag/metrics/server share versions).
The root module is a deprecated shim — anyone adopting must require
`github.com/larsartmann/go-etag/server`, not the root path. No currency risk
while absent; adopt at sub-module latest when the trigger fires.

## Fleet leverage

For: every fleet HTTP API with a pollable snapshot (go-health federation
peers, CV, this hub) re-derives conditional-request logic; a shared two-sided
library (server middleware + client RoundTripper that rebuilds 304→200) is
exactly the kind of plumbing that should exist once. If v2 gives this hub
snapshot identity, the server side is a natural adopter and the _client_ side
becomes interesting for any fleet poller (e.g. a future overview that polls
`/v1/compliance` through `etagclient.NewTransport` and gets 304s for free).
Against: adopting before snapshot identity exists in the contract would bake a
hub-private workaround into the wire behavior — the kind of divergence that
costs MORE fleet maintenance later (two consumers, two notions of "unchanged").
Verdict consistent with waiting.

## Ranked opportunities (impact × ease)

1. **G1 stdlib derived-validator ETag + `Cache-Control: no-cache`** — impact
   low-moderate today, ease high; genuinely optional until a poller complains.
2. **v2 snapshot identity (contract level, already owned by
   v2-wire-design.md)** — impact high long-term, ease medium; the ETag follows
   from it for free, go-etag or not.
3. **go-etag client transport for a future federation poller** — impact
   unknown (no poller exists in-repo), ease high; park until one does.

## How to verify (all commands executed 2026-11-19)

```bash
# Library identity & health
git -C /home/lars/projects/go-etag describe --tags --abbrev=0         # → v0.7.1
git -C /home/lars/projects/go-etag log -1 --format='%h %cs %s'        # → 39f9f37 2026-11-18 chore: auto-commit...
git -C /home/lars/projects/go-etag tag --sort=-creatordate | head -5  # → v0.7.1 metrics/v0.6.2 client/v0.6.2 server/v0.6.2 entitytag/v0.6.2
head -20 /home/lars/projects/go-etag/README.md   # root = deprecated alias shim, removed in v1.0.0

# Consumer absence
rg -n 'larsartmann/go-etag' /home/lars/projects/go-compliance --type go -g '!vendor/**'   # → no matches
rg -n 'go-etag' /home/lars/projects/go-compliance/go.mod /home/lars/projects/go-compliance/hub/go.mod  # → no matches

# What the handler sets today
rg -n 'Cache-Control|ETag|Last-Modified' /home/lars/projects/go-compliance/hub --type go -g '!*_test.go'
#   → ONLY hub/internal/dashboard/dashboard.go:116 (script bundle max-age=86400)

# The decisive finding: per-request timestamp in the body
rg -n 'GeneratedAt' /home/lars/projects/go-compliance/v1.go                 # → :34 (field), :162 (now.UTC() stamp)
rg -n 'BuildV1Report\(' /home/lars/projects/go-compliance/hub/internal/cli/serve.go  # → :91 (time.Now() per request)

# Payload scale
wc -c /home/lars/projects/go-compliance/testdata/buildflow-v1-fixture.json  # → 170792
python3 -c "import json; print(len(json.load(open('/home/lars/projects/go-compliance/testdata/buildflow-v1-fixture.json'))['repos']))"  # → 341

# SSE is a different mechanism
rg -n 'LastEventIDFromRequest|sse.Replay' /home/lars/projects/go-compliance/hub/internal/events/events.go  # → :220 area

# Library API spot-checks
cd /home/lars/projects/go-etag && go doc ./server            # middleware surface
cd /home/lars/projects/go-etag && go doc ./server.ETagConfig # SkipIfPresent, Skip, MaxBufferSize semantics
```

Spot-checks: serve.go:83-102 read in full (GET only, Content-Type at :100, no
other headers); v1.go:115-170 read in full (`GeneratedAt` at :162);
dashboard.go:113-119 read (the only Cache-Control); events.go:213-260 read
(Last-Event-ID replay). The "0% hit rate" claim for the default body-hash mode
follows from the per-request `generated_at` — it is a deduction from the
verified timestamp behavior, not an executed benchmark.

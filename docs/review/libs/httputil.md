# httputil

| Field              | Value                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| ------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Reviewed           | 2026-11-19                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| Group              | C (library-utilization audit, HTTP/wire-protocol family)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| Consumer modules   | core (`github.com/larsartmann/go-compliance`, root) — **absent**; hub (`github.com/larsartmann/go-compliance/hub`) — **absent** (not in either `go.mod`, direct or indirect)                                                                                                                                                                                                                                                                                                                                                                                             |
| Library version    | repo tag `v1.9.0` (base module); last commit `e565697 2026-11-18 "chore: auto-commit 2 changed file(s) (heuristic)"`                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| Recent tags        | `v1.9.0`, `endpoints/v0.1.0`, `v1.8.0`, `v1.7.0`, `server_timing/v1.0.3`, `v1.6.0`, `server_timing/v1.0.2`, `v1.5.0` (newest first; interleaved sub-module tags)                                                                                                                                                                                                                                                                                                                                                                                                         |
| Sub-modules        | base `github.com/larsartmann/httputil` @ v1.9.0 (deps: go-error-family, go-etag/server+entitytag, `justinas/nosurf`, `golang.org/x/time`); `github.com/larsartmann/httputil/server_timing` @ v1.0.3 (zero deps); `github.com/larsartmann/httputil/endpoints` @ v0.1.0 (zero deps)                                                                                                                                                                                                                                                                                    |

## Consumer HTTP surface (what was actually located)

- Route registration happens at three sites, all in `hub/internal/cli/serve.go`:
  dashboard (`board.RegisterRoutesAt` → `GET /`, `GET /events`, `GET /datastar.js`,
  serve.go:75-79, wired in `hub/internal/dashboard/routes.go:55-61`), health
  (`probe.RegisterRoutes(mux, health.DefaultRoutes())` — go-health's own
  `/healthz /readyz /startupz /version`, serve.go:81), and `GET /v1/compliance`
  (serve.go:83). **Every route is GET; the hub exposes zero write endpoints.**
- The mux goes into a hand-rolled `http.Server` with only
  `ReadHeaderTimeout: 10s` (serve.go:22-23, serve.go:104-108). **No middleware
  at all**: no panic recovery, no per-request logging, no security headers, no
  request IDs.
- Bind/serve lifecycle is hand-rolled in `hub/internal/cli/bind.go`:
  `bindAddress` (unix-socket mkdir/stale-remove/chmod, bind.go:31-62) and
  `serveListener` (sd_notify, watchdog, graceful drain, socket cleanup,
  bind.go:95-131) — built on the **go-daemon** building blocks (bind.go:15),
  which httputil does not provide.
- The dashboard page is a real browser surface rendering repo names via templ
  (`hub/internal/dashboard/view.templ`), including a CDN script tag
  (`<script src="https://cdn.tailwindcss.com">`, view.templ:74) and an inline
  script (view.templ:75).
- The discovery client (`hub/internal/discovery/client.go:75-94`) hand-rolls a
  unix-socket `http.Client` with a custom `DialContext` Transport — client-side,
  where base httputil has nothing to offer (its surface is server middleware;
  `ClientIP` is server-side request inspection).

## Verdicts

### Sub-module: base `httputil` — **NOT-ADOPTED-justified**

The hub is a localhost/unix-socket, single-operator, all-GET service. The
middleware families base httputil offers answer threats and needs this consumer
documented away: no cross-origin browser consumers (no CORS), no writes (no
CSRF), no multi-tenant clients (no rate limiting), no proxies (no ClientIP),
localhost payloads (no compression). This mirrors the recorded decision in
STACK.md's *Deliberately excluded* table — go-appkit was rejected with "few
endpoints; `net/http` mux + `RegisterRoutes` is less ceremony. Revisit:
middleware/CORS needs grow" — and that revisit condition has not fired.

Two honest caveats keep this from being a pure shrug:

1. **The dashboard is a browser surface with zero hardening headers.** No
   `X-Content-Type-Options: nosniff`, no CSP. templ escapes content, so this is
   defense-in-depth, not a hole — but `SecurityHeaders` alone (a 3-line wrap)
   is the one base capability with non-trivial expected value.
2. **`httputil.Server` defaults are a trap for this service.**
   `DefaultServerConfig()` sets `WriteTimeout: 30s`, and net/http's write
   deadline is armed when the request headers are read — a 30 s WriteTimeout
   kills the SSE patch stream (`/events`, heartbeat 15 s, expected to live for
   hours) roughly one heartbeat after the deadline. The consumer's minimal
   `http.Server` with only `ReadHeaderTimeout` (serve.go:104-108) is SSE-correct
   by construction. Adopting `NewServer` would be a *Misuse* unless
   `WriteTimeout` is explicitly zeroed — the hand-rolled server is the better
   fit here, and `serveListener`'s sd_notify/watchdog integration
   (bind.go:103-107) has no httputil equivalent.

**Flip triggers:** (a) bind-posture open question g1 (auth) resolves toward
network exposure — then Recovery + Logging + SecurityHeaders + Timeout all
become real and adoption is a one-chain change (see sketch below); (b) an
embedder composes the hub into a larger service with an existing middleware
chain; (c) any cross-origin WebUI consumer appears (CORS); (d) the dashboard
drops the CDN tailwind script and inline script so a nonce CSP becomes
enforceable (view.templ:74-75 is what blocks `Nonce` today, not the library).

### Sub-module: `httputil/server_timing` — **NOT-ADOPTED-justified**

The fleet's reflex ("refresh-path observability") does not map: the refresher
loop (`hub/internal/cli/refresher.go:178-211`) is **background**, and
Server-Timing is a *response* header — it can only instrument request paths.
The request paths that exist are: `/v1/compliance` (per-request
`BuildV1Report` + `MarshalV1` over ~341 repos / ~170 KB — serve.go:91-101), the
dashboard page render (templ over all rows, routes.go:91-93), and the SSE
snapshot send (dashboard.go:136-158). Those are real, and `servertiming` is a
zero-dependency, Flusher/Hijacker-transparent wrapper that composes cleanly —
but for a single operator on loopback, DevTools timing on three handlers is a
nicety, not an observability gap: the actual liveness story is already owned by
the go-health probe (refresh-freshness, discovery, sse-backpressure checks —
`hub/internal/cli/health.go`), which is the fleet standard (STACK.md ★ row).

**Flip trigger:** the hub API surface grows request-path work whose latency an
operator would actually act on (e.g. v2 API with per-request queries), or the
hub gets embedded behind a shared fleet gateway where Server-Timing aggregates
across hops. Cost when it flips is trivial (stdlib-only module).

### Sub-module: `httputil/endpoints` — **NOT-ADOPTED-justified** (closest of the three to a "when")

`endpoints.Registry` solves "the listing cannot drift from what is served" by
making registration the single source of truth. The hub today registers routes
at three sites in two packages (serve.go:75-83) and documents its surface in
AGENTS.md prose — drift is possible and only convention-checked. But the
surface is 7 GET routes, stable, single-operator; the registry's payoff
(self-describing API, attack-surface documentation) is proportionate to route
count. One real friction: `probe.RegisterRoutes` (go-health) registers its own
routes internally, so a hub Registry would cover only the dashboard +
`/v1/compliance` routes unless go-health handlers are re-registered handler-by-
handler — partial adoption, two registration styles coexisting.

**Flip trigger:** the v2 hub API lands. STACK.md pins "v2 lives at the hub API
layer only" and `docs/decisions/v2-wire-design.md` (status: design sketch) will
grow the route count; adopt `Registry` for all *new* v2 routes at that point so
the listing is born drift-free. Second trigger: the hub is ever exposed to a
federation peer that wants to discover its surface.

### Overall repo verdict — **NOT-ADOPTED-justified**

Healthy, active, well-tested fleet library (98.7% coverage claim, v1.9.0,
commit 2026-11-18) whose center of mass — internet-facing middleware hardening
— is deliberately not this consumer's problem. The three sub-modules each have
a credible future trigger (auth/exposure for base, v2 API growth for
endpoints, request-path latency for server_timing), none has fired. Adoption
cost is not zero even beyond code: a new `github.com/larsartmann/*` require
triggers the STACK.md drift check + decision-log entry (STACK.md rules) and the
Nix vendorHash bootstrap ritual at **two flake sites** (AGENTS.md, "Nix / FOD
Lessons") — worth paying when a trigger fires, not before.

## Derived scores

Bands: 85-100 excellent · 70-84 good · 50-69 fair · <50 poor.

### base httputil (counterfactual inventory: what adopting would cover vs what the consumer already solves differently)

| Capability              | Status                    | Library evidence (path)                        | Consumer evidence (path:line)                                                                                                              |
| ----------------------- | ------------------------- | ---------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| Panic recovery (500)    | Partial — 0.5             | `recovery.go`, README "Panic Recovery"         | net/http's built-in per-connection recovery (process survives, connection aborted, no 500 body) — no equivalent in `hub/internal/cli/serve.go` |
| Security headers        | Missed — 0                | `security.go`, README "Security Headers"       | dashboard sets only `Content-Type` (`hub/internal/dashboard/routes.go:91`); no nosniff/CSP anywhere (rg: only Cache-Control hit is dashboard.go:116) |
| CSP nonce               | N/A                       | `nonce.go`                                     | blocked consumer-side: CDN + inline scripts in `hub/internal/dashboard/view.templ:74-75` make nonce CSP unenforceable without a refactor     |
| Request logging         | Missed — 0 (low impact)   | `logging.go`                                   | no per-request log line exists; lifecycle logging only (`hub/internal/cli/serve.go:70`, refresher.go:208)                                   |
| Request timeout         | Partial — 0.5             | `timeout.go`                                   | `ReadHeaderTimeout` set (`hub/internal/cli/serve.go:22-23,104-108`); full `Timeout()` unnecessary — handlers are fast or streaming          |
| CORS                    | N/A                       | `cors.go`                                      | localhost/unix-socket bind posture; no cross-origin consumers (`docs/decisions/bind-posture.md`)                                            |
| Client IP / trust       | N/A                       | `clientip.go`                                  | no reverse proxy in the posture; `isLoopbackBind` guards exposure (bind.go:68-89)                                                           |
| CSRF                    | N/A                       | `csrf.go` (nosurf)                             | zero write endpoints — every route is GET (`hub/internal/cli/serve.go:75-83`)                                                               |
| Rate limiting           | N/A                       | `ratelimit_keyed.go`                           | single operator on loopback/socket                                                                                                          |
| Compression             | N/A                       | `compression.go`                               | ~170 KB max payload on loopback; SSE must not be buffered/compressed anyway                                                                 |
| Body decompression      | N/A                       | `decompression.go`                             | no request bodies consumed anywhere (GET-only surface)                                                                                      |
| Health handlers         | Covered elsewhere — 1     | `health.go` (`RegisterHealth`)                 | go-health probe is the adopted fleet standard, richer (critical/non-critical, federation) — `hub/internal/cli/health.go`, STACK.md ★ row      |
| Server lifecycle        | Covered elsewhere — 1     | `server.go` (`NewServer`/`NewServerListener`)  | `serveListener` + go-daemon sd_notify/watchdog (`hub/internal/cli/bind.go:95-131`) is a *better* fit; `DefaultServerConfig.WriteTimeout=30s` would break `/events` SSE |
| Behavioral spec suite   | Partial — 0.5             | `httpspec/` (19 specs)                        | hand-rolled content-negotiation + wiring tests (`hub/internal/dashboard/routes_test.go:18-143`); nosniff/leak specs would add coverage        |
| Middleware chaining     | N/A                       | `compose.go`, `chain_test.go`                 | zero middleware exists to chain                                                                                                             |
| Language negotiation    | N/A                       | `language.go`                                  | no i18n surface                                                                                                                             |
| ETag composition        | N/A                       | README "ETag / Conditional Requests"           | see `go-etag.md` in this directory — separate verdict, separate library                                                                    |

Inputs: N=17, A=7 (Partial 0.5×3 + Missed 0×2 + Covered 1×2), Σ=3.5 →
**score = round(100×3.5/7) = 50 (fair)**. Interpretation: where the hub has a
real need, it is ~half covered by minimal stdlib code; the uncovered remainder
(security headers, request logging) is low-impact *under the current bind
posture* — which is exactly why the verdict is justified-no rather than
should-adopt. Publish: **3.5 of 7**.

### server_timing (counterfactual)

| Capability                     | Status                  | Library evidence (path)              | Consumer evidence (path:line)                                                                 |
| ------------------------------ | ----------------------- | ------------------------------------ | --------------------------------------------------------------------------------------------- |
| Per-request phase timing       | Missed — 0 (low impact) | `README.md` "Quick Start"            | `/v1/compliance` does BuildV1Report+MarshalV1 per request (`hub/internal/cli/serve.go:91-101`) — unmeasured |
| Total/TTFB metric              | Missed — 0 (low impact) | `README.md` (`total` auto-record)    | no TTFB measurement anywhere                                                                  |
| Conditional instrumentation    | N/A                     | `ServerTimingMiddlewareWhen`        | no public/debug request distinction exists (all requests are the operator's)                  |
| Background-refresh observability | N/A (wrong tool)      | — (Server-Timing is a response header) | refresh loop is background: `hub/internal/cli/refresher.go:178-211`; freshness is a go-health check instead (health.go) |

Inputs: N=4, A=2, Σ=0 → **score = 0 (poor)** as a counterfactual utilization
score — i.e., adopting today would cover ~nothing the consumer actually needs;
the need it *could* serve (request-path timing) has no operator demand yet.
Publish: **0 of 2**. This low number is the quantitative restatement of
"justified no".

### endpoints (counterfactual)

| Capability                  | Status                   | Library evidence (path)           | Consumer evidence (path:line)                                                     |
| --------------------------- | ------------------------ | --------------------------------- | --------------------------------------------------------------------------------- |
| Drift-free route registry   | Missed — 0 (low impact)  | `endpoints/README.md` "Quick Start" | three registration sites, prose documentation only (`hub/internal/cli/serve.go:75-83`) |
| Self-describing `/endpoints` | Missed — 0 (low impact) | `endpoints/README.md`             | no listing exists; surface documented in AGENTS.md prose                          |
| Content-negotiated JSON/HTML listing | N/A             | `endpoints/README.md`             | consumer already hand-rolls Accept q-value negotiation for the page (`hub/internal/dashboard/routes.go:99-134`) — but for a different resource |
| Zero-dep stdlib-only        | — (cost note, not graded)| `endpoints/go.mod`                | n/a                                                                               |

Inputs: N=4, A=2, Σ=0 → **score = 0 (poor)**; publish **0 of 2**. Same reading
as server_timing: no current demand, cheap to adopt at the v2-API trigger.

### Repo aggregate

Weighted by applicability (A_base=7, A_st=2, A_ep=2): Σ=3.5 of A=11 →
**round(100×3.5/11) = 32 (poor)**. The aggregate is deliberately pessimistic:
it measures counterfactual coverage, and this consumer's minimalism is a
recorded decision, not an accident.

## Gap analysis (actionable findings)

### G1 — Dashboard ships without `X-Content-Type-Options: nosniff` (only actionable finding today)

Severity: low (templ escapes content; loopback posture), ease: high. Do it with
3 lines of stdlib — adopting base httputil just for this is overkill:

```go
// hub/internal/cli/serve.go, before building the server
mux := http.NewServeMux()
// ...
server := &http.Server{
	Addr: flags.Addr,
	Handler: withBaselineHeaders(mux), // X-Content-Type-Options: nosniff on everything
	ReadHeaderTimeout: readHeaderTimeout,
}
```

```go
func withBaselineHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}
```

(If base httputil ever gets adopted for a bigger reason, `SecurityHeaders` +
`Chain` replaces this — the sketch is the stdlib stopgap, not a reason to
adopt.)

### G2 — Flip-trigger sketch: the g1-auth/exposure chain

When bind-posture g1 resolves toward exposure, this is the entire change
(single site, serve.go:74), illustrating why base httputil is worth keeping
warm:

```go
handler := httputil.Chain(mux,
	httputil.Recovery(slog.Default()),     // today: net/http aborts the conn, no 500
	httputil.Logging(slog.Default()),      // today: silent
	httputil.RequestID(httputil.DefaultRequestIDConfig()),
	httputil.SecurityHeaders(httputil.DefaultSecurityHeadersConfig()),
	// NO httputil.Timeout and NO NewServer defaults: WriteTimeout would kill /events SSE
)
server := &http.Server{Addr: flags.Addr, Handler: handler, ReadHeaderTimeout: readHeaderTimeout}
```

### G3 — endpoints at the v2 boundary (deferred, recorded)

When v2 routes land: wrap *only* the new registrations in
`endpoints.Registry.HandleFunc(mux, pattern, handler, summary)` and
`endpoints.Register(mux, routes)`; leave go-health's internal registration
alone (its handlers are not individually exposed). The registry then documents
exactly the surface the hub itself owns.

## Version currency

Pinned: nothing (absent from both consumer go.mod files). Latest base tag
`v1.9.0` is 1 day old at review (commit 2026-11-18) — the library is under
active development, tags roughly monthly (v1.5.0→v1.9.0 in the recent-5
window), sub-modules stable (`server_timing/v1.0.3`, `endpoints/v0.1.0` young
but zero-dep). No currency risk because there is no pin; when a flip trigger
fires, adopt at whatever is newest — no migration concerns visible in the API
table.

## Fleet leverage

Both directions weighed: **for adoption** — Lars runs many services (CV is
already an endpoints consumer per that module's README); every service that
hand-rolls recovery/logging/security-header chains is plumbing fixed N times.
Base httputil is the fleet's canonical answer and this repo eventually growing
an exposure surface would rejoin it. **against adoption** — fleet maintenance
cost is not only "did we import it" but "did we import it where it fits";
adopting middleware into a service whose threat model explicitly excludes the
middleware's threats (bind-posture.md decides localhost/socket *is* the
boundary) creates dead configuration every future reader must reason about,
plus the concrete vendorHash/STACK.md ritual cost per dep. The honest fleet
answer: this repo's HTTP surface is *deliberately* not middleware-shaped;
consolidation pressure should be spent where the hub actually duplicates
plumbing (see go-retry.md: it doesn't; see go-etag.md: it doesn't yet either).

## Ranked opportunities (impact × ease)

1. **G1 nosniff stopgap** — impact low, ease high. Do it in 5 minutes, no dep.
2. **httpspec test-only adoption** — impact low-moderate (19 behavioral specs
   for free on dashboard + /v1/compliance), ease moderate (hub go.mod test dep
   + STACK.md row + the nosniff spec will fail until G1 lands). Optional.
3. **G3 endpoints at v2-API time** — impact moderate at that point, ease high
   (zero-dep module). Deferred by design.
4. **G2 base middleware chain** — impact high *iff* g1 lands on exposure, ease
   high once triggered. Do nothing today.

## How to verify (all commands executed 2026-11-19 unless noted)

```bash
# Library identity & health
git -C /home/lars/projects/httputil describe --tags --abbrev=0        # → v1.9.0
git -C /home/lars/projects/httputil log -1 --format='%h %cs %s'       # → e565697 2026-11-18 chore: auto-commit...
git -C /home/lars/projects/httputil tag --sort=-creatordate | head -5 # → v1.9.0 endpoints/v0.1.0 v1.8.0 v1.7.0 server_timing/v1.0.3
cat /home/lars/projects/httputil/go.mod            # base deps incl. go-etag/server, nosurf, x/time
cat /home/lars/projects/httputil/server_timing/go.mod  # zero requires (stdlib-only)
cat /home/lars/projects/httputil/endpoints/go.mod     # zero requires (stdlib-only)

# Consumer absence
rg -n 'larsartmann/httputil' /home/lars/projects/go-compliance --type go -g '!vendor/**'   # → no matches
rg -n 'httputil' /home/lars/projects/go-compliance/go.mod /home/lars/projects/go-compliance/hub/go.mod  # → no matches

# Consumer surface citations
rg -n 'RegisterRoutesAt|RegisterRoutes\(|GET /v1/compliance' /home/lars/projects/go-compliance/hub/internal/cli/serve.go  # → :75 :81 :83
rg -n 'mux.HandleFunc|mux.Handle' /home/lars/projects/go-compliance/hub --type go -g '!*_test.go'                        # → only dashboard routes.go:58-60 + serve.go:83 (all GET)
rg -n 'ReadHeaderTimeout' /home/lars/projects/go-compliance/hub/internal/cli/serve.go                                    # → :107 (const readHeaderTimeout at :22-23)
rg -n 'WriteTimeout' /home/lars/projects/go-compliance/hub --type go -g '!*_test.go'                                     # → none (SSE-safe by construction)
rg -n 'Cache-Control' /home/lars/projects/go-compliance/hub --type go -g '!*_test.go'                                    # → dashboard.go:116 only
rg -n 'cdn.tailwindcss.com|<script>' /home/lars/projects/go-compliance/hub/internal/dashboard/view.templ                 # → :74 :75
rg -n 'func serveListener|func bindAddress' /home/lars/projects/go-compliance/hub/internal/cli/bind.go                   # → :95 :31

# WriteTimeout-vs-SSE claim (library side)
rg -n 'WriteTimeout' /home/lars/projects/httputil/README.md   # → DefaultServerConfig table: 30s
rg -n 'Flush|Hijack' /home/lars/projects/httputil/server_timing/README.md  # SSE passthrough claims
```

Spot-checks performed: serve.go:83 handler verified by reading the file (GET
only, `Content-Type` at :100, nothing else); bind.go:103-107 sd_notify
ordering verified; routes.go:58-61 registration verified; refresher.go:178-211
background-loop claim verified; view.templ:74-75 script tags verified. The
`WriteTimeout` net/http semantics claim (deadline armed at header read, kills
long streams) follows from the `http.Server` documentation and is the reason
`/events` cannot adopt default timeouts; it was reasoned, not executed — no
code was run against the consumer (read-only audit).

# go-health-dashboard

- Reviewed: 2026-11-19 (group A library-utilization audit)
- Consumer module: hub; **NOT in the consumer's module graph** — no require in `hub/go.mod`, zero matches in `hub/go.sum`/`go.work.sum` (`rg -c "go-health-dashboard" hub/go.sum go.work.sum` → no match), not in `go list -m all`.
- Library reviewed at: `git -C /home/lars/projects/go-health-dashboard describe --tags --abbrev=0` → `v0.13.3`; `git log -1` → `39c3465 2026-11-18 fix: hub unit reads OIDC secrets directly, not via LoadCredential`.
- Newest tags: `v0.13.3`, `v0.13.2`, `v0.13.1`, `v0.13.0`, `v0.12.2`. Tag `v0.13.3` is the newest release, **but HEAD is meaningfully ahead**: native OIDC login (`WithOIDC`, commit `8e4c070`) and the `39c3465` fix exist only post-tag — `git show v0.13.3:oidc.go` fails (file absent at the tag). No pin exists to be current or stale; this report evaluates the skip.

## Verdict: NOT-ADOPTED — justified

STACK.md records the exclusion: "same architecture, different domain (live service probes vs fleet evidence windows); mirror the pattern, do not import it" — and for the thing it was considered for (the fleet-evidence dashboard) that reason is structurally correct: this library renders a go-health `Prober`'s check map (`dashboard.go:28-42` — `CachedResponse()`/`RefreshInterval()`/probe handlers are the entire data-source interface); it cannot render fleet repo rows, so the hub's `internal/dashboard` is not duplicated work but a different domain. Two STACK claims verify as honest pattern-borrowing, not covert re-implementation: the module split rationale (dep-light core vs UI-heavy hub) and the embedder surface (`dashboard.RegisterRoutes(mux)` at `hub/internal/dashboard/routes.go:49-51` mirrors the lib's `RegisterRoutes` at `go-health-dashboard/handlers.go:280`). The genuinely open question — should the hub mount the lib **for its own health surface** (visualizing the 5-check probe at an HTML `/health`, next to the fleet dashboard, which today shows no self-health state: `rg -n 'readyz|healthz|probe' hub/internal/dashboard/` finds nothing) — resolves to "no" on fleet topology: the library ships `cmd/health-hub`, a federation hub binary whose stated purpose is rendering every service's probes from one place, and STACK's go-health row already banks on the "federation precedent" — the compliance-hub's `/readyz` JSON is a _remote_ for that central view, which subsumes a per-service HTML duplicate (including per-remote roll-ups and trend). Adoption would be cheap in dependencies (templ, templ-components, go-datastar, go-sse — all already hub deps, verified against `hub/go.mod` vs the lib's `go.mod`) but would still add a second SSE pusher/heartbeat/connection surface to harden and gain an audience of one (the single operator) a view they already have centrally. Under the fleet constraint, a well-fitting lib amortizes fixes; a redundant UI surface in every service does the opposite — this is the rare case where NOT importing the fleet lib IS the consolidation.

**Flip trigger (any one):** (a) bind-posture open question g1 lands on "expose/auth" and the operator wants a hardened status page reachable without the SSH tunnel — the lib's `WithOIDC` + `WithPublicMode` + `WithRateLimit`/`WithPerIPRateLimit` + strict CSP (`csp.go:85`) is exactly that path, but note it is **unreleased** (post-v0.13.3), so this trigger implies waiting for ≥v0.14; (b) the central health-hub stops being run (then hub-local HTML of refresh-freshness/discovery flap history via `WithTrend` + evidence log becomes the only view); (c) the operator wants a self-health card _inside_ the existing fleet dashboard — which is better built directly from `probe.CachedResponse()` into the hub's own templ components (no second SSE machinery), keeping this verdict unchanged.

## Derived score

For a justified non-adoption the score measures _utilization of applicable capabilities_, not decision quality — the verdict carries the decision. Two capabilities are conditionally applicable (nothing structural prevents mounting them; the justification is fleet topology, i.e. a decision), the rest are genuinely N/A.

- Capability inventory rows: **N = 10**
- N/A rows: 8 → **applicable A = 2**
- Fully = 0, Partial = 0, Missed = 2 × 0 → **Σ = 0**
- **Score = round(100 × 0 / 2) = 0 — low band (utilization). Publish "0 of 2".** Read with the verdict: the zero is the intended consequence of the skip, not an oversight.

## Capability inventory

| Capability                                                                                                         | Status                                                                                                                       | Library evidence (path)                                                              | Consumer evidence (path:line)                                                                                                                                                               |
| ------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Live probe dashboard at `/health` (HTML + `Accept: application/json` negotiation, severity-grouped tables, badges) | Missed Opportunity (declined — see verdict)                                                                                  | `handlers.go:23`, `view.templ`, `go-health-dashboard/handlers.go:280 RegisterRoutes` | hub probe exists (`hub/internal/cli/health.go:136`) but serves JSON-only via go-health handlers (`serve.go:81`); the fleet dashboard (`hub/internal/dashboard/`) shows no self-health state |
| Trend/evidence history for own checks (`WithTrend`, evidence log, `/health/trend`+`/health/export`)                | Missed Opportunity (declined — flap history exists centrally + in slog transitions)                                          | `history.go`, `evidence.go`, `trend.go`                                              | transitions slog'd via `WithEvaluationHook` (`health.go:139`, `checkTransitions` health.go:149-186); no history retained hub-side                                                           |
| Fleet-evidence dashboard replacement                                                                               | N-A — structural: the lib's data source is a `Prober` check map (`dashboard.go:28-42`); fleet repo rows are not probe checks | `dashboard.go:28`                                                                    | `hub/internal/dashboard/dashboard.go` (rows/stat-cards patches from `events.Hub`)                                                                                                           |
| Federation hub (`cmd/health-hub` binary + NixOS module)                                                            | N-A — the fleet already has this role centrally; a second federation surface inside compliance-hub duplicates it             | `cmd/health-hub/`, `nix/nixos-module.nix`                                            | STACK.md go-health row: "federation precedent"; the hub is a federation remote                                                                                                              |
| OIDC login / public mode / rate limits / strict CSP / connection caps                                              | N-A — bind posture is localhost-only until g1 (docs/decisions/bind-posture.md); also `WithOIDC` is unreleased at v0.13.3     | `oidc.go:195`, `options.go:328,344,283,300`, `csp.go:85`                             | `bind.go` loopback warning + bind-posture decision                                                                                                                                          |
| Prometheus metrics endpoint (`WithMetrics`)                                                                        | N-A — no scraper; hub exposes no metrics surface at all                                                                      | `metrics.go:59`                                                                      | —                                                                                                                                                                                           |
| Webhook change notifications (`WithWebhook`+HMAC)                                                                  | N-A — no receiver exists; status flips already slog                                                                          | `webhook.go`                                                                         | `health.go:149-186`                                                                                                                                                                         |
| Embedded/self-hosted DataStar SDK serving                                                                          | N-A — the hub already self-hosts the bundle for its own dashboard (independent need, same pinned module)                     | `options.go:574 WithEmbeddedDatastarSDK`                                             | `hub/internal/dashboard/dashboard.go:113-121 ScriptHandler` (serves `go-datastar/static` bytes with day-long cache)                                                                         |
| `RegisterRoutes(mux)` embedder pattern                                                                             | N-A — the pattern was deliberately borrowed, the library not (STACK's own wording); pattern use is not library utilization   | `go-health-dashboard/handlers.go:280`                                                | `hub/internal/dashboard/routes.go:49-51 RegisterRoutes`                                                                                                                                     |
| HTML/JSON content negotiation on one URL                                                                           | N-A — same: pattern borrowed (q-value negotiation)                                                                           | `handlers.go:23`                                                                     | `hub/internal/dashboard/routes.go:65-93 pageHandler` + `preferJSON` (called at 70, defined at 96-99)                                                                                        |

## Gap analysis

### 1. No HTML view of the hub's own 5 checks (the declined "miss")

What the lib offers: `dashboard.New(probe, dashboard.WithTitle("compliance-hub"))` + `dash.RegisterRoutes(mux)` renders the probe — status banner, severity groups (`evidence`/`refresh-freshness`/`discovery`/`sse-backpressure`/`cache-disk`), per-check `since`/duration, live SSE — in ~6 lines and zero new module deps (all its deps are already hub deps).

What the consumer does instead: `/readyz` JSON only (`serve.go:81`); flap visibility = slog transition lines (`health.go:149-186`) + whatever the central health-hub shows.

Impact: for a single operator on a localhost/socket-bound service whose probe is a federation remote of the central hub: ~none. The cost side is real though — a second DataStar pusher, its own SSE connection accounting, and 30+ `With*` hardening knobs to consider on every upgrade — machinery whose value is multi-viewer, multi-tenant deployments this hub explicitly is not (STACK profile: "Single operator… one single-binary hub service").

If the flip trigger ever fires, the shape is:

```go
// serve.go, after probe.RegisterRoutes(mux, health.DefaultRoutes()):
dash := ghdashboard.New(probe,
	ghdashboard.WithTitle("compliance-hub"),
	ghdashboard.WithRoutes(ghdashboard.Routes{ // only the view; probe routes stay go-health's
		Dashboard: "/health", SSE: "/health/sse",
	}),
)
_ = dash.Start(ctx)
defer dash.Shutdown()
dash.RegisterRoutes(mux)
```

### 2. Self-health inside the fleet dashboard — the cheaper third path (recommended if wanted)

If all that's wanted is "is the hub itself healthy?" on the existing page, this needs no library: `probe.CachedResponse()` (go-health, `cache.go:21`) already returns the classified snapshot, and the hub's templ components can render one badge + check chips into the existing SSE model. That keeps one dashboard, one pusher, one dependency posture — the fleet-consolidating shape.

## Version currency

- No pin. For future reference: newest tag `v0.13.3` (2026-11-12); HEAD adds native OIDC (`8e4c070`) + a secrets-handling fix (`39c3465`) — i.e. the auth machinery that would matter for flip trigger (a) is not yet in any release. If ever adopted, wait for the release cut; adopting from a pseudo-version of a UI library drags CSP/OIDC churn into every bump.

## Fleet leverage: LEAVE (adopt nowhere in this repo)

The correct fleet posture: keep the library as the fleet's _central_ probe visualizer (health-hub) and the hub as one of its remotes; keep per-service HTML views out of single-operator services. Revisit only on the flip triggers. When borrowing its _pattern_ again (as this repo did twice — embedder surface, content negotiation), continue to do it in-repo rather than importing — the pattern is free, the machinery is not.

## Ranked opportunities (impact × ease)

1. **Nothing now** — the justified-skip verdict means the ranked list is empty by design.
2. (Conditional, on flip trigger c) **Self-health badge from `probe.CachedResponse()` in the existing dashboard** — low effort, small operator win, no new deps, no second SSE surface.
3. (Conditional, on flip trigger a + a ≥v0.14 release) **Mount the dashboard with OIDC/public-mode as the exposed status page** — medium effort, only if the hub ever leaves the localhost posture.

## Appendix: how to verify

```bash
# absence from the module graph
rg -c 'go-health-dashboard' /home/lars/projects/go-compliance/hub/go.sum /home/lars/projects/go-compliance/hub/go.mod /home/lars/projects/go-compliance/go.work.sum   # no match
cd /home/lars/projects/go-compliance/hub && GOWORK=off go list -m all | grep -c go-health-dashboard     # 0
# the pattern-borrow claims
rg -n 'RegisterRoutes' /home/lars/projects/go-compliance/hub/internal/dashboard/routes.go               # 49 (pattern mirror)
rg -n 'preferJSON' /home/lars/projects/go-compliance/hub/internal/dashboard/routes.go                   # 70, 96, 99 (negotiation mirror)
# the hub's own health surface is JSON-only; no self-health in the dashboard
rg -n 'readyz|healthz|probe' /home/lars/projects/go-compliance/hub/internal/dashboard --glob '*.go'     # no match
rg -n 'probe.RegisterRoutes' /home/lars/projects/go-compliance/hub/internal/cli/serve.go                # 81
# dependency-cost claim (all lib deps already hub deps)
rg -n 'a-h/templ|go-datastar|go-sse|templ-components' /home/lars/projects/go-compliance/hub/go.mod | head
# library side
git -C /home/lars/projects/go-health-dashboard describe --tags --abbrev=0        # v0.13.3
git -C /home/lars/projects/go-health-dashboard tag --sort=-creatordate | head -5 # v0.13.3 v0.13.2 v0.13.1 v0.13.0 v0.12.2
git -C /home/lars/projects/go-health-dashboard show v0.13.3:oidc.go >/dev/null 2>&1 && echo in-tag || echo post-tag  # post-tag (OIDC unreleased)
git -C /home/lars/projects/go-health-dashboard log --oneline v0.13.3..HEAD | head -3
sed -n '28,42p' /home/lars/projects/go-health-dashboard/dashboard.go            # Prober interface (why fleet rows don't fit)
```

## Citation drift addendum (2026-11-19 post-audit wave)

`go-health-dashboard/handlers.go:280` cites the LIBRARY's own
RegisterRoutes (go-health-dashboard source; the library keeps it in
handlers.go, not routes.go) — the library-qualified path also avoids the
basename shadow of in-repo `hub/internal/dashboard/routes.go` (verified
by `scripts/check-citations.sh`).

# go-health

- Reviewed: 2026-11-19 (group A library-utilization audit)
- Consumer module: hub; pin `github.com/larsartmann/go-health v0.10.0` (direct, `hub/go.mod`). Note: the repo has 3 go.mod files (root + `testdata/httputil-interop` + `tools/doanalyzerv2`), but `checks`, `aggregate`, `federation` are packages **of the root module** — the hub's `go-health/checks` import needs no separate require, and none exists.
- Library reviewed at: `git -C /home/lars/projects/go-health describe --tags --abbrev=0` → `v0.10.0`; `git log -1` → `2c9c743 2026-11-15 chore: auto-commit 1 changed file(s) (heuristic)`.
- Newest tags: `v0.10.0`, `v0.9.0`, `v0.8.1`, `v0.8.0`, `v0.7.0`. **Pin is current.** Post-tag `git diff v0.10.0..HEAD --name-only` touches only CI/docs/`flake.lock`/`tools/` — no library `.go` changes.

## Verdict: ADOPTED

This is the best-utilized library of the four. STACK.md's reason — "the hub is a service; fleet probe standard; federation precedent (22 consumers)" — is fully cashed out: the hub builds its probe with the no-DI constructor the library provides for exactly this posture (`health.NewChecks`, health.go:136), encodes its warn-vs-fail verdicts through the library's critical/non-critical classifier rather than hand-rolled status mapping (`WithCriticalServices(evidenceCheckName)`, health.go:138 — verdict g1's "warn stays 200" is literally go-health's `StatusWarn` semantics, `classifier.go:47-50`), uses the designed observability seam (`WithEvaluationHook` → `checkTransitions()` slog, health.go:139/149-186), mounts the library's own HTTP handlers rather than hand-rolling endpoints (`probe.RegisterRoutes(mux, health.DefaultRoutes())`, serve.go:81 — `/healthz`, `/readyz`, `/startupz`, `/version`), consumes a battery from the checks catalog (`checks.Disk`, health.go:133), and follows the documented shutdown ordering (`MarkShuttingDown()` before `Shutdown()` and listener drain, serve.go:122-123, matching `lifecycle.go:210`). Panic recovery, the 5 s batch deadline, and per-check parallelism are inherited for free (`checks.go:55-92`). The single soft spot: `health.Off` — the library's purpose-built "deliberately unconfigured dependency" visibility state — is bypassed by conditionally registering `cache-disk`.

**Flip trigger:** none needed for the verdict. Micro-trigger for the one Partial row: if an operator is ever confused by `cache-disk` appearing/disappearing from `/readyz` depending on `--db`, switch to always-registering with `health.Off` (below).

## Derived score

- Capability inventory rows: **N = 15**
- N/A rows: 7 → **applicable A = 8**
- Fully = 7 × 1, Partial = 1 × 0.5, Missed = 0 → **Σ = 7.5**
- **Score = round(100 × 7.5 / 8) = 94 — very high band. Publish "7.5 of 8".**

## Capability inventory

| Capability | Status | Library evidence (path) | Consumer evidence (path:line) |
| --- | --- | --- | --- |
| Three-probe HTTP handlers + route mounting (`/healthz` `/readyz` `/startupz`, method-qualified) | Fully Leveraged | `handlers.go:37-172` (`DefaultRoutes`, `RegisterRoutes`) | `serve.go:81` — the hub serves via go-health's handlers, nothing hand-rolled |
| `GET /version` endpoint (DefaultRoutes mounts it; panics on unstamped probe) | Fully Leveraged | `version.go:56`, `handlers.go:14-36` | `serve.go:81` + `WithVersion(Version)` at health.go:137 (`Version` ldflags-overridable, app.go:17-19) |
| No-DI constructor `NewChecks` (map of `CheckFunc`) | Fully Leveraged | `checks.go:44` | `health.go:70-136` — five checks wired as plain closures |
| Critical/non-critical classification (critical fail → 503; non-critical fail → warn/200) | Fully Leveraged | `classifier.go:47-70` | `health.go:138` (`evidence` critical; `refresh-freshness`/`discovery`/`sse-backpressure`/`cache-disk` non-critical — verdict g1/g2 semantics) |
| `WithEvaluationHook` observability seam | Fully Leveraged | `options.go:64-73`; `prometheus_example_test.go` shows the exposition pattern | `health.go:139` + `checkTransitions()` (health.go:149-186): status flips slog'd, first batch logged as baseline |
| checks catalog battery: `checks.Disk` | Fully Leveraged | `checks/checks.go:27` | `health.go:131-134` (128 MiB floor on the SQLite cache dir) |
| Shutdown awareness (`MarkShuttingDown` one-way drain + `Shutdown`) | Fully Leveraged | `lifecycle.go:170-210` | `serve.go:122-123`, in the order AGENTS.md documents |
| Panic recovery + per-batch deadline + parallel check execution | Fully Leveraged | `checks.go:55-92`, `99-129` (`runBoundedCheck`) | inherited via `NewChecks` — the consumer's own `refreshTimeout` (refresher.go:44) shows it reasons about the same hung-check class for the source query |
| `Off`/`StatusOff` unconfigured-dependency visibility | Partially Used | `types.go:106` (`Off(detail)` → visibility, never verdict; classifier.go:10-14) | `health.go:131`: `cache-disk` is registered only `if cacheDir != ""` — absent (invisible) instead of off-with-reason when `--db` unset |
| samber/do wiring (`New(injector)`, `WithCriticalService[T]`, `DoServiceName`) | N-A — STACK decision: the hub wires manually; `NewChecks` is the library's sanctioned non-DI path (this is precisely why the manual-wiring decision costs nothing here) | `probe.go:161`, `service_name.go` | `docs/architecture-understanding/2026-11-18_19-39_samber-do-health-architecture.md` |
| `aggregate` (merge multiple in-process probes) | N-A — one probe in the hub process; nothing to merge | `aggregate/aggregate.go:74` | — |
| `federation` (merge remote probes over HTTP) | N-A — the hub is a federation **remote** (its `/readyz` is what a central health-hub fetches), not a hub-of-probes; mounting `federation` here would duplicate the fleet's existing `health-hub` role | `federation/federation.go:147` | STACK.md go-health row: "federation precedent" |
| checks catalog beyond Disk (`Memory`/`HTTP`/`Database`) | N-A — no `*sql.DB` handle is exposed by `go-sqlitestore`'s Store; the discovery daemon is better served by the consumer's stamp-based freshness check (refresher.go `LastDiscoverySuccess`) than a periodic `checks.HTTP` ping (it's a unix socket, and liveness ≠ freshness); `Memory` ceiling is not a fleet convention yet (stretch idea, see opportunities) | `checks/checks.go:58,93,127` | `health.go:63-90` (discovery), `refresher.go:180-186` |
| Replica/observability knobs (`WithInstanceID`, `WithBootTime`, `WithLiveThrottle`, `WithAllowedMethods`) | N-A — single instance on one machine; defaults (GET-only, no throttle) match the threat model | `options.go:40-126` | — |
| `Healthz()` combined "route traffic here?" endpoint | N-A — no load balancer fronts the localhost/socket-bound hub; the embedder question it answers doesn't arise (docs/decisions/bind-posture.md) | `accessors.go:165` | — |

## Gap analysis

### 1. `cache-disk`: invisible vs `Off` (the one Partial row)

What the lib offers: `health.Off(detail string) *OffError` (types.go:106) — a check returning it grades to `StatusOff`: shown in `/readyz` as deliberately unconfigured, excluded from the failure roll-up, satisfies the startup latch (classifier.go:10-14, 49). Built exactly for "this dependency isn't configured in this deployment".

What the consumer does instead: `if cacheDir != "" { checksByName["cache-disk"] = checks.Disk(...) }` (health.go:131-134) — without `--db`, the check silently doesn't exist, so `/readyz` gives no signal that disk headroom is simply unconfigured (vs forgotten).

Impact: small — an operator comparing `serve --source file` vs `serve --source cache --db …` boots sees a check appear/vanish with no stated reason. One line makes the deployment posture legible:

```go
checksByName["cache-disk"] = func(ctx context.Context) error {
	if cacheDir == "" {
		return health.Off("no SQLite cache configured (source != cache: --db unset)")
	}
	return checks.Disk(cacheDir, minFreeBytes)(ctx)
}
```

(Keeps the map key stable; `Off` never affects status codes.)

### 2. Not gaps — verified non-issues

- **Handlers:** `/healthz` `/readyz` `/startupz` are go-health's own handlers via `RegisterRoutes(mux, health.DefaultRoutes())` (serve.go:81) — no hand-rolled probe endpoints anywhere (`rg -n 'healthz|readyz' hub/internal/` matches only serve.go's comment and go-health call sites).
- **The load-bearing evidence contract:** `Hub.Rows() == nil` before first publish vs non-nil empty fleet (health.go:75; AGENTS.md "verdict g2") lives on the consumer side by design — go-health is the reporting machinery, the hub owns check semantics. The `evidence` check gate + regression pins (`events_test.go`, `cli/health_internal_test.go`) are the correct home.
- **do-based wiring:** STACK's manual-wiring decision would matter if the hub used `New(injector)`; with `NewChecks` there is no DI surface to forgo — the decision is free, and go-health v0.10's do integration is irrelevant to this consumer (contrary to what "samber/do is transitive via go-health" might suggest).

## Version currency

- Pin `v0.10.0` = newest tag (2026-11-04). Post-tag commits are docs/CI/tooling only — no upgrade pending, zero risk.
- The v0.10.0 release itself added the default `/version` mount + "honest-stamp" contract (`docs/status/2026-11-04_20-41_v0100-release-…md` in-repo) — the consumer already consumes both via `DefaultRoutes` + `WithVersion`.

## Fleet leverage: LEAVE (already near-maximal)

The hub is what a 22-consumer fleet standard looks like when followed: semantics encoded through the library's classifier, observability through its hook, endpoints through its handlers, shutdown through its documented ordering. The only fleet-relevant note is upstream-facing, not consumer-facing: the consumer's delta-based `sse-backpressure` check (health.go:107-127 — warns only on NEW drops so past bursts self-clear) is a genuinely reusable check shape; donating it to the `checks` catalog (as e.g. `checks.SSEDropped`) would amortize it to every SSE-serving consumer in the fleet. That's a nice-to-have, not owed.

## Ranked opportunities (impact × ease)

1. **`health.Off` for `cache-disk`** — low impact, trivial ease (one closure).
2. **Donate the delta-based SSE-backpressure check upstream** — medium fleet impact, low ease (extract + test in go-health).
3. **Stretch: `checks.Memory` ceiling** — arguable for a long-running SSE hub; only if a leak is ever observed (would be fleet-novel, not standard).

## Appendix: how to verify

```bash
# pin + module graph (checks/aggregate/federation are root-module packages — no separate pins)
cd /home/lars/projects/go-compliance/hub && GOWORK=off go list -m github.com/larsartmann/go-health   # v0.10.0
rg -n 'go-health' /home/lars/projects/go-compliance/hub/go.mod                                       # single require line
# usage sites
rg -n '"github.com/larsartmann/go-health' /home/lars/projects/go-compliance/hub --glob '!vendor/**' --glob '*.go' --glob '!*_test.go'  # health.go:13-14, serve.go:11
# handlers are the library's (no hand-rolled probe endpoints)
rg -n 'healthz|readyz|startupz' /home/lars/projects/go-compliance/hub --glob '!vendor/**' --glob '*.go'  # serve.go comment + RegisterRoutes call only
rg -n 'RegisterRoutes|NewChecks|WithCriticalServices|WithEvaluationHook|checks.Disk|MarkShuttingDown' /home/lars/projects/go-compliance/hub/internal/cli/health.go /home/lars/projects/go-compliance/hub/internal/cli/serve.go
# 81,122,123,133,136,138,139
# the load-bearing gate
rg -n 'hub.Rows\(\) == nil' /home/lars/projects/go-compliance/hub/internal/cli/health.go             # 75
# library side
git -C /home/lars/projects/go-health describe --tags --abbrev=0        # v0.10.0
git -C /home/lars/projects/go-health tag --sort=-creatordate | head -5 # v0.10.0 v0.9.0 v0.8.1 v0.8.0 v0.7.0
git -C /home/lars/projects/go-health diff v0.10.0..HEAD --name-only | grep -v '^docs/\|^.github/\|flake\|^tools/'   # empty
sed -n '47,70p' /home/lars/projects/go-health/classifier.go           # warn-vs-fail classification the hub leans on
sed -n '100,110p' /home/lars/projects/go-health/types.go              # Off()
```

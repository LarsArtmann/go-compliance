# go-appkit

- Reviewed: 2026-11-19 (group E framework audit)
- Consumer modules: core (root `github.com/larsartmann/go-compliance`, deps only go-branded-id v0.7.1 + go-error-family v0.12.0) and hub (`hub/`, cmdguard/v5 v5.1.0, go-daemon e8e1a1f, go-health v0.10.0, go-sse v0.6.2, go-datastar v0.6.2, go-sqlitestore v0.1.1, templ-components v1.27.0). **go-appkit: absent from both** (`rg 'go-appkit' go.mod hub/go.mod` → nothing; the only repo mention is the STACK.md exclusion row).
- Library reviewed at: `git -C /home/lars/projects/go-appkit describe --tags --abbrev=0` → `preset/v0.1.1`; `git log -1` → `a2a5c49 2026-11-18 chore(deps): bump go-etag server v0.6.1 -> v0.6.2 (httpx)`. Remote tags fetched (`git fetch --tags` no-op — local == origin).
- Main module (root `github.com/larsartmann/go-appkit`) tags are unprefixed: newest `v0.12.0` (2026-11-07). Newest overall tags: `preset/v0.1.1` (2026-11-13), `httpx/v0.2.0`, `cqrs/v0.9.0`, `realtime/v0.2.0`, `systemd/v0.2.0` (2026-11-08), `config/v0.1.1`, `otel/v0.4.0`, `health/v0.3.2`, `v0.12.0` (2026-11-07).
- 15 `go.mod` sub-modules: root (core) + config, cqrs, docs, errorpages, flightrecorder, flightrecorderhealth, health, httpx, integration (never released by charter), otel, preset, realtime, security, systemd.

## Verdict: NOT-ADOPTED-justified

go-appkit is a production HTTP **service skeleton**: it assembles lifecycle (Start/Drain/Shutdown hooks, graceful drain), a default middleware stack (ServerTiming→Recovery→RequestID→Logging→Timeout→SecurityHeaders), health endpoints, charmbracelet logging, and opt-in legs (systemd sd_notify, go-health bridge + dashboard, otel, metrics, error pages, SSE hub, security batteries, env-first config), capped by `preset.New` one-call composition. The hub assembles the _same concerns_ deliberately from smaller fleet primitives (cmdguard + go-daemon + go-health + manual DI — recorded decisions STACK.md:48-53) and needs almost none of the battery: zero middleware is the bind-posture decision (localhost/unix, STACK.md:22), no metrics/otel/error-pages requirement exists, and — the hard structural fact — **appkit core binds TCP only** (`service.go:133-134`: `Listen(ctx, "tcp", s.cfg.Addr)`; no listener-injection field in `ServiceConfig`), while the hub's unix-socket bind mode (`hub/internal/cli/bind.go:31-62`, `unix:` prefix) is a load-bearing posture feature. Adopting core today would _keep_ bind.go hand-rolled anyway and _add_ a framework layer between the hub and the primitives (go-health, go-sse) it already consumes directly. Non-adoption is correct; this is also the closest of the three audited frameworks to a future flip.

**Flip trigger:** the STACK.md:78 recorded trigger — "middleware/CORS needs grow" (i.e. the hub ever serves non-loopback clients and wants the security battery) — _plus_ appkit core gaining a listener-injection or unix-socket seam. Either alone is insufficient: middleware could be layered on the bare mux without appkit, and a unix-socket seam without middleware needs would still swap one small hand-roll for a framework. If a second fleet service standardizes on `preset` and the hub's ~135 LOC serve/bind becomes the fleet outlier, re-run this audit.

## Derived score

Counterfactual rubric (NOT-adopted framework): a row is Fully (1) when the consumer's existing solution is equivalent-or-better for THIS consumer's needs, Partial (0.5) when adoption would add real marginal value today, Missed (0) when the consumer lacks a capability it genuinely needs. N/A rows carry a one-line reason.

- Capability inventory rows: **N = 12**
- N/A rows: 5 (metrics, otel, error pages, httpx handler battery, cqrs sub-module) → **applicable A = 7**
- Fully = 6 × 1, Partial = 1 × 0.5, Missed = 0 → **Σ = 6.5**
- **Score = round(100 × 6.5 / 7) = 93 — very high band (≥85). Publish "6.5 of 7".**

## Capability inventory

| Capability                                                                                                       | Status                                                                                                                              | Library evidence (path)                                                                                | Consumer evidence (path:line)                                                                                                                                                                                                          |
| ---------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Service lifecycle: bind, StartHooks post-listen, drain delay, ShutdownHooks, graceful shutdown                   | Fully (solved differently; unix socket makes appkit's seam a non-fit)                                                               | `service.go:132-211` (Start: TCP listen, delegated server, hooks), `service.go:214-324` (Run/Shutdown) | `hub/internal/cli/serve.go:104-132` + `bind.go:95-131`: hand-rolled bind→NotifyReady→watchdog→serve→STOPPING→5s drain→socket unlink, incl. `unix:` mode (`bind.go:31-62`) appkit cannot express (`service.go:133` tcp-only)            |
| sd_notify integration (READY=1 post-listen, STOPPING=1 at drain, WatchdogSec/2 keepalive, no-op outside systemd) | Fully (isomorphic coverage via go-daemon)                                                                                           | `systemd/README.md` (ordering contract table); rides core hooks, requires core ≥v0.10.0                | `bind.go:103-114`: `daemon.NotifyReady` / `WatchdogInterval`+`Watchdog` / `NotifyStopping` — same contract, 4 calls, no framework (see the go-daemon.md group-A audit)                                                                 |
| Default middleware stack (recovery, request-id, logging, timeout, security headers, Server-Timing)               | Partial (zero-middleware is deliberate today; recovery/headers become real value only if the bind posture changes)                  | `README.md` (stack order), `middleware.go`                                                             | `serve.go:104-108`: bare mux, `ReadHeaderTimeout` only; posture = localhost/unix warning (`serve.go:110-116`, `docs/decisions/bind-posture.md`)                                                                                        |
| Health surface (go-health probes + real-time dashboard, one wiring call)                                         | Fully (hub wires go-health directly; the dashboard half is deliberately declined)                                                   | `health/README.md` (bridges go-health + go-health-dashboard)                                           | `hub/internal/cli/health.go:74-78` (critical `evidence` check), `:138` (`WithCriticalServices`); go-health-dashboard rejected at STACK.md:77 ("mirror the pattern, do not import it")                                                  |
| SSE hub + canonical endpoint + Last-Event-ID replay + heartbeat                                                  | Fully (consumer's hub is domain-coupled and already built/tested)                                                                   | `realtime/hub.go:17-21` (thin `Broadcaster`+`EventStore` wrapper), `realtime/handler.go`               | `hub/internal/events/events.go:71-132` (`Hub`: broadcaster + ring store + heartbeat + seq), `events.go:213-244` (snapshot-on-connect, heartbeat); domain payloads (`Snapshot`/`Update`/`DiffRows`) appkit's generic hub does not carry |
| Configuration (flag > env > file > default, typed FieldErrors)                                                   | Fully (cmdguard owns this concern; recorded)                                                                                        | `config/README.md` (`Load[T]`)                                                                         | STACK.md:48 (cmdguard/v5: "typed struct-tag per-command flags + DI + config files… vs raw cobra"); `hub/internal/cli/flags.go`                                                                                                         |
| One-call composition (preset spine: core+systemd+health dashboard+errorpages+version; env-switch legs)           | Fully (manual DI is the recorded decision)                                                                                          | `preset/README.md` (spine + `APPKIT_*` grammar)                                                        | STACK.md:50 (samber/do transitive only, manual wiring) + `docs/architecture-understanding/2026-11-18_19-39_samber-do-health-architecture.md`                                                                                           |
| Opt-in Prometheus metrics (`GET /metrics`, request histogram, build info)                                        | N/A — single operator watches the dashboard + /readyz; no metrics consumer exists for a localhost tool                              | `README.md` (Metrics section)                                                                          | —                                                                                                                                                                                                                                      |
| OpenTelemetry (provider setup, otelhttp, trace-correlated logs)                                                  | N/A — no collector in the deployment shape; nothing traces a 3-route localhost service                                              | `otel/README.md`                                                                                       | —                                                                                                                                                                                                                                      |
| Pretty classified error pages (HTML + JSON contracts)                                                            | N/A — the hub's error surface is one 503 plain-text (`serve.go:86`) + JSON errors; no browser-facing failure UX need                | `errorpages/README.md`                                                                                 | —                                                                                                                                                                                                                                      |
| Handler-DX battery (`ResultHandler`, bind/validate, negotiate, conditional, per-route deadlines)                 | N/A — "few endpoints; net/http mux + RegisterRoutes is less ceremony" (the recorded exclusion, STACK.md:78); 3 route families exist | `httpx/README.md`                                                                                      | `hub/internal/dashboard/routes.go:55-61`, `serve.go:83-102`                                                                                                                                                                            |
| CQRS/ES facade over go-cqrs-lite `system` engine                                                                 | N/A — the hub is not a command-sourced domain (see go-cqrs-lite.md, this group)                                                     | `cqrs/README.md`                                                                                       | —                                                                                                                                                                                                                                      |

## Gap analysis

Nothing actionable. The one genuine overlap — appkit/systemd ≈ go-daemon's sd_notify surface — resolves in go-daemon's favor because the hub already pins it (STACK.md:52) and appkit/systemd is coupled to core's hook seams (its go.mod requires core v0.10.0+), so it cannot be adopted standalone. The serve/bind hand-roll itself (~135 LOC) is separately audited in go-daemon.md, where the group-A finding recommends migrating it onto `daemon.Server.Start` — the primitive layer, not a service framework.

Sketch of what adoption _would_ look like (for the flip-trigger file, not a proposal): `preset.New(preset.Config{Addr: …})` replaces serve.go's mux/lifecycle, `APPKIT_REALTIME=on` mounts `/events` — but the hub's stream is DataStar patches over domain snapshot/update events, so `realtime.Mount` would be bypassed anyway; `systemd.Install` replaces bind.go's three notify calls; and the unix-socket bind would still need the hand-rolled listener + `svc` internals appkit does not expose (Start owns the listener, `service.go:133`). Net: replaces ~135 LOC, adds a framework dep tree (httputil, charmbracelet/log — note the hub's logging policy row says "slog + charm.land/log/v2", STACK.md:54, while appkit core pins `github.com/charmbracelet/log v1.0.0`), and leaves the hardest 60 lines (unix socket) in place. That math only turns positive when the middleware/security battery becomes a real need.

## Version currency

- Main module `v0.12.0` tagged 2026-11-07; 14 released sub-modules, all 0.x with an "additions-only API diff" release ritual (AGENTS.md release ritual step 1) — pre-1.0 but disciplined.
- Extremely active: 363 commits since 2026-10-20; HEAD 2026-11-18; newest tag 2026-11-13. Deps swept continuously (HEAD is a dependency bump).
- Known consumers per its AGENTS.md: cqrs-htmx `setup` (`RunWithAppkit`, pins core v0.10.0) and PapDashboard (core v0.4.0) — i.e. exactly the GOTH-stack fleet services, not read-model tools like this hub.

## Fleet leverage

go-appkit is the fleet's service-skeleton bet: collective improvements (drain ordering, watchdog contracts, security batteries, the preset grammar — e.g. the 2026-11-13 preset GOTCHA fixing a real drain deadlock) amortize across every service that rides it. The honest two-way accounting: the hub contributes its fixes one layer _down_, to the shared primitives (go-sse, go-datastar, go-health, templ-components) that appkit itself consumes through its bridges — primitives-layer fixes reach strictly more fleet consumers than assembly-layer ones. Adopting appkit here would not consolidate anything the hub duplicates with other services (the hub duplicates nothing they use; its serve/bind hand-roll is go-daemon's lane), it would only add the framework between hub and primitives. The hub is simply not in appkit's target shape (embedded GOTH web service, TCP, users, middleware) — it is a unix-socket, userless, read-side projection service. Revisit when that shape converges.

## Ranked opportunities (impact × ease)

1. **None.** Keep the STACK.md:78 row; tighten its wording to record the second flip condition: "Revisit: middleware/CORS needs grow OR appkit core gains a unix-socket/listener-injection seam." (impact low × ease trivial — a one-line doc edit, worth doing only when STACK.md is next touched).
2. Watch appkit's roadmap for a listener seam (`Deferred Register` in its AGENTS.md lists a core TLS option but no listener injection today) — if it lands, the cost side of the flip calculation changes.

## How to verify

Executed 2026-11-19, all green unless noted:

```bash
# 1. Absence in both consumer modules (exclude docs/ — this report mentions the name)
rg -l 'go-appkit' --glob '!vendor/**' --glob '!docs/**' /home/lars/projects/go-compliance
# → /home/lars/projects/go-compliance/STACK.md only (the "Deliberately excluded" row), exit 0

# 2. Library version metadata
git -C /home/lars/projects/go-appkit describe --tags --abbrev=0   # preset/v0.1.1
git -C /home/lars/projects/go-appkit log -1 --format='%h %cs %s'  # a2a5c49 2026-11-18 …
git -C /home/lars/projects/go-appkit tag --sort=-creatordate | head -16
git -C /home/lars/projects/go-appkit ls-files '*go.mod' | wc -l   # 15

# 3. TCP-only listen (the structural mismatch)
sed -n '132,137p' /home/lars/projects/go-appkit/service.go        # Listen(ctx, "tcp", s.cfg.Addr)

# 4. Consumer's unix-socket bind + sd_notify surface
sed -n '31,62p' /home/lars/projects/go-compliance/hub/internal/cli/bind.go
sed -n '103,114p' /home/lars/projects/go-compliance/hub/internal/cli/bind.go

# 5. Zero-middleware posture + manual server assembly
sed -n '104,108p' /home/lars/projects/go-compliance/hub/internal/cli/serve.go

# 6. Realtime Hub is a thin go-sse wrapper (transport only, no domain semantics)
sed -n '11,25p' /home/lars/projects/go-appkit/realtime/hub.go

# 7. Recorded decisions this audit relies on
rg -n 'go-appkit|cmdguard/v5|go-daemon|samber/do v2' /home/lars/projects/go-compliance/STACK.md
# → :48 cmdguard, :50 samber/do, :52 go-daemon, :78 go-appkit exclusion row
```

Spot-checks confirmed: `service.go:133-134` is the tcp-only `Listen`; `bind.go:103-114` is the go-daemon notify surface; STACK.md rows at lines 48/50/52/73/77/78 as cited.

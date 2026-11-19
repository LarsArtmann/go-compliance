# go-daemon

- Reviewed: 2026-11-19 (group A library-utilization audit)
- Consumer module: hub; pin `github.com/LarsArtmann/go-daemon v0.4.1-0.20261103232705-e8e1a1fff502` (direct, `hub/go.mod`; also flake input `rev=e8e1a1fff5028d1e7fc292d0eeb87f8d65e2cf36` in `flake.nix:20` + `postPatchExtra` replace at `flake.nix:146-147`). The hub's ONLY private dep (STACK.md errata 2026-11-18; AGENTS.md documents the PAT-gated CI).
- Library reviewed at: `git -C /home/lars/projects/go-daemon describe --tags --abbrev=0` → `v0.4.0`; `git log -1` → `b6af9f1 2026-11-18 chore: update flake.lock nixpkgs input to latest revision`.
- Newest tags: `v0.4.0`, `v0.3.0`, `v0.2.0`, `v0.1.1`, `v0.1.0`. **The pin is AHEAD of the newest tag** — pseudo-version on commit `e8e1a1f` (2026-11-04). It is Go-code-current: `git diff e8e1a1fff502..HEAD --stat -- '*.go'` is EMPTY (6 post-pin commits are flake.lock/auto-commit chores).

## What the pinned commit adds over the v0.4.0 tag

`git diff v0.4.0..e8e1a1f --stat -- '*.go' go.mod`: `jsonv2.go` (+67, new `JSONv2Codec` — `encoding/json/v2` adapter with caller-chosen marshal options), `negotiation.go` (+102/−22: injectable `WithJSONCodec`/`WithCBORCodec`, `ResolveRequestCodec`/`ResolveResponseCodec`/`CodecForContentType` hardening), `negotiation_contract_test.go` (+168), `notify.go` (nil-logger defaulting consolidated into `loggerOrDefault`, no API change), go.mod dep bumps (Go 1.27.2, prometheus 1.25, error-family 0.12). **None of the gap content is used by this consumer** — the hub calls only the four notify functions, whose API is identical at v0.4.0. Socket activation (`LISTEN_FDS`) is NOT in the library at any revision (ROADMAP raw idea only — `docs/status/2026-10-08*:57`).

## Verdict: UNDERUSED — severe

STACK.md's recorded reason is "unix-socket + sd_notify/systemd building blocks (project-discovery-daemon precedent) vs bare main.go" — and the consumer read "building blocks" maximally narrowly: exactly four function calls (`NotifyReady`, `NotifyStopping`, `WatchdogInterval`, `Watchdog`, all in `hub/internal/cli/bind.go:103-114`). Everything those functions are building blocks FOR — the ordered `Server.Start(ctx, Lifecycle)` contract whose ordering is "proven by the project-discovery-daemon shutdown-regression tests" (`socket.go:192`, contract at 197-234) — is hand-rolled in the consumer: `bindAddress` (bind.go:31-68) re-implements go-daemon's `listen()` (mkdir, chmod, stale removal), and `serveListener` (bind.go:95-127) re-implements `Server.Start`'s serve/drain/unlink select. The hand-roll is not just duplication; it is strictly weaker: `bind.go:46` removes an existing socket file **unconditionally**, whereas go-daemon's `CleanStaleSocket` (socket.go:394-409) probes the old socket's health endpoint first and refuses to steal a live daemon's socket. Under the fleet constraint this is the canonical anti-pattern: the ordering bug class go-daemon exists to amortize is re-opened per-project. Bonus duplication: `hub/internal/discovery/client.go` hand-rolls both a unix-socket `http.Transport` (client.go:84-92 ≙ `daemon.NewHTTPClient`/`NewHTTPClientNoTimeout`) and a single-line-only SSE `data:` scanner (client.go:184-198 ≙ `daemon.ParseSSEData`, WHATWG-conformant) — importing them adds no new privacy or dependency cost because go-daemon is already the hub's direct dep.

**Flip trigger:** any of (a) a second concurrent `serve` instance ever silently steals the first one's unix socket (the bind.go:46 blind-remove), (b) the hub grows a second long-running surface (metrics/second listener) that needs the OnListen/OnShutdown hook contract, (c) the next go-daemon lifecycle fix (idle-timeout class, drain ordering) has to be re-applied here by hand. Any one makes migrating `serveListener` onto `daemon.Server.Start` the cheapest correctness win available.

## Derived score

- Capability inventory rows: **N = 10**
- N/A rows: 2 (content negotiation, flight recorder) → **applicable A = 8**
- Fully = 3 × 1, Partial = 0, Missed = 5 × 0 → **Σ = 3**
- **Score = round(100 × 3 / 8) = 38 — low band. Publish "3 of 8".**

## Capability inventory

| Capability | Status | Library evidence (path) | Consumer evidence (path:line) |
| --- | --- | --- | --- |
| sd_notify readiness (`NotifyReady`) | Fully Leveraged | `notify.go:59` | `hub/internal/cli/bind.go:103` (listener-first ordering, matching the lib's contract) |
| sd_notify stopping (`NotifyStopping`) | Fully Leveraged | `notify.go:66` | `bind.go:114` |
| Watchdog keepalive (`WatchdogInterval` + `Watchdog`, WatchdogSec/2) | Fully Leveraged | `notify.go:75,102` | `bind.go:105-106` |
| `Server`/`SocketServer` ordered lifecycle (`Start(ctx, Lifecycle)`: Startup→listen→OnListen→READY→watchdog→serve→STOPPING→drain→unlink→OnShutdown) | Missed Opportunity | `socket.go:158-335` (`NewServer`, `NewSocketServer`, `Lifecycle` at 213-234) | hand-rolled: `bind.go:31-68` (bindAddress) + `bind.go:95-127` (serveListener) + `serve.go:104-127`; consumer builds its own `http.Server` (`serve.go:104-108`) |
| Stale-socket liveness probe (`CleanStaleSocket`/`IsSocketAlive`) | Missed Opportunity | `socket.go:394-433` (probes health path; leaves LIVE sockets alone) | `bind.go:46` removes unconditionally — a second instance steals a live daemon's socket |
| Unix-socket HTTP client transports (`NewHTTPClient`/`NewHTTPClientNoTimeout`) | Missed Opportunity | `transport.go:25,41` (pooled keep-alive; no-timeout SSE variant) | `hub/internal/discovery/client.go:84-92` hand-rolls `DialContext("unix", …)` transport |
| WHATWG SSE framing parser (`ParseSSEData`) | Missed Opportunity | `sse.go:42` (multi-line data joins, heartbeats, size limits; WPT corpus-pinned) | `client.go:184-198` hand-rolled single-line-only `data:` scanner (multi-line frames would be dropped as malformed JSON) |
| JSON/CBOR content negotiation (`ResolveRequestCodec`/`WriteEncoded`/`JSONv2Codec` — the pin's headline addition) | N-A — hub owns a frozen v1 JSON wire contract; `encoding/json/v2` stdlib is used directly (STACK JSON policy); no CBOR demand | `negotiation.go:73,98`, `jsonv2.go` | — |
| `FlightRecorder` (trace capture, slow-request trigger, `TraceHandler`) | N-A — no tracing requirement in a localhost single-operator hub; the CLI-side equivalent is cmdguard's `flightrecorder` sub-module, also unused | `flightrecorder.go:198` | — |
| sd_notify observability counters (`ReadNotifyCounters`) | Missed Opportunity | `notify.go:46` (docs: expose from /metrics) | never read; the hub exposes no metrics endpoint at all |

## Gap analysis

### 1. `serveListener` should be `daemon.Server.Start` (the big one)

What the lib offers: `NewServer(network, addr, opts...)` + `Lifecycle{Handler, Startup, OnListen, OnShutdown}` runs the exact sequence the consumer hand-rolled — including things the hand-roll lacks: `IdleTimeout` on the `http.Server` (`serve.go:104-108`, consumer sets only `ReadHeaderTimeout`, `serve.go:107`), stale-socket probing before removal, socket unlink + `NotifyStopping` on BOTH exit paths (the consumer's serve-error path at bind.go:125-127 returns without `NotifyStopping` or socket removal — it exits the process, so systemd notices anyway, but the socket file lingers until next boot's stale-removal), and a hook contract so background loops (refresher, discovery watch, currently `go`-launched at `serve.go:62-63`) become ordered `OnListen`/`OnShutdown` citizens.

What the consumer does instead: ~90 lines across `bind.go:31-127` + `serve.go:104-127` re-deriving that ordering, plus its own docs comment re-stating the ordering rationale (bind.go:90-95) — a maintenance shadow of `socket.go:213-234`'s contract comment.

Impact: deletes the riskiest plumbing in the hub; future lifecycle fixes land once for every go-daemon consumer. The consumer's deliberate divergences are all expressible as options: socket mode `0o660` (bind.go:23) → `WithSocketMode(0o660)`; health probe path → `WithHealthPath("/healthz")`; drain window 5 s (bind.go:25) → `WithShutdownGracePeriod` (same default); "listening on" to stdout → keep the `fmt.Fprintf` before `Start`, or accept the lib's slog line (`WithLogger`).

```go
// serve.go — replaces bindAddress+serveListener+the manual http.Server:
network, addr := "tcp", flags.Addr
if path, ok := strings.CutPrefix(flags.Addr, "unix:"); ok {
	network, addr = "unix", path
}

srv := daemon.NewServer(network, addr,
	daemon.WithSocketMode(0o660),            // bind.go's group-accessible choice
	daemon.WithHealthPath("/healthz"),       // probe endpoint the hub actually serves
	daemon.WithVersion(Version),
	daemon.WithShutdownGracePeriod(shutdownGracePeriod),
)

err := srv.Start(ctx, daemon.Lifecycle{
	Handler: mux,
	OnListen: func(ctx context.Context) {
		go refresher.Run(ctx)
		go runDiscoveryWatch(ctx, discovery.New(discovery.ResolveSocketPath()), refresher)
	},
	OnShutdown: func() {
		probe.MarkShuttingDown()
		probe.Shutdown()
		_ = hub.Shutdown(context.Background())
	},
})
```

(TCP support is in the pinned revision: `NewServer("tcp", …)` at `socket.go:158`, verified via `git show e8e1a1f:socket.go`.)

### 2. Discovery client: two building blocks re-implemented

What the lib offers: `NewHTTPClient(socketPath, timeout)` / `NewHTTPClientNoTimeout(socketPath)` for pooled unix-socket HTTP (transport.go:25-41 — the no-timeout variant exists precisely for SSE streams), and `ParseSSEData(r, maxBytes, emit)` (sse.go:42) for WHATWG-correct SSE framing.

What the consumer does instead: a hand-rolled `http.Transport` with unix `DialContext` (client.go:84-92 — one client serves both the 60 s-bounded Discover and the unbounded Watch), and a `bufio.Scanner` that keeps only lines with a `data:` prefix (client.go:184-198).

Impact: the privacy posture cannot get worse — go-daemon is already the hub's private dep, so the client.go:1-4 comment's "no SDK dependency" rationale is untouched. `ParseSSEData` fixes a real latent mismatch: a multi-line `data:` frame (legal SSE, and what json/v2-free encoders can emit) is silently dropped by the hand-rolled scanner. The project-discovery daemon currently emits single-line frames, so this is correctness-insurance, not an active bug.

```go
// client.go — transport:
c.http = daemon.NewHTTPClientNoTimeout(socketPath) // Watch stream; Discover already wraps ctx with 60s
// client.go — frame loop:
if err := daemon.ParseSSEData(resp.Body, maxScanLine, func(data string) {
	var event Event
	if json.Unmarshal([]byte(data), &event) != nil {
		return // a malformed frame never kills the stream
	}
	select {
	case events <- event:
	case <-ctx.Done():
	}
}); err != nil { /* stream ended */ }
```

### 3. `ReadNotifyCounters` unexposed

The lib instruments every sd_notify outcome for a `/metrics` endpoint (notify.go:31-46). The hub exposes no metrics at all, so READY/WATCHDOG send failures are visible only as slog warnings. Fold into any future metrics surface (go-health's evaluation-hook exposition would be the fleet-consistent carrier); not worth building a metrics endpoint for alone.

## Version currency

- Pin `v0.4.1-0.20261103232705-e8e1a1fff502` vs newest tag `v0.4.0` (2026-10-03): the pin adds the json/v2 codec + negotiation hardening + dep bumps (see header) — none used here, so v0.4.0 would serve this consumer identically. Post-pin: zero Go-file changes.
- Upgrade risk: none. Recommendation: ask upstream to cut `v0.4.1` at (or after) `e8e1a1f` and repin to the tag — a pseudo-version pin forces every reader to diff commits to learn "what am I on", and the flake (`flake.nix:20`) already pins the same rev explicitly, so a shared tag makes the two pin sites self-documenting.

## Fleet leverage: ADOPT (expand)

go-daemon is young (v0.4.x; bank-sync noted as first v0.4.0 consumer in-repo), and this hub is its most natural next heavy consumer: a systemd-notified, unix-or-loopback-TCP, gracefully-draining server is exactly `Server.Start`'s shape. Every lifecycle bug fixed upstream after the migration is one bug Lars does not fix here. Conversely, leaving 90 hand-rolled lines means this repo silently forks the ordering contract — the exact outcome the fleet constraint says to avoid. The two discovery-client adoptions are strictly additive (no posture change, deletes hand-rolled parsing). Leave: negotiation codecs and flight recorder (genuinely out of profile).

## Ranked opportunities (impact × ease)

1. **Swap `serveListener`+`bindAddress` for `daemon.Server.Start`** — high impact (correctness: live-socket probe, both-path cleanup, IdleTimeout; deletes ~90 LOC), medium ease (tests in `bind_internal_test.go` need re-pointing).
2. **Use `NewHTTPClientNoTimeout` + `ParseSSEData` in discovery/client.go** — medium impact (WHATWG-correct frames), high ease (~20 LOC).
3. **Request a `v0.4.1` tag; repin go.mod + flake to it** — low impact (legibility), trivial ease.
4. **Expose `ReadNotifyCounters`** when any metrics surface exists — low impact now, trivial later.

## Appendix: how to verify

```bash
# pin + module graph
cd /home/lars/projects/go-compliance/hub && GOWORK=off go list -m github.com/LarsArtmann/go-daemon  # v0.4.1-0.20261103232705-e8e1a1fff502
# the only usage sites (four calls)
rg -n 'daemon\.(NotifyReady|NotifyStopping|Watchdog)' /home/lars/projects/go-compliance/hub --glob '!vendor/**' --glob '*.go'  # bind.go:103,105,106,114
# the hand-rolled lifecycle it duplicates
rg -n 'MkdirAll|os.Remove|Chmod|ListenConfig|Shutdown' /home/lars/projects/go-compliance/hub/internal/cli/bind.go  # 33,42,46,50,55,121,124
# hand-rolled unix transport + SSE scan in discovery
rg -n 'DialContext|bufio.NewScanner|"data:"' /home/lars/projects/go-compliance/hub/internal/discovery/client.go  # 86-89, 184, 189
# pin-vs-tag gap and post-pin drift
git -C /home/lars/projects/go-daemon diff v0.4.0..e8e1a1fff502 --stat -- '*.go' go.mod   # jsonv2.go, negotiation.go, notify.go
git -C /home/lars/projects/go-daemon diff e8e1a1fff502..HEAD --stat -- '*.go'            # empty
git -C /home/lars/projects/go-daemon tag --sort=-creatordate | head -5                   # v0.4.0 v0.3.0 v0.2.0 v0.1.1 v0.1.0
git -C /home/lars/projects/go-daemon show e8e1a1f:socket.go | grep -n 'func NewServer'   # tcp-capable NewServer at the pin
# the lifecycle contract being skipped
sed -n '212,234p' /home/lars/projects/go-daemon/socket.go   # Lifecycle ordering contract (struct at 213)
# flake pin sites
rg -n 'go-daemon|postPatchExtra' /home/lars/projects/go-compliance/flake.nix | head -6       # 19-20, 35, …, 146-147
```

---

## Addendum 2026-11-19 (executed): opportunities 1+2 landed

Post-audit Wave 1 (plan `docs/planning/2026-11-19_07-55_*`, T1+T2) adopted the
two highest-ranked opportunities:

- **`ParseSSEData`** now frames `/v1/watch` in `discovery/client.go` —
  multi-line `data:` frames reassemble, over-cap lines fail loudly (logged,
  never truncated), read errors surface instead of masquerading as clean EOF.
  Pinned by two new specs (multi-line reassembly, over-cap failure) in
  `client_test.go`.
- **`IsSocketAlive` probe-first cleanup** in `bind.go`
  (`removeUnownedSocket`): a second `serve` against a live hub's socket now
  fails with "refusing to steal its socket" instead of silently killing the
  first instance. Repro-first test in `bind_internal_test.go`.
- **Full `Server.Start` adoption declined** (recorded in STACK.md decision
  log): the hand-rolled lifecycle keeps the M27-pinned sd_notify ordering and
  the minimal `http.Server`; the hand-roll is no longer *strictly weaker* now
  that the steal bug is fixed.
- Score inputs change: "Stale-socket liveness probe" and "ParseSSEData" rows
  move from Missed Opportunity to USED; verdict moves toward ADOPTED
  (38 → ~55 counterfactual). `NewHTTPClientNoTimeout` remains unused (rank-3
  opportunity) — flip trigger unchanged.

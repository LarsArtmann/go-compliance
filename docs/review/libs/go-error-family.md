# go-error-family

- Reviewed: 2026-11-19 (group D library-utilization audit)
- Consumer modules + pins: core `github.com/larsartmann/go-error-family v0.12.0` (**direct** — `errors.go`, `errors_test.go`; root `go.mod` require block); hub `v0.12.0` (**indirect**, via `github.com/larsartmann/go-compliance`; `hub/go.mod` indirect block).
- Library reviewed at: `git -C /home/lars/projects/go-error-family describe --tags --abbrev=0` → `v0.12.0`; `git log -1` → `898004e 2026-11-15 chore: auto-commit 3 changed file(s) (heuristic)`.
- Newest tags: `v0.12.0`, `v0.11.0`, `diagnose/v0.2.6`, `v0.10.3`, `agent/v0.2.5`. **Pin is current** (tags freshly fetched). Post-tag `git log v0.12.0..HEAD --oneline` = 12 commits, all chore (auto-commit CI/website/flake.lock + one golangci-lint config converge); `git diff v0.12.0..HEAD --stat` touches no library `.go` files outside the website/tests — no reason to chase HEAD.
- Module shape: this repo is a workspace of **7 go.mod sub-modules** — root (classification core, zero third-party deps, the only one the consumer pins), `agent` (experimental LLM-ish error analysis over diagnostic context), `bridge` (samber/oops ↔ error-family adapters), `diagnose` (deterministic diagnostic rules: Filesystem/Network), `diagnose/git`, `diagnose/postgres` (rule packs), `examples` (runnable demos). Only the root module is in the consumer's graph; that is correct — every sub-module is explicitly experimental (v.x) and none matches a consumer need.

## Verdict: ADOPTED (core) / UNDERUSED-slight (hub boundaries)

Core cashes out the library's central doctrine — "share the protocol, not the implementation" (README, *Complementary, not competing*) — exactly the way the README prescribes for libraries: `*compliance.Error` is the consumer's own error type implementing the library's interfaces (`Coded` at errors.go:88, `Classified` at errors.go:92–104, with the documented "families for integration with Classify()" comment at errors.go:91), every boundary in both modules constructs family-typed errors via the core constructors rather than bare wraps (`hub/internal/cache/cache.go:50,69,75,104,109,120,192…`; `hub/internal/discovery/client.go:107–174`), and the Classify mapping is pinned by tests (errors_test.go:38–50). STACK.md's reason ("fleet classification/exit-code standard, 135 consumers") is honored on the *production* side. The under-use is on the *consumption* side: the hub never asks a family question. `rg 'compliance\.Err' hub/internal -g '!*_test.go'` → **zero hits**: no `errors.Is(err, compliance.ErrIO)`, no `errorfamily.IsRetryable`, no `errorfamily.ExitCode`, no `errorfamily.HTTPStatus`, no `errorfamily.LogError`. The CLI exit code is delegated to cmdguard's own `ExitCoder` protocol (`hub/cmd/compliance-hub/main.go:14`), which `*compliance.Error` does not implement — so a transient cache-open failure (`cache.Open` → `NewIOError` → Transient) exits 1 (generic) where the fleet standard says 75 (EX_TEMPFAIL); and `/v1/compliance` hand-rolls its 503/500 (serve.go:86,95). Classification is built and paid for, then only ever rendered as a string.

**Flip trigger:** the next time a hub boundary needs to *decide* based on an error rather than log it — a retry that should distinguish transient from permanent, an exit code that should distinguish "cache unavailable" (75) from "bad flag" (1), or an HTTP handler that returns errors instead of snapshots — wire `errorfamily.IsRetryable` / `ExitCode` / `HTTPStatus` at that boundary. Concretely: if `serve` boot failures ever get classified for systemd (`Restart=on-failure` + exit 75 vs 1), adopt `errorfamily.ExitCode` in `main.go` (sketch below).

## Derived score

- Capability inventory rows: **N = 16**
- N/A rows: 8 → **applicable A = 8**
- Fully = 4 × 1, Partial = 4 × 0.5, Missed = 0 → **Σ = 6.0**
- **Score = round(100 × 6.0 / 8) = 75 — high band. Publish "6.0 of 8".**

## Capability inventory

| Capability | Status | Library evidence (path) | Consumer evidence (path:line) |
| --- | --- | --- | --- |
| Own error type implementing `Classified` (library-classification integration) | Fully Leveraged | `interfaces.go:23-27`; README "Implementing Your Own Error Type" | `errors.go:92-104` (`ErrorFamily()` maps 5 categories → 3 families; comment at :91 names the intent) |
| Own error type implementing `Coded` | Fully Leveraged | `interfaces.go:16-21` | `errors.go:88` (`ErrorCode() → "compliance.<category>"`), pinned errors_test.go:52-70 |
| `Classify()` works on consumer errors (category→family mapping tested) | Fully Leveraged | `classify.go` (precedence: `Classified` interface step 2) | `errors_test.go:38-50` — every category pinned to Rejection/Transient/Infrastructure via `errorfamily.Classify`; codes pinned at errors_test.go:52-70 |
| Family-typed construction at every IO/parse boundary (library-imports-only doctrine) | Fully Leveraged | README "Complementary, not competing" (libraries import the root module only) | `cache.go:69,75,120,192,248,263,268` + `queryRows`/`wrapIterateErr` single homes (cache.go:299,310); `discovery/client.go:120-174`; `events/events.go:321` — zero bare `fmt.Errorf` on data paths (bare `fmt.Errorf` survives only at CLI-shell/plumbing boundaries: bind.go:35-58, flags.go:58, app.go:35,60,103, health.go:88,101,123) |
| Exit-code mapping at CLI boundary | Partially Used | `classify.go:78-92` `ExitCode(err)` (family → BSD sysexits; Transient → 75) | `hub/cmd/compliance-hub/main.go:14` uses cmdguard `v5.ExitCode`, which honors only its own `ExitCoder` interface (`cli_errors_json.go:39-53` of cmdguard v5.1.0) — `*compliance.Error` implements no `ExitCode() int`, so every failure exits 1 |
| HTTP status mapping (`Family.HTTPStatus()`, `HTTPStatus(err)`) | Partially Used | `family.go:211-243` (Rejection→400, Transient→503, Infrastructure→503…), `http.go:17-28` | `serve.go:86` 503 is state-based ("evidence not published", not an error — legitimate); `serve.go:95` 500 for a `MarshalV1` failure is hand-rolled and *diverges* from the lib (Internal/Infrastructure → 503 per family.go:119 — defensible either way, but it is an untracked local decision) |
| Structured error logging with family/code fields (`LogError`/`LogErrorContext`) | Partially Used | `log.go:22-28` | `refresher.go:208,226,241` log `"error", err` via slog — family/code/retryable fields not attached (message string only) |
| Per-error structured context (`Contextual`, `WithContext`) | Partially Used | `interfaces.go:31-36`; `constructors.go` `WithContext` | Core embeds context in message format args (`errors.go` constructors: `"lock evidence file %s"`) — equivalent information, stringly-typed; no `errors.AsType[errorfamily.Contextual]` consumer |
| `IsRetryable` retry decisions | N-A — both retry loops are unconditional **by design**: the refresher retries on the next trigger/safety-net tick whatever the family (refresher.go:203-208), and the discovery watch retries a unix-socket daemon whose failures are transient by nature (refresher.go:218-261, fixed 30s backoff). There is no conditional retry loop to feed | `classify.go` (`IsRetryable`); README "backoff/jitter/idempotency are yours" | — |
| `errorfamilytest` assertion helpers | N-A — consumer policy is ginkgo/gomega (STACK.md Testing row); the existing DescribeTable at errors_test.go:38-50 performs the identical `Classify` assertions in fleet-idiomatic form | `errorfamilytest/errorfamilytest.go:31-60` | `errors_test.go:38-50` |
| `Registry` / `RegisterStdlibDefaults` (scoped registration, stdlib sentinel table) | N-A — every third-party error (modernc sqlite, exec, fs) is wrapped family-typed at the single-home boundaries before it can escape; there is no raw third-party error reaching a classifier | `registry.go`, `stdlib.go` | `cache.go:299-310` (`queryRows`/`wrapIterateErr`), `adapter.go` (CommandSource bounded read) |
| Multi-error worst-family (`errors.Join` + Classify) | N-A — no `errors.Join` call site in either module | `classify.go` (precedence step 1) | — |
| `HandleError` CLI boundary (What/Why/Fix/WayOut, stderr formatting) | N-A — cmdguard owns the CLI boundary and has its own error envelope (JSON `error.type/message/hint`, `classifyError` walks cmdguard's sentinel facts; cmdguard v5.1.0 `cli_errors_json.go`) — stacking `HandleError` would double-render | `handle.go:121-186` | `hub/cmd/compliance-hub/main.go:14`, `app.go:36-63` |
| `WrapOnce` idempotent boundary wrap | N-A — consumer wraps once at creation (constructors), never re-wraps at boundaries | README "What It Gives You" | `errors.go` constructors |
| Diagnostic rules (`diagnose`, `diagnose/git`, `diagnose/postgres`) | N-A — hub errors are local SQLite/exec/socket; the zero-dep Filesystem/Network rules answer questions this operator answers with `ls` | `diagnose/diagnose.go` (experimental) | — |
| `bridge` (samber/oops) / `agent` | N-A — no oops in the consumer (enrichment stack decision), no agent surface | `bridge/bridge.go`, `agent/agent.go` (both experimental submodules) | — |

## Gap analysis

### 1. CLI exit codes ignore the family (Partial → the one worth closing)

What the lib offers: `errorfamily.ExitCode(err)` (classify.go:78) — walks for `ExitCoder`, else classifies and maps family → BSD sysexits (Transient → 75 EX_TEMPFAIL, Rejection/Conflict → 1, Corruption → 65, Infrastructure → 69, Orchestration → 70). This is the fleet standard the STACK row names ("classification/**exit-code** standard").

What the consumer does: `hub/cmd/compliance-hub/main.go:14` — `os.Exit(v5.ExitCode(run(...)))`. cmdguard's `extractExitCode` (cmdguard v5.1.0, `cli_errors_json.go:39-53`) checks only its own `ExitCoder` interface and otherwise returns the generic 1. The comment in cmdguard even names go-error-family as a third-party type it deliberately treats as "no override". So `compliance-hub report --source cache --db /unreadable/path` (→ `NewIOError` → Transient → would be 75) and `compliance-hub report --source bogus` (→ `NewValidationError` → Rejection → 1) both exit 1. A systemd unit or wrapper script cannot tell "retry later" from "fix your flags".

Impact: low-moderate — the hub CLI is single-operator muscle memory today, but exit codes are the cheapest machine contract a CLI has, and the classification is already attached to every error the CLI returns. One line closes it:

```go
// hub/cmd/compliance-hub/main.go
func main() {
	err := run(context.Background(), os.Args[1:])
	if err == nil {
		os.Exit(0)
	}
	// cmdguard renders the message; errorfamily decides the exit code
	// (Transient → 75 EX_TEMPFAIL, Rejection → 1, …) — fleet standard.
	os.Exit(errorfamily.ExitCode(err))
}
```

(`*compliance.Error` reaches `main` unwrapped through `cli.Run`; `Classify` sees it via the `Classified` implementation, errors.go:92. Keep cmdguard's rendering — only the code derivation changes.)

### 2. `/v1/compliance` 500 vs the family table (Partial, informational)

serve.go:95 returns 500 for a `MarshalV1` failure. `*Error` from MarshalV1 would be Internal → family.go:119 maps Infrastructure → **503**. The consumer's 500 is arguably *more* correct for a marshalling bug (nobody should retry), but it is an unrecorded divergence from the fleet table — either add a one-line comment citing the divergence or switch to `errorfamily.HTTPStatus(err)`. The 503 at serve.go:86 is not an error classification at all (nil snapshot = boot state) and should stay hand-rolled.

### 3. Family fields absent from failure logs (Partial)

`refresher.go:208` logs `"compliance-hub: refresh failed", "error", err`. `errorfamily.LogErrorContext(ctx, err, slog.Default())` (log.go:28) would add `family=transient`, `code=compliance.io`, `retryable=true` — the exact fields that make a grepped fleet log triageable ("was the 03:00 refresh storm transient or permanent?"). Impact small (single operator, two log sites); adopt opportunistically at the next touch, not as a chore.

## Version currency

Pin `v0.12.0` = newest tag (fetched fresh). Post-tag history is chores only (website pnpm locks, CI pins, golangci config converge; `git log v0.12.0..HEAD --oneline` verified). Sub-module tags (`diagnose/v0.2.6`, `agent/v0.2.5`) are irrelevant to the consumer (nothing imports them). No action.

## Fleet leverage

**Leave the pin; expand consumption at hub boundaries, not the import surface.** The core module is one of the 135 consumers that make this the fleet standard, and its integration is the canonical *pattern* other libraries should copy (own type + two interfaces + tested mapping — precisely what the README's "Implementing Your Own Error Type" section asks for). Re-implementing any of that per-project is exactly the plumbing duplication the fleet constraint exists to prevent. The flip side of the constraint — a poor-fit expansion also costs — applies to the *sub-modules*: `agent`/`bridge`/`diagnose` are experimental and would drag semantics this hub doesn't need; importing the root module deeper (`ExitCode`, `HTTPStatus`, `LogError`) costs zero new dependencies (already direct in core, indirect in hub) and pays fleet dividends every time a wrapper script or systemd unit branches on an exit code. The one-line `main.go` change above is the highest-leverage move; it also makes go-compliance the reference consumer for the "exit-code standard" half of the STACK row, which is currently aspirational.

## Ranked opportunities

1. **`errorfamily.ExitCode` in `hub/cmd/compliance-hub/main.go`** (impact: low-moderate — machine-readable CLI outcomes, matches STACK row; ease: trivial, one line + a test asserting 75 for a transient source failure).
2. **`LogErrorContext` at refresher.go:208/226/241** (impact: small — greppable family fields; ease: trivial, mechanical).
3. **Pin the serve.go:95 divergence or adopt `HTTPStatus`** (impact: small — consistency vs fleet table; ease: one comment or one call).

## How to verify

```bash
# Library state
git -C /home/lars/projects/go-error-family fetch --tags && git -C /home/lars/projects/go-error-family describe --tags --abbrev=0   # v0.12.0
git -C /home/lars/projects/go-error-family log v0.12.0..HEAD --oneline | wc -l                          # 12 (all chore)
fd -t f '^go\.mod$' /home/lars/projects/go-error-family | sort                                     # 7 sub-modules (agent, bridge, diagnose, diagnose/git, diagnose/postgres, examples, root)

# Consumer pins
cd /home/lars/projects/go-compliance && rg 'go-error-family' go.mod hub/go.mod                          # core direct v0.12.0; hub indirect v0.12.0
cd hub && GOWORK=off go mod why github.com/larsartmann/go-error-family                                  # via github.com/larsartmann/go-compliance

# Consumer usage (core)
cd /home/lars/projects/go-compliance && rg -n 'errorfamily' errors.go errors_test.go                    # errors.go:7,88,92; errors_test.go:7-49
rg -n 'ErrorFamily|ErrorCode' errors.go                                                                 # :88, :92

# The gap: zero family *consumption* in hub
rg -n 'compliance\.Err|errorfamily\.' hub/internal hub/cmd -g '!*_test.go'                               # no hits
sed -n '39,53p' /home/lars/.cache/go-mod/github.com/larsartmann/cmdguard/v5@v5.1.0/pkg/cmdguard/v5/cli_errors_json.go   # extractExitCode → generic 1
sed -n '78,92p' /home/lars/projects/go-error-family/classify.go                                         # ExitCode(err)
rg -n 'StatusServiceUnavailable|StatusInternalServerError' hub/internal/cli/serve.go                    # :86, :95
rg -n 'slog\.Warn' hub/internal/cli/refresher.go                                                        # :208, :226, :241
```

All commands executed 2026-11-19; outputs as annotated. Every file:line cited above was read directly.

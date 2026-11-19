# cmdguard

- Reviewed: 2026-11-19 (group A library-utilization audit)
- Consumer module: hub (`github.com/larsartmann/go-compliance/hub`); pin `github.com/larsartmann/cmdguard/v5 v5.1.0` (direct, `hub/go.mod`). Core module: no cmdguard.
- Library reviewed at: `git -C /home/lars/projects/cmdguard describe --tags --abbrev=0` → `v5.1.0`; `git log -1` → `5290526 2026-11-17 chore: auto-commit 1 changed file(s) (heuristic)`.
- Newest tags (`git tag --sort=-creatordate | head -5`): `v5.1.0`, `v4.3.1`, `v4.3.0`, `lint/v0.3.0`, `schema/v0.1.0`. **Pin is current** (v5.1.0, tagged 2026-11-11, is the newest major tag).
- Library contract reference: `~/.config/crush/skills/cmdguard/SKILL.md` (verified against `~/projects/cmdguard` source).

## Verdict: UNDERUSED — moderate

STACK.md chose cmdguard over "raw cobra + package-level flag vars (BuildFlow compliance's current wart)", and on that axis the hub delivers: zero raw cobra imports (`rg -n "cobra" hub/ --glob '!vendor/**'` finds only the `// indirect` line in `hub/go.mod`), typed per-command flag structs with correct scoping (`--addr` exists only on `serve`), and the exit-code contract wired through `go-error-family`'s structural `ExitCoder`. But the recorded reason — "typed struct-tag per-command flags **+ DI + config files** (36 consumers)" — names two pillars the consumer doesn't touch: the root `Config` is an empty struct (`hub/internal/cli/app.go:23`) and `WithEnvPrefix("COMPLIANCE_HUB_")` (app.go:32) is inert because not one field carries an `env:` tag. The CLI uses cmdguard's happy path (NewCLI/NewCommand/AddCommands/ExecuteWithArgs) and leaves the production machinery — config files, middleware, output-format flags, the Builder, cmdguard-lint in CI — on the table. Some of that is honest deferral (app.go:21: "config-file and env knobs land with the service hardening pass"), some is missed fleet standardization (no cmdguard-lint gate in `.github/workflows/ci.yml` despite the hub satisfying every rule it would check).

**Flip trigger:** the "service hardening pass" lands (then `WithConfigFile` + `env:` tags become owed work, not optional), OR a second consumer starts scripting `compliance-hub report` in pipelines (then adopt the post-v5.1.0 `pkg/testutil.AssertDescribeContract` harness to pin the CLI contract, and add `--output` for `table`). Either moves this to ADOPTED-leaning; neither has happened yet.

## Derived score

- Capability inventory rows: **N = 15**
- N/A rows: 2 (prompts sub-module, DI container wiring) → **applicable A = 13**
- Fully = 6 × 1, Partial = 3 × 0.5, Missed = 4 × 0 → **Σ = 7.5**
- **Score = round(100 × 7.5 / 13) = 58 — moderate band. Publish "7.5 of 13".**

## Capability inventory

| Capability | Status | Library evidence (path) | Consumer evidence (path:line) |
| --- | --- | --- | --- |
| Typed struct-tag flags (`flag`/`default`/`help`) | Fully Leveraged | `pkg/cmdguard/v5/flags.go`, `config_parsing.go` | `hub/internal/cli/flags.go:17-44` (QueryFlags/ServeFlags) |
| Per-command F-struct flag scoping | Fully Leveraged | `pkg/cmdguard/v5/cli_command.go:217`, `cli_flag_scoping_test.go` | `hub/internal/cli/app.go:38,45,52` (QueryFlags per command; `Addr` only in ServeFlags flags.go:39-43) |
| Tag vocabulary: `values`/`validate`/`required` | Partially Used | `config_parsing.go:120-302` | `values:"file,buildflow,cache"` flags.go:26 is the only constraint tag; no `validate`/`required`/`env` anywhere (`rg 'env:"\|required:"\|validate:"' hub/internal/cli/` → no match) |
| Exit-code contract (`ExitCode`/`ExitCoder`) | Fully Leveraged | `pkg/cmdguard/v5` errors/exit docs; `ExitCode(err)` semantics | `hub/cmd/compliance-hub/main.go:12` (`os.Exit(v5.ExitCode(run(...)))`); hub errors are `go-error-family` `*Error` with `ExitCode() int` (mod cache `go-error-family@v0.12.0/error.go:324`) — structurally satisfies cmdguard's contract |
| Signal handling + graceful shutdown (v5 default) | Fully Leveraged | `cli.go:175-179`, `cli_graceful_shutdown_test.go` | load-bearing: `runServe`'s whole drain path (`hub/internal/cli/serve.go:105-127`) triggers off cmdguard's ctx cancel |
| Batteries: `version`/`doctor`/`describe` + strict validation | Fully Leveraged | `cli_batteries.go`, `doctor.go`, `describe.go` | `app.go:31` `WithCLIVersion(Version)`; every command has `WithShort` (app.go:40,47,54) satisfying strict validation |
| Env prefix + `env:` fallback | Partially Used | `flags_parse.go:46-71` (flag → env → default) | `app.go:32` sets `WithEnvPrefix` but zero `env:` tags exist → prefix is dead weight today |
| Config files (`WithConfigFile`, strict keys, `config explain`) | Missed Opportunity | `cli_options.go:283-294`, `config_explain.go` | `app.go:23` `type Config struct{}` with the deferral comment at app.go:21-22 |
| Builder API (`NewBuilder`/`Add`/`Try`/`Check`/`Build`) | Missed Opportunity | `pkg/cmdguard/v5/cli_builder.go` (v5-preferred shape) | `app.go:38-59` hand-ladders three `NewCommand` errors + `AddCommands` (v4-style, still supported) |
| Middleware (spinner/telemetry/auditlog) | Missed Opportunity | `auditlog.go:232 AuditMiddleware`, sub-modules `spinner/`, `telemetry/` | none (`rg "Middleware\|Spinner\|WithConfigFile" hub/` → no match outside vendor/tests) |
| Output-format integration (`--output`/`WithOutputFormat` → go-output formats) | Missed Opportunity | `pkg/cmdguard/v5/cli_output.go:11-60`, `cli_options.go:277` | `table` renders terminal-only (`hub/internal/cli/table.go:47-83`); `report` is JSON-only (`report.go:19`) |
| cmdguard-lint gate (CG001-CG008) in CI | Missed Opportunity | `lint/` sub-module, rules CG001-CG008 | `.github/workflows/ci.yml` — no cmdguard-lint step (`rg -n "cmdguard-lint" /home/lars/projects/go-compliance` → no match) |
| Value types (`Duration` codec) | Partially Used | `pkg/cmdguard/v5/types_duration.go:16` (plain `time.ParseDuration`, no day suffix) | `hub/internal/cli/flags.go:113-133` hand-rolls `parseDurationSupportingDays` ("7d" BuildFlow muscle memory) because the lib codec lacks `Nd` |
| Prompts sub-module (`prompt:` tags, huh) | N-A — non-interactive service CLI; no missing-value prompt flows exist or are wanted | `prompts/prompts.go` | — |
| DI container wiring (do scope, `WithCleanup`, graceful DI shutdown) | N-A — STACK decision: hub wires manually; cmdguard's CLI-level T-config needs no container | `pkg/cmdguard/v5` DI docs; `samber-do-best-practices` skill | `docs/architecture-understanding/2026-11-18_19-39_samber-do-health-architecture.md` |

## Gap analysis

### 1. cmdguard-lint not in CI (cheapest close)

What the lib offers: `lint/` sub-module + `cmdguard-lint` binary, rules CG001-CG008 (execute bypass, stale major, constructor panics, runtime SetVersion, duplicate version options, error re-print, missing WithShort, missing WithCLIVersion). The hub passes all of them today (ExitCode used at main.go:12; every `New*` error checked in app.go; `WithShort` and `WithCLIVersion` set).

What the consumer does instead: nothing — no gate; a future edit could regress (e.g. a constructor panic via ignored `NewCommand` error, or re-printing the execute error) with no signal.

Impact: low immediate value (code is clean), high standardization value across Lars's fleet: one CI step keeps every consumer honest against the same contract, and rule improvements amortize to all 36 consumers simultaneously.

Recommended usage:

```yaml
# .github/workflows/ci.yml (new step in lint-hub job)
- name: cmdguard-lint
  run: go run github.com/larsartmann/cmdguard/lint/cmd/cmdguard-lint@v0.3.0 ./...
  working-directory: hub
```

### 2. Output-format flag for `table`

What the lib offers: `WithOutputFormat(output.FormatTable)` + auto `--output` flag whose help lists go-output's registered formats (`cli_output.go:26-41`); handlers read `cli.OutputFormat()`. go-output already emits JSON/CSV/MD/HTML (STACK row "Terminal output").

What the consumer does instead: `runTable` always renders the lipgloss table (table.go:47-83); machine consumers must use `report` (v1 JSON) which lacks the derived RATE/VERDICT column projection the table shows.

Impact: `compliance-hub table --output md` would drop into READMEs/issues for ~5 lines; consistency with the 51-consumer go-output convention. Not urgent — the v1 JSON contract already covers scripting.

```go
// app.go — one option:
cli, err := v5.NewCLI("compliance-hub", "...", Config{},
	v5.WithCLIVersion(Version),
	v5.WithEnvPrefix("COMPLIANCE_HUB_"),
	v5.WithOutputFormat(output.FormatTable), // registers --output on the CLI
)
// table.go handler:
switch cli.OutputFormat() {
case output.FormatJSON, output.FormatCSV, output.FormatMarkdown:
	// go-output marshal of the same row model
}
```

### 3. Config files / env tags — owed by the consumer's own roadmap

What the lib offers: `WithConfigFile(paths...)` (JSON/YAML/TOML, `$VAR`/`~` expansion), strict keys with "did you mean" (v5 default), `config explain` opt-in.

What the consumer does instead: `Config struct{}` + a comment promising this lands with the service hardening pass (app.go:21-22).

Impact: `serve` currently takes ~8 flags on every invocation; a systemd unit or shell alias re-states them. A `compliance-hub.yaml` (or `COMPLIANCE_HUB_*` env tags on `ServeFlags`) is exactly the pass cmdguard was chosen for. Deliberate, documented, and the single biggest latent item.

### 4. Duration codec without day suffixes — a fleet-level ask, not a consumer fix

What the lib offers: `Duration` value type (`types_duration.go`) validated and enum-able, but parsed by plain `time.ParseDuration`.

What the consumer does instead: 30 lines of `parseDurationSupportingDays`/`parseDaysSuffix` (flags.go:113-133) to keep "7d"/"30d" BuildFlow muscle memory.

Impact: every fleet CLI with day-window flags re-implements this. The consolidation-correct fix is in cmdguard (accept `Nd` in `ParseDuration` or add a `DayDuration` codec), amortizing the deletion across 36 consumers — not more hand-rolling in go-compliance. File the upstream ask; leave the consumer code until cmdguard ships it.

### 5. Builder instead of the error ladder (cosmetic)

`app.go:38-59` is a 3-rung ladder plus `AddCommands`; v5's `Builder` collapses it. Pure style; zero behavior change. Do it only when touching app.go anyway.

## Version currency

- Pin `v5.1.0` = newest tag (2026-11-11). Post-tag `git log v5.1.0..HEAD --name-only -- '*.go'`: `examples/docs-generator/main.go`, `lint/ast_helpers.go`, and **`pkg/testutil/describe_golden.go`** — the describe-contract snapshot harness (`testutil.AssertDescribeContract`, commit `39be10c`, 2026-11-11, NOT in v5.1.0) that pins a CLI's `describe` JSON against a committed golden.
- Gap content: consumer-usable tooling only; no breaking changes to the pinned API surface.
- Upgrade risk: none pending. When a v5.2/v6 tags, the harness becomes adoptable to pin the hub's CLI contract (protects `report`'s flags for scripts/agents).

## Fleet leverage: EXPAND (two cheap moves now, one when the pass lands)

1. Now: add the cmdguard-lint CI gate (rule improvements then flow to this repo for free).
2. Now: upstream the `Nd` duration-suffix ask to cmdguard — the highest-leverage single change because "7d" windows are a BuildFlow-fleet idiom, and every consumer currently pays the same 30 LOC.
3. When the service hardening pass lands: config file + `env:` tags via cmdguard's native machinery (do NOT hand-roll env parsing — that forfeits the recorded reason cmdguard was chosen).
4. Leave: prompts, DI wiring, middleware — genuinely inapplicable to a 3-command non-interactive service CLI; adopting them would be ceremony, not consolidation.

## Ranked opportunities (impact × ease)

1. **cmdguard-lint in CI** — low effort, low-but-durable impact (fleet gate parity). 
2. **Upstream `Nd` Duration codec, then delete flags.go:113-133** — medium effort, fleet-wide impact.
3. **Config file + env tags at the hardening pass** — medium effort, high impact for `serve` operators (already promised in-code).
4. **`--output` on `table`** — low effort, low impact (v1 JSON already scriptable).
5. **Builder migration** — trivial effort, cosmetic impact.

## Appendix: how to verify

```bash
# pin + module graph
cd /home/lars/projects/go-compliance/hub && GOWORK=off go list -m github.com/larsartmann/cmdguard/v5   # v5.1.0
# no raw cobra usage outside the indirect go.mod line
rg -n '"github.com/spf13/cobra' /home/lars/projects/go-compliance/hub --glob '!vendor/**' --glob '*.go'  # no match
rg -n 'spf13/cobra' /home/lars/projects/go-compliance/hub/go.mod                                        # // indirect only
# usage sites
rg -n 'cmdguard/v5' /home/lars/projects/go-compliance/hub --glob '!vendor/**' --glob '*.go'             # main.go:8, app.go:11
# absent features: env/required/validate tags, middleware, config file, ExecuteAndExit
rg -n 'env:"|required:"|validate:"|prompt:"|secret:"' /home/lars/projects/go-compliance/hub/internal/cli # no match
rg -n 'Middleware|Spinner|WithConfigFile|ExecuteAndExit' /home/lars/projects/go-compliance/hub --glob '!vendor/**'  # no non-test match
# no lint gate
rg -n 'cmdguard-lint' /home/lars/projects/go-compliance                                                                                # no match
# library side
git -C /home/lars/projects/cmdguard describe --tags --abbrev=0        # v5.1.0
git -C /home/lars/projects/cmdguard tag --sort=-creatordate | head -5 # v5.1.0 v4.3.1 v4.3.0 lint/v0.3.0 schema/v0.1.0
git -C /home/lars/projects/cmdguard diff v5.1.0..HEAD --name-only -- '*.go' | grep -v website  # examples/, lint/ast_helpers.go, pkg/testutil/describe_golden.go
sed -n '16,25p' /home/lars/projects/cmdguard/pkg/cmdguard/v5/types_duration.go  # plain time.ParseDuration, no Nd
```

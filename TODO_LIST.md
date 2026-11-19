# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, use ROADMAP.md.
> Items are ranked by impact. Status is verified, not assumed.

Seeded from `docs/planning/2026-11-17_20-01_go-compliance-pareto-execution-plan.md`
(M17/F078) plus the deliberate debts in
`docs/status/2026-11-18_06-47_go-compliance-m10-m20-session-report.md` §f.
The plan is a snapshot: this file is the living backlog.

## Status legend

| Status           | Meaning                                                 |
| ---------------- | ------------------------------------------------------- |
| 🔴 `TODO`        | Not started. Needs doing.                               |
| 🟡 `IN_PROGRESS` | Actively being worked on.                               |
| 🔵 `BLOCKED`     | Cannot proceed, external dependency or decision needed. |
| 🟢 `DONE`        | Completed. Remove from this list and log in CHANGELOG.  |

## High Impact

| Task                                                                                                                             | Status    | Impact | Effort | Evidence                                    |
| -------------------------------------------------------------------------------------------------------------------------------- | --------- | ------ | ------ | ------------------------------------------- |
| Adopt go-filewatcher: watch the evidence source's directory (SQLite `-wal` sidecars / reporter atomic renames) so cross-process writes refresh in ~1-2s instead of the 5-min safety net; note the DEFAULT `--source file` has NO write trigger at all today | 🔴 `TODO` | High   | 2h     | docs/review/libs/go-filewatcher.md (2026-11-19 audit; NOT-ADOPTED-but-should) |
| Replace the discovery client's hand-rolled SSE `data:` scanner with `daemon.ParseSSEData` (already a direct dep; current parser drops multi-line frames and skips `scanner.Err()`, gopls-flagged) | 🔴 `TODO` | High   | 1h     | docs/review/libs/go-daemon.md + go-sse.md (both groups independently flagged client.go:184-201) |
| Stop unconditional unix-socket removal on bind (`bind.go:46` removes a live daemon's socket; adopt `daemon.CleanStaleSocket` probe-first behavior or the `Server.Start` lifecycle) | 🔴 `TODO` | High   | 1-2h   | docs/review/libs/go-daemon.md (UNDERUSED-severe) |
| Fleet exit-code contract: map `*compliance.Error` classifications through `errorfamily.ExitCode(err)` in `main.go` (hub currently exits generic 1 on transient failures) | 🔴 `TODO` | Med    | 10m    | docs/review/libs/go-error-family.md (hub boundaries) |

## Medium Impact

| Task                                                                                                                             | Status    | Impact | Effort | Evidence                                    |
| -------------------------------------------------------------------------------------------------------------------------------- | --------- | ------ | ------ | ------------------------------------------- |
| Address GO-2026-6859 (net/http MaxBytesReader header race): bump Go toolchain as needed, re-run govulncheck, confirm 0 reachable | 🔴 `TODO` | Med    | 20m    | 23-40 §f10; re-appears every buildflow run  |
| Create `SECURITY.md` (root, `.github/`, or `docs/` per the tool's lookup paths)                                                  | 🔴 `TODO` | Med    | 10m    | 23-40 §f6; error-severity finding every run |

## Low Impact

| Task                                                                                                                                             | Status    | Impact | Effort | Evidence                                                                    |
| ------------------------------------------------------------------------------------------------------------------------------------------------ | --------- | ------ | ------ | --------------------------------------------------------------------------- |
| Drop hub `replace` directive when core tags v0.1.0                                                                                               | 🔴 `TODO` | Low    | 10m    | `hub/go.mod` replace block                                                  |
| `templ fmt` gate for `.templ` sources in CI (canonical templ formatting; generated `*_templ.go` already committed)                               | 🔴 `TODO` | Low    | 10m    | 06-47 §39; `go tool templ fmt ./internal/dashboard/`; dprint excludes templ |
| Archive sweep: `git mv` aged status reports to `docs/status/archived/` + archived README manifest (cadence gate threshold 15; currently 10)      | 🔴 `TODO` | Low    | 15m    | 18-06 §8 remainder; docs-cadence gate                                       |
| Investigate io-history-ratchet anomaly (43 files hashed vs median 16) with `buildflow --profile-io`                                              | 🔴 `TODO` | Low    | 30m    | 23-40 §f30                                                                  |
| Confirm lychee's 4 path-excluded unscanned files are intentionally excluded                                                                      | 🔴 `TODO` | Low    | 10m    | 23-40 §f41                                                                  |
| Core suite `-race` parity: run root module with `-race -count=1` + add it to the AGENTS.md commands block (hub already canonical)                | 🔴 `TODO` | Low    | 5m     | 00-53 §b1/c1                                                                |
| Pin README health thresholds (128 MiB / 10 min) in the `/readyz` payload contract test or loosen the prose, so README drift fails CI             | 🔴 `TODO` | Low    | 10m    | 00-53 §d5; README "Health surface" vs health.go/refresher.go constants      |
| Add devShell build to the standing verification ritual (AGENTS.md commands block)                                                                | 🔴 `TODO` | Low    | 5m     | 23-40 §f5; 00-53 §c5                                                        |
| `nix flake check --all-systems` once; record aarch64-darwin/aarch64-linux status                                                                 | 🔴 `TODO` | Low    | 10m    | 00-53 §c4; nix omitted-systems warning                                      |
| Semantic audit of the 23-40 report's ~22 token-"matched" open items (docs-gates token-match green ≠ harvest); strike or route each               | 🔴 `TODO` | Low    | 30m    | 00-53 §b3; docs-freshness matched-items list                                |
| Exit-code policy in AGENTS.md commands block: redirect-to-file then `echo $?`, never `cmd \| tail; echo $?`                                      | 🔴 `TODO` | Low    | 5m     | 00-53 §d1 (3 in-session instances of the known trap)                        |
| Goroutine window on probe-construction failure: clarifying comment in serve.go or ROADMAP conditional (embedder-only; never a code change today) | 🔴 `TODO` | Low    | 5m     | 00-53 §b6; serve.go:62-65 ordering                                          |
| `X-Content-Type-Options: nosniff` baseline header wrap in serve.go (5-line stdlib middleware; no httputil adoption needed) | 🔴 `TODO` | Low    | 5m     | docs/review/libs/httputil.md G1                                            |
| Bump go-sse v0.6.2 → v0.7.0 (additive) and go-datastar v0.6.x → v0.7.0/static v0.8.0 in lockstep (v0.7 `ScriptTagVersioned` fixes a 24h stale-JS-client window) | 🔴 `TODO` | Low    | 30m    | docs/review/libs/go-sse.md + go-datastar.md version currency               |
| go-datastar de-dup: replace hand-rolled `ringStore` with `datastar.MemoryStore`; use `ScriptHandlerWith` ETag/304 contract for `/datastar.js` | 🔴 `TODO` | Low    | 1h     | docs/review/libs/go-datastar.md                                            |
| Add `--format` flag to the fleet-table CLI path (go-output already in graph with 16 formats via cmdguard → samber-do-auditlog; ~25 lines, zero new deps) | 🔴 `TODO` | Low    | 30m    | docs/review/libs/go-output.md                                              |
| cmdguard follow-ups: cmdguard-lint CI gate; `pkg/testutil.AssertDescribeContract` at next tag; upstream the `Nd` day-suffix duration codec | 🔴 `TODO` | Low    | 1h     | docs/review/libs/cmdguard.md                                               |
| Use `health.Off` for the unconfigured `cache-disk` check instead of conditional registration | 🔴 `TODO` | Low    | 10m    | docs/review/libs/go-health.md                                              |
| templ-components cheap wins: `errorpage` for default 404s, `layout.ThemeToggle`; migrate off the Tailwind Play CDN (known CSP debt); fix view.templ:46 embedlit hint | 🔴 `TODO` | Low    | 2h     | docs/review/libs/templ-components.md                                       |
| Citation-check script (verify every file:line cite in docs/review/libs reports) + run over all 20 reports + fix stale cites | 🔴 `TODO` | Med    | 1h     | 07-41 status §e1; plan T10                                                 |
| Screenshot docs/review/libs/index.html into docs/assets + link from index (firefox permission needed) | 🔴 `TODO` | Low    | 10m    | 07-41 status §b1; plan T10.4                                               |
| Link docs/review/libs/ from README docs section (discoverability) | 🔴 `TODO` | Low    | 5m     | 07-41 status §c4; plan T12.2                                               |
| Utilization reports for the two used-but-unlisted libs: go-branded-id, go-sqlitestore | 🔴 `TODO` | Low    | 1h     | 07-41 status §c2; plan T15                                                 |
| Flip-triggers table (11 justified-no libs) → ROADMAP "conditions watched" section | 🔴 `TODO` | Low    | 12m    | 07-41 status §f24; plan T12.1                                              |

## Blocked (owner decisions)

| Task                                                                   | Status       | Impact | Effort | Evidence                                                |
| ---------------------------------------------------------------------- | ------------ | ------ | ------ | ------------------------------------------------------- |
| M26: overview endgame (replace vs coexist) + migration note            | 🔵 `BLOCKED` | Med    | 30m    | owner question g2 unanswered                            |
| Create `FLAKE_PAT` secrets (GitHub + Forgejo) for CI hub/flake jobs    | 🔵 `BLOCKED` | High   | 10m    | owner-only; README CI section documents the requirement |
| Forgejo repo visibility (lars/go-compliance → public?)                 | 🔵 `BLOCKED` | Low    | 5m     | manual UI action, owner-only                            |
| Push master to remotes (github mirror 37+ commits behind)              | 🔵 `BLOCKED` | Med    | 5m     | owner push policy — never push unless asked             |
| `sqlite3 ~/.local/state/buildflow/buildflow.db VACUUM` (1.03 GB)       | 🔵 `BLOCKED` | Low    | 10m    | 23-40 §f25; owner machine hygiene                       |
| Delete the expendable 0.51 GB cache DB (`~/.cache/buildflow/cache.db`) | 🔵 `BLOCKED` | Low    | 5m     | 23-40 §f26; owner data deletion, never agent-executed   |

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

| Task | Status | Impact | Effort | Evidence |
| ---- | ------ | ------ | ------ | -------- |

## Medium Impact

| Task | Status | Impact | Effort | Evidence |
| ---- | ------ | ------ | ------ | -------- |

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

## Blocked (owner decisions)

| Task                                                                   | Status       | Impact | Effort | Evidence                                                |
| ---------------------------------------------------------------------- | ------------ | ------ | ------ | ------------------------------------------------------- |
| M26: overview endgame (replace vs coexist) + migration note            | 🔵 `BLOCKED` | Med    | 30m    | owner question g2 unanswered                            |
| Create `FLAKE_PAT` secrets (GitHub + Forgejo) for CI hub/flake jobs    | 🔵 `BLOCKED` | High   | 10m    | owner-only; README CI section documents the requirement |
| Forgejo repo visibility (lars/go-compliance → public?)                 | 🔵 `BLOCKED` | Low    | 5m     | manual UI action, owner-only                            |
| Push master to remotes (github mirror 37+ commits behind)              | 🔵 `BLOCKED` | Med    | 5m     | owner push policy — never push unless asked             |
| `sqlite3 ~/.local/state/buildflow/buildflow.db VACUUM` (1.03 GB)       | 🔵 `BLOCKED` | Low    | 10m    | 23-40 §f25; owner machine hygiene                       |
| Delete the expendable 0.51 GB cache DB (`~/.cache/buildflow/cache.db`) | 🔵 `BLOCKED` | Low    | 5m     | 23-40 §f26; owner data deletion, never agent-executed   |

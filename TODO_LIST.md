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

| Task                                                      | Status    | Impact | Effort | Evidence                                                    |
| --------------------------------------------------------- | --------- | ------ | ------ | ----------------------------------------------------------- |
| M27: watch mode — re-derive on store writes, debounce     | 🔴 `TODO` | High   | 90m    | `hub/internal/cache/cache.go:221` no write hook yet         |
| M27: fix `queryLatestRunCounts` N+1 (single window query) | 🔴 `TODO` | High   | 30m    | `hub/internal/cache/cache.go:221` per-repo query loop       |

## Medium Impact

| Task                                                      | Status    | Impact | Effort | Evidence                                                            |
| --------------------------------------------------------- | --------- | ------ | ------ | ------------------------------------------------------------------- |
| Race-detector pass over hub suites                        | 🔴 `TODO` | Med    | 20m    | new concurrent code in `hub/internal/events` untested under `-race` |
| AGENTS.md: hub conventions + verified library API facts   | 🔴 `TODO` | Med    | 30m    | no `AGENTS.md` in repo                                              |
| `go mod tidy` in hub: go-daemon require sits in the indirect block | 🔴 `TODO` | Low | 5m | `hub/go.mod`; direct import in `bind.go` — tidy moves it (touches vendorHash) |

## Low Impact

| Task                                                        | Status    | Impact | Effort | Evidence                                                           |
| ----------------------------------------------------------- | --------- | ------ | ------ | ------------------------------------------------------------------ |
| Dedupe `waitForSubscriber` test helper across 2 packages    | 🔴 `TODO` | Low    | 15m    | duplicated in events + dashboard test files                        |
| Verify RelativeTime purity inside DataStar patches          | 🔴 `TODO` | Low    | 15m    | `hub/internal/dashboard/rows.templ` uses RelativeTime              |
| .gitignore note: generated `*_templ.go` files ARE committed | 🔴 `TODO` | Low    | 5m     | fleet precedent (go-health-dashboard)                              |
| Drop hub `replace` directive when core tags v0.1.0          | 🔴 `TODO` | Low    | 10m    | `hub/go.mod` replace block                                         |
| Fixture pins exactly 341 repos — document refresh ritual    | 🔴 `TODO` | Low    | 10m    | `testdata/buildflow-v1-fixture.json`, `scripts/refresh-fixture.sh` |

## Blocked (owner decisions)

| Task                                                        | Status       | Impact | Effort | Evidence                     |
| ------------------------------------------------------------ | ------------ | ------ | ------ | ---------------------------- |
| M26: overview endgame (replace vs coexist) + migration note | 🔵 `BLOCKED` | Med    | 30m    | owner question g2 unanswered |
| Create `FLAKE_PAT` secrets (GitHub + Forgejo) for CI hub/flake jobs | 🔵 `BLOCKED` | High | 10m | owner-only; README CI section documents the requirement |
| Forgejo repo visibility (lars/go-compliance → public?)      | 🔵 `BLOCKED` | Low    | 5m     | manual UI action, owner-only |

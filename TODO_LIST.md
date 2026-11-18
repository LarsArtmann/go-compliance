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

| Task                                                                                           | Status    | Impact | Effort | Evidence                                                                       |
| ---------------------------------------------------------------------------------------------- | --------- | ------ | ------ | ------------------------------------------------------------------------------ |
| Close the watch-stream response body on the success path (`discovery.Client.Watch`, bodyclose) | 🔴 `TODO` | High   | 15m    | `hub/internal/discovery/client.go:165`; `docs/status/2026-11-18_17-49_…md` §f5 |

## Medium Impact

| Task                                                                                                                                        | Status    | Impact | Effort | Evidence                                                                                                      |
| ------------------------------------------------------------------------------------------------------------------------------------------- | --------- | ------ | ------ | ------------------------------------------------------------------------------------------------------------- |
| Regression test: `?state=` filter accepts ANY valid `State` constant (table over the 4-state enum)                                          | 🔴 `TODO` | Med    | 15m    | `hub/internal/dashboard/routes.go:145` → `State.IsValid`; routes_test.go:105 covers `failing` only; 17-49 §f3 |
| Commit `.art-dupl-baseline.json` + wire `art-dupl check` as a CI/pre-commit gate (hash-pinned; replaces line-proximity directive fragility) | 🔴 `TODO` | Med    | 30m    | no baseline in tree; AGENTS.md Dedup Audit; 17-49 §f4                                                         |

## Low Impact

| Task                                                                                                               | Status    | Impact | Effort | Evidence                                                             |
| ------------------------------------------------------------------------------------------------------------------ | --------- | ------ | ------ | -------------------------------------------------------------------- |
| Drop hub `replace` directive when core tags v0.1.0                                                                 | 🔴 `TODO` | Low    | 10m    | `hub/go.mod` replace block                                           |
| varnamelen: rename `r` in `runDiscoveryWatch` (or nolint)                                                          | 🔴 `TODO` | Low    | 5m     | `hub/internal/cli/refresher.go:159`; 17-49 §f8                       |
| `queryLatestRecords` 3-slice return → named result struct                                                          | 🔴 `TODO` | Low    | 15m    | `hub/internal/cache/cache.go:209`; 17-49 §f25                        |
| Dead second `err != nil` check (nilness) + `infertypeargs` cleanup in app.go                                       | 🔴 `TODO` | Low    | 15m    | `hub/internal/cli/app.go:98`, `:26`; 17-49 §f6/f9                    |
| Fix or file the art-dupl HTML template leak (self-scan accept comments leak into foreign reports) (repo: art-dupl) | 🔴 `TODO` | Low    | 30m    | art-dupl v0.9.1 observed 2026-11-18; 17-49 §f14                      |
| Record the gofmt doc-comment vs `//tool:` directive lesson in references/lessons.md (repo: crush-config)           | 🔴 `TODO` | Low    | 10m    | AGENTS.md Dedup Audit section; 17-49 §f19                            |
| Verify each of the dedup session's 5 heuristic daemon commits compiles                                             | 🔴 `TODO` | Low    | 10m    | 17-49 §f21 (fold.go helper and adapter.go usage were separate edits) |

## Blocked (owner decisions)

| Task                                                                | Status       | Impact | Effort | Evidence                                                |
| ------------------------------------------------------------------- | ------------ | ------ | ------ | ------------------------------------------------------- |
| M26: overview endgame (replace vs coexist) + migration note         | 🔵 `BLOCKED` | Med    | 30m    | owner question g2 unanswered                            |
| Create `FLAKE_PAT` secrets (GitHub + Forgejo) for CI hub/flake jobs | 🔵 `BLOCKED` | High   | 10m    | owner-only; README CI section documents the requirement |
| Forgejo repo visibility (lars/go-compliance → public?)              | 🔵 `BLOCKED` | Low    | 5m     | manual UI action, owner-only                            |
| Push master to remotes (github mirror 37+ commits behind)           | 🔵 `BLOCKED` | Med    | 5m     | owner push policy — never push unless asked             |

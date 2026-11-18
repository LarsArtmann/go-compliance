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

| Task                                                                                                                                                                                                                         | Status    | Impact | Effort | Evidence                                                             |
| ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------- | ------ | ------ | -------------------------------------------------------------------- |
| Drop hub `replace` directive when core tags v0.1.0                                                                                                                                                                           | 🔴 `TODO` | Low    | 10m    | `hub/go.mod` replace block                                           |
| `queryLatestRecords` 3-slice return → named result struct                                                                                                                                                                    | 🔴 `TODO` | Low    | 15m    | `hub/internal/cache/cache.go:209`; 17-49 §f25                        |
| Fix or file the art-dupl HTML template leak (self-scan accept comments leak into foreign reports) (repo: art-dupl)                                                                                                           | 🔴 `TODO` | Low    | 30m    | art-dupl v0.9.1 observed 2026-11-18; 17-49 §f14                      |
| Record the gofmt doc-comment vs `//tool:` directive lesson in references/lessons.md (repo: crush-config)                                                                                                                     | 🔴 `TODO` | Low    | 10m    | AGENTS.md Dedup Audit section; 17-49 §f19                            |
| Verify each of the dedup session's 5 heuristic daemon commits compiles                                                                                                                                                       | 🔴 `TODO` | Low    | 10m    | 17-49 §f21 (fold.go helper and adapter.go usage were separate edits) |
| Re-prove the full battery (`-race` + lint both modules + `nix flake check`) — hub `-race` suite green again 19:39 session after Rows()/app.go fixes; core suite, core lint, and flake check still unproven since those edits | 🔴 `TODO` | Low    | 15m    | `docs/status/2026-11-18_20-02_samber-do-health-review-status.md` §c  |
| Verify STACK.md + CONTRIBUTING.md claims (inventoried, not verified this pass)                                                                                                                                               | 🔴 `TODO` | Low    | 10m    | STACK drift-check PASS was last proven 2026-11-18 13:31              |
| Probe observability: log the startup-latch pass + adopt `WithEvaluationHook` to slog check transitions (warn↔pass flapping becomes visible)                                                                                  | 🔴 `TODO` | Low    | 30m    | go-health v0.10.0 seams; 20-02 status §f19-20                        |
| Port two lessons to crush-config `references/lessons.md` (needs a commit there): never pipe when the exit code matters; version-pin checkout reads against the compiled dep tag                                              | 🔴 `TODO` | Low    | 10m    | 20-02 status §d1-d2, §e1-e2                                          |
| Row-by-row DOMAIN_LANGUAGE.md verification (spot-checked only this pass)                                                                                                                                                     | 🔴 `TODO` | Low    | 15m    | terms sampled against code; not every row checked                    |
| Annotate 13-31 report §f items 2–7 (done at, evidence = 15-00 report) to unblock the archive sweep                                                                                                                           | 🔴 `TODO` | Low    | 10m    | unstruck-but-done items are invisible archive debt                   |
| Reconcile the 42-vs-40 annotate count (15-00 report claim vs docs-gates struck count)                                                                                                                                        | 🔴 `TODO` | Low    | 5m     | 15-00 §a9 vs `docs-gates run` annotate line                          |

## Blocked (owner decisions)

| Task                                                                | Status       | Impact | Effort | Evidence                                                |
| ------------------------------------------------------------------- | ------------ | ------ | ------ | ------------------------------------------------------- |
| M26: overview endgame (replace vs coexist) + migration note         | 🔵 `BLOCKED` | Med    | 30m    | owner question g2 unanswered                            |
| Create `FLAKE_PAT` secrets (GitHub + Forgejo) for CI hub/flake jobs | 🔵 `BLOCKED` | High   | 10m    | owner-only; README CI section documents the requirement |
| Forgejo repo visibility (lars/go-compliance → public?)              | 🔵 `BLOCKED` | Low    | 5m     | manual UI action, owner-only                            |
| Push master to remotes (github mirror 37+ commits behind)           | 🔵 `BLOCKED` | Med    | 5m     | owner push policy — never push unless asked             |

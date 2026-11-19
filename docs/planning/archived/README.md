# Archived execution plans

Closed history: all three go-compliance execution plans ran end-to-end and
every task row carries an inline verdict (`done — <evidence>` / routed / NOT-DO).
Do not reopen them; the living backlog is TODO_LIST.md and the decision
context lives in the status reports each plan links.

<!-- check-rows:ledger -->
| Archived plan                                             | Closed     | Why archived                                                                                                                                         |
| --------------------------------------------------------- | ---------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| 2026-11-17_20-01_go-compliance-pareto-execution-plan.md   | 2026-11-19 | All 27 medium + 109 fine tasks executed (106 tree-verified by the 15-00 final-pass audit; M26 pair + Forgejo-visibility routed to TODO blocked rows) |
| 2026-11-18_20-29_integrity-first-pareto-execution-plan.md | 2026-11-19 | All 19 WPs executed per the 23-09 report (g1-g3 decided; WP-16 tag shipped 77881b3; replace-drop + owner desk routed to TODO rows)                   |
| 2026-11-19_07-55_post-audit-pareto-execution-plan.md      | 2026-11-19 | All waves + T-packages executed per the 14-43/15-50/16-12 reports (OG owner gate + T21 replace-drop routed to TODO/ROADMAP)                          |

Archive rule (docs-health skill): a plan lands here only after every task row
carries an inline verdict and both row gates (`check-rows --strict`,
`bare-rows`) ran clean before the `git mv`.

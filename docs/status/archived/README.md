# Archived status reports

Closed history: every open item in these reports carries an inline verdict
(`done at` / `Won't implement` / `routed`); do not reopen them. The living
backlog is `TODO_LIST.md`; context for decisions lives in the report bodies
below and the commit that archived them.

<!-- check-rows:ledger -->

| Archived report                                                  | Closed     | Why archived                                                                                                                                             |
| ---------------------------------------------------------------- | ---------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 2026-11-18_13-31_m18-m25-closed-m27-midflight.md                 | 2026-11-19 | M18-M25 closed + M27 midflight; all items struck (final one routed to the owner gate OG.2), completion confirmed by the 15-00 report                     |
| 2026-11-18_15-00_final-pass-plan-complete.md                     | 2026-11-19 | Final pass complete; zero open §f items at archive time (the 11-17 plan's M27/M18 items closed per this report)                                          |
| 2026-11-17_19-57_go-compliance-stack-selection-scaffold.md       | 2026-11-19 | Scaffold-session report; every §b/§c/§f item closed by the M-series build (tree-verified by the 15-00 audit); g2/visibility routed to TODO blocked rows  |
| 2026-11-17_20-44_go-compliance-m1-m10-progress.md                | 2026-11-19 | M1-M10 session; §f build tasks all shipped (15-00 audit), M26 + visibility routed, fixture-pin guidance lives in scripts/refresh-fixture.sh              |
| 2026-11-18_06-47_go-compliance-m10-m20-session-report.md         | 2026-11-19 | M10-M20 session; §f tasks shipped, M26/tag-train remainder routed to TODO rows                                                                           |
| 2026-11-18_07-19_go-compliance-m17-m18-flake-bootstrap.md        | 2026-11-19 | M17-M18 arc; M18 closed per 13-31, M19-M27 shipped, the go-nix-helpers upstream fix routed to ROADMAP (owner-gated push)                                 |
| 2026-11-18_11-21_m18-hash-flip-m19-lint-midflight.md             | 2026-11-19 | Mid-flight M18/M19 report; every item closed (vendorHash flip root-caused → _core_only; hub lint 0; CI live); CI-drift-gate ideas resolved NOT-DO/routed |
| 2026-11-18_17-49_art-dupl-dedup-audit-status.md                  | 2026-11-19 | Dedup audit; all findings closed (gates WP-04/09/10, calibration ideas KEEP RAW in ROADMAP, leak NOT REPRODUCIBLE)                                       |
| 2026-11-18_18-06_docs-health-audit-status.md                     | 2026-11-19 | Prior docs-health AUDIT; its open rows either shipped (WP-12) or carry scoped verdicts (ROADMAP KEEP RAW / NOT-DO / TODO rows)                           |
| 2026-11-18_20-02_samber-do-health-review-status.md               | 2026-11-19 | samber/do + health review; g1-g3 RESOLVED same day, §f items closed by the integrity-first pass, conditionals closed as moot/documented                  |
| 2026-11-19_07-41_library-utilization-fleet-audit-status.md       | 2026-11-19 | 20-library fleet audit; all 37 §f items executed (T1-T18) or routed (ROADMAP owner decisions / fleet waves / TODO rows)                                  |
| 2026-11-19_14-43_post-audit-pareto-waves-1-3-execution-status.md | 2026-11-19 | Waves 1-3 execution; every §f item closed by Wave 4 + T14 (15-50/16-12), owner questions routed to ROADMAP/TODO                                          |
| 2026-11-19_15-50_post-audit-pareto-plan-fully-executed.md        | 2026-11-19 | Plan-fully-executed report; §f remainders are TODO_LIST rows, owner questions routed to ROADMAP owner decisions                                          |

Archive rule (docs-health skill): a report lands here only when ANNOTATE
resolved every item; the docs-gates cadence gate keeps the unarchived count
under its threshold (15). The 2026-11-19 pass archived 11 reports + all 3
planning docs (see docs/planning/archived/README.md) after every numbered
item carried an inline verdict and both row gates (`check-rows --strict`,
`bare-rows`) ran clean on every file BEFORE the `git mv`.

# Library Utilization Audit notes (2026-11-19 run)

Relocated from AGENTS.md 2026-11-20 (canonical-gate line-count rule); content unchanged.

`docs/review/libs/` holds 22 per-library utilization reports + `index.html`
(fleet run: 5 sub-agent groups + T15 completion reports for go-branded-id
and go-sqlitestore, both ADOPTED). Verdict line: go-health / go-sse /
go-error-family(core) / go-branded-id / go-sqlitestore ADOPTED;
go-filewatcher ADOPTED 2026-11-19 (plan T3, was the audit's
NOT-ADOPTED-but-should); templ-components recovering after the T9 pass
(errorpage/ThemeToggle/compiled CSS); cmdguard/go-datastar/go-output/
go-daemon UNDERUSED (go-daemon surface widened by the waves); 9
NOT-ADOPTED-justified (each with a recorded flip trigger in its report).
Actionable findings routed to TODO_LIST (4 High/Med + 7 Low rows).
Re-review triggers: a NOT-ADOPTED lib's flip trigger fires, a pinned dep
moves a major, or new fleet-wide consumers change the leverage math. For
NOT-adopted libs the report score is counterfactual (100 = zero unmet need).
Citation integrity: `scripts/check-citations.sh` verifies every
`file:line` cite in the reports against the tree (ratchet via
`docs/review/libs/.citations-baseline` — rerun after ANY refactor that
moves cited code; new drift fails, baselined drift prints ACCEPTED).
Tag-watch hook: when a reviewed lib (starting with go-daemon) cuts a new
tag, append an executed-addendum to its report the next time it is
touched — the reports pin versions in their verdicts and go stale
silently otherwise.

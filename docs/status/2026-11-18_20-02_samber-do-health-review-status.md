# Status: samber/do + Health Architecture Deep Review (session of 2026-11-18 19:30–20:02)

**Scope line — what this report VERIFIED vs NOT:**
Verified this session, with commands/artifacts: every hub wiring/liveness file read end-to-end; samber/do traced to zero direct usage (`go mod graph`, `go mod why`); go-health v0.10.0 semantics read from source — and this closing check confirmed the sibling checkout (`v0.10.0-4-g2c9c743`) differs from the pinned tag only in 4 docs/chore commits, `git diff --stat v0.10.0..HEAD` on all cited files is empty, so every citation in the review report is exact for the compiled version; hub build + vet + full `-race` suite green; gofmt clean; lint delta zero-new-findings. NOT verified: core-module test run, `nix flake check`, FEATURES.md cross-check, LSP diagnostics cleared (nilness warning persisted in the panel after the fix — file content proven correct). This report covers ONLY this session's work, per instruction.

---

## a) FULLY DONE

1. **Skill discipline:** `samber-do-best-practices` loaded with all three references read completely before any edit.
2. **samber/do verdict proven, not assumed:** zero `do.*` calls repo-wide; transitive paths = cmdguard/v5, go-health, samber-do-auditlog (graph quoted in the report). STACK.md:50's recorded decision found and honored in the verdict.
3. **Wiring/lifecycle map:** main, app, serve, bind, flags, health, refresher, events (incl. Shutdown/SubscribeFeed), cache (Store/OnWrite/Close), discovery client, dashboard call sites — all read; shutdown ordering and ownership tabled with file:line.
4. **go-health internals verified:** Probe.Start (critical-name validation, one-way Shutdown, two-phase MarkShuttingDown), NewChecks injector-free contract, Status/Check wire shapes — from source, not FEATURES.md claims alone.
5. **Defect #1 found, root-caused, FIXED:** `Hub.Rows()` could never return nil (`make([]T,0) != nil`, proven with an executable check) → the `evidence` readiness/startup gate and the SSE snapshot gates were inert. Fixed at the root (events.go:134); nil-before-first-publish vs empty-published-fleet semantics documented as load-bearing.
6. **Regression pins:** new events spec (nil/non-nil/empty split) + new `health_internal_test.go` (gate fail→pass through the REAL probe's `Evaluate`).
7. **Defect #2 FIXED:** unreachable duplicate `err != nil` in `queryEvidence` removed (gopls nilness).
8. **DO-1 → DO-6 equivalence audit:** all eight rules (incl. service-locator and container-storage smells) pass by construction — table with evidence in the report.
9. **Deliverable written:** `docs/architecture-understanding/2026-11-18_19-39_samber-do-health-architecture.md` — verdicts, scorecard, 8 ranked findings, adopt-or-not decision with revisit triggers, verification log.
10. **Docs updated:** TODO_LIST (+4 rows, 1 row split), CHANGELOG (2 Fixed entries), AGENTS.md (health-surface contract, DI posture, known gap).
11. **Verification battery:** hub `go build`, `go vet`, full `go test -race -count=1 ./...` sequential per policy — all green; gofmt clean on changed files.

## b) PARTIALLY DONE

1. **P1 refresh-freshness guard** (hung `QueryEvidence` wedges the loop with health green): designed completely (per-refresh timeout + `lastRefreshSuccess` + non-critical warn check, report §5.3), TODO_LIST row created — **not implemented**, deliberately: classification and thresholds are an owner call (question g1).
2. **Lint posture:** 16 pre-existing golangci findings documented and sourced (none from this diff) — triage/fix/config decision not made (question g3).
3. **go-health citation honesty:** initial reads came from an unpinned working checkout; closed at 20:02 by diffing v0.10.0..HEAD over every cited file (empty). The miss is recorded below, the risk is retired.
4. **docs-gates/FEATURES.md sync after the behavior fix:** the readiness gate changed from decorative to real; whether FEATURES.md or older status reports claimed behavior that was actually inert was NOT cross-checked this session.

## c) NOT STARTED

1. Core-module test run (module untouched by the diff, but no proof run either).
2. `nix flake check` after the session's Go-file edits (hub-only files; vendorHash should be unaffected by the `_core_only` isolation — unproven).
3. All pre-existing TODO_LIST rows (bodyclose, `?state=` table test, art-dupl baseline, etc.) — out of scope, untouched.
4. Discovery/SSE observability checks (designed in report §5.4/§5.5, rows exist, no code).

## d) TOTALLY FUCKED UP (all caught in-session; none shipped)

1. **Pipe swallowed the lint exit code:** first run was `golangci-lint … | tail; echo $?` → reported "exit: 0" and I briefly treated lint as passing. Caught one step later, rerun with output redirection → true exit 1 with 16 findings. Lesson: when the exit code is the point, redirect; never pipe.
2. **Cited an unpinned source:** read go-health from `/home/lars/projects/go-health` without first pinning it to the v0.10.0 the hub actually compiles. A verify-external-claims discipline slip — the citations happened to be exact (proved at 20:02), but that was luck, not method.
3. **Let a red flag linger cosmetically:** the gopls nilness warning kept rendering after the fix; I proved the file correct but never restarted the LSP, so the diagnostics panel lied for the rest of the session.

## e) WHAT WE SHOULD IMPROVE

1. **Exit-code discipline:** no `cmd | tail` when `$?` matters (d1). Cheap habit, expensive failure mode: a false "lint green" in a report is exactly the unverified-gate class this repo's own AGENTS.md warns about.
2. **Version-pin external reads:** when auditing a pinned dependency, read `$GOMODCACHE/…@vX.Y.Z` or first `git describe` the checkout (d2). One command turns a citation from "probably" into "exact".
3. **Post-edit battery completeness:** after Go edits, the AGENTS.md battery includes core + flake check, not just the touched module (c1/c2). I optimized for the diff's blast radius and skipped the cheap full proof.
4. **LSP hygiene:** restart the client after edits when a diagnostic provably contradicts file content (d3).
5. **Test depth on gates:** the new gate test exercises `Evaluate` directly; the not-ready window is still unprovable over real HTTP because serve aborts on boot-refresh failure — a design fact worth a deliberate decision, not an accident (f-item 12).
6. **Empty-fleet semantics were decided unilaterally by me:** I made "0 published repos = ready" the contract. Plausible, tested, documented — but it is a domain judgment I could not derive from code (question g2).

## f) Next things (ranked; ✎ = already a TODO_LIST row, rest are new fuel)

1. ✎ Refresh-freshness guard (per-refresh timeout + warn check) — P1, design ready.
2. ✎ Lint posture decision for the 16 standing findings.
3. Run core suite + `nix flake check` to close the post-edit battery.
4. Verify FEATURES.md + prior status reports for claims about the (previously inert) readiness gate; correct where they oversold.
5. ✎ bodyclose fix in `discovery.Client.Watch`.
6. ✎ Non-critical `discovery` health check (daemon-down = standing warn).
7. ✎ Surface `Hub.Dropped()`/`SubscriberCount()` in readiness payload or a check.
8. Decide tolerant-boot mode: should serve start (and report not-ready) when the initial refresh fails, instead of exiting? Makes the readiness window real over HTTP.
9. Empty-fleet readiness semantics: confirm or overturn "0 repos = ready" (g2).
10. ✎ `?state=` filter table test over all four states.
11. ✎ art-dupl baseline commit + CI gate.
12. Add a `/readyz` payload contract test (JSON shape stability for federation consumers).
13. ✎ infertypeargs cleanup app.go:26.
14. ✎ varnamelen `r` in runDiscoveryWatch; makezero ×3; ireturn in resolveSource — fold into the lint triage.
15. Consider `go-health` federation/aggregate adoption once a second probe-owning service exists (YAGNI today — do not install).
16. Record the version-pin-external-reads lesson in crush-config `references/lessons.md` (cross-project; needs a commit there — session writes are read-only by design).
17. Record the pipe-swallows-exit-code lesson alongside it (same file, same commit).
18. ✎ `queryLatestRecords` named-result struct (cache readability).
19. Startup-latch observability: log when `/startupz` latches (one line at boot; cheap operator signal).
20. Probe `WithEvaluationHook` adoption to feed slog on check transitions (warn↔pass flapping becomes visible) — opt-in, zero-risk seam.
21. README: document the health surface (`/healthz` `/readyz` `/startupz` `/version`) for embedders if not already there.
22. ✎ Drop hub `replace` when core tags v0.1.0.
23. ✎ M26 overview endgame; FLAKE_PAT secrets; Forgejo visibility; master push (all owner-blocked).
24. Consider a `Store.OnWrite` fan-out if a second consumer ever appears (documented single-slot today).
25. Re-run `stack-drift-check.sh` after any of the above dependency-adjacent changes.
26. Next full-code-review pass should visit the files this session only read (table.go, view.go, rows.templ internals) — reviewed as call sites, not line-by-line.
27. Pin the `evidence` check name as a typed constant shared with tests (it is one already — `evidenceCheckName`; keep it that way when adding checks).
28. If tolerant-boot lands (item 8), add an E2E: source fails at boot → `/readyz` 503 → source recovers → 200.

## g) Questions I cannot answer myself

1. **Freshness check classification:** when the refresh loop goes stale, should `/readyz` degrade to **warn (stays 200, visible in payload)** — my recommendation, matching the "serve last good state" posture — or **fail (503, drain the hub)**? And is ~2× the 5-minute safety net (10 min) the right threshold?
2. **Empty-fleet semantics:** is a published fleet of **zero repos** genuinely healthy (my chosen contract: ready), or suspicious enough to degrade readiness (it can mean discovery + evidence both went dark while the process hums along)?
3. **Lint posture:** for the 16 standing findings — fix the class (nolint-with-reason on the idiomatic globals, close the bodyclose closure shape), disable the offending linters, or adopt a baseline file? Owner taste; a permanently red lint gate guards nothing either way.

**RESOLVED same day (2026-11-18 ~22:15, execution pass; owner green-lit autonomous execution of the whole plan):**
g1 = **warn, stays 200, threshold 10 min** (recommendation adopted; `refreshFreshnessWarn`, pinned in health_internal_test.go). g2 = **empty published fleet IS ready** (chosen contract confirmed as final; pinned + documented in AGENTS.md/FEATURES.md). g3 = **real fixes + nolint-with-reason, NO baseline file** — both modules at 0 findings (see the 22:xx CHANGELOG entries; note: the bodyclose "closure shape" fix turned out to be a false positive — the stream goroutine owns the close via defer, now documented at the call site).

---

_Written 2026-11-18 20:02. Auto-commit daemon owns the commit; harness forbids manual commits. Waiting for instructions._

# Post-Audit Pareto — Brutal Self-Review + Full Execution Status

**Date:** 2026-11-19 16:09 CET · **Session:** the resume session that
finished Wave 3 and executed all of Wave 4 + T14 (the 15:50 report covers
WHAT shipped; this one covers what I did WRONG, forgot, and should improve).
**Scope of this report:** this session only, judged against its own work.

## a) FULLY DONE (this session, all gated green)

~~1. templ `fmt -fail` gate green; T9.5 embedlit adjudicated a templ-LSP FALSE POSITIVE with a scratch-program proof (Go forbids elided embedded-field literal types); AGENTS.md carries the verdict.~~ done — Verified 2026-11-28: the verdict is in the tree at AGENTS.md:88-90 ("The templ LSP `embedlit` hint is a false positive: keep `BaseProps: utils.BaseProps{...}` explicit"); the gate is the CI job "templ fmt and generate gate" (ci.yml:131)
~~2. vendorHash rituals ×2 (go-output/toolchain window, then errorpage window) — both converged first paste; `nix flake check` green after each; final `--all-systems` exit 0 (x86_64 + aarch64, 6 checks).~~ done — Verified 2026-11-28: git log pins the day's flake.nix ritual commits ab77ef2 (15:07), 1292ccf (15:11), 17171ff (15:27); the ritual is codified AGENTS.md:200-204 ("go.mod changes flip `vendorHash` — tidy per module, then `buildflow -s nix-hash-fix --fix`, and finish with `nix flake check`"); the local `nix flake check` runs themselves are not re-runnable in this session
~~3. SECURITY.md response-time rewording → securitymd 0 findings.~~ done — Verified 2026-11-28: SECURITY.md:23-25 carries the reworded response-time commitment ("acknowledged within a week and receives a fix or a mitigation plan before any public disclosure")
~~4. T11.3 threshold pins (10-min warn ×2; `cacheDiskMinFreeBytes` const hoisted + 128 MiB pin), T11.4 goroutine-window comment, AGENTS command rows (core -race, devShell build, templ fmt gate, exit-code probe).~~ done — Verified 2026-11-28: hub/internal/cli/health.go:45 `const cacheDiskMinFreeBytes = 128 << 20` (used at :155); AGENTS.md:19-20 -race command rows with CI parity (ci.yml:37-39, 60); goroutine-window rule AGENTS.md:172-175 ("Never sleep to 'let a watch arm' ... `armWatch` in `watcher_internal_test.go`")
~~5. Re-verifies: go-structure-linter `-race` finding cleared; art-dupl 0 new clones (held throughout).~~ done — Verified 2026-11-28: -race is the CI build+test gate (ci.yml:39, 60), green since the first real runs (CHANGELOG.md:90); art-dupl is the CI dupl-gate ratchet over `.art-dupl-baseline.json` (ci.yml:100; acceptance policy DECIDED per ROADMAP.md:55-62; AGENTS.md:184-186); structure-linter items are decided warning triage (AGENTS.md:208-212)
~~6. T10: `scripts/check-citations.sh` + `.citations-baseline` ratchet + 11 report drift addenda + `citation-gate` CI job + AGENTS tag-watch hook. 414+ cites mechanically verified per run.~~ done — Verified 2026-11-28: scripts/check-citations.sh runs green — checked=427 ok=427 stale=0 (my run); ratchet baseline lives at docs/review/libs/.citations-baseline (script :35); citation-gate CI job ci.yml:158-164; the wave-drift addenda are recorded in CHANGELOG.md:78; the AGENTS tag-watch hook line is gone from the slimmed AGENTS.md (its :198 now carries the library-audit pointer)
~~7. T8: `cmdguard usage gate` CI job (verified clean via direct CLI run); AssertDescribeContract confirmed master-only → deferred to next tag.~~ done — Verified 2026-11-28: cmdguard usage gate job ci.yml:166-181 (shallow clone of pinned `lint/v0.3.0`; the published-module replace bug documented at ci.yml:177-178); AssertDescribeContract still awaits the next tag (TODO_LIST.md:40 Low row; zero uses under hub/ — grep 2026-11-28)
~~8. T9: errorpage 404 (exact-root `/{$}` + catch-all + noindex), ThemeToggle/ThemeScript dark mode, Play CDN + inline tailwind.config DELETED → compiled CSS pipeline (regen-ui.sh, library inventory from go.mod truth, content-hash `/app.css` ETag/304/immutable, css-freshness CI gate, devShell tailwindcss_4) + CSP/patch-budget regression specs.~~ done — Verified 2026-11-28: hub/internal/dashboard/routes.go:71 `page = "/{$}"`; ThemeToggle/ThemeScript in view.templ + app.css; Play CDN banned by AGENTS.md:91-93 ("a compiled Tailwind artifact, never the Play CDN"); scripts/regen-ui.sh + content-hash versioning (styles.go:25-30, routes.go:150); css-freshness gate ci.yml:204; devShell tailwindcss_4 (flake.nix:109, 208); CSP headers in the FEATURES.md:45 row
~~9. T15: go-branded-id (ADOPTED 90) + go-sqlitestore (ADOPTED 85) reports; index.html → 22 reports, flipped verdicts reflected.~~ done — Verified 2026-11-28: docs/review/libs/ holds exactly 22 .md reports including go-branded-id.md and go-sqlitestore.md; docs/review/libs/index.html links 22 reports; README.md:188 "22 audited"
~~10. T18: 7 audit-all items struck with evidence; annotate-gate spacing fix; lychee 0 findings (one real stale link depth fixed); the "4 path-excluded files" concern moot under current config.~~ done — Verified 2026-11-28: the 18-06 report sits archived and fully struck (docs/status/archived/README.md:20 manifest row); the annotate gate runs in the docs-gates battery; the lychee-0-findings and 4-path-excluded halves are owner-machine BuildFlow scanner results, not re-runnable here — the digest still tracks "Re-run `buildflow -s lychee` + `-s vulnix`" and "Confirm lychee's 4 path-excluded files are intentional"
~~11. T13: 2 fully-resolved reports archived + README manifest; cadence green at 13.~~ done — Verified 2026-11-28: docs/status/archived/README.md:12-13 manifest rows close 13-31 and 15-00 ("Closed 2026-11-19", the T13 pass); "cadence green at 13" was true at report time
~~12. T17: io anomaly explained from `io_metrics` (files_hashed tracks tree churn) — NOT-DO verdict recorded.~~ done — Verified 2026-11-28: the NOT-DO verdict is on record twice — archived 14-43 report :207-208 and 23-40 §f30 (archived file :131: "files_hashed correlates with files changed since the previous run ... No ratchet defect; no action")
~~13. T14: both modules build/vet/lint-0/race-green; docs-gates 5/5; govulncheck 0 both modules; TODO_LIST reconcile (27 rows struck, CHANGELOG appended); final status report written.~~ done — Verified 2026-11-28: CHANGELOG.md `## [0.1.0] - 2026-11-19` (:109) carries the wave rows; govulncheck triage at AGENTS.md:208-212 ("the run shows zero called vulnerabilities"); CI build+test/lint with -race parity green since first runs (ci.yml:37-39, 60; CHANGELOG.md:90); the 27-row TODO strike and 5/5 battery were that session's state
~~14. Post-hoc fix-on-sight (found BY this self-review): STACK.md wiring row missing go-error-family + go-filewatcher (stack-drift-check was RED — would have failed CI on next push; now PASS), AGENTS.md audit counts (20→22 reports, 10→9 NOT-ADOPTED), FEATURES.md rows (--format, exit codes, MemoryStore, probe-first, dark mode/404/CSS, dir-watch).~~ done — Verified 2026-11-28: STACK.md wiring rows exist — go-error-family (:43), go-filewatcher/v2 (:53); FEATURES.md rows: --format/exit codes (:41), MemoryStore (:44), dark mode/404/CSS (:45), probe-first (:50), dir-watch (:52); the AGENTS audit-count line was later slimmed away (README.md:188 now carries the "22 audited" count)

## b) PARTIALLY DONE

~~1. **CI end-to-end verification**: four new jobs (citation-gate, css-freshness, cmdguard-usage, templ-fmt) are actionlint-clean and locally de-risked (each command ran green here), but ZERO real runner executions. `css-freshness` has the most novel surface (nix shell on the runner + module downloads); `citation-gate` assumes `rg` exists on ubuntu-latest (it is in the runner image, but I did not pin it).~~ done — Overtaken by events, Verified 2026-11-28: real runner executions happened — CHANGELOG.md:90 ("the owner-ordered push ran the workflow for the FIRST time ... full matrix 13/13 in run 44612235355; the four newest gates now hold real green runner executions (citation-gate, css-freshness, cmdguard-usage, templ-fmt)"); the rg assumption was made moot by the grep -ERE fallback (scripts/check-citations.sh:53-60); the remaining CI debt is the docs-gates audit-all row (TODO_LIST.md:25)
~~2. **T10.4 index screenshot**: firefox blocked by the session permission hook — needs the owner (or `permissions safe firefox`).~~ done — Verified 2026-11-28: docs/assets/libs-index.png exists (2026-11-27), captured headlessly by scripts/screenshot-page.sh and linked from the audit index (11-52 report :138: "capture rendered and checked by eye") — no firefox permission was ultimately needed
3. **Upstream findings queue**: I PROVED two new upstream bugs (published **→ routed to ROADMAP "Fleet adoption waves (raw, from the 2026-11-19 utilization audit)", owner-decision row "mandate for consumer sessions filing upstream TODOs in library repos" (lists the cmdguard/lint published-go.mod replace-directive bug and the a-h/templ LSP embedlit false positive). Verified 2026-11-28: both drafts sit voice-checked in docs/drafts/ (TODO_LIST.md:40)**
   `cmdguard/lint` go.mod carries local-dev replace directives breaking
   `go run @version`; templ LSP `embedlit` suggests invalid Go) — recorded
   in repo docs, but filing them upstream is owner-mandate-gated.

## c) NOT STARTED (deliberately untouched, not in the plan's waves)

~~1. Hub `table --sort-by` flag parity (pre-existing TODO row).~~ done — Verified 2026-11-28: shipped 2026-11-20 — hub/internal/cli/flags.go:51 `--sort-by` (values findings,failures,name, tiebreak inside the state rank) + table_test.go:255-262 specs; CHANGELOG.md:81
~~2. T21.2 hub replace-drop — version-currency gated on the next core tag.~~ done — Executed at v0.2.0 instead of v0.1.1, Verified 2026-11-28: tag v0.2.0 = f906ae5 (2026-11-20 17:12 +0100, `git rev-parse v0.2.0^{commit}`); hub/go.mod:12 requires go-compliance v0.2.0 with zero replace directives; flake.nix:150 "dropped at the v0.2.0 release); no _core_only copy needed."
3. DOMAIN_LANGUAGE.md additions for the CSS-pipeline artifact terms
   (borderline: the glossary is domain language, not build tooling).

## d) TOTALLY FUCKED UP (my mistakes this session — all caught + fixed)

~~1. **STACK.md wiring drift shipped and nearly missed**: I landed the errorpage/go-error-family/go-filewatcher dep changes WITHOUT running `stack-drift-check.sh` (a documented flake/CI gate). The self-review caught it 30 minutes later; CI would have gone red on push. Root cause: I treated "nix flake check" as the whole nix gate and forgot the repo's OWN drift gate exists.~~ done — Verified 2026-11-28: the drift gate is now CI-enforced (ci.yml:149 "stack drift check") and STACK.md carries the wiring rows (go-error-family :43, go-filewatcher/v2 :53, boot contract :107-108)
~~2. **The annotate-writer mis-strike**: struck item numbers without the `@substring` disambiguator in two reports with duplicate item numbers across sections — the writer hit the FIRST match (wrong lines, wrong verdicts on them). docs-gates caught it; one repair cycle burned.~~ done — Verified 2026-11-28: both affected reports (14-43, 15-50) sit archived with correct per-item strikes (14-43 :206-209, 15-50 :68-70) and the lesson is codified AGENTS.md:193 (`N@substring` disambiguator)
~~3. **`go get errorpage@v1.9.0` downgraded the root templ-components module** to v1.9.0 (shared tag namespace) — the lockstep rule was in the loaded skill; I applied the fix immediately but should have read the tag list before the get.~~ done — Verified 2026-11-28: hub/go.mod pins templ-components v1.27.2 uniformly across all seven module requires (:28-31, :89-90) — no downgrade present
~~4. **Mangled CI checkout SHA** (`...da90b1` garbage) written into ci.yml — caught and fixed in the same minute, but I pasted a checksum by hand instead of copying it.~~ done — Verified 2026-11-28: ci.yml:171-172 carry full 40-char pinned action SHAs with version comments; a grep for the old garbage ("da90b1") in ci.yml finds nothing
~~5. **Broken-first code**: styles.go's const-then-init-reassign (consts cannot be assigned), `appCSSEVersion` typo, `"/{$" + "}"` split string; a tautological `Expect(128<<20).To(Equal(128<<20))` placeholder I started writing before doing the honest const-hoist. All caught by compile/review within minutes — but three of them were avoidable by thinking before typing.~~ done — Verified 2026-11-28: health.go:45 `const cacheDiskMinFreeBytes = 128 << 20` (the honest const-hoist shipped); styles.go:25-30 `appCSSVersion` (the typo is gone); routes.go:71 `page = "/{$}"` (single string)
~~6. **Two bash-script bugs in check-citations.sh before first green run** (`((x++))` returning 1 under `set -e`; a heredoc-mangled `do`), plus the `rg -rn`/`rg -rln` permission hook fired on me AGAIN (twice) — the same documented trap, despite the warning being in my context.~~ done — Verified 2026-11-28: scripts/check-citations.sh is shellcheck 0.11.0 clean (my run, exit 0) and green (checked=427 ok=427 stale=0, my run); the rg permission-hook trap is session folklore, not a repo defect
~~7. **A baseline row for my OWN minutes-old report**: wrote `branded.go:14-40` (file is 30 lines), baselined the drift instead of fixing my own fresh text, then reversed course. Fresh files should be FIXED, never baselined.~~ done — Verified 2026-11-28: docs/review/libs/.citations-baseline (mtime 2026-11-19 15:44) carries NO branded.go row — the minutes-old report was fixed, not baselined
~~8. **AGENTS.md/FEATURES.md left stale through the whole session** while I updated README/CHANGELOG/TODO_LIST/index — the audit-section counts (20 reports, 10 NOT-ADOPTED) and the feature rows drifted until this self-review. docs-gates routing could not catch class of staleness.~~ done — Verified 2026-11-28: the fix is in the tree — STACK.md:43,53 wiring rows; FEATURES.md:41-52 feature rows; README.md:188 "22 audited"; AGENTS.md:198 library-audit pointer

## e) WHAT WE SHOULD IMPROVE (process, from this session's scars)

~~1. **A "dep-change checklist" gate**: any session that edits go.mod must~~ done — Codified, Verified 2026-11-28: AGENTS.md:200-204 (the vendorHash/nix-hash-fix/flake-check block) + the CI "stack drift check" job (ci.yml:149) enforces the drift half on every push
   run stack-drift-check + vendorHash ritual + govulncheck as ONE block
   before declaring green (today: drift was forgotten until self-review).
~~2. **Annotate-writer discipline**: ALWAYS `@substring` keys in reports~~ done — Codified, Verified 2026-11-28: AGENTS.md:193 "Multi-section annotate keys need the `N@substring` disambiguator; run `emit-keys` and `--dry-run` first"
   with multiple numbered sections; a number-only key is a footgun.
   Worth an AGENTS.md line in the docs-health workflow.
3. **Copy pinned SHAs from the file, never from memory** (the ci.yml
   mangling); better: a YAML anchor or a single reusable step.
4. **Living-doc sweep as part of "done"**: when a session changes
   user-visible behavior, FEATURES.md + AGENTS.md counts update in the
   SAME breath as CHANGELOG — not at self-review time.
~~5. **CI jobs need a first-run observation note**: 4 new jobs, 0 runner~~ done — Overtaken by events, Verified 2026-11-28: all four jobs hold real green runner executions (CHANGELOG.md:90; FEATURES.md:60 "all with real green runner executions") — "pushed but never run" is no longer true of any gate
   executions — treat "pushed but never run" as PARTIALLY DONE honestly
   (this report does).
~~6. **Pin `rg` in the citation-gate job** (or switch the script to~~ done — Done differently 2026-11-20, Verified 2026-11-28: no pin needed — `extract_cites` prefers ripgrep and falls back to `grep -ERE` (scripts/check-citations.sh:53-60; CHANGELOG.md:83 "runner-image drift cannot break the gate and no install step is needed")
   grep-able extraction) so runner-image drift can't break the gate.
~~7. **shellcheck the two new scripts** (they run green, but are not~~ done — Verified 2026-11-28: shellcheck 0.11.0 clean on scripts/check-citations.sh (my run, exit 0); the second script (stack-drift-check.sh) is fetched by CI from the tooling repo (ci.yml:154), not in-tree
   shellcheck-verified; dprint does not cover shell).
8. **templ LSP false-positive upstream**: the embedlit proof is worth an **→ routed to ROADMAP "Fleet adoption waves (raw, from the 2026-11-19 utilization audit)", owner-decision row "a-h/templ LSP embedlit false positive" (upstream-mandate queue). Verified 2026-11-28: the in-tree FP verdict stands at AGENTS.md:88-90**
   a-h/templ issue (blocker: upstream mandate question).

## f) Top things to get done next (ordered)

~~1. Push to remotes (owner policy — 40+ commits behind on the GitHub~~ done — Done, Verified 2026-11-28: the push happened and CI ran — CHANGELOG.md:90 (first push run 44611411587 12/13 green; full matrix 13/13 in run 44612235355); the residual CI debt is the docs-gates audit-all row (TODO_LIST.md:25 High)
   mirror) and WATCH the four new CI jobs' first real runs.
~~2. `--sort-by` flag parity on hub `table` (findings/failures/name).~~ done — Done, Verified 2026-11-28: hub/internal/cli/flags.go:51 `--sort-by` (findings, failures, name — tiebreak inside the state rank) + table_test.go:255-262; CHANGELOG.md:81
~~3. Tag core v0.1.1 (current tree), then execute the hub replace-drop row~~ done — Executed at v0.2.0 instead, Verified 2026-11-28: v0.2.0 = f906ae5 (2026-11-20 17:12 +0100); hub/go.mod:12 requires core v0.2.0, no replace; flake.nix:150 "no _core_only copy needed"
   (require v0.1.1, drop `=> ../`, tidy, vendorHash ritual, simplify the
   `_core_only` postPatchExtra story).
4. cmdguard next tag: adopt `AssertDescribeContract`; re-evaluate the **→ routed to TODO_LIST Low Impact row "cmdguard follow-ups: cmdguard-lint CI gate DONE 2026-11-19 ..." (AssertDescribeContract verified on master but NOT in v5.1.0 — adopt at the next tag; Nd-codec + lint-replaces drafts voice-checked; filing waits on the OG.5 mandate). Verified 2026-11-28: zero AssertDescribeContract uses under hub/**
   lint CI recipe if upstream drops the replace directives.
~~5. Screenshot `docs/review/libs/index.html` (owner or `permissions safe~~ done — Done, Verified 2026-11-28: docs/assets/libs-index.png (2026-11-27) captured by scripts/screenshot-page.sh, linked from the audit index (11-52 report :138)
   firefox`).
6. go-humanize adopt/suppress/accept decision (gates 6 buildflow **→ routed to docs/status/owner-decisions-digest.md, row "go-humanize adopt/accept-with-reason" (linter/policy verdicts table). Verified 2026-11-28: hub/go.mod:10 still requires go-humanize v1.1.0 directly**
   error-severity findings every full run).
7. samber/lo adoption-or-suppress policy (fleet-wide). **→ routed to docs/status/owner-decisions-digest.md, row "samber/lo adopt-or-suppress" (owner verdict; ROADMAP owner queue). Verified 2026-11-28: samber/do is still a transitive dependency only (`hub/go.mod`, through cmdguard)**
8. Fleet SSE-parser canonical owner decision (ParseSSEData vs sseparse). **→ routed to ROADMAP "Fleet adoption waves (raw, from the 2026-11-19 utilization audit)", row "Owner decisions pending: fleet SSE-parser canonical owner (`daemon.ParseSSEData` vs `go-sse/sseparse`)"**
9. Upstream-TODO mandate decision — queue now holds FOUR candidates: **→ routed to ROADMAP "Fleet adoption waves (raw, from the 2026-11-19 utilization audit)", row "mandate for consumer sessions filing upstream TODOs in library repos" (queue: cmdguard Nd codec, cmdguard/lint replace-directive bug, go-datastar MarshalSignals, templ-components FilterChips, a-h/templ embedlit FP)**
   cmdguard Nd codec, go-datastar MarshalSignals determinism,
   templ-components FilterChips, cmdguard/lint replace-directive bug,
   (+ templ LSP embedlit false positive as a fifth).
10. FLAKE_PAT secrets (CI hub/flake jobs still skip without them). **→ routed to TODO_LIST Blocked (owner decisions) row "Create `FLAKE_PAT` secrets (GitHub + Forgejo) for CI hub/flake jobs" (FEATURES.md:60: hub/flake jobs skip until the secret exists)**
11. M26 overview endgame (replace vs coexist) — owner question g2. **→ routed to TODO_LIST Blocked (owner decisions) row "M26: overview endgame (replace vs coexist) + migration note" (also ROADMAP "Endgame question (owner, g2 — BLOCKED)")**
12. Repo visibility decision; VACUUM buildflow state DB (1.04 GB); **→ routed to TODO_LIST Blocked (owner decisions) rows "Forgejo repo visibility (lars/go-compliance → public?)", "`sqlite3 ~/.local/state/buildflow/buildflow.db VACUUM` (1.03 GB)", "Delete the expendable 0.51 GB cache DB (`~/.cache/buildflow/cache.db`)"**
    delete the 0.51 GB cache DB (owner machine hygiene).
~~13. `check-rows --strict` + `bare-rows` validation pass over the~~ done — Standing practice, re-verified 2026-11-28: `docs-gates rows --strict` + `--bare` over docs/status/archived/*.md + docs/planning/archived/*.md → "20 file(s) clean", exit 0 (my runs)
    remaining aged status reports as their items close (archive cadence).
~~14. shellcheck + `rg` pin hardening for `scripts/check-citations.sh` /~~ done — Done, Verified 2026-11-28: CHANGELOG.md:83 records the 2026-11-20 hardening (rg-preferred, grep -ERE fallback, shellcheck clean); re-verified today: shellcheck 0.11.0 clean + checked=427 stale=0 (my runs)
    CI job.
15. Consider verifying EXTERNAL (lib-side) citations against local
    checkouts when `~/projects/<lib>` exists (140 cites currently
    presumed).
~~16. Watch the cross-process watcher spec on CI for timing flakiness~~ done — Overtaken, Verified 2026-11-28: the watcher spec rides CI's build+test jobs (core/hub -race, ci.yml:37-39, 60), green in the first real runs (CHANGELOG.md:90 full matrix 13/13); no watcher-spec flake recorded in CI to date
    (2s budget; loosen only with evidence).
17. io_metrics ratchet: now that the churn correlation is understood,
    consider a per-run churn-normalized expectation (upstream buildflow
    idea, not this repo's to fix).
18. app.min.css + library-classes.txt (3.5 MB) churn on every **→ routed to ROADMAP "Dashboard/UI raw ideas (2026-11-19 self-review)", row "a per-bundle diff-size budget for the `app.min.css` + `library-classes.txt` churn on templ-components bumps"**
    templ-components bump — acceptable per fleet pattern, but a
    per-bundle diff-size budget would catch accidental explosions.
~~19. CSS: `errorpage`'s `text-[8rem]` etc. ride the shared artifact; a~~ done — Shipped stronger, Verified 2026-11-28: scripts/screenshot.sh renders light/dark/just-wide/narrow headlessly and fails on console errors, theme, empty-fleet table, overflow, WCAG-AA contrast, layout, pinned-header and keyboard findings (script header :3-11); the visual smoke exists well beyond a byte-gate
    visual smoke (screenshot diff) would catch unstyled-class regressions
    the byte-gate cannot see.
20. HEAD-request explicit test for `/app.css` (the handler relies on net/ **→ routed to ROADMAP "Dashboard/UI raw ideas (2026-11-19 self-review)", row "an explicit HEAD-request test for `/app.css`". Verified 2026-11-28: still no explicit HEAD test under hub/ (grep empty)**
    http's implicit HEAD body suppression; ScriptHandler handles HEAD
    explicitly — match that bar if it ever matters).
~~21. FEATURES.md: consider a row for the citation/stack/css CI gates~~ done — Done, Verified 2026-11-28: FEATURES.md:60 CI row names the 14-job matrix incl. dupl-gate, templ-fmt, citation-gate, cmdguard-usage, branded-ID, css-freshness — "all with real green runner executions"
    (infra inventory completeness).
~~22. AGENTS.md docs-health workflow line for the @substring rule (from §e2).~~ done — Done, Verified 2026-11-28: AGENTS.md:193 carries the `N@substring` rule

## g) Questions I can NOT answer myself

1. **go-humanize + samber/lo policy** (23-40 §g2, carried): adopt, **→ routed to docs/status/owner-decisions-digest.md, rows "samber/lo adopt-or-suppress" and "go-humanize adopt/accept-with-reason" (linter/policy verdicts table). Verified 2026-11-28: hub/go.mod:10 (go-humanize direct) and :115 (samber/do/v2 indirect only)**
   suppress with baseline, or accept-with-reason? Gates 6 error-severity
   findings on every buildflow full run — the ONLY thing keeping
   `buildflow --build-mode=full` from exit 0 on this repo.
~~2. **Core v0.1.1 timing** (new): v0.1.0 predates the Wave 1-3 core~~ **Superseded — superseded by docs/status/2026-11-20_12-06_v020-release-run-status.md — the question answered itself: core tagged v0.2.0 (f906ae5, 2026-11-20 17:12 +0100) and the hub replace-drop executed (hub/go.mod:12, no replace); no v0.1.1 was ever needed.**
   splits (no filestore.go, no toolchain line). Tag v0.1.1 as-is so the
   hub replace-drop can execute, or hold for a review pass / more
   changes first?
3. **Upstream-filing mandate** (carried Q3, now 5 candidates): may **→ routed to ROADMAP "Fleet adoption waves (raw, from the 2026-11-19 utilization audit)", row "mandate for consumer sessions filing upstream TODOs in library repos" (both proven findings listed; drafts voice-checked per TODO_LIST.md:40)**
   sessions file issues/PRs in library repos (cmdguard, go-datastar,
   templ-components, a-h/templ) in your name? Two of this session's
   findings (lint-module replace directives; templ embedlit suggesting
   invalid Go) are proven, reproducible, and waiting on this answer.

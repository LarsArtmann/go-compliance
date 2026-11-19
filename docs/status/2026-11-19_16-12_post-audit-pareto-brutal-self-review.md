# Post-Audit Pareto — Brutal Self-Review + Full Execution Status

**Date:** 2026-11-19 16:09 CET · **Session:** the resume session that
finished Wave 3 and executed all of Wave 4 + T14 (the 15:50 report covers
WHAT shipped; this one covers what I did WRONG, forgot, and should improve).
**Scope of this report:** this session only, judged against its own work.

## a) FULLY DONE (this session, all gated green)

1. templ `fmt -fail` gate green; T9.5 embedlit adjudicated a templ-LSP
   FALSE POSITIVE with a scratch-program proof (Go forbids elided
   embedded-field literal types); AGENTS.md carries the verdict.
2. vendorHash rituals ×2 (go-output/toolchain window, then errorpage
   window) — both converged first paste; `nix flake check` green after
   each; final `--all-systems` exit 0 (x86_64 + aarch64, 6 checks).
3. SECURITY.md response-time rewording → securitymd 0 findings.
4. T11.3 threshold pins (10-min warn ×2; `cacheDiskMinFreeBytes` const
   hoisted + 128 MiB pin), T11.4 goroutine-window comment, AGENTS command
   rows (core -race, devShell build, templ fmt gate, exit-code probe).
5. Re-verifies: go-structure-linter `-race` finding cleared; art-dupl
   0 new clones (held throughout).
6. T10: `scripts/check-citations.sh` + `.citations-baseline` ratchet +
   11 report drift addenda + `citation-gate` CI job + AGENTS tag-watch
   hook. 414+ cites mechanically verified per run.
7. T8: `cmdguard usage gate` CI job (verified clean via direct CLI run);
   AssertDescribeContract confirmed master-only → deferred to next tag.
8. T9: errorpage 404 (exact-root `/{$}` + catch-all + noindex),
   ThemeToggle/ThemeScript dark mode, Play CDN + inline tailwind.config
   DELETED → compiled CSS pipeline (regen-ui.sh, library inventory from
   go.mod truth, content-hash `/app.css` ETag/304/immutable,
   css-freshness CI gate, devShell tailwindcss_4) + CSP/patch-budget
   regression specs.
9. T15: go-branded-id (ADOPTED 90) + go-sqlitestore (ADOPTED 85) reports;
   index.html → 22 reports, flipped verdicts reflected.
10. T18: 7 audit-all items struck with evidence; annotate-gate spacing
    fix; lychee 0 findings (one real stale link depth fixed); the
    "4 path-excluded files" concern moot under current config.
11. T13: 2 fully-resolved reports archived + README manifest; cadence
    green at 13.
12. T17: io anomaly explained from `io_metrics` (files_hashed tracks tree
    churn) — NOT-DO verdict recorded.
13. T14: both modules build/vet/lint-0/race-green; docs-gates 5/5;
    govulncheck 0 both modules; TODO_LIST reconcile (27 rows struck,
    CHANGELOG appended); final status report written.
14. Post-hoc fix-on-sight (found BY this self-review): STACK.md wiring
    row missing go-error-family + go-filewatcher (stack-drift-check was
    RED — would have failed CI on next push; now PASS), AGENTS.md audit
    counts (20→22 reports, 10→9 NOT-ADOPTED), FEATURES.md rows
    (--format, exit codes, MemoryStore, probe-first, dark mode/404/CSS,
    dir-watch).

## b) PARTIALLY DONE

1. **CI end-to-end verification**: four new jobs (citation-gate,
   css-freshness, cmdguard-usage, templ-fmt) are actionlint-clean and
   locally de-risked (each command ran green here), but ZERO real runner
   executions. `css-freshness` has the most novel surface (nix shell on
   the runner + module downloads); `citation-gate` assumes `rg` exists on
   ubuntu-latest (it is in the runner image, but I did not pin it).
2. **T10.4 index screenshot**: firefox blocked by the session permission
   hook — needs the owner (or `permissions safe firefox`).
3. **Upstream findings queue**: I PROVED two new upstream bugs (published
   `cmdguard/lint` go.mod carries local-dev replace directives breaking
   `go run @version`; templ LSP `embedlit` suggests invalid Go) — recorded
   in repo docs, but filing them upstream is owner-mandate-gated.

## c) NOT STARTED (deliberately untouched, not in the plan's waves)

1. Hub `table --sort-by` flag parity (pre-existing TODO row).
2. T21.2 hub replace-drop — version-currency gated on the next core tag.
3. DOMAIN_LANGUAGE.md additions for the CSS-pipeline artifact terms
   (borderline: the glossary is domain language, not build tooling).

## d) TOTALLY FUCKED UP (my mistakes this session — all caught + fixed)

1. **STACK.md wiring drift shipped and nearly missed**: I landed the
   errorpage/go-error-family/go-filewatcher dep changes WITHOUT running
   `stack-drift-check.sh` (a documented flake/CI gate). The self-review
   caught it 30 minutes later; CI would have gone red on push. Root
   cause: I treated "nix flake check" as the whole nix gate and forgot
   the repo's OWN drift gate exists.
2. **The annotate-writer mis-strike**: struck item numbers without the
   `@substring` disambiguator in two reports with duplicate item numbers
   across sections — the writer hit the FIRST match (wrong lines, wrong
   verdicts on them). docs-gates caught it; one repair cycle burned.
3. **`go get errorpage@v1.9.0` downgraded the root templ-components
   module** to v1.9.0 (shared tag namespace) — the lockstep rule was in
   the loaded skill; I applied the fix immediately but should have read
   the tag list before the get.
4. **Mangled CI checkout SHA** (`...da90b1` garbage) written into
   ci.yml — caught and fixed in the same minute, but I pasted a checksum
   by hand instead of copying it.
5. **Broken-first code**: styles.go's const-then-init-reassign (consts
   cannot be assigned), `appCSSEVersion` typo, `"/{$" + "}"` split
   string; a tautological `Expect(128<<20).To(Equal(128<<20))`
   placeholder I started writing before doing the honest const-hoist.
   All caught by compile/review within minutes — but three of them were
   avoidable by thinking before typing.
6. **Two bash-script bugs in check-citations.sh before first green run**
   (`((x++))` returning 1 under `set -e`; a heredoc-mangled `do`), plus
   the `rg -rn`/`rg -rln` permission hook fired on me AGAIN (twice) —
   the same documented trap, despite the warning being in my context.
7. **A baseline row for my OWN minutes-old report**: wrote
   `branded.go:14-40` (file is 30 lines), baselined the drift instead of
   fixing my own fresh text, then reversed course. Fresh files should be
   FIXED, never baselined.
8. **AGENTS.md/FEATURES.md left stale through the whole session** while I
   updated README/CHANGELOG/TODO_LIST/index — the audit-section counts
   (20 reports, 10 NOT-ADOPTED) and the feature rows drifted until this
   self-review. docs-gates routing could not catch class of staleness.

## e) WHAT WE SHOULD IMPROVE (process, from this session's scars)

1. **A "dep-change checklist" gate**: any session that edits go.mod must
   run stack-drift-check + vendorHash ritual + govulncheck as ONE block
   before declaring green (today: drift was forgotten until self-review).
2. **Annotate-writer discipline**: ALWAYS `@substring` keys in reports
   with multiple numbered sections; a number-only key is a footgun.
   Worth an AGENTS.md line in the docs-health workflow.
3. **Copy pinned SHAs from the file, never from memory** (the ci.yml
   mangling); better: a YAML anchor or a single reusable step.
4. **Living-doc sweep as part of "done"**: when a session changes
   user-visible behavior, FEATURES.md + AGENTS.md counts update in the
   SAME breath as CHANGELOG — not at self-review time.
5. **CI jobs need a first-run observation note**: 4 new jobs, 0 runner
   executions — treat "pushed but never run" as PARTIALLY DONE honestly
   (this report does).
6. **Pin `rg` in the citation-gate job** (or switch the script to
   grep-able extraction) so runner-image drift can't break the gate.
7. **shellcheck the two new scripts** (they run green, but are not
   shellcheck-verified; dprint does not cover shell).
8. **templ LSP false-positive upstream**: the embedlit proof is worth an
   a-h/templ issue (blocker: upstream mandate question).

## f) Top things to get done next (ordered)

1. Push to remotes (owner policy — 40+ commits behind on the GitHub
   mirror) and WATCH the four new CI jobs' first real runs.
2. `--sort-by` flag parity on hub `table` (findings/failures/name).
3. Tag core v0.1.1 (current tree), then execute the hub replace-drop row
   (require v0.1.1, drop `=> ../`, tidy, vendorHash ritual, simplify the
   `_core_only` postPatchExtra story).
4. cmdguard next tag: adopt `AssertDescribeContract`; re-evaluate the
   lint CI recipe if upstream drops the replace directives.
5. Screenshot `docs/review/libs/index.html` (owner or `permissions safe
   firefox`).
6. go-humanize adopt/suppress/accept decision (gates 6 buildflow
   error-severity findings every full run).
7. samber/lo adoption-or-suppress policy (fleet-wide).
8. Fleet SSE-parser canonical owner decision (ParseSSEData vs sseparse).
9. Upstream-TODO mandate decision — queue now holds FOUR candidates:
   cmdguard Nd codec, go-datastar MarshalSignals determinism,
   templ-components FilterChips, cmdguard/lint replace-directive bug,
   (+ templ LSP embedlit false positive as a fifth).
10. FLAKE_PAT secrets (CI hub/flake jobs still skip without them).
11. M26 overview endgame (replace vs coexist) — owner question g2.
12. Repo visibility decision; VACUUM buildflow state DB (1.04 GB);
    delete the 0.51 GB cache DB (owner machine hygiene).
13. `check-rows --strict` + `bare-rows` validation pass over the
    remaining aged status reports as their items close (archive cadence).
14. shellcheck + `rg` pin hardening for `scripts/check-citations.sh` /
    CI job.
15. Consider verifying EXTERNAL (lib-side) citations against local
    checkouts when `~/projects/<lib>` exists (140 cites currently
    presumed).
16. Watch the cross-process watcher spec on CI for timing flakiness
    (2s budget; loosen only with evidence).
17. io_metrics ratchet: now that the churn correlation is understood,
    consider a per-run churn-normalized expectation (upstream buildflow
    idea, not this repo's to fix).
18. app.min.css + library-classes.txt (3.5 MB) churn on every
    templ-components bump — acceptable per fleet pattern, but a
    per-bundle diff-size budget would catch accidental explosions.
19. CSS: `errorpage`'s `text-[8rem]` etc. ride the shared artifact; a
    visual smoke (screenshot diff) would catch unstyled-class regressions
    the byte-gate cannot see.
20. HEAD-request explicit test for `/app.css` (the handler relies on net/
    http's implicit HEAD body suppression; ScriptHandler handles HEAD
    explicitly — match that bar if it ever matters).
21. FEATURES.md: consider a row for the citation/stack/css CI gates
    (infra inventory completeness).
22. AGENTS.md docs-health workflow line for the @substring rule (from §e2).

## g) Questions I can NOT answer myself

1. **go-humanize + samber/lo policy** (23-40 §g2, carried): adopt,
   suppress with baseline, or accept-with-reason? Gates 6 error-severity
   findings on every buildflow full run — the ONLY thing keeping
   `buildflow --build-mode=full` from exit 0 on this repo.
2. **Core v0.1.1 timing** (new): v0.1.0 predates the Wave 1-3 core
   splits (no filestore.go, no toolchain line). Tag v0.1.1 as-is so the
   hub replace-drop can execute, or hold for a review pass / more
   changes first?
3. **Upstream-filing mandate** (carried Q3, now 5 candidates): may
   sessions file issues/PRs in library repos (cmdguard, go-datastar,
   templ-components, a-h/templ) in your name? Two of this session's
   findings (lint-module replace directives; templ embedlit suggesting
   invalid Go) are proven, reproducible, and waiting on this answer.

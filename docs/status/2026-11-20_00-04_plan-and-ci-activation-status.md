# Status: Pareto Plan 21-25 + CI Activation — first real CI runs, three latent bugs, 13/13 green

**Created:** 2026-11-20 00:04 CET · author: Crush (glm-5.3-flash) planning/CI session ·
backlog basis: the 21-25 corpus-drain & CI-activation plan (13 packages / 50 fine
tasks), TODO_LIST after this session's reconcile (14 open rows), the four CI runs
this session produced.

**Scope line — what this session VERIFIED vs NOT:**

Verified with commands/artifacts: `pareto-planning` skill loaded before planning;
Step-0 errata pass over TODO_LIST (3 rows removed with evidence: importance-column
= DECIDED producer-side-forever per GC5, `nix flake check --all-systems` = DONE
per 16-12 §a2, push-master = ordered by this directive); the plan written to
`docs/planning/2026-11-19_21-25_corpus-drain-ci-activation-pareto-plan.md` (md +
mermaid — the owner's format directive overrides the skill's HTML default, same
precedent as the 2026-11-17 plan); commits `c994836` (plan + reconcile),
`01762cb`/`156610a`/`6f31054` (CI fixes), `7433adf` (CI-watch close-out); pushes
to Forgejo origin AND the GitHub mirror (both at `7433adf`; the mirror was 40+
commits behind and is now current); FOUR real CI executions watched end-to-end:
44608863751 = 10/13, 44610073442 = 12/13, 44611411587 = 12/13,
**44612235355 = 13/13 GREEN**; failed-job logs pulled with `gh run view
--log-failed` for every red job; each fix validated locally before push
(actionlint, dprint check, `go mod download` + `.Dir` resolution proofs); fleet
sibling dprint.json files read (go-branded-id, go-health-dashboard, cmdguard,
go-sqlitestore — all four use the same https plugin URLs, proving the drift);
docs-gates 5/5 and full-repo `dprint check` green at every commit point.

NOT verified: the flake truth on CI — the `nix flake check` job "passes" in 3-9s,
which is a FLAKE_PAT-guarded skip, not a real sandboxed flake build (hub/flake
truth remains proven locally only, as documented since M18); `go list -m` was
silent-empty on the cold runner and the exact silence mechanism (vs a loud
stderr) is empirically papered over, not root-caused; T2–T11 of the plan not
executed (this session's mandate was plan + commit + push — T1 executed live as
a side effect of the push).

---

## a) FULLY DONE

~~1. **Plan written per the owner's format** — `docs/planning/2026-11-19_21-25_corpus-drain-ci-activation-pareto-plan.md`: 1%/4%/20%/other-20% breakdown, 13 medium packages (≈7h agent work), 50 fine tasks (≤12m), mermaid execution graph, errata + anti-verschlimmbessern rules, ALL open TODOs included (the three §c owner items and demand-gated ROADMAP ideas are listed, not scheduled).~~ done — verified 2026-11-28: the plan file exists with the 1/4/20 breakdown and mermaid graph at docs/planning/2026-11-19_21-25_corpus-drain-ci-activation-pareto-plan.md; committed in `c994836` (2026-11-19; `git show --stat`: the plan + TODO_LIST reconcile)

~~2. **Step-0 errata with evidence** — 3 stale TODO rows removed (importance = decided-not-blocked; all-systems = done 2026-11-19; push = ordered by this directive), 4 rows minted (CI watch, corpus drain, docs-gates verify, cmdguard readiness), 1 CHANGELOG all-systems gap covered by annotating 00-53 §f4 done at the tree.~~ done — at `c994836` (2026-11-19): `git show --stat c994836` shows TODO_LIST.md rewritten (rows removed and minted) plus the 00-53 §f4 house-grammar fix. Verified 2026-11-28

~~3. **The push executed plan task T1 LIVE** — both remotes at `7433adf`; the push triggered the repo's FIRST real CI execution ever (every previous "CI green" claim had zero runner evidence — the four newest gates existed only as YAML).~~ done — at `7433adf` (2026-11-19, "CI first-run watch closes green — all 13 jobs pass"); `gh run view 44612235355` shows the run green and both remote master branches contain the commit. Verified 2026-11-28

~~4. **Three latent first-run bugs found and fixed across three fix commits:** - dprint job: `GITHUB_PATH` pointed at `$PWD/bin`; the dprint.dev installer targets `~/.dprint/bin` → exit 127. Fixed; actionlint's CWD-invocation hardened in the same step. - css-freshness: `gen-library-classes.sh` resolves templ-components via `go list -m -f '{{.Dir}}'`; the cold runner returned silent-empty. Fixed by warming ALL FIVE module zips (parent-only warm-up was my intermediate mistake, see §d1). - dprint.json: plugin array pointed at absolute local BuildFlow-store paths (`/home/lars/.local/state/buildflow/...`) — fleet drift vs all four sibling repos, which use the fleet-canonical https URLs (same versions, so formatting is byte-identical; proven by full-repo `dprint check` green).~~ done — at `01762cb`, `156610a`, `6f31054` (all 2026-11-19; `git show --stat`: ci.yml PATH + five-zip warm-up, the 00-53 strike rewrite, dprint.json https plugin URLs). Verified 2026-11-28

~~5. **A formatter-unstable strike caught before it could distort history**: my 00-53 §f4 annotation struck the LIST MARKER (`~~4. …~~`); dprint's markdown parser then wanted to indent the line as a paragraph and RENUMBER items 5→4, 6→5, 7→6. Rewritten to house grammar (`4. ~~item~~ verdict`); the lesson (a strike that swallows the digit is unstable exactly inside an otherwise-unstruck list) is recorded in the fix commit.~~ done — corrected state stands: docs/status/archived/2026-11-19_00-53_samber-do-health-rereview-status.md:69 is house grammar `4. ~~item~~ done — …`; the fix commit `156610a` touches exactly that file + ci.yml. Verified 2026-11-28

~~6. **CI-watch closed green**: TODO row removed as DONE, CHANGELOG carries the CI-activation entry (10/13 → 12/13 → 12/13 → 13/13 with each latent bug named), all gates green at every commit point (docs-gates 5/5, full-repo dprint, actionlint).~~ done — at `7433adf`: CHANGELOG.md:90 carries the CI-activation entry naming all three latent bugs and the 13/13 matrix in run 44612235355; the commit removes the watch TODO row. Verified 2026-11-28

## b) PARTIALLY DONE

~~1. **Plan execution: T1 of 13.** The push-and-watch package ran to green as a side effect of the commit directive; T2 (CHANGELOG cut), T3 (bench), T4–T7, T8–T10, T11 are untouched — deliberately, this session's mandate was plan + commit + push, not full execution.~~ **Superseded — the owner's same-day blanket order executed T2–T11 end-to-end on 2026-11-20 (docs/status/2026-11-20_11-30_execution-t2-t11-status.md).**

~~2. **The css cold-cache mechanism**: empirically fixed (warm all five zips → green) but NOT root-caused — why `go list -m` printed NOTHING (no stderr) instead of a loud fetch error on the runner is unexplained; the CI comment claims the warm step "makes fetch problems loud", which is untested for the failure mode it claims to catch.~~ done — standing fact, re-verified 2026-11-28: the warm step still enumerates all five module zips (ci.yml:220) and the comment still records the silent empty .Dir (ci.yml:214-219); no later report root-caused the silence — but the loud-fetch-problems claim is now TESTED (§f item 10)

~~3. **The flake truth on CI**: `nix flake check` shows green in 3-9s — a guarded skip, not evidence. The hub module + FOD builds remain locally-proven only until FLAKE_PAT exists (unchanged standing block).~~ done — standing fact, re-verified 2026-11-28: every hub/flake CI step is still guarded `if: env.FLAKE_PAT != ''` (ci.yml:55, 89, 253) — the flake truth on CI remains a skip; FLAKE_PAT still uncreated (digest g3 row OPEN)

~~4. **00-53's drain**: one item struck (§f4); the rest of its verifiably-done subset (§f1-3, §f5-7, §f13-20-class) is still open — T8's job.~~ done — T8 ran 2026-11-20: 00-53 is drained and archived at docs/status/archived/2026-11-19_00-53_samber-do-health-rereview-status.md, the owner-decisions digest minted, 23-40 archived with it (11-30 execution report §a7). Verified 2026-11-28

## c) NOT STARTED (plan items deliberately left for the next sessions)

~~1. T2 — CHANGELOG `[0.1.0]` cut at the 77881b3 boundary (agent-executable, no gate; simply not in this session's mandate).~~ done — 2026-11-20: `## [0.1.0] - 2026-11-19` exists in CHANGELOG.md:109 at the 77881b3 boundary (11-30 execution report §a1). Verified 2026-11-28

2. T3 — bench verify (+ re-record per the owner's g2 answer). **→ routed to TODO_LIST Low Impact row "Re-record `benchmarks/baseline.txt` (owner call g2)" — the VERIFY half is done (claim PROVEN 2026-11-20: fold @341 = 107µs vs the ≈150µs claim; 11-30 report §a2); Verified 2026-11-28: row still BLOCKED**

~~3. T4 — hub `table --sort-by` parity; T5 — `queryRows` error-path test; T6 — check-citations shellcheck + `rg` pin; T7 — embedder-docs five-checks sweep.~~ done — 2026-11-20 (T4–T7, 11-30 report §a3-§a6): hub/internal/cli/table_sort.go + README "Fleet table" section (README.md:120); error-path pins at hub/internal/cache/cache_test.go:40-57 and the wrapIterateErr DescribeTable (queries_internal_test.go:13); citation hardening at `a204282`; the five-checks table at docs/guides/integrate.md:86-90. Verified 2026-11-28

~~4. T8 — open-corpus drain + owner-decisions digest + archive 23-40/00-53; T9 — docs-gates installed-pin vs skill surface; T10 — cmdguard next-tag readiness; T11 — reconcile + full battery.~~ done — 2026-11-20 (T8–T11, 11-30 report §a7-§a10): digest minted (docs/status/owner-decisions-digest.md), 23-40 + 00-53 archived, T9 verdict recorded, cmdguard drafts + adoption sketch written, TODO_LIST reconciled, full battery green. Verified 2026-11-28

~~5. T21 — hub replace-drop (gated on the v0.1.1 go-signal).~~ done — resolved 2026-11-20, differently from the gate: the replace directive was dropped at the `v0.2.0` tag (`5010082`) instead of a v0.1.1 tag; hub/go.mod:12 requires go-compliance v0.2.0 with no replace. Verified 2026-11-28

6. Owner desk: FLAKE_PAT, M26 (g2), Forgejo visibility, go-humanize/samber-lo verdicts, upstream-TODO mandate (5 candidates), VACUUM + cache-DB, firefox. **→ routed — each row lives in a living doc (verified 2026-11-28): FLAKE_PAT → TODO_LIST Blocked "Create `FLAKE_PAT` secrets"; M26 → TODO_LIST Blocked "M26: overview endgame"; Forgejo visibility → TODO_LIST Blocked "Forgejo repo visibility"; go-humanize/samber-lo → owner-decisions-digest Linter/policy rows; upstream mandate → ROADMAP "Fleet adoption waves" owner-decisions item; VACUUM + cache DB → TODO_LIST Blocked rows; firefox → ROADMAP same item**

## d) TOTALLY FUCKED UP (all caught; the CI log is the receipts)

~~1. **I verschlimmbessert my own fix (the session's worst moment).** Run 1's css failure → I added a warm-download step listing all five modules → then "simplified" it to parent-only ("the parent pulls the whole source tree") → run 2 failed on the FIRST submodule (`datastar`), because the submodules are SEPARATE requires in hub/go.mod with their own module zips. My local "validation" proved the parent download but never modeled the cold-cache submodule resolution the runner actually needed. The first instinct was right; the cleverness cost a full red run.~~ done — historical narrative, still accurate: the shipped warm step enumerates all five zips explicitly (ci.yml:220), not parent-only; the story is also in CHANGELOG.md:90. Verified 2026-11-28

~~2. **Scoped-check discipline gap**: I ran `dprint check <touched files>` all session; CI runs it repo-wide. The 00-53 marker-strike drift passed every local check I made and was caught by CI (run 2), not by me. After run 2 I switched to full-repo checks — should have been the standing shape from the first push.~~ done — historical narrative; the corrective shape is standing: AGENTS.md:190-192 runs the battery before AND after every annotation sweep with same-minute dprint fmt. Verified 2026-11-28

~~3. **Struck a list marker** (`~~4. …~~`), violating the house grammar's own examples (`4. ~~item~~ verdict`), and dprint's repair for it would have RENUMBERED a historical list — the exact history-distortion class the annotation canon exists to prevent. CI caught it; I had already read the marker examples in the skill this session.~~ done — historical narrative; corrected state stands: the 00-53 strike is house grammar `4. ~~item~~ done — …` (archived 00-53:69) and the skill canon pins the writer shape (SKILL.md:281). Verified 2026-11-28

~~4. **dprint.json fleet drift sat in the tree unnoticed by me across TWO docs-health passes** — absolute local store paths cannot work on any runner or any other machine. One `head -3` over a sibling's dprint.json in either earlier pass would have caught it before it ever reached CI. Fixed in minutes once compared; the miss was not comparing.~~ done — fixed at `6f31054` (2026-11-19): dprint.json:4-7 now carries the fleet-canonical https plugin URLs. Verified 2026-11-28

~~5. **Three pushes to converge one green run** — each fix revealed the next latent layer (binary PATH → repo-wide format → config paths). A pre-push audit of ALL job configs against the fleet norm (siblings' dprint.json + installer targets) would very plausibly have landed green on push one.~~ done — historical narrative; the three fix commits `01762cb`/`156610a`/`6f31054` are in master's history and run 44612235355 closed the first matrix green (`gh run view`, checked 2026-11-28)

## e) WHAT WE SHOULD IMPROVE

~~1. **Local gates must mirror CI's scope**: full-repo `dprint check` + full docs-gates battery before EVERY push, never file-scoped; the scoped shape hides exactly the drift class CI then reports twenty minutes later.~~ done — standing rule, held: AGENTS.md:190-192 requires the battery before AND after every annotation sweep with same-minute dprint fmt; the 11-20 execution pass ran the FULL battery (11-30 report §d). Verified 2026-11-28

~~2. **Model the cold runner**: any CI step that touches the Go module cache needs its deps enumerated EXPLICITLY (all five zips), and local validation should run from a cold-equivalent context (`GOMODCACHE` in a temp dir) when the claim is "works on the runner".~~ done — half shipped, half unrecorded (verified 2026-11-28): the explicit enumeration is live (ci.yml:220 warms all five zips); the cold-equivalent GOMODCACHE-temp local validation ritual works (used to test §e6 today) but is recorded in no living doc

3. **Fix latent-class audits in one pass**: when a first-run gate fails on
   plumbing (PATH, config paths), audit every other job's plumbing in the SAME
   commit — three red runs each exposed one layer of the same neglected class. **→ routed to TODO_LIST Low Impact row "Record the first-execution rituals in the AGENTS.md"**
4. **Read fleet siblings BEFORE first execution**: the dprint.json norm, the
   installer target, the CWD-invocation pattern — all were one sibling-peek
   away and none were taken before push one. **→ routed to TODO_LIST Low Impact row "Record the first-execution rituals in the AGENTS.md"**

~~5. **House grammar is also formatter law**: a strike that includes the list digit is parser-hostile; annotate via the writer's canonical shape only (`N. ~~item~~ verdict`), and run the formatter BEFORE declaring an annotation pass done (the canon said this; I scoped it away).~~ done — adopted: the docs-health skill canon pins the writer shape `N. ~~item~~ verdict` and the annotate→gate→fmt→gate order (SKILL.md:281, 299). Verified 2026-11-28

~~6. **Empirical fixes need failure-mode claims that are tested**: the css warm step's comment promises "loud fetch errors" — either test that shape once (break the proxy deliberately) or soften the comment to what is proven.~~ done — tested 2026-11-28 (the deliberate shape this item asked for): cold GOMODCACHE + broken proxy (GOPROXY=http://127.0.0.1:9) → the warm `go mod download` fails LOUD — exit 1, visible `connect: connection refused`; the comment's claim holds (ci.yml:214-220)

~~7. **The `docs-gates rows`/`--pre-flight` verb mismatch** (21-17 §e5) remains open — T9 resolves it; do not re-try the ghost verb before then.~~ done — resolved by T9 (2026-11-20): the pin-vs-skill diff lives in TODO_LIST Low row "Fleet docs-gates pin bump past `e305834`" and the digest Fleet-layer row; AGENTS.md:187-197 documents the CI build vs the PATH wrapper. Verified 2026-11-28

## f) Up to 50 things we should get done next

_The plan (21-25) is the execution source; its fine tasks are referenced here,
not duplicated. New items beyond the plan are marked ★._

**Agent-executable, unblocked (plan order):**

~~1. T2: CHANGELOG `[0.1.0]` cut at 77881b3 (2.1–2.4 in the plan).~~ done — 2026-11-20: CHANGELOG.md:109 `## [0.1.0] - 2026-11-19` (11-30 execution report §a1). Verified 2026-11-28

2. T3: bench verify + re-record per g2 answer (3.1–3.3).~~~~ **→ routed to TODO_LIST Low Impact row "Re-record `benchmarks/baseline.txt` (owner call g2)"; VERIFY half done 2026-11-20 — claim PROVEN (fold @341 = 107µs vs ≈150µs; 11-30 report §a2). Verified 2026-11-28: row still BLOCKED**

~~3. T4: `--sort-by` parity (4.1–4.5).~~ done — 2026-11-20: hub/internal/cli/table_sort.go + README "Fleet table" section (README.md:120-129). Verified 2026-11-28

~~4. T5: `queryRows` error-path specs (5.1–5.3).~~ done — 2026-11-20: error-path pins hub/internal/cache/cache_test.go:40-57 (ErrIO + Transient family) + the wrapIterateErr DescribeTable (queries_internal_test.go:13). Verified 2026-11-28

~~5. T6: check-citations shellcheck + `rg` pin (6.1–6.5).~~ done — citation-gate hardening, `a204282` (2026-11-20): `extract_cites` prefers ripgrep and falls back to `grep -oE`, shellcheck clean (CHANGELOG [0.1.0])

~~6. T7: embedder-docs five-checks sweep (7.1–7.3).~~ done — 2026-11-20: the five-checks table with criticality at docs/guides/integrate.md:86-90. Verified 2026-11-28

~~7. T8: corpus drain — strike the done-subset of 23-09/23-40/00-53, owner digest, archive two (8.1–8.5).~~ done — 2026-11-20: digest minted at docs/status/owner-decisions-digest.md; 23-40 + 00-53 archived under docs/status/archived/ (both present 2026-11-28); 68 rows annotated (11-30 report §a7)

~~8. T9: docs-gates pin-vs-skill verdict (9.1–9.3).~~ done — 2026-11-20: T9 verdict recorded in AGENTS.md (fleet pin e305834 lacks ONLY the rows verb; its freshness surface is complete; writer aliases cover the row gates) + TODO_LIST carries the BLOCKED bump row

~~9. T10: cmdguard next-tag readiness + upstream drafts (10.1–10.3).~~ done — 2026-11-20: docs/drafts/cmdguard-nd-duration-codec.md + cmdguard-lint-published-replaces.md + the adoption sketch docs/notes/cmdguard-describe-adoption-sketch.md (all present 2026-11-28); filing gated on the OG.5 mandate (TODO_LIST Low cmdguard row)

~~10. **→ 21-25 plan T11** — TODO/ROADMAP reconcile + FULL battery (11.1–11.4), executed with the wave that precedes it.~~ done — executed 2026-11-20 with the T2–T11 wave: reconcile + full battery green (11-30 execution report §a10 and §d). Verified 2026-11-28

~~11. ★ Re-run the FULL CI matrix once more after T2–T11 land (the run history now exists — keep it green, not "green last week").~~ done — re-ran continuously since — NOT all green today: gh run list shows failure runs on 2026-11-27/28 (e.g. 45974600959, 46105857772) — 13/14 jobs pass, docs-gates red on audit-all over 30 open items in ten reports 2026-11-18..25, this report among them (TODO_LIST High Impact row tracks the fix). Verified 2026-11-28

~~12. ★ Soften or test the css warm-step comment's "loud errors" claim (§e6).~~ done — tested 2026-11-28: cold GOMODCACHE + deliberately broken proxy (GOPROXY=http://127.0.0.1:9) → the warm `go mod download` fails LOUD (exit 1, `connect: connection refused`); the comment's claim holds, comment unchanged (ci.yml:214-220)

~~13. ★ Confirm the FLAKE_PAT-guarded jobs' skip shape is still exactly "skip-not-fail" on a real run (observed green-skip this session; pin the observation in README if it differs).~~ done — skip shape unchanged and pinned: every hub/flake step guarded `if: env.FLAKE_PAT != ''` (ci.yml:55-96, 253) and README.md:180-183 states "skip (not fail)". Verified 2026-11-28

**Owner-gated (rows exist):**

14. FLAKE_PAT secrets — ALSO the only way CI ever proves the flake truth (§b3); until then CI's flake green is a skip.~~~~ **→ routed to TODO_LIST Blocked (owner decisions) row "Create `FLAKE_PAT` secrets (GitHub + Forgejo) for CI hub/flake jobs" (digest §00-04 §g row "g3: FLAKE_PAT creation"). Verified 2026-11-28: still OPEN, ci.yml still guards**

~~15. v0.1.1 go-signal → T21 replace-drop executes (~34m).~~ done — the replace directive was dropped at the `v0.2.0` tag (`5010082`, 2026-11-20) instead of a v0.1.1 tag; `hub/go.mod` requires `go-compliance v0.2.0` with no replace

16. go-humanize + samber/lo verdicts (the findings-gate blocker).~~~~ **→ routed to docs/status/owner-decisions-digest.md Linter/policy rows "samber/lo adopt-or-suppress" and "go-humanize adopt/accept-with-reason" (ROADMAP owner queue). Verified 2026-11-28: both rows open**

17. M26 overview endgame (g2).~~~~ **→ routed to TODO_LIST Blocked (owner decisions) row "M26: overview endgame (replace vs coexist) + migration note" (ROADMAP section "Endgame question (owner, g2 — BLOCKED)"). Verified 2026-11-28: open**

18. Forgejo visibility click.~~~~ **→ routed to TODO_LIST Blocked (owner decisions) row "Forgejo repo visibility (lars/go-compliance → public?)". Verified 2026-11-28: open**

19. Upstream-TODO filing mandate (5 queued candidates; T10 pre-writes drafts).~~~~ **→ routed to ROADMAP "Fleet adoption waves" owner-decisions item "mandate for consumer sessions filing upstream TODOs in library repos" (the 5 candidates listed there); the drafts are READY (TODO_LIST Low cmdguard row). Verified 2026-11-28: mandate still pending**

20. VACUUM buildflow DB + delete the 0.51 GB cache DB.~~~~ **→ routed to TODO_LIST Blocked (owner decisions) rows "`sqlite3 ~/.local/state/buildflow/buildflow.db VACUUM`" and "Delete the expendable 0.51 GB cache DB". Verified 2026-11-28: both open**

21. Firefox permission → index.html screenshot row closes.~~~~ **→ routed to ROADMAP "Fleet adoption waves" owner-decisions item "firefox permission for agent sessions (index.html screenshot)". Verified 2026-11-28: the ask still stands there; scripts/screenshot.sh meanwhile also serves screenshots headless**

**Watch / standing:**

22. ~~Every future push: full-repo dprint + docs-gates battery pre-push (§e1).~~ **→ routed to TODO_LIST Low Impact row "Record the first-execution rituals in the AGENTS.md"**

~~23. The CI run history now anchors "the gates have RUN" — FEATURES' zero-runs note and the README CI section should be softened to "first runs green 2026-11-19" (one-line doc fix, fold into T11).~~ done — in the T11-era doc pass: FEATURES.md:60 CI row now reads "all with real green runner executions" and README.md:167-175 lists the full matrix without a zero-runs caveat. Verified 2026-11-28

~~24. Keep the annotate→gate→fmt→gate order; add "full-repo fmt scope" to the canon (§e1) — crush-config lesson candidate (owner commit).~~ done — adopted: the docs-health skill's formatter-order canon is SKILL.md:299 (annotate → check-rows → format → check-rows again) plus the 2026-11-20 keyed-spec lesson at SKILL.md:303; repo-side rule at AGENTS.md:190-192. Verified 2026-11-28

~~25. 16-12 §f16's watcher-spec CI flake watch: the watcher spec has now run on a real runner twice (green) — the row's evidence base improved; revisit the TODO row wording at T11.~~ done — resolved at T11 (2026-11-20): the reconcile rebuilt TODO_LIST and the watcher watch row is gone (today's TODO_LIST carries no watcher row; 11-30 report §a10); the hub test job that runs the watcher spec is green on the latest CI runs — only docs-gates is red (TODO_LIST High row; gh run list 2026-11-28)

(25 items; slots 26-50 are the existing TODO_LIST rows — 14 open — and the
plan's fine tables, referenced not duplicated.)

## g) Questions I cannot figure out myself

~~1. **T2 timing**: cut the CHANGELOG `[0.1.0]` section NOW (splits [Unreleased] at the 77881b3 boundary), or hold until the v0.1.1 tag so one cut covers both releases? The plan schedules T2 now; your call changes the split.~~ done — resolved — the digest's "Still-gated from the CI-activation report (00-04 §g)" row 1 is struck done: "resolved by the owner's 2026-11-20 blanket execution order; the `[0.1.0]` cut landed same-day at the 77881b3 boundary" (owner-decisions-digest.md:65-66; CHANGELOG.md:109). Verified 2026-11-28

2. **Bench baseline policy (carried)**: when T3 runs — re-record `benchmarks/baseline.txt` on the current (untagged) tree, or verify-only and keep the M27-era baseline until v0.1.1? **→ routed to docs/status/owner-decisions-digest.md section "Still-gated from the CI-activation report (00-04 §g)", row "g2: bench baseline re-record policy" (VERIFY done; re-record is the owner's optional call). Verified 2026-11-28: row open**
3. **FLAKE_PAT now?** CI's flake green is currently a guarded skip (§b3) — the hub module and FOD builds have still never executed on a real runner. Creating the secret un-skips them; the first real run may surface hub-side fetch/config surprises exactly like the dprint/css jobs did. Create it now, or keep CI partially-skipped until you want that surface probed? **→ routed to docs/status/owner-decisions-digest.md section "Still-gated from the CI-activation report (00-04 §g)", row "g3: FLAKE_PAT creation (GitHub + Forgejo, read-only Contents on go-daemon)". Verified 2026-11-28: row OPEN, ci.yml still guards the flake job**

---

_Gates at report time: CI 13/13 green (run 44612235355); docs-gates 5/5 PASS;
full-repo dprint clean; both remotes at `7433adf`; TODO_LIST 14 open rows
(11 Low + 3 Blocked-class owner gates consolidated in ROADMAP). Daemon owns
most commits; the five session commits are explicitly owner-ordered. WAITING
for instructions._

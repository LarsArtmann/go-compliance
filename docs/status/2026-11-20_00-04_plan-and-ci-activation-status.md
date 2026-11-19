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

1. **Plan written per the owner's format** — `docs/planning/2026-11-19_21-25_corpus-drain-ci-activation-pareto-plan.md`:
   1%/4%/20%/other-20% breakdown, 13 medium packages (≈7h agent work), 50 fine
   tasks (≤12m), mermaid execution graph, errata + anti-verschlimmbessern rules,
   ALL open TODOs included (the three §c owner items and demand-gated ROADMAP
   ideas are listed, not scheduled).
2. **Step-0 errata with evidence** — 3 stale TODO rows removed (importance =
   decided-not-blocked; all-systems = done 2026-11-19; push = ordered by this
   directive), 4 rows minted (CI watch, corpus drain, docs-gates verify,
   cmdguard readiness), 1 CHANGELOG all-systems gap covered by annotating
   00-53 §f4 done at the tree.
3. **The push executed plan task T1 LIVE** — both remotes at `7433adf`; the push
   triggered the repo's FIRST real CI execution ever (every previous "CI green"
   claim had zero runner evidence — the four newest gates existed only as YAML).
4. **Three latent first-run bugs found and fixed across three fix commits:**
   - dprint job: `GITHUB_PATH` pointed at `$PWD/bin`; the dprint.dev installer
     targets `~/.dprint/bin` → exit 127. Fixed; actionlint's CWD-invocation
     hardened in the same step.
   - css-freshness: `gen-library-classes.sh` resolves templ-components via
     `go list -m -f '{{.Dir}}'`; the cold runner returned silent-empty. Fixed by
     warming ALL FIVE module zips (parent-only warm-up was my intermediate
     mistake, see §d1).
   - dprint.json: plugin array pointed at absolute local BuildFlow-store paths
     (`/home/lars/.local/state/buildflow/...`) — fleet drift vs all four sibling
     repos, which use the fleet-canonical https URLs (same versions, so
     formatting is byte-identical; proven by full-repo `dprint check` green).
5. **A formatter-unstable strike caught before it could distort history**: my
   00-53 §f4 annotation struck the LIST MARKER (`~~4. …~~`); dprint's markdown
   parser then wanted to indent the line as a paragraph and RENUMBER items
   5→4, 6→5, 7→6. Rewritten to house grammar (`4. ~~item~~ verdict`); the
   lesson (a strike that swallows the digit is unstable exactly inside an
   otherwise-unstruck list) is recorded in the fix commit.
6. **CI-watch closed green**: TODO row removed as DONE, CHANGELOG carries the
   CI-activation entry (10/13 → 12/13 → 12/13 → 13/13 with each latent bug
   named), all gates green at every commit point (docs-gates 5/5, full-repo
   dprint, actionlint).

## b) PARTIALLY DONE

1. **Plan execution: T1 of 13.** The push-and-watch package ran to green as a
   side effect of the commit directive; T2 (CHANGELOG cut), T3 (bench), T4–T7,
   T8–T10, T11 are untouched — deliberately, this session's mandate was plan +
   commit + push, not full execution.
2. **The css cold-cache mechanism**: empirically fixed (warm all five zips →
   green) but NOT root-caused — why `go list -m` printed NOTHING (no stderr)
   instead of a loud fetch error on the runner is unexplained; the CI comment
   claims the warm step "makes fetch problems loud", which is untested for the
   failure mode it claims to catch.
3. **The flake truth on CI**: `nix flake check` shows green in 3-9s — a guarded
   skip, not evidence. The hub module + FOD builds remain locally-proven only
   until FLAKE_PAT exists (unchanged standing block).
4. **00-53's drain**: one item struck (§f4); the rest of its verifiably-done
   subset (§f1-3, §f5-7, §f13-20-class) is still open — T8's job.

## c) NOT STARTED (plan items deliberately left for the next sessions)

1. T2 — CHANGELOG `[0.1.0]` cut at the 77881b3 boundary (agent-executable, no
   gate; simply not in this session's mandate).
2. T3 — bench verify (+ re-record per the owner's g2 answer).
3. T4 — hub `table --sort-by` parity; T5 — `queryRows` error-path test; T6 —
   check-citations shellcheck + `rg` pin; T7 — embedder-docs five-checks sweep.
4. T8 — open-corpus drain + owner-decisions digest + archive 23-40/00-53;
   T9 — docs-gates installed-pin vs skill surface; T10 — cmdguard next-tag
   readiness; T11 — reconcile + full battery.
5. T21 — hub replace-drop (gated on the v0.1.1 go-signal).
6. Owner desk: FLAKE_PAT, M26 (g2), Forgejo visibility, go-humanize/samber-lo
   verdicts, upstream-TODO mandate (5 candidates), VACUUM + cache-DB, firefox.

## d) TOTALLY FUCKED UP (all caught; the CI log is the receipts)

1. **I verschlimmbessert my own fix (the session's worst moment).** Run 1's css
   failure → I added a warm-download step listing all five modules → then
   "simplified" it to parent-only ("the parent pulls the whole source tree")
   → run 2 failed on the FIRST submodule (`datastar`), because the submodules
   are SEPARATE requires in hub/go.mod with their own module zips. My local
   "validation" proved the parent download but never modeled the cold-cache
   submodule resolution the runner actually needed. The first instinct was
   right; the cleverness cost a full red run.
2. **Scoped-check discipline gap**: I ran `dprint check <touched files>` all
   session; CI runs it repo-wide. The 00-53 marker-strike drift passed every
   local check I made and was caught by CI (run 2), not by me. After run 2 I
   switched to full-repo checks — should have been the standing shape from
   the first push.
3. **Struck a list marker** (`~~4. …~~`), violating the house grammar's own
   examples (`4. ~~item~~ verdict`), and dprint's repair for it would have
   RENUMBERED a historical list — the exact history-distortion class the
   annotation canon exists to prevent. CI caught it; I had already read the
   marker examples in the skill this session.
4. **dprint.json fleet drift sat in the tree unnoticed by me across TWO
   docs-health passes** — absolute local store paths cannot work on any runner
   or any other machine. One `head -3` over a sibling's dprint.json in either
   earlier pass would have caught it before it ever reached CI. Fixed in
   minutes once compared; the miss was not comparing.
5. **Three pushes to converge one green run** — each fix revealed the next
   latent layer (binary PATH → repo-wide format → config paths). A pre-push
   audit of ALL job configs against the fleet norm (siblings' dprint.json +
   installer targets) would very plausibly have landed green on push one.

## e) WHAT WE SHOULD IMPROVE

1. **Local gates must mirror CI's scope**: full-repo `dprint check` + full
   docs-gates battery before EVERY push, never file-scoped; the scoped shape
   hides exactly the drift class CI then reports twenty minutes later.
2. **Model the cold runner**: any CI step that touches the Go module cache
   needs its deps enumerated EXPLICITLY (all five zips), and local validation
   should run from a cold-equivalent context (`GOMODCACHE` in a temp dir) when
   the claim is "works on the runner".
3. **Fix latent-class audits in one pass**: when a first-run gate fails on
   plumbing (PATH, config paths), audit every other job's plumbing in the SAME
   commit — three red runs each exposed one layer of the same neglected class.
4. **Read fleet siblings BEFORE first execution**: the dprint.json norm, the
   installer target, the CWD-invocation pattern — all were one sibling-peek
   away and none were taken before push one.
5. **House grammar is also formatter law**: a strike that includes the list
   digit is parser-hostile; annotate via the writer's canonical shape only
   (`N. ~~item~~ verdict`), and run the formatter BEFORE declaring an
   annotation pass done (the canon said this; I scoped it away).
6. **Empirical fixes need failure-mode claims that are tested**: the css warm
   step's comment promises "loud fetch errors" — either test that shape once
   (break the proxy deliberately) or soften the comment to what is proven.
7. **The `docs-gates rows`/`--pre-flight` verb mismatch** (21-17 §e5) remains
   open — T9 resolves it; do not re-try the ghost verb before then.

## f) Up to 50 things we should get done next

_The plan (21-25) is the execution source; its fine tasks are referenced here,
not duplicated. New items beyond the plan are marked ★._

**Agent-executable, unblocked (plan order):**

1. T2: CHANGELOG `[0.1.0]` cut at 77881b3 (2.1–2.4 in the plan).
2. T3: bench verify + re-record per g2 answer (3.1–3.3).
3. T4: `--sort-by` parity (4.1–4.5).
4. T5: `queryRows` error-path specs (5.1–5.3).
5. T6: check-citations shellcheck + `rg` pin (6.1–6.5).
6. T7: embedder-docs five-checks sweep (7.1–7.3).
7. T8: corpus drain — strike the done-subset of 23-09/23-40/00-53, owner
   digest, archive two (8.1–8.5).
8. T9: docs-gates pin-vs-skill verdict (9.1–9.3).
9. T10: cmdguard next-tag readiness + upstream drafts (10.1–10.3).
10. **→ 21-25 plan T11** — TODO/ROADMAP reconcile + FULL battery (11.1–11.4), executed with the wave that precedes it.
11. ★ Re-run the FULL CI matrix once more after T2–T11 land (the run history
    now exists — keep it green, not "green last week").
12. ★ Soften or test the css warm-step comment's "loud errors" claim (§e6).
13. ★ Confirm the FLAKE_PAT-guarded jobs' skip shape is still exactly
    "skip-not-fail" on a real run (observed green-skip this session; pin the
    observation in README if it differs).

**Owner-gated (rows exist):**

14. FLAKE_PAT secrets — ALSO the only way CI ever proves the flake truth
    (§b3); until then CI's flake green is a skip.
15. v0.1.1 go-signal → T21 replace-drop executes (~34m).
16. go-humanize + samber/lo verdicts (the findings-gate blocker).
17. M26 overview endgame (g2).
18. Forgejo visibility click.
19. Upstream-TODO filing mandate (5 queued candidates; T10 pre-writes drafts).
20. VACUUM buildflow DB + delete the 0.51 GB cache DB.
21. Firefox permission → index.html screenshot row closes.

**Watch / standing:**

22. Every future push: full-repo dprint + docs-gates battery pre-push (§e1).
23. The CI run history now anchors "the gates have RUN" — FEATURES' zero-runs
    note and the README CI section should be softened to "first runs green
    2026-11-19" (one-line doc fix, fold into T11).
24. Keep the annotate→gate→fmt→gate order; add "full-repo fmt scope" to the
    canon (§e1) — crush-config lesson candidate (owner commit).
25. 16-12 §f16's watcher-spec CI flake watch: the watcher spec has now run on
    a real runner twice (green) — the row's evidence base improved; revisit
    the TODO row wording at T11.

(25 items; slots 26-50 are the existing TODO_LIST rows — 14 open — and the
plan's fine tables, referenced not duplicated.)

## g) Questions I cannot figure out myself

1. **T2 timing**: cut the CHANGELOG `[0.1.0]` section NOW (splits [Unreleased]
   at the 77881b3 boundary), or hold until the v0.1.1 tag so one cut covers
   both releases? The plan schedules T2 now; your call changes the split.
2. **Bench baseline policy (carried)**: when T3 runs — re-record
   `benchmarks/baseline.txt` on the current (untagged) tree, or verify-only
   and keep the M27-era baseline until v0.1.1?
3. **FLAKE_PAT now?** CI's flake green is currently a guarded skip (§b3) — the
   hub module and FOD builds have still never executed on a real runner.
   Creating the secret un-skips them; the first real run may surface hub-side
   fetch/config surprises exactly like the dprint/css jobs did. Create it now,
   or keep CI partially-skipped until you want that surface probed?

---

_Gates at report time: CI 13/13 green (run 44612235355); docs-gates 5/5 PASS;
full-repo dprint clean; both remotes at `7433adf`; TODO_LIST 14 open rows
(11 Low + 3 Blocked-class owner gates consolidated in ROADMAP). Daemon owns
most commits; the five session commits are explicitly owner-ordered. WAITING
for instructions._

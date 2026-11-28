# Status Report: v0.2.0 Release Run — 2026-11-20 12:06 CET

**Session scope:** "Publish a new version!" → go-release skill run on go-compliance (multi-module: core root + hub/).

**VERIFIED this session (with command/artifact):** core+hub `GOWORK=off go build/vet/test -race -count=1` all green; golangci 0+0 both modules; `templ fmt -fail` green; art-dupl "no new clones"; docs-gates 5/5; dprint was clean pre-subagent and is NOW **2 files unformatted**; `nix build .#compliance-hub` converged on new vendorHash `sha256-t1Kx9BVBgHbwRf4nO9Bqwen8eiNm6iAXCEql9oeZHbE=`; `nix flake check` all passed; BD003 brandid-lint exit 0; proxy `@v/list` = `[v0.1.0]`; core exported API v0.1.0→HEAD = 15→15 symbols (additive, via grep over `^(func|type) [A-Z]`).

**NOT verified:** no tag pushed, no proxy v0.2.0 (release NOT published yet); citation gate still RED (stale 30→15, subagent mid-work); CI red on master (latest runs failed on citation gate; a new run is queued on the daemon's partial citation commit); I did NOT run the canonical plain `buildflow` full gate (AGENTS.md calls it canonical; I ran per-tool drill-downs only); stack-drift-check NOT run locally (CI-only in my battery); exit-code probe (75 contract) not re-run.

---

## Session narrative (what happened)

1. Loaded go-release skill + multi-module reference. Counted go.mod files → multi-module shape; checked tag history: **only `v0.1.0` exists, core-only** (hub deliberately untagged; binary via nix).
2. Assessment: 62 commits since v0.1.0; real Go delta = 37 files +1871/−688; hub gained features (filewatcher adoption, `--format`, `--sort-by`, 404/dark-mode/compiled-CSS, sysexits, nosniff). Version call: **v0.2.0** (MINOR; core API verified additive — reporter.go's −208 lines was a pure `FileStore`→filestore.go split, 15 exported symbols before and after).
3. **Discovered CI red on master** (citation gate + css-freshness, broken ~1h before the session by the earlier T2–T11 execution's daemon commits). A red tag commit stays red forever → repair became release-blocking step 1.
4. Fixed css-freshness myself: ran `scripts/regen-ui.sh` (same tailwindcss_4 as CI), regenerated `app.min.css` + `library-custom.css`, daemon committed (199d349).
5. Dispatched `roberto-citation-adjudicator` (write subagent) for the 30 drifted citations across 12 lib reports — root-caused to T9/T11 file splits (reporter.go, cache.go→queries.go, events.go→handler.go, etc.).
6. Found the recorded TODO_LIST row: **"Drop hub `replace` directive — execute with the NEXT core tag"**. This reshaped the plan: bump hub require → tag → THEN drop replace as post-tag commit (avoids the chicken-and-egg of tidy-before-tag).
7. Cut CHANGELOG `[0.2.0] - 2026-11-20` from `[Unreleased]` (empty placeholders left).
8. `go mod edit` hub require v0.1.0 → v0.2.0; tidy both modules zero-diff (go.sum untouched — replace posture confirmed).
9. vendorHash bootstrap ritual: fake hash → `got: sha256-t1Kx9BVB…` → pasted → rebuild converged → `nix flake check` all passed. Core hash (L202) stable as predicted.
10. Full local battery green (see scope line).
11. **Parallel-session collisions observed:** 80afd3a (daemon) mixed MY flake.nix hash with ANOTHER session's `ci.yml` (new brandid-lint job) + `.buildflow.yml` (BD003 opt-in). I pre-verified BD003 locally (exit 0) instead of reverting (unexpected ≠ wrong).
12. Drafted GitHub release notes at `/tmp/release-notes.md`. Fetched proxy baseline (`v0.1.0` only).

## Brutal self-review

**What did I forget?**

- The release-notes draft says "13-job matrix" — the parallel session made it **14** (brandid-lint). The CHANGELOG 0.2.0 section was cut BEFORE their commits, so it misses the brandid gate + the css-regen fix + the citation re-adjudication. A release cut on a live daemon-committed tip is stale the moment another session is active.
- Did not give the citation subagent a **dprint/formatting constraint** — its edits currently leave 2 files unformatted, which will keep the dprint CI job red even after citations pass.
- Did not check for an ACTIVE parallel session before planning (the untracked brutal-self-review status doc in the env snapshot was the tell; the 11:51 `.buildflow.yml` was the confirmation).
- Ran the canonical-gate drill-downs but never plain `buildflow` (the AGENTS.md canonical gate); the 2026-11-18 incident proved manual-green sweeps miss findings.
- `gh run list` came third in my assessment order; CI-red should have been checked first — it invalidates everything else.

**Did I lie?** No. Every green claim above has a command behind it from this session. The one soft claim — "core hash stable as predicted" — was confirmed by `nix flake check` passing, not assumed.

**Ghost systems?** None created. Existing one noticed: `examples/reporter` demos the reporter SDK and is the only consumer-shaped doc for it — fine, wired via docs/guides/integrate.md.

**Split brains found (this session):**

1. hub/go.mod `require v0.2.0` + `replace => ../` + flake `postPatchExtra` sed repointing that replace at `_core_only` — THREE places encoding "how core resolves". The post-tag replace-drop collapses all three (already planned).
2. TODO_LIST says "v0.1.1-class tag", I shipped v0.2.0 — number drift inside a recorded row (class ≠ pin, but the row text is now wrong).
3. CHANGELOG/release-notes "13 jobs" vs ci.yml's 14 jobs.
4. Release notes live in /tmp (ephemeral) while the in-repo CHANGELOG is canonical — after `gh release create` the /tmp copy should die.

**Stupid things we do anyway:** CI has NO tag-triggered workflow — a tag can freeze a tree whose CI state nobody will ever see red-or-green on the tag ref itself; the flake-check CI job silently no-ops without FLAKE_PAT, so vendorHash has ZERO CI coverage (only my local ritual protects it); release prep can never form a clean dedicated commit under the continuous daemon (69bc125 got lucky with 3/3 my files, 80afd3a mixed theirs).

**Fucked up (d) honestly:** Nothing destroyed, no re-tags, no proxy poisoning, no data loss. The genuinely bad: **the tree was NOT release-ready when the user asked** — CI had been red for ~1.5h from earlier sessions, and ~1h into the "publish" run we are still gated on citation adjudication. The release is late because it inherited debt, not because this session broke anything — but it's late.

---

## a) FULLY DONE

| #  | Item                                                                                          | Evidence                                                           |
| -- | --------------------------------------------------------------------------------------------- | ------------------------------------------------------------------ |
| 1  | CSS-freshness gate repaired (regenerated `app.min.css` + `library-custom.css`)                | committed 199d349; regen rerun byte-stable                         |
| 2  | Version determination: **v0.2.0 MINOR**                                                       | 62-commit delta reviewed; core API 15→15 exported symbols verified |
| 3  | CHANGELOG `[0.2.0] - 2026-11-20` cut, `[Unreleased]` reset with placeholders                  | committed 69bc125                                                  |
| 4  | hub require → `go-compliance v0.2.0` via `go mod edit`                                        | committed 69bc125; tidy zero-diff                                  |
| 5  | Hub vendorHash bootstrapped (both build passes + convergence)                                 | `nix build` green; hash `t1Kx9BVB…` in flake.nix, committed        |
| 6  | `nix flake check` ALL PASSED (hub package + core-build + templ-gen + devShell)                | command output this session                                        |
| 7  | Full test battery: core+hub build/vet/test `-race -count=1`                                   | all ok, sequential                                                 |
| 8  | Lint: golangci 0 issues both modules; templ fmt -fail; art-dupl no new clones; docs-gates 5/5 | all run this session                                               |
| 9  | BD003 brandid gate pre-verified for the parallel session's new CI job                         | `go run …brandid-lint@linter/v0.5.1 -rules BD003 .` exit 0         |
| 10 | GitHub release notes drafted                                                                  | `/tmp/release-notes.md`                                            |
| 11 | Citation drift root-caused and fix dispatched                                                 | subagent `roberto-citation-adjudicator`, stale 30→15 and falling   |
| 12 | Proxy baseline established                                                                    | `@v/list` = v0.1.0 only                                            |

## b) PARTIALLY DONE

| # | Item                                   | State                                                                                  |
| - | -------------------------------------- | -------------------------------------------------------------------------------------- |
| 1 | **The release itself**                 | All prep committed (69bc125); NOT tagged, NOT pushed, NOT published                    |
| 2 | Citation gate repair                   | 30→15 stale; subagent still running; dprint-clean pass still needed after              |
| 3 | CI green-on-tag-commit precondition    | Red master; queued run on partial citation commit will fail; need push-after-fix cycle |
| 4 | Release-notes accuracy                 | Drafted but contains the stale "13 jobs" claim; CHANGELOG missing 3 in-window items    |
| 5 | hub replace-drop (TODO_LIST row)       | Planned + sequenced (post-tag), NOT executed                                           |
| 6 | Post-tag TODO_LIST/ROADMAP row updates | Planned, not started                                                                   |

## c) NOT STARTED

| # | Item                                                                                                           |
| - | -------------------------------------------------------------------------------------------------------------- |
| 1 | Push master (github + forgejo) → CI green → annotated tag `v0.2.0` → push tag both remotes                     |
| 2 | Post-push verification: proxy `.info`, pkg.go.dev fetch, clean-dir `go get`                                    |
| 3 | `gh release create v0.2.0` (notes + `--prerelease` decision pending)                                           |
| 4 | Retro-create v0.1.0 GitHub Release (v0.1.0 shipped tag-only, zero notes on GitHub)                             |
| 5 | Consumer propagation: md-go-validator + BuildFlow bound core at v0.1.0 → bump to v0.2.0 (go-ecosystem-upgrade) |
| 6 | Canonical plain `buildflow` full run (AGENTS.md canonical gate — skipped this session)                         |
| 7 | AGENTS.md memory updates: require-bump-vendorHash lesson; release-on-live-repo protocol                        |
| 8 | docs-health HARVEST of this report's §f into TODO_LIST (blocked: "wait for instructions")                      |

## d) TOTALLY FUCKED UP

Nothing destroyed or corrupted. The honest entries:

1. **Release latency:** ~1h into "publish", still untagged — inherited red CI + live parallel sessions, but the run is objectively late.
2. **The frozen-tag blind spot:** CI never runs on tag refs and flake-check no-ops without FLAKE_PAT — the ecosystem's two hardest invariants (tag CI green, vendorHash correct) have NO mechanical protection. v0.2.0 relies entirely on my local ritual.
3. **Release-notes/CHANGELOG cut on a moving tip** — already stale by one feature (brandid gate).

## e) WHAT WE SHOULD IMPROVE

1. **Pre-flight collision check** for every task on a daemon-committed repo: `git status`, last-3 commits w/ files, recent mtimes, active-session tells BEFORE the first edit. The env snapshot's untracked status doc was the warning; I read it and moved on.
2. **CI-first release ordering:** `gh run list` before anything else; Phase 4.4 exists precisely for this.
3. **Subagent briefs need explicit constraints:** time budget, dprint/format gate, "rerun gate X to exit 0" acceptance. The citation brief had acceptance but not formatting.
4. **Cut release content LAST:** notes/CHANGELOG after all gates and all sessions quiesce — not in parallel with repair work.
5. **Tag-triggered CI** (see §f) so the #1 rule (tag green) is mechanical, not ritual.
6. **Canonical gate discipline:** run plain `buildflow` once before tagging; drill-downs are not a substitute (2026-11-18: manual sweep missed 79 findings).
7. **Questions first, work second** — §g below contains a ship/window decision I guessed at (tag tip as-is, absorbing the parallel session's unreviewed-by-me CI change) when I could have asked.

## f) Up to 50 things to get done next (ordered: in-flight → release path → follow-ups → hardening)

**In-flight (finish before anything else):**

1. Citation subagent completes → verify stale=0, exit 0
2. `dprint fmt` pass over the subagent's edited reports (dprint gate currently 2 files unformatted)
3. Re-run citation gate + dprint + docs-gates together post-fix
4. CHANGELOG 0.2.0: add the 3 in-window items (brandid-lint gate, css-regen repair, citation re-adjudication)
5. Fix release-notes "13 jobs" → 14, and re-verify every claim against the final CHANGELOG
6. Confirm the parallel session is quiescent (no new daemon commits for ~10 min) before tagging
7. Push master to github; watch CI to green (14 jobs); push forgejo for parity
8. Create annotated tag `v0.2.0` at the green commit; push tag to both remotes
9. Post-push: proxy `.info` shows v0.2.0 (wait 2–10 min propagation)
~~10. pkg.go.dev fetch trigger + render check~~ done — pkg.go.dev renders the module at v0.2.0, published 2026-11-20 (README, examples, and godoc verified by fetch on 2026-11-28)
11. Clean-dir `go get github.com/larsartmann/go-compliance@v0.2.0` (GONOSUMDB='*' then without)
12. `gh release create v0.2.0` with notes — decide `--prerelease` (§g)
13. Retro-create v0.1.0 GitHub Release so the releases page isn't a hole

**Post-tag (the recorded row):**
14. Drop `replace github.com/larsartmann/go-compliance => ../` from hub/go.mod
15. `go mod tidy` hub (now resolves core v0.2.0 from proxy; go.sum gains checksums)
16. vendorHash ritual #2 (fake → got → paste → converge) at the hub site
17. Simplify flake `postPatchExtra` `_core_only` story (replace-drop makes the sed dead)
18. Full battery re-run after replace-drop (tests, lint, `nix flake check`)
19. Update the flake comment about "replace => ../ resolves core in-tree" (now lies)
~~20. STACK.md decision-log entry for the replace-drop~~ done — recorded 2026-11-28 in the STACK.md decision log ("The hub consumes the released core, not a replace"), together with the stale wiring note
21. AGENTS.md Module Rules rewrite: go-daemon/_local_deps + replace wiring text goes stale
22. TODO_LIST: tick the replace-drop row (also fix its "v0.1.1-class" text → done-as-v0.2.0)
23. ROADMAP: update the "core v0.1.1 tag" owner-decision line to shipped-v0.2.0
24. go.work: re-verify `use .` + `use ./hub` still keeps local dev live post-replace (it should)
25. Exit-code contract probe: `compliance-hub report --db /nonexistent/x.db; echo $?` → 75
26. Built-binary smoke: `nix build .#compliance-hub && ./result/bin/compliance-hub --help`

**Consumer + docs propagation:**
27. md-go-validator bump to core v0.2.0 (go-ecosystem-upgrade)
28. BuildFlow core binding bump v0.1.0 → v0.2.0
29. README install/`go get` snippets → v0.2.0
30. docs/guides/integrate.md version references
31. FEATURES.md CI-gate row → 14 jobs
32. AGENTS.md: record require-bump-vendorHash lesson (modules.txt version string flips hub hash; core hash stable)
33. AGENTS.md: record release-on-live-repo protocol (collision check, cut-last, CI-first)
34. docs-health HARVEST of this report's §f into TODO_LIST/ROADMAP
35. Verify release-notes claims == CHANGELOG exactly (no invented claims shipped)

**Hardening (release machinery):**
36. Add tag-triggered CI workflow (`on: push: tags: v*`) running the battery on the frozen tag
37. Investigate FLAKE_PAT: deliberate absence or missing secret? If added, hub vendorHash gains CI coverage
38. Make the stack-drift-check + exit-code probe part of my local battery script (CI-only today)
39. Run plain `buildflow` canonical gate once pre-tag (this session's skipped step)
40. Release runbook doc: distill this session into docs/notes/release-runbook.md (collision check → CI first → battery → tag → post-push → replace-drop)
41. Consider a `scripts/pre-release-check.sh` in-repo (the go-release skill assumes one; this repo has none)
42. Baseline-vs-ratchet hygiene: after replace-drop moves cited code again, rerun check-citations (accept-directives are line-proximity fragile)
43. Tag-watch: check go-daemon/go-filewatcher for new tags → executed-addenda to lib reports (AGENTS.md hook)
44. cmdguard DescribeContract sketch still gates on unpublished cmdguard tag (upstream watch)
45. Benchmark re-record decision (107µs vs 127µs baseline; owner g2 call, evidence already in CHANGELOG)
46. Decide Forgejo release mirroring (releases exist on GitHub only today; §g)
47. Sum-check `.buildflow.yml` under dprint (yml formatting; landed mid-session from parallel work)
48. Consider pinning CI to `--all-systems` flake check or an aarch64 runner eventually (current check omits both darwin/arm64)
49. Retire /tmp/release-notes.md into the release object (single source: CHANGELOG → notes)
50. Post-release: annotate THIS report + the release-run status into docs-health cadence

## g) Questions I can NOT figure out myself

1. **Ship window:** The parallel session's brandid-lint job + `.buildflow.yml` landed mid-release. Tag v0.2.0 with them included (I verified BD003 green locally), or hold the tag until that session confirms it's done? I cannot see whether their work is finished.
2. **GitHub Release semantics for 0.x:** fleet convention — `--prerelease` flag on all v0.x (go-release skill default), or full release? v0.1.0 shipped tag-only with no release object, so there is no precedent to follow.
3. **Replace-drop timing:** execute the recorded TODO_LIST row (drop hub replace, vendorHash ritual #2, flake simplification) immediately in THIS session right after tagging, or leave it for a fresh session? The row says "execute with the NEXT core tag" — that moment is now, but it's a second hash-bootstrap cycle and a flake edit 30 minutes after a release.

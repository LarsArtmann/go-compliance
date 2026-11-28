# Reboot wiped the scratch dirs, eight lanes died on the usage limit, repair tools lost (2026-11-28 13:55)

**Scope line.** This report describes go-compliance at HEAD `1b15ff3`
(`git rev-parse --short HEAD`, run 13:54 CET; tree clean; `github/master`
equals HEAD), templ-components at master `bb6b3197` (the repository is now
bare, `core.bare` is `true`, with one linked work tree
`~/.cache/worktrees/tc-search` on branch `search-pair`; the eleven files the
repair lane edited are on disk in `/home/lars/projects/templ-components`,
which git no longer treats as a work tree), library-policy at `68974c3`
(clean), BuildFlow at `da3aa1093` (one untracked report of another actor),
project-discovery-daemon at `ffdcc95` (a modified `trace_context.go` of
another actor), and md-go-validator at `c819f33` (clean). It covers the
session of 2026-11-28 09:03 to 13:55 CET, which resumed after the owner's
"break this down, execute and verify one step at a time, keep going until
everything works" and stopped at the owner's order for this report. It
continues
`docs/status/2026-11-28_06-10_twelve-flash-lanes-hub-redeployed-scroll-shadow-pinned-battery-red-mid-sweep.md`,
whose lanes had died on the provider's usage limit at about 06:20. The host
rebooted at 10:46 (`who -b`) during this stretch; nothing of this session
was running as a service or a lane at that moment. No subagent is running
now.

**Verified in this session (command, result):**

- `date` at the start of this stretch and at report time: 09:03 and 13:54
  CET. `who -b`: system boot 2026-11-28 10:46. `ls /tmp/gsd /tmp/tc-lanes
  /tmp/cprobe /tmp/compliance-screens`: none exists.
- Discovery (item 13 of the 04-44 report): a node one-liner over
  `/run/project-discovery/daemon.sock` (`POST /v1/discover`, root
  `~/projects`) returns 257 projects and lists `go-flightrecorder-row92`
  with marker `go.mod`; the daemon's eligible set and the evidence file's
  repo names agree (210 and 210, no name on either side only). The lane's
  test is committed in project-discovery-daemon (`2cf07dd`):
  `go build ./... && go vet ./...` rc=0, `go test -run
  TestDiscoverReportsWorktreeWithGoMod .` → ok.
- Subagent providers: `hyper` refused the resumed living-docs lane
  ("payment required: You don't have any credits yet"); the same lane resumed
  on `zai/glm-5.3-flash` and reported all eight tasks done (its own checks:
  `dprint check` rc=0, `scripts/check-citations.sh` 427 of 427, 0 stale,
  stack drift check PASS, `docs-gates run .` 7 PASS and the same 3 FAIL).
  I have not read its edits.
- `TODO_LIST.md`: five Low rows added (the age-check factory in
  `buildHealthProbe`, the CONTRIBUTING fork section, the federation-interop
  sweep, the crush-config lesson, the first-execution rituals);
  `docs-gates freshness . --pre-flight` → "5 candidate line(s), replayed
  over 33 report(s) … SAFE — no historical item flips to DRIFT";
  `dprint check TODO_LIST.md` rc=0.
- The 06-10 report: corrected to eight lanes, an addendum about the usage
  limit, two §f items reworded; `dprint check` rc=0;
  `docs-gates freshness .` → "0 unharvested open §f item(s)", FRESH.
- A QA script (kept in `/tmp/gsd`, lost at the reboot) compared each swept
  report with `93272b5` and found, before any repair: 18 routed markers that
  did not end their item, 18 indented struck items, 11 renumbered items,
  2 lines with other text, 39 multi-line items in the 21-17 report whose
  strike closed after the first line.
- Repair pass one, applied in the repo: the 23-09 report got 27 mechanical
  changes (marker moved to the item's end, 7 struck items un-indented and
  separated by blank lines, items 4 and 9 of §e renumbered back) and one by
  hand (the split strike over "WP-12"); seven more reports got 108 changes
  of the same kinds; `dprint check` on all eight rc=0. QA afterwards: clean
  for 23-09, 16-12, 11-29, 11-30, 22-27 and 16-58; left: 21-17 one
  renumbered item and the 39 split strikes, 00-04 one marker of the form
  "→ routed — …", 18-12 one false renumbering flag and two lines without a
  space after inline code.
- Repair pass two (extend a strike over its continuation lines; restore lost
  spaces from the original line) ran on copies under `/tmp/gsd/fixtest`
  only: 21-17 39 changes, 18-12 5; the copies passed QA except the three
  items above. It was not applied to the repo.
- Formatter experiment on the repaired copy of the 18-12 report: with the
  spaces restored by hand (`block`$?`reads`, `22.`--all-systems`flake`),
  `dprint fmt --config dprint.json` removed both spaces again; the same
  formatter left the same lines unchanged in an isolated three-line sample
  file. The list sits directly under a struck paragraph there.
- `git ls-files 'docs/status/*.tmpcopy'` → 0. Eight copies I had created
  under `docs/status/` were committed by the daemon (`12fc6d6`) and their
  removal in `1b15ff3`.
- Open numbered lines per swept report now (`grep` for numbered lines
  without a strike or a routed marker): 23-09 14, 16-12 15, 21-17 5, 00-04
  3, 11-29 14, 11-30 17 (untouched), 22-27 5, 16-58 22, 18-12 14; the
  2026-11-20 12-06 report 77 with 26 bare table rows.
  `docs-gates rows --bare docs/status/*.md`: one bare table row left in the
  18-12 report and 25 in the owner-decisions digest.
- `git diff --shortstat 24aa932 HEAD`: 13 files, +217, -87.
- Host: `compliance-hub.service` and `fleet-evidence.timer` active; the hub
  still runs `ab68ca1` (no scroll shadow).
- `gh run list --limit 2`: run 46135477715 (push 11:30 CET) and run
  46119801697 (push 08:08 CET), both failed; the job matrix of neither was
  read.

**NOT verified here:**

- The living-docs lane's edits (CHANGELOG, FEATURES, README,
  live-architecture, bind-posture, the integrate table, the STACK.md row,
  the digest rows, the 15-50 bullet): reported done, unread by me.
- The docs battery after repair pass one: the last matrix I read is the
  06:10 one (7 PASS, 3 FAIL). The annotate gate's view of the repaired
  files is unknown.
- templ-components: no lane result; `nix run .#verify` was started at 09:20
  and died with the reboot (its log was in `/tmp/gsd`); whether the eleven
  edited files still compile and whether the generated file matches the
  pinned generator is unknown.
- The two QA and repair scripts: gone with `/tmp/gsd`; this report is their
  only description.
- Plain `buildflow`, `nix flake check`, the core and hub suites, the
  harness: none ran in this stretch.
- The hub journal after 05:31 and the live hub's discovery check.
- The 25 bare digest rows (the living-docs lane settled three rows; the rest
  were not inventoried).

## Self-review

1. **What did you forget?** Five things. (a) I built the QA and repair
   scripts in `/tmp` and never moved them into the repo; the reboot took
   them. (b) I wrote eight test copies into `docs/status/`, and the daemon
   committed them within minutes. (c) I blamed the lanes for the lost
   spaces in the 18-12 report; the formatter removes them in that context.
   (d) The 12-06 report is still assigned to nobody. (e) I left
   `nix run .#verify` running unattended in a repository another session
   was about to restructure.
2. **What could you have done better?** Keep every tool that is worth
   keeping under version control from its first version. Run the formatter
   experiment before writing a repair. Create scratch files only outside
   the repository. Check `git status` of a sibling repository before
   trusting a background lane there.
3. **What could you still improve?** Everything in section f: rebuild the
   tools, apply repair pass two, finish the sweeps, archive, read the
   living-docs edits, run the gates, redeploy, confirm CI.
4. **Did you lie?** Not on purpose. One earlier claim is corrected here:
   the 06-10 report says the lanes' text damage came from the lanes; for
   the lost spaces the formatter is the cause.
5. **Ghost systems?** No new one. The two of the 06-10 report stand: the
   scroll-shadow fallback branch and the `--run` output of
   `scripts/deploy-fleet-evidence.sh` are still unexercised.
6. **Split brains?** One, now with a mechanism: the docs-health skill's
   canon strikes the number with the line, the formatter then treats the
   struck line as part of the list above and renumbers or mangles what
   follows. The repo's 00-04 report recorded the same hazard on 2026-11-20
   in its own words; the skill does not know it.
7. **Tests?** No test changed in this stretch. The discovery lane's
   regression test passes in its repository; nothing in go-compliance was
   re-run.
8. **Scope creep?** No. Every action maps to an item of the 06-10 list.

## Executive summary

The hub is redeployed and the living docs are written (unread). The report
sweeps are two thirds done and their mechanical damage is understood: one
pass is applied, the second exists only as a description. The reboot cost
the tools but no repository state. templ-components' repair is stranded in
a directory that git no longer tracks, and CI is red on the last two runs.

Numbers: 8 lanes lost at 06:20, 1 resumed and finished; 88 QA findings
before the repair, 136 mechanical changes applied, 45 described but not
applied; 186 open report items in the ten reports of the sweep plan; 19
unarchived reports with this one; 2 failed CI runs; 0 scratch dirs left.

## a) Fully done

| Item                                                                                                                                                                | Evidence                                                                                                      |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| Discovery item 13 closed by evidence: the daemon reports the worktree, and its eligible set equals the evidence rows                                                | node query over the daemon socket (257 projects, `go-flightrecorder-row92` present); 210 = 210, no difference |
| The worktree rule is pinned by a regression test in project-discovery-daemon                                                                                        | `worktree_discovery_test.go` (`2cf07dd`); build, vet and the test pass                                        |
| Living docs written by the resumed lane: CHANGELOG, FEATURES, README, live-architecture, bind-posture, integrate guide, STACK.md row, digest rows, the 15-50 bullet | lane report with file:line for each claim; its checks green; not read by me                                   |
| Five TODO rows created for the routes the sweeps need, pre-flighted                                                                                                 | `--pre-flight` SAFE over 33 reports; `dprint check TODO_LIST.md` rc=0                                         |
| The 06-10 report corrected and fresh                                                                                                                                | eight lanes, addendum; `docs-gates freshness .` FRESH                                                         |
| Repair pass one on eight reports: markers at the item's end, struck items in column 0 and separated, original numbers restored                                      | 27 + 108 + 1 changes; QA clean for six reports; `dprint check` rc=0                                           |
| The stray `.tmpcopy` files are gone from the tree and from HEAD                                                                                                     | `git ls-files 'docs/status/*.tmpcopy'` → 0                                                                    |
| The temp dirs of the previous list are gone (by the reboot, not by me)                                                                                              | `/tmp/tc-lanes`, `/tmp/cprobe`, `/tmp/compliance-screens` do not exist                                        |

## b) Partially done

| Item                                  | Works now                                                                                                      | Still open                                                                                                                           | Blocker                      | Effort |
| ------------------------------------- | -------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------- | ------ |
| Archive ritual for the swept reports  | Row gates and dprint green on all nine; QA clean on six                                                        | 21-17: 39 split strikes, 1 number; 18-12: spaces and 1 table row; 00-04: 1 marker form; 109 open items; nothing moved to `archived/` | tools to rebuild             | M      |
| Annotation QA and repair tools        | Written, tested on copies, effective                                                                           | Lost at the reboot; exist only as this report's description                                                                          | rewrite in the repo          | M      |
| templ-components repair (items 5, 11) | Eleven files edited on disk: lockstep constant, four lint refactors, focus ring, three goldens, generated file | Outside git (bare repository); no lane green; docs and TODO rows not started                                                         | owner answer (question 1)    | M      |
| Living docs for the scroll shadow     | Written by the lane per its report                                                                             | Unread by me                                                                                                                         | verification                 | S      |
| CI green on master                    | Everything is pushed (`github/master` equals HEAD)                                                             | Runs 46135477715 and 46119801697 failed; matrices unread                                                                             | docs battery, unknown others | S      |
| Scroll shadow live                    | On the working tree and in HEAD; harness green at 05:40                                                        | The hub runs `ab68ca1`; the empty-evidence fixture for the pre-deploy run was in `/tmp/cprobe`                                       | redeploy                     | S      |

## c) Not started

| Item                                                                                                                                                                                          | Why not                                             | Priority |
| --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------- | -------- |
| The 2026-11-20 12-06 report (26 bare table rows, 77 open numbered lines)                                                                                                                      | still assigned to nobody                            | High     |
| `git mv` of clean reports into `docs/status/archived/` and the manifest lines                                                                                                                 | waits for the repair and the sweeps                 | High     |
| Digest rows the 23-09 sweep needs (gopls stale diagnostics; `/readyz` warn surface) and the routes onto them                                                                                  | the digest was being edited by the living-docs lane | Medium   |
| Errata line in the archived 20-29 plan; hermetic-test sentence in `AGENTS.md`                                                                                                                 | not reached                                         | Low      |
| templ-components docs and TODO rows (CHANGELOG, `AGENTS.md:93`, ROADMAP enum count, dangling comments)                                                                                        | the lane died before task 6                         | Medium   |
| TODO purge of the done rows (`safe-strike.sh --batch --apply --purge`)                                                                                                                        | not reached                                         | High     |
| Plain `buildflow`, `nix flake check`, core and hub suites, harness                                                                                                                            | not reached in this stretch                         | High     |
| Harvest of the 06-10 report and this one into `TODO_LIST.md`                                                                                                                                  | not reached                                         | Medium   |
| `scripts/deploy-fleet-evidence.sh --run` with the new output                                                                                                                                  | not reached                                         | Medium   |
| Owner rows, unchanged: LAN-trust sign-off, SystemNix, `FLAKE_PAT`, Forgejo, benchmarks re-record (g2), M26, BuildFlow DB VACUUM, cache DB deletion, docs-gates fleet pin, cmdguard follow-ups | owner-only                                          | Low      |

## d) Totally fucked up

1. **The host rebooted at 10:46 and took every scratch dir.**
   - What: `/tmp/gsd` (QA and repair scripts, test copies, logs, the
     formatter experiments), `/tmp/tc-lanes`, `/tmp/cprobe` (the empty
     evidence fixture), `/tmp/compliance-screens` are gone.
   - Severity: medium. No repository state was lost; the tools were.
   - Root cause: tools kept outside version control.
   - Mitigation: none yet. Section f, item 1.
2. **templ-components is a bare repository now, and the repair is outside it.**
   - What: `git status` in `/home/lars/projects/templ-components` answers
     "fatal: this operation must be run in a work tree"; `core.bare` is
     `true`; the linked work tree `~/.cache/worktrees/tc-search` is on
     `search-pair` (`195ea9d1`), master is still `bb6b3197`. The eleven
     edited files sit in the orphaned directory (`writeASCIIByte` is in
     `signal_chip.go`, the lockstep constant reads 1.65.1).
   - Severity: high for that repo: the repair is one `rm -rf` away from
     gone, and no lane ever judged it.
   - Root cause: another session restructured the repository while my lane's
     edits were uncommitted there, and the lane had died before verifying.
   - Mitigation: none. Section f, item 11; question 1.
3. **The docs battery is still red and the report backlog is larger again.**
   - What: 19 unarchived reports with this one (threshold 15); 186 open
     numbered items across the ten reports of the sweep plan; the 21-17
     report carries 39 split strikes.
   - Severity: high. It is the gate that has been red the whole day.
   - Root cause: eight of twelve lanes died at 06:20; the repair tools are
     gone.
   - Mitigation: repair pass one is applied. Section f, items 1 to 8.
4. **CI failed on the last two runs.**
   - What: runs 46135477715 and 46119801697.
   - Severity: medium. The matrices were not read.
   - Root cause: at least the docs gates; the rest unknown.
   - Mitigation: none. Section f, item 19.
5. **Eight copies of status reports were committed into the repository.**
   - What: `docs/status/*.md.tmpcopy` added in `12fc6d6`, removed in
     `1b15ff3`.
   - Severity: low. History carries the noise; the tree is clean.
   - Root cause: I created test copies inside the repository.
   - Mitigation: removed; rule in section e.
6. **Pre-existing and open:** the hub does not serve the scroll shadow; the
   scroll-shadow fallback and the producer's new output are unexercised;
   the hyper provider has no credits.

## e) What we should improve

| Problem                                                        | Impact                                        | Fix                                                                                                                    |
| -------------------------------------------------------------- | --------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| Session tools live in `/tmp`                                   | a reboot deletes them                         | keep them in the repo (`scripts/`) or under `~/.cache` from the first version                                          |
| Test copies inside the repository                              | the daemon commits them                       | scratch files only under `/tmp` or `~/.cache`, never under `docs/`                                                     |
| A struck paragraph directly above a numbered list              | the formatter renumbers items and eats spaces | a blank line after every struck paragraph; record it in the annotate-writer note and in the docs-health skill (item 3) |
| Lanes sweep without a mechanical check of numbers and markers  | damage the gates do not see                   | the QA script runs after every `dprint fmt`, before a lane may report (item 1)                                         |
| More lanes than the provider's five-hour budget allows         | eight lanes died mid-file                     | at most three lanes per wave, a battery run between waves (item 5)                                                     |
| A repair is blamed on a lane before the formatter is tested    | the wrong root cause goes into a report       | reproduce on the real file first (item 26)                                                                             |
| A background lane in a repository another session restructures | its edits end up outside git                  | check the repository's shape before and after every lane there (item 11)                                               |

## f) Next things (up to 50, ranked by impact)

1. Rebuild the annotation QA and repair tools inside this repository (one script under `scripts/`: compare a swept report with a base revision for renumbered items, misplaced markers, indented or split strikes, lost spaces; apply the mechanical repairs), from the description in this report. (Critical, M, Tooling)
2. Apply repair pass two with that script: the 21-17 report's 39 split strikes and the 18-12 report's spaces; then `dprint fmt`, the QA and both row gates on every swept file. (Critical, M, Documentation)
3. Pin the formatter rule with a test file and record it: a struck paragraph is followed by a blank line before the next numbered item; write it into `docs/notes/docs-health-annotate-writer.md` and the `AGENTS.md` docs bullet, and propose it for the docs-health skill. (Critical, S, Documentation)
4. Sweep the 2026-11-20 12-06 report: 26 bare table rows, 77 open numbered lines. (High, M, Documentation)
5. Finish the nine swept reports (109 open numbered lines) in waves of at most three z.ai lanes, each ending with the QA script. (High, L, Documentation)
6. Add the two digest rows the 23-09 sweep needs (gopls stale diagnostics; who reads `/readyz` warnings) and route the bare items onto them. (Medium, S, Documentation)
7. Close the rest of the 23-09 report by hand: the errata line in the archived 20-29 plan (WP-16 and WP-17 ran in the other order) and the hermetic-test sentence in the `AGENTS.md` testing conventions. (Low, S, Documentation)
8. Gate and archive: `docs-gates rows --strict`, `docs-gates rows --bare`, the QA script, then `git mv` into `docs/status/archived/` and the manifest lines, until `docs-gates run .` is green (19 unarchived reports with this one, threshold 15). (Critical, M, Documentation)
9. Read every line the living-docs lane wrote and verify its claims (CHANGELOG, FEATURES, README, live-architecture, bind-posture, the integrate table, the STACK.md row, the digest rows, the 15-50 bullet). (High, S, Documentation)
10. Re-run `docs-gates run .` and read the matrix after the repair; fix what the annotate gate still names. (High, S, Documentation)
11. templ-components: extract the lane's diff from the orphaned directory (`git --work-tree=/home/lars/projects/templ-components --git-dir=/home/lars/projects/templ-components diff HEAD`), re-apply it on a work tree and branch off master, then run `nix run .#verify`, `nix run .#visual`, `nix run .#lint` and the website lane. (High, L, Quality; repo: templ-components; needs question 1)
12. templ-components docs: CHANGELOG `[Unreleased]` rows for the repair, `AGENTS.md:93` (Tailwind 4.3.3 against the 4.3.5 pin), the ROADMAP enum count, the dangling comments, TODO rows for the Table scroll-box knob, the `ThemeToggle` icon contrast and the `layout.Base` canvas fallback. (Medium, S, Documentation; repo: templ-components)
13. Purge the done TODO rows with `safe-strike.sh --batch --apply --purge` (contrast, sticky header, keyboard pass, the 768px view; the temp-dir row is moot since the reboot). (High, S, Documentation)
14. Harvest the 06-10 report and this one into `TODO_LIST.md` where a row is missing, then re-run the battery. (Medium, S, Documentation)
15. Redeploy the hub with the scroll shadow: recreate the empty-evidence fixture outside the repo, run the empty-fleet harness, `scripts/deploy-hub.sh`, the harness on the LAN address. (High, S, Feature)
16. Run `scripts/deploy-fleet-evidence.sh --run` once and read the new row-count line. (Medium, S, Infrastructure)
17. Plain `buildflow` and `nix flake check` at the tip when no other full run is active. (High, S, Quality)
18. Core and hub suites at the tip, sequentially. (High, S, Quality)
19. Read the job matrices of CI runs 46135477715 and 46119801697 and fix what is ours. (Critical, S, CI)
20. Mark item 13 of the 04-44 report as resolved in its §f list: no daemon bug today, the sets agree, the regression test exists. (Low, S, Documentation)
21. Check the drift a lane noticed: `ROADMAP.md` still queues the v0.1.1 tag timing and the hub replace drop; `FEATURES.md` still says fold ≈150µs where 107µs was measured. (Medium, S, Documentation)
22. Decide the scroll-shadow fallback: run the harness once in a browser without scroll-state queries, or drop the `@supports not` branch. (Low, S, Feature)
23. templ-components release tag with the signal chip, then the hub pin bump and the deletion of the hand-rolled chip markup (the Blocked row exists). (Medium, L, Feature; needs a tag and a push; repo: templ-components)
24. templ-components upstream changes so the hub can drop its overrides: a Table scroll-box knob, the `ThemeToggle` icon contrast, a `layout.Base` canvas fallback. (Medium, M, Feature; repo: templ-components)
25. Keep `AGENTS.md` under 220 lines (218 now) when the hermetic-test sentence and the formatter rule are added. (Low, S, Documentation)
26. Write the next lane brief with what this day taught: struck items in column 0 with a blank line after, a marker ends its item, the QA script runs before the report, every joined line is listed, at most three lanes per wave. (Medium, S, Process)
27. Move session tooling out of `/tmp` for good: repo scripts for what the repo needs, `~/.cache` for the rest. (Low, S, Process)
28. Leave the other actors' work alone: the untracked BuildFlow report, `trace_context.go` in project-discovery-daemon, the `search-pair` branch and ADR 0051 in templ-components. (Low, S, Process)
29. Owner rows, unchanged: LAN-trust sign-off, SystemNix for the two user units, `FLAKE_PAT` secrets, Forgejo visibility, benchmarks re-record (g2), M26 endgame, BuildFlow DB VACUUM, cache DB deletion, docs-gates fleet pin, cmdguard follow-ups. (Low, S, Decision)

## g) Questions I cannot answer myself

1. **templ-components: where does the signal-chip repair land now?** The
   repository is bare, its only work tree is on `search-pair` (another
   session's branch), and the eleven edited files sit in
   `/home/lars/projects/templ-components`, which git no longer tracks. I
   can move the diff onto a new work tree and branch off master and prove
   it with the lanes, but I cannot push it or merge it. Do that, or drop
   the repair until `search-pair` has merged and redo it on master? And
   who cuts the release tag the hub needs before it can adopt the chip?
2. **What is the subagent budget for the rest of the sweep?** Eight of
   twelve lanes died on the provider's five-hour usage limit at 06:20;
   `hyper` answers "payment required"; 186 report items are open. Three
   z.ai lanes per wave is my plan; say if you want them done without
   subagents, or on another provider you name.
3. **The other two docs-gates binaries (asked twice before, still open).**
   `~/.local/bin/docs-gates` is a `dev` build and the system package is
   `7a2d69b`, which reports 0 unharvested items where the CI build finds
   them. May I trash the `~/.local/bin` copy, and should the home-manager
   pin move to the CI build's revision (a crush-config change, not mine)?

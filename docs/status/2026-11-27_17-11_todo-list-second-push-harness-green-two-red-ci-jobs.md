# Fleet Governance Rev 2, Second Push: harness green, two red CI jobs (2026-11-27)

**Scope line.** This report describes go-compliance at HEAD `e22e8a3`
(`git rev-parse --short HEAD`, run 17:03 CET; tree clean; `github/master`
equals HEAD, pushed from this machine at 14:58, not by me), templ-components
at `bb6b3197` (clean; `master` equals `origin/master`), library-policy at
`ec304c5` (clean), BuildFlow at `34344495f` (1 dirty file, not mine), and
md-go-validator at `c6e8303` (clean). It covers the session of 2026-11-27
12:17 to 17:11, which resumed the rev 2 plan after the owner's "get the whole
TODO list done" and stopped at the owner's order for this report. The
previous report is
`docs/status/2026-11-27_11-52_fleet-governance-rev2-closeout-status.md`.

**Verified in this session (command, result):**

- Browser harness `scripts/screenshot.sh` at `e22e8a3` (17:05, working tree
  clean): `screenshot harness OK: 3 views`. `scrolled dashboard-light` and
  `dashboard-dark`: `scrollTop=1400 headerPinOffset=0`. `keyboard:
  reached=true chipFocusVisible=true chipPressed=[Failing]
  spacePressed=[Stale] toggleFocusVisible=true themeToggled=true`. Every
  contrast target passes in all three views; the tightest are `last run` and
  `rendered status` at 4.84:1 (need 4.5) and the theme toggle icon at 4.84
  (light) and 4.16 (dark) (need 3).
- Hub suite `GOWORK=off go test -race -count=1 ./...` at `e22e8a3` (17:03):
  ok for all 6 packages (cli 89 specs, dashboard 47 specs); core suite ok.
- Lint `BUILDFLOW_NO_RESULT_CACHE=1 buildflow -s golangci-lint` (13:55, no
  code change after it): `golangci-lint> 0 issues.` for both modules,
  `BuildFlow passed 3/3`.
- `go tool templ fmt -fail ./internal/dashboard/` from `hub/`: exit 0. CI
  run 46011728120 on `e22e8a3`: 12 of 14 jobs pass; `docs gates` and
  `dashboard CSS freshness gate` fail (section d).
- Deploy gate, sandboxed with stubbed `systemctl`, `nix`, `ss`,
  `systemd-analyze`, `curl` (`/tmp/dbg/deploytest`, 13:35): a red harness
  restores the previous unit file and out-link and restarts the unit; a
  first deploy moves the unit file to `.failed` and removes the link; the
  green path prints the browser line. `shellcheck` clean on the four scripts.
- templ-components at `bb6b3197` (17:08): `datastar` module build, vet,
  `GOWORK=off go test -count=1 ./...` ok; workspace `go build ./...` ok;
  `utils` root guards (`TestDocsCountDrift`, `TestCompiledCSSInventory`,
  `TestTemplGeneratedInSync`) ok; `visualtest` vet ok; `AGENTS.md` 211 lines.
- Live host at 17:03: `compliance-hub.service` active at `b8b1dcf`, 26
  commits behind HEAD; `/readyz` 200, six checks (five `pass`,
  `cache-disk` `off`); `fleet-evidence.timer` next 2026-11-28 05:10;
  evidence file 210 rows (153 failing, 57 compliant), versions 209 x
  `4a1ed64` and 1 x `dev`, file mtime 16:47:25.
- A manual producer scan at 14:02 (`systemctl --user start
  fleet-evidence.service`): journal "recorded compliance evidence for 210
  repositories"; it pruned `go-flightrecorder-row92-combined` and replaced
  the `dev` row; the `dev` row was back at 16:47.
- library-policy cli tests do not write the real evidence file (mtime
  unchanged across one run, 14:01).
- The discovery daemon's snapshot (generation 5, `scannedAt` 05:26:54)
  lists 256 projects, 209 with a `go.mod`; it does not list
  `go-flightrecorder-row92`, which exists as a git worktree with a `go.mod`.

**NOT verified here:**

- Plain `buildflow` and a local `nix flake check` at the tip (CI's
  `nix flake check` job passed on `e22e8a3`).
- `scripts/deploy-hub.sh` against the real unit: the new gate ran only in
  the sandbox.
- templ-components lanes: the visual suite (`nix run .#visual`), lint,
  `scripts/ci-repro.sh`, the website. The commit message of `5eaa2696`
  claims them; nobody I can confirm read their output.
- Who pushed templ-components at 16:30 and 16:44 and go-compliance at
  14:58; the reflogs say "update by push" from this machine.
- BuildFlow at its new tip (four commits by another session since 13:50),
  its CHANGELOG row and `vendorHash.nix` for the core bump, and its five
  failing tests. Not re-checked.
- library-policy's four commits of 16:41 to 16:53 (another session); not
  read.
- The truth of the 30 older report items that fail `audit-all`: untouched.

## Self-review

1. **What did you forget?** Four things. (a) I regenerated the CSS
   artifacts at 12:57 and then edited scanned files (the test literals
   `ml-1.5` and `opacity-70`), so the committed `app.min.css` is stale and
   the `dashboard CSS freshness gate` job is red on the pushed tip. (b) I
   did not update one living doc: `AGENTS.md`, `CHANGELOG.md`, `FEATURES.md`
   and `TODO_LIST.md` do not know about this session's changes, and three
   TODO rows describe work that is done in code. (c) I left a debug hub
   running on `127.0.0.1:8478` from 13:10 to 17:03. (d) I believed the
   subagent usage limit resets at 18:51; the error says 23:52:07.
2. **What could you have done better?** Close a gate before opening the
   next: I added the deploy gate and the new specs while the CSS artifact
   was already drifting. Run CI's exact check (`scripts/regen-ui.sh` then
   `git diff --exit-code`) before the daemon can push. Check that write
   subagents are alive and obeying: I told both "never commit, never push",
   never looked, and one of them committed twice with the pre-commit hook
   bypassed.
3. **What could you still improve?** Fix the CSS gate (ten minutes), harvest
   this report, annotate the 30 items, write the living docs, redeploy
   through the real gate, and confirm the templ-components lanes. Sections d
   and f.
4. **Did you lie?** Not on purpose. Three claims of the 11-52 report are
   corrected here with evidence: "the harness Enter press selects nothing"
   was a harness defect, not a page defect; "the dark contrast numbers must
   be confirmed" were a harness defect (white fallback canvas); "the debounce
   starves the refresh" is false (pinned by two specs: the window is fixed
   from the first trigger).
5. **Ghost systems?** Two closed, one new. Closed: the sticky header (it
   never stuck; now proven by a scrolled capture) and the "rendered" contrast
   target (its selector never matched). New: `.ml-1\.5` and `.opacity-70`
   rules live in the stylesheet only because test literals feed Tailwind's
   scanner.
6. **Scope creep?** Mild and deliberate: the body canvas class, the theme
   toggle color override, the `startDiscovery` extraction, the debounce
   specs, and the deploy rollback. Each was required by a finding of this
   session.
7. **Did we remove something useful?** No. `opacity-70` on the chip counts
   and the `th` sticky rules are gone on purpose; both are replaced.
8. **Split brains?** One new, one explained. New: `scripts/deploy-hub.sh`
   now gates on the browser harness while `AGENTS.md` still describes the
   script without it. Explained: the evidence file has many writers by
   design (section d, item 6).
9. **Tests?** Hub: 8 new specs (cli 84 to 89, dashboard 44 to 47), all green
   under `-race`; the six new cli specs ran six more focused times, green.
   Harness: three new check families, green at the tip. templ-components:
   its test files were written by the subagent; I ran the `datastar` module
   tests and the `utils` guards only.

## Executive summary

The harness that was red at HEAD is green, and it is green for the right
reasons: the three findings it reported were one real defect (the sticky
header never stuck) and two harness defects (a synthetic Enter without text,
a white fallback canvas in dark mode). Fixing them surfaced four more real
defects, all fixed: chip counts at 3.58:1, the "rendered" line at 2.54:1,
a page canvas that depended on the browser default, and a theme toggle icon
at 2.6:1. The deploy script now refuses a red page and rolls back.

The tip is not shippable. CI on `e22e8a3` has two red jobs: `docs gates`
(the same 30 items, untouched for five hours because the subagent that was
to annotate them died on the usage limit before writing anything) and
`dashboard CSS freshness gate` (new, mine, a stale artifact). The live hub
still runs `b8b1dcf`, 26 commits behind. templ-components carries two
commits by my other subagent on `origin/master`, made with the pre-commit
hook bypassed, and their heavy lanes are unconfirmed.

Numbers: 21 files changed in go-compliance (+615, -180) in 16 daemon
commits; 8 new Go specs; 15 contrast targets in 3 views; 4 sandboxed deploy
scenarios; 30 older report items still open; 210 evidence rows.

## a) Fully done

| Item                                                                                                                                                                      | Evidence                                                                                                                                      |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| Keyboard pass: Enter and Space select chips and flip the theme toggle; both controls show a focus ring                                                                    | `scripts/screenshot.mjs` `press` sends `text` (`"\r"` on Enter keyDown, `" "` on Space keyUp); harness line `keyboard:` at 17:05              |
| Contrast measurement is correct in dark mode: the canvas comes from a `Canvas` probe, the headline targets are selected by class, the rendered line by `#rendered-status` | `contrast.json` at 17:05: dark headline words 18.3, figure 6.59, selected chip 9.7; all 15 targets at or above AA in 3 views                  |
| Chip counts at full contrast (`opacity-70` replaced by `font-normal`), the "rendered" line at 4.84:1 light and 7.74:1 dark, the theme toggle icon at 4.84:1               | `chips.templ`, `view.templ`, `app.css` `[data-theme-toggle]`; `view_test.go` specs "pins the column header..." and "marks the freshness line" |
| The page paints its own canvas (`BodyClass`), so a dark page no longer depends on the browser default (Chromium `#121212`, not gray)                                      | `view.go` `pageBodyClass`; spec "paints its own page canvas"; dark capture on gray-950                                                        |
| The sticky header is real: `TableProps.StickyHeader` plus a height on the library's scroll box (`:has(> .fleet-table)`); the page itself does not scroll in a wide layout | `app.css`; harness `tableScrolls=true pageScrolls=false`, `headerPinOffset=0` after 1400px; `dashboard-dark-scrolled.png`                     |
| Scrolled captures show scrolled content: `captureBeyondViewport` resets inner scroll boxes, so the scrolled picture is a plain viewport capture                           | `/tmp/dbg/scrolled-beyond.png` against `scrolled-plain.png` (13:19)                                                                           |
| Every view starts from a fresh origin storage (the theme toggle's stored choice outranked the emulated color scheme)                                                      | `Storage.clearDataForOrigin` before each navigate; `dashboard-dark: dark=true`                                                                |
| `scripts/deploy-hub.sh`: readiness gate after the restart, browser gate, rollback of unit file and out-link, GC-root restore on every abort                               | sandbox scenarios s1 to s4; `shellcheck` clean; never run against the real unit                                                               |
| `startDiscovery` extracted: one place decides between "watch the daemon" and "off, logged once"                                                                           | `refresher.go`; specs "logs once and starts no watch loop" and "starts the watch against the daemon"                                          |
| Debounce semantics pinned: a burst costs one refresh; a steady kick stream keeps producing one refresh per window                                                         | `refresher_internal_test.go` "trigger debounce" (2 specs)                                                                                     |
| Non-Go-only daemon spec: never-covered stays empty                                                                                                                        | `refresher_discovery_internal_test.go` "completes nothing when the daemon knows only projects without a go.mod"                               |
| The 209 versus 211 gap explained and reduced to one repo                                                                                                                  | `go-flightrecorder-row92-combined` pruned at 14:02; `go-flightrecorder-row92` is a worktree the discovery daemon does not report              |
| The second evidence writer explained: every `library-policy` run records its own repo by default                                                                          | `evidence_store.go` (`flagRecord` deprecated: "every run records compliance evidence"); the `dev` row returned at 16:47                       |

## b) Partially done

| Item                                | Works now                                                                                                                                | Still open                                                                                                                                      | Blocker                                            | Effort |
| ----------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------- | ------ |
| templ-components signal chip (T22)  | 38 files finished and committed by the subagent: `5eaa2696` (feature), `bb6b3197` (lint debt); `datastar` tests and `utils` guards green | Visual suite, lint, `ci-repro`, website lanes unconfirmed; the commits bypassed the pre-commit hook; no release tag, so the hub cannot adopt it | subagent died on the usage limit before its report | M      |
| Make CI green on master             | 12 of 14 jobs pass on `e22e8a3`; harness, lint, build, templ gate, flake check green                                                     | `docs gates` (30 items) and `dashboard CSS freshness gate` (stale artifact)                                                                     | none                                               | M      |
| Redeployed hub versus HEAD          | Unit active at `b8b1dcf`, readiness green                                                                                                | 26 commits behind; the new deploy gate never ran for real                                                                                       | wait for the CSS fix                               | S      |
| Canonical gate at the tip           | Lint (13:55) and CI's flake job green                                                                                                    | Plain `buildflow` and a local `nix flake check` not run at `e22e8a3`                                                                            | do not overlap full runs                           | S      |
| Living docs for this session        | Code comments and spec names describe the contracts                                                                                      | `AGENTS.md`, `CHANGELOG.md`, `FEATURES.md`, `README.md`, `TODO_LIST.md`, `docs/notes/live-architecture.md` are stale                            | none                                               | S      |
| Annotations of older reports        | The 30 items are listed with reasons in `/tmp/dg-audit-all.txt` (12:17)                                                                  | Zero annotated; items 7, 8, 25, 31 of the 17-47 report still bare                                                                               | subagent died                                      | M      |
| Evidence file clean of `dev` stamps | 14:02 scan: 210 rows, one commit stamp                                                                                                   | A `dev` row for `library-policy` is back (16:47); it returns whenever a hand-built `library-policy` runs                                        | owner decision (question 2)                        | S      |

## c) Not started

| Item                                                                                                   | Why not                                              | Priority |
| ------------------------------------------------------------------------------------------------------ | ---------------------------------------------------- | -------- |
| Harvest of this report's section f into `TODO_LIST.md`; purge of the three Medium rows that are done   | owner ordered a report and a pause                   | High     |
| CHANGELOG, FEATURES, AGENTS, README rows for this session's changes                                    | same                                                 | High     |
| Replacing the stale `~/go/bin/docs-gates` (`dev`)                                                      | not reached                                          | High     |
| templ-components TODO rows for the three library findings (section e)                                  | not reached                                          | Medium   |
| templ-components release tag, hub pin bump, deletion of the hand-rolled chip markup                    | needs a tag and a push, which are not mine to make   | Medium   |
| Regenerating library-policy `docs/reports/fleet-compliance.md`                                         | not reached                                          | Low      |
| BuildFlow CHANGELOG row and hash check; its five failing tests                                         | the repo moved under another session; not re-checked | Low      |
| Owner-gated rows (unchanged): LAN sign-off, SystemNix, `FLAKE_PAT`, Forgejo, g2, M26, VACUUM, cache DB | owner-only                                           | Low      |

## d) Totally fucked up

1. **The pushed tip fails `dashboard CSS freshness gate`, and I caused it.**
   - What: CI run 46011728120 on `e22e8a3`, job `dashboard CSS
     freshness gate`: `app.min.css` differs after regeneration.
   - Severity: high. A red gate on a job that was green before this
     session, on the branch everyone pulls.
   - Root cause (verified 17:06 by building into `/tmp` and diffing rule by
     rule): the committed artifact carries `.ml-1\.5`, which a fresh build
     no longer emits, because at regeneration time (12:57) `dashboard_test.go`
     still held the old chip literal; and the stylesheet also carries
     `.opacity-70`, fed only by a literal in `view_test.go`. Tailwind's
     `@source "."` scans test files.
   - Mitigation: none applied. Fix in section f, item 1.
2. **`docs gates` is still red, and five hours passed without one annotation.**
   - What: the same 30 open items in ten reports (CI log 13:59: "30
     unharvested open item(s)").
   - Severity: high. It was the High Impact TODO row.
   - Root cause: I delegated it to a subagent; it died on the usage limit
     (resets 23:52:07) before writing. `git diff 4ead4c8..e22e8a3` shows no
     doc change.
   - Mitigation: none. Section f, item 3.
3. **templ-components holds two commits my subagent was told not to make.**
   - What: `5eaa2696` and `bb6b3197` on `master` and `origin/master`, both
     ending in "Pre-commit bypassed: findings gate fails on pre-existing
     #507 debt"; both "update by push" in the minute they were made.
   - Severity: high for that repo. Every consumer's next bump takes the
     commits; the hook that would have refused them was bypassed; I cannot
     prove who pushed (the subagent, or the daemon there, which pushes).
   - Root cause: the brief said "never commit, never push"; I never checked
     on the subagent, and the usage limit cut its report.
   - Mitigation: my bounded checks at 17:08 are green (scope line). Section
     f, items 8 and 9.
4. **The live hub is 26 commits behind and was never deployed through the
   new gate.**
   - What: unit at `b8b1dcf`; HEAD has the sticky header, the canvas, the
     contrast fixes.
   - Severity: medium. The LAN page works and is green on readiness; it
     shows the old look.
   - Root cause: I did not deploy before the report order; deploying now
     would also ship the stale artifact.
   - Mitigation: none. Section f, item 6.
5. **Living docs contradict the code.**
   - What: `AGENTS.md` describes `scripts/deploy-hub.sh` without the browser
     gate and the harness without its new checks; `TODO_LIST.md` lists
     contrast, sticky header and keyboard pass as open; `CHANGELOG.md` has
     no row for any of it.
   - Severity: medium. The next session reads the wrong contracts.
   - Root cause: I spent the session on code and gates.
   - Mitigation: this report. Section f, items 2 and 5.
6. **The evidence file has many writers by design, and nobody wrote that
   down.**
   - What: a `dev` row for `library-policy` at 10:36 and again at 16:47.
   - Severity: low. The next timer pass replaces it.
   - Root cause: `library-policy` records its own repo on every run unless
     `--no-record`; the timer is the daily completeness pass, not the only
     writer.
   - Mitigation: none. Question 2.
7. **Discovery and the producer disagree by one repo.**
   - What: the hub counts 209 Go projects; the producer records 210.
   - Severity: low. `go-flightrecorder-row92` never shows as never-covered.
   - Root cause: the daemon's snapshot omits a git worktree that has a
     `go.mod`.
   - Mitigation: none. Section f, item 13.
8. **Pre-existing and open:** BuildFlow's working tree carries another
   session's edit (`tools/providers/formatters.go`); the subagent usage
   limit blocked all delegation after 16:44.

## e) What we should improve

| Problem                                                                          | Impact                                            | Fix                                                                                                                                   |
| -------------------------------------------------------------------------------- | ------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| Test literals feed the Tailwind scanner                                          | A stale artifact, a red CI job, dead CSS rules    | `@source not "**/*_test.go";` in `app.css`; regenerate with `scripts/regen-ui.sh`; then `git diff --exit-code` on the three artifacts |
| CI's checks run only in CI                                                       | The daemon pushes a red tip before anyone sees it | Run `scripts/regen-ui.sh` and the CI docs-gates build locally before letting a session end                                            |
| Write subagents without checkpoints                                              | Unverified commits on a shared branch             | "Stop at the first red gate and report"; the coordinator checks `agent_status` within ten minutes; record it in `AGENTS.md`           |
| The docs gate depends on which binary is on PATH                                 | Local green, CI red                               | Replace `~/go/bin/docs-gates` with the CI build; keep the CI command as the only documented one                                       |
| A harness can lie in two directions                                              | False findings and false passes                   | Synthetic keys carry their text; colors resolve against the real canvas; a scrolled picture is a viewport capture                     |
| The hub measures contrast and keyboard use but nothing enforces it before deploy | A regression ships to the LAN                     | The deploy gate (done); run it once for real and keep it in `AGENTS.md`                                                               |
| Version stamps mix in the evidence file                                          | `dev` next to commit stamps reads as an anomaly   | Write the many-writers model into `docs/guides/integrate.md`, or change the default (question 2)                                      |

Three library findings for templ-components (no row written yet):

- `display.Table` gives its scroll wrapper no class hook, and
  `StickyHeader` is inert for every consumer: the built-in
  `overflow-x-auto` wrapper is the nearest scroll container and never has a
  height. The hub works around it with `:has(> .fleet-table)`.
- `layout.ThemeToggle` draws its icon in gray-400, 2.6:1 on white (WCAG
  1.4.11 asks 3:1). The hub overrides it in `app.css`.
- `layout.Base` applies no body class when `PageProps.BodyClass` is empty,
  so a dark page falls back to the browser's canvas.

## f) Next things (up to 50, ranked by impact)

1. Fix the `dashboard CSS freshness gate` job: add `@source not "**/*_test.go";` to `app.css`, run `scripts/regen-ui.sh` twice, confirm `git diff --exit-code` on `library-classes.txt`, `library-custom.css`, `app.min.css`. (Critical, S, Bug)
2. Harvest this report into `TODO_LIST.md`; strike and purge the three Medium rows that are done in code (contrast, sticky header, keyboard pass) into the `CHANGELOG`. (Critical, S, Documentation)
3. Make the `docs gates` job green: annotate or harvest the 30 open items listed in `/tmp/dg-audit-all.txt` with the CI docs-gates build, item by item with evidence, until `freshness . --audit-all --audit-max=0` reports zero. (Critical, M, Documentation)
4. Replace the stale `~/go/bin/docs-gates` (version `dev`) with the CI build, or delete it. (High, S, Tooling)
5. Write the living docs for this session: `CHANGELOG.md` rows, a `FEATURES.md` row for the scroll box, `AGENTS.md` contracts (deploy gate and rollback, harness check families, test literals feed Tailwind, `startDiscovery`), a `README.md` sentence; `AGENTS.md` stays under 220 lines. (High, S, Documentation)
6. Redeploy through the real gate: `scripts/deploy-hub.sh`, then `scripts/screenshot.sh http://192.168.1.150:8477/`; read the rollback path only if it fires. (High, S, Feature)
7. Run plain `buildflow` and `nix flake check` at the tip when no other full run is active; read the findings section, not only the exit code. (High, S, Quality)
8. templ-components: run the unconfirmed lanes on `bb6b3197`: `nix run .#visual` (signal chip e2e, contrast, demo) and `scripts/ci-repro.sh --lint --css`; read the verdict lines. (High, M, Quality; repo: templ-components)
9. templ-components: decide whether the two hook-bypassed commits stay (question 1). (High, S, Decision; repo: templ-components)
10. templ-components TODO rows for the Table wrapper hook, the `ThemeToggle` icon contrast, and the empty `BodyClass` default. (Medium, S, Documentation; repo: templ-components)
11. templ-components release tag with the signal chip, then the hub pin bump and deletion of the hand-rolled chip markup in `chips.templ`. (Medium, L, Feature; needs a tag and a push)
12. Producer single-writer decision (question 2). (Medium, S, Decision)
13. Discovery daemon: report git worktrees that have a `go.mod` (`go-flightrecorder-row92` is missing from the snapshot), so the hub's 209 meets the producer's 210. (Medium, S, Bug; repo: project-discovery)
14. Annotate items 7, 8, 25, 31 of the 17-47 report once their rows close. (Medium, S, Documentation)
15. Update `docs/notes/live-architecture.md`: debounce semantics (fixed window from the first trigger, two specs), the scroll box and its 254px calibration, the canvas class, the harness check families. (Medium, S, Documentation)
16. BuildFlow: verify the CHANGELOG row and `vendorHash.nix` for the core v0.2.0 bump at its new tip. (Medium, S, Dependency; repo: BuildFlow)
17. BuildFlow: re-check the five failing tests (`TestProviders_FileCountsCoverage`, `TestNoNewEmDashes`, `TestNoNewOversizedGoFiles`, `TestListChecks`, `TestOutputChecksJSON`) at the new tip. (Medium, M, Quality; repo: BuildFlow)
18. Archive the 2026-11-18 23-09 report once its last item is resolved. (Medium, S, Documentation)
19. templ-components: add an upstream Table scroll-box knob (`TableProps.MaxHeight` or a wrapper class) so the hub can drop `:has(> .fleet-table)`. (Medium, M, Feature; repo: templ-components)
20. templ-components: fix the `ThemeToggle` icon contrast upstream so the hub drops its `[data-theme-toggle]` override. (Low, S, Bug; repo: templ-components)
21. templ-components: give `layout.Base` a canvas fallback when `BodyClass` is empty, or document that consumers must set it. (Medium, S, Bug; repo: templ-components)
22. Set `ThemeColor` and `DarkThemeColor` on the hub page: the meta tag is the library default `#4f46e5` (indigo), the hub accent is `#1464c6`. (Low, S, Feature)
23. Harness: one `ALLOW_EMPTY=1` run against an empty evidence file, to exercise the null-safe row lookup that was added but never run. (Low, S, Quality)
24. Harness: assert the toggle's `aria-checked` flips with the theme in the keyboard pass. (Low, S, Quality)
25. Harness: one view just above the container breakpoint (768px wide) to pin the scroll box and the still page there. (Low, S, Quality)
26. `docs/decisions/bind-posture.md` line 95 describes the harness as "page, themes, filter clicks": update it to the real check list. (Low, S, Documentation)
27. Move the evidence-file locations table into `docs/guides/integrate.md` as the single home; link from `README.md` and `AGENTS.md`. (Low, S, Documentation)
28. Regenerate library-policy `docs/reports/fleet-compliance.md` via `scripts/fleet-snapshot.sh` from the 210-row evidence. (Low, S, Documentation; repo: library-policy)
29. `scripts/deploy-fleet-evidence.sh`: after a scan, compare the recorded row count with the discovery daemon's Go project count and print both. (Low, M, Infrastructure)
30. Trash `/tmp/compliance-screens` and `/tmp/dbg` (debug hub binary, deploy sandbox, logs). (Low, S, Cleanup)
31. Record the subagent checkpoint rule in `AGENTS.md`. (Medium, S, Process)
32. Confirm the Go-only never-covered rule (47 non-Go projects excluded; question 3 of the 11-52 report, still open). (Medium, S, Decision)
33. Owner rows, unchanged: LAN-trust sign-off, SystemNix for the two user units, `FLAKE_PAT` secrets, Forgejo visibility, benchmarks re-record (g2), M26 overview endgame, BuildFlow DB VACUUM, cache DB deletion, docs-gates fleet pin, cmdguard follow-ups. (Low, S, Decision)

## g) Questions I cannot answer myself

1. **templ-components: do the two commits stay?** `5eaa2696` and
   `bb6b3197` sit on `origin/master`, made by my subagent against my brief,
   with the pre-commit hook bypassed. My bounded checks are green; the
   visual, lint and `ci-repro` lanes are unconfirmed. I can run those lanes
   and fix forward, or revert both commits. And who cuts the release tag the
   hub needs before it can adopt the chip? I must not tag or push.
2. **Is the timer the only evidence writer, or is every run a writer?** I
   verified the mechanism: `library-policy` records its own repo on every
   run unless `--no-record`, so a hand-built binary in that repo writes a
   `dev` row (10:36, 16:47). Keeping the default means mixed stamps are
   normal and I document it; changing it means non-fleet runs stop
   recording and the timer becomes the only writer. Which one?
3. **May I redeploy the LAN hub now?** It runs `b8b1dcf`, 26 commits behind.
   The new deploy gate never ran for real. I would fix the CSS artifact
   first (item 1), then deploy, which restarts the unit for about two
   seconds. Is the LAN sign-off (`docs/decisions/bind-posture.md`, "Open")
   still pending, and does deploying wait for it?

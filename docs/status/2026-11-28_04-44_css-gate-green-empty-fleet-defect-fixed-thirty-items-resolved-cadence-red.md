# CSS gate green, empty-fleet defect fixed, 30 report items resolved, cadence red (2026-11-28)

**Scope line.** This report describes go-compliance at HEAD `af268c6`
(`git rev-parse --short HEAD`, run 04:44 CET; tree clean; `github/master`
is two commits behind the local tip), templ-components at `bb6b3197` (clean;
no file changed there), library-policy at `ec304c5` (clean), BuildFlow at
`382949cb9` (one dirty path, another session), and md-go-validator at
`c6e8303` (clean). It covers the session of 2026-11-28 02:29 to 04:45 CET,
which resumed the list of
`docs/status/2026-11-27_17-11_todo-list-second-push-harness-green-two-red-ci-jobs.md`
after the owner's "get the whole TODO list done" and stopped at the owner's
order for this report. Against the tip that report described (`4271983`),
this session changed 27 files (+522, -314), all committed by the daemon.

**Verified in this session (command, result):**

- Browser harness `scripts/screenshot.sh` on the working tree (03:15, after
  the last dashboard change): `screenshot harness OK: 4 views` (light, dark,
  just-wide 768px, narrow). `scrolled dashboard-light`, `dashboard-dark`,
  `dashboard-just-wide`: `scrollTop=1400 headerPinOffset=0`.
  `keyboard: reached=true chipFocusVisible=true chipPressed=[Failing]
  spacePressed=[Stale] toggleFocusVisible=true themeToggled=true
  toggleAria=[false,true,false] toggleDark=[false,true,false]`.
  `console-errors.json`: `pageErrors: []`, `browserNoise: []`.
- The same harness on an empty fleet (an empty v1 document, discovery off:
  `SHOTS_EVIDENCE=<empty> SHOTS_DISCOVER_PATHS= ALLOW_EMPTY=1`):
  `screenshot harness OK: 4 views`, `rows=0 rowDisplay=table-row-group`,
  `filter All: visible=0 table=none messages=[No compliance evidence yet.
  The first refresh publishes it here.]`.
- Tests: core `GOWORK=off go test -race -count=1 .` ok and `go vet` rc=0 at
  04:37, after the last core edit. Hub `GOWORK=off go test -race -count=1
  ./...` ok for all 6 packages at 02:56; the dashboard package was re-run
  ok after each of its later changes (last 03:12); hub `go vet` rc=0 at
  04:37. The other five hub packages were not re-run after 02:56; none of
  their files changed after it.
- CI run 46103353820 on `eb9726f` (04:32 CET): 12 of 14 jobs pass, among
  them `dashboard CSS freshness gate`, the job that was red in the previous
  report. `docs gates` and `dprint + actionlint` fail (section d).
- Docs battery with the CI build (`docs-gates` is now a wrapper that runs
  `nix run github:Artmann-Technologies/docs-gates`, version `bfdefba`), at
  `6264995` (04:36): 9 PASS (freshness, audit-all, annotate, routing, stub,
  citations, stamps, vocabulary, agents), 1 FAIL (`cadence`: 16 unarchived
  reports in `docs/status/`, threshold 15).
  `docs-gates freshness . --audit-all --audit-max=0`: 31 reports replayed,
  0 unharvested open items (30 at 02:50).
- `dprint check` on the repository: rc=0 at 04:36, after `dprint fmt`
  reformatted 9 files. `actionlint .github/workflows/ci.yml` rc=0.
  `scripts/check-citations.sh`: 0 stale. `stack-drift-check.sh`: PASS.
- Tailwind toolchains: the rev CI used to hardcode (`0bb82723`) and the
  `flake.lock` rev (`2786a2a7`) both ship `tailwindcss v4.3.5`, and a
  regeneration with each leaves the three artifacts unchanged
  (`git diff --exit-code` rc=0 for both).
- Extension probe (a throwaway DevTools script against a fresh profile):
  16 to 19 extension service workers start without `--disable-extensions`,
  2 (Chromium's own) with it. Source: the host policy
  `/etc/chromium/policies/managed/extra.json` force-installs 22 extensions.
- Mutation test of the corrupt-document trip-wire, in a scratch copy of the
  core module under `/tmp` (04:01): unmutated, 1 spec passes; with
  `rememberMiss` made a no-op, the spec fails at `adapter_test.go:102`
  (`[parse] parse compliance document: jsontext: invalid character`, "to
  match error" `<io>`).
- The original flake artifact, read-only from BuildFlow's state DB:
  `step_outputs` row `20261120-105922-fc096d59`, step `test-race`, git head
  `80afd3a`, `command_output` 1855 bytes, middle truncated.
- BuildFlow, through a read-only subagent: `go build ./...` and `go vet
  ./...` rc=0 at `1620fc1a1` and again at `f1d847014`; the five tests named
  in the previous report pass at both commits; `CHANGELOG.md` line 9 carries
  the core bump row; `scripts/vendorhash-tripwire.sh` prints fresh.
- templ-components at `bb6b3197`, through a read-only subagent; logs in
  `/tmp/tc-lanes/`: `nix run .#visual` rc=1; `nix run .#verify` rc=1;
  `scripts/ci-repro.sh --lint --css --website` (inside `nix develop`) rc=1
  with no verdict line; per-module tests: `utils` rc=1, the five other
  modules rc=0; `nix run .#lint` rc=1 (details in section b).
- Live host at 04:40: `compliance-hub.service` and `fleet-evidence.timer`
  active; the unit's store path ends in `compliance-hub-b8b1dcf`, 45
  commits behind the tip; `/readyz` 200; timer next 05:10; evidence file
  210 rows (153 failing, 57 compliant).

**NOT verified here:**

- Plain `buildflow` and `nix flake check` at the current tip (CI's flake
  job passed on `eb9726f`).
- `scripts/deploy-hub.sh` against the real unit; the hub was not redeployed.
- CI on the local tip (`af268c6`, not pushed at 04:44): `dprint` and
  `audit-all` are green locally, not in CI.
- templ-components on GitHub: no CI result was read. Lanes a and b aborted
  at their first failing step, so their lint, coverage, art-dupl, erraudit,
  and website steps never ran through the script (the subagent ran the
  website lane and `nix run .#lint` on their own).
- The truth of the bare items in the 16 unarchived reports beyond the 30
  that `audit-all` flagged. The other bare items pass the gate by token
  match; nobody re-read them.
- Whether the stacked narrow layout is usable with a screen reader (its
  cells carry no column labels; not tested).
- The dark-mode colour of the pressed signal chip's focus ring in
  templ-components (light mode was measured; dark mode was not).

## Self-review

1. **What did you forget?** Six things. (a) I resolved 30 items and only
   then ran the whole battery: `cadence` is red because 16 reports sit
   unarchived, so the docs job stays red for a different reason. (b) Two
   routed verdicts named a `TODO_LIST.md` row and a digest row that did not
   exist until 04:35; the 11-52 report confessed the same mistake, and I
   repeated it for ten minutes before catching it. (c) The three done
   Medium `TODO_LIST.md` rows are still listed, `CHANGELOG.md`, `AGENTS.md`,
   `FEATURES.md` and `README.md` do not know about this session, and
   `buildflow`, `nix flake check` and the redeploy were not reached. (d) I
   ran a symbol replace on a block that was not a symbol; it replaced the
   `filterChecks` declaration, and the daemon committed the broken harness
   as `1bc119b` before I repaired it. (e) My annotations were committed
   unformatted, so CI's `dprint` job went red on `eb9726f`. (f) The temp
   dirs grew: `/tmp/cprobe`, `/tmp/app.min.css.run2`, `/tmp/tc-lanes`
   (206 MB, including a copy of the templ-components tree).
2. **What could you have done better?** Close one gate completely before
   opening the next. Run `dprint fmt` in the same minute as a markdown
   edit and `docs-gates run .` before and after an annotation sweep. Use
   exact-text edits for blocks. Write the row first and the route second.
   The scratch-copy mutation test is the pattern to keep.
3. **What could you still improve?** Everything in section f: the archive
   ritual, the redeploy, the templ-components fixes, the living docs, the
   TODO purge, the full gate run.
4. **Did you lie?** Not on purpose. Two beliefs from earlier reports are
   corrected with evidence: a locally installed docs-gates can stand in for
   the CI build (false: the system package `7a2d69b` reports 0 unharvested
   where the CI build reports 30), and the harness measures a clean profile
   (false on this host: policy-installed extensions ran inside it).
5. **Ghost systems?** Three closed, one open. Closed: the `.ml-1\.5` and
   `.opacity-70` rules that only test literals produced; the in-table
   "No compliance evidence yet" row, which the first patch of an empty fleet
   erased; the handwritten Tailwind rev in CI, five days behind the lock.
   Open: `layout.ThemeToggle` hardcodes `aria-checked="false"` and its own
   script corrects it; the harness now pins that it does.
6. **Split brains?** Three found, one open. The two nixpkgs revs (fixed:
   CI reads the lock). `STACK.md` against `go.mod` (fixed for the rows
   named in section a; the `samber/do` row still names a library the hub
   does not require directly). Three docs-gates binaries on `PATH` (one
   replaced by a wrapper; `~/.local/bin/docs-gates`, a `dev` build, and the
   system package remain).
7. **Tests?** Dashboard package: one spec added (theme colors), one
   rewritten (an empty fleet hides its table), one expectation updated (the
   chip message expression). Core: no test changed. Harness: a fourth view
   and two new check families (the layout each view must produce, the
   toggle's `aria-checked` trail), and filter checks that judge against the
   fleet the page rendered.
8. **Scope creep?** Deliberate, each part forced by a finding: the
   empty-fleet fix (found by the requested `ALLOW_EMPTY` run), the
   extension policy (the harness was red for a host reason), the lock-derived
   Tailwind rev (the two drift-check items asked for it), `STACK.md` and
   `CONTRIBUTING.md` (an archived item asked for the verification).

## Executive summary

The CSS gate is green in CI. The harness runs four views, is green on the
real fleet and on an empty fleet, and no longer fails on scripts that the
host's browser policy injects. The empty-fleet run found a real defect: the
first patch of an empty fleet erased the page's only explanation and left a
bare column header. It is fixed and pinned. All 30 items that failed the
docs audit are resolved inline, each with evidence. `STACK.md` and
`CONTRIBUTING.md` say what the code does.

The tip is still not shippable as a green CI run: `docs gates` is red on
`cadence` (16 unarchived reports, 17 with this one), the local tip is not
pushed, the live hub runs a build 45 commits behind, and templ-components'
master fails its own visual and test lanes. The living docs do not describe
this session.

Numbers: 27 files changed (+522, -314); 30 items resolved (23 done, 5
routed, 1 not done, 1 won't implement); 4 harness views; 1 defect fixed; 45
commits between the live hub and the tip; 16 unarchived reports.

## a) Fully done

| Item                                                                                                                                                                  | Evidence                                                                                                                                                                                                                      |
| --------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| The `dashboard CSS freshness gate` job is green: test files no longer feed the Tailwind scanner                                                                       | `@source not "**/*_test.go";` in `app.css`; three regenerations byte-identical; `.ml-1\.5` and `.opacity-70` gone from `app.min.css`; CI run 46103353820, job green                                                           |
| The harness ignores browser extensions: Chromium starts with `--disable-extensions`, and an exception from a `chrome-extension://` script counts as browser noise     | `scripts/screenshot.mjs`; probe: 16 to 19 extension workers without the flag, 2 with it; `console-errors.json` has no page error                                                                                              |
| A fourth harness view just above the stacking breakpoint (768px wide) pins the scroll box and the still page                                                          | `dashboard-just-wide`: `thead=table-header-group tableScrolls=true pageScrolls=false`, `headerPinOffset=0` after 1400px                                                                                                       |
| Layout checks key on the layout the page produced (the column header is hidden or not) and on each view's declared layout, not on a guess from the viewport width     | `scripts/screenshot.mjs` truth checks; new fact `tableOverflows`: the box must scroll exactly when the table is taller than the box's height limit                                                                            |
| The keyboard pass asserts that the theme toggle's `aria-checked` follows the theme                                                                                    | `toggleAria=[false,true,false]` equals `toggleDark=[false,true,false]` before, after Enter, after Space                                                                                                                       |
| Filter checks judge against the fleet the page rendered: an empty selection hides the table and is named, whatever the fleet holds                                    | `scripts/screenshot.mjs` `stateFilterFailure`; green on 210 rows and on an empty fleet; `SHOTS_DISCOVER_PATHS` in `scripts/screenshot.sh` switches discovery off                                                              |
| The `ALLOW_EMPTY=1` run against an empty evidence file was run, and it is green                                                                                       | `rows=0 rowDisplay=table-row-group` in all four views; one message on the page, whatever the chip                                                                                                                             |
| An empty fleet hides its table and explains itself outside the rows the stream replaces (the old message row was erased by the first patch)                           | `view.templ`: `tableShowExpression` starts with `$total > 0`, `noEvidenceShowExpression` block, `chipEmptyShowExpression`; spec "hides the table for a fleet without evidence and keeps the message outside the patched rows" |
| The page names its browser-UI colors: the hub accent for the light scheme, the dark canvas for the dark scheme                                                        | `view.go` `pageThemeColor` `#1464c6` (accent-600), `pageDarkThemeColor` `#030712`; spec "tints the browser UI with the hub palette, not the library's indigo"                                                                 |
| Mutation test of the corrupt-document trip-wire (16-58 report, item 3)                                                                                                | scratch copy, red at `adapter_test.go:102` with `rememberMiss` disabled, green unmutated                                                                                                                                      |
| The original flake artifact was traced (16-58 report, item 2: not recoverable)                                                                                        | BuildFlow `step_outputs` row `20261120-105922-fc096d59`: head-and-tail excerpt, middle truncated                                                                                                                              |
| `WithMissingBackoff` godoc states the zero value (16-58 report, item 36)                                                                                              | `adapter.go`                                                                                                                                                                                                                  |
| CI reads the Tailwind tool's nixpkgs rev from `flake.lock`; the handwritten rev had been behind the lock since 2026-11-23 (21-24 report item 7, 22-27 report item 29) | `.github/workflows/ci.yml` CSS gate job; both revs regenerate byte-identical artifacts; `actionlint` rc=0                                                                                                                     |
| `STACK.md` and `CONTRIBUTING.md` were verified against the code and corrected (archived 11-18 report, item 2)                                                         | `STACK.md`: stale replace and indirect notes gone, rows for `go-humanize` and `lipgloss/v2` + `colorprofile`, Logging row is stdlib `slog`, two decision-log entries; `CONTRIBUTING.md`: current commands                     |
| The 30 unharvested report items are resolved inline: 23 done with evidence, 5 routed, 1 not done, 1 won't implement                                                   | 25 through the annotate writer (`repair`), 5 by hand (four wrapped items, one archived item); `freshness . --audit-all --audit-max=0`: 0 unharvested                                                                          |
| `~/go/bin/docs-gates` is the CI build, not a stale `dev` binary                                                                                                       | wrapper script that runs `nix run github:Artmann-Technologies/docs-gates`; `docs-gates --version` prints `bfdefba`; the old binary is in the trash                                                                            |
| BuildFlow rows of the previous report (items 16 and 17): the changelog row exists, the vendor hash is fresh, the five tests pass                                      | subagent report: `CHANGELOG.md` line 9; `vendorhash-tripwire.sh` fresh; five tests PASS at `1620fc1a1` and `f1d847014`                                                                                                        |

## b) Partially done

| Item                                              | Works now                                                                                                                               | Still open                                                                                                                                                                                             | Blocker                   | Effort |
| ------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------- | ------ |
| Make CI green on master                           | CSS gate green on `eb9726f`; `audit-all` 0 and `dprint check` rc=0 locally                                                              | `cadence` red (16 unarchived reports); the local tip `af268c6` is not pushed, so CI has not seen the formatted annotations                                                                             | archive ritual            | M      |
| templ-components signal chip (T22)                | The three signal-chip browser tests pass; generated files in sync with the pinned generator; CSS artifact byte-fresh; counts consistent | Five route goldens not regenerated (from `5eaa2696`), two site goldens stale, four lint findings left by the "lint debt cleared" commit, the pressed chip's focus ring is white on white in light mode | owner answer (question 1) | M      |
| Redeployed hub against the tip                    | Unit active, readiness 200                                                                                                              | Unit runs `b8b1dcf`, 45 commits behind; the empty-fleet fix and the harness work are not live                                                                                                          | next step, not reached    | S      |
| `TODO_LIST.md` as the living backlog              | One Blocked row added (signal-chip adoption)                                                                                            | Three done Medium rows and two Low rows that are done or nearly done are still listed; this report's rows are not harvested                                                                            | purge not reached         | S      |
| The owner-decisions digest                        | One row added (go-idempotency Q2 and Q3)                                                                                                | Rows that the 2026-11-27 triage settled (branching-flow, inline vendor hashes) and the `go-humanize` row are stale                                                                                     | not reached               | S      |
| Items 7, 8, 25, 31 of the 2026-11-26 17-47 report | Status known: 31 done (BuildFlow warning triage in `AGENTS.md`), 25 partial (contributed, not adopted)                                  | Not annotated; 7 and 8 (full `buildflow` run, local `nix flake check`) are not done                                                                                                                    | follow the gate run       | S      |
| templ-components lanes on `bb6b3197`              | All four lanes were run and read; root and per-module tests pass except `utils`                                                         | Every lane is red at its first failing step; nothing was fixed (the repo was read-only for this session)                                                                                               | owner answer (question 1) | M      |

## c) Not started

| Item                                                                                                                                                                                          | Why not                                                                 | Priority |
| --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------- | -------- |
| Purge of the done `TODO_LIST.md` rows into the `CHANGELOG`                                                                                                                                    | not reached before the report order                                     | High     |
| `CHANGELOG.md`, `FEATURES.md`, `AGENTS.md`, `README.md`, `docs/notes/live-architecture.md` rows for this session                                                                              | same                                                                    | High     |
| Plain `buildflow` and `nix flake check` at the tip                                                                                                                                            | same                                                                    | High     |
| Redeploy through the real gate                                                                                                                                                                | same                                                                    | High     |
| Archive ritual for `docs/status/`                                                                                                                                                             | noticed only when the battery ran after the annotation sweep            | High     |
| templ-components fixes (goldens, lint, focus ring, playwright lockstep constant) and its three library TODO rows                                                                              | needs the owner's keep-or-revert answer                                 | High     |
| templ-components upstream changes (Table scroll-box knob, `ThemeToggle` icon contrast, `layout.Base` canvas fallback)                                                                         | not reached                                                             | Medium   |
| `docs/decisions/bind-posture.md` line 95 and the evidence-locations table in `docs/guides/integrate.md`                                                                                       | not reached                                                             | Low      |
| library-policy `scripts/fleet-snapshot.sh` regeneration; the row-count line in `scripts/deploy-fleet-evidence.sh`                                                                             | not reached                                                             | Low      |
| Discovery daemon: git worktrees that have a `go.mod`                                                                                                                                          | another repo; not reached                                               | Medium   |
| Trash `/tmp/tc-lanes`, `/tmp/cprobe`, `/tmp/compliance-screens`, `/tmp/app.min.css.run2`                                                                                                      | `/tmp/tc-lanes` still belongs to a running report; the rest not reached | Low      |
| The subagent checkpoint rule in `AGENTS.md`; the `samber/do` row in `STACK.md`                                                                                                                | not reached                                                             | Medium   |
| Producer single-writer decision; Go-only never-covered rule                                                                                                                                   | owner decisions                                                         | Medium   |
| Owner rows, unchanged: LAN-trust sign-off, SystemNix, `FLAKE_PAT`, Forgejo, benchmarks re-record (g2), M26, BuildFlow DB VACUUM, cache DB deletion, docs-gates fleet pin, cmdguard follow-ups | owner-only                                                              | Low      |

## d) Totally fucked up

1. **`docs gates` is still red, now on `cadence`.**
   - What: `docs-gates run .` fails one gate: 16 unarchived reports in
     `docs/status/`, threshold 15; this report makes it 17.
   - Severity: high. It is the gate the previous report called the High
     Impact row, and it stays red after 30 items were resolved.
   - Root cause: the 2026-11-27 17-11 report was the sixteenth; nobody ran
     the archive ritual. I did not run the whole battery before the sweep.
   - Mitigation: none. Section f, item 1.
2. **CI on `eb9726f` shows two red jobs that are green locally.**
   - What: `dprint + actionlint` failed on my unformatted annotations;
     `docs gates` counted 4 items that I annotated by hand after that
     commit.
   - Severity: medium. The tip that CI judged is not the tip that exists.
   - Root cause: markdown edits sat unformatted until the daemon committed
     them.
   - Mitigation: `dprint fmt` (9 files) and `dprint check` rc=0 at 04:36;
     `audit-all` 0. Unproven in CI: `af268c6` is not pushed.
3. **A broken harness was committed and pushed under me.**
   - What: commit `1bc119b` (02:55) holds `scripts/screenshot.mjs` with the
     filter checks in place of the `filterChecks` declaration.
   - Severity: medium. It does not build the product and CI does not run
     the harness, but the tip was wrong for four minutes and the daemon
     pushed it.
   - Root cause: I used a symbol replace on a block that was not a symbol.
   - Mitigation: repaired in `1c28b49`; every later harness change was
     checked with `node --check` and a real run.
4. **The empty-fleet page was broken for eight days, and no test saw it.**
   - What: with an empty fleet, the first SSE patch replaced the table body
     and erased the "No compliance evidence yet" row; the page showed a
     column header above nothing.
   - Severity: medium. Only an operator without a producer hits it, and
     that is exactly who needs the explanation.
   - Root cause: the message lived inside the region the stream replaces;
     the spec checked the server-rendered HTML only.
   - Mitigation: fixed and pinned (section a). The harness has the
     empty-fleet run. Not deployed.
5. **templ-components master fails its own lanes, and I cannot fix it
   without an answer.**
   - What: `nix run .#visual` fails on five route goldens that `5eaa2696`
     never regenerated and two site goldens that were already stale;
     `TestCrossbrowserLockstep` fails since `d5e25025` (2026-11-26: the
     spike pins playwright-core 1.65.1, the test wants 1.64.1);
     `nix run .#lint` reports four findings (two in
     `datastar/signal_chip.go`, one in `utils/golden/golden.go`, one in the
     visualtest e2e file); the pressed chip's focus ring is white on white
     in light mode (measured in Chromium by the subagent).
   - Severity: high for that repo. Two of the four findings came in with
     the hook-bypassed commits.
   - Root cause: commits landed on master without the lanes.
   - Mitigation: none. Question 1; section f, items 4 and 5.
6. **Pre-existing and open:** the live hub is 45 commits behind; the LAN
   sign-off is still "Open" in `docs/decisions/bind-posture.md`;
   `~/.local/bin/docs-gates` (a `dev` build) and the system package
   `7a2d69b` (which reports 0 unharvested for this corpus) are still on
   `PATH` behind the wrapper.

## e) What we should improve

| Problem                                                      | Impact                                                   | Fix                                                                                                       |
| ------------------------------------------------------------ | -------------------------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| The docs battery runs once, at the end                       | a sweep that closes one gate trips the next              | run `docs-gates run .` before and after every annotation sweep                                            |
| Markdown edits sit unformatted until the daemon commits them | CI's `dprint` job goes red on a correct change           | `dprint fmt` in the same minute as the edit; a format step before the daemon commits (owner's daemon)     |
| A symbol replace can hit the wrong declaration               | a broken file is committed and pushed within minutes     | exact-text edits for blocks; `node --check` after every harness edit                                      |
| The harness's Chromium inherits host policies                | third-party scripts run inside the measured page         | `--disable-extensions` plus the noise classification (done); write it into `AGENTS.md`                    |
| CI held a second copy of a rev that lives in `flake.lock`    | five days of silent drift                                | read the lock at run time (done); no other hardcoded nixpkgs rev was found                                |
| A routed verdict can name a row that does not exist          | a pointer into nothing, found only by the next reader    | write the row first, the route second (the rule existed; I broke it for ten minutes)                      |
| The empty-fleet page state had no test                       | a wiped message shipped                                  | the `ALLOW_EMPTY` harness run (done); run it before every deploy                                          |
| `stack-drift-check.sh` audits first-party requires only      | third-party dependencies drift unseen (three had no row) | a row-reading pass whenever `hub/go.mod` gains a direct require (recorded in the `STACK.md` decision log) |
| Commits bypass the hook on a shared library's master         | every consumer's next bump takes unverified code         | a branch and a pull request per templ-components' `AGENTS.md`; a stop-and-report rule for write subagents |

Two notes from the `modern-web-guidance` guides I read, neither applied:
the table box could show a scroll shadow through a `scroll-state` container
query (CSS only, no markup), and the stacked narrow layout injects no
column labels (its cell texts are self-describing: "23 findings", "12 hours
ago").

## f) Next things (up to 50, ranked by impact)

1. Run the archive ritual for `docs/status/`: 17 unarchived reports against a cadence threshold of 15; start with the report that has the fewest bare items (`2026-11-20_11-30`, 17), gate each file with the row checks before `git mv`, until `docs-gates run .` is all green. (Critical, M, Documentation)
2. Confirm CI on the pushed tip: the dprint job and the `audit-all` replay are green locally on `af268c6` and unproven in CI; read the next run's matrix. (Critical, S, CI)
3. Redeploy the hub through the real gate: `scripts/deploy-hub.sh`, then `scripts/screenshot.sh http://192.168.1.150:8477/`; the unit runs `b8b1dcf`, 45 commits behind. (High, S, Feature)
4. templ-components: decide whether `5eaa2696` and `bb6b3197` stay and are fixed forward or are reverted (question 1). (High, S, Decision; repo: templ-components)
5. templ-components, if the commits stay: regenerate the five route goldens and the two site goldens (`nix run .#visual -- -update`), clear the four lint findings (`signal_chip.go` cyclomatic complexity and a one-letter variable, `utils/golden/golden.go` magic number, the visualtest e2e `prealloc`), give the pressed chip's focus ring a visible colour, and move `expectedPlaywrightVersion` to 1.65.1. (High, M, Quality; repo: templ-components)
6. Write the living docs for this session: `CHANGELOG.md` rows, a `FEATURES.md` row for the scroll box and the empty-fleet message, `AGENTS.md` contracts (deploy gate and rollback, harness check families and the extension policy, test literals feed Tailwind, `startDiscovery`, the docs-gates wrapper, the empty-fleet run; stay under 220 lines), a `README.md` sentence. (High, S, Documentation)
7. Purge the done `TODO_LIST.md` rows with `docs-gates purge . --apply`: the three Medium rows (contrast, sticky header, keyboard pass) and the 768px harness row; the temp-dir row after item 23. (High, S, Documentation)
8. Run plain `buildflow` and `nix flake check` at the tip when no other full run is active; read the findings section. (High, S, Quality)
9. Harvest this report into `TODO_LIST.md` and `ROADMAP.md` (the rows this list names), then re-run the battery. (High, S, Documentation)
10. templ-components release tag with the signal chip, then the hub pin bump and the deletion of the hand-rolled chip markup (the Blocked row exists). (Medium, L, Feature; needs a tag and a push)
11. templ-components TODO rows for the three library findings: the Table scroll-box knob, the `ThemeToggle` icon contrast, the empty `BodyClass` canvas. (Medium, S, Documentation; repo: templ-components)
12. Producer single-writer decision (question 3), then either document the many-writers model in `docs/guides/integrate.md` or change the default in library-policy. (Medium, S, Decision)

13. ~~Discovery daemon: report git worktrees that have a `go.mod` (`go-flightrecorder-row92` is missing from the snapshot), so the hub's 209 meets the producer's 210. (Medium, S, Bug; repo: project-discovery)~~ done — resolved 2026-11-28: no daemon bug; the daemon reports the worktree (257 projects over the daemon socket, `go-flightrecorder-row92` present) and its eligible set equals the evidence rows (210 = 210); the rule is pinned by a regression test in the project-discovery-daemon repo (worktree_discovery_test.go, commit 2cf07dd). Details: docs/status/2026-11-28_13-55_reboot-wiped-scratch-lanes-died-on-usage-limit-repair-tools-lost.md, section a.

14. Annotate items 7, 8, 25, 31 of the 2026-11-26 17-47 report: 31 done, 25 partial, 7 and 8 after item 8 of this list. (Medium, S, Documentation)
15. Update `docs/notes/live-architecture.md`: the debounce window, the scroll box and its 254px calibration, the empty-fleet block outside the rows region, the harness check families. (Medium, S, Documentation)
16. Owner-decisions digest upkeep: strike the rows the 2026-11-27 triage settled (branching-flow, inline vendor hashes), reword the `go-humanize` row (adopted; the `STACK.md` row exists), and sharpen the `samber/do` row in `STACK.md` (transitive only). (Medium, S, Documentation)
17. BuildFlow: close the two rows of the 2026-11-27 17-11 report (items 16 and 17) and item 21 of the 11-52 report as done, with the subagent's evidence. (Medium, S, Documentation)
18. BuildFlow: two comments still say "(tagged v0.1.0)" (`internal/cli/compliance_view.go:15`, `modules/dbstore/compliance_test.go:189`) and its `AGENTS.md:51` records only the v0.1.0 bind. (Low, S, Documentation; repo: BuildFlow)
19. templ-components upstream changes so the hub can drop its overrides: a Table scroll-box knob, the `ThemeToggle` icon contrast, a `layout.Base` canvas fallback. (Medium, M, Feature; repo: templ-components)
20. Record the subagent checkpoint rule in `AGENTS.md`: a write subagent stops at the first red gate and reports; the coordinator checks it within ten minutes. (Medium, S, Process)
21. Confirm the Go-only never-covered rule (47 non-Go projects are excluded by design). (Medium, S, Decision)
22. `docs/decisions/bind-posture.md` line 95: describe the harness by its real check list. (Low, S, Documentation)
23. Move the evidence-file locations table into `docs/guides/integrate.md` as its single home; link from `README.md` and `AGENTS.md`. (Low, S, Documentation)
24. Regenerate library-policy `docs/reports/fleet-compliance.md` via `scripts/fleet-snapshot.sh` from the 210-row evidence. (Low, S, Documentation; repo: library-policy)
25. `scripts/deploy-fleet-evidence.sh`: after a scan, print the recorded row count next to the discovery daemon's Go project count. (Low, M, Infrastructure)
26. Trash `/tmp/tc-lanes` (206 MB), `/tmp/cprobe`, `/tmp/compliance-screens`, `/tmp/app.min.css.run2`, and ask the owner about `~/.local/bin/docs-gates` and the home-manager pin (question 2). (Low, S, Cleanup)
27. The archived 2026-11-19 15-50 report carries a WARN: its bullet "the lint-module replace-directive upstream fix (BLOCKED row)" reads as done and is not an item the gate can count. (Low, S, Documentation)
28. Table box: a scroll shadow through a `scroll-state` container query, so a fleet whose last row is not cut off still shows that the box scrolls. (Low, S, Feature)
29. The editor's TypeScript diagnostics for `scripts/screenshot.mjs` report three phantom errors beyond the end of the file (711 lines, valid for `node --check`); find out why or exclude the file. (Low, S, Tooling)
30. Owner rows, unchanged: LAN-trust sign-off, SystemNix for the two user units, `FLAKE_PAT` secrets, Forgejo visibility, benchmarks re-record (g2), M26 overview endgame, BuildFlow DB VACUUM, cache DB deletion, docs-gates fleet pin, cmdguard follow-ups. (Low, S, Decision)

## g) Questions I cannot answer myself

1. **templ-components: do the two commits stay?** `5eaa2696` and `bb6b3197`
   sit on `origin/master`, made with the pre-commit hook bypassed. The
   signal chip itself passes its browser tests; what is red is fixable in
   about an hour (seven goldens, four lint findings, one focus-ring colour,
   one version constant that was already wrong before them). I recommend
   keeping them and fixing forward. Reverting throws away passing work and
   does not fix the constant. And who cuts the release tag the hub needs
   before it can adopt the chip? I must not tag or push.
2. **What should happen to the other two docs-gates binaries?**
   `~/.local/bin/docs-gates` is a `dev` build from 2026-11-23 and the system
   package is `7a2d69b`, which reports 0 unharvested items for a corpus
   where the CI build reports 30. My wrapper in `~/go/bin` wins on this
   account's `PATH`; it is not a fix for other accounts or for a shell that
   omits `~/go/bin`. May I trash the `~/.local/bin` copy, and should the
   home-manager pin move to the CI build's revision (that is a
   crush-config change, not mine to make)?
3. **Is every library-policy run an evidence writer, or only the timer?**
   Every run records its own repo unless `--no-record`, so a hand-built
   binary writes a `dev` version stamp next to the timer's commit stamps
   (the row returned at 16:47 yesterday; the timer replaces it at 05:10).
   Keeping the default means mixed stamps are normal and I document it;
   changing it means only the timer writes. Which one?

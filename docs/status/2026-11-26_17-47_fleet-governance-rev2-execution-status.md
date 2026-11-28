# Fleet Governance Rev 2 Execution Status (2026-11-26 17:47)

**Scope line.** This report describes go-compliance HEAD `8d4153e` and
library-policy HEAD `fab6d1b` (`git rev-parse --short HEAD`, run at 17:47 CET
on 2026-11-26). The auto-commit daemon moves HEAD, so re-pin before relying on
any claim. The session behind it started on 2026-11-25 at 01:35 CET and was
suspended several times; wall-clock gaps of 3 to 26 hours sit between some
steps.

**Verified here** (2026-11-26 17:40 to 17:55 CET unless another time is
named):

- `date`: `2026-11-26 17:47:08 CET`. `who -b`: system boot 2026-11-25 08:58.
- go-compliance: `git status --short` empty. `git log 8473626..8d4153e`: one
  commit, `go.work.sum` plus one line.
- Hub unit: `systemctl --user is-active compliance-hub.service` prints
  `active`. The installed unit names `compliance-hub-bbc7913/bin`;
  `ActiveEnterTimestamp` is 2026-11-25 11:07:01 and `NRestarts` is 0.
- Live page: `scripts/screenshot.sh http://192.168.1.150:8477/` exited 0.
  It rendered 209 rows (154 failing, 55 compliant) in light, dark, and a
  390px viewport; clicking Stale showed 0 rows and "No stale repos"; Failing
  showed 154; All showed 209. No console errors, no failed requests.
- `/readyz` (read 2026-11-25 11:04): status `warn`, and only the
  `discovery` check warns ("silent too long: last contact 2 hours ago").
  `evidence`, `evidence-watch`, `refresh-freshness`, and
  `sse-backpressure` pass.
- `journalctl --user -u compliance-hub -b`: every start logs
  `discovery snapshot failed ... 400 Bad Request` (03:06 and 11:07 on
  2026-11-25), and the discovery check flips from pass to warn exactly ten
  minutes after each start.
- CI: run 45541531352 at `8473626` is 13 of 14 green, `report citation
  gate` red. Run 45669873581 at `8d4153e` is red on `report citation gate`
  and `docs gates`.
- `scripts/check-citations.sh` at HEAD: `checked=426 ok=425 stale=1`. The
  stale cite is `AGENTS.md:165-166` in
  `docs/review/libs/go-credentials-store.md`; `AGENTS.md` has 161 lines now.
- CI `docs gates` (run 45669873581) fails its `freshness` and `audit-all`
  gates on UNHARVESTED items of the 2026-11-25 00-01 report (items 8 and 10
  among them) and of the 2026-11-18 23-09 report (item 21, "Tag v0.1.0
  now?").
- Local corpus gates: `docs-gates run .` PASS 10 of 10 (local pin);
  `dprint check` rc 0.
- Evidence file `~/.local/state/overview/compliance.json`: mtime
  2026-11-26 17:24:57, 209 repos, every row `buildflow_versions: ["dev"]`,
  and a repo name starting with a dot is still present.
  `~/projects/.bf-jscpd-worktree` exists on disk.
- The `fleet-evidence` timer is not installed: no unit files in
  `~/.config/systemd/user/`, and `~/.local/state/compliance-fleet` does not
  exist.
- library-policy: `discoverFleetRepos` now lives in
  `cmd/library-policy/internal/cli/fleet_record.go:146-168` and skips
  dot-directories at line 155. Another actor's refactor moved it there
  (commits `7e5015a`, `5e1b66d`). My two tests are still in the tree. The
  working tree is clean.
- Discovery daemon (probed 2026-11-25 03:15 with the hub's own client):
  roots `/home/lars/projects` answer 251 projects, `$HOME` answers 1, and no
  roots answer 400. The daemon runs with
  `PROJECT_DISCOVERY_SEARCH_PATHS=/home/lars/projects`. Its
  `validateDiscoverOptions` (`server.go:1220`) rejects zero search paths
  with `errSearchPathsRequired`.
- Earlier in the session (2026-11-25 01:35 to 03:16): hub `-race` suite ok
  (02:58); hub `golangci-lint` 0 issues (03:02); core `-race` suite ok
  (01:54); dashboard package 44 specs ok; `buildflow -s nix-hash-fix --fix`
  verified 3 of 3 targets; `buildflow -s go-structure-linter` reports no
  `AGENTS.md` finding; `scripts/deploy-hub.sh` exited 0 at `1240a9f`,
  `9318c63`, and `bbc7913`.

**NOT verified here:**

- The canonical full `buildflow` gate was not run after any change of this
  session. Only the single-step drill-downs named above ran.
- `nix flake check` did not run locally after the hub `vendorHash` changed.
  CI skips that job without `FLAKE_PAT` and still reports success.
- Core `golangci-lint` was not re-run. Core is untouched except
  `go.work.sum`.
- library-policy: I last ran the `cli` package tests and lint at 02:14 on
  2026-11-25. Another actor refactored the file afterwards. I did not re-run
  them on `fab6d1b`.
- The contrast of the re-skinned colors is estimated from OKLCH lightness,
  not measured with a tool.
- Nobody looked at the page between 2026-11-25 03:06 and 2026-11-26 17:50.
- Unknown actors: who restarted the unit at 11:07 on 2026-11-25, who
  regenerated the evidence at 17:24 on 2026-11-26 with a `dev` binary, and
  who added the `go.work.sum` line in `8d4153e`.

## Executive summary

1. ~~T he h ub ru ns a s a syste md us er un it on `192.168.1.150:8477 `, bui lt from
   the committed tree, and survived a reboot plus one restart (30 hours of
   uptime at report time).~~ done — standing fact, re-verified by the later reports: the unit survived the 2026-11-28 10:46 boot too and runs a current rev (13-55 report scope line)

2. ~~T he dashboa rd w as rework ed e nd to e nd a nd is liv e: o ne headli ne wi th the
   failing ratio, a state bar, filter chips that carry counts, a quiet
   theme, a stacked layout for narrow screens, and empty states.~~ done — standing fact: the reworked page is live and the living docs describe it now (§f11 and §f12 at `77c2c45`)

3. ~~T he rework 's real-brows er harne ss fou nd th at t he fle et filt er h ad never
   worked: the chips used `data-on-click`, which the pinned DataStar runtime
   drops without a log line. That and eight other defects are fixed.~~ done — standing fact: the harness is the standing visual gate (§e row 1; the pre-deploy habit is in AGENTS.md)

4. ~~T he produc er ha lf of t he flywhe el is n ot installe d: no durable
   library-policy build, no timer. The evidence was last written by hand
   with a `dev` binary, and the `.bf-jscpd-worktree` row is still in it.~~ done — resolved 2026-11-27: the timer is installed and runs, and the prune keeps the worktree row out (§f2, §f3; closeout §a); the hand-written dev row is TODO_LIST Blocked (owner decisions) row "Producer single-writer decision"

5. ~~HE AD is r ed in CI on t wo gate s: o ne citati on drift ed wh en `AGENTS.md`
   shrank, and the CI docs-gates version finds unharvested report items.~~ done — the citation gate went green at `77c2c45` (§f1) and the unharvested items were harvested (§f6); the remaining local-vs-CI docs-gates split is TODO_LIST High Impact row "Make CI green on master"

6. ~~T he never-cover ed cou nt h as be en de ad sin ce t he featu re shippe d: t he hub
   asks the discovery daemon for a snapshot without search roots, and the
   daemon answers 400 on every start.~~ done — fixed at `77c2c45` (§f4): the serve flag sends real roots and the stub-daemon specs pin the contract; the journal shows the watch connecting without the 400 (closeout §a)

## Self-review

**What did you forget?**

- I did not run `scripts/check-citations.sh` after moving cited code, which
  `AGENTS.md` demands. CI has been red on that gate since 2026-11-25 03:39.
- I polished the dashboard for two hours before building the producer timer,
  although rev 2 ranks both in the same tier and the timer decides whether
  the data is fresh.
- I did not harvest this work into `TODO_LIST.md`, `FEATURES.md`, or
  `CHANGELOG.md`. The living docs still describe stat cards and a pending
  templ-components bump.
- I did not run the canonical `buildflow` gate once. I ran drill-downs and
  called the suites green.
- I left probe files in `/tmp` (`/tmp/compliance-status-shots`,
  `/tmp/compliance-status-live`), after planning to keep scratch out of it.
- I wrote `StartLimitIntervalSec` into `[Service]` in my own unit template
  and shipped it once before `systemd-analyze` caught it.
- I ran `templ generate` from the repo root by accident, the exact footgun
  `AGENTS.md` warns about, and had to regenerate from `hub/` to repair the
  `FileName` paths.

**What is something stupid that we do anyway?**

- The dashboard's reactive behavior had no real-browser test for a week, so
  an attribute typo shipped a dead filter. String tests cannot see a runtime
  that ignores an attribute.
- The hub asked the daemon for a snapshot with no roots and logged a 400 on
  every start for a week. The log line was read as noise.
- A healthy hub reports `warn` on `/readyz` forever after ten minutes,
  because a quiet discovery stream never refreshes the contact stamp.
- The auto-commit daemon bundles every actor's files into one heuristic
  commit and lagged up to an hour in library-policy under IO pressure.
- Two plans and at least three sessions touched the same two repositories
  within one day, and nothing enforces a single writer.

**What could you have done better?**

- Order by value: timer and producer first, then the unit, then visuals.
- Run the canonical gate after each task, as the plan's own guardrail says.
- Measure contrast with a tool instead of estimating it.
- Check `git log` and running processes after every step, not only at the
  start. The library-policy refactor moved my fix between two of my checks.

**What could you still improve?**

- A real contract test for the discovery request, against a stub daemon
  that rejects empty roots the way the real one does.
- A health semantic that treats a connected stream as contact.
- A producer exit code that lets the timer notice a failed scan.
- Pruning for the file store, so a repo that leaves the fleet disappears.

**Did you lie?** No. Two statements need correcting. I wrote in the previous
report that T08's padding step was "already in place". It was not: the
library's compact padding reached the header only, because the hub renders
its own cells. And I said the templ-components bump needed a vendorHash
re-bootstrap; it did, and `buildflow -s nix-hash-fix --fix` did it, but
`nix flake check` was never run to prove the hash.

**Ghost systems found:**

- The fleet filter: five chips rendered, none ever reacted (fixed).
- Never-covered: wired end to end, dead at the first request (open).
- The `stateSignals` patch: sent on every state change, read by nobody.
  The summary is patched as HTML instead.
- `--source buildflow`, the documented default source: this host runs
  `--source file`; no process was found that pipes a BuildFlow document.
- The v2 reporter ingress: history and severities have no path from an
  external reporter into the hub (owner-gated).

**Split brains found:**

- README says `nix run .#compliance-hub -- serve`; the runbook and
  `AGENTS.md` now say `scripts/deploy-hub.sh`. The README is still stale.
- `FEATURES.md` says "stat cards"; the page has a headline and chips.
- The evidence default path: core's `/var/lib/overview/compliance.json`
  versus the XDG path used live (open, T19).
- Local `docs-gates` (older pin) passes where CI's `docs-gates` (latest)
  fails. They do not judge the same rules.

**Tests:** 44 specs in the dashboard package, 20 of them new; the library
has two new tests for dot-directory skipping; the browser harness checks
themes, the favicon, filter clicks, and the empty selection on every run.
Missing: a discovery contract test, a measured contrast check, and a
keyboard-focus check.

**Scope creep:** yes, and all of it came from defects the harness or the
deploy exposed: the favicon route, the duplicated theme script, the live
region, the version stamp, the CI templ gate, the CI checkout depth, the
reference unit, a displaced `nolint` in library-policy. None of it is in the
CHANGELOG yet.

**Removed anything useful?** The four stat cards. The headline and the
count-carrying chips carry the same numbers, and on a phone the table moved
from 1390px to 615px down the page. `AGENTS.md` lost 82 lines of story; the
story moved to `docs/notes/live-architecture.md`.

**Inherited claims re-verified:** the rev-2 plan's T07, T02, T10, and T12
verdicts held. T08 did not (see above). The library's claim that a single
`SDKScript` tag makes DataStar work held; my first reading of the bundle
(auto-init missing) was wrong and I corrected it with a listener trace.

## a) Fully done

| Item                                                                                                                                                                                                                                                  | Evidence                                                                                                                             |
| ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| ~~Containment of `b0a8b7a`: root `go.mod`, `go.sum`, `go.work.sum` restored to `6e16257`; templ regenerated with the pinned v0.3.1100 from `hub/`~~ done — commit `2b4ada2`; the restore held (the closeout re-verified the inherited rev-2 verdicts) | ~~commit `2b4ada2`; `GOWORK=off go mod tidy -diff` empty in both modules at 01:39~~                                                  |
| ~~T07 findings cell: count fallback, pluralized through `go-humanize` (BuildFlow `go-humanize-linter` error)~~ done — dashboard specs green; live renders verified at report time ("85 findings", "2 findings")                                       | ~~dashboard specs green; live: "85 findings", "2 findings"~~                                                                         |
| ~~T05 templ-components v1.27.0 to v1.27.2 (six modules), CSS regen, vendorHash `sha256-ioJ7Njpl...`~~ done — commits `9d7da9c`, `f2d81d2`; hub `-race` suite ok at 01:53                                                                              | ~~commits `9d7da9c`, `f2d81d2`; hub `-race` suite ok at 01:53~~                                                                      |
| ~~T10.1 screenshot harness: `scripts/screenshot.sh` and `scripts/screenshot.mjs`~~ done — both scripts at HEAD; exit 0 verified against the tree build and the LAN URL                                                                                | ~~exit 0 against the tree-built hub and against the live LAN URL~~                                                                   |
| ~~T01 hub as a systemd user unit: `scripts/deploy-hub.sh`, `deploy/compliance-hub.user.service`~~ done — deploy/compliance-hub.user.service at HEAD; the unit ran from 2026-11-25 11:07:01 across the 08:58 boot                                      | ~~unit active since 2026-11-25 11:07:01, survived the 08:58 boot; hardened (`ProtectSystem=strict`); watchdog pings in the journal~~ |
| ~~Hub version stamp: the flake passes the short rev to `cli.Version`~~ done — the stamp shipped; the unit was redeployed from newer revs since (b8b1dcf on 2026-11-27 and later)                                                                      | ~~`compliance-hub --version` prints `bbc7913`; `/version` agrees~~                                                                   |
| ~~T16 LAN recipe as `bind-posture.md` Appendix A, cross-linked from the runbook and `AGENTS.md`~~ done — commit `8473626`; the Appendix A recipe stands in docs/decisions/bind-posture.md                                                             | ~~commit `8473626`; `dprint check` rc 0~~                                                                                            |
| ~~T20 auth posture: threat sketch and recommendation, marked awaiting sign-off~~ done — the sketch stands in docs/decisions/bind-posture.md; the sign-off ask lives in TODO_LIST Blocked (owner decisions) row "LAN-trust sign-off"                   | ~~`docs/decisions/bind-posture.md`~~                                                                                                 |
| ~~T04 hierarchy: headline with the failing ratio, state bar, chips with counts~~ done — live and harness-checked at report time                                                                                                                       | ~~live, harness-checked; 27 rows above the fold instead of 24~~                                                                      |
| ~~T03 quiet theme: one accent, muted `red`/`amber`/`green` ramps, system font stack, in `app.css` `@theme` only~~ done — the single-accent theme shipped in the app.css @theme block                                                                  | ~~compiled CSS carries one `--color-accent-600` definition (mine)~~                                                                  |
| ~~T08 table: cell padding on custom rows, findings column right-aligned, count-first severity tokens~~ done — shipped; screenshots at 1440px and 390px                                                                                                | ~~screenshots at 1440px and 390px~~                                                                                                  |
| ~~T09 typography: `tabular-nums` on counts, 400/500/600 weights~~ done — shipped with the theme work                                                                                                                                                  | ~~screenshots~~                                                                                                                      |
| ~~T10 visual QA and T21 empty states: "No stale repos" for an empty selection, stacked rows under 40rem, `LiveOff` on the table region~~ done — the empty-selection and stacked-row behavior is harness-verified                                      | ~~harness: `filter Stale: visible=0 table=none messages=[No stale repos]`~~                                                          |
| ~~Filter chips fixed: `data-on:click`, object-form `data-class`, live `aria-pressed`, server-rendered selection~~ done — the fleet filter works; the harness pin records the pressed state                                                            | ~~harness: `pressed=[Failing] highlighted=true`~~                                                                                    |
| ~~Favicon route (`/favicon.svg`, embedded, ETag/304)~~ done — shipped with the routes spec                                                                                                                                                            | ~~no 404 in the harness; `routes_test.go` spec~~                                                                                     |
| ~~Duplicate pre-paint theme script removed~~ done — a spec counts exactly one theme read                                                                                                                                                              | ~~spec counts one `localStorage.getItem('theme')`~~                                                                                  |
| ~~`?state=` pages: headline and chips count the whole fleet~~ done — the routes spec covers the whole-fleet counts                                                                                                                                    | ~~`routes_test.go` spec~~                                                                                                            |
| ~~T06 dot-directory skip in library-policy, table test plus BDD spec~~ done — the library-policy repo skips dot-directories in its fleet walk (line 155 at the time); table test plus BDD spec                                                        | ~~`fleet_record.go:155` at `fab6d1b`; proved failing without the fix~~                                                               |
| ~~library-policy: displaced `//nolint:gosec` and magic `0o750` repaired~~ done — library-policy golangci-lint 0 issues on the cli package at 02:14 on 2026-11-25                                                                                      | ~~`golangci-lint` 0 issues on the `cli` package at 02:14 on 2026-11-25~~                                                             |
| ~~CI templ gate regenerates and diffs; CI docs-gates checks out full history~~ done — .github/workflows/ci.yml carries both jobs; actionlint rc 0                                                                                                     | ~~`.github/workflows/ci.yml`; `actionlint` rc 0; both jobs green in run 45541531352 except the citation gate~~                       |
| ~~Reference unit reconciled: `StartLimitIntervalSec` moved to `[Unit]`, false NixOS-module claim removed, ineffective `Environment=` line removed~~ done — deploy/compliance-hub.service reconciled                                                   | ~~`deploy/compliance-hub.service`~~                                                                                                  |
| ~~`AGENTS.md` slimmed from 243 to 161 lines; story moved to `docs/notes/live-architecture.md`~~ done — the story lives at docs/notes/live-architecture.md; later sessions grew AGENTS.md again (219 lines on 2026-11-28)                              | ~~`buildflow -s go-structure-linter`: no `AGENTS.md` finding~~                                                                       |
| ~~Rev-1 plan annotated; rev 2 written (carried over from the previous report)~~ done — commits `526ffa3`, `6e16257`; the rev 2 plan docs/planning/2026-11-24_23-15_fleet-governance-80-20-execution-plan-rev2.md is annotated too                     | ~~commits `526ffa3`, `6e16257`~~                                                                                                     |

## b) Partially done

| Item                                                                                                                                                                                                                                               | Works now                                                                                                                                                 | Still open                                                                    | Blocker                                                  | Effort |
| -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------- | -------------------------------------------------------- | ------ |
| ~~T07 close-out~~ done — CHANGELOG and FEATURES rows landed at `77c2c45` (§f9, §f11); the plain buildflow run exited 0 on 2026-11-27                                                                                                               | ~~Code, tests, deployment~~                                                                                                                               | ~~CHANGELOG, FEATURES, and TODO rows; canonical gate run~~                    | ~~None~~                                                 | ~~S~~  |
| ~~T01 redeploy~~ done — redeployed from the discovery-fix commit (§f10 at `77c2c45`); later deploys `b8b1dcf` and newer                                                                                                                            | ~~Unit live at `bbc7913`~~                                                                                                                                | ~~Hub sources equal HEAD; a redeploy is needed only after the discovery fix~~ | ~~None~~                                                 | ~~S~~  |
| ~~T02 fleet evidence timer~~ done — installed, started, and run (§f2 at `77c2c45`); active with the next run 2026-11-28 (closeout §a); the hand-writer question moved to TODO_LIST Blocked (owner decisions) row "Producer single-writer decision" | ~~`deploy/fleet-evidence.service`, `deploy/fleet-evidence.timer`, `scripts/deploy-fleet-evidence.sh` written, shellcheck clean, `systemd-analyze` clean~~ | ~~Never run: no build, no install, no scan~~                                  | ~~Question 3 (another session writes evidence by hand)~~ | ~~M~~  |
| ~~T13 durable library-policy build~~ done — the timer's nix build runs library-policy at its committed HEAD (closeout §a)                                                                                                                          | ~~Script builds `git+file://<repo>?rev=HEAD#library-policy`~~                                                                                             | ~~Never executed; `~/.local/state/compliance-fleet` absent~~                  | ~~Same as T02~~                                          | ~~S~~  |
| ~~T14 producer version stamp~~ done — the evidence rows carry the nix build's commit stamp (210 of 211 at the closeout; the one dev row is the single-writer TODO)                                                                                 | ~~Nothing~~                                                                                                                                               | ~~Verify `buildflow_versions` after a nix-built scan~~                        | ~~Depends on T13~~                                       | ~~S~~  |
| ~~T11 discovery 400~~ done — the serve flag shipped with the stub-daemon regression specs (§f4 at `77c2c45`); the journal shows the watch connecting (closeout §a)                                                                                 | ~~Root cause found and reproduced; fix designed (`--discover-paths`, default `~/projects`)~~                                                              | ~~Flag, plumbing, regression test, redeploy~~                                 | ~~Question 2 (fleet definition)~~                        | ~~S~~  |
| ~~T19 evidence path docs~~ done — README, docs/guides/integrate.md, and the bind-posture decision state both defaults (§f20 at `77c2c45`)                                                                                                          | ~~`AGENTS.md` names both defaults~~                                                                                                                       | ~~Decision docs and README not aligned~~                                      | ~~None~~                                                 | ~~S~~  |
| ~~T26 hygiene~~ done — the /tmp probe dirs went with the reboot (§f27); the TODO strike-and-purge landed (§f26 at `77c2c45`); CHANGELOG rows in (§f9)                                                                                              | ~~Scratch from the first hours went with the reboot~~                                                                                                     | ~~New probe dirs in `/tmp`; TODO strike and purge; CHANGELOG rows~~           | ~~None~~                                                 | ~~M~~  |
| ~~library-policy gates~~ done — cli tests and golangci-lint green at core v0.2.0 after the refactor (§f33 at `77c2c45`)                                                                                                                            | ~~`cli` package green at 02:14 on 2026-11-25~~                                                                                                            | ~~Re-run tests and lint on `fab6d1b` after another actor's refactor~~         | ~~None~~                                                 | ~~S~~  |

## c) Not started

| Task                                                                                                                                                                                                            | Subject                                                                         | Why not                                                                           | Priority   |
| --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------- | --------------------------------------------------------------------------------- | ---------- |
| ~~T18~~ done — §f13 at `77c2c45`: library-policy, md-go-validator, and BuildFlow require core v0.2.0                                                                                                            | ~~Consumer sweep to core v0.2.0 in library-policy, md-go-validator, BuildFlow~~ | ~~Deferred behind the flywheel; other sessions active in two of the three repos~~ | ~~Medium~~ |
| ~~T15~~ done — §f19 at `77c2c45`; the library-policy repo has scripts/fleet-snapshot.sh writing docs/reports/fleet-compliance.md                                                                                | ~~Fleet snapshot persisted to library-policy `docs/reports/`~~                  | ~~Depends on a nix-built producer run~~                                           | ~~Low~~    |
| ~~T22~~ routed — TODO_LIST Blocked (owner decisions) row "Adopt the templ-components signal chip once it is tagged" (the chip sits untagged on that repo's master; the hub still renders its hand-rolled chips) | ~~DataStar signal-mode chip contributed to templ-components~~                   | ~~Rest tier~~                                                                     | ~~Low~~    |
| ~~T28~~ done — §f21 at `77c2c45`: docs/decisions/v2-wire-design.md carries the "Decision ask" section                                                                                                           | ~~v2 reporter ingress decision ask~~                                            | ~~Rest tier~~                                                                     | ~~Medium~~ |
| ~~T27~~ done — §f40 at `77c2c45`: the consolidated ask is the closing section of docs/status/2026-11-27_17-11_todo-list-second-push-harness-green-two-red-ci-jobs.md                                            | ~~Consolidated owner ask~~                                                      | ~~This report's section g is the first draft~~                                    | ~~Medium~~ |
| ~~README~~ done — §f12 at `77c2c45`: the user-unit recipe replaced the nix run recipe and the screenshot was refreshed                                                                                          | ~~Replace the `nix run` recipe; refresh `docs/assets/dashboard.png`~~           | ~~Not reached~~                                                                   | ~~Medium~~ |
| ~~CSP~~ done — §f15 at `77c2c45`: Content-Security-Policy and Referrer-Policy on every hub response                                                                                                             | ~~The hub sets only `X-Content-Type-Options`~~                                  | ~~Found during the threat sketch; not in the plan~~                               | ~~Medium~~ |

## d) Totally fucked up

1. ~~**CI is red at HEAD on two gates.**~~ done — resolved 2026-11-27: the cite re-pointed and the citation gate went green (§f1), the unharvested items were harvested (§f6), and this corpus sweep strikes the remainder; the binary-version split is TODO_LIST High Impact row "Make CI green on master"
   - What: `report citation gate` fails on one stale cite
     (`docs/review/libs/go-credentials-store.md` cites `AGENTS.md:165-166`,
     a file of 161 lines). `docs gates` fails in CI on `freshness` and
     `audit-all` (unharvested items of two older reports), while the local
     `docs-gates` pin passes.
   - Severity: medium. Every "green on master" claim is false until fixed.
   - Root cause: I shrank `AGENTS.md` without re-pointing the cite. The
     docs-gates failure is a tool-version split: local is older than CI.
   - Mitigation: none applied. Both are small edits.

2. ~~**Never-covered has never worked.**~~ done — fixed at `77c2c45` (§f4): the serve flag passes real roots and the stub-daemon regression specs pin the contract (closeout §a)
   - What: the hub calls `Discover` with no search roots; the daemon
     rejects that with 400 on every start, so the never-covered set is
     always empty and the dashboard shows 0 by construction.
   - Severity: high. A headline state of the domain model is dead.
   - Root cause: `hub/internal/cli/refresher.go` passes `nil`; the daemon's
     `validateDiscoverOptions` requires at least one root.
   - Mitigation: none. The fix is a serve flag plus plumbing.

3. ~~**`/readyz` is `warn` forever after ten minutes.**~~ done — fixed at `77c2c45` (§f5): a connected stream passes the discovery check by itself
   - What: with a healthy, quiet discovery stream the contact stamp is
     never refreshed, and the check degrades exactly ten minutes after
     every start (journal 03:16:59 and 11:17:02 on 2026-11-25).
   - Severity: medium. A permanent warning hides a real one.
   - Root cause: `markDiscoverySuccess` runs on connect and on a successful
     snapshot only.
   - Mitigation: none.

4. ~~**The evidence pipeline has no durable producer.**~~ done — resolved 2026-11-27 (§f2, §f3): the timer is installed and running and the prune keeps the worktree row out; the unknown dev-binary writer is TODO_LIST Blocked (owner decisions) row "Producer single-writer decision"
   - What: no timer, no nix build. The evidence was regenerated on
     2026-11-26 at 17:24 by an unknown actor with a `dev` binary. The
     `.bf-jscpd-worktree` row is still in the file, because the file store
     merges and never prunes.
   - Severity: high. This is the "1% that delivers 51%" of rev 2.
   - Root cause: my ordering (visuals first), then the interruptions.
   - Mitigation: the script and units exist and are one run away.

5. ~~**Pre-existing and open:** 94 BuildFlow findings at warning and info~~ done — each part closed: the findings triaged (§f31), the go floor a deliberate NOT-DO (§f30), the pin split routed to TODO_LIST Low Impact row "Fleet docs-gates pin bump past `e305834`", README and FEATURES updated at `77c2c45` (§f11, §f12)
   level (branching-flow, erraudit, go-auto-upgrade, nix-checker);
   `hub/go.mod` declares `go 1.27.2`, a patch floor; the local docs-gates
   pin is behind CI; `README.md` and `FEATURES.md` describe the old page.

## e) What we should improve

| Problem                                                                                                                                                                                                                              | Impact                                                                   | Fix                                                                                                                                                |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| ~~A dashboard's reactive contract had no browser-level test~~ done — scripts/screenshot.sh is the standing visual gate (AGENTS.md pre-deploy habit); the keyboard extension is routed to TODO_LIST Medium Impact row "Keyboard pass" | ~~A dead filter shipped and nobody noticed for a week~~                  | ~~Keep `scripts/screenshot.sh` as the gate; add it to the pre-deploy habit in `AGENTS.md` (done) and extend it with a keyboard pass~~              |
| ~~Moving cited code has no local reminder~~ done — the citation gate is green since `77c2c45` (§f1) and CI keeps enforcing it                                                                                                        | ~~CI red for a day~~                                                     | ~~Make `scripts/check-citations.sh` part of the documented after-refactor checklist; add it to the harness `buildflow` runs if a provider exists~~ |
| ~~The discovery contact stamp ignores a live stream~~ done — a connected stream counts as contact (§f5 at `77c2c45`; closeout §a)                                                                                                    | ~~Permanent `warn`~~                                                     | ~~Stamp on stream events and treat a connected stream as contact~~                                                                                 |
| ~~`library-policy fleet` exits 0 when it scanned nothing or recorded nothing~~ done — §f17's verdict stands: the library already exits non-zero when the evidence is incomplete                                                      | ~~A timer cannot notice a broken scan~~                                  | ~~Non-zero exit on zero repos or on zero recorded rows~~                                                                                           |
| ~~The file store never prunes~~ done — §f18 at `77c2c45`: the library-policy prune mode, default on, with three tests                                                                                                                | ~~Stale rows outlive their repos~~                                       | ~~A prune or replace mode, or a documented regenerate ritual in the producer's help~~                                                              |
| ~~Two `docs-gates` versions judge the same corpus~~ routed — TODO_LIST Low Impact row "Fleet docs-gates pin bump past `e305834`" (the PATH binary now runs CI's build; the fleet pin bump stays the owner row)                       | ~~Local green, CI red~~                                                  | ~~Bump the fleet pin (existing BLOCKED TODO row)~~                                                                                                 |
| ~~Sessions collide in one tree~~ routed — TODO_LIST Blocked (owner decisions) row "Producer single-writer decision"                                                                                                                  | ~~A refactor moved my fix; an unknown hand-built binary wrote evidence~~ | ~~A single-writer rule per repo; the check is in `AGENTS.md` now, enforcement is not~~                                                             |
| ~~The deploy unit is a plain user file~~ routed — TODO_LIST Blocked (owner decisions) row "Move both user units" (the SystemNix reproduction row)                                                                                    | ~~Not reproduced by a system rebuild~~                                   | ~~Owner decision: move both units into SystemNix~~                                                                                                 |

## f) Next things (up to 50, ranked by impact)

~~1. Re-point the stale cite in `docs/review/libs/go-credentials-store.md` and rerun `scripts/check-citations.sh` (Critical, S, Bug).~~ done at `77c2c45` 2026-11-27: cite re-pointed to AGENTS.md:59-61; check-citations stale=0

~~2. Run `scripts/deploy-fleet-evidence.sh --run`: nix build of library-policy at its committed HEAD, install the timer, one scan (Critical, M, Feature).~~ done at `77c2c45` 2026-11-27: timer installed and started, one scan ran, 210 repos recorded

~~3. Remove the `.bf-jscpd-worktree` row: trash the evidence file right before the timed scan, because the file store never prunes (Critical, S, Bug).~~ done at `77c2c45` 2026-11-27: library-policy `fleet --prune` drops it on every run

~~4. Add `--discover-paths` to `serve` (default `~/projects`), pass it to `Discover`, and add a regression test against a stub daemon that rejects empty roots (High, S, Bug).~~ done at `77c2c45`: serve --discover-paths with stub-daemon regression specs

~~5. Fix the discovery health stamp so a connected stream counts as contact (High, S, Bug).~~ done at `77c2c45`: a connected stream passes the discovery check by itself

~~6. Harvest this report and the 2026-11-25 00-01 report into `TODO_LIST.md` until the CI `docs gates` job is green (High, M, Documentation).~~ done at `77c2c45` 2026-11-27: docs-gates run . is all green, audit-all clean

7. ~~Run the canonical full `buildflow` gate at HEAD and fix every error-severity finding (High, M, Quality).~~ superseded by docs/status/2026-11-28_13-55_reboot-wiped-scratch-lanes-died-on-usage-limit-repair-tools-lost.md — its §f17 owns the at-the-tip re-run; plain `buildflow` exit 0 on 2026-11-27 06:0x with zero error findings (closeout §b), warnings triaged at AGENTS.md "BuildFlow warning triage"

8. ~~Run `nix flake check` locally (High, S, Quality).~~ done — 2026-11-27 06:1x: all checks passed on the rev-2 tree (docs/status/2026-11-27_11-52_fleet-governance-rev2-closeout-status.md, scope line and the "Canonical gate at the tip" row); flake.nix moved afterwards — the standing at-the-tip re-run lives in the successor report's §f17

~~9. Write the CHANGELOG rows for every fix and feature of 2026-11-25 (High, S, Documentation).~~ done at `77c2c45` 2026-11-27: CHANGELOG Unreleased carries Added, Fixed, and Security rows

~~10. Redeploy the hub from the commit that carries items 4 and 5, then run the harness against the LAN URL (High, S, Feature).~~ done at `77c2c45`: unit runs 77c2c45, scripts/screenshot.sh green on the LAN URL

~~11. Update `FEATURES.md`: replace "stat cards" with the fleet summary and define `DOCUMENTED` in the legend (Medium, S, Documentation).~~ done at `77c2c45` 2026-11-27: fleet summary row and the DOCUMENTED legend entry

~~12. Update `README.md`: the user unit instead of `nix run`, and a current `docs/assets/dashboard.png` (Medium, S, Documentation).~~ done at `77c2c45` 2026-11-27: user unit recipe, discover-paths, refreshed docs/assets/dashboard.png

~~13. T18 consumer sweep to core v0.2.0 in library-policy, md-go-validator, BuildFlow (Medium, L, Dependency).~~ done at `77c2c45` 2026-11-27: library-policy, md-go-validator, and BuildFlow require core v0.2.0; tests match each baseline

~~14. Decide the never-covered fleet definition (question 2), then document it in `docs/DOMAIN_LANGUAGE.md` (Medium, S, Decision).~~ done at `77c2c45` 2026-11-27: fleet member = direct subdirectory with a go.mod; docs/DOMAIN_LANGUAGE.md and the README state it

~~15. Add a CSP header to the hub responses (Medium, S, Security).~~ done at `77c2c45`: Content-Security-Policy and Referrer-Policy on every response

16. ~~Annotate rev 2: T04 shipped without stat cards; T08's padding was absent on custom rows (Medium, S, Documentation).~~ done — the rev 2 plan docs/planning/2026-11-24_23-15_fleet-governance-80-20-execution-plan-rev2.md carries the correction strikes (43 strike markers on disk, 2026-11-28); the closeout §a logs the annotation pass

~~17. Give `library-policy fleet` a non-zero exit when it scans or records nothing (Medium, S, Bug).~~ NOT-DO/DUPLICATE — already shipped in library-policy on 2026-11-26: fleet exits non-zero when evidence is incomplete

~~18. Add a prune or replace mode to the file store, or document the regenerate ritual (Medium, M, Feature).~~ done at `77c2c45` 2026-11-27: library-policy `fleet --prune` (default on)

~~19. T15: persist the fleet snapshot into library-policy `docs/reports/`, whitespace-normalized (Low, S, Documentation).~~ done at `77c2c45` 2026-11-27: library-policy scripts/fleet-snapshot.sh writes docs/reports/fleet-compliance.md

~~20. T19: align the evidence path defaults across README and the decision docs (Low, S, Documentation).~~ done at `77c2c45` 2026-11-27: README, docs/guides/integrate.md, and bind-posture.md state both defaults

~~21. T28: write the v2 reporter-ingress decision ask (Medium, S, Decision).~~ done at `77c2c45` 2026-11-27: docs/decisions/v2-wire-design.md, section Decision ask

~~22. File a verified templ-components issue: `SDKScript` consumers get no signal that `data-on-click` is dropped (Medium, S, Upstream).~~ NOT-DO/DUPLICATE — superseded by the contribution itself: the signal chip component in templ-components bakes in data-on:click with a guard test (item 25)

~~23. File a verified BuildFlow issue: the prettier lane rewrites committed compiled artifacts (Medium, S, Upstream).~~ NOT-DO — deliberate non-fix recorded here: .buildflow.yml keeps the formatters off the generated CSS artifacts and the CI css-freshness gate guards byte-freshness; no issue filed

24. ~~Bump the local docs-gates pin past `e305834` (Medium, S, Tooling).~~ **→ routed to TODO_LIST Low Impact row "Fleet docs-gates pin bump past `e305834`"**

25. ~~T22: contribute the signal-mode chip to templ-components and adopt it in `chips.templ` (Low, L, Feature).~~ **→ routed to TODO_LIST Blocked (owner decisions) row "Adopt the templ-components signal chip once it is tagged"**

~~26. Strike and purge the done `TODO_LIST.md` rows, including the 4 unwrapped DONE rows (Medium, S, Cleanup).~~ done at `77c2c45` 2026-11-27: 15 rows struck and purged into the CHANGELOG

~~27. Trash the probe directories `/tmp/compliance-status-shots` and `/tmp/compliance-status-live` (Low, S, Cleanup).~~ done at `77c2c45` 2026-11-27: both were already gone after the reboot; the harness output dir was trashed too

28. ~~Measure the contrast of badge, chip, and severity text in both themes with a tool (Low, S, Quality).~~ **→ routed to TODO_LIST Medium Impact row "Fix the text contrast the harness measures under WCAG AA"**

29. ~~Add a keyboard-focus pass to the harness (chips, theme toggle, table region) (Low, S, Quality).~~ **→ routed to TODO_LIST Medium Impact row "Keyboard pass"**

~~30. Change `hub/go.mod` to `go 1.27` (BuildFlow `go-version-auto-configure` warning) (Low, S, Quality).~~ NOT-DO — a dependency declares go 1.27.2 and go mod tidy restores the patch floor on the next run; the directive is forced, not chosen

31. ~~Triage the 94 BuildFlow findings: fix or record a deliberate non-fix per tool (Medium, M, Quality).~~ done — triaged 2026-11-27: plain `buildflow` exits 0 and the warnings carry recorded deliberate non-fixes at AGENTS.md "BuildFlow warning triage" (AGENTS.md:207; closeout §a)

32. ~~Sticky header for the 200-row table (Low, M, Feature).~~ **→ routed to TODO_LIST Medium Impact row "Make the sticky fleet-table header real"**

~~33. Re-run the library-policy `cli` package tests and lint on its current HEAD (Medium, S, Quality).~~ done at `77c2c45` 2026-11-27: cli package tests and golangci-lint green at core v0.2.0

~~34. Record the discovery flag and the never-covered definition in `AGENTS.md` once landed (Low, S, Documentation).~~ done at `77c2c45` 2026-11-27: AGENTS.md Deployment section

~~35. Move both user units and the library-policy pin into SystemNix (Medium, M, Infrastructure).~~ → routed to TODO_LIST Blocked (owner decisions) SystemNix row

36. ~~Annotate the 2026-11-24 UI-audit report: "zero `@theme` overrides" and the stat-card verdict are history now (Low, S, Documentation).~~ done — docs/status/2026-11-24_22-27_fleet-lan-serve-and-dashboard-ui-audit.md carries the strikes (verified 2026-11-28: its numbered items are struck done)

~~37. Add a spec that `compliance-hub describe` still lists every serve flag after the discovery flag lands (Low, S, Quality).~~ NOT-DO/DUPLICATE — the serve flag contract is pinned by a struct-tag spec on ServeFlags.DiscoverPaths in refresher_discovery_internal_test.go

~~38. Add the citation check to the release runbook's pre-tag sequence (Low, S, Process).~~ done at `77c2c45` 2026-11-27: docs/notes/release-runbook.md step 4

~~39. Make the harness fail on `warn` readiness during a deploy probe, so a permanent warning cannot hide again (Low, S, Quality).~~ done at `77c2c45` 2026-11-27: scripts/deploy-hub.sh fails when any readiness check is not pass after 10s

~~40. T27: send the owner ask in one list (sign-off, fleet definition, SystemNix, `FLAKE_PAT`, Forgejo visibility, docs-gates pin) (Medium, S, Decision).~~ done at `77c2c45` 2026-11-27: the ask is the closing section of the 2026-11-27 execution status report

## g) Questions I cannot answer myself

1. Is another session owning the producer side now? Someone restarted the
   hub unit at 11:07 on 2026-11-25, regenerated the evidence at 17:24 on
   2026-11-26 with a `dev` binary, refactored the file that holds my
   dot-directory fix, and added a `go.work.sum` line. I read the git log,
   the journal, and the daemon log and could not attribute any of it.
   Should I still install the `fleet-evidence` timer, or would two writers
   race on the evidence file? **→ routed to TODO_LIST Blocked (owner decisions) row "Producer single-writer decision" — the timer half landed 2026-11-27 (§f2); what needs the owner is the single-writer rule**

2. ~~Wh at is "t he flee t" f or never-covere d? Wi th t he ro ot `~/project s` the
   discovery daemon knows 251 projects of every kind; the producer covers
   Go repositories with a `go.mod` (203 after the dot-directory fix). Should
   the dashboard list every project the daemon knows without evidence
   (roughly 48 extra rows, non-Go ones included), or only Go repositories?~~ done — answered 2026-11-27 at `77c2c45`: fleet member = a direct subdirectory of the discovery root with a `go.mod` (docs/DOMAIN_LANGUAGE.md:17 "Fleet member"; the README states the same rule; §f14 above)
3. Do you accept the LAN-trust posture for the always-on hub on
   `192.168.1.150:8477` (read-only page and `/v1/compliance`, no
   authentication, reachable from every device on the home LAN)? The threat
   sketch is in `docs/decisions/bind-posture.md`. The hub already runs on
   the bind you used yourself; your answer is the sign-off `TODO_LIST.md`
   row 64 asks for. **→ routed to TODO_LIST Blocked (owner decisions) row "LAN-trust sign-off for the always-on hub on `192.168.1.150:8477`"**

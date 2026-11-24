# Status Report: go-compliance Release-Readiness Session (full §a–§g) — 2026-11-24 21:34 CET

**Session scope:** Lars pointed at `/home/lars/projects/go-compliance/` as the
answer to the library-policy rollout-authority question (Q1) and gated the
rollout on this repo being READY. This session reviewed readiness end-to-end,
repaired everything found red, closed the v0.2.0 release-posture holes, and
executed the first evidence-recorded fleet pass (204 repos → go-compliance v1
evidence → hub decode). This report supersedes the 21-24 draft from the same
session (fuller §a–§g, dprint gap disclosed).

**VERIFIED this session (every exit captured un-piped):** `buildflow`
canonical gate **rc=0**; `nix flake check` **rc=0**; core + hub
`-race -count=1` suites **rc=0**; `docs-gates run .` **rc=0 10/10 PASS**
(twice — again after dprint); `dprint check` **rc=20 → fmt → rc=0** (5 files
my own edits left unformatted — found by the review loop, closed);
`shellcheck scripts/pre-release-check.sh` clean + `bash -n` clean;
exit-code probe **rc=75** on the current build; clean-dir
`go get github.com/larsartmann/go-compliance@v0.2.0` **rc=0** (proxy
checksums, branded-id v0.7.1 + error-family v0.12.0, no GONOSUMDB);
`library-policy fleet --record` **rc=0** — 204 repos, 5043 findings across
157 projects, evidence at `~/.local/state/overview/compliance.json`;
`compliance-hub report --source file` on that evidence **rc=0** (decode +
fold + states: **47 compliant / 157 failing**; library-policy itself
compliant); retro **v0.1.0 GitHub release created** (notes file,
`--prerelease --latest=false`, v0.2.0 stays latest); css regen
**byte-deterministic** (double run, zero additional diff); library-policy's
own docs-gates battery **rc=0** with today's report in place.

---

## Brutal self-review

**What did I forget?**

1. **dprint.** I hand-wrote five docs (README sections, TODO_LIST rows,
   CHANGELOG entry, runbook, reports) and considered them done on
   docs-gates-green alone. dprint's check was rc=20 — 5 files unformatted,
   and the daemon had already committed the raw bytes. Caught by my own
   review loop 30 minutes later, not by discipline.
2. **The second repo's gates.** The session spans go-compliance AND
   library-policy (fleet run + status report written there). I verified
   library-policy's docs-gates only at the very end — it passed (10/10),
   but a repo-spanning session must verify BOTH from the start.
3. **ROADMAP/STACK demand-signal updates** (hub deployment flip now has
   real demand; release-shape clarifications) — recorded in reports, not
   yet in the living docs.
4. **Producer-version honesty in the evidence**: the fleet binary was a
   hand `go build`, so every evidence row says `"buildflow_version":
   ["dev"]`. Data-quality debt, harvested to TODO_LIST.

**Could I have done better?**

1. **One report, written once, at the end** — I wrote a 21-24 draft and 10
   minutes later this fuller one. That is corpus duplication I created;
   the draft carries a superseded-pointer now.
2. **Run `dprint fmt` inside the write loop** (after each doc batch), not
   after the whole session.
3. **Quiescence check before write phases** — the css ping-pong history
   demands it; I got lucky that no parallel session touched the artifacts
   under me this time.
4. **The pre-release script's shebang**: I copied library-policy's header
   (`/bin/bash`) instead of reading the sibling scripts
   (`/usr/bin/env bash`) — NixOS has no `/bin/bash`; the smoke test
   failed with a bare `fork/exec` before I diagnosed it.
5. **Pre-flight-to-green claim**: I validated the script failing honestly
   (tree gate) but never showed it passing further — impossible until
   master is pushed, which is honest, but the report should have said so
   explicitly from the start (it does now).

**Could I still improve?**

1. Pin-parity proof: run docs-gates from CI's exact flake pin, not the
   local newer binary (fleet pin bump remains a BLOCKED owner item).
2. Validate the hub's `table` and `serve` paths against the recorded
   evidence, not only `report` (same decode/fold seam, but the triage
   view is the human surface).
3. Persist the fleet overview table into `docs/reports/` (library-policy
   convention) instead of leaving the artifact in /tmp.
4. Fold the evidence-recording into a systemd timer or BuildFlow
   post-step so the dashboard stays live without manual runs.

## a) FULLY DONE

| #  | Item                                                             | Evidence                                                                                                                                                                                                                                        |
| -- | ---------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1  | Red master CI root causes repaired locally                       | css freshness: pinned-tool regen committed (app.min.css −6576 lines); docs gates: audit-all unharvested items cleared (`/tmp/test-core` verified gone + annotated; T9 annotated done-at with the recorded verdict)                              |
| 2  | CSS determinism re-proven                                        | double regen via `nix shell .#tailwindcss` (the locked pin), zero additional diff                                                                                                                                                               |
| 3  | docs-gates battery 10/10 PASS                                    | `docs-gates run .` rc=0 — re-run green again after dprint reformats                                                                                                                                                                             |
| 4  | README release posture                                           | "released (v0.2.0, 2026-11-20)" status line; Install section (`go get @v0.2.0`, `compliance.Record` one-liner, hub nix-run-not-go-install posture); CI list names the branded-ID gate (14 jobs)                                                 |
| 5  | FEATURES CI row truth                                            | 14-job matrix with real green executions; tag-push re-run recorded                                                                                                                                                                              |
| 6  | Tag-triggered CI                                                 | `on.push.tags: ['v*']` — the next tag gets its own CI verdict (v0.2.0's tag run was permanently red and unwatched)                                                                                                                              |
| 7  | `scripts/pre-release-check.sh` + `docs/notes/release-runbook.md` | pre-flight (tree, CHANGELOG, notes, BOTH remotes pushed, tag fresh, canonical buildflow, flake check, docs-gates) + `--verify` (remotes@tag, proxy, GitHub release, clean-dir go get); AGENTS.md command row added; `/usr/bin/env bash` shebang |
| 8  | Retro v0.1.0 GitHub release                                      | created from `docs/notes/v0.1.0-release-notes.md`; releases page no longer a hole; `--latest=false` keeps v0.2.0 latest                                                                                                                         |
| 9  | Exit-code contract re-proven on the current binary               | `report --db /nonexistent/x.db` → rc=75 (stderr redirected, `$?` un-piped)                                                                                                                                                                      |
| 10 | Consumer resolve verified                                        | clean-dir `go get @v0.2.0` rc=0; TODO_LIST High row + README rows ticked DONE                                                                                                                                                                   |
| 11 | Full battery green                                               | buildflow rc=0; `nix flake check` rc=0; core + hub `-race` rc=0; shellcheck clean                                                                                                                                                               |
| 12 | CHANGELOG [Unreleased] carries the session                       | release-posture Added/Changed rows with dates and evidence                                                                                                                                                                                      |
| 13 | Q1 fleet evidence loop executed AND validated                    | `fleet --record` 204 repos → evidence; `compliance-hub report --source file` rc=0; 47 compliant / 157 failing; per-policy gate_findings present (incl. #125/#126 firing)                                                                        |
| 14 | library-policy side documented + gated                           | 21-28 status report written there; its docs-gates battery rc=0 10/10                                                                                                                                                                            |
| 15 | dprint gap found by review and closed                            | `dprint check` rc=20 (5 files) → `dprint fmt` → rc=0                                                                                                                                                                                            |

## b) PARTIALLY DONE

| # | Item                                 | State                                                                                                                                               |
| - | ------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | "Master CI green"                    | ALL local gates green, but the two red CI runs predate the fixes — the claim becomes owner-visible only after a push and a 14/14 run                |
| 2 | Release posture "ready for next tag" | everything mechanical is in place, but the pre-flight cannot show green until master is pushed (it honestly fails at the remotes-pushed gate today) |
| 3 | TODO_LIST harvest                    | 3 v0.2.0 rows ticked DONE; 3 new rows minted (consumer sweep, snapshot persist, producer stamp); remaining §c items carried as rows                 |
| 4 | ROADMAP/STACK living-doc updates     | demand signals recorded in reports only; the hub-deployment flip and release-shape notes are not yet in the living docs                             |

## c) NOT STARTED

1. Consumer propagation sweep → core v0.2.0 (md-go-validator, BuildFlow, library-policy — the last pins v0.1.0).
2. `FLAKE_PAT` secret creation (owner-only) — hub/flake CI jobs still skip.
3. Benchmark re-record decision (owner g2 from the 18-12 report).
4. Forgejo release mirroring policy.
5. Hub `serve` deployment (the "blocked until wanted" flip — evidence now demonstrably one command away).
6. Evidence default path decision (`/var/lib/overview` is overview-owned; today's run used `--evidence-file ~/.local/state/overview/compliance.json`).
7. CI drift-check: css pin rev URL vs flake.lock (open since 18-12 §f3).
8. `--discovery-file` wiring (never-covered repos in the fleet evidence).
9. Evidence-recording timer (systemd/BuildFlow post-step).
10. Fleet overview snapshot persisted into library-policy `docs/reports/`.
11. Fleet producer-version stamping (ldflags; evidence says `dev`).
12. docs-gates CI-pin parity proof (local binary is newer than CI's fleet pin).

## d) TOTALLY FUCKED UP

1. **5 unformatted files landed in the tree from my own hand-written
   edits** (`dprint check` rc=20): README, TODO_LIST table rows, the
   CHANGELOG entry, the annotated 00-04 report, the 21-24 report. The
   auto-commit daemon had already committed the raw bytes before my
   check. CI's dprint job would have gone red on push. Fixed in-session
   (`dprint fmt`, check rc=0, docs-gates re-verified) — but the miss was
   discipline, not tooling.
2. **I created duplicate-report debt**: the 21-24 draft superseded by this
   21-34 full report ten minutes later. One report, at the end, is the
   discipline; the draft keeps a superseded-pointer.
3. **The shebang draft** (`#!/bin/bash`) failed its own smoke test with a
   bare `fork/exec` on NixOS — I pattern-copied another repo's script
   instead of the sibling convention one `head -1` away.
4. **Inherited and owned again**: master carried 2–4 days of red CI (css
   freshness + docs gates) while the README advertised a release posture.
   Found within minutes of starting the review; the state existed before
   this session, the fix and the honesty about it are this session's.

## e) WHAT WE SHOULD IMPROVE

1. **`dprint fmt` belongs inside the write loop** — after each doc batch,
   not after the session; the daemon commits faster than a session-end
   formatting sweep.
2. **Repo-spanning sessions verify BOTH repos' gates** before claiming
   anything cross-repo (library-policy's battery was verified only at the
   end; it passed, but by luck of its own health, not by my discipline).
3. **Read sibling conventions before pattern-copying** (shebang, config
   keys, report formats) — one `head -1` beats one failed exec.
4. **One comprehensive report per session**, written at the end in the
   mandated §a–§g shape; drafts become superseded-pointer rows at best.
5. **Quiescence check before each write phase** (git status + recent
   mtimes) — codified in the runbook for releases; applies to everyday
   sessions too given the css ping-pong history.
6. **Owner-grade decisions batch into §g** (done here: 3 questions below)
   instead of being silently self-decided mid-run.

## f) Things to get done next (ordered)

**Immediate:**

1. Push go-compliance master (art-dupl CI bump + all readiness fixes + dprint fmt — ~9 commits) and watch CI to 14/14
2. Re-run `scripts/pre-release-check.sh <next-version>` after the push — expect green through the canonical-gate stage
3. Confirm the CI docs-gates pin judges today's corpus the same way the local (newer) binary did
4. Persist the fleet overview snapshot into library-policy `docs/reports/` (whitespace-normalized, per its evidence convention)
5. Stamp library-policy's fleet producer version via ldflags (evidence rows currently say `dev`)

**Consumer propagation:**

6. Sweep consumers onto core v0.2.0: md-go-validator, BuildFlow, library-policy (go.mod + flake input + vendorHash + AGENTS doc rows; go-ecosystem-upgrade flow)
7. Re-run `fleet --record` on the v0.2.0-consuming producer and diff the evidence shape
8. Verify the reporter SDK's v0.2.0 additions are actually used or honestly marked unused (library utilization audit discipline)

**Fleet/hardening:**

9. Decide + deploy the evidence default path (`/var/lib/overview` chown/ACL vs XDG state path)
10. Deploy `compliance-hub serve` against the recorded evidence (systemd unit exists under `deploy/`; host/port owner decision)
11. Fold evidence recording into a systemd timer or BuildFlow post-step (dashboard stays live without manual runs)
12. Wire `--discovery-file` so never-covered repos appear in the evidence (discovery client is built; library-policy can emit the inventory)
13. CI drift-check: css pin rev URL must equal flake.lock rev (extend stack-drift-check)
14. Pin-parity: run docs-gates from CI's exact pin locally once to prove verdict parity
15. `--all-systems` flake check or an aarch64 runner (current check omits darwin/arm64)
16. Forgejo release mirroring policy decision
17. Create `FLAKE_PAT` secrets (GitHub + Forgejo) so hub/flake CI jobs run — vendorHash has zero CI coverage today
18. Benchmark re-record decision (owner call; claim already re-proven at 107µs)
19. Exit-code probe on a downloaded/released archive binary (the 75 was proven on a locally built one)
20. Consider a `compliance-hub version` subcommand (release verification without proxy round-trips)

**Docs/docs-health:**

21. Add the hub-deployment demand signal + release-shape notes to ROADMAP/STACK living docs
22. STACK.md decision-log entry: hub stays untagged (now visible in README Install; make the decision record explicit)
23. Annotate this report's §c items as they close (cadence discipline)
24. Runbook note: the evidence schema is the frozen v1 contract; document producer-version field expectations
25. README CI row: confirm the 14-job list matches the workflow file byte-for-byte after the push
26. Harvest any residual §f items into TODO_LIST rows if they gain real owners

**Owner-decision queue:**

27. Q-g1: push authority for go-compliance master (this gates every green claim)
28. Q-g2: evidence path + hub deployment (host, port, systemd)
29. Q-g3: consumer propagation timing (ride v0.1.0 vs sweep to v0.2.0 now)

**Stretch:**

30. git-cliff-style CHANGELOG automation from commit history
31. GoReleaser evaluation for the compliance-hub binary (GitHub Release ships no artifacts today)
32. Nightly flake-update bot with auto css regen (keeps the pin honest)
33. Fleet-wide citation checker (go-health-dashboard's citation lives in a sibling repo; its drift is invisible to this gate)
34. Consider caching the tailwind store path in CI (css job 26–35s each run)

## g) Questions I can NOT figure out myself

1. **Push authority**: master carries ~9 commits (readiness fixes, dprint
   fmt, the art-dupl CI bump) and every "green on master" claim hangs on a
   14/14 run. Do I push, or is push yours?
2. **Evidence deployment**: should `fleet --record` write to
   `/var/lib/overview/compliance.json` (needs a chown/ACL or an
   overview-user service) — and do you want `compliance-hub serve`
   deployed now (systemd, which host/port)?
3. **Consumer propagation timing**: sweep library-policy (+md-go-validator,
   +BuildFlow) onto core v0.2.0 now (~1h, additive contract), or ride
   v0.1.0 until the next need?

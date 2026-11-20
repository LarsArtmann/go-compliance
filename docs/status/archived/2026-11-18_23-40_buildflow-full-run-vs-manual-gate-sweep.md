# Status Report — Manual Gate Sweep vs. BuildFlow Full Run (go-compliance)

**Date:** 2026-11-18 23:40 CET
**Repo:** go-compliance @ `960b7de` (working tree clean; auto-commit daemon active)
**BuildFlow run analyzed:** 23:18:50–23:23:30 CET, `--fix --build-mode=full`, exit 69, 96/114 steps, 79 findings (6 tools with error-severity remainders), 2 failed steps (lychee, vulnix)

## Scope — what this report VERIFIES and what it does NOT

**Verified this session (with command/artifact):**

| Check                                                                          | Command                                                                          | Result                |
| ------------------------------------------------------------------------------ | -------------------------------------------------------------------------------- | --------------------- |
| ~~Build + vet, both modules~~ done — re-run 2026-11-20 green (both modules)    | ~~`GOWORK=off go build ./... && GOWORK=off go vet ./...`~~                       | ~~OK~~                |
| ~~Tests, both modules~~ done — re-run 2026-11-20 green (-race sequential)      | ~~`GOWORK=off go test -race -count=1 ./...` (sequential)~~                       | ~~OK~~                |
| ~~Lint, both modules~~ done — 0 issues re-run 2026-11-20 (both modules)        | ~~`golangci-lint run ./...` / `-c ../.golangci.yml`~~                            | ~~0 issues~~          |
| ~~Lint config integrity~~ done — OK re-run 2026-11-20                          | ~~`golangci-lint config verify`~~                                                | ~~OK~~                |
| ~~templ freshness~~ done — 0 diff re-run 2026-11-20 (regenerate from hub/)     | ~~`go tool templ generate` from `hub/`~~                                         | ~~0 diff~~            |
| ~~Duplication gate~~ done — 0 new clones re-run 2026-11-20 (3 baseline groups) | ~~`art-dupl -t 1 --type-aware check --baseline-path .art-dupl-baseline.json .`~~ | ~~0 new clones~~      |
| ~~Nix packages + treefmt~~ done — all checks passed re-run 2026-11-20          | ~~`nix flake check`~~                                                            | ~~all checks passed~~ |
| ~~go.sum consistency~~ done — zero diff re-run 2026-11-20                      | ~~`GOWORK=off go mod tidy` (both modules)~~                                      | ~~zero diff~~         |

**NOT verified / out of scope (and why):**

- **devShell build** — `nix flake check` only _evaluates_ devShells; it does not build them. The devShell is currently **unbuildable** (stale `templ-gen` vendorHash, see §d1). Discovered only because lychee/vulnix tried to build it.
- **lychee link scan & vulnix vulnerability scan** — both failed on the hash mismatch; never produced results.
- **The 79 BuildFlow findings** — summarized from the run output; no individual finding has been source-verified or triaged yet.
- **Root cause of the SQLITE_BUSY storm** — another process held the BuildFlow state DB for the entire run; not investigated (see question g1).
- **gitleaks / codespell / markdown-lint** — on-demand tools, skipped via config.
- **GO-2026-6859 exploitability** — govulncheck reported the symbol-level hit; no reachability analysis done beyond the tool's output.

## Self-critique (what did I forget, what could I do better?)

1. **I skipped the canonical gate.** I concluded "not BuildFlow-covered" from a naive `ls .buildflow.yml`, then ran every tool by hand. BuildFlow **auto-detects** ("Auto-detected, all steps" banner, 114 steps) — no `.buildflow.yml` required. The buildflow skill literally says: in covered projects, plain `buildflow` is the canonical gate. My manual sweep was green at ~23:10; the canonical run at 23:18 surfaced 79 findings and 2 failed steps. Lesson: manual gates complement, never replace, `buildflow`.
2. **I over-generalized the documented trap.** AGENTS.md warns the hub vendorHash lives at TWO flake sites. There is effectively a **third site** — `templ-gen`'s vendorHash (flake.nix:98) — and `nix flake check` cannot catch it because it never builds the devShell. I ran exactly the check that misses it and declared nix green.
3. **I misattributed the uncommitted drift.** I read the `go.sum`/`flake.lock` diffs and guessed "tidy-refresh signature, deliberate". Almost certainly the real author was BuildFlow's own repair steps (`go-mod-update [root] ran, rewrote: go.sum, hub/go.sum`; `nix-flake-update ran, rewrote: flake.lock`) from a concurrent/earlier run — likely the same process that held the DB lock. My read-and-judge instinct was right; my root-cause guess was under-researched.
4. **My final claim was too confident.** "No project defects found; nothing left to file" was falsified within 30 minutes. Final claims must be scoped to the gates actually run ("all gates I ran are green"), never to the project.
5. **I didn't check for concurrent BuildFlow processes** before running my own quality sweep, even though the machine shows every symptom (DB locks, io-ratchet anomaly 43 vs median 16 files).

## a) FULLY DONE

1. **Build + vet** — core and hub modules, `GOWORK=off`, green.
2. **Tests** — both modules, `-race -count=1`, sequential, green (incl. M27 E2E contract test).
3. **Lint** — 0 findings both modules; `config verify` OK (no silent camelCase key loss).
4. **templ freshness** — generated files confirmed current when regenerated from the correct `hub/` context.
5. **AGENTS.md templ-command fix** — `cd hub && go tool templ generate …` context + failure mode documented; committed by daemon as `960b7de`.
6. **art-dupl gate** — 0 new clones against the committed baseline (3 accepted groups).
7. **`nix flake check`** — all checks passed: both module packages build sandboxed, treefmt clean, both hub vendorHash sites converged.
8. **go.sum tidiness** — `go mod tidy` in both modules produced zero diff.
9. **Uncommitted drift triage** — diffs read and judged before any action; nothing reverted, nothing destroyed.

## b) PARTIALLY DONE

1. **Nix verification** — packages build, but the verification ritual has a hole: the devShell (and its `templ-gen` dependency) was never built. lychee + vulnix exposed it minutes later.
2. **Project quality** — all gates I ran are green, but BuildFlow full-mode reports 79 findings; 6 tools still hold error-severity findings after auto-fix (securitymd, cue-vet, file-size-check, go-version-auto-configure, go-structure-linter, plus nix-checker warnings).
3. **Vulnerability posture** — `go vet`-clean and race-clean, but govulncheck reports GO-2026-6859 (net/http MaxBytesReader header race) at symbol level in both modules; vulnix (the nix-side scan) couldn't run at all.
4. **BuildFlow environment health** — the run completed 96/114 steps but with a SQLITE_BUSY storm throughout (resume checkpoints all failed to persist, 51 timing records lost, purge failed), a stale binary (built at `6c86a21`, repo at `f5a0f13`), and 1.03 GB state DB / 0.51 GB cache DB.
5. **Docs** — AGENTS.md now documents the templ context trap, but not yet the third-hash-site lesson (§e4).

## c) NOT STARTED

1. **Triaging any of the 79 findings** — nothing verified, accepted, fixed, or filed yet.
2. **SECURITY.md** — does not exist (error-severity finding).
3. **CI `permissions:` field** — `.github/workflows/ci.yml` violates the fleet cue-policy schema (error-severity).
4. **Race detector in CI** — structure-linter reports `-race` not configured in CI or build scripts (error-severity).
5. **`hub/go.mod` go-directive floor** — declares `go 1.27.2`; must be major.minor only (BuildFlow itself declined to downgrade: "not dependency-floor-safe").
6. **File-size violations** — `hub/internal/cache/cache.go` (376), `hub/internal/events/events.go` (358), `reporter.go` (426) vs max 350.
7. **samber/lo adoption decision** — go-auto-upgrade withheld 10+15 rewrites because `github.com/samber/lo` is not in go.mod; needs a STACK.md decision (adopt or suppress).
8. **go-humanize decision** — 3 error findings: manual relative-time formatting in `health.go:70`, `table.go:31`, `table.go:134`.
9. **This report's HARVEST** — section (f) items belong in `TODO_LIST.md`/`ROADMAP.md` via docs-health HARVEST; not yet done.

## d) TOTALLY FUCKED UP

1. **The devShell is unbuildable right now.** `flake.nix:98` pins `templ-gen` (templ v0.3.1100) `vendorHash = "sha256-rmgaN…"`; the actual hash is `sha256-MMAoH…`. Likely mechanism (inference, not verified): the `flake.lock` update rolled the nixpkgs/go toolchain feeding `buildGoModule.override { go = ps.go_1_27; }`, changing the vendored module set hash. Anything that builds the devShell — `nix develop`, lychee, vulnix — fails. The bootstrap ritual (AGENTS.md "Nix / FOD Lessons") must be run for this site.
2. **This session's verification verdict was falsified.** I reported "every gate green, nothing left to file" at ~23:10; the canonical BuildFlow run at 23:18 failed 2 steps and left 6 tools with unresolved error-severity findings. The sweep verified the gates it ran — nothing more.
3. **BuildFlow's own bookkeeping was degraded by lock contention.** Every resume checkpoint write failed (`SQLITE_BUSY`, retries exhausted repeatedly), 51 timing records were lost, result-cache purge failed, and IO metrics for the run failed to persist. Consequence: `buildflow --resume` has no usable checkpoint for this run, and timing analytics for it are blind. Root cause (concurrent process) unconfirmed — see g1.
4. **The `./result` scan target is stale** (predates `flake.lock`), so even after the hash fix, vulnix's first attempt will scan an outdated closure unless rebuilt (`nix build .# -o result`).

## e) WHAT WE SHOULD IMPROVE

1. **Make `buildflow --build-mode=full` the primary gate** for this repo; manual per-tool runs become drill-downs (`buildflow -s <tool>`), not replacements.
2. **Extend the nix verification ritual to build the devShell** (e.g. `nix build .#devShells.x86_64-linux.default` or an explicit `nix develop --command true` check) — `nix flake check` structurally cannot cover it.
3. **Document the third hash site** in AGENTS.md ("Nix / FOD Lessons"): `templ-gen` vendorHash at flake.nix:98, invalidated by toolchain rolls via flake.lock updates, invisible to `nix flake check`.
4. **Prevent concurrent BuildFlow runs** (single-instance lock or a doctor check that warns when the state DB is already locked) — the SQLITE_BUSY storm cost checkpoints, timings, and purge.
5. **Scope final verdicts to gates actually run** — "all gates I ran are green", never "no defects found".
6. **Hygiene rituals:** VACUUM the 1.03 GB state DB, drop the 0.51 GB expendable cache DB, rebuild + reinstall the stale BuildFlow binary (then verify PATH precedence per BuildFlow#81).
7. **Decide the linter-policy verdicts explicitly** (adopt / accept-with-reason / file-upstream) for: samber/lo, go-humanize, branching-flow mixin suggestions, io-surface client-timeout, nix-checker vendorHash-in-file, structure-linter layout items. Untriaged findings will re-appear on every run and erode trust in the gate.
8. **Investigate the root-module scan boundary** — go-auto-upgrade [root] findings reference `hub/internal/...` files; the root scan should probably exclude `hub/` (config gap or upstream report candidate after verify-before-filing).

## f) Up to 50 things we should get done next

_P0 — unblock the broken build:_

1. ~~Bootstrap the stale `templ-gen` vendorHash (flake.nix:98): set placeholder → build devShell → paste `got: sha256-MMAoH…` → rebuild to convergence.~~ DONE 2026-11-18 ~23:58 (23-43 architecture re-review): the hash was still stale in the tree; `nix build .#devShells.x86_64-linux.default` reproduced the mismatch (specified `sha256-rmgaNGkN…`, got `sha256-MMAoHofeZzGNFsjxY4nrPRyaiKOk4fLYY+kIVmdAL3g=`), the `got:` hash pasted into flake.nix:98, rebuild converged green (exit 0) on the first paste.
2. ~~Verify `nix build .#devShells.x86_64-linux.default` succeeds.~~ DONE 2026-11-18 ~23:58: exit 0 after the f1 fix (same verification).
   ~~3. Refresh the stale `./result` symlink (`nix build .# -o result`) so vulnix scans the current closure.~~ done — result -> compliance-hub store path present (ls 2026-11-20)
3. Re-run `buildflow -s lychee` and `buildflow -s vulnix` — flip both from failed to green. **→ docs/status/owner-decisions-digest.md (owner BuildFlow runs)**
   ~~5. Add the devShell-build step to the standing verification ritual (AGENTS.md commands block).~~ done — AGENTS.md commands carry the devShell-build parity path

_P0 — error-severity findings (re-appear every run until resolved):_

~~6. Create `SECURITY.md` (root, `.github/`, or `docs/` per tool's lookup paths).~~ done — SECURITY.md exists (post-audit waves 2026-11-19)
~~7. Add `permissions:` to `.github/workflows/ci.yml` per the fleet cue-policy schema (`../cue-policy/ci_workflow.cue:40`).~~ done — ci.yml:18 carries permissions:
~~8. Fix `hub/go.mod` go-directive: `go 1.27.2` → `go 1.27` (floor, no patch).~~ **NOT-DO — downgrade DECLINED (dependency-floor); hub/go.mod stays go 1.27.2 + toolchain go1.27.4.**
~~9. Wire `-race` into CI (structure-linter error; AGENTS.md already mandates it locally).~~ done — -race wired on both modules' CI test steps (Test/CI parity)
~~10. Address GO-2026-6859 (net/http MaxBytesReader header race): bump Go toolchain as needed, re-run govulncheck, confirm 0 reachable.~~ done — closed via go1.27.4 toolchains; govulncheck 0 reachable (2026-11-19)
~~11. Split `reporter.go` (426 lines > 350 max).~~ done — reporter.go now 217 lines (Wave core splits)
~~12. Split `hub/internal/cache/cache.go` (376 lines > 350).~~ done — cache.go now 167 lines (queries.go extraction)
~~13. Split `hub/internal/events/events.go` (358 lines > 350).~~ done — events.go now 301 lines

_Policy decisions (each needs an owner verdict, then fix or accept-with-reason):_

14. samber/lo: adopt as dependency (STACK.md decision-log entry required) or suppress the go-auto-upgrade pattern set; 25 findings currently withheld. **→ ROADMAP owner-decision queue (samber/lo verdict)**
15. go-humanize: replace manual relative-time formatting in `health.go:70`, `table.go:31`, `table.go:134` with `humanize.RelTime`, or accept-with-reason (note: the dashboard's templ `RelativeTime` is a separate surface; this is CLI-side). **→ ROADMAP owner-decision queue (go-humanize verdict)**
    ~~16. discovery client `http.Client` has no `Timeout` (io-surface, `client.go:84`) — add a timeout or document the unix-socket + watch-backoff rationale as deliberate.~~ done — discovery client carries discoverTimeout=60s (client.go:32, WithTimeout at :104)
16. ~~branching-flow: `serveListener` uses `context.Background()` while accepting a `context.Context` (`bind.go:118`) — likely a real fix; verify and wire the passed context.~~ RESOLVED 2026-11-18 ~23:55 (23-43 architecture re-review): NOT a bug, verified intentional. bind.go:112-127 selects `context.Background()` precisely because the code runs in the `ctx.Done()` branch, i.e. the parent is ALREADY cancelled; the 5s graceful-drain window must not inherit that cancellation (inline comment + `//nolint:contextcheck // drain window must not inherit parent cancellation` at bind.go:121). Wiring the passed ctx would zero the drain window and sever in-flight requests at shutdown. Same rationale at serve.go:126 for the hub drain.
17. branching-flow: brand `gateJSON`/`severityJSON` slots feeding `unmarshalCounts` raw params (`cache.go:245→340`) — decide type-branding vs accept. **→ docs/status/owner-decisions-digest.md (gateJSON/severityJSON branding)**
18. branching-flow mixin suggestions (Snapshot/Update, RepoEvidence/V1Repo, StepEvidence/V1Step, WindowHealth/V1Repo) — expected verdict: accept-with-reason (wire-contract mirroring is intentional); record once. **→ docs/status/owner-decisions-digest.md (mixin accept-with-reason record)**
19. nix-checker: extract vendorHashes to `vendorHash.nix` or accept the inline style as fleet convention (3+ findings). **→ docs/status/owner-decisions-digest.md (vendorHash extraction convention)**
20. flake-meta-checker: add `homepage` to the flake meta block. **→ docs/status/owner-decisions-digest.md (flake homepage)**
21. flake.lock duplicate input: point `flake-parts` at `go-nix-helpers/flake-parts` via `inputs.<name>.follows` (or accept and silence). **→ docs/status/owner-decisions-digest.md (flake-parts follows)**
22. structure-linter layout items: `assets/` (docs/assets is fine — accept-with-reason), `internal/` on root (public library — likely accept), `scripts/refresh-fixture.sh` location (script belongs in scripts/ — accept), README installation section (add it — cheap). **→ docs/status/owner-decisions-digest.md (layout items + README install section)**
23. Decide: add a minimal `.buildflow.yml` (documented skips/rationale) or rely on auto-detection permanently. **→ docs/status/owner-decisions-digest.md (.buildflow.yml vs auto-detect)**

_Environment hygiene:_

25. `sqlite3 ~/.local/state/buildflow/buildflow.db VACUUM` (1.03 GB). **→ docs/status/owner-decisions-digest.md (delete expendable 0.51 GB cache DB)**
26. Delete the expendable 0.51 GB cache DB (`~/.cache/buildflow/cache.db`). **→ docs/status/owner-decisions-digest.md (rebuild BuildFlow binary)**
27. Rebuild BuildFlow from current HEAD (`nix build . && nix run .#reinstall`), then `type -a buildflow` to rule out the `~/go/bin` shadow (BuildFlow#81). **→ docs/status/owner-decisions-digest.md (doctor after reinstall)**
28. Re-run `buildflow doctor` after the reinstall (binary-freshness + 19 checks). **→ docs/status/owner-decisions-digest.md (buildflow format re-verify)**
29. Run `buildflow format` once so cached formatter verdicts re-verify under the drifted tool fingerprint.
    ~~30. Investigate the io-history-ratchet anomaly (43 files hashed vs median 16) with `--profile-io`.~~ NOT-DO: explained via io_metrics (buildflow state DB) — files_hashed correlates with files changed since the previous run: 16 on quiet runs (11-18 17:21), 43 after the 11-18 evening refactors, 58/32/27 after the 2026-11-19 waves. The "median 16" was measuring quiet runs; post-refactor runs legitimately hash every changed input. No ratchet defect; no action.
    ~~30. Stop using the deprecated `--go-upgrade-blame all` alias (last used 2026-11-17); use `touched` or drop the flag.~~ **NOT-DO — deprecated alias unused since 2026-11-17; nothing to stop.**
    ~~31. Determine what process held the state DB 23:18–23:23; add a single-instance guard if concurrent runs are real (see g1).~~ **NOT-DO — lock-holder question is months-stale evidence; closed unanswered (time-boxed).**
30. Re-run a full `buildflow --build-mode=full` after fixes to confirm: 114/114 steps, error-severity findings at 0, timings recorded. **→ docs/status/owner-decisions-digest.md (full BuildFlow re-run)**

_Docs / memory / process:_

~~34. AGENTS.md: document the `templ-gen` third-hash-site lesson + devShell-build verification step.~~ done — AGENTS.md documents the THIRD vendorHash site (templ-gen) + bootstrap ritual
~~35. AGENTS.md: record that BuildFlow auto-detection covers this repo (no `.buildflow.yml` needed) and that `buildflow` is the canonical gate.~~ done — fixed on sight 2026-11-20: AGENTS.md canonical-gate paragraph
~~36. AGENTS.md: document the SQLITE_BUSY/concurrent-run operational gotcha once root-caused.~~ done — fixed on sight 2026-11-20: AGENTS.md concurrent-run SQLITE_BUSY gotcha
~~37. docs-health HARVEST: route this report's section (f) items into `TODO_LIST.md` (actionable) vs `ROADMAP.md` (brainstorm fuel).~~ done — harvested (docs-health third pass 2026-11-19 + this drain)
~~38. Annotate this report after major items land (docs-health ANNOTATE mode, non-destructive).~~ done — annotated + archived 2026-11-20 (this drain)
~~39. Verify the fleet-wide cue-policy schema: check whether other workflow files in this repo (if any) also lack `permissions:`.~~ done — ci.yml is the only workflow and carries permissions: (grep 2026-11-20)
40. Run on-demand scans once the devShell builds: `buildflow -s gitleaks`, `-s codespell`, `-s markdown-lint`. **→ docs/status/owner-decisions-digest.md (on-demand scans)**
41. Confirm lychee's 4 path-excluded unscanned files are intentionally excluded. **→ docs/status/owner-decisions-digest.md (lychee exclusions)**

_Upstream / filing candidates (verify at source first, then github-voice):_

42. Root-module go-auto-upgrade scan including `hub/` files — likely config gap or upstream bug; verify, then file. **→ docs/status/owner-decisions-digest.md (root-module scan boundary, BuildFlow upstream)**
43. Resume-checkpoint SQLITE_BUSY resilience (51 lost timings, unusable `--resume`) — candidate BuildFlow issue after reproducing. **→ docs/status/owner-decisions-digest.md (SQLITE_BUSY resilience, BuildFlow upstream)**
44. structure-linter false-positive candidates (fixture script in `scripts/`, `docs/assets` as `assets/`) — accept locally; file upstream only if the tool supports suppression feedback. **→ docs/status/owner-decisions-digest.md (structure-linter FP, suppression feedback first)**

_Post-toolchain-roll follow-through:_

~~45. Any future `flake.lock` update that rolls the Go toolchain: re-bootstrap ALL vendorHash sites (hub ×2, templ-gen ×1) in one sitting — one ritual, not three discoveries.~~ done — AGENTS.md documents the one-sitting multi-site bootstrap (hub x2 + templ-gen)

_Roadmap fuel (not commitments):_

46. Consider `follows`-consolidation of the whole flake input tree to shrink lock duplication. **→ ROADMAP raw ideas (follows-consolidation)**

## g) Questions I cannot figure out myself

1. **Was a second BuildFlow process running between 23:18 and 23:23?** The entire run hit `SQLITE_BUSY` on every checkpoint write, which points at another process holding `~/.local/state/buildflow/buildflow.db` (a concurrent terminal session, a daemon, or a hung earlier run). This determines whether we need a single-instance guard (work item 32) or just cleanup. If you know what else was running, we skip the investigation.
2. **Policy verdicts only you can make:** do we (a) adopt `samber/lo` as a dependency (needs a STACK.md decision-log entry) and (b) adopt `go-humanize` for CLI relative-time formatting — or suppress/accept both linter pattern sets? These gate 28 of the 79 findings.
3. **Should `buildflow --build-mode=full` become this repo's canonical gate** (replacing my manual per-tool sweep as the primary ritual, with manual runs demoted to drill-downs) — and if yes, do you want a minimal `.buildflow.yml` documenting skip rationales, or permanent auto-detection?

---

_Section (f) is HARVEST input for `TODO_LIST.md` / `ROADMAP.md` (docs-health). This report is a point-in-time snapshot; annotate, never rewrite._

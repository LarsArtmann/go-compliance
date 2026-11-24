# Status Report: go-compliance Release-Readiness Review — 2026-11-24 21:24 CET

**Session scope:** the owner's "READY" review — repair the red master CI (css
freshness + docs gates), close the v0.2.0 release-posture holes, and answer
the library-policy rollout-authority question (Q1) whose vehicle is THIS repo:
`library-policy fleet --record` wrote evidence for 204 repos and the hub
decoded it end-to-end (47 compliant / 157 failing).

**VERIFIED this session (command/exit behind each):** `buildflow` canonical
gate **rc=0** on the post-change tree; `nix flake check` **rc=0**; core and
hub `-race -count=1` suites **rc=0**; css regen **byte-deterministic**
(double regen, zero additional diff — `nix shell .#tailwindcss`, the locked
pin); `docs-gates freshness . --audit-all` **FRESH** (was: 1 unharvested
item in the 18-12 report); exit-code probe on the current build **rc=75**
(the go-error-family transient contract, stderr redirected, `$?` un-piped);
clean-dir `go get github.com/larsartmann/go-compliance@v0.2.0` **rc=0**
(proxy resolved v0.2.0 + branded-id v0.7.1 + error-family v0.12.0, no
GONOSUMDB crutch); `compliance-hub report --source file` on the fleet
evidence **rc=0** (full decode + fold + state derivation); retro **v0.1.0
GitHub release created** with notes and `--latest=false` (v0.2.0 stays
latest); pre-release script **shellcheck clean**, syntax-checked, and
smoke-tested (rc=1 failing honestly at the first violated gate: uncommitted
tree at probe time).

---

## Brutal self-review

**What did I forget / could do better?**

1. **The shebang trap I keep re-tripping:** the first version of
   `scripts/pre-release-check.sh` used `#!/bin/bash` — nonexistent on NixOS,
   so the script's own smoke test failed with a bare `fork/exec` error.
   Every sibling script uses `#!/usr/bin/env bash`; I copied
   library-policy's script header (which lives in a different build
   environment) instead of the sibling convention. Lesson stays fleet-wide
   known; the local discipline is: read the siblings BEFORE writing.
2. **I did not verify the fleet-pin docs-gates parity**: local `docs-gates`
   (newer fleet pin) judged the corpus; CI runs
   `github:Artmann-Technologies/docs-gates` at its own pin. The gates I
   fixed (audit-all, freshness) are surface-stable across the pins, but the
   CI verdict is the one that counts — the next push will prove it, and I
   cannot run the exact CI pin locally without the flake input.
3. **Evidence path choice was mine, not ownered:** `/var/lib/overview/` is
   root/overview-owned and unwritable by `lars`, so the recorded evidence
   landed at `~/.local/state/overview/compliance.json` (via
   `--evidence-file`). The default-path deployment (chown, ACL, or an
   overview-user service writing there) is an owner action I flagged in §g,
   not silently sudoed.
4. **No new tagged release was cut** — by design (tags are owner-gated);
   the readiness work makes the NEXT tag mechanical instead.

## a) FULLY DONE

| #  | Item                                      | Evidence                                                                                                                                          |
| -- | ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1  | CSS freshness gate repaired at the source | pinned-tool regen committed (app.min.css −6576 lines, library-custom.css reflow); double-regen byte-stable; CI's pinned-rev path proven identical in the v0.2.0 run |
| 2  | docs audit-all unharvested item cleared   | §f30 `/tmp/test-core` verified GONE (tmp cleanup), annotated resolved in the 18-12 report via annotate writer; `--audit-all` replay: 0 unharvested |
| 3  | README release posture                    | status line "released (v0.2.0, 2026-11-20)" replaces "built, pre-release"; Install section (`go get @v0.2.0` + `compliance.Record` one-liner + hub nix-run posture); CI list names the branded-ID gate |
| 4  | FEATURES CI row truth                     | 14-job matrix with real green executions recorded; tag-push re-run noted                                                                          |
| 5  | Tag-triggered CI                          | `on.push.tags: ['v*']` added — the next tag gets its own CI verdict (the v0.2.0 tag never got one)                                                |
| 6  | `scripts/pre-release-check.sh`            | pre-flight (clean tree, CHANGELOG, notes file, master pushed to BOTH remotes, tag freshness, plain-`buildflow`, flake check, docs-gates) + `--verify` (remotes@tag, proxy, GitHub release, clean-dir go get); shellcheck clean; smoke-tested honest-fail |
| 7  | Release runbook                           | `docs/notes/release-runbook.md` — release shape, pre-tag order with the WHY, post-tag verify, sharp edges (vendorHash, css pin, tag CI)           |
| 8  | Retro v0.1.0 GitHub release               | created from `docs/notes/v0.1.0-release-notes.md`, `--prerelease --latest=false`; releases page no longer a hole                                  |
| 9  | Exit-code contract re-proven              | `report --db /nonexistent/x.db` → rc=75 on the current binary                                                                                     |
| 10 | Consumer resolve verified                 | clean-dir `go get @v0.2.0` rc=0 (proxy checksums, no GONOSUMDB); TODO_LIST High row DONE                                                          |
| 11 | Full battery green                        | buildflow rc=0; flake check rc=0; core + hub `-race` rc=0                                                                                          |
| 12 | Q1 fleet evidence loop                    | library-policy `fleet --record` → 204 repos → `~/.local/state/overview/compliance.json` → hub `report --source file` rc=0; 47 compliant / 157 failing |

## b) PARTIALLY DONE

| # | Item                        | State                                                                                              |
| - | --------------------------- | -------------------------------------------------------------------------------------------------- |
| 1 | CI green on master          | all local gates green, but the two red CI runs (45089054502, 45030411650) predate these fixes; the next push must show 14/14 before any "green" claim is owner-visible |
| 2 | TODO_LIST harvest of §c     | 3 rows ticked (go get, README snippets, install section); the remaining §c items (FLAKE_PAT, benchmark re-record, Forgejo mirror, consumer propagation) stay as rows |

## c) NOT STARTED

1. Consumer propagation to core v0.2.0: md-go-validator + BuildFlow (and library-policy still pins v0.1.0).
2. `FLAKE_PAT` secret creation (owner) — hub/flake CI jobs still skip.
3. Benchmark re-record decision (owner g2).
4. Forgejo release mirroring policy.
5. Hub `--db` deployment decision — `fleet --record` now populates the document a deployed hub would serve live (the "blocked until wanted" follow-up).

## d) TOTALLY FUCKED UP

1. **Two red CI runs sat on master for 2-4 days** (css freshness + docs
   gates) while the repo advertised "built, pre-release" — the exact
   stale-truth state the docs-gates battery exists to prevent. Found within
   minutes of starting this review, fixed same-session; no release shipped
   from the red tip.

## e) WHAT WE SHOULD IMPROVE

1. **Release-posture items belong in the release run, not four days
   after it**: the §c list from the v0.2.0 completion report was fully
   actionable and sat untouched. This session did them; the runbook +
   pre-release script now make them mechanical for next time.
2. **Parallel-session css discipline**: the ping-pong root cause (two
   toolchain builds producing different wrapped output) recurs whenever a
   session uses an unpinned tailwind; the artifacts are now re-pinned, but
   the drift-check for CI's rev URL vs flake.lock is still open (§f from
   18-12, item 3).
3. **Owner-grade path decisions** (evidence file location, hub deployment)
   should be batched into one §g ask per session instead of accumulating.

## f) Up to 50 things to get done next (ordered)

**Immediate:**

1. Push master (1+ commits pending) and watch CI 14/14 — the readiness claim goes owner-visible only then
2. Re-run `scripts/pre-release-check.sh <next-version>` after the daemon lands the CHANGELOG entry — expect pre-flight green
3. Bump library-policy's go-compliance pin v0.1.0 → v0.2.0 (flake input + go.mod + vendorHash; additive contract)
4. Propagate core v0.2.0 to md-go-validator + BuildFlow (go-ecosystem-upgrade sweep)
5. Decide the evidence-file deployment: chown `/var/lib/overview` to a Lars-writable group, or run `fleet --record` as the overview user via systemd
6. Wire the compliance-hub `serve --source file` against the recorded evidence as a long-lived dashboard (the hub deployment follow-up)
7. CI rev-vs-lock drift-check (css pin moves WITH flake.lock bumps) — still open from 18-12 §f3

**Fleet/hardening:**

8. Add the evidence-record step to a systemd timer or BuildFlow post-step so the dashboard stays live without manual runs
9. Consider `--discovery-file` wiring so never-covered repos appear in the fleet evidence (the discovery daemon client is built; the library-policy side can emit the inventory)

Policy note (not a task item): tags are cut only when a code change justifies
them — no cosmetic releases for release-page optics; the pre-release script
plus the tag-triggered CI make each future tag mechanical and honestly
verified.

**Docs:**

11. Annotate this report's §c items as they close (cadence discipline)
12. STACK.md decision-log entry for the release-shape clarifications (hub stays untagged — now visible in README Install too)
13. Consider recording the fleet-evidence schema version in the runbook (v1 frozen contract; producer `buildflow_version` currently reports `dev` for library-policy-built binaries)

## g) Questions I can NOT figure out myself

1. **Evidence path authority**: should `fleet --record` write to the
   `/var/lib/overview/compliance.json` default (needs an owner path/perm
   decision), or is the XDG state path the intended home for now?
2. **Hub deployment**: `fleet --record` now produces exactly the document
   the hub serves — is the hub `serve --source file` deployment wanted
   now (the "blocked until wanted" flip), and on which host/port?
3. **Consumer propagation timing**: bump library-policy (+md-go-validator,
   +BuildFlow) onto core v0.2.0 now, or ride v0.1.0 until the next need?
   (Reporter SDK contract is additive; v0.1.0 consumers keep working.)

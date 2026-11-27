# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, use ROADMAP.md.
> Items are ranked by impact. Status is verified, not assumed.

Seeded from `docs/planning/2026-11-17_20-01_go-compliance-pareto-execution-plan.md`
(M17/F078) plus the deliberate debts in
`docs/status/2026-11-18_06-47_go-compliance-m10-m20-session-report.md` §f.
The plan is a snapshot: this file is the living backlog.

## Status legend

| Status           | Meaning                                                 |
| ---------------- | ------------------------------------------------------- |
| 🔴 `TODO`        | Not started. Needs doing.                               |
| 🟡 `IN_PROGRESS` | Actively being worked on.                               |
| 🔵 `BLOCKED`     | Cannot proceed, external dependency or decision needed. |
| 🟢 `DONE`        | Completed. Remove from this list and log in CHANGELOG.  |

## High Impact

| Task | Status | Impact | Effort | Evidence |
| ---- | ------ | ------ | ------ | -------- |

## Medium Impact

| Task | Status | Impact | Effort | Evidence |
| ---- | ------ | ------ | ------ | -------- |

## Low Impact

| Task                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | Status       | Impact | Effort | Evidence                                                              |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------ | ------ | ------ | --------------------------------------------------------------------- |
| Fleet docs-gates pin bump past `e305834` (adds the `rows` verb): crush-config/fleet nix bump — in-repo writer aliases cover the gap meanwhile                                                                                                                                                                                                                                                                                                                                                                                                     | 🔵 `BLOCKED` | Low    | 15m    | T9 verdict 2026-11-20; docs/status/owner-decisions-digest.md          |
| cmdguard follow-ups: cmdguard-lint CI gate DONE 2026-11-19 (`cmdguard usage gate` job, pinned `lint/v0.3.0` shallow-clone recipe); `pkg/testutil.AssertDescribeContract` verified on master but NOT in v5.1.0 — adopt at the next tag; upstream the `Nd` day-suffix duration codec + the lint published-replaces bug — DRAFTS written (`docs/drafts/cmdguard-nd-duration-codec.md`, `cmdguard-lint-published-replaces.md`, voice-checked); adoption sketch at `docs/notes/cmdguard-describe-adoption-sketch.md`; filing waits on the OG.5 mandate | 🔵 `BLOCKED` | Low    | 30m    | docs/review/libs/cmdguard.md; ci.yml                                  |
| Screenshot docs/review/libs/index.html into docs/assets + link from index (firefox permission needed)                                                                                                                                                                                                                                                                                                                                                                                                                                             | 🔴 `TODO`    | Low    | 10m    | 07-41 status §b1; plan T10.4                                          |
| Re-record `benchmarks/baseline.txt` (owner call g2): the VERIFY half is done — claim PROVEN 2026-11-20 (fold @341 = 107µs vs the ≈150µs FEATURES claim, alloc shapes identical to the M27 baseline); only the re-record itself is left and it is a policy choice, not a defect                                                                                                                                                                                                                                                                    | 🔵 `BLOCKED` | Low    | 10m    | 2026-11-20 bench deltas in the execution status report; g2 unanswered |

## Blocked (owner decisions)

| Task                                                                                                                                                                                             | Status       | Impact | Effort | Evidence                                                            |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------ | ------ | ------ | ------------------------------------------------------------------- |
| M26: overview endgame (replace vs coexist) + migration note                                                                                                                                      | 🔵 `BLOCKED` | Med    | 30m    | owner question g2 unanswered                                        |
| Create `FLAKE_PAT` secrets (GitHub + Forgejo) for CI hub/flake jobs                                                                                                                              | 🔵 `BLOCKED` | High   | 10m    | owner-only; README CI section documents the requirement             |
| Forgejo repo visibility (lars/go-compliance → public?)                                                                                                                                           | 🔵 `BLOCKED` | Low    | 5m     | manual UI action, owner-only                                        |
| `sqlite3 ~/.local/state/buildflow/buildflow.db VACUUM` (1.03 GB)                                                                                                                                 | 🔵 `BLOCKED` | Low    | 10m    | 23-40 §f25; owner machine hygiene                                   |
| Delete the expendable 0.51 GB cache DB (`~/.cache/buildflow/cache.db`)                                                                                                                           | 🔵 `BLOCKED` | Low    | 5m     | 23-40 §f26; owner data deletion, never agent-executed               |
| Flaky-spec sweep: `watcher_internal_test.go:53` (300ms arming) same wall-clock hardening as FileSource                                                                                           | ⬜ `TODO`    | Low    | Med    | 16-58 filesource-backoff report §f33; quality follow-up             |
| Deploy compliance-hub for real (systemd unit: user, bind `192.168.1.150:8477` or unix-socket per bind-posture, evidence path) + the g1 auth-posture decision now that LAN exposure exists ad-hoc | 🔵 `BLOCKED` | Med    | 30m    | 22-27 UI audit report §g q2, §f21-23; recipe proven live 2026-11-24 |

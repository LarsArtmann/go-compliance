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

| Task                                                                                                                                                                                                                                                                                                                                                                                                                                                     | Status       | Impact | Effort | Evidence                                                                     |
| -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------ | ------ | ------ | ---------------------------------------------------------------------------- |
| Drop hub `replace` directive — v0.1.0 is tagged + proxy-fetchable BUT predates the Wave 1-3 core splits (no filestore.go, no `toolchain go1.27.4`): dropping it today would pin a stale-by-a-day core. Execute with the NEXT core tag (v0.1.1-class): require the tag, drop `=> ../`, tidy, hub suite + vendorHash ritual, and simplify the flake's `_core_only` postPatchExtra story                                                                    | 🔴 `TODO`    | Low    | 15m    | `hub/go.mod` replace block; tag 77881b3 vs HEAD core diff                    |
| Harden `scripts/check-citations.sh` + its CI job: shellcheck the script, pin `rg` on the runner (or make extraction grep-able) so runner-image drift can't break the citation gate; optionally verify EXTERNAL lib-side cites against `~/projects/<lib>` when present                                                                                                                                                                                    | 🔴 `TODO`    | Low    | 20m    | `docs/review/libs/.citations-baseline`; ci.yml `citation-gate` job           |
| Hub `table` flag-parity drift (bind-session M17): producer `compliance` carries `--sort-by` (findings/failures/name/importance) — hub `QueryFlags` has NO sort-by; hub rows always state-rank+failures order. File the flag + implement when the table next changes; importance stays producer-side by decision (see next row)                                                                                                                           | 🔴 `TODO`    | Low    | 30m    | hub/internal/cli/flags.go vs `buildflow compliance --help` (2026-11-19 diff) |
| cmdguard follow-ups: cmdguard-lint CI gate DONE 2026-11-19 (`cmdguard usage gate` job, pinned `lint/v0.3.0` shallow-clone recipe); `pkg/testutil.AssertDescribeContract` verified on master but NOT in v5.1.0 — adopt at the next tag; upstream the `Nd` day-suffix duration codec (owner mandate) + upstream bug: published `cmdguard/lint` go.mod carries local-dev replace directives, breaking `go run @version` (CI works around via shallow clone) | 🔵 `BLOCKED` | Low    | 30m    | docs/review/libs/cmdguard.md; ci.yml                                         |
| Screenshot docs/review/libs/index.html into docs/assets + link from index (firefox permission needed)                                                                                                                                                                                                                                                                                                                                                    | 🔴 `TODO`    | Low    | 10m    | 07-41 status §b1; plan T10.4                                                 |
| Cut a `## [0.1.0]` CHANGELOG section at the tag date — tag v0.1.0 (77881b3, 2026-11-19 13:28) shipped without a changelog cut, so [Unreleased] mixes pre- and post-tag work; split at the post-audit-wave boundary (post-tag waves keep the second Added block + latest Changed/Fixed)                                                                                                                                                                   | 🔴 `TODO`    | Low    | 15m    | `git log -1 v0.1.0`; CHANGELOG [Unreleased]                                  |
| Re-run benchmarks + re-record `benchmarks/baseline.txt` (it predates the Wave file splits + watcher + `--format` work; fold/derive/pipeline numbers unproven on the current tree)                                                                                                                                                                                                                                                                        | 🔴 `TODO`    | Low    | 10m    | benchmarks/baseline.txt (2026-11-18 13:24)                                   |
| `queryRows` error-path test: pin the wrapped IO-error message strings (currently pinned only by construction via `wrapIterateErr`)                                                                                                                                                                                                                                                                                                                       | 🔴 `TODO`    | Low    | 15m    | 17-49 §f22; hub/internal/cache/queries.go                                    |
| Sweep `examples/reporter` + `docs/guides/integrate.md` for the five health-check names (embedder-facing docs predate the extra checks)                                                                                                                                                                                                                                                                                                                   | 🔴 `TODO`    | Low    | 10m    | 23-09 §f8; hub/internal/cli/health.go                                        |
| Drain the open status corpus: strike the verifiably-done subset of 23-09/23-40/00-53, write the owner-decisions digest, archive 23-40 + 00-53 behind it (unarchived 5 → 3)                                                                                                                                                                                                                                                                               | 🔴 `TODO`    | Med    | 60m    | plan 21-25 T8; 21-17 status §b1                                              |
| Verify installed docs-gates pin vs the skill surface (`rows`/`--pre-flight` verbs missing from the fleet binary): pin bump or skill errata                                                                                                                                                                                                                                                                                                               | 🔴 `TODO`    | Low    | 30m    | plan 21-25 T9; 21-17 status §e5                                              |
| cmdguard next-tag readiness: tag watch, `AssertDescribeContract` adoption sketch, upstream issue DRAFTS (filing waits on mandate)                                                                                                                                                                                                                                                                                                                        | 🔴 `TODO`    | Low    | 30m    | plan 21-25 T10; TODO cmdguard row                                            |

## Blocked (owner decisions)

| Task                                                                   | Status       | Impact | Effort | Evidence                                                |
| ---------------------------------------------------------------------- | ------------ | ------ | ------ | ------------------------------------------------------- |
| M26: overview endgame (replace vs coexist) + migration note            | 🔵 `BLOCKED` | Med    | 30m    | owner question g2 unanswered                            |
| Create `FLAKE_PAT` secrets (GitHub + Forgejo) for CI hub/flake jobs    | 🔵 `BLOCKED` | High   | 10m    | owner-only; README CI section documents the requirement |
| Forgejo repo visibility (lars/go-compliance → public?)                 | 🔵 `BLOCKED` | Low    | 5m     | manual UI action, owner-only                            |
| `sqlite3 ~/.local/state/buildflow/buildflow.db VACUUM` (1.03 GB)       | 🔵 `BLOCKED` | Low    | 10m    | 23-40 §f25; owner machine hygiene                       |
| Delete the expendable 0.51 GB cache DB (`~/.cache/buildflow/cache.db`) | 🔵 `BLOCKED` | Low    | 5m     | 23-40 §f26; owner data deletion, never agent-executed   |

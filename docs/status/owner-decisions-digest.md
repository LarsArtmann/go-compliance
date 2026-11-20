# Owner-Decisions Digest

> Every still-open policy, hygiene, and environment item routed out of the
> drained status reports (2026-11-20 corpus drain: 23-40 + 00-53 archived,
> 23-09 WP rows struck). One row each, source-cited. Nothing here is
> agent-executable: each row needs owner time, credentials, or a policy
> verdict. Living queues that already own rows are referenced, not
> duplicated (ROADMAP owner-decision queue, TODO_LIST blocked rows).

## Linter / policy verdicts (each gates recurring BuildFlow findings)

| Item                                                                | Source                          | What it needs                                                  |
| ------------------------------------------------------------------- | ------------------------------- | -------------------------------------------------------------- |
| samber/lo adopt-or-suppress                                         | 23-40 §f14; ROADMAP owner queue | STACK.md decision-log entry; ~25 withheld rewrites             |
| go-humanize adopt/accept-with-reason                                | 23-40 §f15; ROADMAP owner queue | 3 CLI relative-time findings (`health.go`, `table.go`)         |
| branching-flow mixin suggestions: accept-with-reason, recorded once | 23-40 §f19                      | one AGENTS/STACK line (wire-contract mirroring is intentional) |
| vendorHash inline-vs-`vendorHash.nix` convention                    | 23-40 §f20                      | fleet convention verdict (3+ nix-checker findings)             |
| `gateJSON`/`severityJSON` branding vs accept                        | 23-40 §f18                      | type-branding decision at the cache scan seam                  |

## Flake / repo-shape conventions

| Item                                                                                | Source          | What it needs                                                       |
| ----------------------------------------------------------------------------------- | --------------- | ------------------------------------------------------------------- |
| `homepage` in the flake meta block                                                  | 23-40 §f21      | a URL decision (docs site does not exist yet)                       |
| `flake-parts` duplicate input → `follows` go-nix-helpers, or silence                | 23-40 §f22      | input-graph decision                                                |
| structure-linter layout items (`assets/`, root `internal/`, README install section) | 23-40 §f23      | accept-with-reason record; README install section is the cheap half |
| minimal `.buildflow.yml` (documented skips) vs permanent auto-detection             | 23-40 §f24, §g3 | canonical-gate policy call                                          |

## Environment hygiene (owner machine, BuildFlow-owned)

| Item                                                                                 | Source     |
| ------------------------------------------------------------------------------------ | ---------- |
| `VACUUM` the 1.04 GB BuildFlow state DB                                              | 23-40 §f25 |
| Delete the expendable 0.51 GB cache DB                                               | 23-40 §f26 |
| Rebuild the BuildFlow binary from HEAD; verify PATH precedence (BuildFlow#81)        | 23-40 §f27 |
| `buildflow doctor` after the reinstall                                               | 23-40 §f28 |
| `buildflow format` once (stale formatter fingerprint)                                | 23-40 §f29 |
| Re-run `buildflow -s lychee` + `-s vulnix` (both failed on the old hash)             | 23-40 §f4  |
| On-demand scans once the devShell builds: gitleaks, codespell, markdown-lint         | 23-40 §f40 |
| Confirm lychee's 4 path-excluded files are intentional                               | 23-40 §f41 |
| Full `buildflow --build-mode=full` re-run: target 114/114, 0 error-severity findings | 23-40 §f32 |

## BuildFlow upstream candidates (verify-before-filing, then github-voice; needs the upstream mandate from ROADMAP)

| Candidate                                                                                          | Source     |
| -------------------------------------------------------------------------------------------------- | ---------- |
| Root-module go-auto-upgrade scan including `hub/` files (scan-boundary gap)                        | 23-40 §f42 |
| Resume-checkpoint SQLITE_BUSY resilience (lost timings, unusable `--resume` under lock contention) | 23-40 §f43 |
| structure-linter FP feedback path (fixture script in `scripts/`, `docs/assets` as `assets/`)       | 23-40 §f44 |

## Fleet-layer (outside this repo)

| Item                                                                                | Source                                               | What it needs                                                           |
| ----------------------------------------------------------------------------------- | ---------------------------------------------------- | ----------------------------------------------------------------------- |
| docs-gates fleet pin bump past `e305834` (adds the `rows` verb the skill documents) | AGENTS annotate-writer errata, T9 verdict 2026-11-20 | crush-config / fleet nix pin bump; in-repo writer aliases cover the gap |

## Still-gated from the CI-activation report (00-04 §g)

| Question                                                                   | State                                                                                                                                                                                                                              |
| -------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| g2: bench baseline re-record policy                                        | VERIFY done 2026-11-20 — claim PROVEN on the current tree (fold @341 = 107µs vs the ≈150µs FEATURES claim; alloc shapes identical). `benchmarks/baseline.txt` stays M27-era; re-record is the owner's call and otherwise optional. |
| g3: FLAKE_PAT creation (GitHub + Forgejo, read-only Contents on go-daemon) | OPEN — un-skips the CI hub/flake jobs (currently green-but-skipping).                                                                                                                                                              |

~~1. T2/CHANGELOG-cut timing~~ done — resolved by the owner's 2026-11-20 blanket
execution order; the `[0.1.0]` cut landed same-day at the 77881b3 boundary.

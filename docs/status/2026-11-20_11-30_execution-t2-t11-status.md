# Execution Status — Corpus-Drain Plan T2–T11 (CHANGELOG cut, sort-by, cache specs, gate hardening, corpus drain, cmdguard readiness)

**Date:** 2026-11-20 ~11:30 CET
**Repo:** go-compliance @ master (auto-commit daemon owns commits)
**Plan:** `docs/planning/2026-11-19_21-25_corpus-drain-ci-activation-pareto-plan.md` — packages T2–T11 executed end-to-end (T1 closed 2026-11-20 00:04; T21 + OG stay owner-gated)

## a) FULLY DONE

1. **T2 — CHANGELOG `[0.1.0]` cut** at the 77881b3 boundary: pre-tag `[Unreleased]` moved verbatim into `## [0.1.0] - 2026-11-19` (Added/Changed/Fixed/Security), post-tag work stays unreleased (second Added block + the CI-activated/third-pass Changed rows, whitespace normalized). docs-gates 5/5 after.
2. **T3 — bench verify**: claim PROVEN on the current tree. fold @341 = 107µs (M27 baseline 127µs; FEATURES claims ≈150µs → 40% headroom); alloc shapes identical everywhere (fold 62/359/1025 B·allocs, decode+fold+derive 4778); derive zero-alloc intact. `benchmarks/baseline.txt` untouched — re-record is the owner's g2 call (TODO row rewritten to BLOCKED-on-g2).
3. **T4 — hub `table --sort-by`**: findings/failures/name retiebreak INSIDE the state rank, mirroring the producer (importance = producer-only, rejected at the flag boundary). Key design fact discovered by test failure: the v1 wire cannot carry severities, so the findings tiebreak sorts on the wire-portable `GateFindings` totals (`findingTotalsByRepo`), not severity cells. Default `failures` keeps output byte-identical. 5 CLI specs (default-order regression pin, findings/name retiebreak, styled+machine parity, values-tag rejection) + parse DescribeTable. README gained the "Fleet table" section.
4. **T5 — cache error-path specs**: dead-store query + write family-pinned as IO/Transient (exit-75 class) with context and driver cause surviving the wrap; empty-fleet contract pinned; `wrapIterateErr` unit DescribeTable (internal). Migration-idempotence was already pinned by the reopen spec.
5. **T6 — citation-gate hardening**: `extract_cites` prefers rg, falls back to `grep -oE` (ERE-compatible regex by design); shellcheck clean; fallback verified BYTE-IDENTICAL to the rg path across the full 414-cite corpus (30 accepted, 141 presumed lib-side, exit 0) in an rg-less PATH. No CI change needed — the script is now self-sufficient.
6. **T7 — embedder-docs sweep**: `docs/guides/integrate.md` gained the five-checks table (`evidence`, `cache-disk`, `refresh-freshness`, `discovery`, `sse-backpressure`) + the four probe routes + warn-stays-200 semantics; names verified against `cli/health.go`; reporter example rebuilt and executed.
7. **T8 — corpus drain**: 68 rows annotated via the writer (dry-run → apply): 23-09's 19 WP rows re-verified against today's tree; 23-40's 8-gate matrix RE-RUN green today (build/vet/`-race`/lint/`config verify`/templ 0-diff/dupl/`nix flake check`/tidy zero-diff) + §f struck done/routed/NOT-DO; `docs/status/owner-decisions-digest.md` minted (policy verdicts, flake conventions, env hygiene, BuildFlow upstream candidates, fleet pin bump — all source-cited); 23-40 + 00-53 archived (manifest updated; row gates green before AND after dprint). Unarchived reports 6 → 4 (+digest). AGENTS gained the canonical-gate paragraph (plain `buildflow` is the gate; concurrent full runs SQLITE_BUSY-storm — fixes 23-40 §f35/f36 on sight).
8. **T9 — docs-gates surface verify**: installed fleet pin `e305834` lacks ONLY the `rows` verb — its freshness surface is COMPLETE (`--pre-flight`, `--margins`, `--dump-items`, `--explain-item`, `--audit-all`, `--audit-max`, `--replace`, `--duplicates`, `--land-parity`). Verdict recorded in AGENTS (the on-disk file did NOT carry the errata — the session-context copy was stale, disk wins); fleet pin bump routed to the digest.
9. **T10 — cmdguard readiness**: latest tag `v5.1.0`; `AssertDescribeContract` verified present at `v5.1.0-33` (UNPUBLISHED — 33 commits ahead), adoption sketch written patch-ready (`docs/notes/cmdguard-describe-adoption-sketch.md`, internal-test attachment at `buildApp`). Both upstream claims re-verified at source before drafting (`types_duration.go:14` raw `time.ParseDuration` passthrough; `lint/go.mod:81-87` four `=> ../` replaces) and both DRAFTS written + voice-checker-clean (`docs/drafts/cmdguard-nd-duration-codec.md`, `cmdguard-lint-published-replaces.md`). Filing stays gated on OG.5.
10. **T11 — reconcile + battery**: TODO_LIST 8 completed rows removed (logged in CHANGELOG), bench row rewritten to its owner-gated remainder, cmdguard row updated with draft pointers, 2 rows minted (README install section; fleet docs-gates bump BLOCKED). CHANGELOG `[Unreleased]` gained 5 Added + 3 Changed entries carrying the harvest tokens — freshness gate green again after it correctly flagged the unharvested T4/T6 §f items mid-reconcile.

## b) PARTIALLY DONE

1. **T3 re-record** — verify done, claim proven; the re-record itself waits on g2 (owner). The stale baseline is now KNOWN-stale with evidence, not unknown-stale.
2. **T21 (hub replace-drop)** — unchanged, gated on the v0.1.1 go-signal (OG.8). Nothing this session changed its prerequisites.

## c) NOT STARTED (owner-gated, unchanged)

FLAKE_PAT (g3), Forgejo visibility, M26 endgame, VACUUM + cache-DB hygiene, firefox permission, upstream filing mandate (drafts now READY behind it).

## d) Battery (final, sequential, all green)

core build+vet+`-race` (1 ok) · hub build+vet+`-race` (6 ok) · golangci 0+0 · templ fmt -fail (changed=0) · art-dupl 0 new (3 baseline) · actionlint clean · dprint full-repo clean · docs-gates 5/5 · `nix flake check` all-passed (16-12 §a2 shape) · citation-gate exit 0.

## e) Design notes worth keeping

1. **Wire-contract constraint shapes the sort**: `--sort-by findings` cannot use severity cells (file-sourced evidence has none — the contract can't carry them). Sorting on `GateFindings` totals is the only source-uniform truth; documented at `findingTotalsByRepo`.
2. **The freshness gate works**: it flagged exactly the two §f items this session completed before any living doc carried the proof. Harvest-after-execute is the correct order — the gate caught the gap within minutes.
3. **Stale-context trap hit twice** (AGENTS annotate-writer section; dprint CHANGELOG exclusion) — project-context copies lag the disk; re-read before editing.

## f) Next

1. Owner answers: g2 (bench re-record — now optional, claim proven), g3 (FLAKE_PAT), OG.8 (v0.1.1 go-signal → T21), OG.5 (upstream mandate → the two drafts are ready to file).
2. Agent-executable remainders: README install section (new TODO row); everything else drains to `docs/status/owner-decisions-digest.md`.

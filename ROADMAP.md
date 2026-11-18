# Roadmap

> Long-term direction and raw ideas not yet refined into actionable
> TODO_LIST tasks. When an item becomes actionable, it graduates to
> TODO_LIST and loses its row here.

## Adoption triggers (from the 2026-11-17 stack-selection scaffold)

- **Tag consumers when adopted** — once real repos consume
  go-compliance (core or hub), revisit the module split and the v1
  contract's importance ranking against actual usage; adjust before
  v2 anything.
- **Hub v2 wire** — `docs/decisions/v2-wire-design.md` sketches v2
  (v1 stays emittable forever). Build only when v1 has been solid in
  production for a full adoption cycle.

## Federation (from docs/notes/federation-interop.md, 2026-11-18)

Default posture is passive (A): go-health-dashboard scrapes our
`/health`-equivalent surface; stale-must-not-page holds. Build further
only on a demand signal:

- **B** — source-adapter in go-health-dashboard reading our JSON
  (gated on that repo's maintainers wanting it).
- **C** — link-level federation (hub-to-hub evidence exchange).
- **D** — library convergence (go-compliance core as the shared
  evidence model inside go-health).

## Deduplication and tooling ideas (2026-11-18 art-dupl audit)

Raw ideas harvested from `docs/status/2026-11-18_17-49_art-dupl-dedup-audit-status.md` §f —
not yet refined into tasks:

- art-dupl calibration: a `--show-suppressed` pass to judge what the
  actionability filter hides, hash-based detection (`-m hash`) as an audit
  complement, and trend-tracking total clone tokens per audit.
- Data-model candidate: a `WorstStepLimit` branded type with a defaulting
  constructor (making invalid caps unrepresentable) — weigh public-API churn
  before touching the v1-adjacent surface.
- templ-components upstream candidate: an attribute-only "tickable"
  RelativeTime mode so consumers stop hand-rolling
  `templ.Attributes{"data-tc-relative": true}` (2026-11-18 final pass).
- Acceptance policy DECIDED (2026-11-18 execution pass): BOTH mechanisms,
  layered — per-site `// art-dupl:accept <why>` directives stay the
  in-source record of intentional clones, AND a committed
  `.art-dupl-baseline.json` + `art-dupl -t 1 --type-aware check` CI gate
  (job `dupl-gate`, tool pinned v0.9.1-0.20261017225421-93432fd910fb)
  ratchets out hash-known groups so only NEW clones can ever surface.
  Regenerating the baseline is a deliberate act, never a way to silence
  a review finding.

## Endgame question (owner, g2 — BLOCKED)

M26: does go-compliance replace the overview system or coexist?
Answer shapes the migration note and whether an overview-compat
decoder stays a supported surface. See TODO_LIST blocked row.

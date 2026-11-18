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

## Endgame question (owner, g2 — BLOCKED)

M26: does go-compliance replace the overview system or coexist?
Answer shapes the migration note and whether an overview-compat
decoder stays a supported surface. See TODO_LIST blocked row.

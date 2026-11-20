# Dedup Audit Policy (art-dupl)

Relocated from AGENTS.md 2026-11-20 (canonical-gate line-count rule); content unchanged.

- Audit with `art-dupl -t 1 --type-aware`; target "0 shown". Accept
  leftover intentional clones with `// art-dupl:accept <why>` — gofmt
  normalizes the directive to the spaced form and art-dupl still honors
  it, but it must sit within 5 lines ABOVE the clone start (function-doc
  distance is too far). Every suppression should be pinned by a directive
  so future audits only ever litigate NEW clones (2026-11-18 audit).
- CI gate (2026-11-18): `.art-dupl-baseline.json` is committed and
  `dupl-gate` runs `art-dupl -t 1 --type-aware check --baseline-path
  .art-dupl-baseline.json .` (tool pinned v0.9.1-0.20261017225421). The
  baseline ratchets known hashes; only NEW clones fail CI. Accept
  directives document WHY a clone is intentional; the baseline records
  THAT it exists. After a deliberate refactor that grows duplication
  knowingly, regenerating the baseline is a decision, not a fix.
- Single homes for shared policies: SQL query + IO-error wrap =
  `cache.Store.queryRows` + `wrapIterateErr` (hub/internal/cache); the
  worst-steps defaulting rule = `worstStepLimitOrDefault` (fold.go);
  dashboard filter validation delegates to `State.IsValid` so new core
  states are filterable without touching routes.go.
- Re-open triggers: a 4th cache query site (extract the scan-callback
  collector then); accept directives are line-proximity fragile (5-line
  window) — expect drift after refactors near them, re-audit on sight.
- One `// art-dupl:accept` directive suppresses its WHOLE clone group
  (both occurrences — observed empirically on the refresher group;
  mechanism undocumented upstream, don't rely on per-occurrence
  behavior).

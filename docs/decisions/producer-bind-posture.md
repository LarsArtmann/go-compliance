# Decision: BuildFlow binds the core

Date: 2026-11-19
Status: decided and executed

## Context

BuildFlow was the origin of this library's contract: its
`compliance --json` output IS the frozen v1 wire document, and the
extraction (2026-11-18) left BuildFlow as producer-only with ZERO imports
- compat held by dual golden tests (its fixture decodes real producer
output). The open question was whether BuildFlow would ever IMPORT the
dep-light core (state model, wire shapes, derive/marshal) or stay
permanently decoupled with the goldens as the only tie.

The deep-dive audit of 2026-11-19 scored the posture 33/100 precisely
because of that gap: a ~180-line twin of a frozen-forever contract,
re-verified byte-equal on every change instead of shared.

## Decision

**Bind.** Owner verdict 2026-11-19: BuildFlow imports the core at the
first tag, deletes the twin, and keeps only host-specific concerns local
(SQL evidence fold, missing-on-disk join, table enrichment).

Executed same day:

1. `v0.1.0` tagged (annotated, core module) + pushed to both remotes;
   proxy-fetchable (`go list -m github.com/larsartmann/go-compliance@v0.1.0`
   resolves through proxy.golang.org).
2. BuildFlow branch `xp2-bind-core`: root require `v0.1.0`, adapter in
   `internal/cli/compliance_view.go` (dbstore stays library-free by
   design), wire structs + state consts + `DeriveComplianceState` +
   worst-first sort deleted in favor of `V1Report` / `BuildV1Report` /
   `MarshalV1` / `State.Rank`.
3. Gates: both golden suites unchanged (BuildFlow's golden compares
   decoded trees - MarshalV1's sorted map keys are the documented
   canonical subset), `dryrun-repro.sh`, decode-roundtrip test added
   (producer stdout -> `DecodeV1Report` -> re-marshal tree-equal),
   mutation-verified verdict mapping.

## Consequences

- ONE implementation of the contract; schema changes need a conscious
  v2 bump here and a matching consumer move - no third copy can drift.
- BuildFlow's human table keeps its extra columns (severities, last-run
  version, project-meta importance) as local view enrichment; the wire
  stays frozen v1.
- The `never-covered` state remains consumer-side only (discovery); the
  producer path cannot emit it (documented in `DeriveState`).
- Adoption trigger from ROADMAP: revisit the module split now that a
  real consumer exists.

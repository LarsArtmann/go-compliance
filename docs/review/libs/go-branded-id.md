# go-branded-id

| Field            | Value                                                                                              |
| ---------------- | -------------------------------------------------------------------------------------------------- |
| Reviewed         | 2026-11-19 (T15 completion report — closes the audit's dep-graph coverage gap)                     |
| Group            | T15 (post-audit completion; sole-author verification, same contract as the 5-group fleet run)      |
| Consumer modules | core — **direct** (`go.mod` dep 1 of 2); hub — indirect via core                                    |
| Library version  | `v0.7.1` (pinned; equals latest tag — current)                                                     |
| Recent tags      | `v0.7.1`, `v0.7.0`, `v0.6.0`, …                                                                    |
| Modules          | single module `github.com/larsartmann/go-branded-id` (generic `ID[B, V]` + codecs, no sub-modules) |

## What the consumer uses (verified)

- **Two brands, two aliases, two constructors** — the entire domain-ID layer
  is `branded.go` (30 lines): `RepoBrand`/`SourceBrand` with `Name()`
  ("Repo:go-finding" rendering), `RepoName = id.ID[RepoBrand, string]`,
  `SourceName = id.ID[SourceBrand, string]`, `NewRepoName`/`NewSourceName`
  (`branded.go:14-40`).
- **The JSON codecs ARE the frozen v1 wire contract**: `MarshalV1` /
  `DecodeV1Report` serialize `RepoName` through the lib's JSON v1/v2
  implementations (`id_json_v1.go` / `id_json_v2.go`), including the
  documented zero-value-means-unset → `null` behavior the wire depends on
  (`branded.go:22-24`, pinned by the byte-exact golden fixture test).
- **Type-level separation actually earns its keep**: `WriteEvidence(ctx,
  repo RepoName, …)` (`reporter.go:48`, `filestore.go:40`) cannot be called
  with a source name; the hub's `OnWrite` hook carries `compliance.RepoName`
  across the module boundary (`hub/internal/cache/cache.go:61,90,100`).
- **String rendering in logs/errors** via `String()` (brand-prefixed), used
  throughout slog calls in the hub.

## What the consumer does not use (and why that is fine)

- `ValidateID` / `MustValidateID`: the frozen v1 wire tolerates the zero
  value by design (redacted/unknown repos in producer output); boundary
  validation would tighten `DecodeV1Report` against empty repo names, but
  that is a v2-wire decision, not a retrofit (see
  `docs/decisions/v2-wire-design.md`).
- `Compare` / `Or` / `FromPtr`: rows sort by derived state ranking, not ID
  order; no optional-ID fields exist.
- gob / binary / SQL / text codecs: the hub cache stores repo names as
  `TEXT` via its own scan (`queries.go:95,149` re-brands on read); the SQL
  `Valuer`/`Scanner` would only matter if the cache ever bound parameters
  as `RepoName` directly — it binds strings today.

## Verdict — **ADOPTED**

Utilization score **90** (4.5 of 5 weighted rows: the codec-and-brand core
is the product's type spine; the unused remainder is genuinely unneeded).
The library is doing exactly what a dep-light core wants: zero transitive
dependencies, one generic type, and codecs the wire contract leans on.

**Flip triggers:** a v2 wire that rejects empty/unknown repo names (adopt
`ValidateID` at the decode boundary); the cache binding `RepoName`
parameters directly (SQL codec becomes load-bearing); fleet-wide integer
IDs (out of scope for a repo-name domain).

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

- Core module: compliance state model (`compliant`/`failing`/`stale`/`never-covered`), branded types (`RepoName`, `SourceName`), and sentinel errors classified through go-error-family.
- Core module: frozen v1 wire contract — marshal/decode byte-exact against a real BuildFlow `compliance --json` fixture (golden tests); `schema_version >= 2` rejected like the overview decoder.
- Core module: source-agnostic evidence fold with `worstStepLimit` capping.
- Core module: `Source` adapter interface, `FileSource` (piped v1 document with missing-file TTL backoff), `CommandSource` (spawns `buildflow compliance --json`).
- Core module: reporter SDK one-liner `compliance.Record` with memory and file writers.
- Hub module: cmdguard CLI with `serve`, `report`, `table` commands and BuildFlow-parity flags (`--since`, `--stale-after`, `--worst-steps`, `--step`, `--source file|buildflow|cache`).
- Hub module: go-output fleet table (ERR/WARN/INFO severity cells, state labels, worst-first sort).
- Hub module: SQLite evidence cache via go-sqlitestore with versioned migrations and window queries.
- Hub module: discovery daemon client over a unix socket (`/v1/discover`) with never-covered derivation.
- Hub module: SSE events hub (`/events`) with Last-Event-ID replay, ring-buffer store, and heartbeats (go-sse).
- Hub module: realtime dashboard — templ-components page, stat cards, state filter chips, DataStar patch stream with three-way row diffing (go-datastar), embedded SDK bundle.
- Hub module: content negotiation (HTML for browsers, JSON for clients) and `?state=` filtering on the page route.
- Hub module: go-health probe surface (`/healthz` `/readyz` `/startupz` `/version`) with evidence-published readiness check.
- Hub module: sd_notify readiness/stopping/watchdog via go-daemon; hardened systemd unit under `deploy/`.
- Hub module: bind posture — localhost-only default, `unix:` socket mode (0o660, stale-socket removal, cleanup on shutdown), non-localhost warning.
- Repo hygiene: dprint gate, TODO_LIST, FEATURES, DOMAIN_LANGUAGE, planning + status docs.
- Nix: two-module flake (devShell with pinned templ generator + dprint + actionlint; `compliance-hub` package; core build+test check) — `nix flake check` green. The hub vendorHash isolates a Go-files-only core copy (`_core_only`): `go mod vendor` copies directory-replaced modules verbatim, so vendoring the repo root would feed flake.nix (which carries the vendorHash) into the FOD — an unfixable self-reference flip-flop, root-caused by diffing two stored go-modules outputs.
- CI: `.github/workflows/ci.yml` — per-module build+test and lint (golangci-lint pinned to the devShell 2.15.0), dprint + actionlint, stack-drift-check, docs-gates, `nix flake check`; hub/flake jobs skip (not fail) until the `FLAKE_PAT` secret exists. Valid on GitHub and Forgejo runners.
- Lint: `.golangci.yml` (fleet-trimmed); both modules at 0 issues with tests green.
- Docs: `examples/reporter` (runnable one-line integration demo), `docs/guides/integrate.md`, README screenshots (terminal table + live dashboard), `docs/decisions/v2-wire-design.md` (compat policy: v1 emittable forever), `docs/notes/federation-interop.md` (hub ↔ go-health-dashboard options), STACK.md errata (templ pin; hub's private go-daemon consequence).
- Hub module: live re-derive (M27) — `refresher` (trigger-coalescing channel, 500ms debounce, 5-minute safety-net ticker) keeps served rows live: in-process store writes fire an `OnWrite` hook that kicks a re-query → `CompleteWithDiscovery` → `PublishRefresh`; a discovery `/v1/watch` subscription completes never-covered repos on fleet changes and re-kicks (30s backoff on daemon-down, 1s re-subscribe on stream end). `/v1/compliance` serves the latest evidence snapshot (503 before the first successful refresh). The store's latest-records query is one correlated-subquery statement (the per-repo follow-up loop was an N+1). E2E contract test: a store write mid-stream reaches the browser as an append patch on the rows region.
- Benchmarks: fold/derive/full-pipeline at 50/341/1000 repos (`bench_test.go`); baseline recorded under `benchmarks/` (fold @341 ≈ 150µs, derive @341 ≈ 2.3µs zero-alloc, decode+fold+derive @341 ≈ 0.9ms).

### Changed

### Deprecated

### Removed

### Fixed

### Security

- Bind posture decision enforced in code: loopback default, explicit opt-in for wider binds (`docs/decisions/bind-posture.md`).

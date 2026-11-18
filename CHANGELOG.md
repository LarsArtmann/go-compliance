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

### Changed

### Deprecated

### Removed

### Fixed

### Security

- Bind posture decision enforced in code: loopback default, explicit opt-in for wider binds (`docs/decisions/bind-posture.md`).

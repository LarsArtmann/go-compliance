# AGENTS.md - go-compliance

## Project Overview

**go-compliance** is a fleet compliance hub: a frozen v1 wire contract
(byte-compatible with `buildflow compliance --json`), a source-agnostic
evidence fold + state model (`compliant` / `failing` / `stale` /
`never-covered`), and a realtime dashboard that projects evidence into the
browser via SSE + DataStar patches. Two modules: **core** (root, public
deps only — the publishable library) and **hub** (`hub/`, the server +
CLI). Coordinates: `go.work`; per-module work uses `GOWORK=off`.

## Commands

```bash
# Per module (ALWAYS GOWORK=off; the workspace hides replace-directive drift)
cd hub && GOWORK=off go build ./... && GOWORK=off go vet ./...
cd hub && GOWORK=off go test -race -count=1 ./...
golangci-lint run ./...                       # core (root .golangci.yml)
cd hub && golangci-lint run -c ../.golangci.yml ./...
nix flake check                                # both modules, sandboxed
nix build .#compliance-hub                     # the server binary
go tool templ generate ./internal/dashboard/   # after ANY .templ edit
```

Run `go test` invocations SEQUENTIALLY on this machine (shared
`/mnt/buildcache` GOCACHE; concurrent runs produce transient
`[build failed]`). The auto-commit daemon commits continuously — never
author commits manually, never push unless asked; expect "file modified
since read" edit races (re-read and retry, or use asserted python
replaces).

## Module Rules

- **Core keeps zero private deps.** Hub's only private dep is go-daemon
  (sd_notify), wired through `_local_deps` + the flake's `postPatchExtra`
  replace. New private deps need a STACK.md decision-log entry.
- **One root `.golangci.yml` for both modules** (v2 has no `extends`);
  hub runs with `-c ../.golangci.yml`. Wrapcheck keys are v2 kebab-case
  (`ignore-package-globs`, `extra-ignore-sigs`): old camelCase keys are
  SILENTLY DROPPED by `run` — only `golangci-lint config verify` catches
  the loss. Policy: first-party LarsArtmann globs are already-wrapped
  boundaries; stdlib boundaries get real wraps or nolint-with-reason.
- **templ**: the generator is pinned (v0.3.1100, tool directive);
  generated `*_templ.go` files ARE committed (flake + CI consume them
  without a generate step — see .gitignore note). After editing any
  `.templ`, run `go tool templ generate` and commit the `_templ.go` diff
  together. The `templ` package identifier is implicitly available in
  .templ files — do NOT import `github.com/a-h/templ` explicitly in a
  .templ import block (the generated file redeclares it).

## Testing Conventions

- ginkgo/gomega; one suite bootstrap per package
  (`xxx_suite_test.go` with `RunSpecs`).
- SSE/DataStar handlers are tested with `datastartest`
  (`CollectWithTimeout(GinkgoTB(), handler, timeout, WithPath(...))`),
  NEVER by polling an `httptest.ResponseRecorder` from another goroutine
  — the recorder is unsynchronized and `-race` flags it (fixed
  2026-11-18). A real request lifecycle is the only honest shape.
- Cross-package test helpers live in `hub/internal/testsupport`
  (external `_test` packages cannot share `_test.go` code);
  `testsupport.WaitForSubscriber(hub)` before publishing to a fresh SSE
  subscriber (snapshot delivery happens before feed subscription).
- The M27 E2E contract test (`hub/internal/cli/refresher_internal_test.go`)
  pins: store write → OnWrite hook → Kick → debounce → refresh →
  PublishRefresh → append patch on `#fleet-rows`.

## Live Architecture (M27)

`refresher` (hub/internal/cli/refresher.go) owns liveness: trigger
channel (cap 1, coalescing), 500ms debounce, 5-minute safety-net ticker
(covers cross-process SQLite writes — the OnWrite hook only sees THIS
process's writes; documented limitation), discovery `/v1/watch` with 30s
backoff. `/v1/compliance` serves the refresher's snapshot (503 before the
first successful refresh).

Dashboard patches: rows append/replace/remove by diff
(`events.DiffRows`); the page-level `rendered X ago` RelativeTime is the
ONE AutoRefresh carrier (ships the 30s tick script); row cells opt into
ticking via the bare `data-tc-relative` attribute WITHOUT per-row
scripts (RelativeTimeProps couples AutoRefresh to a per-instance
`<script>` — rows re-render inside every patch; per-row scripts would
bloat patches and accumulate in the tbody). Pinned by a view_test.go
regression test.

## Nix / FOD Lessons

- **The vendorHash FOD self-reference flip-flop** (root-caused
  2026-11-17): `go mod vendor` copies directory-replaced modules
  verbatim, so hub's `replace go-compliance => ../` vendored the repo
  root INCLUDING flake.nix — the FOD hashed its own placeholder. Fix:
  `postPatchExtra` builds `_core_only` (Go files + go.mod/go.sum only)
  and repoints the hub replace at it. Consequence: doc-only edits no
  longer change the hub vendorHash.
- go.mod/go.sum bumps invalidate `vendorHash`: tidy per module FIRST,
  then bootstrap (`vendorHash = "sha256-AAA..."` → build → paste the
  `got:` hash → rebuild to confirm convergence) and `nix flake check`.
- CI hub/flake jobs need the `FLAKE_PAT` secret (private go-daemon);
  missing secret SKIPS those jobs (step-level `if: env.FLAKE_PAT != ''`
  — job-level env-if is invalid per actionlint).

## Fixture (frozen v1 contract)

`testdata/buildflow-v1-fixture.json` pins 341 redacted repos of real
producer output; `scripts/refresh-fixture.sh` header documents the full
refresh ritual (golden byte-exactness, the @341 benchmark shape,
baseline re-recording).

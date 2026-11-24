# AGENTS.md - go-compliance

## Project Overview

**go-compliance** is a fleet compliance hub: a frozen v1 wire contract
(byte-compatible with `buildflow compliance --json`), a source-agnostic
evidence fold + state model (`compliant` / `failing` / `stale` /
`never-covered`), and a realtime dashboard projecting evidence into the
browser via SSE + DataStar patches. Two modules: **core** (root, public
deps only) and **hub** (server + CLI). Coordinates: `go.work`; per-module
work uses `GOWORK=off`.

## Commands

```bash
# Per module (ALWAYS GOWORK=off — module-isolated, consumer-parity gates)
cd hub && GOWORK=off go build ./... && GOWORK=off go vet ./...
cd hub && GOWORK=off go test -race -count=1 ./...
GOWORK=off go test -race -count=1 ./...      # core: -race parity with hub (CI runs both)
golangci-lint run ./...                       # core (root .golangci.yml)
cd hub && golangci-lint run -c ../.golangci.yml ./...
cd hub && GOWORK=off go tool templ fmt -fail ./internal/dashboard/   # templ formatting gate (CI-enforced; no -check flag exists)
nix flake check                                # both modules, sandboxed
nix build .#compliance-hub                     # the server binary
nix develop -c bash -c 'cd hub && GOWORK=off go build ./...'   # devShell hub build (the CI parity path)
compliance-hub report --db /nonexistent/x.db 2>/dev/null; echo $?   # exit-code contract probe: expect 75 (transient IO); always redirect stderr + echo $? when probing exit codes; NEVER read $? after a pipe — the pipe's last command masks the real exit (3 strikes 2026-11-20: citation gate, brandid gate, CI watch all misread green)
cd hub && go tool templ generate ./internal/dashboard/   # after ANY .templ edit; MUST run from hub/ (tool directive lives in hub/go.mod) — from the root it silently rewrites the embedded templ.Error FileName paths (hub/ prefix churn) instead of failing
scripts/pre-release-check.sh vX.Y.Z   # release pre-flight (clean tree, CHANGELOG, notes file, BOTH remotes pushed, tag fresh, canonical buildflow, flake check, docs-gates — true exits); --verify re-checks post-release; sequence: docs/notes/release-runbook.md (script needs #!/usr/bin/env bash — /bin/bash does not exist on NixOS)
```

Run `go test` invocations SEQUENTIALLY on this machine (shared
`/mnt/buildcache` GOCACHE; concurrent runs produce transient
`[build failed]`). The auto-commit daemon commits continuously — never
author commits manually, never push unless asked; expect "file modified
since read" edit races (re-read and retry, or use asserted python
replaces).

The canonical gate is plain `buildflow` (`.buildflow.yml` exists solely to
opt into brandid-lint BD003 via `tool_options` — everything else is
auto-detected); manual per-tool runs are drill-downs, never
replacements (a manual-green sweep missed 79 findings + 2 failed steps in
the 2026-11-18 BuildFlow full run). Never overlap concurrent full
BuildFlow runs: the second run hits SQLITE_BUSY on every checkpoint write
(lost timings, unusable `--resume`).

## Module Rules

- **Branded IDs (brandid-lint BD003, adopted 2026-11-20)**: identifier-typed
  fields/params use strong types — `RepoName`/`SourceName` (core) and
  `sse.EventID` for `events.Published.ID` / `Publish*` returns (branded in
  go-sse; construct via `sse.NewEventID`, unwrap with `.Get()` for display).
  ssetest records the wire `id:` as a plain string, so SSE-ID test
  expectations compare strings. Deliberate primitive-ID boundaries carry a
  reasoned `//brandid-lint:ignore(BD003)` directive. Gate runs via BuildFlow
  (`tool_options` in `.buildflow.yml`) and the CI `branded-ID gate` job
  (pinned `linter/v0.5.1` — bump manually on new linter tags).
- **Core keeps zero private deps.** Hub's only private dep is go-daemon
  (sd_notify), wired through `_local_deps` + the flake's `postPatchExtra`
  replace. New private deps need a STACK.md decision-log entry.
- **One root `.golangci.yml` for both modules** (v2 has no `extends`);
  hub runs with `-c ../.golangci.yml`. Wrapcheck keys are v2 kebab-case
  (`ignore-package-globs`, `extra-ignore-sigs`): old camelCase keys are
  SILENTLY DROPPED by `run` — only `golangci-lint config verify` catches
  the loss. Policy: first-party LarsArtmann globs are already-wrapped
  boundaries; stdlib boundaries get real wraps or nolint-with-reason.
  Both modules sit at 0 findings (verdict g3, 2026-11-18): real fixes +
  nolint-with-reason for intentional globals/ports, NO baseline file —
  the 0-finding gate stays honest. `golangci-lint fmt` applies the
  configured goimports/gofumpt/golines; golines wraps directives ON a
  code line, breaking suppression — always put `//nolint` block-form on
  its own line ABOVE the statement.
- **templ**: the generator is pinned (v0.3.1100, tool directive);
  generated `*_templ.go` files ARE committed (flake + CI consume them
  without a generate step — see .gitignore note). After editing any
  `.templ`, run `go tool templ generate` and commit the `_templ.go` diff
  together. The `templ` package identifier is implicitly available in
  .templ files — do NOT import `github.com/a-h/templ` explicitly in a
  .templ import block (the generated file redeclares it). The templ LSP's
  `embedlit` hint is a FALSE POSITIVE for props literals: Go forbids
  elided embedded-field composite literal types (proven 2026-11-19), keep
  the explicit `BaseProps: utils.BaseProps{...}` form.
- **dashboard CSS**: compiled Tailwind artifact, NOT the Play CDN. After
  touching any dashboard `.templ`, adopting a templ-components component,
  or bumping the pin: `scripts/regen-ui.sh` (locked tailwind:
  `nix shell .#tailwindcss`;
  rebuilds `library-classes.txt` from go.mod truth + compiles
  `app.min.css`) and commit the three artifacts together — the CI
  `css-freshness` gate regenerates and diffs byte-fresh. New routes: serve
  links `/app.css?v=<content-hash>` (immutable day cache).

## Dedup Audit (art-dupl)

- Audit `art-dupl -t 1 --type-aware`; accept intentional clones with
  `// art-dupl:accept <why>` within 5 lines ABOVE the clone start. CI
  `dupl-gate` ratchets `.art-dupl-baseline.json` — only NEW clones fail
  (accept directives are the WHY, the baseline is the THAT).
- Full policy (single homes, re-open triggers, whole-group suppression
  semantics): `docs/notes/dedup-audit-policy.md`.

## docs-health annotate writer

- Multi-section report keys MUST use the `N@substring` disambiguator;
  always `emit-keys` + `--dry-run` first (a repair cycle burned 2026-11-19).
- Writer verb inventory + the `docs-gates` fleet-pin surface:
  `docs/notes/docs-health-annotate-writer.md`. Battery stays
  `docs-gates run .` (5 gates).

## Testing Conventions

- ginkgo/gomega; one suite bootstrap per package
  (`xxx_suite_test.go` with `RunSpecs`).
- SSE/DataStar handlers are tested with `datastartest`
  (`CollectWithTimeout(GinkgoTB(), handler, timeout, WithPath(...))`),
  NEVER by polling an `httptest.ResponseRecorder` from another goroutine
  — the recorder is unsynchronized and `-race` flags it (fixed
  2026-11-18). A real request lifecycle is the only honest shape.
- Wall-clock specs (backoff/TTL/debounce): work INSIDE the asserted
  window stays microsecond-scale (never fixture IO — a 170KB write flaked
  a race run); sleep ≥2× the TTL for expiry; plant a CORRUPT document
  mid-window so broken backoff fails with ErrParse, not a silent pass.
- Cross-package test helpers live in `hub/internal/testsupport`
  (external `_test` packages cannot share `_test.go` code);
  `testsupport.WaitForSubscriber(hub)` before publishing to a fresh SSE
  subscriber (snapshot delivery happens before feed subscription).
- The M27 E2E contract test (`hub/internal/cli/refresher_internal_test.go`)
  pins: store write → OnWrite hook → Kick → debounce → refresh →
  PublishRefresh → append patch on `#fleet-rows`.

## Library Utilization Audit (2026-11-19)

- `docs/review/libs/` = 22 reports + index; verdicts and re-review
  triggers (flip trigger, major bump, new fleet consumers) live in the
  reports; details: `docs/notes/library-audit-notes.md`.
- Citation integrity: `scripts/check-citations.sh` gate — rerun after
  ANY refactor that moves cited code (ratchet:
  `.citations-baseline`; new drift fails, baselined prints ACCEPTED).
- Tag-watch: a reviewed lib's new tag → executed-addendum to its report
  next touch (verdicts pin versions, they stale silently otherwise).

## Live Architecture (M27)

`refresher` (hub/internal/cli/refresher.go) owns liveness: trigger
channel (cap 1, coalescing), 500ms debounce, 5-minute safety-net ticker
(backstop), discovery `/v1/watch` with 30s backoff, and — since 2026-11-19 —
a go-filewatcher/v2 dir-watch (cli/watcher.go) over the cache DB + WAL
sidecars (and the evidence file's dir for `--source file`): cross-process
SQLite writes and atomic evidence renames refresh in ~1-2s instead of
waiting for the safety net; `evidence-watch` is its non-critical health
check. `/v1/compliance` serves the refresher's snapshot (503 before the
first successful refresh). The watch stream is framed by
`daemon.ParseSSEData` (WHATWG: multi-line data joins, over-cap lines fail
loudly; discovery/client.go), and unix binds are probe-first (bind.go
`removeUnownedSocket` + `daemon.IsSocketAlive`): a second hub NEVER steals
a live sibling's socket — it fails with "refusing to steal".

Boot contract (verdict recorded 2026-11-18): `serve` EXITS when the
boot refresh fails — no tolerant boot. A failed first query means the
evidence source itself is broken; systemd `Restart=on-failure` retries
with backoff and the operator gets paged by the service manager instead
of a green process advertising a 503 window. The readiness machinery
(503-until-publish, MarkShuttingDown ordering) remains for embedders
and graceful shutdown.

Health surface: the `evidence` readiness check (cli/health.go) gates on
`hub.Rows() == nil` — `Hub.Rows()` returning nil BEFORE the first publish
and non-nil for an EMPTY published fleet is a load-bearing contract
(fixed 2026-11-18 after the gate proved inert; regression-pinned in
events_test.go + cli/health_internal_test.go; verdict g2: an empty
published fleet IS ready). The non-critical `refresh-freshness` check
(refresher.LastRefreshSuccess, warn threshold 10 min = 2× safety net;
verdict g1: warn stays 200 — stale view is degraded truth, not service
loss) plus the per-refresh `refreshTimeout` (30s) close the former wedge
hole: a hung `QueryEvidence` now fails ONE refresh (logged, retried) and
degrades /readyz to warn instead of freezing green. The non-critical
`discovery` check (same warn semantics, stamp = last watch connect or
snapshot, threshold 10 min) makes a long-dead daemon visible in /readyz.
DI posture:
samber/do is transitive only (cmdguard v5 + go-health); the hub wires
manually by recorded decision (STACK.md) — see
docs/architecture-understanding/2026-11-18_19-39_samber-do-health-architecture.md.

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
  2026-11-17, retired at v0.2.0 when the hub replace was dropped): `go
  mod vendor` copied directory-replaced modules verbatim, so hub's
  `replace go-compliance => ../` vendored the repo root INCLUDING
  flake.nix — the FOD hashed its own placeholder. The `_core_only`
  postPatchExtra fix is gone with the replace; keep the lesson for any
  future directory replace in a vendored module.
- go.mod/go.sum bumps — and ANY root-module Go-file edit (the hub
  vendored core from the proxy since v0.2.0, so core edits no longer
  flip the HUB hash but still flip the CORE check's) — invalidate
  `vendorHash`: tidy per module FIRST, then bootstrap (`vendorHash =
  "sha256-AAA..."` → build → paste the `got:` hash → rebuild to confirm
  convergence) and `nix flake check`. A past run had the hub hash at TWO
  flake sites — `nix build .#compliance-hub` alone can pass while `nix
  flake check` still fails on the second site, so ALWAYS finish with the
  full check. There is another hash site: templ-gen's devShell
  `buildGoModule` (flake.nix) — its `vendorHash` tracks the upstream templ repo's module set and can go stale with NO
  local change (2026-11-18: devShell build failed on a mismatch; the
  same paste-`got:` ritual fixed it, converged first paste).
- CI hub/flake jobs need the `FLAKE_PAT` secret (private go-daemon);
  missing secret SKIPS those jobs (step-level `if: env.FLAKE_PAT != ''`
  — job-level env-if is invalid per actionlint).

## Fixture (frozen v1 contract)

`testdata/buildflow-v1-fixture.json` pins 341 redacted repos of real
producer output; `scripts/refresh-fixture.sh` documents the refresh
ritual (golden byte-exactness, @341 benchmark shape, baseline re-record).

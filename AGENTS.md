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
GOWORK=off go test -race -count=1 ./...      # core: -race parity with hub (CI runs both)
golangci-lint run ./...                       # core (root .golangci.yml)
cd hub && golangci-lint run -c ../.golangci.yml ./...
cd hub && GOWORK=off go tool templ fmt -fail ./internal/dashboard/   # templ formatting gate (CI-enforced; no -check flag exists)
nix flake check                                # both modules, sandboxed
nix build .#compliance-hub                     # the server binary
nix develop -c bash -c 'cd hub && GOWORK=off go build ./...'   # devShell hub build (the CI parity path)
compliance-hub report --db /nonexistent/x.db 2>/dev/null; echo $?   # exit-code contract probe: expect 75 (transient IO); always redirect stderr + echo $? when probing exit codes
cd hub && go tool templ generate ./internal/dashboard/   # after ANY .templ edit; MUST run from hub/ (tool directive lives in hub/go.mod) — from the root it silently rewrites the embedded templ.Error FileName paths (hub/ prefix churn) instead of failing
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
  .templ import block (the generated file redeclares it).

## Dedup Audit (art-dupl)

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

## Library Utilization Audit (2026-11-19)

`docs/review/libs/` holds 20 per-library utilization reports + `index.html`
(fleet run: 5 sub-agent groups, orchestrator QC). Verdict line: go-health /
go-sse / go-error-family(core) ADOPTED; go-filewatcher ADOPTED 2026-11-19
(plan T3, was the audit's NOT-ADOPTED-but-should); templ-components/cmdguard/
go-datastar/go-output/go-daemon UNDERUSED; 10 NOT-ADOPTED-justified (each
with a recorded flip trigger in its report). Actionable findings routed to
TODO_LIST (4 High/Med + 7 Low rows).
Re-review triggers: a NOT-ADOPTED lib's flip trigger fires, a pinned dep
moves a major, or new fleet-wide consumers change the leverage math. For
NOT-adopted libs the report score is counterfactual (100 = zero unmet need).
Citation integrity: `scripts/check-citations.sh` verifies every
`file:line` cite in the reports against the tree (ratchet via
`docs/review/libs/.citations-baseline` — rerun after ANY refactor that
moves cited code; new drift fails, baselined drift prints ACCEPTED).
Tag-watch hook: when a reviewed lib (starting with go-daemon) cuts a new
tag, append an executed-addendum to its report the next time it is
touched — the reports pin versions in their verdicts and go stale
silently otherwise.

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
  2026-11-17): `go mod vendor` copies directory-replaced modules
  verbatim, so hub's `replace go-compliance => ../` vendored the repo
  root INCLUDING flake.nix — the FOD hashed its own placeholder. Fix:
  `postPatchExtra` builds `_core_only` (Go files + go.mod/go.sum only)
  and repoints the hub replace at it. Consequence: doc-only edits no
  longer change the hub vendorHash.
- go.mod/go.sum bumps — and ANY root-module Go-file edit (the `_core_only`
  replace target is vendored verbatim) — invalidate `vendorHash`: tidy per
  module FIRST, then bootstrap (`vendorHash = "sha256-AAA..."` → build →
  paste the `got:` hash → rebuild to confirm convergence) and `nix flake
  check`. The hub hash appears at TWO flake sites (the `compliance-hub`
  package AND the CI devShell's hub-module build) — bootstrap BOTH;
  `nix build .#compliance-hub` alone can pass while `nix flake check`
  still fails on the second site. The core check's hash is stable (no
  directory replace in its module set). There is a THIRD hash site:
  templ-gen's devShell `buildGoModule` (flake.nix) — its `vendorHash`
  tracks the upstream templ repo's module set and can go stale with NO
  local change (2026-11-18: devShell build failed on a mismatch; the
  same paste-`got:` ritual fixed it, converged first paste).
- CI hub/flake jobs need the `FLAKE_PAT` secret (private go-daemon);
  missing secret SKIPS those jobs (step-level `if: env.FLAKE_PAT != ''`
  — job-level env-if is invalid per actionlint).

## Fixture (frozen v1 contract)

`testdata/buildflow-v1-fixture.json` pins 341 redacted repos of real
producer output; `scripts/refresh-fixture.sh` header documents the full
refresh ritual (golden byte-exactness, the @341 benchmark shape,
baseline re-recording).

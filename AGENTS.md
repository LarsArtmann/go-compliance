# AGENTS.md - go-compliance

## Project Overview

**go-compliance** is a fleet compliance hub: a frozen v1 wire contract
(byte-compatible with `buildflow compliance --json`), a source-agnostic
evidence fold + state model (`compliant` / `failing` / `stale` /
`never-covered`), and a realtime dashboard projecting evidence into the
browser via SSE + DataStar patches. Two modules: **core** (root, public
deps only) and **hub** (server + CLI). Coordinates: `go.work`; per-module
work uses `GOWORK=off`. The longer story behind the contracts below is in
`docs/notes/live-architecture.md`.

## Commands

```bash
# Per module, ALWAYS GOWORK=off (module isolation, consumer parity)
cd hub && GOWORK=off go build ./... && GOWORK=off go vet ./...
cd hub && GOWORK=off go test -race -count=1 ./...          # hub suite
GOWORK=off go test -race -count=1 ./...                    # core suite
golangci-lint run ./...                                    # core lint
cd hub && golangci-lint run -c ../.golangci.yml ./...      # hub lint
buildflow                                                  # canonical gate; --dry-run --verbose first
nix flake check                                            # both modules, sandboxed
scripts/pre-release-check.sh vX.Y.Z                        # release (docs/notes/release-runbook.md)

# Dashboard, in this order, after ANY dashboard .templ edit (always from hub/)
cd hub && go tool templ fmt -fail ./internal/dashboard/ && go tool templ generate ./internal/dashboard/
nix shell .#tailwindcss -c bash scripts/regen-ui.sh        # commit the three artifacts with the edit
scripts/screenshot.sh                                      # real-browser gate: 4 views, scroll, keyboard, filters
SHOTS_EVIDENCE=<empty-v1.json> SHOTS_DISCOVER_PATHS= ALLOW_EMPTY=1 \
  scripts/screenshot.sh                                    # empty-fleet gate: run before every deploy
scripts/screenshot-page.sh <url|file> <out.png> [WxH]      # one static capture for docs/assets

# Deployment (host state; both build from the COMMITTED tree)
scripts/deploy-hub.sh                                      # hub as a systemd user unit on the LAN address
scripts/deploy-fleet-evidence.sh --run                     # producer build, daily timer, one scan now
```

- Run `go test` SEQUENTIALLY: the shared `/mnt/buildcache` fails concurrent
  runs with a transient `[build failed]`.
- The auto-commit daemon commits and pushes; never author commits, never
  push. Under IO pressure it lags 30+ minutes and it bundles every actor's
  files into one commit, so read `git status` and `git log --since` before
  and after each step, and treat changes you did not author as not yours.
- Never overlap two full BuildFlow runs (SQLITE_BUSY on every checkpoint).
  `.buildflow.yml` only opts into brandid-lint BD003 and keeps the
  formatters off the generated CSS artifacts.
- Exit-code probes: redirect stderr and `echo $?`; never read `$?` after a
  pipe (`compliance-hub report --db /nonexistent/x.db` exits 75).
- AGENTS.md stays under 220 lines (structure linter, error level): move
  story to `docs/notes/`, keep contracts here.
- Harness (`scripts/screenshot.mjs`): four views (light, dark, `just-wide`
  768px, narrow); the layout is judged per declared view; checks cover the
  scroll box and pinned header, the keyboard pass (the toggle's
  `aria-checked` must follow the theme), filter clicks against the rendered
  fleet, contrast, and console errors versus browser noise. Chromium starts
  with `--disable-extensions`: a host policy force-installs 22 extensions
  whose scripts fail the console check. After a harness edit run
  `node --check` and a real run; the editor's vtsls errors past EOF are
  phantom.
- Subagents: one file scope per write subagent, never the same file twice;
  a write subagent stops at the first red gate and reports, and the
  coordinator re-runs that gate on its files when the report arrives.

## Module Rules

- **Branded IDs (brandid-lint BD003)**: identifier-typed fields and params
  use strong types (`RepoName`/`SourceName`, `sse.EventID` built with
  `sse.NewEventID`). ssetest records the wire `id:` as a plain string, so
  SSE-ID expectations compare strings. Deliberate primitive boundaries carry
  `//brandid-lint:ignore(BD003)`. Gate: BuildFlow plus the CI job pinned to
  `linter/v0.5.1` (bump by hand on new linter tags).
- **Core keeps zero private deps.** The hub's only private dep is go-daemon
  (sd_notify), wired through `_local_deps` and the flake's `postPatchExtra`.
  A new private dep needs a STACK.md decision-log entry.
- **One root `.golangci.yml` for both modules** (v2 has no `extends`).
  Wrapcheck keys are v2 kebab-case; camelCase keys are SILENTLY DROPPED and
  only `golangci-lint config verify` notices. Both modules sit at 0 findings
  with no baseline file. `//nolint` goes on its own line ABOVE the
  statement: golines wraps a trailing directive onto the code line and
  breaks the suppression.
- **templ**: generator pinned by the hub tool directive (v0.3.1100);
  generated `*_templ.go` files are committed. Run `go tool templ generate`
  from `hub/` ONLY: from the repo root it silently rewrites every
  `templ.Error` FileName to a `hub/`-prefixed path (it bit two sessions on
  2026-11-24/25; the CI templ gate now regenerates and diffs). Never import
  `github.com/a-h/templ` in a `.templ` file. The templ LSP `embedlit` hint
  is a false positive: keep `BaseProps: utils.BaseProps{...}` explicit.
- **dashboard CSS** is a compiled Tailwind artifact, never the Play CDN.
  After touching a dashboard `.templ`, adopting a component, or bumping the
  templ-components pin: `scripts/regen-ui.sh`, then commit
  `library-classes.txt`, `library-custom.css`, and `app.min.css` together
  (the CI `css-freshness` gate regenerates and diffs byte-fresh). Test files
  are excluded from the scan (`@source not "**/*_test.go"`): a class that
  only a `_test.go` string mentions would otherwise be compiled into
  `app.min.css`.
- **DataStar attributes** (runtime 1.0.4): DOM events are `data-on:click`
  with a COLON; `data-on-click` is an unknown plugin and is dropped without
  a log line. `data-class` takes an object of class names to booleans.
  `data-show` elements start hidden through inline `style="display:none"`,
  never the `hidden` class. Details: `docs/notes/live-architecture.md`.

## Live Contracts

- `serve` EXITS when the boot refresh fails (no tolerant boot); systemd
  restarts it with backoff.
- `Hub.Rows()` is nil BEFORE the first publish and non-nil for an EMPTY
  published fleet; the `evidence` readiness check gates on exactly that.
- `/readyz`: `refresh-freshness` and `discovery` are non-critical (warn at
  10 minutes stays 200); a hung `QueryEvidence` fails one refresh (30s
  timeout), never the service.
- Patches: `#fleet-summary` (headline, state bar, chips) is re-rendered on
  every state change; `#fleet-rows` rows are appended, replaced, and removed
  by diff. The header's `rendered X ago` RelativeTime is the ONE AutoRefresh
  script carrier and must stay outside every patched region.
- An empty fleet hides its table (`$total > 0`) and shows the no-evidence
  message OUTSIDE `#fleet-rows`: the first patch replaces that body, so a
  message inside it is erased.
- The v1 wire contract is frozen by owner decision
  (`docs/decisions/v2-wire-design.md`): severities for external reporters
  ride v2, the dashboard falls back to the findings count.

## Deployment (this host)

- **Hub**: `scripts/deploy-hub.sh` builds
  `git+file://<repo>?rev=HEAD#compliance-hub`, GC-roots it at
  `~/.local/state/compliance-hub/current`, renders
  `deploy/compliance-hub.user.service` into `~/.config/systemd/user/`, and
  restarts the unit. It names the concrete store path, binds
  `192.168.1.150:8477` (LAN unicast, never `0.0.0.0`), reads
  `~/.local/state/overview/compliance.json`, and runs on a read-only
  filesystem. The script gates itself (answers, every `/readyz` check, the
  harness on the live address) and rolls the unit file and out-link back on
  a failed gate. Verify afterwards with `scripts/screenshot.sh
  http://192.168.1.150:8477/` and `journalctl --user -u compliance-hub`.
- **Evidence-file locations** have one home, the table in
  `docs/guides/integrate.md` (`DefaultEvidenceFile`, the user path, and the
  `GO_COMPLIANCE_EVIDENCE_FILE` override); link to it, never restate it.
- **Producer**: `scripts/deploy-fleet-evidence.sh --run` builds library-policy
  the same way into `~/.local/state/compliance-fleet/current`, installs and
  starts the `fleet-evidence` timer (daily 05:10, `Persistent=true`) that
  runs `library-policy fleet` on the XDG evidence file, and scans once.
  `enable` alone does not start a user timer; the script starts it.
- **File-store limits**: each run replaces a repo's row, so `runs` stays 1.
  `fleet` prunes by default (`--prune`): rows of repos no longer under the
  root are dropped after a complete recording pass.
- **Never-covered fleet**: `serve --discover-paths` (default `~/projects`)
  sends real roots to the daemon (empty roots are a 400); only projects
  with a `go.mod` count. `discovery` readiness passes while the watch
  stream is connected. No paths switches discovery off by design
  (`startDiscovery`): never-covered stays empty and one log line says so.
- **CSP**: every response carries a Content-Security-Policy; DataStar
  compiles expressions with `new Function`, so `script-src` needs
  `'unsafe-eval'`. Removing it goes inert with no server-side error; only
  `scripts/screenshot.sh` (console-error check) sees it.
- NEVER sudo: root cannot fetch the private Go dependencies and the system
  evidence path belongs to another user. The non-loopback bind warning is
  the documented speed bump (`docs/decisions/bind-posture.md`). The units
  are plain user files, not home-manager managed, pending the owner's
  SystemNix decision.

## Testing Conventions

- ginkgo/gomega; one suite bootstrap per package; black-box `_test`
  packages.
- SSE/DataStar handlers are tested with `datastartest`
  (`CollectWithTimeout(GinkgoTB(), handler, timeout, WithPath(...))`),
  never by polling an `httptest.ResponseRecorder` from another goroutine.
- Wall-clock specs: the work inside the asserted window stays
  microsecond-scale (never fixture IO); sleep at least 2x the TTL for
  expiry; plant a corrupt document mid-window so broken backoff fails
  loudly. Never sleep to "let a watch arm": probe until the effect shows
  (`armWatch` in `watcher_internal_test.go`), and space probes wider than
  `refreshDebounce`, because every kick restarts the debounce window.
- Cross-package helpers live in `hub/internal/testsupport`; call
  `testsupport.WaitForSubscriber(hub)` before publishing to a fresh
  subscriber.
- String tests cannot see a runtime that ignores an attribute: anything
  reactive or visual also gets a `scripts/screenshot.sh` check.

## Gates and Notes

- Dedup: `art-dupl -t 1 --type-aware`; accept intentional clones with
  `// art-dupl:accept <why>` within 5 lines above the clone; CI ratchets
  `.art-dupl-baseline.json`. Policy: `docs/notes/dedup-audit-policy.md`.
- Docs: `nix run github:Artmann-Technologies/docs-gates -- run .` (the CI
  build; on PATH only `~/go/bin/docs-gates` runs it, a wrapper; the
  `~/.local/bin` and system copies are stale builds that report green while
  CI's audit-all fails) and `dprint check`. Run the battery before AND after
  every annotation sweep; `dprint fmt` in the same minute as a markdown edit
  (the daemon commits unformatted files and CI's dprint job goes red).
  Multi-section annotate keys need the `N@substring` disambiguator; run
  `emit-keys` and `--dry-run` first
  (`docs/notes/docs-health-annotate-writer.md`). After EVERY `dprint fmt`
  on a swept report run `node scripts/annotate-qa.mjs --base add --fix
  <file>` (renumbered items, split strikes, eaten spaces; dprint nests a
  struck item that touches a bare neighbour, so keep blank lines around
  struck items). The CI docs-gates
  job checks out the full history because its annotate gate resolves
  `done at <sha>` verdicts.
- Library audits: `docs/review/libs/`; rerun `scripts/check-citations.sh`
  after any refactor that moves cited code.
- Nix: go.mod changes flip `vendorHash` — tidy per module, then
  `buildflow -s nix-hash-fix --fix`, and finish with `nix flake check`
  (three hash sites; CI skips the flake job without `FLAKE_PAT` and still
  reports success).
- Fixture: `testdata/buildflow-v1-fixture.json` pins 341 redacted repos;
  `scripts/refresh-fixture.sh` is the refresh ritual.
- BuildFlow warning triage (2026-11-27, plain `buildflow` exits 0; these
  warning groups are deliberate non-fixes, do not "fix" them again):
  branching-flow and erraudit (style, both modules), go-auto-upgrade
  (dependency pins move only in a bump task), go-version-auto-configure
  (a dependency forces the `go 1.27.2` floor; tidy restores it),
  nix-checker (hashes stay inline: `nix-hash-fix` rewrites them in place),
  go-structure-linter (`docs/assets`, top-level layout, fixtures are
  decided), library-policy companions (this hub IS the health dashboard;
  three headers do not justify the `httputil` dependency), io-surface on
  the discovery `http.Client` (no `Timeout` on purpose: the watch is a
  long-lived SSE stream, every request carries a context), govulncheck (its
  "findings" are the report's own section headers, the run shows zero
  called vulnerabilities) and vulnix (advisories in the nixpkgs closure).

# Live architecture and operations notes

The contracts an AI session must know sit in `AGENTS.md`. This note keeps
the longer story behind them. It moved out of `AGENTS.md` on 2026-11-25,
when that file crossed the 220-line ceiling of the structure linter.

## Liveness (M27)

`refresher` (`hub/internal/cli/refresher.go`) owns liveness: a trigger
channel (cap 1, coalescing), a 500ms debounce (a fixed window from the
first trigger of a burst: a burst costs one refresh, a steady stream keeps
producing one refresh per window, pinned by two specs in
`refresher_internal_test.go`), a 5-minute safety-net ticker
(backstop), discovery `/v1/watch` with a 30s backoff, and, since 2026-11-19,
a go-filewatcher/v2 directory watch (`cli/watcher.go`) over the cache DB and
its WAL sidecars (and over the evidence file's directory for `--source
file`). Cross-process SQLite writes and atomic evidence renames therefore
refresh in about 1-2 seconds instead of waiting for the safety net;
`evidence-watch` is its non-critical health check. `/v1/compliance` serves
the refresher's snapshot and answers 503 before the first successful
refresh. The watch stream is framed by `daemon.ParseSSEData` (WHATWG rules:
multi-line `data:` joins, over-cap lines fail loudly;
`discovery/client.go`). Unix binds are probe-first (`bind.go`
`removeUnownedSocket` plus `daemon.IsSocketAlive`): a second hub never steals
a live sibling's socket and fails with "refusing to steal".

## Boot contract

Verdict recorded 2026-11-18: `serve` EXITS when the boot refresh fails.
There is no tolerant boot. A failed first query means the evidence source
itself is broken; systemd `Restart=on-failure` retries with backoff and the
operator gets paged by the service manager instead of reading a green
process that advertises a 503 window. The readiness machinery
(503-until-publish, the `MarkShuttingDown` ordering) remains for embedders
and for graceful shutdown.

## Health surface

The `evidence` readiness check (`cli/health.go`) gates on
`hub.Rows() == nil`. `Hub.Rows()` returning nil BEFORE the first publish
and non-nil for an EMPTY published fleet is a load-bearing contract. It was
fixed on 2026-11-18 after the gate proved inert, is regression-pinned in
`events_test.go` and `cli/health_internal_test.go`, and carries verdict g2:
an empty published fleet IS ready.

The non-critical `refresh-freshness` check (`refresher.LastRefreshSuccess`,
warn threshold 10 minutes = 2x the safety net) and the per-refresh
`refreshTimeout` (30s) close the former wedge hole together: a hung
`QueryEvidence` fails ONE refresh (logged, retried) and degrades `/readyz`
to warn instead of freezing green. Verdict g1: warn stays 200, because a
stale view is degraded truth, not service loss. The non-critical
`discovery` check (same warn semantics, stamp = last watch connect or
snapshot, threshold 10 minutes) makes a long-dead daemon visible in
`/readyz`.

## DI posture

samber/do is transitive only (cmdguard v5 and go-health bring it). The hub
wires manually by recorded decision (`STACK.md`); the reasoning is in
`docs/architecture-understanding/2026-11-18_19-39_samber-do-health-architecture.md`.

## Dashboard patches

Rows are appended, replaced, and removed by diff (`events.DiffRows`). The
page-level `rendered X ago` RelativeTime in the header is the ONE
AutoRefresh carrier and ships the 30s tick script. Row cells opt into
ticking through the bare `data-tc-relative` attribute and carry no script:
`RelativeTimeProps` couples `AutoRefresh` to a per-instance `<script>`, rows
re-render inside every patch, and per-row scripts would bloat the patches
and accumulate in the tbody. A `view_test.go` regression test pins this.

## Dashboard DataStar contract (2026-11-25)

The pinned runtime is DataStar 1.0.4 (`go-datastar/static` v0.8.0). String
tests cannot see a runtime that ignores an attribute, and the fleet filter
shipped inert for that reason: these rules are load-bearing.

- DOM events are `data-on:click` (colon). `data-on-click` parses as a plugin
  named `on-click`, which does not exist, and is dropped without a log line.
  `data-on-interval`, `data-on-intersect`, and `data-on-signal` are plugin
  names of their own and keep the hyphen.
- `data-class` takes an object of class names to booleans
  (`{ 'a b': $x === 'y', 'c d': $x !== 'y' }`). A class string is iterated
  character by character and yields numeric junk classes.
- The server renders the selected chip classes as well, so the page is right
  without JavaScript; the runtime's class action only adds or removes what
  differs.
- Elements that start hidden and are revealed by `data-show` carry
  `style="display:none"`, never the `hidden` class: the show action removes
  an inline `display`, and a class would keep the element hidden.
- `data-attr:aria-pressed` takes a string expression. A boolean sets the
  attribute to an empty value or removes it.
- `ViewModel.Counts` and `ViewModel.Total` describe the whole fleet, even
  when `?state=` narrowed `Rows`; they also seed `data-signals` so no
  expression reads a missing signal before the stream connects.
- The table region is not a live region (`datastar.LiveOff`): announcing a
  200-row table on every patch would flood assistive technology.

`scripts/screenshot.sh` is the gate for all of the above. It drives a real
headless Chromium over the DevTools protocol through four views (light,
dark, just-wide at 768px, narrow at 390px), each naming the layout it must
produce, with one check family per concern: the layout the page actually
produced (the column header's display, and the table's box scrolling
exactly when the table is taller than the box's height limit); the scroll
of that box and the column-header pin after a 1400px scroll; the keyboard
walk (chips via Enter and Space, the theme toggle with its
`aria-checked` trail, focus rings on everything it stops at); filter
clicks judged against the fleet the page rendered, never against a fleet
the harness assumes; WCAG AA contrast for every text pair; the empty fleet
(`ALLOW_EMPTY=1` with `SHOTS_DISCOVER_PATHS=` switched off); and console
errors, of which only classified browser noise is exempt (exception from
`chrome-extension://` scripts and Chromium's own `runtime.lastError`;
Chromium starts `--disable-extensions` because a managed host's policy
force-installs extensions into every profile).

## Fleet table geometry (2026-11-28)

In a wide layout the fleet table scrolls inside the box the library wraps
it in: a container query at 40rem of container width gives that box
`max-height: max(20rem, calc(100dvh - 254px))` and `overflow: auto`
(`app.css`). The 254px is calibrated against the rest of the page: the
summary block above the table is 222px tall at every width from 640px up,
plus the page's bottom padding, so the box takes exactly the height the
viewport leaves and the page itself never scrolls (the harness fails when
it does). `TableProps.StickyHeader` pins the column header to the top of
that box over 200 rows. Below 40rem a second container query hides the
header, stacks each row into a two-column grid, and hands scrolling back
to the page.

The empty-fleet state shares the patch contract's one hard rule: nothing
the stream replaces may carry a message. The "No compliance evidence yet"
block sits outside the rows region (`view.templ`
`noEvidenceShowExpression`, `$total===0`), the table hides through
`tableShowExpression` (`$total > 0 ...`), and a selected chip whose state
holds no rows gets its own message (`chipEmptyShowExpression`). The first
version kept the message inside the tbody, and the first patch of an
empty fleet erased it (found by the harness's `ALLOW_EMPTY=1` run on
2026-11-28, fixed and pinned by a view spec and the empty-fleet harness
run).

## templ-components verdict

The hub is a component-layer citizen: `layout.Base`, `ThemeToggle`,
`display.Table`, `PageHeader`, `RelativeTime`, `SimpleEmptyState`,
`Badge` through the typed `StateBadge` in `rows.templ`,
`errorpage.NotFound404`, and `datastar.SDKScript` and `LiveRegion`.
`StateBadge` stays local on purpose: it maps the domain enum
`compliance.State` exhaustively at compile time, where
`display.StatusBadgeWith` would map strings.

`chips.templ` is a justified hand-roll. The library ships only link-mode and
radio-mode `FilterChips` and `SegmentedControl`; the fleet filter binds a
DataStar signal and carries live counts.

Theming goes through `@theme` overrides in `app.css` only: one accent (the
`accent-*` family the library defines), muted `red`/`amber`/`green` ramps
(the library hardcodes those names for badges and tiles), and the system
font stack. The consumer never restyles library components in Go.

## Nix and FOD lessons

- **The vendorHash FOD self-reference flip-flop** (root-caused 2026-11-17,
  retired at v0.2.0 when the hub replace was dropped): `go mod vendor`
  copied directory-replaced modules verbatim, so the hub's
  `replace go-compliance => ../` vendored the repo root INCLUDING
  `flake.nix`, and the FOD hashed its own placeholder. The `_core_only`
  `postPatchExtra` fix is gone with the replace; keep the lesson for any
  future directory replace in a vendored module.
- `go.mod`/`go.sum` bumps, and ANY root-module Go-file edit, invalidate
  `vendorHash` (the hub has consumed core from the proxy since v0.2.0, so
  core edits no longer flip the HUB hash but still flip the CORE check's).
  Tidy per module FIRST, then rebuild the hash through
  `buildflow -s nix-hash-fix --fix`, and finish with `nix flake check`. A
  past run had the hub hash at TWO flake sites, so `nix build
  .#compliance-hub` alone can pass while `nix flake check` fails on the
  second site. A third site exists: templ-gen's devShell `buildGoModule` in
  `flake.nix`, whose `vendorHash` tracks the upstream templ repository's
  module set and can go stale with NO local change (2026-11-18).
- CI hub and flake jobs need the `FLAKE_PAT` secret (private go-daemon). A
  missing secret SKIPS those jobs: step-level `if: env.FLAKE_PAT != ''`,
  because a job-level env-if is invalid per actionlint. A skipped
  `nix flake check` job still reports `success`, so CI green does not prove
  the flake; run `nix flake check` locally.
- The `docs-gates` CI job checks out the full history (`fetch-depth: 0`):
  its annotate gate resolves every `done at <sha>` verdict against the
  object store, which a depth-1 clone cannot do.

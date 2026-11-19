# templ-components

| Field                | Value                                                                                              |
| -------------------- | -------------------------------------------------------------------------------------------------- |
| Reviewed             | 2026-11-19                                                                                         |
| Group                | B (realtime-dashboard stack)                                                                       |
| Consumer module      | `hub` (`github.com/larsartmann/go-compliance/hub`) — imports confined to `hub/internal/dashboard`   |
| Pins                 | `templ-components v1.27.0` + `/datastar v1.27.0` + `/icons v1.27.0` + `/utils v1.27.0` (all direct); `/display` and `/layout` imported in `.templ` (compile-time deps of the root module, no separate require); `/htmx v1.27.0` indirect (unused) |
| Library version      | main tag `v1.27.0` (`git -C /home/lars/projects/templ-components describe --tags --abbrev=0` reports `charts/echarts/v1.27.0`, newest tag overall; newest `vX` line tag is `v1.27.0`); HEAD `d5702fc1 2026-11-17 chore: auto-commit`; tags newest-first: `charts/echarts/v1.27.0`, `datastar/v1.27.0`, `errorpage/v1.27.0`, `htmx/v1.27.0`, `icons/v1.27.0` |
| Pin current?         | **Yes** — v1.27.0 is the newest main-line tag. HEAD is ahead of the tag (auto-commits + a DataStar 1.0.4 SDK runtime bump via `static v0.8.0` in `15d836b8`), i.e. a future tag will move the DataStar runtime; nothing released beyond the pin today. |

Library contract read before auditing: `/home/lars/.config/crush/skills/templ-components/SKILL.md`
(138 components / 11 packages, props structs embedding `utils.BaseProps`, typed enums
with `IsValid()`, `utils.Class` merging, icon set, HTMX/DataStar helpers + `wire.Action`,
guard tests, compiled-CSS rule, dark-mode `gray-*` + `dark:` conventions).

## Verdict: UNDERUSED-slight

What the dashboard uses, it uses *well and deeply*: `display.Table` with the
`LazyRows`/`BodyID`/compact-padding props turned into a DataStar patch target
(`view.templ:84-93`), `display.RelativeTime` in the sophisticated two-mode split the
repo pinned a regression test for (page-level `AutoRefresh` carrier vs bare
`data-tc-relative` row attributes — `view.templ:41`, `rows.templ:30-44`,
`view_test.go:53-79`), `Badge` with typed tone enums (`rows.templ:60-75`),
`Grid`+`StatCard`+`icons` for the stat strip, `layout.Base` (which ships the
dark-mode `ThemeScript` by default), and the `/datastar` module's `LiveRegion` +
`SDKScript`. That is 10 of 138 components carrying an entire realtime dashboard —
the STACK.md choice ("fleet server-rendered component library … vs hand-rolled
HTML/CSS") is broadly vindicated. The "slight" is threefold: the filter chips are
hand-rolled (`chips.templ`) where `forms.FilterChips`' zero-JS link design matches
the dashboard's own `?state=` fallback exactly; the `errorpage` module is unused
while the hub serves Go-default 404s; and the page violates the library's
compiled-CSS contract by loading the Tailwind Play CDN at runtime
(`view.templ:74-77`) — the library's theming pipeline, the thing that makes
components render correctly in production, is the one capability misused.

**Flip trigger:** adopt `forms.FilterChips` (or upstream a signal-mode variant),
wire `errorpage`, and ship compiled CSS → ADOPTED; score then lands high-80s.
(Also fixes row 14 — the missing `ThemeToggle` — for free.)

## Derived score: 75 (high) — 12 of 16

- Inventory rows N = 25; applicable A = 16 (9 N/A rows excluded).
- Fully = 10 × 1.0; Partial = 4 × 0.5 (rows 6, 12, 13, 14); Missed/Misused = 2 × 0 (rows 15, 16).
- Σ = 12.0; score = round(100 × 12 / 16) = **75**, band **high** (70-84).

## Capability inventory

| # | Capability | Status | Library evidence (path, at v1.27.0) | Consumer evidence (path:line) |
|---|---|---|---|---|
| 1 | Page shell (`layout.Base`: meta, skip-link, a11y, ThemeScript) | Fully | `layout/base.templ:228,271-272` | `view.templ:18-23` |
| 2 | Content layout (`layout.Container`) | Fully | `layout/container.templ` | `view.templ:24` |
| 3 | `display.PageHeader` | Fully | `display/page_header.templ` | `view.templ:26-29` |
| 4 | `display.Grid` + `display.StatCard` (tones + icon slots) | Fully | `display/grid.templ`, `stat_card.templ` | `statcards.templ:13-38` |
| 5 | `display.Table` incl. `LazyRows`, `BodyID`, compact `CellPadding`, `Striped`/`Hover` | Fully | `display/table.templ:26,74-83,155-156` | `view.templ:84-93` (BodyID feeds `DefaultRowsSelector`) |
| 6 | Table sorting / clickable rows (`Sortable`+`Href` headers, `Row.Href`) | Partial | `display/table.templ:26-28,99-111` | `view.templ:84-93` sets neither; sort order is the fixed worst-first triage invariant (`cli/table.go:56-58` mirrors it) |
| 7 | `display.Badge` (typed `BadgeType` enums) | Fully | `display/badge.templ` | `rows.templ:60-75` (`StateBadge`) |
| 8 | `display.RelativeTime` (+ AutoRefresh singleton pattern) | Fully | `display/relative_time.templ` | `view.templ:39-42` (carrier), `rows.templ:30-44` (bare-attribute rows), regression `view_test.go:53-79` |
| 9 | `/datastar` components (`LiveRegion`, `SDKScript`) | Fully | `datastar/live_region.templ`, `sdk_script.templ` | `view.templ:45-56`, `view.templ:66-73` |
| 10 | `icons` (102 icons) | Fully | `icons/icon_paths.go` | `statcards.templ:18,24,30,36` (Server, ExclamationTriangle, Clock, Search) |
| 11 | `utils` (`BaseProps`, `Class`, `utils/format`) | Fully | `utils/` | `view.templ:46`, `statcards.templ` (BaseProps everywhere); `utils/format` unused — no byte/duration rendering needs yet |
| 12 | `forms.FilterChips` (+ `FilterToggleHref` query builder, `aria-current`) | Partial | `forms/filter_chips.templ:1-45` (zero-JS link chips) | `chips.templ:19-46` hand-rolls 5 chips, 3 class constants, `aria-pressed` logic |
| 13 | Empty-state ladder (`EmptyState`/`SimpleEmptyState`/`DataState`) | Partial | `display/empty_state.templ`, `data_state.templ` | `view.templ:97-104` hand-rolls a `<td colspan=4>` empty row inside the tbody |
| 14 | Dark-mode theme system (`ThemeScript` + `ThemeToggle`) | Partial | `layout/theme_toggle.templ`; `base.templ:182-191,271-272` | script shipped by Base (implicit use); **no `ThemeToggle` rendered** — dark mode follows OS preference only |
| 15 | Theming pipeline (compiled CSS, `@theme` vars; "compiled-CSS rule" in SKILL) | **Misused** | SKILL.md Part 1 "Consumer essentials" + `references/consumer-guide.md` | `view.templ:74-77` loads `https://cdn.tailwindcss.com` (Play CDN, dev-only per Tailwind) + an inline `tailwind.config` script at runtime |
| 16 | `errorpage` module (`ErrorPage`, `NotFound404`, `ErrorHandler`) | Missed | `errorpage/` (4 components + 6 constructors + `ErrorHandler`) | no import anywhere in `hub/`; serve mux serves `net/http` default 404/plain `http.Error` (`cli/serve.go` mux wiring, `dashboard/routes.go:53-61`) |
| 17 | Stream-error surfacing (`datastar.SSEErrorHandling`, `htmx.GlobalErrorHandling`) | N/A | `datastar/sse_error_handling.templ` | deliberate alternative design: the `rendered X ago` carrier going stale IS the dead-stream signal (`view.templ:35-38`, pinned by test) |
| 18 | Navigation suite (`Nav`, `SidebarNav`, `Breadcrumbs`, `Footer`) | N/A | `navigation/` | single-page dashboard |
| 19 | Forms suite (26 components: `Form`, `Input`, `Select`, …) | N/A | `forms/` | no data entry on the dashboard |
| 20 | Feedback suite (`Toast`, `Alert`, `Skeleton*`, `Spinner`, …) | N/A | `feedback/` | server renders full state; no async loading states to bridge |
| 21 | Modal/overlay suite (`Modal`, `Drawer`, `Dropdown`, …) | N/A | `display/` | no overlays needed |
| 22 | `DataTable` | N/A | `display/` | the fleet table is patch-driven via `LazyRows` tbody; ~280 rows need no pagination UI |
| 23 | Charts (`Sparkline`, `BarChart`, …, `charts/echarts`) | N/A (revisit) | `display/sparkline.templ`, `charts/echarts/` | no time-series surface yet; the evidence cache (`hub/internal/cache/cache.go:35` `evidence_runs`) accumulates the history that would make a fail-rate sparkline meaningful — revisit when a history API lands |
| 24 | `/htmx` module | N/A | `htmx/` | DataStar-only by design (correctly left indirect/unused; `view.templ:21` sets `HTMXVersion: ""`) |
| 25 | `KanbanBoard`, `Carousel`, `CommandPalette`, … | N/A | `display/` | no matching use case |

## Gap analysis

### Row 15 — Tailwind Play CDN in production violates the library contract (Misused)

**Library offers:** the compiled-CSS rule — components are authored against
Tailwind v4 utility classes; consumers compile one CSS bundle (scanner over
`.templ` sources) and ship it; the demo's `nix run .#css` is the reference flow
(SKILL.md "Recompile the CSS when classes change"; `references/consumer-guide.md`).

**Consumer does:** `view.templ:74-77`:

```templ
<script src="https://cdn.tailwindcss.com"></script>
<script>
	tailwind.config = { darkMode: 'class' }
</script>
```

with the in-repo acknowledgment "no build step until compiled CSS lands"
(`view.templ:63-64`).

**Impact:** three-fold: (1) Tailwind explicitly marks the Play CDN
development-only — it compiles CSS in the browser on every page load (a ~100 KB+
JS dependency and a runtime flash-of-unstyled-content race against DataStar's
first patches); (2) an offline/air-gapped operator (this is a localhost fleet hub)
gets an unstyled dashboard whenever the CDN is unreachable; (3) the inline
`tailwind.config` script is CSP-unsafe — the library's own guard tests
(`TestAllInlineScriptsHaveNonce`) exist precisely because this pattern is banned
in library code. Every class the components render (`dark:` variants especially)
round-trips through a runtime JIT the library never tests against.

**Sketch:** add a compiled stylesheet the library way and drop the CDN:

```bash
# one-time: scan the hub's templ sources, compile, embed
cd hub && go run github.com/a-h/templ/cmd/templ generate ./internal/dashboard/
tailwindcss -i internal/dashboard/css/app.css -o internal/dashboard/static/app.css --minify
```

```templ
// view.templ — replace the Play CDN block; layout.Stylesheet serves the file
@layout.Base(layout.PageProps{
	Title:   model.Title,
	CSSPath: "/static/app.css",   // compiled bundle, embedded via go:embed + a static handler
	...
})
```

`layout.Base` already supports `CSSPath` (`view.templ:20` passes `""` today). The
`darkMode: 'class'` config becomes moot — compiled v4 CSS + Base's `ThemeScript`
class toggling already interoperate.

### Row 12 — hand-rolled filter chips vs `forms.FilterChips` (Partial)

**Library offers:** `forms.FilterChips` — a row of zero-JS link chips with
`aria-current`, `FilterToggleHref(path, query, key, value)` building the query
string, and library-owned selected/unselected styling
(`forms/filter_chips.templ:1-45`).

**Consumer does:** `chips.templ:19-46` — five hand-rolled `<button>`s wired to
the DataStar `$stateFilter` signal (`data-on-click`, `data-class` ternaries), plus
three hand-maintained class constants and `aria-pressed` logic. The no-JS path
already exists in the consumer (`?state=` query, `routes.go:88-92`) — which is
*exactly* the mode `forms.FilterChips` embodies natively.

**Impact:** moderate fleet cost, low local cost. The signal-driven interactivity
(set-filter-without-reload) is a real requirement `forms.FilterChips` does not
serve — so this is not a pure duplicate — but the chip *styling, ARIA, and
selected-state logic* are duplicated, and any fleet chip restyle will miss this
repo. The gap is bidirectional: the library lacks a DataStar-signal chip variant
that 26 go-datastar-consuming projects would reuse.

**Sketch (upstream-first, adopt-second):**

```templ
// upstream: forms.FilterChips gains a signal mode (zero-JS links stay the default)
@forms.FilterChips(forms.FilterChipsProps{
	Signal: "stateFilter",           // render data-on-click="$stateFilter='x'" buttons
	Chips: []forms.FilterChip{
		{Label: "Failing", Value: "failing", Active: model.StateFilter == "failing"},
		...
	},
})
```

Until that exists, the honest local move is to keep the signal buttons but delete
the hand-rolled class ladder by reusing the library's chip classes (exported via
a `forms.ChipClass(active bool)` helper — also an upstream one-liner).

### Row 16 — `errorpage` unused; the hub serves default 404s (Missed)

**Library offers:** `errorpage.NotFound404`, `ErrorPage`, `FromError`, and
`ErrorHandler`/`WriteNotFound404` mux helpers — branded, dark-mode-aware error
pages with the same Base shell as the dashboard.

**Consumer does:** nothing — `rg errorpage hub/` is empty; unknown paths get Go's
plain-text 404, and the JSON API endpoints use bare `http.Error`
(`cli/serve.go:86-99`). For a service whose entire surface is one dashboard +
API + health, a branded 404 and a rendered 500 (instead of a stack-less white
page) is a 10-line win that keeps the operator UX coherent.

**Sketch:**

```go
// cli/serve.go, after mux wiring
mux.HandleFunc("/", errorpage.NotFound404) // branded fallback for unknown paths
// and wrap the JSON handlers' error returns:
errorpage.WriteError(w, err) // instead of http.Error(w, err.Error(), 500)
```

(Requires adding `templ-components/errorpage` to hub's requires; it is already
resolved at v1.27.0 via the root module.)

### Row 14 — no `ThemeToggle` (Partial)

Base ships the `ThemeScript` that honors `localStorage['theme']` and sets the
`.dark` class (`base.templ:271-272` at v1.27.0), and every dashboard component
carries `dark:` variants — but with no `ThemeToggle` rendered anywhere, a user
who prefers light-mode-on-a-dark-OS (or vice versa) cannot switch. One line in
`view.templ:26-29` (slot verified on the pin: `PageHeaderProps.Action`;
`layout.ThemeToggle(ariaLabel, nonce)` at `layout/theme.templ:53-54`):

```templ
@display.PageHeader(display.PageHeaderProps{
	Title:    model.Title,
	Subtitle: "Real-time gate compliance across the fleet",
	Action:   layout.ThemeToggle("Toggle theme", ""), // renders right of the title row
})
```

### Row 13 — hand-rolled in-table empty state (Partial, low priority)

`view.templ:97-104` renders its own "No compliance evidence yet" colspan row.
`display.EmptyState`/`DataState` target block-level regions, not a `<td>` inside
a patchable tbody — the hand-roll is a reasonable special case. If the library
ever grows a table-native empty row (an `EmptyRow` helper), swap; otherwise leave.

### Row 6 — `Sortable` headers unused (Partial, low priority)

The worst-first order is the documented triage invariant; alternate client-side
sorts would fight the patch stream (server owns ordering). Server-side sort via
`Sortable`+`Href` headers would round-trip through `?sort=` — feasible but adds
page-reload behavior to an otherwise no-reload dashboard. Leave; noted so the
omission is a recorded decision rather than an unknown.

## Version currency

- Pin v1.27.0 == newest main tag. **Current.**
- HEAD beyond the tag is auto-commit noise plus one meaningful commit:
  `15d836b8 Update Datastar SDK runtime to v1.0.4 via go-datastar/static v0.8.0
  (#98)` — i.e. the next templ-components release will move `datastar.SDKScript`
  off DataStar JS 1.0.3. That interacts with the hub's static v0.6.1 pin: when
  the hub bumps go-datastar/static to ≥ v0.7.0 (see the go-datastar report), it
  should bump templ-components in the same motion once tagged, keeping
  `SDKScript`'s rendered runtime and the self-served bundle on one version.
- Upgrade risk on a future v1.28: watch `datastar/version.go` (the 1.0.3 → 1.0.4
  constant rename deprecates `DatastarVersion1_0_3`); the hub renders no explicit
  version (`SDKScriptProps{Src: model.ScriptSrc}` — follows the pin), so risk is
  confined to the bundle/runtime pairing above.

## Fleet leverage

**Expand (selectively).** 40 consumers per STACK.md; this repo demonstrates the
library's hardest pattern (components *inside* an SSE patch stream — patchable
`BodyID` tbody, script-free re-rendered rows) which is exactly the reference
other fleet dashboards should copy; `view_test.go`'s regression pin is the kind
of consumer-side guard that should be upstreamed into docs. The expansions with
fleet payoff: (a) a DataStar-signal variant of `FilterChips` — every go-datastar
project with a filterable table wants it; (b) `errorpage` adoption here (cheap,
visible, reinforces the "no hand-rolled HTML" rule); (c) compiled CSS — the
library's contract only pays off fleet-wide if consumers actually ship compiled
bundles; this repo is currently the off-pattern outlier. **Leave** the htmx
module unused (correct), and leave charts until a history API exists.

## Ranked opportunities (impact × ease)

1. **Ship compiled CSS, drop the Play CDN** (row 15) — highest impact
   (production correctness, offline operation, CSP); moderate ease (build step +
   embed + route; the in-repo comment already reserves the decision).
2. **Wire `errorpage` 404/500 handlers** (row 16) — moderate impact, high ease
   (~10 lines + one require).
3. **Add `layout.ThemeToggle`** (row 14) — small impact, trivial ease.
4. **Upstream signal-mode `FilterChips`** (row 12) — moderate fleet impact,
   medium ease (upstream PR + guard tests + count updates per SKILL).
5. **Note the `Sortable`/`DataTable` leaves as decisions** (rows 6, 13) — zero
   risk, prevents re-litigation.

## How to verify (executed 2026-11-19)

```bash
# Library metadata
git -C /home/lars/projects/templ-components tag | grep '^v1' | sort -V | tail -3   # v1.25.1, v1.26.0, v1.27.0
git -C /home/lars/projects/templ-components log -1 --format='%h %cs %s'            # d5702fc1 2026-11-17 chore: auto-commit 3 changed file(s) (heuristic)
git -C /home/lars/projects/templ-components log v1.27.0..HEAD --oneline | head     # incl. 15d836b8 DataStar 1.0.4 via static v0.8.0
git -C /home/lars/projects/templ-components tag --sort=-creatordate | head -5      # charts/echarts/v1.27.0, datastar/v1.27.0, ...

# Skill contract (read first)
sed -n '1,60p' /home/lars/.config/crush/skills/templ-components/SKILL.md

# Consumer usage — the complete import set (display 6×, utils/layout/icons/datastar 2× each)
cd /home/lars/projects/go-compliance
rg -o 'templ-components/[a-z/]+' hub -g '*.templ' -g '*_templ.go' -g '*.go' | grep -o 'templ-components/[a-z/]*' | sort | uniq -c
rg -n 'display\.\w+|layout\.\w+|datastar\.\w+|icons\.\w+' hub/internal/dashboard/*.templ

# Hand-rolled vs library: chips, empty state, Play CDN, error pages
sed -n '19,46p' hub/internal/dashboard/chips.templ
sed -n '63,78p' hub/internal/dashboard/view.templ
git -C /home/lars/projects/templ-components show v1.27.0:forms/filter_chips.templ | head -45
rg -n 'errorpage' hub -g '*.go' -g '*.templ'        # (empty — module unused)
rg -n 'http.Error' hub/internal/cli/serve.go hub/internal/dashboard/routes.go

# Regression pin for the RelativeTime carrier pattern
sed -n '53,79p' hub/internal/dashboard/view_test.go

# Effective pins
cd hub && GOWORK=off go list -m all | grep templ-components
```

All file:line citations spot-checked against the working tree on 2026-11-19.

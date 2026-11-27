# go-output

| Field           | Value                                                                                                                                                                                                                                                                                                                                                  |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Reviewed        | 2026-11-19                                                                                                                                                                                                                                                                                                                                             |
| Group           | B (realtime-dashboard stack)                                                                                                                                                                                                                                                                                                                           |
| Consumer module | `hub` (`github.com/larsartmann/go-compliance/hub`) — core module does not depend on go-output                                                                                                                                                                                                                                                          |
| Pins            | `github.com/larsartmann/go-output v0.38.7` (direct) + `go-output/table v0.38.7` (direct). 11 further sub-modules resolved at v0.38.7 as **indirect** via `cmdguard/v5` → `samber-do-auditlog` (tree, markdown, delimited, markup, escape, d2, daghtml, graph, plantuml, serialization, testhelpers).                                                   |
| Library version | root line newest `v0.38.7` (`git -C /home/lars/projects/go-output tag \| grep '^v' \| sort -V \| tail -1`); repo-wide newest tag `bdd/v0.38.7` (`git describe --tags --abbrev=0`); HEAD `90fb2a6 2026-11-17 fix(nix): update vendor hash...`; tags newest-first: `bdd/v0.38.7`, `d2/v0.38.7`, `daghtml/v0.38.7`, `delimited/v0.38.7`, `escape/v0.38.7` |
| Pin current?    | **Yes** — v0.38.7 is the newest root-line tag (the 19-module v0.38.7 family was cut together; HEAD commits since are release/nix/lint chores with no API delta).                                                                                                                                                                                       |

## Verdict: UNDERUSED-moderate

The consumer uses go-output for exactly the reason STACK.md names first —
"ANSI-display-width-aware tables (BuildFlow gotcha #298)" — and uses it well:
every cell flows through `go-output/table`'s lipgloss-backed renderer with a
single color-mode authority shared between cell strings and table chrome
(`hub/internal/cli/table.go:52-82`, `colors.go`), with a regression test holding
the alignment invariant. But the second half of STACK.md's own rationale —
"+ JSON/CSV/MD/HTML" — is entirely unrealized: the `table` command has no
`--format` flag, so the library's 16-format `output.RenderTable` surface is
unreachable, and the footer/StyleFunc refinements of the table module itself go
unused. The `report` command correctly does NOT use go-output's serialization
(the frozen v1 contract is stdlib `encoding/json/v2` by explicit policy —
STACK.md "JSON" row: "all third-party JSON libs banned" — and go-output's
`serialization` module is go-faster/jx-based, i.e. exactly the banned class;
leaving it unused is _compliance_, not underuse). The moderate verdict rests on
one axis: the fleet's terminal-output standard is 2 modules deep in this repo
while 13 more sit in the module graph already paid for.

**Flip trigger:** add `--format` to the `table` command (opportunity 1) →
UNDERUSED-slight; adopt the footer totals row → ADOPTED-adjacent.

## Derived score: 57 (moderate) — 4 of 7

- Inventory rows N = 14; applicable A = 7 (7 N/A rows excluded).
- Fully = 3 × 1.0; Partial = 2 × 0.5; Missed/Misused = 2 × 0.
- Σ = 4.0; score = round(100 × 4 / 7) = **57**, band **moderate** (50-69).

## Capability inventory

| #  | Capability                                                                                         | Status                      | Library evidence (path)                                                                                                               | Consumer evidence (path:line)                                                                                                                                                                        |
| -- | -------------------------------------------------------------------------------------------------- | --------------------------- | ------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1  | Terminal table rendering (`table.New`/`SetHeaders`/`AddRow`/`Render`)                              | Fully                       | `table/` (`Table` type, `Render`)                                                                                                     | `hub/internal/cli/table.go:52-54,69-74,77-82`                                                                                                                                                        |
| 2  | ANSI display-width alignment (lipgloss-backed renderer — gotcha #298)                              | Fully                       | `table/` renderer internals; FEATURES.md "Rendering Infrastructure"                                                                   | `table.go:25-30` (comment names the lesson); regression-tested in `table_test.go` / `table_internal_test.go`                                                                                         |
| 3  | ColorMode plumbing (`ColorModeAuto/Never/Always`, single authority)                                | Fully                       | `go-output/color.go`; `table.WithColorMode`                                                                                           | `colors.go:9,32-48` (mode decided once, shared by cells + chrome via `colorMode()` at `colors.go:67-73`)                                                                                             |
| 4  | Color/env detection helpers (`IsNoColor`, `IsCI`)                                                  | Partial                     | `go-output/envdetect.go:20,32`                                                                                                        | `colors.go:33` hand-checks `os.Getenv("NO_COLOR")`; TTY detection hand-rolled via `colorprofile.Detect` (`colors.go:44`) instead of the library's envdetect                                          |
| 5  | Per-cell style functions (`StyleFunc`, `WithFooterStyle`)                                          | Partial                     | `table/table.go` (`StyleFunc(fn func(row, col int) lipgloss.Style)`)                                                                  | cells are pre-colored _strings_ via `colorService` (`table.go:67-74`, `colors.go:54-58`) — works, but couples styling into data cells instead of the render layer                                    |
| 6  | Footer/totals row (`SetFooter`/`Table.Footer`)                                                     | Missed                      | `table/` (`SetFooter` — bold-styled totals; root `Table.Footer` doc: rendered by tabular formats)                                     | `table.go:52-75` builds no footer — FINDINGS/RUNS/FAILS column totals are a one-call add                                                                                                             |
| 7  | Multi-format rendering (`output.RenderTable`, 16 `Format` values incl. csv/tsv/markdown/html/yaml) | Missed                      | `go-output/format.go:7-26` (16 formats); `RenderTable` doc "supports all registered formats when respective sub-modules are imported" | no `--format` flag exists (`flags.go` — QueryFlags only); `report.go:14-31` emits frozen v1 JSON via `compliance.MarshalV1` (stdlib — correct, separate concern)                                     |
| 8  | `TableProvider`/`FromTable` struct-driven tables                                                   | N/A                         | `table/` (`FromTable`, `FooterProvider`)                                                                                              | the CLI assembles rows from `BuildV1Report` output in one loop; variadic `AddRow` fits                                                                                                               |
| 9  | Streaming writers (`WriteRendered*`, `streaming.go`)                                               | N/A                         | `go-output/streaming.go`, `write_rendered.go`                                                                                         | one-shot renders to stdout; no streaming needs                                                                                                                                                       |
| 10 | Cell width caps/truncation                                                                         | N/A                         | not offered by `table/` (no max-width option in its API)                                                                              | `table.go:143-152` hand-rolls `truncateCell` (24-rune cap + ellipsis) — justified, the library has no equivalent                                                                                     |
| 11 | `markdown/` module (docs generation)                                                               | N/A (revisit)               | `markdown/`                                                                                                                           | nothing generates docs today; revisit for README fleet-status snapshots or refresh-report attachments                                                                                                |
| 12 | `tree/` module (hierarchical output)                                                               | N/A                         | `tree/`                                                                                                                               | no hierarchy in the fleet view (worst-steps lists live inside table cells)                                                                                                                           |
| 13 | Graph family (`graph/`, `d2/`, `plantuml/`, `daghtml/`, `delimited/`)                              | N/A                         | `go-output/graph.go`, `d2/`, …                                                                                                        | no graph/diagram surface; `delimited` is subsumed by row 7's `FormatCSV/TSV`                                                                                                                         |
| 14 | `serialization/` module                                                                            | N/A — banned here by policy | `serialization/` (go-faster/jx-based)                                                                                                 | STACK.md JSON policy: stdlib `encoding/json/v2` only, third-party JSON banned; the frozen v1 wire contract (`v1.go`, `MarshalV1`) correctly stays stdlib — this row is a policy alignment, not a gap |

## Gap analysis

### Row 7 — `--format` absent: 16 formats paid for, one used (Missed)

**Library offers:** `output.RenderTable(data *Table, format Format, opts
RenderOptions)` — one call, sixteen formats (table, json, csv, tsv, markdown,
xml, d2, yaml, html, tree, mermaid, dot, jsonl, asciidoc, toml, plantuml;
`format.go:7-26`), driven by `output.ParseFormat` for flag parsing and
`output.EnumAllowedValues(AllFormats, …)` for cmdguard-style `values:"…"` tags.
The format sub-modules the hub would import (`markdown`, `delimited`, `escape`)
are **already in the hub's module graph** as indirect requires via cmdguard →
samber-do-auditlog — promoting them to direct costs zero new dependency weight.

**Consumer does:** `table.go:31-86` renders the ANSI table only; `flags.go`
defines no format flag; the `report` command covers JSON via the frozen contract
(`report.go:23-25`) — which is right and must stay stdlib, but it is the _wire_
concern, not the _human_ one. There is no way to get the fleet view as CSV (for
spreadsheets), Markdown (for status posts/notes), or HTML (for a static
drop-in-page) from the tool that already computed it.

**Impact:** moderate. BuildFlow's `compliance` heritage (flag compatibility is
explicit — `flags.go:12-13` "matching BuildFlow compliance muscle memory") argues
for a _new_ flag, not a changed one: `--format` with default `table` changes no
existing invocation. Fleet-level: this is the single cheapest utilization win in
this audit — the data loop (`table.go:65-75`) already produces every cell.

**Sketch** (`hub/internal/cli/flags.go` + `table.go`):

```go
// skip-validate
// QueryFlags (or a TableFlags wrapper): one new field
Format string `default:"table" flag:"format" help:"Output format for the table command" values:"table,csv,tsv,markdown,html"`
```

```go
// runTable — after the existing row loop, branch once:
if flags.Format != "" && flags.Format != "table" {
	format, err := output.ParseFormat(flags.Format)
	if err != nil {
		return compliance.NewValidationError("invalid --format %q", flags.Format)
	}

	data := output.NewTable(headers) // same header slice as SetHeaders
	for _, row := range allRows {    // collect [][]string next to the existing AddRow loop
		data.AddRow(row)
	}

	return output.RenderTable(data, format, output.RenderOptions{
		Title:  "Repo compliance",
		Writer: out,
	})
}
```

Note: ANSI-colored cell strings must be bypassed for non-terminal formats —
build `allRows` from the plain values and keep `colorService` for the `table`
branch only (the color decision already lives in one place, `colors.go:32-48`).

### Row 6 — no totals footer (Missed)

**Library offers:** `SetFooter(row ...string)` — a bold-styled totals row
(`table/table.go`); the root `Table.Footer` carries into csv/tsv/markdown/html
renders too (`Table` doc).

**Consumer does:** nothing — yet the table's own columns (FINDINGS, RUNS, FAILS)
are countable aggregates, and a fleet triage view that ends with
"FINDINGS 412 · RUNS 1 893 · FAILS 96" answers "how bad is it overall" without a
second command.

**Sketch:** alongside the row loop, accumulate `sumFindings/sumRuns/sumFails`,
then `tbl.SetFooter("", "", "", "", strconv.Itoa(sumFindings), …)`. Three lines
plus the accumulators.

### Row 4 — env detection hand-rolled around the library (Partial)

**Library offers:** `output.IsNoColor()` / `output.IsCI()`
(`envdetect.go:20,32`).

**Consumer does:** `colors.go:33` checks `os.Getenv("NO_COLOR")` directly,
and detects TTY via `colorprofile.Detect(os.Stdout, os.Environ())`
(`colors.go:44`). The colorprofile route is _more_ capable than the
library's (it handles termcap profiles, not just on/off), so this is a
half-duplicate: NO_COLOR should defer to `output.IsNoColor()` (one call, fleet
semantics kept in one place), while the colorprofile detection is a justified
local refinement. Low stakes either way.

### Row 5 — styling lives in cell strings, not the render layer (Partial)

**Library offers:** `StyleFunc(fn func(row, col int) lipgloss.Style)` — style at
render time, keyed by coordinates.

**Consumer does:** pre-colors cell _values_ (`stateLabel`, `severityCells` —
`table.go:90-131`) and relies on the width-aware renderer to measure ANSI
correctly. This works (that's the whole gotcha-#298 point) and keeps the domain
→ string mapping simple, but it means the same table rendered via
`--format csv|markdown` (row 7) would need the plain-value path anyway. If row 7
lands with the two-branch structure sketched above, the current shape is fine;
`StyleFunc` becomes the refactor target only if the colored/plain split starts
duplicating row data.

## Version currency

- Pin v0.38.7 == newest root-line tag → **current**. The repo-wide newest tag
  `bdd/v0.38.7` is the same release family (19 modules cut in lockstep; the
  CHANGELOG records the v0.38.7 release discipline: tag-push limits, proxy
  verification).
- HEAD beyond the tag: `90fb2a6` + chores (vendor hash, dependency refresh,
  release scripts) — no consumer-visible API change (`CHANGELOG [Unreleased]`:
  "Nothing yet").
- Upgrade risk: none pending. Next 0.38.x or 0.39 should be a routine
  `go get github.com/larsartmann/go-output@latest
  github.com/larsartmann/go-output/table@latest` + tidy (GOWORK=off in `hub/`).

## Fleet leverage

**Expand.** 51 consumers per STACK.md; this repo proves the width-aware table
rationale but stops at one command. The expansion case is unusually strong
because the marginal cost is near zero: the format sub-modules are already in
the module graph (paid for by cmdguard), the data loop already exists, and the
fleet convention (BuildFlow muscle memory) is flag-compatible by design. A
`--format markdown` fleet table is also the natural input to status posts —
the workflow this very repo's `reports/` directory suggests exists. **Leave**
`serialization` unused (policy-banned, and the frozen contract has its own
golden tests), **leave** the graph family (no surface), **revisit** `markdown`
via row 7 (which subsumes it) rather than as a separate adoption.

## Ranked opportunities (impact × ease)

1. **`--format` flag on `table` via `output.RenderTable`** (row 7) — high
   impact (16 formats, fleet convention, zero new deps), high ease (~25 lines
   - one flag + tests).
2. **Totals footer via `SetFooter`** (row 6) — moderate impact (triage summary),
   trivial ease.
3. **`output.IsNoColor()` instead of the direct env check** (row 4) — small
   consistency win, trivial ease; do it in the same PR as #1 (both touch
   `colors.go`/`table.go`). One-line change at `colors.go:33`.
4. **README fleet-status snapshot via `--format markdown`** — only after #1;
   documentation value, zero code.

## How to verify (executed 2026-11-19)

```bash
# Library metadata
git -C /home/lars/projects/go-output tag | grep '^v' | sort -V | tail -1   # v0.38.7
git -C /home/lars/projects/go-output describe --tags --abbrev=0            # bdd/v0.38.7
git -C /home/lars/projects/go-output log -1 --format='%h %cs %s'           # 90fb2a6 2026-11-17 fix(nix): update vendor hash...
git -C /home/lars/projects/go-output tag --sort=-creatordate | head -5     # bdd/v0.38.7, d2/v0.38.7, daghtml/v0.38.7, delimited/v0.38.7, escape/v0.38.7
git -C /home/lars/projects/go-output log v0.38.7..HEAD --oneline | head    # chores only

# Format surface + table API
cd /home/lars/projects/go-output && GOWORK=off go doc . RenderTable        # "all 16 formats ... when respective sub-modules are imported"
cd /home/lars/projects/go-output && GOWORK=off go doc . Format | head -30
cd /home/lars/projects/go-output/table && GOWORK=off go doc -all . | grep -E '^func |^type '

# Effective pins + the indirect chain (why the sub-modules are "free")
cd /home/lars/projects/go-compliance/hub && GOWORK=off go list -m all | grep go-output
cd /home/lars/projects/go-compliance/hub && GOWORK=off go mod why github.com/larsartmann/go-output/tree
#   hub/cmd/compliance-hub -> cmdguard/v5 -> samber-do-auditlog -> go-output/tree

# Consumer usage sites (all citations in this report)
cd /home/lars/projects/go-compliance
rg -n 'go-output' hub -g '*.go' | grep -v _test
sed -n '31,86p' hub/internal/cli/table.go
sed -n '9,58p' hub/internal/cli/colors.go
sed -n '14,31p' hub/internal/cli/report.go        # frozen v1 JSON — stdlib, by policy
rg -n 'format|Format' hub/internal/cli/flags.go hub/internal/cli/app.go   # (no --format flag exists)

# Policy anchor for the serialization N/A
grep -n 'JSON' STACK.md
```

All file:line citations spot-checked against the working tree on 2026-11-19.

package cli

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/dustin/go-humanize"

	"github.com/larsartmann/go-output"
	_ "github.com/larsartmann/go-output/delimited" // csv, tsv
	_ "github.com/larsartmann/go-output/markdown"
	_ "github.com/larsartmann/go-output/markup"        // html, xml, asciidoc
	_ "github.com/larsartmann/go-output/serialization" // json, jsonl, yaml, toml
	gotable "github.com/larsartmann/go-output/table"

	"github.com/larsartmann/go-compliance"
)

// maxRepoCellWidth bounds the REPO column; longer names truncate with an
// ellipsis so one long path cannot blow up the whole table.
const maxRepoCellWidth = 24

// tableHeaders orders the fleet table columns; shared by the styled TTY
// path and the machine-format path so both views stay column-identical.
//
//nolint:gochecknoglobals // immutable column metadata, effectively a const
var tableHeaders = []string{
	"REPO", "STATE", "ERR", "WARN", "INFO", "FINDINGS", "RUNS", "FAILS", "RATE", "VERDICT", "LAST RUN",
}

// runTable renders the human fleet view: every repo, its state, its
// latest-run finding severities (ERR/WARN/INFO), window counters, and
// last verdict. Failing repos sort first - the table is the triage view.
//
// The table renders through go-output/table (lipgloss-backed) instead of
// hand-rolled Printf width verbs: state/severity cells carry ANSI color
// and Printf pads on byte length, so every colored cell would shift all
// later columns on a TTY. lipgloss measures ANSI display width, so
// colored and plain cells align identically (the gotcha #298 lesson,
// regression-tested in table_test.go).
//
// --format selects go-output's RenderTable registry instead (json/csv/
// markdown/html/...): machine formats get full untruncated repo names,
// absolute timestamps, and zero ANSI — data, not chrome. cmdguard's own
// --output flag stays CLI-global (error envelopes) because the command
// handler contract does not expose it; --format is the command-local
// twin for the payload itself.
func runTable(out, _ io.Writer) func(context.Context, *Config, TableFlags) error {
	return func(ctx context.Context, _ *Config, flags TableFlags) error {
		evidence, since, staleAfter, err := queryEvidence(ctx, flags.QueryFlags)
		if err != nil {
			return err
		}

		format, err := parseTableFormat(flags.Format)
		if err != nil {
			return err
		}

		order, err := parseSortBy(flags.SortBy)
		if err != nil {
			return err
		}

		if format != output.FormatTable {
			return renderTableData(out, format, evidence, since, staleAfter, order)
		}

		colors := newColorService(output.ColorModeAuto)

		fmt.Fprintf(out, "Fleet compliance: %d repos, since %s (%s), stale after %s\n\n",
			len(evidence), since.Format("2006-01-02 15:04"),
			humanize.Time(since), staleAfter)

		if len(evidence) == 0 {
			fmt.Fprintln(out, "No evidence - the fleet has no runs recorded yet.")

			return nil
		}

		fmt.Fprintf(out, "%s\n\n", colors.Header("Repo compliance"))

		tbl := gotable.New(gotable.WithColorMode(colors.colorMode())).SetHeaders(tableHeaders...)

		// Worst-first base order: the report builder's sort (state rank,
		// then failures desc, then name) is the triage invariant;
		// --sort-by only re-tiebreaks INSIDE the state rank.
		report := compliance.BuildV1Report(evidence, since, staleAfter, Version, time.Now())
		sortReposBy(report.Repos, order, findingTotalsByRepo(evidence))

		severitiesByRepo := make(map[string]map[string]int, len(evidence))
		for _, record := range evidence {
			severitiesByRepo[record.Repo.Get()] = record.FindingSeverities
		}

		for _, row := range report.Repos {
			rate := strconv.FormatFloat(row.FailRate*100, 'f', 0, 64) + "%"
			errs, warns, infos, total := severityCells(colors, severitiesByRepo[row.Repo])

			tbl.AddRow(
				truncateCell(row.Repo, maxRepoCellWidth), stateLabel(colors, compliance.State(row.State)),
				errs, warns, infos, total,
				strconv.FormatInt(row.Runs, 10), strconv.FormatInt(row.Failures, 10), rate,
				row.LastRunVerdict, shortAge(compliance.ParseWireTime(row.LastRunAt)),
			)
		}

		rendered, err := tbl.Render()
		if err != nil {
			return compliance.NewInternalError(err, "failed to render the compliance table")
		}

		fmt.Fprintf(out, "%s\n", rendered)

		return nil
	}
}

// supportedTableFormats is the registry-backed set table.go can render:
// one entry per imported go-output sub-module's self-registered formats.
//
//nolint:gochecknoglobals // immutable format metadata, effectively a const
var supportedTableFormats = map[output.Format]struct{}{
	output.FormatTable:    {},
	output.FormatJSON:     {},
	output.FormatJSONL:    {},
	output.FormatYAML:     {},
	output.FormatTOML:     {},
	output.FormatCSV:      {},
	output.FormatTSV:      {},
	output.FormatMarkdown: {},
	output.FormatHTML:     {},
	output.FormatXML:      {},
	output.FormatAsciiDoc: {},
}

// parseTableFormat validates the --format flag against the formats the
// imported registry actually serves (cmdguard's values tag guards the CLI;
// this guards embedders calling runTable directly).
func parseTableFormat(raw string) (output.Format, error) {
	format, err := output.ParseFormat(raw)
	if err != nil {
		return "", compliance.NewValidationError(
			"invalid --format value %q (supported: table, json, jsonl, yaml, toml, csv, tsv, markdown, html, xml, asciidoc)",
			raw,
		)
	}

	if _, ok := supportedTableFormats[format]; !ok {
		return "", compliance.NewValidationError(
			"--format %q is not wired into the fleet table (supported: table, json, jsonl, yaml, toml, csv, tsv, markdown, html, xml, asciidoc)",
			raw,
		)
	}

	return format, nil
}

// renderTableData renders the fleet table through go-output's RenderTable
// registry for machine consumption: full repo names (no truncation), RFC
// 3339 timestamps, plain cells (no ANSI), same columns and worst-first
// order as the styled view.
func renderTableData(
	out io.Writer, format output.Format, evidence []compliance.RepoEvidence,
	since time.Time, staleAfter time.Duration, order sortBy,
) error {
	report := compliance.BuildV1Report(evidence, since, staleAfter, Version, time.Now())
	sortReposBy(report.Repos, order, findingTotalsByRepo(evidence))

	severitiesByRepo := make(map[string]map[string]int, len(evidence))
	for _, record := range evidence {
		severitiesByRepo[record.Repo.Get()] = record.FindingSeverities
	}

	data := output.NewTable(tableHeaders)

	for _, row := range report.Repos {
		rate := strconv.FormatFloat(row.FailRate*100, 'f', 0, 64) + "%"
		errs, warns, infos, total := plainSeverityCells(severitiesByRepo[row.Repo])

		lastRun := compliance.ParseWireTime(row.LastRunAt)
		lastRunCell := "-"

		if !lastRun.IsZero() {
			lastRunCell = lastRun.UTC().Format(time.RFC3339)
		}

		data.AddRow([]string{
			row.Repo, row.State, errs, warns, infos, total,
			strconv.FormatInt(row.Runs, 10), strconv.FormatInt(row.Failures, 10), rate,
			row.LastRunVerdict, lastRunCell,
		})
	}

	return output.RenderTable(data, format, output.RenderOptions{
		Title:     "Fleet compliance",
		Writer:    out,
		ColorMode: output.ColorModeNever,
	})
}

// plainSeverityCells is severityCells without color: raw counts, "-" for
// empty severities, identical totals.
func plainSeverityCells(severities map[string]int) (string, string, string, string) {
	cell := func(n int) string {
		if n <= 0 {
			return "-"
		}

		return strconv.Itoa(n)
	}

	total := 0
	for _, n := range severities {
		total += n
	}

	totalCell := "-"
	if total > 0 {
		totalCell = strconv.Itoa(total)
	}

	return cell(severities["error"]), cell(severities["warning"]), cell(severities["info"]), totalCell
}

// stateLabel colors the state cell: failing red, stale amber, compliant
// green, never-covered gray-blue; unknown states render verbatim.
func stateLabel(colors *colorService, state compliance.State) string {
	switch state {
	case compliance.StateFailing:
		return colors.Error("failing")
	case compliance.StateStale:
		return colors.Warning("stale")
	case compliance.StateCompliant:
		return colors.Success("compliant")
	case compliance.StateNeverCovered:
		return colors.Info("never-covered")
	default:
		return state.String()
	}
}

// severityCells renders the ERR/WARN/INFO/FINDINGS cells: counts when
// present, "-" when a severity carries nothing. The total counts every
// finding the latest run recorded, severity known or not.
func severityCells(colors *colorService, severities map[string]int) (string, string, string, string) {
	cell := func(n int, color func(string) string) string {
		if n <= 0 {
			return "-"
		}

		return color(strconv.Itoa(n))
	}

	total := 0
	for _, n := range severities {
		total += n
	}

	totalCell := "-"
	if total > 0 {
		totalCell = strconv.Itoa(total)
	}

	return cell(severities["error"], colors.Error),
		cell(severities["warning"], colors.Warning),
		cell(severities["info"], colors.Info),
		totalCell
}

// shortAge renders a last-run time as a compact age.
func shortAge(t time.Time) string {
	if t.IsZero() {
		return "-"
	}

	return humanize.Time(t)
}

// truncateCell shortens s to maxLen runes, appending the ellipsis when it
// had to cut (fleet cell convention).
func truncateCell(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}

	runes := []rune(s)

	return string(runes[:maxLen-1]) + "…"
}

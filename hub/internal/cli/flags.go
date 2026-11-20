package cli

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/larsartmann/go-compliance"
	"github.com/larsartmann/go-compliance/hub/internal/cache"
)

// QueryFlags are the evidence-query flags shared by report and table,
// matching BuildFlow compliance muscle memory.
type QueryFlags struct {
	Since      string `default:"7d"                                flag:"since"       help:"Evidence window (e.g. 24h, 7d, 30d)"`
	StaleAfter string `default:"336h"                              flag:"stale-after" help:"Time since a repo's last run after which it counts as stale (e.g. 168h, 336h, 14d)"`
	WorstSteps int    `default:"3"                                 flag:"worst-steps" help:"Worst steps listed per repository"`
	Step       string `default:""                                  flag:"step"        help:"Scope the view to ONE step (EXACT name; fan-out keys count)"`
	Source     string `default:"file"                              flag:"source"      help:"Evidence source"                                                                    values:"file,buildflow,cache"`
	File       string `default:"/var/lib/overview/compliance.json" flag:"file"        help:"Piped v1 evidence document (source=file)"`
	DB         string `default:""                                  flag:"db"          help:"SQLite evidence cache path (source=cache)"`
}

// ServeFlags are the hub server flags. The bind default is localhost
// only: fleet evidence stays local until the auth posture is decided
// (docs/decisions/bind-posture.md).
type ServeFlags struct {
	QueryFlags

	Addr string `default:"127.0.0.1:8477" flag:"addr" help:"Listen address (localhost-only default by policy)"`
}

// TableFlags extends the query flags with the table command's output
// format and the --sort-by tiebreak. The supported format set is exactly
// the formats whose go-output sub-modules table.go imports for
// self-registration; --sort-by mirrors the producer's flag semantics
// (tiebreak INSIDE the state rank, state always sorts first).
type TableFlags struct {
	QueryFlags

	Format string `default:"table" flag:"format" help:"Output format for the fleet table" values:"table,json,jsonl,yaml,toml,csv,tsv,markdown,html,xml,asciidoc"`
	SortBy string `default:"failures" flag:"sort-by" help:"Tiebreak inside the state rank (state always sorts first): findings, failures, name" values:"findings,failures,name"`
}

// resolveSource builds the Source the flags select. The cache store is
// returned alongside so long-running callers (serve) can close it on
// shutdown; one-shot callers ignore it.
//
//nolint:ireturn // compliance.Source IS the port; callers depend on the interface by design
func resolveSource(ctx context.Context, flags QueryFlags) (compliance.Source, io.Closer, error) {
	switch flags.Source {
	case "", "file":
		if flags.File == "" {
			return nil, nil, compliance.NewValidationError("--file must not be empty for --source file")
		}

		return compliance.NewFileSource(flags.File), nil, nil
	case "buildflow":
		return compliance.NewBuildFlowCommandSource(), nil, nil
	case "cache":
		if flags.DB == "" {
			return nil, nil, compliance.NewValidationError("--db must not be empty for --source cache")
		}

		store, err := cache.Open(ctx, flags.DB)
		if err != nil {
			return nil, nil, fmt.Errorf("open cache %s: %w", flags.DB, err)
		}

		return store, store, nil
	default:
		return nil, nil, compliance.NewValidationError(
			"invalid --source value %q (supported: file, buildflow, cache)", flags.Source)
	}
}

// parseSinceDuration parses a shorthand duration string like "7d", "24h",
// "30d" and returns the cutoff time (now - duration). Port of BuildFlow's
// timings_cmd flag helper, kept flag-compatible.
func parseSinceDuration(raw string) (time.Time, error) {
	window, err := parseSinceWindow(raw)
	if err != nil {
		return time.Time{}, err
	}

	return time.Now().Add(-window), nil
}

// parseSinceWindow parses the same shorthand into the window LENGTH: the
// live serve loop re-derives with since = now - window on every refresh.
func parseSinceWindow(raw string) (time.Duration, error) {
	if raw == "" {
		return 0, nil
	}

	d, err := parseDurationSupportingDays(raw)
	if err != nil {
		return 0, compliance.NewValidationError(
			"invalid --since value: %s (supported: 24h, 7d, 30d, 2h45m)", raw)
	}

	return d, nil
}

// parseStaleAfter accepts Go duration syntax (336h) plus the d-suffix the
// fleet uses (14d); must be positive.
func parseStaleAfter(raw string) (time.Duration, error) {
	d, err := parseDurationSupportingDays(raw)
	if err != nil || d <= 0 {
		return 0, compliance.NewValidationError(
			"invalid --stale-after value: %s (supported: 168h, 336h, 14d)", raw)
	}

	return d, nil
}

// parseDurationSupportingDays parses Go duration syntax plus the Nd day
// suffix fleet style (7d, 30d), converting days to hours so a single
// ParseDuration call serves flags that accept both spellings.
func parseDurationSupportingDays(raw string) (time.Duration, error) {
	durationStr := raw

	if days, ok := parseDaysSuffix(raw); ok {
		durationStr = fmt.Sprintf("%dh", days*24)
	}

	d, err := time.ParseDuration(durationStr)
	if err != nil {
		return 0, fmt.Errorf("parse duration %q: %w", raw, err)
	}

	return d, nil
}

// parseDaysSuffix extracts a day count from strings like "7d", "30d".
func parseDaysSuffix(s string) (int, bool) {
	if !strings.HasSuffix(s, "d") {
		return 0, false
	}

	n, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
	if err != nil || n <= 0 {
		return 0, false
	}

	return n, true
}

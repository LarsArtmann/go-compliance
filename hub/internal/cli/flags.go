package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/larsartmann/go-compliance"
)

// QueryFlags are the evidence-query flags shared by report and table,
// matching BuildFlow compliance muscle memory.
type QueryFlags struct {
	Since string `flag:"since" default:"7d" help:"Evidence window (e.g. 24h, 7d, 30d)"`
	StaleAfter string `flag:"stale-after" default:"336h" help:"Time since a repo's last run after which it counts as stale (e.g. 168h, 336h, 14d)"`
	WorstSteps int    `flag:"worst-steps" default:"3" help:"Worst steps listed per repository"`
	Step       string `flag:"step" default:"" help:"Scope the view to ONE step (EXACT name; fan-out keys count)"`
	Source     string `flag:"source" default:"file" help:"Evidence source" values:"file,buildflow"`
	File       string `flag:"file" default:"/var/lib/overview/compliance.json" help:"Piped v1 evidence document (source=file)"`
}

// ServeFlags are the hub server flags. The bind default is localhost
// only: fleet evidence stays local until the auth posture is decided
// (docs/decisions/bind-posture.md).
type ServeFlags struct {
	Addr string `flag:"addr" default:"127.0.0.1:8477" help:"Listen address (localhost-only default by policy)"`
	QueryFlags
}

// resolveSource builds the Source the flags select.
func resolveSource(flags QueryFlags) (compliance.Source, error) {
	switch flags.Source {
	case "", "file":
		if flags.File == "" {
			return nil, compliance.NewValidationError("--file must not be empty for --source file")
		}

		return compliance.NewFileSource(flags.File), nil
	case "buildflow":
		return compliance.NewBuildFlowCommandSource(), nil
	default:
		return nil, compliance.NewValidationError(
			"invalid --source value %q (supported: file, buildflow)", flags.Source)
	}
}

// parseSinceDuration parses a shorthand duration string like "7d", "24h",
// "30d" and returns the cutoff time (now - duration). Port of BuildFlow's
// timings_cmd flag helper, kept flag-compatible.
func parseSinceDuration(raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, nil
	}

	d, err := parseDurationSupportingDays(raw)
	if err != nil {
		return time.Time{}, compliance.NewValidationError(
			"invalid --since value: "+raw+" (supported: 24h, 7d, 30d, 2h45m)")
	}

	return time.Now().Add(-d), nil
}

// parseStaleAfter accepts Go duration syntax (336h) plus the d-suffix the
// fleet uses (14d); must be positive.
func parseStaleAfter(raw string) (time.Duration, error) {
	d, err := parseDurationSupportingDays(raw)
	if err != nil || d <= 0 {
		return 0, compliance.NewValidationError(
			"invalid --stale-after value: "+raw+" (supported: 168h, 336h, 14d)")
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

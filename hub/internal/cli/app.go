// Package cli is the compliance-hub command line: report (the v1 JSON
// producer), table (the human fleet view), and serve (the realtime hub).
package cli

import (
	"context"
	"fmt"
	"io"
	"time"

	v5 "github.com/larsartmann/cmdguard/v5/pkg/cmdguard/v5"

	"github.com/larsartmann/go-compliance"
)

// Version is the hub binary version (overridden at release time).
var Version = "dev"

// Config is the hub's global configuration (empty for now; config-file
// and env knobs land with the service hardening pass).
type Config struct{}

// buildApp composes the production CLI. out and errW are injectable for
// tests; production passes os.Stdout/os.Stderr.
func buildApp(out, errW io.Writer) (*v5.CLI[Config], error) {
	cli, err := v5.NewCLI[Config](
		"compliance-hub", "Fleet compliance evidence: report, table, and the realtime hub",
		Config{},
		v5.WithCLIVersion(Version),
		v5.WithEnvPrefix("COMPLIANCE_HUB_"),
	)
	if err != nil {
		return nil, fmt.Errorf("build compliance-hub CLI: %w", err)
	}

	reportCmd, err := v5.NewCommand("report", QueryFlags{}, runReport(out, errW),
		v5.WithShort("Emit the frozen v1 JSON contract (stdout)"),
	)
	if err != nil {
		return nil, err
	}

	tableCmd, err := v5.NewCommand("table", QueryFlags{}, runTable(out, errW),
		v5.WithShort("Render the human fleet compliance table"),
	)
	if err != nil {
		return nil, err
	}

	serveCmd, err := v5.NewCommand("serve", ServeFlags{}, runServe(out, errW),
		v5.WithShort("Run the realtime hub server"),
	)
	if err != nil {
		return nil, err
	}

	if err := v5.AddCommands(cli, reportCmd, tableCmd, serveCmd); err != nil {
		return nil, fmt.Errorf("register commands: %w", err)
	}

	return cli, nil
}

// Run executes the CLI with the given args (split from main for tests).
func Run(ctx context.Context, args []string, out, errW io.Writer) error {
	cli, err := buildApp(out, errW)
	if err != nil {
		return err
	}

	return cli.ExecuteWithArgs(ctx, args)
}

// queryEvidence resolves the configured source and queries it, mapping
// flag values onto the shared options. It also returns the parsed window
// (since, staleAfter) so handlers never re-parse flags.
func queryEvidence(
	ctx context.Context, flags QueryFlags,
) ([]compliance.RepoEvidence, time.Time, time.Duration, error) {
	since, err := parseSinceDuration(flags.Since)
	if err != nil {
		return nil, time.Time{}, 0, err
	}

	staleAfter, err := parseStaleAfter(flags.StaleAfter)
	if err != nil {
		return nil, time.Time{}, 0, err
	}

	source, closer, err := resolveSource(ctx, flags)
	if err != nil {
		return nil, time.Time{}, 0, err
	}

	if closer != nil {
		defer func() { _ = closer.Close() }()
	}

	if err != nil {
		return nil, time.Time{}, 0, err
	}

	evidence, err := source.QueryEvidence(ctx, since, flags.WorstSteps, flags.Step)
	if err != nil {
		return nil, time.Time{}, 0, fmt.Errorf("source %s: %w", source.Name().Get(), err)
	}

	return evidence, since, staleAfter, nil
}

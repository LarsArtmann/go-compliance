package cli

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/larsartmann/go-compliance"
)

// runTable renders the human fleet view (the triage surface). The M7
// plain rendering lands the semantics; M8 upgrades the cells to the
// go-output fleet table (severity cells, state labels, worst-first sort).
func runTable(out, _ io.Writer) func(context.Context, *Config, QueryFlags) error {
	return func(ctx context.Context, _ *Config, flags QueryFlags) error {
		evidence, since, staleAfter, err := queryEvidence(ctx, flags)
		if err != nil {
			return err
		}

		fmt.Fprintf(out, "Fleet compliance: %d repos, since %s, stale after %s\n\n",
			len(evidence), since.Format("2006-01-02 15:04"), staleAfter)

		if len(evidence) == 0 {
			fmt.Fprintln(out, "No evidence - the fleet has no runs recorded yet.")

			return nil
		}

		w := tabwriter.NewWriter(out, 0, 2, 2, ' ', 0)

		fmt.Fprintln(w, "REPO\tSTATE\tRUNS\tFAILS\tRATE\tVERDICT\tLAST RUN")

		for _, record := range evidence {
			state := compliance.DeriveState(record, time.Now(), staleAfter)
			fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%.0f%%\t%s\t%s\n",
				record.Repo.Get(), state,
				record.Runs, record.Failures, record.FailRate*100,
				record.LastRunVerdict,
				shortTime(record.LastRunAt))
		}

		return w.Flush()
	}
}

// shortTime renders a last-run time as a compact age.
func shortTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}

	return time.Since(t).Round(time.Hour).String() + " ago"
}

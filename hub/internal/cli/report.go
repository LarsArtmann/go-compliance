package cli

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/larsartmann/go-compliance"
)

// runReport emits the frozen v1 JSON contract to stdout - the producer
// command, pipe-for-pipe compatible with `buildflow compliance --json`.
func runReport(out, _ io.Writer) func(context.Context, *Config, QueryFlags) error {
	return func(ctx context.Context, _ *Config, flags QueryFlags) error {
		evidence, since, staleAfter, err := queryEvidence(ctx, flags)
		if err != nil {
			return err
		}

		report := compliance.BuildV1Report(evidence, since, staleAfter, Version, time.Now())

		encoded, err := compliance.MarshalV1(report)
		if err != nil {
			return err
		}

		_, err = fmt.Fprintln(out, string(encoded))

		return err
	}
}

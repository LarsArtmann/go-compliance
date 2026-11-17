// Command compliance-hub is the realtime fleet compliance hub binary.
package main

import (
	"context"
	"os"

	v5 "github.com/larsartmann/cmdguard/v5/pkg/cmdguard/v5"

	"github.com/larsartmann/go-compliance/hub/internal/cli"
)

func main() {
	os.Exit(v5.ExitCode(run(context.Background(), os.Args[1:])))
}

func run(ctx context.Context, args []string) error {
	return cli.Run(ctx, args, os.Stdout, os.Stderr)
}

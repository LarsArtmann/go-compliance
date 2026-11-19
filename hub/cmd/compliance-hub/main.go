// Command compliance-hub is the realtime fleet compliance hub binary.
package main

import (
	"context"
	"os"

	errorfamily "github.com/larsartmann/go-error-family"

	"github.com/larsartmann/go-compliance/hub/internal/cli"
)

func main() {
	// errorfamily.ExitCode (not cmdguard's): it classifies compliance errors
	// through their errorfamily integration — transient cache failures exit
	// 75 (EX_TEMPFAIL), validation 64/65, internal 70 — so systemd restart
	// semantics and operator muscle memory get the fleet-standard codes.
	os.Exit(errorfamily.ExitCode(run(context.Background(), os.Args[1:])))
}

func run(ctx context.Context, args []string) error {
	return cli.Run(ctx, args, os.Stdout, os.Stderr)
}

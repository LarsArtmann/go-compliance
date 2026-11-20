# cmdguard `AssertDescribeContract` adoption sketch (go-compliance hub)

**Status:** patch-ready, NOT applied — gated on the next cmdguard tag that
publishes `pkg/testutil` (the helper lives at `v5.1.0-33`, 33 commits ahead
of `v5.1.0` as of 2026-11-20; `git describe` in `~/projects/cmdguard`).
**Flip trigger:** the next cmdguard tag (watch: `git ls-remote --tags
https://github.com/LarsArtmann/cmdguard`).

## What it buys

One pinned test fails the hub build the moment the CLI's machine contract
drifts: every command's flags, env names, defaults, required/secret
markers, and the schema version — the same describe JSON the hidden
`describe` command emits. The hub's fleet-table/serve surface becomes a
reviewed artifact: flag additions, renames, and default changes show up as
a reviewable golden diff instead of silent drift.

## The patch (after the tag lands)

1. `cd hub && go get github.com/larsartmann/cmdguard/v5@<next-tag> && GOWORK=off go mod tidy`
2. New file `hub/internal/cli/describe_contract_internal_test.go`:

```go
package cli

import (
	"testing"

	v5test "github.com/larsartmann/cmdguard/v5/pkg/testutil"
)

func init() { v5test.RegisterUpdateGoldensFlag() }

func TestDescribeContract(t *testing.T) {
	cli, err := buildApp(io.Discard, io.Discard)
	if err != nil {
		t.Fatalf("buildApp: %v", err)
	}
	v5test.AssertDescribeContract(t, cli, "testdata/describe.json")
}
```

Internal test file on purpose: `buildApp` is unexported and is the only
handle on the composed `*v5.CLI[Config]` (`app.go:27`); the external
`cli_test` package has no hook to reach it. `RegisterUpdateGoldensFlag`
is double-registration-safe (no-ops when another package in the binary
already owns the flag).

3. Record the golden: `cd hub && GOWORK=off go test ./internal/cli/
   -run TestDescribeContract -update-golden`, commit
   `hub/internal/cli/testdata/describe.json` with the test.
4. No CI wiring needed — the repo's per-module `go test ./...` job already
   runs every internal test; the golden diff IS the gate.
5. `go.sum`/`vendorHash` ritual: hub `vendorHash` re-bootstrap at BOTH
   flake sites after the dep bump (AGENTS "Nix / FOD Lessons").

## Maintenance rule

A CLI-shape change (new flag, renamed command, changed default) must
regenerate the golden in the SAME commit: `go test ./internal/cli/ -run
TestDescribeContract -update-golden`. The diff then documents the contract
change for reviewers instead of hiding it.

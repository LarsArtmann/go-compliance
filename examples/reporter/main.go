// Command reporter demonstrates the one-line reporter integration: a tool
// finishes a run, records evidence, done. Run it, then inspect the JSON
// document it wrote — the same shape `compliance-hub serve --source file`
// consumes.
//
//	go run ./examples/reporter [evidence-file.json]
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/larsartmann/go-compliance"
)

func main() {
	path := "compliance-example.json"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	ctx := context.Background()
	store := compliance.NewFileStore(path)

	// The integration: your runner's last step records the run. A clean
	// window (no failures) derives compliant.
	err := compliance.RecordTo(ctx, store, compliance.NewRepoName("my-repo"),
		compliance.RunResult{
			Verdict:      compliance.VerdictPassed,
			Runs:         3,
			StepRows:     42,
			GateFindings: map[string]int{"lint": 2},
		})
	if err != nil {
		log.Fatalf("record my-repo: %v", err)
	}

	fmt.Printf("recorded my-repo (passed) -> %s\n", path)

	// A failing repo lands in the same document; the hub derives the
	// failing state from it. Note the honesty rule: ANY failure in the
	// window keeps a repo failing, even when the latest verdict passed.
	err = compliance.RecordTo(ctx, store, compliance.NewRepoName("flaky-repo"),
		compliance.RunResult{
			Verdict:  compliance.VerdictFailed,
			Runs:     2,
			StepRows: 10,
			Failures: 5,
		})
	if err != nil {
		log.Fatalf("record flaky-repo: %v", err)
	}

	fmt.Printf("recorded flaky-repo (failed) -> %s\n", path)

	doc, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("read back evidence: %v", err)
	}

	fmt.Printf("\n%s\n", doc)
}

package compliance_test

import (
	"context"
	"fmt"
	"os"

	"github.com/larsartmann/go-compliance"
)

// ExampleRecord shows the one-liner integration: report one repository's
// run evidence into the fleet document. Point
// GO_COMPLIANCE_EVIDENCE_FILE at your evidence sink (a CI step, a
// post-build hook, a timer) and every dashboard consuming the v1 contract
// sees the repository.
func ExampleRecord() {
	dir, _ := os.MkdirTemp("", "compliance-example-*")
	defer func() { _ = os.RemoveAll(dir) }()

	_ = os.Setenv(compliance.EnvEvidenceFile, dir+"/compliance.json")
	defer func() { _ = os.Unsetenv(compliance.EnvEvidenceFile) }()

	err := compliance.Record(context.Background(), compliance.NewRepoName("my-repo"),
		compliance.RunResult{
			Verdict:         compliance.VerdictPassed,
			GateFindings:    map[string]int{"lint": 2},
			ProducerVersion: "v1.4.2",
		})
	fmt.Println("recorded:", err)

	report, _ := compliance.NewFileStore(dir + "/compliance.json").Load(context.Background())
	fmt.Println("repos:", len(report.Repos), "state:", report.Repos[0].State)
	// Output:
	// recorded: <nil>
	// repos: 1 state: compliant
}

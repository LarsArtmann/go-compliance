package compliance_test

import (
	"os"
	"testing"
	"time"

	"github.com/larsartmann/go-compliance"
)

// Benchmarks pin the fleet-scale query path: the fixture carries the real
// 341-repo document shape (regenerable via scripts/refresh-fixture.sh).

const benchFixture = "testdata/buildflow-v1-fixture.json"

// benchFoldInput builds a synthetic FoldInput at fleet scale: every repo
// carries window health, a latest run, gate findings, severities, and two
// worst steps — the heaviest legal shape a source can produce.
func benchFoldInput(n int) compliance.FoldInput {
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	input := compliance.FoldInput{WorstStepLimit: 3}

	for i := range n {
		repo := compliance.NewRepoName("repo-" + itoa4(i))

		input.WindowHealth = append(input.WindowHealth, compliance.WindowHealth{
			Repo: repo, Runs: 12, StepRows: 240, Failures: int64(i % 7), FailRate: 0.03,
		})
		input.LatestRuns = append(input.LatestRuns, compliance.LatestRun{
			Repo: repo, LastRunAt: now.Add(-time.Duration(i%200) * time.Hour),
			AnyFailed: i%3 == 0, HasVerdict: true, Version: "v1.2.3",
		})
		input.GateFindings = append(input.GateFindings, compliance.RepoFindings{
			Repo: repo, Findings: map[string]int{"lint": 2, "test": 1},
		})
		input.FindingSeverities = append(input.FindingSeverities, compliance.RepoSeverities{
			Repo: repo, Severities: map[string]int{"error": 1, "warning": 2},
		})
		input.WorstSteps = append(input.WorstSteps, compliance.RepoSteps{
			Repo: repo,
			Steps: []compliance.StepEvidence{
				{Step: "lint", FailRate: 0.5, Failures: 9, TotalRuns: 18},
				{Step: "test", FailRate: 0.1, Failures: 2, TotalRuns: 20},
			},
		})
	}

	return input
}

func itoa4(i int) string {
	digits := []byte{'0' + byte(i/1000%10), '0' + byte(i/100%10), '0' + byte(i/10%10), '0' + byte(i%10)}
	return string(digits)
}

func BenchmarkFoldEvidence(b *testing.B) {
	for _, n := range []int{50, 341, 1000} {
		b.Run(itoa4(n)+"repos", func(b *testing.B) {
			input := benchFoldInput(n)

			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				if out := compliance.FoldEvidence(input); len(out) != n {
					b.Fatalf("folded %d repos, want %d", len(out), n)
				}
			}
		})
	}
}

func BenchmarkDeriveState(b *testing.B) {
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	staleAfter := 336 * time.Hour

	for _, n := range []int{50, 341, 1000} {
		b.Run(itoa4(n)+"repos", func(b *testing.B) {
			evidence := compliance.FoldEvidence(benchFoldInput(n))

			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				for _, repo := range evidence {
					_ = compliance.DeriveState(repo, now, staleAfter)
				}
			}
		})
	}
}

// BenchmarkV1DecodeFoldDerive measures the file-source query path end to
// end: decode the 341-repo document, fold it to evidence, derive every
// state.
func BenchmarkV1DecodeFoldDerive(b *testing.B) {
	data, err := os.ReadFile(benchFixture)
	if err != nil {
		b.Skipf("fixture missing: %v", err)
	}

	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		report, err := compliance.DecodeV1Report(data)
		if err != nil {
			b.Fatalf("decode: %v", err)
		}

		evidence := compliance.V1ReportToEvidence(report, 3)

		for _, repo := range evidence {
			_ = compliance.DeriveState(repo, now, 336*time.Hour)
		}
	}
}

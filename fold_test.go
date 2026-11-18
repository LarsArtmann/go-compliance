package compliance_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-compliance"
)

// Package-scope fixtures: DescribeTable entry arguments and shared Its
// evaluate at tree construction, before any BeforeEach would run.
var (
	foldNow        = time.Date(2026, 11, 17, 20, 0, 0, 0, time.UTC)
	foldQuietRepo  = compliance.NewRepoName("quiet")
	foldActiveRepo = compliance.NewRepoName("active")
)

var _ = Describe("FoldEvidence", func() {
	It("keeps repositories that never dropped evidence for being quiet", func() {
		now := foldNow
		evidence := compliance.FoldEvidence(compliance.FoldInput{
			LatestRuns: []compliance.LatestRun{{
				Repo: foldQuietRepo, LastRunAt: now.Add(-30 * 24 * time.Hour),
				HasVerdict: true, Version: "v1",
			}},
		})

		Expect(evidence).To(HaveLen(1))
		Expect(evidence[0].Repo.Get()).To(Equal("quiet"))
		Expect(evidence[0].HasEvidence()).To(BeTrue())

		By("deriving stale through the report")
		report := compliance.BuildV1Report(evidence, now.Add(-7*24*time.Hour), 14*24*time.Hour, "v", now)
		Expect(report.Repos[0].State).To(Equal("stale"))
	})

	It("unions window health with all-history latest runs", func() {
		now := foldNow
		evidence := compliance.FoldEvidence(compliance.FoldInput{
			WindowHealth: []compliance.WindowHealth{{
				Repo: foldActiveRepo, Runs: 9, StepRows: 40, Failures: 2, FailRate: 0.05,
			}},
			LatestRuns: []compliance.LatestRun{{
				Repo: foldActiveRepo, RepoHash: "abc", LastRunAt: now.Add(-time.Hour),
				AnyFailed: true, HasVerdict: true, Version: "v2",
			}},
		})

		Expect(evidence).To(HaveLen(1))
		record := evidence[0]
		Expect(record.Runs).To(Equal(int64(9)))
		Expect(record.StepRows).To(Equal(int64(40)))
		Expect(record.Failures).To(Equal(int64(2)))
		Expect(record.RepoHash).To(Equal("abc"))
		Expect(record.LastRunVerdict).To(Equal(compliance.VerdictFailed))
		Expect(record.LastRunVersion).To(Equal("v2"))
	})

	It("returns a deterministic, name-sorted order", func() {
		evidence := compliance.FoldEvidence(compliance.FoldInput{
			WindowHealth: []compliance.WindowHealth{
				{Repo: compliance.NewRepoName("zulu"), Runs: 1},
				{Repo: compliance.NewRepoName("alpha"), Runs: 1},
				{Repo: compliance.NewRepoName("mike"), Runs: 1},
			},
		})

		names := []string{evidence[0].Repo.Get(), evidence[1].Repo.Get(), evidence[2].Repo.Get()}
		Expect(names).To(Equal([]string{"alpha", "mike", "zulu"}))
	})

	Describe("worst-step caps", func() {
		steps := func(repo compliance.RepoName, n int) compliance.RepoSteps {
			out := compliance.RepoSteps{Repo: repo}
			for i := range n {
				out.Steps = append(out.Steps, compliance.StepEvidence{
					Step: string(rune('a' + i)), Failures: int64(n - i), TotalRuns: 10,
				})
			}
			return out
		}

		DescribeTable("cap and default",
			func(limit int, expected int) {
				evidence := compliance.FoldEvidence(compliance.FoldInput{
					WindowHealth:   []compliance.WindowHealth{{Repo: foldActiveRepo, Runs: 1}},
					WorstSteps:     []compliance.RepoSteps{steps(foldActiveRepo, 5)},
					WorstStepLimit: limit,
				})
				Expect(evidence[0].WorstSteps).To(HaveLen(expected))
			},
			Entry("cap at 2", 2, 2),
			Entry("cap at 5 keeps all", 5, 5),
			Entry("zero falls back to the default 3", 0, 3),
			Entry("negative falls back to the default 3", -1, 3),
			Entry("no cap set falls back to the default 3", 0, 3),
		)

		It("truncates from the worst end (source order preserved)", func() {
			evidence := compliance.FoldEvidence(compliance.FoldInput{
				WindowHealth: []compliance.WindowHealth{{Repo: foldActiveRepo, Runs: 1}},
				WorstSteps:   []compliance.RepoSteps{steps(foldActiveRepo, 5)},
			})
			Expect(evidence[0].WorstSteps[0].Step).To(Equal("a"))
			Expect(evidence[0].WorstSteps[2].Step).To(Equal("c"))
		})
	})

	Describe("attach-only semantics", func() {
		It("never materializes a repo from versions alone", func() {
			evidence := compliance.FoldEvidence(compliance.FoldInput{
				Versions: []compliance.RepoVersions{{
					Repo: compliance.NewRepoName("ghost"), Versions: []string{"v1"},
				}},
			})
			Expect(evidence).To(BeEmpty())
		})

		It("attaches versions to already-materialized repositories", func() {
			evidence := compliance.FoldEvidence(compliance.FoldInput{
				LatestRuns: []compliance.LatestRun{{Repo: foldQuietRepo, LastRunAt: foldNow, HasVerdict: true}},
				Versions:   []compliance.RepoVersions{{Repo: foldQuietRepo, Versions: []string{"v1", "v2"}}},
			})
			Expect(evidence).To(HaveLen(1))
			Expect(evidence[0].Versions).To(Equal([]string{"v1", "v2"}))
		})

		It("skips empty findings and severity attachments", func() {
			evidence := compliance.FoldEvidence(compliance.FoldInput{
				WindowHealth:      []compliance.WindowHealth{{Repo: foldActiveRepo, Runs: 1}},
				GateFindings:      []compliance.RepoFindings{{Repo: foldActiveRepo, Findings: map[string]int{}}},
				FindingSeverities: []compliance.RepoSeverities{{Repo: foldActiveRepo, Severities: map[string]int{}}},
			})
			Expect(evidence[0].GateFindings).To(BeNil())
			Expect(evidence[0].FindingSeverities).To(BeNil())
		})

		It("attaches non-empty findings and severities", func() {
			evidence := compliance.FoldEvidence(compliance.FoldInput{
				WindowHealth:      []compliance.WindowHealth{{Repo: foldActiveRepo, Runs: 1}},
				GateFindings:      []compliance.RepoFindings{{Repo: foldActiveRepo, Findings: map[string]int{"lint": 3}}},
				FindingSeverities: []compliance.RepoSeverities{{Repo: foldActiveRepo, Severities: map[string]int{"error": 1}}},
			})
			Expect(evidence[0].GateFindings).To(Equal(map[string]int{"lint": 3}))
			Expect(evidence[0].FindingSeverities).To(Equal(map[string]int{"error": 1}))
		})
	})

	It("leaves an empty fold empty", func() {
		Expect(compliance.FoldEvidence(compliance.FoldInput{})).To(BeEmpty())
	})
})

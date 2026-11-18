package compliance_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-compliance"
)

// Package-scope fixtures: DescribeTable entry arguments evaluate at tree
// construction, before any BeforeEach would run.
var (
	v1Now        = time.Date(2026, 11, 17, 20, 0, 0, 0, time.UTC)
	v1Since      = v1Now.Add(-7 * 24 * time.Hour)
	v1StaleAfter = 14 * 24 * time.Hour

	v1Failing = compliance.RepoEvidence{
		Repo: compliance.NewRepoName("failing-repo"), Runs: 10, StepRows: 40, Failures: 3,
		FailRate: 0.3, LastRunAt: v1Now.Add(-1 * time.Hour), LastRunVerdict: compliance.VerdictFailed,
		GateFindings: map[string]int{"samber-linter": 4}, Versions: []string{"v1.2.3"},
		WorstSteps: []compliance.StepEvidence{{Step: "go-test", Failures: 2, TotalRuns: 10, FailRate: 0.2}},
	}
	v1Stale = compliance.RepoEvidence{
		Repo: compliance.NewRepoName("stale-repo"), Runs: 2, LastRunAt: v1Now.Add(-20 * 24 * time.Hour),
		LastRunVerdict: compliance.VerdictPassed,
	}
	v1Fresh = compliance.RepoEvidence{
		Repo: compliance.NewRepoName("fresh-repo"), Runs: 5, LastRunAt: v1Now.Add(-2 * time.Hour),
		LastRunVerdict: compliance.VerdictPassed, GateFindings: map[string]int{}, Versions: []string{},
	}
)

var _ = Describe("BuildV1Report", func() {

	It("derives states and normalizes nil maps and slices", func() {
		report := compliance.BuildV1Report(
			[]compliance.RepoEvidence{v1Fresh, v1Failing, v1Stale}, v1Since, v1StaleAfter, "v0.1.0", v1Now)

		Expect(report.SchemaVersion).To(Equal(1))
		Expect(report.GeneratedAt).To(Equal("2026-11-17T20:00:00Z"))
		Expect(report.Since).To(Equal("2026-11-10T20:00:00Z"))
		Expect(report.StaleAfter).To(Equal("336h0m0s"))
		Expect(report.BuildFlowVersion).To(Equal("v0.1.0"))

		By("sorting worst-first: failing, stale, then compliant")
		Expect(report.Repos).To(HaveLen(3))
		Expect(report.Repos[0].Repo).To(Equal("failing-repo"))
		Expect(report.Repos[0].State).To(Equal("failing"))
		Expect(report.Repos[1].Repo).To(Equal("stale-repo"))
		Expect(report.Repos[1].State).To(Equal("stale"))
		Expect(report.Repos[2].Repo).To(Equal("fresh-repo"))
		Expect(report.Repos[2].State).To(Equal("compliant"))

		By("normalizing nil maps and slices to {} and []")
		emptyRow := compliance.BuildV1Report(
			[]compliance.RepoEvidence{{Repo: compliance.NewRepoName("bare"), Runs: 1}}, v1Since, v1StaleAfter, "", v1Now)
		encoded, err := compliance.MarshalV1(emptyRow)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(encoded)).To(ContainSubstring(`"gate_findings":{}`))
		Expect(string(encoded)).To(ContainSubstring(`"buildflow_versions":[]`))
		Expect(string(encoded)).To(ContainSubstring(`"worst_steps":[]`))
	})

	It("normalizes an empty fleet to an empty repos array", func() {
		report := compliance.BuildV1Report(nil, v1Since, v1StaleAfter, "v", v1Now)
		Expect(report.Repos).To(Equal([]compliance.V1Repo{}))
	})

	It("pins the exact wire shape, key names, and key order", func() {
		report := compliance.BuildV1Report(
			[]compliance.RepoEvidence{{
				Repo: compliance.NewRepoName("shape"), Runs: 1, LastRunAt: v1Now,
				WorstSteps: []compliance.StepEvidence{{Step: "s", Failures: 1, TotalRuns: 2, FailRate: 0.5}},
			}},
			v1Since, v1StaleAfter, "1.2.3", v1Now)

		encoded, err := compliance.MarshalV1(report)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(encoded)).To(MatchJSON(`{
			"schema_version": 1,
			"generated_at": "2026-11-17T20:00:00Z",
			"since": "2026-11-10T20:00:00Z",
			"stale_after": "336h0m0s",
			"buildflow_version": "1.2.3",
			"repos": [{
				"repo": "shape",
				"state": "compliant",
				"runs": 1,
				"step_rows": 0,
				"failures": 0,
				"fail_rate": 0,
				"last_run_at": "2026-11-17T20:00:00Z",
				"last_run_verdict": "",
				"gate_findings": {},
				"buildflow_versions": [],
				"worst_steps": [{
					"step": "s",
					"failures": 1,
					"total_runs": 2,
					"fail_rate": 0.5
				}]
			}]
		}`))
	})

	It("emits keys in the frozen declaration order", func() {
		encoded, err := compliance.MarshalV1(compliance.BuildV1Report(nil, v1Since, v1StaleAfter, "v", v1Now))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(encoded)).To(HavePrefix(
			`{"schema_version":1,"generated_at":`))
	})

	It("sorts ties by failures desc then name", func() {
		report := compliance.BuildV1Report([]compliance.RepoEvidence{
			{Repo: compliance.NewRepoName("b"), Failures: 1, Runs: 1, LastRunAt: v1Now},
			{Repo: compliance.NewRepoName("a"), Failures: 5, Runs: 1, LastRunAt: v1Now},
			{Repo: compliance.NewRepoName("c"), Failures: 5, Runs: 1, LastRunAt: v1Now},
		}, v1Since, v1StaleAfter, "", v1Now)

		names := []string{report.Repos[0].Repo, report.Repos[1].Repo, report.Repos[2].Repo}
		Expect(names).To(Equal([]string{"a", "c", "b"}))
	})
})

var _ = Describe("MarshalV1 and DecodeV1Report", func() {
	v1Now := time.Date(2026, 11, 17, 20, 0, 0, 0, time.UTC)

	build := func() compliance.V1Report {
		return compliance.BuildV1Report([]compliance.RepoEvidence{{
			Repo: compliance.NewRepoName("round-trip"), Runs: 3, Failures: 1,
			LastRunAt: v1Now, LastRunVerdict: compliance.VerdictFailed,
			GateFindings: map[string]int{"z-step": 1, "a-step": 2},
			Versions:     []string{"v9"},
			WorstSteps:   []compliance.StepEvidence{{Step: "x", Failures: 1, TotalRuns: 1, FailRate: 1}},
		}}, v1Now.Add(-168*time.Hour), 336*time.Hour, "v0", v1Now)
	}

	It("round-trips losslessly through the wire format", func() {
		encoded, err := compliance.MarshalV1(build())
		Expect(err).NotTo(HaveOccurred())

		decoded, err := compliance.DecodeV1Report(encoded)
		Expect(err).NotTo(HaveOccurred())
		Expect(decoded.Repos).To(HaveLen(1))
		Expect(decoded.Repos[0].Repo).To(Equal("round-trip"))
		Expect(decoded.Repos[0].State).To(Equal("failing"))
		Expect(decoded.Repos[0].GateFindings).To(Equal(map[string]int{"z-step": 1, "a-step": 2}))
		Expect(decoded.Repos[0].WorstSteps).To(Equal([]compliance.V1Step{
			{Step: "x", Failures: 1, TotalRuns: 1, FailRate: 1},
		}))

		reEncoded, err := compliance.MarshalV1(decoded)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(reEncoded)).To(Equal(string(encoded)), "marshal must be deterministic")
	})

	It("accepts BuildFlow-style second-precision timestamps", func() {
		encoded, err := compliance.MarshalV1(compliance.BuildV1Report(nil, v1Now, time.Hour, "v", v1Now))
		Expect(err).NotTo(HaveOccurred())

		decoded, err := compliance.DecodeV1Report(encoded)
		Expect(err).NotTo(HaveOccurred())
		Expect(compliance.ParseWireTime(decoded.GeneratedAt)).To(Equal(v1Now))
	})

	It("rejects schema_version 2 like the overview consumer", func() {
		_, err := compliance.DecodeV1Report([]byte(`{"schema_version":2,"repos":[]}`))
		Expect(err).To(MatchError(compliance.ErrUnsupportedSchema))
		Expect(err.Error()).To(ContainSubstring("unsupported compliance schema version 2 (want 1)"))
	})

	It("rejects empty and malformed documents", func() {
		_, err := compliance.DecodeV1Report([]byte("   \n"))
		Expect(err).To(MatchError(compliance.ErrParse))

		_, err = compliance.DecodeV1Report([]byte("{not json"))
		Expect(err).To(MatchError(compliance.ErrParse))
	})
})

var _ = Describe("Wire time helpers", func() {
	It("format the zero time as empty and back", func() {
		Expect(compliance.FormatWireTime(time.Time{})).To(Equal(""))
		Expect(compliance.ParseWireTime("")).To(Equal(time.Time{}))
	})

	It("round-trip timestamps through format and parse", func() {
		t := time.Date(2026, 5, 1, 10, 30, 0, 123456789, time.FixedZone("CET", 3600))
		parsed := compliance.ParseWireTime(compliance.FormatWireTime(t))
		Expect(parsed.Equal(t.UTC())).To(BeTrue())
	})

	It("return the zero time for malformed timestamps", func() {
		Expect(compliance.ParseWireTime("yesterday").IsZero()).To(BeTrue())
	})
})

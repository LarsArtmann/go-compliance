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
	stateNow        = time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	stateFreshRun   = stateNow.Add(-24 * time.Hour)
	stateAncientRun = stateNow.Add(-30 * 24 * time.Hour)
)

var _ = Describe("DeriveState", func() {
	DescribeTable(
		"the v1 state model",
		func(evidence compliance.RepoEvidence, staleAfter time.Duration, expected compliance.State) {
			Expect(compliance.DeriveState(evidence, stateNow, staleAfter)).To(Equal(expected))
		},
		Entry(
			"no evidence at all is never-covered",
			compliance.RepoEvidence{},
			336*time.Hour,
			compliance.StateNeverCovered,
		),
		Entry(
			"never-covered wins over staleness",
			compliance.RepoEvidence{},
			time.Nanosecond,
			compliance.StateNeverCovered,
		),
		Entry(
			"any failure in the window is failing",
			compliance.RepoEvidence{
				Runs:      5,
				Failures:  1,
				LastRunAt: stateFreshRun,
			},
			336*time.Hour,
			compliance.StateFailing,
		),
		Entry(
			"failure wins over staleness",
			compliance.RepoEvidence{
				Runs:      5,
				Failures:  1,
				LastRunAt: stateAncientRun,
			},
			336*time.Hour,
			compliance.StateFailing,
		),
		Entry("a quiet-but-once-failed step row still fails",
			compliance.RepoEvidence{StepRows: 3, Failures: 2}, 336*time.Hour, compliance.StateFailing),
		Entry("fresh evidence without failures is compliant",
			compliance.RepoEvidence{Runs: 5, LastRunAt: stateFreshRun}, 336*time.Hour, compliance.StateCompliant),
		Entry("a last run older than staleAfter is stale",
			compliance.RepoEvidence{Runs: 5, LastRunAt: stateAncientRun}, 14*24*time.Hour, compliance.StateStale),
		Entry(
			"a last run exactly staleAfter old stays compliant (strictly older required)",
			compliance.RepoEvidence{
				Runs:      5,
				LastRunAt: stateNow.Add(-336 * time.Hour),
			},
			336*time.Hour,
			compliance.StateCompliant,
		),
		Entry("staleness disabled by staleAfter <= 0 keeps ancient evidence compliant",
			compliance.RepoEvidence{Runs: 5, LastRunAt: stateAncientRun}, time.Duration(0), compliance.StateCompliant),
		Entry("negative staleness disabled keeps ancient evidence compliant",
			compliance.RepoEvidence{Runs: 5, LastRunAt: stateAncientRun}, -time.Hour, compliance.StateCompliant),
		Entry("window rows without a last-run recording are compliant",
			compliance.RepoEvidence{Runs: 2, StepRows: 4}, 336*time.Hour, compliance.StateCompliant),
		Entry(
			"clock skew (last run in the future) stays compliant",
			compliance.RepoEvidence{
				Runs:      1,
				LastRunAt: stateNow.Add(time.Hour),
			},
			336*time.Hour,
			compliance.StateCompliant,
		),
	)

	Describe("precedence", Ordered, func() {
		It("never-covered only for rows without any evidence", func() {
			Expect(compliance.RepoEvidence{Runs: 1}.HasEvidence()).To(BeTrue())
			Expect(compliance.RepoEvidence{StepRows: 1}.HasEvidence()).To(BeTrue())
			Expect(compliance.RepoEvidence{LastRunAt: stateFreshRun}.HasEvidence()).To(BeTrue())
			Expect(compliance.RepoEvidence{Failures: 1}.HasEvidence()).To(BeFalse(),
				"failures alone cannot exist without rows; such a row is still treated as never-covered")
		})
	})
})

var _ = Describe("State", func() {
	DescribeTable("validation",
		func(raw string, valid bool) {
			_, err := compliance.ParseState(raw)
			if valid {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(MatchError(compliance.ErrValidation))
			}
		},
		Entry("compliant", "compliant", true),
		Entry("failing", "failing", true),
		Entry("stale", "stale", true),
		Entry("never-covered", "never-covered", true),
		Entry("empty", "", false),
		Entry("unknown", "healthy", false),
		Entry("case-mismatch", "Compliant", false),
	)

	Describe("Rank", func() {
		It("orders the triage view: failing, stale, never-covered, compliant", func() {
			Expect(compliance.StateFailing.Rank()).To(BeNumerically("<", compliance.StateStale.Rank()))
			Expect(compliance.StateStale.Rank()).To(BeNumerically("<", compliance.StateNeverCovered.Rank()))
			Expect(compliance.StateNeverCovered.Rank()).To(BeNumerically("<", compliance.StateCompliant.Rank()))
			Expect(compliance.State("bogus").Rank()).To(BeNumerically(">", compliance.StateCompliant.Rank()))
		})
	})
})

var _ = Describe("Verdict", func() {
	It("uses the wire vocabulary with unknown empty", func() {
		Expect(compliance.VerdictPassed.String()).To(Equal("passed"))
		Expect(compliance.VerdictFailed.String()).To(Equal("failed"))
		Expect(compliance.VerdictUnknown.String()).To(Equal(""))
	})
})

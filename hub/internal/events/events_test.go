package events_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-compliance"
	"github.com/larsartmann/go-compliance/hub/internal/events"
	"github.com/larsartmann/go-compliance/hub/internal/testsupport"
	"github.com/larsartmann/go-sse"
	ssetest "github.com/larsartmann/go-sse/ssetest"
)

var _ = Describe("rows projection", func() {
	DescribeTable("deriving row state",
		func(evidence compliance.RepoEvidence, expected compliance.State) {
			rows := events.RowsFromEvidence([]compliance.RepoEvidence{evidence},
				time.Unix(1_000_000, 0), 24*time.Hour)
			Expect(rows).To(HaveLen(1))
			Expect(rows[0].State).To(Equal(expected))
		},
		Entry("compliant recent passing", compliance.RepoEvidence{
			Repo: compliance.NewRepoName("a"), Runs: 3,
			LastRunAt: time.Unix(1_000_000-60, 0), LastRunVerdict: compliance.VerdictPassed,
		}, compliance.StateCompliant),
		Entry("failing verdict", compliance.RepoEvidence{
			Repo: compliance.NewRepoName("a"), Runs: 3, Failures: 2,
			LastRunAt: time.Unix(1_000_000-60, 0), LastRunVerdict: compliance.VerdictFailed,
		}, compliance.StateFailing),
		Entry("stale last run", compliance.RepoEvidence{
			Repo: compliance.NewRepoName("a"), Runs: 3,
			LastRunAt: time.Unix(0, 0), LastRunVerdict: compliance.VerdictPassed,
		}, compliance.StateStale),
		Entry("never covered zero evidence", compliance.RepoEvidence{
			Repo: compliance.NewRepoName("a"),
		}, compliance.StateNeverCovered),
	)

	It("projects findings, severities, and worst steps", func() {
		evidence := compliance.RepoEvidence{
			Repo:              compliance.NewRepoName("fleet"),
			Runs:              7,
			Failures:          1,
			GateFindings:      map[string]int{"build": 2, "lint": 3},
			FindingSeverities: map[string]int{"error": 2, "warning": 3},
			WorstSteps:        []compliance.StepEvidence{{Step: "lint"}, {Step: "build"}},
			LastRunVerdict:    compliance.VerdictFailed,
		}

		rows := events.RowsFromEvidence([]compliance.RepoEvidence{evidence}, time.Now(), 0)
		Expect(rows[0].Findings).To(Equal(5))
		Expect(rows[0].Severities).To(Equal(map[string]int{"error": 2, "warning": 3}))
		Expect(rows[0].WorstSteps).To(Equal([]string{"lint", "build"}))
		Expect(rows[0].LastVerdict).To(Equal(compliance.VerdictFailed))
	})
})

var _ = Describe("state diffing", func() {
	baseRow := func(repo string, state compliance.State) events.RepoRow {
		return events.RepoRow{Repo: repo, State: state, Severities: map[string]int{}}
	}

	It("splits added and changed repos, omitting unchanged ones", func() {
		previous := []events.RepoRow{
			baseRow("stable", compliance.StateCompliant),
			baseRow("flipped", compliance.StateCompliant),
		}
		next := []events.RepoRow{
			baseRow("stable", compliance.StateCompliant),
			baseRow("flipped", compliance.StateFailing),
			baseRow("fresh", compliance.StateNeverCovered),
		}

		added, changed, removed := events.DiffRows(previous, next)
		Expect(added).To(HaveLen(1))
		Expect(added[0].Repo).To(Equal("fresh"))
		Expect(changed).To(HaveLen(1))
		Expect(changed[0].Repo).To(Equal("flipped"))
		Expect(changed[0].State).To(Equal(compliance.StateFailing))
		Expect(removed).To(BeEmpty())
	})

	It("reports severity and verdict moves as changes", func() {
		row := baseRow("r", compliance.StateFailing)
		row.Severities = map[string]int{"error": 1}

		moved := baseRow("r", compliance.StateFailing)
		moved.Severities = map[string]int{"error": 2}

		_, changed, _ := events.DiffRows([]events.RepoRow{row}, []events.RepoRow{moved})
		Expect(changed).To(HaveLen(1))
	})

	It("reports disappeared repos as removed", func() {
		added, changed, removed := events.DiffRows(
			[]events.RepoRow{baseRow("gone", compliance.StateStale), baseRow("kept", compliance.StateStale)},
			[]events.RepoRow{baseRow("kept", compliance.StateStale)},
		)
		Expect(added).To(BeEmpty())
		Expect(changed).To(BeEmpty())
		Expect(removed).To(Equal([]string{"gone"}))
	})
})

var _ = Describe("events hub endpoint", func() {
	rowsFor := func(repo string, state compliance.State) []events.RepoRow {
		return []events.RepoRow{{Repo: repo, State: state, Severities: map[string]int{}}}
	}

	It("sends the current snapshot first, then live updates", func() {
		hub := events.NewHub()
		Expect(hub.PublishSnapshot(rowsFor("fleet", compliance.StateCompliant))).To(Equal("1"))

		handler := hub.Handler()

		published := make(chan error, 1)

		go func() {
			testsupport.WaitForSubscriber(hub)

			_, err := hub.PublishRefresh(rowsFor("fleet", compliance.StateFailing))
			published <- err
		}()

		received := ssetest.CollectN(GinkgoTB(), handler, 2, ssetest.WithPath("/events"))

		Expect(<-published).NotTo(HaveOccurred())

		ssetest.RequireEventCount(GinkgoTB(), received, 2)
		ssetest.RequireEventType(GinkgoTB(), received[0], events.EventSnapshot)
		ssetest.RequireDataContains(GinkgoTB(), received[0], `"repo":"fleet"`)
		ssetest.RequireEventType(GinkgoTB(), received[1], events.EventUpdate)
		ssetest.RequireDataContains(GinkgoTB(), received[1], `"repo":"fleet"`)
		ssetest.RequireDataContains(GinkgoTB(), received[1], `"state":"failing"`)
	})

	It("replays missed events after a Last-Event-ID before live state", func() {
		hub := events.NewHub()
		first, err := hub.PublishSnapshot(rowsFor("a", compliance.StateCompliant))
		Expect(err).NotTo(HaveOccurred())
		_, err = hub.PublishRefresh(rowsFor("a", compliance.StateFailing))
		Expect(err).NotTo(HaveOccurred())

		received := ssetest.CollectN(GinkgoTB(), hub.Handler(), 1,
			ssetest.WithPath("/events"),
			ssetest.WithLastEventID(sse.NewEventID(first)),
		)

		Expect(received[0].ID).To(Equal("2"))
		ssetest.RequireEventType(GinkgoTB(), received[0], events.EventUpdate)
		ssetest.RequireDataContains(GinkgoTB(), received[0], `"state":"failing"`)
	})

	It("publishes diffs only when refreshes change displayed state", func() {
		hub := events.NewHub()

		published, err := hub.PublishRefresh(rowsFor("a", compliance.StateCompliant))
		Expect(err).NotTo(HaveOccurred())
		Expect(published).To(BeTrue())

		published, err = hub.PublishRefresh(rowsFor("a", compliance.StateCompliant))
		Expect(err).NotTo(HaveOccurred())
		Expect(published).To(BeFalse())

		published, err = hub.PublishRefresh(rowsFor("a", compliance.StateStale))
		Expect(err).NotTo(HaveOccurred())
		Expect(published).To(BeTrue())
	})

	It("keeps the latest rows queryable and detached from caller mutation", func() {
		hub := events.NewHub()
		rows := rowsFor("a", compliance.StateCompliant)
		rows[0].Severities["error"] = 1

		_, err := hub.PublishSnapshot(rows)
		Expect(err).NotTo(HaveOccurred())

		rows[0].Severities["error"] = 99

		current := hub.Rows()
		Expect(current).To(HaveLen(1))
		Expect(current[0].Severities).To(Equal(map[string]int{"error": 1}))
	})

	It("returns nil rows only until the first publish; an empty fleet is published", func() {
		hub := events.NewHub()
		Expect(hub.Rows()).To(BeNil())

		_, err := hub.PublishSnapshot(nil)
		Expect(err).NotTo(HaveOccurred())

		Expect(hub.Rows()).NotTo(BeNil())
		Expect(hub.Rows()).To(BeEmpty())
	})

	It("carries state counts and totals on snapshots and updates", func() {
		hub := events.NewHub()
		_, err := hub.PublishSnapshot([]events.RepoRow{
			{Repo: "a", State: compliance.StateFailing, Severities: map[string]int{}},
			{Repo: "b", State: compliance.StateFailing, Severities: map[string]int{}},
			{Repo: "c", State: compliance.StateCompliant, Severities: map[string]int{}},
		})
		Expect(err).NotTo(HaveOccurred())

		received := ssetest.CollectN(GinkgoTB(), hub.Handler(), 1, ssetest.WithPath("/events"))
		ssetest.RequireDataContains(GinkgoTB(), received[0], `"total":3`)
		ssetest.RequireDataContains(GinkgoTB(), received[0], `"failing":2`)
		ssetest.RequireDataContains(GinkgoTB(), received[0], `"compliant":1`)
	})

	It("numbers published events monotonically", func() {
		hub := events.NewHub()

		first, err := hub.PublishSnapshot(nil)
		Expect(err).NotTo(HaveOccurred())
		second, err := hub.PublishUpdate(nil, nil, []string{"gone"})
		Expect(err).NotTo(HaveOccurred())

		Expect(first).To(Equal("1"))
		Expect(second).To(Equal("2"))
	})
})

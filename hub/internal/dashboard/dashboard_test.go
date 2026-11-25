package dashboard_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-compliance"
	"github.com/larsartmann/go-compliance/hub/internal/dashboard"
	"github.com/larsartmann/go-compliance/hub/internal/events"
	"github.com/larsartmann/go-compliance/hub/internal/testsupport"
	"github.com/larsartmann/go-datastar"
	datastartest "github.com/larsartmann/go-datastar/datastartest"
	dstarstatic "github.com/larsartmann/go-datastar/static"
)

func TestDashboard(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dashboard Patch Suite")
}

var _ = Describe("datastar patch stream", func() {
	rowFor := func(repo string, state compliance.State) events.RepoRow {
		return events.RepoRow{Repo: repo, State: state, Severities: map[string]int{}}
	}

	It("sends signals with counts, then rows into the fleet region", func() {
		hub := events.NewHub()
		_, err := hub.PublishSnapshot([]events.RepoRow{
			rowFor("alpha", compliance.StateCompliant),
			rowFor("beta", compliance.StateFailing),
		})
		Expect(err).NotTo(HaveOccurred())

		received := datastartest.CollectN(GinkgoTB(), dashboard.New(hub).StreamHandler(), 3,
			datastartest.WithPath("/events"))

		datastartest.RequireEventCount(GinkgoTB(), received, 3)
		Expect(received[0].IsSignals()).To(BeTrue())
		datastartest.RequireSignalsContain(GinkgoTB(), received[0], "counts")
		datastartest.RequireSignalsContain(GinkgoTB(), received[0], "total")

		var decoded struct {
			Total  int            `json:"total"`
			Counts map[string]int `json:"counts"`
		}
		Expect(received[0].UnmarshalSignals(&decoded)).To(Succeed())
		Expect(decoded.Total).To(Equal(2))
		Expect(decoded.Counts).To(Equal(map[string]int{"compliant": 1, "failing": 1}))

		Expect(received[1].IsElements()).To(BeTrue())
		datastartest.RequireElementsContains(GinkgoTB(), received[1],
			"#"+dashboard.FleetSummaryID, string(datastar.ElementPatchModeInner), `1</span><span>of 2 repos is failing</span>`)
		datastartest.RequireElementsContains(GinkgoTB(), received[1],
			"#"+dashboard.FleetSummaryID, string(datastar.ElementPatchModeInner), `Never covered</span><span class="ml-1.5 tabular-nums opacity-70">0</span>`)
		Expect(received[2].IsElements()).To(BeTrue())
		datastartest.RequireElementsContains(GinkgoTB(), received[2],
			dashboard.DefaultRowsSelector, string(datastar.ElementPatchModeInner), `<tr data-repo="alpha" data-show=`)
		datastartest.RequireElementsContains(GinkgoTB(), received[2],
			dashboard.DefaultRowsSelector, string(datastar.ElementPatchModeInner), `<tr data-repo="beta" data-show=`)
		datastartest.RequireElementsContains(GinkgoTB(), received[2],
			dashboard.DefaultRowsSelector, string(datastar.ElementPatchModeInner), `compliant`)
	})

	It("projects refreshes as appends, replaces, and removes", func() {
		hub := events.NewHub()
		_, err := hub.PublishSnapshot([]events.RepoRow{
			rowFor("stable", compliance.StateStale),
			rowFor("flipped", compliance.StateCompliant),
			rowFor("doomed", compliance.StateStale),
		})
		Expect(err).NotTo(HaveOccurred())

		board := dashboard.New(hub)
		stream := board.StreamHandler()

		published := make(chan error, 1)

		go func() {
			testsupport.WaitForSubscriber(hub)

			_, err := hub.PublishRefresh([]events.RepoRow{
				rowFor("stable", compliance.StateStale),
				rowFor("flipped", compliance.StateFailing),
				rowFor("newcomer", compliance.StateNeverCovered),
			})
			published <- err
		}()

		// 3 initial (signals + summary + rows) + 5 update events
		// (signals, summary, append newcomer, replace flipped, remove
		// doomed).
		received := datastartest.CollectN(GinkgoTB(), stream, 8, datastartest.WithPath("/events"))

		Expect(<-published).NotTo(HaveOccurred())

		signals, found := datastartest.FindSignals(received[3:])
		Expect(found).To(BeTrue())

		var decoded struct {
			Total  int            `json:"total"`
			Counts map[string]int `json:"counts"`
		}
		Expect(signals.UnmarshalSignals(&decoded)).To(Succeed())
		Expect(decoded.Total).To(Equal(3))
		Expect(decoded.Counts).To(Equal(map[string]int{"failing": 1, "never-covered": 1, "stale": 1}))

		replaced := datastartest.FindAllElements(received, `tr[data-repo="flipped"]`)
		Expect(replaced).To(HaveLen(1))
		Expect(replaced[0].Mode()).To(Equal(string(datastar.ElementPatchModeReplace)))
		datastartest.RequireElementsContains(GinkgoTB(), replaced[0],
			`tr[data-repo="flipped"]`, string(datastar.ElementPatchModeReplace), `failing</span>`)

		removals := datastartest.FindAllElements(received, `tr[data-repo="doomed"]`)
		Expect(removals).To(HaveLen(1))
		Expect(removals[0].Mode()).To(Equal(string(datastar.ElementPatchModeRemove)))

		appends := datastartest.FindAllElements(received, dashboard.DefaultRowsSelector)
		Expect(appends).To(HaveLen(2)) // initial inner patch + update append patch
		Expect(appends[1].Mode()).To(Equal(string(datastar.ElementPatchModeAppend)))
		datastartest.RequireElementsContains(GinkgoTB(), appends[1],
			dashboard.DefaultRowsSelector, string(datastar.ElementPatchModeAppend), `never-covered</span>`)
	})

	It("serves the pinned DataStar bundle", func() {
		board := dashboard.New(events.NewHub())

		recorder := httptest.NewRecorder()
		board.ScriptHandler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/datastar.js", nil))

		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(recorder.Header().Get("Content-Type")).To(ContainSubstring("javascript"))
		Expect(recorder.Body.Bytes()).To(Equal(dstarstatic.Bytes()))
	})
})

package cli

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	datastar "github.com/larsartmann/go-datastar"
	datastartest "github.com/larsartmann/go-datastar/datastartest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-compliance"
	"github.com/larsartmann/go-compliance/hub/internal/cache"
	"github.com/larsartmann/go-compliance/hub/internal/dashboard"
	"github.com/larsartmann/go-compliance/hub/internal/discovery"
	"github.com/larsartmann/go-compliance/hub/internal/events"
)

// The M27 end-to-end contract: an evidence write reaches the browser as a
// DataStar patch — store hook -> refresher kick -> diff -> SSE patch.
var _ = Describe("live re-derive", func() {
	It("publishes a browser patch event when evidence is written", func() {
		ctx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)

		store, err := cache.Open(ctx, filepath.Join(GinkgoTB().TempDir(), "evidence.db"))
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { _ = store.Close() })

		Expect(store.WriteEvidence(ctx, compliance.NewRepoName("boot-repo"),
			compliance.RunResult{Verdict: compliance.VerdictPassed})).To(Succeed())

		hub := events.NewHub()
		refresher := newRefresher(hub, store, 24*time.Hour, 336*time.Hour, 3, "")
		Expect(refresher.Refresh(ctx)).To(Succeed())
		store.OnWrite(func(compliance.RepoName) { refresher.Kick() })

		go refresher.Run(ctx)

		board := dashboard.New(hub)

		go func() {
			// After the stream below is subscribed: write evidence and
			// expect the diff to arrive as a patch event.
			time.Sleep(200 * time.Millisecond)

			_ = store.WriteEvidence(ctx, compliance.NewRepoName("late-repo"),
				compliance.RunResult{Verdict: compliance.VerdictFailed, Failures: 2, StepRows: 10})
		}()

		received := datastartest.CollectWithTimeout(GinkgoTB(), board.StreamHandler(), 5*time.Second,
			datastartest.WithPath("/events"))

		// Initial snapshot: signals + fleet summary + the full rows patch.
		Expect(len(received)).To(BeNumerically(">=", 3),
			"expected the initial snapshot patches; events: %s", describeEvents(received))

		bootRendered, lateAppended := false, false

		for _, evt := range received {
			if !evt.IsElements() || evt.Selector() != dashboard.DefaultRowsSelector {
				continue
			}

			switch evt.Mode() {
			case string(datastar.ElementPatchModeInner):
				datastartest.RequireElementsContains(GinkgoTB(), evt,
					dashboard.DefaultRowsSelector, string(datastar.ElementPatchModeInner), "boot-repo")

				bootRendered = true
			case string(datastar.ElementPatchModeAppend):
				datastartest.RequireElementsContains(GinkgoTB(), evt,
					dashboard.DefaultRowsSelector, string(datastar.ElementPatchModeAppend), "late-repo")

				lateAppended = true
			}
		}

		Expect(bootRendered).To(BeTrue(),
			"no inner rows patch carried the initial boot-repo row; events: %s", describeEvents(received))
		Expect(lateAppended).To(BeTrue(),
			"no appended rows patch carried the late-repo row; events: %s", describeEvents(received))
	})
})

var _ = Describe("refresh freshness guard", func() {
	It("bounds a hung query: the refresh fails fast, nothing is stamped, and the next refresh works", func() {
		hub := events.NewHub()

		hanging := compliance.SourceFunc(
			func(ctx context.Context, _ time.Time, _ int, _ string) ([]compliance.RepoEvidence, error) {
				<-ctx.Done()

				return nil, ctx.Err()
			},
		)
		good := compliance.SourceFunc(
			func(_ context.Context, _ time.Time, _ int, _ string) ([]compliance.RepoEvidence, error) {
				return []compliance.RepoEvidence{}, nil
			},
		)

		refresher := newRefresher(hub, hanging, time.Hour, 24*time.Hour, 3, "")
		refresher.queryTimeout = 50 * time.Millisecond

		Expect(refresher.Refresh(context.Background())).To(MatchError(context.DeadlineExceeded))
		Expect(refresher.LastRefreshSuccess()).To(BeZero())

		refresher.source = good
		Expect(refresher.Refresh(context.Background())).To(Succeed())
		Expect(refresher.LastRefreshSuccess()).NotTo(BeZero())
	})

	It("stamps LastRefreshSuccess only after query AND publish succeed", func() {
		hub := events.NewHub()

		failing := compliance.SourceFunc(
			func(_ context.Context, _ time.Time, _ int, _ string) ([]compliance.RepoEvidence, error) {
				return nil, context.Canceled
			},
		)

		refresher := newRefresher(hub, failing, time.Hour, 24*time.Hour, 3, "")

		Expect(refresher.Refresh(context.Background())).NotTo(Succeed())
		Expect(refresher.LastRefreshSuccess()).To(BeZero())
	})
})

var _ = Describe("discovery contact stamping", func() {
	It("keeps the last contact stamp on a failed snapshot and never regresses it", func() {
		refresher := newRefresher(events.NewHub(), nil, time.Hour, 24*time.Hour, 3, "")
		dead := discovery.New(filepath.Join(GinkgoTB().TempDir(), "missing.sock"))

		Expect(refresher.LastDiscoverySuccess()).To(BeZero())
		Expect(refresher.refreshDiscovered(context.Background(), dead, nil)).NotTo(Succeed())
		Expect(refresher.LastDiscoverySuccess()).To(BeZero())

		refresher.markDiscoverySuccess()
		first := refresher.LastDiscoverySuccess()

		Expect(refresher.refreshDiscovered(context.Background(), dead, nil)).NotTo(Succeed())
		Expect(refresher.LastDiscoverySuccess()).To(Equal(first))
	})
})

func describeEvents(events []datastartest.Event) string {
	parts := make([]string, 0, len(events))

	for _, evt := range events {
		parts = append(parts, evt.Type+": "+strings.Join(evt.DataLines, " "))
	}

	return strings.Join(parts, "\n")
}

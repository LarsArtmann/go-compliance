package cli

import (
	"context"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-compliance"
	"github.com/larsartmann/go-compliance/hub/internal/cache"
	"github.com/larsartmann/go-compliance/hub/internal/events"
)

// The realtime promise, closed: a reporter in ANOTHER process appends to the
// same SQLite cache, and the hub re-derives within ~2s — the directory watch
// on the WAL sidecars replaces the 5-minute safety-net wait.
var _ = Describe("evidence watch", func() {
	It("refreshes when another process writes the cache (WAL sidecar event)", func() {
		ctx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)

		dbPath := filepath.Join(GinkgoTB().TempDir(), "evidence.db")

		hubStore, err := cache.Open(ctx, dbPath)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { _ = hubStore.Close() })

		Expect(hubStore.WriteEvidence(ctx, compliance.NewRepoName("boot-repo"),
			compliance.RunResult{Verdict: compliance.VerdictPassed})).To(Succeed())

		hub := events.NewHub()
		refresher := newRefresher(hub, hubStore, 24*time.Hour, 336*time.Hour, 3, "")
		Expect(refresher.Refresh(ctx)).To(Succeed())
		Expect(hub.Rows()).To(HaveLen(1))

		go refresher.Run(ctx)

		// Deliberately NO OnWrite wiring on the hub store: the only trigger
		// path for the write below is the directory watch — exactly what a
		// cross-process reporter produces.
		watcher := watchEvidence(ctx, QueryFlags{Source: "cache", DB: dbPath}, refresher)
		Expect(watcher).NotTo(BeNil())

		// The "other process": an independent connection whose writes are
		// invisible to the hub store's in-process OnWrite hook.
		other, err := cache.Open(ctx, dbPath)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { _ = other.Close() })

		// Let the watch arm (fsnotify add + initial walk settle).
		time.Sleep(300 * time.Millisecond)

		Expect(other.WriteEvidence(ctx, compliance.NewRepoName("cross-process-repo"),
			compliance.RunResult{Verdict: compliance.VerdictFailed, Failures: 1})).To(Succeed())

		Eventually(func() int {
			return len(hub.Rows())
		}, "2s", "20ms").Should(Equal(2),
			"a cross-process WAL write must refresh within the 2s realtime budget")
	})

	DescribeTable("watch targets per source",
		func(flags QueryFlags, wantDirs []string, wantNames map[string]struct{}) {
			dirs, names := evidenceWatchTargets(flags)
			Expect(dirs).To(Equal(wantDirs))
			Expect(names).To(Equal(wantNames))
		},
		Entry("cache watches the directory for db and WAL siblings",
			QueryFlags{Source: "cache", DB: "/tmp/x/evidence.db"},
			[]string{"/tmp/x"},
			map[string]struct{}{
				"evidence.db": {}, "evidence.db-wal": {}, "evidence.db-shm": {},
			}),
		Entry("file watches the evidence file's directory (atomic renames land there)",
			QueryFlags{Source: "file", File: "/var/lib/overview/compliance.json"},
			[]string{"/var/lib/overview"},
			map[string]struct{}{"compliance.json": {}}),
		Entry("buildflow has no filesystem truth to watch",
			QueryFlags{Source: "buildflow"}, []string(nil), map[string]struct{}{}),
	)
})

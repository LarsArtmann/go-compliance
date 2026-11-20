package cache_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-compliance"
	"github.com/larsartmann/go-compliance/hub/internal/cache"
)

var _ = Describe("SQL seam error paths", func() {
	var store *cache.Store

	BeforeEach(func() {
		dbDir := mustTempDir()

		DeferCleanup(func() { _ = os.RemoveAll(dbDir) })

		var err error

		store, err = cache.Open(context.Background(), filepath.Join(dbDir, "cache.db"))
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { _ = store.Close() })
	})

	It("family-types a dead-store query as IO with the query context in the message", func() {
		ctx := context.Background()

		Expect(store.Close()).To(Succeed())

		evidence, err := store.QueryEvidence(ctx, time.Now().Add(-time.Hour), 0, "")
		Expect(err).To(HaveOccurred())
		Expect(errors.Is(err, compliance.ErrIO)).To(BeTrue(), "the queryRows wrap home must keep the IO sentinel")
		Expect(errorfamily.Classify(err)).To(Equal(errorfamily.Transient), "IO maps to the retryable family (exit 75)")
		Expect(err.Error()).To(ContainSubstring("query window health"), "the what context must survive the wrap")
		Expect(err.Error()).To(ContainSubstring("closed"), "the driver cause must survive the wrap")
		Expect(evidence).To(BeNil())
	})

	It("family-types a dead-store write as IO with the repo in the message", func() {
		ctx := context.Background()

		Expect(store.Close()).To(Succeed())

		err := store.WriteEvidence(ctx, compliance.NewRepoName("late-repo"), compliance.RunResult{
			Verdict: compliance.VerdictPassed,
		})
		Expect(err).To(HaveOccurred())
		Expect(errors.Is(err, compliance.ErrIO)).To(BeTrue())
		Expect(errorfamily.Classify(err)).To(Equal(errorfamily.Transient))
		Expect(err.Error()).To(ContainSubstring("late-repo"))
	})

	It("keeps the empty-fleet contract: a fresh cache queries to zero evidence, no error", func() {
		evidence, err := store.QueryEvidence(context.Background(), time.Now().Add(-time.Hour), 0, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(evidence).To(BeEmpty(), "an empty store is an empty fleet, not a failure")
	})
})

var _ = Describe("evidence cache", func() {
	var store *cache.Store

	BeforeEach(func() {
		dbDir := mustTempDir()

		DeferCleanup(func() { _ = os.RemoveAll(dbDir) })

		var err error

		store, err = cache.Open(context.Background(), filepath.Join(dbDir, "cache.db"))
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { _ = store.Close() })
	})

	It("round-trips recorded evidence through the window query", func() {
		ctx := context.Background()
		now := time.Now()

		Expect(store.WriteEvidence(ctx, compliance.NewRepoName("repo-a"), compliance.RunResult{
			At: now.Add(-2 * time.Hour), Verdict: compliance.VerdictFailed,
			StepRows: 10, Failures: 2, ProducerVersion: "v1",
			GateFindings:      map[string]int{"lint": 4},
			FindingSeverities: map[string]int{"error": 3, "warning": 1},
		})).To(Succeed())

		Expect(store.WriteEvidence(ctx, compliance.NewRepoName("repo-b"), compliance.RunResult{
			At: now.Add(-30 * 24 * time.Hour), Verdict: compliance.VerdictPassed, ProducerVersion: "v0.9",
		})).To(Succeed())

		evidence, err := store.QueryEvidence(ctx, now.Add(-7*24*time.Hour), 0, "")
		Expect(err).NotTo(HaveOccurred())

		byRepo := map[string]compliance.RepoEvidence{}
		for _, record := range evidence {
			byRepo[record.Repo.Get()] = record
		}

		Expect(byRepo).To(HaveLen(2), "quiet repos stay visible via their all-history latest run")

		a := byRepo["repo-a"]
		Expect(a.Runs).To(Equal(int64(1)))
		Expect(a.StepRows).To(Equal(int64(10)))
		Expect(a.Failures).To(Equal(int64(2)))
		Expect(a.FailRate).To(Equal(0.2))
		Expect(a.LastRunVerdict).To(Equal(compliance.VerdictFailed))
		Expect(a.GateFindings).To(Equal(map[string]int{"lint": 4}))
		Expect(a.FindingSeverities).To(Equal(map[string]int{"error": 3, "warning": 1}))
		Expect(a.Versions).To(Equal([]string{"v1"}))
		Expect(compliance.DeriveState(a, now, compliance.DefaultStaleAfter)).To(Equal(compliance.StateFailing))

		b := byRepo["repo-b"]
		Expect(compliance.DeriveState(b, now, compliance.DefaultStaleAfter)).To(Equal(compliance.StateStale))
	})

	It("keeps only the newest record's verdict, findings, and versions", func() {
		ctx := context.Background()
		now := time.Now()

		Expect(store.WriteEvidence(ctx, compliance.NewRepoName("repo"), compliance.RunResult{
			At: now.Add(-2 * time.Hour), Verdict: compliance.VerdictFailed, Failures: 1,
			GateFindings: map[string]int{"old": 9}, ProducerVersion: "v1",
		})).To(Succeed())

		Expect(store.WriteEvidence(ctx, compliance.NewRepoName("repo"), compliance.RunResult{
			At: now.Add(-1 * time.Hour), Verdict: compliance.VerdictPassed,
			GateFindings: map[string]int{"lint": 2}, ProducerVersion: "v2",
		})).To(Succeed())

		evidence, err := store.QueryEvidence(ctx, now.Add(-7*24*time.Hour), 0, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(evidence).To(HaveLen(1))

		record := evidence[0]
		Expect(record.LastRunVerdict).To(Equal(compliance.VerdictPassed))
		Expect(record.GateFindings).To(Equal(map[string]int{"lint": 2}))
		Expect(record.Versions).To(ConsistOf("v1", "v2"), "versions are window-wide, not latest-run")
		Expect(record.Failures).To(Equal(int64(1)), "window failures accumulate honestly")
	})

	It("satisfies the Source interface", func() {
		var source compliance.Source = store
		Expect(source.Name().Get()).To(Equal("cache"))
	})

	It("is a valid EvidenceWriter for the reporter SDK", func() {
		ctx := context.Background()
		Expect(compliance.RecordTo(ctx, store, compliance.NewRepoName("sdk-repo"),
			compliance.RunResult{Verdict: compliance.VerdictPassed, ProducerVersion: "sdk"})).To(Succeed())

		evidence, err := store.QueryEvidence(ctx, time.Now().Add(-time.Hour), 0, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(evidence).To(HaveLen(1))
		Expect(evidence[0].Repo.Get()).To(Equal("sdk-repo"))
	})

	It("survives reopen with migrations applied exactly once", func() {
		ctx := context.Background()
		reopenDir := mustTempDir()

		DeferCleanup(func() { _ = os.RemoveAll(reopenDir) })

		path := filepath.Join(reopenDir, "reopen.db")

		first, err := cache.Open(ctx, path)
		Expect(err).NotTo(HaveOccurred())
		Expect(first.WriteEvidence(ctx, compliance.NewRepoName("persist"), compliance.RunResult{
			Verdict: compliance.VerdictPassed,
		})).To(Succeed())
		Expect(first.Close()).To(Succeed())

		second, err := cache.Open(ctx, path)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { _ = second.Close() })

		evidence, err := second.QueryEvidence(ctx, time.Now().Add(-time.Hour), 0, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(evidence).To(HaveLen(1))
		Expect(evidence[0].Repo.Get()).To(Equal("persist"))
	})
})

func mustTempDir() string {
	dir, err := os.MkdirTemp("", "compliance-cache-*")
	if err != nil {
		panic(err)
	}

	return dir
}

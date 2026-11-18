package compliance_test

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-compliance"
)

var _ = Describe("MemoryStore", func() {
	It("records the newest result per repository", func() {
		store := compliance.NewMemoryStore()
		repo := compliance.NewRepoName("repo")

		Expect(compliance.RecordTo(context.Background(), store, repo, compliance.RunResult{
			Verdict: compliance.VerdictFailed, Failures: 3, ProducerVersion: "v1",
		})).To(Succeed())

		Expect(compliance.RecordTo(context.Background(), store, repo, compliance.RunResult{
			Verdict: compliance.VerdictPassed, ProducerVersion: "v2",
		})).To(Succeed())

		evidence := store.Evidence()
		Expect(evidence).To(HaveLen(1))
		Expect(evidence[0].LastRunVerdict).To(Equal(compliance.VerdictPassed))
		Expect(evidence[0].Versions).To(Equal([]string{"v2"}))
	})

	It("ignores older results, keeping the newest run", func() {
		store := compliance.NewMemoryStore()
		repo := compliance.NewRepoName("repo")
		newer := time.Now()
		older := newer.Add(-time.Hour)

		Expect(compliance.RecordTo(context.Background(), store, repo, compliance.RunResult{
			At: newer, Verdict: compliance.VerdictPassed,
		})).To(Succeed())

		Expect(compliance.RecordTo(context.Background(), store, repo, compliance.RunResult{
			At: older, Verdict: compliance.VerdictFailed, Failures: 9,
		})).To(Succeed())

		Expect(store.Evidence()[0].LastRunVerdict).To(Equal(compliance.VerdictPassed))
	})

	It("fills zero counters: runs default to 1, fail rate derives", func() {
		store := compliance.NewMemoryStore()

		Expect(compliance.RecordTo(context.Background(), store, compliance.NewRepoName("r"), compliance.RunResult{
			StepRows: 10, Failures: 2,
		})).To(Succeed())

		record := store.Evidence()[0]
		Expect(record.Runs).To(Equal(int64(1)))
		Expect(record.FailRate).To(Equal(0.2))
		Expect(record.LastRunAt).NotTo(BeZero(), "record time defaults to now")
	})

	It("keeps first-write order across repositories", func() {
		store := compliance.NewMemoryStore()
		for _, name := range []string{"c", "a", "b"} {
			Expect(compliance.RecordTo(context.Background(), store, compliance.NewRepoName(name),
				compliance.RunResult{})).To(Succeed())
		}

		names := make([]string, 0, len(store.Evidence()))
		for _, record := range store.Evidence() {
			names = append(names, record.Repo.Get())
		}

		Expect(names).To(Equal([]string{"c", "a", "b"}))
	})

	It("is safe for concurrent reporters", func() {
		store := compliance.NewMemoryStore()

		var wg sync.WaitGroup
		for range 16 {
			wg.Go(func() {
				Expect(compliance.RecordTo(context.Background(), store,
					compliance.NewRepoName("concurrent"), compliance.RunResult{
						Verdict: compliance.VerdictPassed,
					})).To(Succeed())
			})
		}

		wg.Wait()

		Expect(store.Evidence()).To(HaveLen(1))
	})
})

var _ = Describe("FileStore", func() {
	var (
		dir  string
		path string
	)

	BeforeEach(func() {
		var err error

		dir, err = os.MkdirTemp("", "compliance-evidence-*")
		Expect(err).NotTo(HaveOccurred())

		path = filepath.Join(dir, "compliance.json")

		DeferCleanup(func() { _ = os.RemoveAll(dir) })
	})

	It("starts a fresh valid document when none exists", func() {
		store := compliance.NewFileStore(path)

		report, err := store.Load(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(report.SchemaVersion).To(Equal(1))
		Expect(report.Repos).To(Equal([]compliance.V1Repo{}))
	})

	It("records runs into the document, newest-wins, worst-first", func() {
		store := compliance.NewFileStore(path)
		ctx := context.Background()

		Expect(store.WriteEvidence(ctx, compliance.NewRepoName("ok-repo"), compliance.RunResult{
			Verdict: compliance.VerdictPassed, ProducerVersion: "v1",
		})).To(Succeed())

		Expect(store.WriteEvidence(ctx, compliance.NewRepoName("bad-repo"), compliance.RunResult{
			Verdict: compliance.VerdictFailed, Failures: 2, StepRows: 4,
		})).To(Succeed())

		report, err := store.Load(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(report.Repos).To(HaveLen(2))
		Expect(report.Repos[0].Repo).To(Equal("bad-repo"), "failing sorts first")
		Expect(report.Repos[0].State).To(Equal("failing"))
		Expect(report.Repos[0].Runs).To(Equal(int64(1)))
		Expect(report.Repos[0].FailRate).To(Equal(0.5))
		Expect(report.Repos[1].Repo).To(Equal("ok-repo"))
		Expect(report.Repos[1].State).To(Equal("compliant"))

		By("ignoring an older result")

		stale := time.Now().Add(-time.Hour)
		Expect(store.WriteEvidence(ctx, compliance.NewRepoName("bad-repo"), compliance.RunResult{
			At: stale, Verdict: compliance.VerdictPassed,
		})).To(Succeed())

		report, err = store.Load(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(report.Repos[0].LastRunVerdict).To(Equal("failed"))
	})

	It("writes atomically: readers never see a torn document", func() {
		store := compliance.NewFileStore(path)
		ctx := context.Background()

		var wg sync.WaitGroup

		wg.Add(2)
		go func() {
			defer wg.Done()

			for range 25 {
				Expect(store.WriteEvidence(ctx, compliance.NewRepoName("writer"), compliance.RunResult{
					Verdict: compliance.VerdictPassed, ProducerVersion: "v",
				})).To(Succeed())
			}
		}()
		go func() {
			defer wg.Done()

			for range 25 {
				report, err := store.Load(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(report.SchemaVersion).To(Equal(1))
			}
		}()

		wg.Wait()
	})

	It("preserves an existing piped document's other repositories", func() {
		existing := compliance.BuildV1Report([]compliance.RepoEvidence{{
			Repo: compliance.NewRepoName("piped"), Runs: 9, LastRunAt: time.Now(),
		}}, time.Now().Add(-time.Hour), compliance.DefaultStaleAfter, "buildflow", time.Now())

		encoded, err := compliance.MarshalV1(existing)
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(path, encoded, 0o644)).To(Succeed())

		store := compliance.NewFileStore(path)
		Expect(store.WriteEvidence(context.Background(), compliance.NewRepoName("sdk"), compliance.RunResult{
			Verdict: compliance.VerdictPassed,
		})).To(Succeed())

		report, err := store.Load(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(report.Repos).To(HaveLen(2))
	})

	It("rejects a corrupted document rather than overwriting it", func() {
		Expect(os.WriteFile(path, []byte("{corrupt"), 0o644)).To(Succeed())

		store := compliance.NewFileStore(path)
		err := store.WriteEvidence(context.Background(), compliance.NewRepoName("x"), compliance.RunResult{})
		Expect(err).To(MatchError(compliance.ErrParse))
	})

	It("reports read failures for unreadable files", func() {
		blocker := filepath.Join(dir, "not-a-dir")
		Expect(os.WriteFile(blocker, []byte("file in the way"), 0o644)).To(Succeed())

		store := compliance.NewFileStore(filepath.Join(blocker, "compliance.json"))
		_, err := store.Load(context.Background())
		Expect(err).To(MatchError(compliance.ErrIO))
	})
})

var _ = Describe("Record default target", func() {
	It("honors GO_COMPLIANCE_EVIDENCE_FILE", func() {
		dir, err := os.MkdirTemp("", "compliance-env-*")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { _ = os.RemoveAll(dir) })

		path := filepath.Join(dir, "evidence.json")
		Expect(os.Setenv(compliance.EnvEvidenceFile, path)).To(Succeed())
		DeferCleanup(func() { _ = os.Unsetenv(compliance.EnvEvidenceFile) })

		Expect(compliance.Record(context.Background(), compliance.NewRepoName("env-repo"),
			compliance.RunResult{Verdict: compliance.VerdictPassed})).To(Succeed())

		report, err := compliance.NewFileStore(path).Load(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(report.Repos).To(HaveLen(1))
		Expect(report.Repos[0].Repo).To(Equal("env-repo"))
	})
})

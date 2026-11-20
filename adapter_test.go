package compliance_test

import (
	"context"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-compliance"
)

var _ = Describe("V1ReportToEvidence", func() {
	It("re-derives every state of the real producer fixture identically", func() {
		_, report := loadGoldenFixture()
		evidence := compliance.V1ReportToEvidence(report, 0)

		Expect(evidence).To(HaveLen(len(report.Repos)))

		// The producer derived states at document generation time with its
		// own staleAfter; re-deriving from the normalized evidence under
		// the same conditions must reproduce every single state - the
		// strongest compatibility assertion available.
		generatedAt := compliance.ParseWireTime(report.GeneratedAt)
		staleAfter, err := time.ParseDuration(report.StaleAfter)
		Expect(err).NotTo(HaveOccurred())

		for i, row := range report.Repos {
			Expect(compliance.DeriveState(evidence[i], generatedAt, staleAfter).String()).
				To(Equal(row.State), "repo %s: re-derived state diverges from producer state", row.Repo)
		}
	})

	It("truncates worst steps to the requested limit", func() {
		_, report := loadGoldenFixture()
		evidence := compliance.V1ReportToEvidence(report, 1)

		for _, record := range evidence {
			if len(record.WorstSteps) > 0 {
				Expect(record.WorstSteps).To(HaveLen(1))
			}
		}
	})
})

var _ = Describe("FileSource", func() {
	var dir string

	BeforeEach(func() {
		var err error

		dir, err = os.MkdirTemp("", "compliance-filesource-*")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { _ = os.RemoveAll(dir) })
	})

	It("reads the piped document and normalizes it", func() {
		data, report := loadGoldenFixture()
		path := filepath.Join(dir, "compliance.json")
		Expect(os.WriteFile(path, data, 0o644)).To(Succeed())

		evidence, err := compliance.NewFileSource(path).QueryEvidence(
			context.Background(), time.Time{}, 0, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(evidence).To(HaveLen(len(report.Repos)))
	})

	It("names itself file, and the buildflow preset buildflow", func() {
		Expect(compliance.NewFileSource("x").Name().Get()).To(Equal("file"))
		Expect(compliance.NewBuildFlowFileSource().Name().Get()).To(Equal("buildflow"))
		Expect(compliance.NewBuildFlowFileSource().Path()).To(Equal(compliance.DefaultEvidenceFile))
	})

	It("errors with guidance when the file does not exist yet", func() {
		_, err := compliance.NewFileSource(filepath.Join(dir, "absent.json")).QueryEvidence(
			context.Background(), time.Time{}, 0, "")
		Expect(err).To(MatchError(compliance.ErrIO))
		Expect(err.Error()).To(ContainSubstring("does not exist yet"))
	})

	It("rejects a corrupted document", func() {
		path := filepath.Join(dir, "corrupt.json")
		Expect(os.WriteFile(path, []byte("{nope"), 0o644)).To(Succeed())

		_, err := compliance.NewFileSource(path).QueryEvidence(context.Background(), time.Time{}, 0, "")
		Expect(err).To(MatchError(compliance.ErrParse))
	})

	It("backs off repeated reads of a missing file", func() {
		path := filepath.Join(dir, "absent.json")
		source := compliance.NewFileSource(path).WithMissingBackoff(100 * time.Millisecond)

		_, err := source.QueryEvidence(context.Background(), time.Time{}, 0, "")
		Expect(err).To(MatchError(compliance.ErrIO))

		By("planting a corrupt document inside the backoff window")
		Expect(os.WriteFile(path, []byte("{nope"), 0o644)).To(Succeed())

		_, err = source.QueryEvidence(context.Background(), time.Time{}, 0, "")
		Expect(err).To(MatchError(compliance.ErrIO),
			"the cached miss must still be served (a parse error would prove the disk was consulted)")

		By("consulting the file again once the TTL expires")
		time.Sleep(200 * time.Millisecond)

		data, report := loadGoldenFixture()
		Expect(os.WriteFile(path, data, 0o644)).To(Succeed())

		evidence, err := source.QueryEvidence(context.Background(), time.Time{}, 0, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(evidence).To(HaveLen(len(report.Repos)))
	})
})

var _ = Describe("CommandSource", func() {
	var dir string

	BeforeEach(func() {
		var err error

		dir, err = os.MkdirTemp("", "compliance-cmdsource-*")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { _ = os.RemoveAll(dir) })
	})

	fixturePath := func() string {
		data, _ := loadGoldenFixture()
		path := filepath.Join(dir, "fixture.json")
		Expect(os.WriteFile(path, data, 0o644)).To(Succeed())

		return path
	}

	It("spawns the producer and decodes its stdout", func() {
		evidence, err := compliance.NewCommandSource("cat", fixturePath()).QueryEvidence(
			context.Background(), time.Time{}, 0, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(evidence).To(HaveLen(341))
	})

	It("maps the window onto producer flags", func() {
		argsPath := filepath.Join(dir, "args.txt")

		script := "printf '%s\\n' \"$@\" > " + argsPath + "; cat " + fixturePath()
		source := compliance.NewCommandSource("sh", "-c", script, "sh")

		_, err := source.QueryEvidence(context.Background(),
			time.Now().Add(-7*24*time.Hour), 5, "samber-linter")
		Expect(err).NotTo(HaveOccurred())

		raw, err := os.ReadFile(argsPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(raw)).To(ContainSubstring("--since"))
		Expect(string(raw)).To(ContainSubstring("168h0m0s"))
		Expect(string(raw)).To(ContainSubstring("--worst-steps"))
		Expect(string(raw)).To(ContainSubstring("5"))
		Expect(string(raw)).To(ContainSubstring("--step"))
		Expect(string(raw)).To(ContainSubstring("samber-linter"))
	})

	It("wraps producer failures with stderr context", func() {
		_, err := compliance.NewCommandSource("sh", "-c", "echo boom >&2; exit 3").QueryEvidence(
			context.Background(), time.Time{}, 0, "")
		Expect(err).To(MatchError(compliance.ErrIO))
		Expect(err.Error()).To(ContainSubstring("boom"))
		Expect(err.Error()).To(ContainSubstring("stderr"))
	})

	It("names itself after the command binary", func() {
		Expect(compliance.NewCommandSource("/usr/local/bin/buildflow", "compliance", "--json").
			Name().Get()).To(Equal("buildflow"))
		Expect(compliance.NewBuildFlowCommandSource().Name().Get()).To(Equal("buildflow"))
	})

	It("rejects producer output with a foreign schema", func() {
		path := filepath.Join(dir, "v2.json")
		Expect(os.WriteFile(path, []byte(`{"schema_version":2,"repos":[]}`), 0o644)).To(Succeed())

		_, err := compliance.NewCommandSource("cat", path).QueryEvidence(
			context.Background(), time.Time{}, 0, "")
		Expect(err).To(MatchError(compliance.ErrUnsupportedSchema))
	})
})

package compliance_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-compliance"
)

// goldenFixture is real BuildFlow producer output (341 repositories,
// captured live 2026-11-17) with repository names redacted; every other
// byte of shape is real. Regenerate with scripts/refresh-fixture.sh.
const goldenFixture = "buildflow-v1-fixture.json"

func loadGoldenFixture() ([]byte, compliance.V1Report) {
	data, err := os.ReadFile(filepath.Join("testdata", goldenFixture))
	Expect(err).NotTo(HaveOccurred())

	report, err := compliance.DecodeV1Report(data)
	Expect(err).NotTo(HaveOccurred())

	return data, report
}

var _ = Describe("Golden: BuildFlow producer compatibility", Ordered, func() {
	var (
		data   []byte
		report compliance.V1Report
	)

	BeforeAll(func() {
		data, report = loadGoldenFixture()
	})

	It("accepts the real producer document", func() {
		Expect(report.SchemaVersion).To(Equal(1))
		Expect(report.Repos).To(HaveLen(341))
	})

	It("carries only the three v1 producer states", func() {
		states := map[compliance.State]int{}

		for _, repo := range report.Repos {
			state, err := compliance.ParseState(repo.State)
			Expect(err).NotTo(HaveOccurred(), "repo %s: state %q", repo.Repo, repo.State)

			states[state]++
		}

		Expect(states).To(HaveKey(compliance.StateFailing))
		Expect(states).To(HaveKey(compliance.StateCompliant))
		Expect(states).To(SatisfyAny(
			HaveKey(compliance.StateStale),
			Not(HaveKey(compliance.StateStale)),
		), "stale presence depends on the capture moment")
		Expect(states).NotTo(HaveKey(compliance.StateNeverCovered),
			"the v1 producer never emits never-covered; it is a consumer-side extension")
	})

	It("carries only known verdicts", func() {
		for _, repo := range report.Repos {
			Expect(repo.LastRunVerdict).To(SatisfyAny(
				Equal("passed"), Equal("failed"), Equal("")))
		}
	})

	It("round-trips the fixture byte-identically through our canonical marshal", func() {
		encoded, err := compliance.MarshalV1(report)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(encoded)).To(Equal(string(data)),
			"marshal(decode(fixture)) must reproduce the fixture bytes")
	})

	It("keeps every field through decode and re-decode", func() {
		encoded, err := compliance.MarshalV1(report)
		Expect(err).NotTo(HaveOccurred())

		again, err := compliance.DecodeV1Report(encoded)
		Expect(err).NotTo(HaveOccurred())
		Expect(again).To(Equal(report))
	})

	It("parses every timestamp", func() {
		Expect(compliance.ParseWireTime(report.GeneratedAt)).NotTo(BeZero())
		Expect(compliance.ParseWireTime(report.Since)).NotTo(BeZero())

		for _, repo := range report.Repos {
			if repo.LastRunAt != "" {
				Expect(compliance.ParseWireTime(repo.LastRunAt)).NotTo(BeZero(),
					"repo %s: unparseable last_run_at %q", repo.Repo, repo.LastRunAt)
			}
		}
	})
})

var _ = Describe("Golden: overview consumer mirror", func() {
	It("accepts producer output the way overview does (schema_version 1)", func() {
		_, report := loadGoldenFixture()
		Expect(report.SchemaVersion).To(Equal(compliance.SchemaVersionV1))
	})

	It("rejects schema_version 2 exactly like the overview decoder", func() {
		_, report := loadGoldenFixture()
		report.SchemaVersion = 2

		encoded, err := compliance.MarshalV1(report)
		Expect(err).NotTo(HaveOccurred())

		_, err = compliance.DecodeV1Report(encoded)
		Expect(err).To(MatchError(compliance.ErrUnsupportedSchema))
	})

	It("rejects a truncated document rather than guessing", func() {
		data, _ := loadGoldenFixture()
		_, err := compliance.DecodeV1Report(data[:len(data)/2])
		Expect(err).To(MatchError(compliance.ErrParse))
	})
})

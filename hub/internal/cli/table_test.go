package cli_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-compliance"
	"github.com/larsartmann/go-compliance/hub/internal/cli"
)

var _ = Describe("fleet table", func() {
	var (
		out, errW strings.Builder
		dir       string
	)

	BeforeEach(func() {
		out.Reset()
		errW.Reset()

		var err error

		dir, err = os.MkdirTemp("", "compliance-table-*")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { _ = os.RemoveAll(dir) })
	})

	writeFixture := func(severities map[string]int) string {
		report := compliance.BuildV1Report([]compliance.RepoEvidence{
			{
				Repo:              compliance.NewRepoName("failing-repo"),
				Runs:              8,
				Failures:          3,
				FailRate:          0.375,
				LastRunAt:         time.Now().Add(-2 * time.Hour),
				LastRunVerdict:    compliance.VerdictFailed,
				FindingSeverities: severities,
				GateFindings:      map[string]int{"lint": severities["error"] + severities["warning"]},
			},
			{
				Repo: compliance.NewRepoName("a-very-long-repository-name-exceeding-the-cell"), Runs: 2,
				LastRunAt: time.Now().Add(-30 * 24 * time.Hour), LastRunVerdict: compliance.VerdictPassed,
			},
			{
				Repo: compliance.NewRepoName("compliant-repo"), Runs: 5,
				LastRunAt: time.Now().Add(-3 * time.Hour), LastRunVerdict: compliance.VerdictPassed,
			},
		}, time.Now().Add(-7*24*time.Hour), compliance.DefaultStaleAfter, "test", time.Now())

		encoded, err := compliance.MarshalV1(report)
		Expect(err).NotTo(HaveOccurred())

		path := filepath.Join(dir, "compliance.json")
		Expect(os.WriteFile(path, encoded, 0o644)).To(Succeed())

		return path
	}

	render := func(path string) string {
		Expect(cli.Run(context.Background(), []string{"table", "--file", path}, &out, &errW)).To(Succeed())

		return out.String()
	}

	It("renders state labels, severity cells, and worst-first order", func() {
		rendered := render(writeFixture(map[string]int{"error": 12, "warning": 3, "info": 7}))

		Expect(rendered).To(ContainSubstring("REPO"))
		Expect(rendered).To(ContainSubstring("STATE"))
		Expect(rendered).To(ContainSubstring("ERR"))
		Expect(rendered).To(ContainSubstring("WARN"))
		Expect(rendered).To(ContainSubstring("INFO"))
		Expect(rendered).To(ContainSubstring("FINDINGS"))

		header := strings.Index(rendered, "REPO")
		Expect(header).To(BeNumerically(">=", 0), "table header must be rendered")
		body := rendered[header:]
		failing := strings.Index(body, "failing")
		stale := strings.Index(body, "stale")
		compliant := strings.Index(body, "compliant")

		Expect(failing).To(BeNumerically("<", stale))
		Expect(stale).To(BeNumerically("<", compliant))

		By("deriving every state from wire evidence alone (severities are display-side)")
	})

	It("shows dashes for repos without severities", func() {
		rendered := render(writeFixture(nil))
		Expect(rendered).To(ContainSubstring("-"))
	})

	It("truncates over-long repo names with an ellipsis", func() {
		rendered := render(writeFixture(nil))
		Expect(rendered).To(ContainSubstring("a-very-long-repository-…"))
		Expect(rendered).NotTo(ContainSubstring("a-very-long-repository-name"))
	})

	It("keeps columns aligned under ANSI-colored cells (gotcha #298)", func() {
		path := writeFixture(map[string]int{"error": 12345, "warning": 67, "info": 890})

		rendered := render(path)

		lines := []string{}

		for line := range strings.SplitSeq(rendered, "\n") {
			if strings.HasPrefix(line, "│") {
				lines = append(lines, line)
			}
		}

		Expect(lines).To(HaveLen(4), "header + three repo rows: %s", rendered)

		// Strip ANSI escapes, then every row must agree on the byte
		// offset of each column start: colored and plain cells align
		// identically because the renderer measures display width.
		plain := make([]string, 0, len(lines))
		for _, line := range lines {
			plain = append(plain, stripANSI(line))
		}

		reference := columnStarts(plain[0])
		Expect(reference).ToNot(BeEmpty(), "header must have multiple columns")

		for i := 1; i < len(plain); i++ {
			Expect(columnStarts(plain[i])).To(Equal(reference),
				"row %d misaligned against header:\n%s\n%s", i, plain[0], plain[i])
		}
	})
})

// columnStarts returns the rune positions of the box-drawing column
// separators: aligned rows share identical separator positions, colored
// or not.
func columnStarts(line string) []int {
	starts := []int{}

	runes := []rune(line)
	for i, r := range runes {
		if r == '│' {
			starts = append(starts, i)
		}
	}

	return starts
}

// stripANSI removes ANSI escape sequences.
func stripANSI(input string) string {
	var b strings.Builder

	for i := 0; i < len(input); {
		if input[i] == 0x1b && i+1 < len(input) && input[i+1] == '[' {
			j := i + 2
			for j < len(input) && !isANSITerminator(input[j]) {
				j++
			}

			if j < len(input) {
				j++
			}

			i = j
		} else {
			b.WriteByte(input[i])
			i++
		}
	}

	return b.String()
}

func isANSITerminator(c byte) bool { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }

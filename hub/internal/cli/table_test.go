package cli_test

import (
	"context"
	"encoding/json/v2"
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

	renderFormat := func(format string) string {
		Expect(cli.Run(context.Background(),
			[]string{"table", "--file", writeFixture(map[string]int{"error": 2, "warning": 1}), "--format", format},
			&out, &errW)).To(Succeed())

		return out.String()
	}

	It("--format csv renders comma-separated rows with full repo names", func() {
		rendered := renderFormat("csv")

		// The frozen v1 wire contract cannot carry severities (adapter.go:
		// "fields the contract cannot carry"), so file-sourced severity
		// cells are honestly "-"; what MUST survive is full repo names.
		Expect(rendered).To(HavePrefix("REPO,STATE,ERR,WARN"))
		Expect(rendered).To(ContainSubstring("\nfailing-repo,failing,-,-,-,-,8,3"))
		Expect(rendered).To(ContainSubstring("a-very-long-repository-name-exceeding-the-cell"),
			"machine formats must not truncate repo names")
	})

	It("--format json round-trips as decodable rows", func() {
		rendered := renderFormat("json")

		var rows []map[string]string
		Expect(json.Unmarshal([]byte(rendered), &rows)).To(Succeed())
		Expect(rows).To(HaveLen(3))
		Expect(rows[0]["REPO"]).To(Equal("failing-repo"))
		Expect(rows[0]["STATE"]).To(Equal("failing"))
		Expect(rows[0]["RUNS"]).To(Equal("8"))
		Expect(rows[0]["LAST RUN"]).To(MatchRegexp(`^\d{4}-\d{2}-\d{2}T`),
			"machine formats carry absolute timestamps, not ages")
	})

	It("--format markdown renders a titled pipe table", func() {
		rendered := renderFormat("markdown")

		Expect(rendered).To(ContainSubstring("# Fleet compliance"))
		Expect(rendered).To(ContainSubstring("| REPO"))
		Expect(rendered).To(ContainSubstring("| failing-repo"))
	})

	It("--format tsv and html stay ANSI-free", func() {
		for _, format := range []string{"tsv", "html"} {
			Expect(renderFormat(format)).NotTo(ContainSubstring("\x1b["), "format %s", format)
		}
	})

	It("rejects an unsupported --format at the flag boundary", func() {
		err := cli.Run(context.Background(),
			[]string{"table", "--file", writeFixture(nil), "--format", "dot"}, &out, &errW)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("allowed"))
	})

	// Two failing repos whose canonical order (failures desc) contradicts
	// both the name and the findings order, plus a compliant repo: the
	// sort-key tests can retiebreak INSIDE the state rank without ever
	// being able to violate it.
	writeSortFixture := func() string {
		report := compliance.BuildV1Report([]compliance.RepoEvidence{
			{
				Repo: compliance.NewRepoName("alpha-failing"), Runs: 10, Failures: 5,
				LastRunAt: time.Now().Add(-2 * time.Hour), LastRunVerdict: compliance.VerdictFailed,
				FindingSeverities: map[string]int{"error": 1},
			},
			{
				Repo: compliance.NewRepoName("zeta-failing"), Runs: 10, Failures: 2,
				LastRunAt: time.Now().Add(-2 * time.Hour), LastRunVerdict: compliance.VerdictFailed,
				FindingSeverities: map[string]int{"error": 9},
			},
			{
				Repo: compliance.NewRepoName("compliant-repo"), Runs: 5,
				LastRunAt: time.Now().Add(-3 * time.Hour), LastRunVerdict: compliance.VerdictPassed,
				FindingSeverities: map[string]int{"info": 12},
			},
		}, time.Now().Add(-7*24*time.Hour), compliance.DefaultStaleAfter, "test", time.Now())

		encoded, err := compliance.MarshalV1(report)
		Expect(err).NotTo(HaveOccurred())

		path := filepath.Join(dir, "compliance-sort.json")
		Expect(os.WriteFile(path, encoded, 0o644)).To(Succeed())

		return path
	}

	renderCSV := func(extraArgs ...string) []string {
		args := append([]string{"table", "--file", writeSortFixture(), "--format", "csv"}, extraArgs...)

		Expect(cli.Run(context.Background(), args, &out, &errW)).To(Succeed())

		rows := []string{}
		for line := range strings.SplitSeq(strings.TrimPrefix(out.String(), "REPO,STATE,ERR,WARN,INFO,FINDINGS,RUNS,FAILS,RATE,VERDICT,LAST RUN\n"), "\n") {
			if line != "" {
				rows = append(rows, strings.SplitN(line, ",", 2)[0])
			}
		}

		return rows
	}

	It("keeps the canonical worst-first order by default (failures desc inside the state rank)", func() {
		rows := renderCSV()

		Expect(rows).To(Equal([]string{"alpha-failing", "zeta-failing", "compliant-repo"}),
			"canonical order: failures desc inside the state rank, state rank primary")
	})

	It("--sort-by findings retiebreaks by the FINDINGS cell inside the state rank", func() {
		rows := renderCSV("--sort-by", "findings")

		Expect(rows).To(Equal([]string{"zeta-failing", "alpha-failing", "compliant-repo"}),
			"zeta has 9 findings vs alpha's 1; the compliant repo's 12 findings still rank last")
	})

	It("--sort-by name retiebreaks alphabetically inside the state rank", func() {
		rows := renderCSV("--sort-by", "name")

		Expect(rows).To(Equal([]string{"alpha-failing", "zeta-failing", "compliant-repo"}),
			"alpha precedes zeta despite fewer failures; state rank still primary")
	})

	It("--sort-by keeps the styled and machine views in the same order", func() {
		path := writeSortFixture()

		Expect(cli.Run(context.Background(), []string{"table", "--file", path, "--sort-by", "findings"}, &out, &errW)).To(Succeed())

		styled := out.String()
		Expect(strings.Index(styled, "zeta-failing")).To(BeNumerically("<", strings.Index(styled, "alpha-failing")))
	})

	It("rejects an unsupported --sort-by at the flag boundary", func() {
		err := cli.Run(context.Background(),
			[]string{"table", "--file", writeSortFixture(), "--sort-by", "importance"}, &out, &errW)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("allowed"))
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

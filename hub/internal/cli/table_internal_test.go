package cli

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-output"

	"github.com/larsartmann/go-compliance"
)

var _ = Describe("table cell helpers", func() {
	var colors *colorService

	BeforeEach(func() {
		colors = newColorService(output.ColorModeNever)
	})

	DescribeTable("severity cells",
		func(severities map[string]int, errs, warns, infos, total string) {
			e, w, i, t := severityCells(colors, severities)
			Expect(e).To(Equal(errs))
			Expect(w).To(Equal(warns))
			Expect(i).To(Equal(infos))
			Expect(t).To(Equal(total))
		},
		Entry("nil severities render all dashes", nil, "-", "-", "-", "-"),
		Entry("counts render per severity, total sums everything",
			map[string]int{"error": 12, "warning": 3, "info": 7}, "12", "3", "7", "22"),
		Entry("unknown severities count toward the total only",
			map[string]int{"unknown": 4}, "-", "-", "-", "4"),
		Entry("zero-count severities render dashes",
			map[string]int{"error": 0}, "-", "-", "-", "-"),
	)

	DescribeTable("state labels",
		func(state compliance.State, expected string) {
			Expect(stateLabel(colors, state)).To(Equal(expected))
		},
		Entry("failing", compliance.StateFailing, "failing"),
		Entry("stale", compliance.StateStale, "stale"),
		Entry("compliant", compliance.StateCompliant, "compliant"),
		Entry("never-covered", compliance.StateNeverCovered, "never-covered"),
		Entry("unknown passes through", compliance.State("weird"), "weird"),
	)

	DescribeTable("cell truncation",
		func(in string, width int, expected string) {
			Expect(truncateCell(in, width)).To(Equal(expected))
		},
		Entry("short names pass through", "go-finding", 24, "go-finding"),
		Entry("exact-width names pass through", "abcdef", 6, "abcdef"),
		Entry("long names cut with an ellipsis", "a-very-long-repository-name", 24, "a-very-long-repository-…"),
	)

	DescribeTable("--sort-by parsing",
		func(raw string, expectedErr bool) {
			order, err := parseSortBy(raw)
			if expectedErr {
				Expect(err).To(MatchError(ContainSubstring("invalid --sort-by value")))
				Expect(order).To(Equal(sortBy("")))

				return
			}

			Expect(err).NotTo(HaveOccurred())
			Expect(order).To(Equal(sortBy(raw)))
		},
		Entry("findings", "findings", false),
		Entry("failures", "failures", false),
		Entry("name", "name", false),
		Entry("empty rejected (the flag default substitutes failures)", "", true),
		Entry("importance is producer-only", "importance", true),
		Entry("garbage rejected", "bogus", true),
	)
})

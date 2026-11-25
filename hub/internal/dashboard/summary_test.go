package dashboard_test

import (
	"bytes"
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-compliance/hub/internal/dashboard"
)

var _ = Describe("fleet summary", func() {
	renderSummary := func(counts map[string]int, total int, selected string) string {
		GinkgoHelper()

		var rendered bytes.Buffer
		Expect(dashboard.FleetSummary(counts, total, selected).Render(context.Background(), &rendered)).To(Succeed())

		return rendered.String()
	}

	Describe("the headline", func() {
		It("states the failing ratio in one sentence, the failing count as the only colored word", func() {
			html := renderSummary(map[string]int{"failing": 157, "compliant": 47}, 204, "all")

			Expect(html).To(MatchRegexp(
				`text-red-700 dark:text-red-400">157</span>\s*<span>of 204 repos are failing</span>`))
		})

		It("agrees with a single failing repo in the singular", func() {
			html := renderSummary(map[string]int{"failing": 1, "compliant": 203}, 204, "all")

			Expect(html).To(MatchRegexp(`>1</span>\s*<span>of 204 repos is failing</span>`))
		})

		It("announces a clean fleet without a warning color", func() {
			html := renderSummary(map[string]int{"compliant": 47}, 47, "all")

			Expect(html).To(MatchRegexp(`>All</span>\s*<span class="text-inherit">47</span>\s*<span>repos are compliant</span>`))
			Expect(html).NotTo(ContainSubstring("text-red-"))
		})

		It("says so when no evidence exists yet and draws no bar", func() {
			html := renderSummary(map[string]int{}, 0, "all")

			Expect(html).To(ContainSubstring("No fleet evidence yet"))
			Expect(html).NotTo(ContainSubstring(`role="img"`))
		})
	})

	Describe("the state bar", func() {
		It("describes the split in exact counts, worst first, for assistive technology", func() {
			html := renderSummary(
				map[string]int{"failing": 157, "compliant": 46, "stale": 1, "never-covered": 0}, 204, "all")

			Expect(html).To(ContainSubstring(`aria-label="157 failing, 1 stale, 46 compliant of 204 repos"`))
			Expect(html).To(ContainSubstring(`<div class="flex h-1.5`))
		})

		It("sizes every share by its count, so the shares always add up", func() {
			html := renderSummary(map[string]int{"failing": 1, "compliant": 3}, 4, "all")

			Expect(html).To(ContainSubstring(`style="flex:1 1 0%;"`))
			Expect(html).To(ContainSubstring(`style="flex:3 1 0%;"`))
		})

		It("leaves out the states that hold no repo", func() {
			html := renderSummary(map[string]int{"compliant": 5}, 5, "all")

			Expect(html).To(ContainSubstring(`style="flex:5 1 0%;"`))
			Expect(html).NotTo(ContainSubstring("bg-red-"))
		})
	})

	Describe("the filter chips", func() {
		It("carry the count each one selects", func() {
			html := renderSummary(map[string]int{"failing": 157, "compliant": 47}, 204, "all")

			Expect(html).To(ContainSubstring(">All"))
			Expect(html).To(MatchRegexp(`>All\s*<span[^>]*>204</span>`))
			Expect(html).To(MatchRegexp(`>Failing\s*<span[^>]*>157</span>`))
			Expect(html).To(MatchRegexp(`>Stale\s*<span[^>]*>0</span>`))
			Expect(html).To(MatchRegexp(`>Never covered\s*<span[^>]*>0</span>`))
			Expect(html).To(MatchRegexp(`>Compliant\s*<span[^>]*>47</span>`))
		})

		It("show the selected state on first paint, without JavaScript", func() {
			html := renderSummary(map[string]int{"failing": 3}, 3, "failing")

			Expect(html).To(MatchRegexp(` class="min-h-8[^"]*bg-accent-50[^"]*" aria-pressed="true">Failing`))
			Expect(html).To(MatchRegexp(` class="min-h-8[^"]*bg-white[^"]*" aria-pressed="false">All`))
			Expect(html).To(MatchRegexp(` class="min-h-8[^"]*bg-white[^"]*" aria-pressed="false">Compliant`))
		})

		It("keep the pressed state and the look in step with the filter signal", func() {
			html := renderSummary(map[string]int{"failing": 3}, 3, "failing")

			Expect(html).To(MatchRegexp(
				`data-attr:aria-pressed="\$stateFilter===(?:&#39;|')failing(?:&#39;|') \? (?:&#39;|')true(?:&#39;|') : (?:&#39;|')false(?:&#39;|')"`))
			Expect(html).To(MatchRegexp(`data-on-click="\$stateFilter=(?:&#39;|')stale(?:&#39;|')"`))

			// DataStar's class action toggles an object of class names to
			// booleans; a class string would be iterated character by
			// character and the selection would never show.
			Expect(html).To(MatchRegexp(
				`data-class="\{ (?:&#39;|')border-accent-600 bg-accent-50[^"=]*: \$stateFilter===(?:&#39;|')failing(?:&#39;|'), `+
					`(?:&#39;|')border-gray-200 bg-white[^"=]*: \$stateFilter!==(?:&#39;|')failing(?:&#39;|') \}"`))
		})
	})
})

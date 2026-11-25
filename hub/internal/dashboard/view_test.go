package dashboard_test

import (
	"bytes"
	"context"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-compliance"
	"github.com/larsartmann/go-compliance/hub/internal/dashboard"
	"github.com/larsartmann/go-compliance/hub/internal/events"
)

var _ = Describe("fleet page render", func() {
	// pageWith builds a view model over the whole fleet, the way the page
	// handler does, so the summary speaks for every row.
	pageWith := func(rows ...events.RepoRow) dashboard.ViewModel {
		GinkgoHelper()

		return dashboard.ViewModel{
			Title:  "Fleet compliance",
			Rows:   rows,
			Counts: events.CountStates(rows),
			Total:  len(rows),
		}
	}

	It("renders the shell, fleet summary, live region, and patch targets", func() {
		model := dashboard.ViewModel{
			Title:     "Fleet compliance",
			StreamURL: "/compliance/events",
			ScriptSrc: "/datastar.js",
			Rows: []events.RepoRow{{
				Repo:        "go-finding",
				State:       compliance.StateFailing,
				Findings:    3,
				Severities:  map[string]int{"error": 3},
				WorstSteps:  []string{"lint", "test"},
				LastRunAt:   time.Date(2026, 11, 17, 20, 0, 0, 0, time.UTC),
				LastVerdict: compliance.VerdictFailed,
			}},
			}
			model.Counts = events.CountStates(model.Rows)
			model.Total = len(model.Rows)

		var rendered bytes.Buffer
		Expect(dashboard.View(model).Render(context.Background(), &rendered)).To(Succeed())

		html := rendered.String()
		Expect(html).To(ContainSubstring("<title>Fleet compliance</title>"))
		Expect(html).To(ContainSubstring(`id="fleet-summary"`))
		Expect(html).To(ContainSubstring(`id="fleet-region"`))
		Expect(html).To(ContainSubstring(`id="fleet-rows"`))
		Expect(html).To(ContainSubstring(`data-repo="go-finding"`))
		Expect(html).To(ContainSubstring("of 1 repo is failing"))
		Expect(html).To(ContainSubstring("Never covered"))
		Expect(html).To(ContainSubstring("3 ERR"))
		Expect(html).To(ContainSubstring("2 worst steps"))
		Expect(html).To(ContainSubstring("<li>lint</li>"))
		Expect(html).To(ContainSubstring(`src="/datastar.js"`))
		Expect(html).To(ContainSubstring("/compliance/events"))
		Expect(html).To(ContainSubstring(`datetime="2026-11-17T20:00:00Z"`))
	})

	It("ships the relative-time tick script exactly once and ticks rows by attribute", func() {
		model := dashboard.ViewModel{
			Title:     "Fleet compliance",
			StreamURL: "/events",
			Rows: []events.RepoRow{
				{Repo: "one", State: compliance.StateCompliant, LastRunAt: time.Now()},
				{Repo: "two", State: compliance.StateFailing, LastRunAt: time.Now()},
			},
			RenderedAt: time.Now(),
		}

		var rendered bytes.Buffer
		Expect(dashboard.View(model).Render(context.Background(), &rendered)).To(Succeed())

		// The page-level carrier is the ONLY AutoRefresh instance: one
		// script, guarded by the library's window flag. Rows opt into
		// the 30s tick via the bare attribute (patches stay script-free).
		html := rendered.String()
		Expect(strings.Count(html, "if(!window.tcRelativeTimeAttached)")).To(Equal(1),
			"exactly one relative-time script (the carrier); html: %s", html)
		// 3 time elements (carrier + two rows) plus the script's own
		// querySelector literal.
		Expect(strings.Count(html, "data-tc-relative")).To(Equal(4),
			"two row cells plus the carrier; html: %s", html)
		// Row cells carry the bare attribute with no trailing script
		// (patches must stay script-free).
		Expect(strings.Count(html, "data-tc-relative>")).To(Equal(2),
			"row time cells; html: %s", html)
	})

	It("renders the empty state while no evidence exists", func() {
		var rendered bytes.Buffer
		Expect(dashboard.View(dashboard.ViewModel{Title: "Fleet compliance"}).
			Render(context.Background(), &rendered)).To(Succeed())

		Expect(rendered.String()).To(ContainSubstring("No compliance evidence yet"))
		Expect(rendered.String()).To(ContainSubstring("No fleet evidence yet"))
	})

	It("ships an icon, so the browser logs no 404 for /favicon.ico", func() {
		var rendered bytes.Buffer
		Expect(dashboard.View(dashboard.ViewModel{Title: "Fleet compliance"}).
			Render(context.Background(), &rendered)).To(Succeed())

		Expect(rendered.String()).To(ContainSubstring(`<link rel="icon" href="/favicon.svg">`))
	})

	It("keeps a word space between the freshness label and its relative time", func() {
		var rendered bytes.Buffer
		Expect(dashboard.View(dashboard.ViewModel{Title: "Fleet compliance", RenderedAt: time.Now()}).
			Render(context.Background(), &rendered)).To(Succeed())

		Expect(rendered.String()).To(MatchRegexp(`>rendered</span>\s*<time`))
	})

	It("emits the pre-paint theme script once", func() {
		var rendered bytes.Buffer
		Expect(dashboard.View(dashboard.ViewModel{Title: "Fleet compliance"}).
			Render(context.Background(), &rendered)).To(Succeed())

		Expect(strings.Count(rendered.String(), "localStorage.getItem('theme')")).To(Equal(1),
			"html: %s", rendered.String())
	})

	It("names the fleet table region and gives it the narrow-layout hook", func() {
		var rendered bytes.Buffer
		Expect(dashboard.View(pageWith(events.RepoRow{Repo: "one"})).
			Render(context.Background(), &rendered)).To(Succeed())

		Expect(rendered.String()).To(ContainSubstring(`aria-label="Fleet compliance by repository"`))
		Expect(rendered.String()).To(ContainSubstring("fleet-table"))
	})

	It("seeds the filter and the counts, so no expression reads a missing signal", func() {
		model := pageWith(
			events.RepoRow{Repo: "broken", State: compliance.StateFailing},
			events.RepoRow{Repo: "ok", State: compliance.StateCompliant},
		)
		model.StateFilter = "failing"

		var rendered bytes.Buffer
		Expect(dashboard.View(model).Render(context.Background(), &rendered)).To(Succeed())

		Expect(rendered.String()).To(ContainSubstring(
			`data-signals="{&#34;stateFilter&#34;:&#34;failing&#34;,&#34;total&#34;:2,&#34;counts&#34;:{&#34;compliant&#34;:1,&#34;failing&#34;:1}}"`))
		Expect(rendered.String()).To(MatchRegexp(`>Failing\s*<span[^>]*>1</span>`))
	})

	Describe("a state selection that holds no repo", func() {
		It("hides the table at first paint and names the empty selection", func() {
			model := pageWith(events.RepoRow{Repo: "broken", State: compliance.StateFailing})
			model.StateFilter = "stale"
			model.Rows = nil

			var rendered bytes.Buffer
			Expect(dashboard.View(model).Render(context.Background(), &rendered)).To(Succeed())

			html := rendered.String()
			Expect(html).To(MatchRegexp(`<table[^>]*style="display:none"`))
			Expect(html).To(MatchRegexp(`data-show="\$stateFilter===&#39;stale&#39; &amp;&amp; !\$counts\[&#39;stale&#39;\]">\s*<div class="text-center py-8"><p[^>]*>No stale repos</p>`))
			Expect(html).To(MatchRegexp(`data-show="\$stateFilter===&#39;compliant&#39; &amp;&amp; !\$counts\[&#39;compliant&#39;\]" style="display:none">`))
		})

		It("keeps the table visible for the all selection, even with no evidence", func() {
			var rendered bytes.Buffer
			Expect(dashboard.View(pageWith()).Render(context.Background(), &rendered)).To(Succeed())

			Expect(rendered.String()).NotTo(ContainSubstring(`style="display:none"`))
			Expect(rendered.String()).To(ContainSubstring("No compliance evidence yet"))
		})
	})

	It("keeps row patches script-free and inside the patch budget", func() {
		// FleetRow is the replace-patch unit: it re-renders inside every
		// update, so a script or a size regression here multiplies across
		// the whole stream. The templ-components pass (ThemeToggle,
		// ThemeScript) must stay page-level.
		var rowBuf bytes.Buffer
		Expect(dashboard.FleetRow(events.RepoRow{
			Repo:      "budget-probe",
			State:     compliance.StateFailing,
			LastRunAt: time.Now(),
		}).Render(context.Background(), &rowBuf)).To(Succeed())

		row := rowBuf.String()
		Expect(row).NotTo(ContainSubstring("<script"),
			"row patches must stay script-free; row: %s", row)
		Expect(len(row)).To(BeNumerically("<", 2048),
			"per-row patch budget breached: %d bytes; row: %s", len(row), row)
	})

	Describe("the findings cell", func() {
		renderRow := func(row events.RepoRow) string {
			GinkgoHelper()

			var rendered bytes.Buffer
			Expect(dashboard.FleetRow(row).Render(context.Background(), &rendered)).To(Succeed())

			return rendered.String()
		}

		DescribeTable("shows what a row with findings was flagged for",
			func(row events.RepoRow, want string) {
				html := renderRow(row)

				Expect(html).To(ContainSubstring(want), "html: %s", html)
				Expect(html).NotTo(ContainSubstring("—"),
					"a row with findings must never render as empty; html: %s", html)
			},
			Entry("a file-backed row with one finding falls back to its gate-findings count",
				events.RepoRow{Repo: "single", State: compliance.StateFailing, Findings: 1},
				"1 finding"),
			Entry("a file-backed row with many findings",
				events.RepoRow{Repo: "many", State: compliance.StateFailing, Findings: 85},
				"85 findings"),
			Entry("a cache-backed row keeps its severity breakdown",
				events.RepoRow{
					Repo:       "rich",
					State:      compliance.StateFailing,
					Findings:   3,
					Severities: map[string]int{"error": 3},
				},
				"3 ERR"),
		)

		It("stays muted for a row that has no findings", func() {
			Expect(renderRow(events.RepoRow{Repo: "clean", State: compliance.StateCompliant})).
				To(ContainSubstring("—"))
		})
	})
})

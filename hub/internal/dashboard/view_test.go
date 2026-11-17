package dashboard_test

import (
	"bytes"
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-compliance"
	"github.com/larsartmann/go-compliance/hub/internal/dashboard"
	"github.com/larsartmann/go-compliance/hub/internal/events"
)

var _ = Describe("fleet page render", func() {
	It("renders the shell, stat cards, live region, and patch targets", func() {
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

		var rendered bytes.Buffer
		Expect(dashboard.View(model).Render(context.Background(), &rendered)).To(Succeed())

		html := rendered.String()
		Expect(html).To(ContainSubstring("<title>Fleet compliance</title>"))
		Expect(html).To(ContainSubstring(`id="stat-cards"`))
		Expect(html).To(ContainSubstring(`id="fleet-region"`))
		Expect(html).To(ContainSubstring(`id="fleet-rows"`))
		Expect(html).To(ContainSubstring(`data-repo="go-finding"`))
		Expect(html).To(ContainSubstring("Total repos"))
		Expect(html).To(ContainSubstring("Never covered"))
		Expect(html).To(ContainSubstring("ERR 3"))
		Expect(html).To(ContainSubstring("worst step(s)"))
		Expect(html).To(ContainSubstring("<li>lint</li>"))
		Expect(html).To(ContainSubstring(`src="/datastar.js"`))
		Expect(html).To(ContainSubstring("/compliance/events"))
		Expect(html).To(ContainSubstring(`datetime="2026-11-17T20:00:00Z"`))
	})

	It("renders the empty state while no evidence exists", func() {
		var rendered bytes.Buffer
		Expect(dashboard.View(dashboard.ViewModel{Title: "Fleet compliance"}).
			Render(context.Background(), &rendered)).To(Succeed())

		Expect(rendered.String()).To(ContainSubstring("No compliance evidence yet"))
	})
})

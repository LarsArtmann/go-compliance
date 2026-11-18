package dashboard_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	datastartest "github.com/larsartmann/go-datastar/datastartest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-compliance"
	"github.com/larsartmann/go-compliance/hub/internal/dashboard"
	"github.com/larsartmann/go-compliance/hub/internal/events"
)

var _ = Describe("routes and content negotiation", func() {
	newBoardWithRows := func() *dashboard.Dashboard {
		hub := events.NewHub()
		_, err := hub.PublishSnapshot([]events.RepoRow{
			{Repo: "ok", State: compliance.StateCompliant, Severities: map[string]int{}},
			{Repo: "broken", State: compliance.StateFailing, Severities: map[string]int{}},
		})
		Expect(err).NotTo(HaveOccurred())

		return dashboard.New(hub)
	}

	It("wires page, stream, and script in one call", func() {
		board := newBoardWithRows()
		mux := http.NewServeMux()
		board.RegisterRoutes(mux)

		page := httptest.NewRecorder()
		mux.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/", nil))
		Expect(page.Code).To(Equal(http.StatusOK))
		Expect(page.Header().Get("Content-Type")).To(ContainSubstring("text/html"))
		Expect(page.Body.String()).To(ContainSubstring(`id="fleet-rows"`))
		Expect(page.Body.String()).To(ContainSubstring("data-repo=\"broken\""))

		script := httptest.NewRecorder()
		mux.ServeHTTP(script, httptest.NewRequest(http.MethodGet, "/datastar.js", nil))
		Expect(script.Code).To(Equal(http.StatusOK))

		// A real request lifecycle, not a shared recorder: the handler
		// goroutine writing while the test polls the same recorder is a
		// data race (httptest.ResponseRecorder is unsynchronized).
		events := datastartest.CollectWithTimeout(GinkgoTB(), mux, 2*time.Second,
			datastartest.WithPath("/events"))

		signals := false

		for _, evt := range events {
			signals = signals || evt.IsSignals()
		}

		Expect(signals).To(BeTrue(), "stream route produced no signals patch; events: %v", events)
	})

	It("serves JSON to clients whose Accept prefers it", func() {
		board := newBoardWithRows()
		mux := http.NewServeMux()
		board.RegisterRoutes(mux)

		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.Header.Set("Accept", "application/json;q=0.9, text/html;q=0.8")

		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, request)

		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
		Expect(recorder.Body.String()).To(ContainSubstring(`"repo":"broken"`))
		Expect(recorder.Body.String()).To(ContainSubstring(`"total":2`))
	})

	It("prefers HTML for browser and wildcard accepts", func() {
		board := newBoardWithRows()
		mux := http.NewServeMux()
		board.RegisterRoutes(mux)

		for _, accept := range []string{
			"",
			"*/*",
			"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			"application/json;q=0.8, text/html;q=0.9",
		} {
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.Header.Set("Accept", accept)

			recorder := httptest.NewRecorder()
			mux.ServeHTTP(recorder, request)

			Expect(recorder.Header().Get("Content-Type")).To(ContainSubstring("text/html"), "accept=%s", accept)
		}
	})

	It("filters the initial render server-side via ?state=", func() {
		board := newBoardWithRows()
		mux := http.NewServeMux()
		board.RegisterRoutes(mux)

		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/?state=failing", nil))

		body := recorder.Body.String()
		Expect(body).To(ContainSubstring("data-repo=\"broken\""))
		Expect(body).NotTo(ContainSubstring("data-repo=\"ok\""))
		Expect(body).To(ContainSubstring(`data-signals="{&#34;stateFilter&#34;:&#34;failing&#34;}"`))
	})

	It("renders filter chips wired to the stateFilter signal", func() {
		board := newBoardWithRows()
		mux := http.NewServeMux()
		board.RegisterRoutes(mux)

		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

		body := recorder.Body.String()
		Expect(body).To(ContainSubstring(`data-on-click="$stateFilter=&#39;failing&#39;"`))
		Expect(body).To(ContainSubstring(`data-on-click="$stateFilter=&#39;all&#39;"`))
		Expect(body).To(ContainSubstring(`aria-pressed="true"`))
		Expect(strings.Count(body, `data-show="$stateFilter===&#39;all&#39;`)).To(Equal(2))
	})

	It("supports custom route prefixes", func() {
		board := newBoardWithRows()
		mux := http.NewServeMux()
		board.RegisterRoutesAt(dashboard.Routes{
			Page:   "/compliance/",
			Stream: "/compliance/events",
			Script: "/compliance/datastar.js",
		}, mux)

		page := httptest.NewRecorder()
		mux.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/compliance/", nil))
		Expect(page.Code).To(Equal(http.StatusOK))
		Expect(page.Body.String()).To(ContainSubstring("/compliance/events"))
		Expect(page.Body.String()).To(ContainSubstring("/compliance/datastar.js"))
	})
})

var _ = Describe("?state= filter contract", func() {
	// One row per state: every filter has positive and negative matches.
	newFourStateBoard := func() *dashboard.Dashboard {
		hub := events.NewHub()
		_, err := hub.PublishSnapshot([]events.RepoRow{
			{Repo: "ok", State: compliance.StateCompliant, Severities: map[string]int{}},
			{Repo: "broken", State: compliance.StateFailing, Severities: map[string]int{}},
			{Repo: "quiet", State: compliance.StateStale, Severities: map[string]int{}},
			{Repo: "ghost", State: compliance.StateNeverCovered, Severities: map[string]int{}},
		})
		Expect(err).NotTo(HaveOccurred())

		return dashboard.New(hub)
	}

	filterBody := func(filter string) string {
		mux := http.NewServeMux()
		newFourStateBoard().RegisterRoutes(mux)

		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/?state="+filter, nil))

		Expect(recorder.Code).To(Equal(http.StatusOK))

		return recorder.Body.String()
	}

	expectOnly := func(body, present string) {
		Expect(body).To(ContainSubstring(`data-repo="`+present+`"`), "expected %s to be rendered", present)

		for _, repo := range []string{"ok", "broken", "quiet", "ghost"} {
			if repo == present {
				continue
			}

			Expect(body).NotTo(ContainSubstring(`data-repo="`+repo+`"`), "%s leaked through state=%s", repo, present)
		}
	}

	DescribeTable("accepts every core State constant",
		func(state compliance.State, present string) {
			expectOnly(filterBody(string(state)), present)
		},
		Entry("compliant", compliance.StateCompliant, "ok"),
		Entry("failing", compliance.StateFailing, "broken"),
		Entry("stale", compliance.StateStale, "quiet"),
		Entry("never-covered", compliance.StateNeverCovered, "ghost"),
	)

	It("keeps every row for all and for an invalid filter", func() {
		for _, filter := range []string{"all", "bogus"} {
			body := filterBody(filter)

			for _, repo := range []string{"ok", "broken", "quiet", "ghost"} {
				Expect(body).To(ContainSubstring(`data-repo="`+repo+`"`), "%s must survive state=%s", repo, filter)
			}
		}
	})

	It("renders an empty table for the dashboard-only unknown filter", func() {
		body := filterBody("unknown")
		Expect(body).NotTo(ContainSubstring(`data-repo="`))
	})
})

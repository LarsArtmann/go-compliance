package dashboard_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/larsartmann/go-datastar"
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

		// The self-hosted bundle URL is version-pinned: upgrades become new
		// cache keys, so no browser serves yesterday's client inside the
		// bundle's 24h Cache-Control window.
		Expect(page.Body.String()).To(ContainSubstring(`src="/datastar.js?v=` + datastar.Version() + `"`))

		script := httptest.NewRecorder()
		mux.ServeHTTP(script, httptest.NewRequest(http.MethodGet, "/datastar.js", nil))
		Expect(script.Code).To(Equal(http.StatusOK))

		// ScriptHandler answers conditional requests (ETag/304): the pinned
		// ETag comes from the handler, not a hand-rolled header set.
		etag := script.Header().Get("ETag")
		Expect(etag).NotTo(BeEmpty())

		conditional := httptest.NewRequest(http.MethodGet, "/datastar.js", nil)
		conditional.Header.Set("If-None-Match", etag)

		revalidated := httptest.NewRecorder()
		mux.ServeHTTP(revalidated, conditional)
		Expect(revalidated.Code).To(Equal(http.StatusNotModified))

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
		Expect(body).To(ContainSubstring(`data-signals="{&#34;stateFilter&#34;:&#34;failing&#34;,&#34;total&#34;:2,`))

		// The summary speaks for the whole fleet, not the narrowed rows.
		Expect(body).To(ContainSubstring(`>1</span> <span>of 2 repos is failing</span>`))
		Expect(body).To(MatchRegexp(`>Compliant\s*<span[^>]*>1</span>`))
	})

	It("renders filter chips wired to the stateFilter signal", func() {
		board := newBoardWithRows()
		mux := http.NewServeMux()
		board.RegisterRoutes(mux)

		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

		body := recorder.Body.String()
		Expect(body).To(ContainSubstring(`data-on:click="$stateFilter=&#39;failing&#39;"`))
		Expect(body).To(ContainSubstring(`data-on:click="$stateFilter=&#39;all&#39;"`))
		Expect(body).NotTo(ContainSubstring(`data-on-click=`))
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

	It("links the compiled stylesheet versioned and stays off the Tailwind Play CDN", func() {
		board := newBoardWithRows()
		mux := http.NewServeMux()
		board.RegisterRoutes(mux)

		page := httptest.NewRecorder()
		mux.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/", nil))
		body := page.Body.String()

		// The compiled artifact is content-hash versioned (immutable day
		// cache + fresh key per regeneration).
		Expect(body).To(ContainSubstring(`href="/app.css?v=`))

		// CSP regression: the Play CDN script and its inline tailwind.config
		// block are gone for good.
		Expect(body).NotTo(ContainSubstring("cdn.tailwindcss.com"))
		Expect(body).NotTo(ContainSubstring("tailwind.config"))

		// Dark mode ships as the pre-paint ThemeScript + a real toggle.
		Expect(body).To(ContainSubstring("classList.add('dark')"))
		Expect(body).To(ContainSubstring(`aria-label="Toggle dark mode"`))
	})

	It("serves the compiled stylesheet with ETag/304 revalidation", func() {
		board := newBoardWithRows()
		mux := http.NewServeMux()
		board.RegisterRoutes(mux)

		css := httptest.NewRecorder()
		mux.ServeHTTP(css, httptest.NewRequest(http.MethodGet, "/app.css", nil))
		Expect(css.Code).To(Equal(http.StatusOK))
		Expect(css.Header().Get("Content-Type")).To(ContainSubstring("text/css"))
		Expect(css.Header().Get("Cache-Control")).To(ContainSubstring("immutable"))
		Expect(css.Body.Bytes()).NotTo(BeEmpty())

		etag := css.Header().Get("ETag")
		Expect(etag).NotTo(BeEmpty())

		conditional := httptest.NewRequest(http.MethodGet, "/app.css", nil)
		conditional.Header.Set("If-None-Match", etag)

		revalidated := httptest.NewRecorder()
		mux.ServeHTTP(revalidated, conditional)
		Expect(revalidated.Code).To(Equal(http.StatusNotModified))
	})

	It("serves the site icon the page links, so browsers log no 404", func() {
		board := newBoardWithRows()
		mux := http.NewServeMux()
		board.RegisterRoutes(mux)

		icon := httptest.NewRecorder()
		mux.ServeHTTP(icon, httptest.NewRequest(http.MethodGet, "/favicon.svg", nil))
		Expect(icon.Code).To(Equal(http.StatusOK))
		Expect(icon.Header().Get("Content-Type")).To(Equal("image/svg+xml"))
		Expect(icon.Body.String()).To(HavePrefix("<svg"))

		conditional := httptest.NewRequest(http.MethodGet, "/favicon.svg", nil)
		conditional.Header.Set("If-None-Match", icon.Header().Get("ETag"))

		revalidated := httptest.NewRecorder()
		mux.ServeHTTP(revalidated, conditional)
		Expect(revalidated.Code).To(Equal(http.StatusNotModified))

		page := httptest.NewRecorder()
		mux.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/", nil))
		Expect(page.Body.String()).To(ContainSubstring(`<link rel="icon" href="/favicon.svg">`))
	})

	It("returns the styled 404 for unknown paths, not the dashboard", func() {
		board := newBoardWithRows()
		mux := http.NewServeMux()
		board.RegisterRoutes(mux)

		// The serve wiring: dashboard owns the root exactly, the host's
		// catch-all renders the 404.
		mux.Handle("/", board.NotFoundHandler())

		missing := httptest.NewRecorder()
		mux.ServeHTTP(missing, httptest.NewRequest(http.MethodGet, "/no/such/page", nil))
		Expect(missing.Code).To(Equal(http.StatusNotFound))
		Expect(missing.Header().Get("X-Robots-Tag")).To(Equal("noindex"))
		Expect(missing.Header().Get("Content-Type")).To(ContainSubstring("text/html"))
		Expect(missing.Body.String()).To(ContainSubstring("Page not found"))
		Expect(missing.Body.String()).To(ContainSubstring(`href="/app.css?v=`),
			"the 404 renders inside the styled shell")

		// The root itself still serves the fleet.
		root := httptest.NewRecorder()
		mux.ServeHTTP(root, httptest.NewRequest(http.MethodGet, "/", nil))
		Expect(root.Code).To(Equal(http.StatusOK))
		Expect(root.Body.String()).To(ContainSubstring(`id="fleet-rows"`))
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

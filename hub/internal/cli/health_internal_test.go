package cli

import (
	"context"
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-health"

	"github.com/larsartmann/go-compliance"
	"github.com/larsartmann/go-compliance/hub/internal/events"
)

// booted returns a probe over an empty-but-published fleet (verdict g2:
// the evidence gate passes), so spec bodies can control freshness and
// discovery stamps directly.
func booted(refresher *refresher) *health.Probe {
	hub := events.NewHub()
	_, err := hub.PublishSnapshot(nil)
	Expect(err).NotTo(HaveOccurred())

	probe, err := buildHealthProbe(context.Background(), "test", hub, refresher, nil, "")
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(probe.Shutdown)

	return probe
}

var _ = Describe("health check transition logging", func() {
	It("logs each check's boot baseline once, then exactly one line per flip", func() {
		logs := captureLogs()

		hook := checkTransitions()

		evaluate := func(statuses map[string]health.Status) {
			checks := map[string]health.Check{}
			for name, status := range statuses {
				checks[name] = health.Check{Status: status}
			}

			hook(health.Response{Status: health.StatusPass, Checks: checks})
		}

		five := map[string]health.Status{
			evidenceCheckName:         health.StatusPass,
			refreshFreshnessCheckName: health.StatusPass,
			discoveryCheckName:        health.StatusPass,
			sseBackpressureCheckName:  health.StatusPass,
			"cache-disk":              health.StatusOff,
		}

		evaluate(five)
		Expect(strings.Count(logs(), "health check status changed")).To(Equal(5),
			"the first batch logs all five checks once; got: %s", logs())

		evaluate(five)
		Expect(strings.Count(logs(), "health check status changed")).To(Equal(5),
			"a stable evaluation must not log")

		flipped := map[string]health.Status{
			evidenceCheckName:         health.StatusPass,
			refreshFreshnessCheckName: health.StatusPass,
			discoveryCheckName:        health.StatusPass,
			sseBackpressureCheckName:  health.StatusWarn,
			"cache-disk":              health.StatusOff,
		}

		evaluate(flipped)
		Expect(strings.Count(logs(), "health check status changed")).To(Equal(6),
			"exactly the sse-backpressure flip may log; got: %s", logs())
		Expect(logs()).To(ContainSubstring("check=sse-backpressure"))

		evaluate(flipped)
		Expect(strings.Count(logs(), "health check status changed")).To(Equal(6),
			"a stable evaluation must not log")
	})
})

var _ = Describe("readiness HTTP contract", func() {
	It("pins the payload shape over real HTTP: fail is 503, warn stays 200, checks carry the five names", func() {
		hub := events.NewHub()
		refresher := newRefresher(hub, nil, time.Hour, 24*time.Hour, 3, "")
		refresher.lastRefreshSuccess.Store(time.Now().UnixNano())
		refresher.markDiscoverySuccess()

		probe, err := buildHealthProbe(context.Background(), "contract-test", hub, refresher, nil, "")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(probe.Shutdown)

		mux := http.NewServeMux()
		probe.RegisterRoutes(mux, health.DefaultRoutes())

		before := httptest.NewRecorder()
		mux.ServeHTTP(before, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		Expect(
			before.Code,
		).To(Equal(http.StatusServiceUnavailable), "nothing published: the critical evidence gate must fail readiness")

		_, err = hub.PublishSnapshot(nil)
		Expect(err).NotTo(HaveOccurred())
		refresher.lastRefreshSuccess.Store(time.Now().Add(-refreshFreshnessWarn - time.Minute).UnixNano())

		var body []byte

		Eventually(func() int {
			recorder := httptest.NewRecorder()
			mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))

			body = recorder.Body.Bytes()

			return recorder.Code
		}, "2s", "100ms").Should(Equal(http.StatusOK),
			"warn (stale refresh) must degrade readiness, never fail it")

		var payload struct {
			Status  string `json:"status"`
			Version string `json:"version"`
			Checks  map[string]struct {
				Status string `json:"status"`
				Error  string `json:"error,omitempty"`
			} `json:"checks"`
		}

		Expect(json.Unmarshal(body, &payload)).To(Succeed())
		Expect(payload.Status).To(Equal("warn"))
		Expect(payload.Version).To(Equal("contract-test"))
		Expect(
			payload.Checks,
		).To(HaveLen(5), "evidence, refresh-freshness, discovery, sse-backpressure, cache-disk(off without a db dir)")
		Expect(payload.Checks[evidenceCheckName].Status).To(Equal("pass"))
		Expect(payload.Checks[refreshFreshnessCheckName].Status).To(Equal("warn"))
		Expect(payload.Checks[refreshFreshnessCheckName].Error).To(ContainSubstring("last successful refresh"))
		Expect(payload.Checks[discoveryCheckName].Status).To(Equal("pass"))
		Expect(payload.Checks[sseBackpressureCheckName].Status).To(Equal("pass"))
		Expect(payload.Checks["cache-disk"].Status).To(Equal("off"))
		Expect(payload.Checks["cache-disk"].Error).To(ContainSubstring("--db"),
			"the off row carries its enable recipe")
	})
})

// The README's health section promises "warn when their loop stalls past
// 10 minutes" for refresh-freshness and discovery. Pin both thresholds so
// a future constant edit fails here first and forces the README along.
var _ = Describe("README readiness thresholds", func() {
	It("pins refresh-freshness and discovery warn thresholds at 10 minutes", func() {
		Expect(refreshFreshnessWarn).To(Equal(10*time.Minute),
			"README.md 'Checks' section promises a 10-minute stall warning")
		Expect(discoveryStaleWarn).To(Equal(10*time.Minute),
			"README.md 'Checks' section promises a 10-minute stall warning")
	})

	It("pins the cache-disk free-space floor at 128 MiB", func() {
		Expect(cacheDiskMinFreeBytes).To(Equal(128<<20),
			"README.md 'Checks' section promises a 128 MiB floor for cache-disk")
	})
})

var _ = Describe("health probe sse-backpressure check", func() {
	It("warns on new drops since the last evaluation and clears when they stop", func() {
		hub := events.NewHub()

		_, unsubscribe := hub.SubscribeFeed()
		DeferCleanup(unsubscribe)

		refresher := newRefresher(hub, nil, time.Hour, 24*time.Hour, 3, "")
		refresher.lastRefreshSuccess.Store(time.Now().UnixNano())
		refresher.markDiscoverySuccess()

		probe, err := buildHealthProbe(context.Background(), "test", hub, refresher, nil, "")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(probe.Shutdown)

		Expect(probe.Evaluate(context.Background()).Checks[sseBackpressureCheckName].Status).
			To(Equal(health.StatusPass))

		row := events.RepoRow{Repo: "backpressure", State: compliance.StateFailing}
		for range 200 {
			_, _ = hub.PublishSnapshot([]events.RepoRow{row})
		}

		warned := probe.Evaluate(context.Background())
		Expect(warned.Checks[sseBackpressureCheckName].Status).To(Equal(health.StatusWarn))
		Expect(warned.Checks[sseBackpressureCheckName].Error).To(ContainSubstring("dropped"))
		Expect(warned.Checks[sseBackpressureCheckName].Error).To(ContainSubstring("subscribers 1"))

		Expect(probe.Evaluate(context.Background()).Checks[sseBackpressureCheckName].Status).
			To(Equal(health.StatusPass))
	})
})

var _ = Describe("health probe discovery check", func() {
	It("warns — never fails — while the daemon has never been contacted, then passes on contact", func() {
		refresher := newRefresher(nil, nil, time.Hour, 24*time.Hour, 3, "")
		probe := booted(refresher)

		before := probe.Evaluate(context.Background())
		Expect(before.Checks[discoveryCheckName].Status).To(Equal(health.StatusWarn))
		Expect(before.Checks[discoveryCheckName].Error).
			To(ContainSubstring(ErrDiscoveryNeverContacted.Error()))
		Expect(before.Status).To(Equal(health.StatusWarn))

		refresher.markDiscoverySuccess()

		after := probe.Evaluate(context.Background())
		Expect(after.Checks[discoveryCheckName].Status).To(Equal(health.StatusPass))
	})

	It("warns with the silent age once the last contact exceeds the threshold", func() {
		refresher := newRefresher(nil, nil, time.Hour, 24*time.Hour, 3, "")
		refresher.lastDiscoverySuccess.Store(time.Now().Add(-discoveryStaleWarn - time.Minute).UnixNano())
		probe := booted(refresher)

		evaluated := probe.Evaluate(context.Background())
		Expect(evaluated.Checks[discoveryCheckName].Status).To(Equal(health.StatusWarn))
		Expect(evaluated.Checks[discoveryCheckName].Error).To(ContainSubstring(ErrDiscoverySilent.Error()))
	})

	It("passes on a connected stream, however old the last contact stamp", func() {
		refresher := newRefresher(nil, nil, time.Hour, 24*time.Hour, 3, "")
		refresher.lastDiscoverySuccess.Store(time.Now().Add(-24 * time.Hour).UnixNano())
		refresher.discoveryState.Store(discoveryLive)
		probe := booted(refresher)

		Expect(probe.Evaluate(context.Background()).Checks[discoveryCheckName].Status).
			To(Equal(health.StatusPass))

		refresher.discoveryState.Store(discoveryIdle)

		Expect(probe.Evaluate(context.Background()).Checks[discoveryCheckName].Status).
			To(Equal(health.StatusWarn))
	})

	It("passes, with no daemon, when discovery is disabled", func() {
		refresher := newRefresher(nil, nil, time.Hour, 24*time.Hour, 3, "")
		refresher.DisableDiscovery()
		probe := booted(refresher)

		Expect(probe.Evaluate(context.Background()).Checks[discoveryCheckName].Status).
			To(Equal(health.StatusPass))
	})
})

var _ = Describe("health probe evidence gate", func() {
	It("fails readiness until evidence is published, then passes", func() {
		hub := events.NewHub()

		refresher := newRefresher(hub, nil, time.Hour, 24*time.Hour, 3, "")

		probe, err := buildHealthProbe(context.Background(), "test", hub, refresher, nil, "")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(probe.Shutdown)

		before := probe.Evaluate(context.Background())
		Expect(before.Checks).To(HaveKey(evidenceCheckName))
		Expect(before.Checks[evidenceCheckName].Status).To(Equal(health.StatusFail))
		Expect(before.Checks[evidenceCheckName].Error).To(ContainSubstring(ErrEvidenceNotPublished.Error()))

		_, err = hub.PublishSnapshot(nil)
		Expect(err).NotTo(HaveOccurred())

		after := probe.Evaluate(context.Background())
		Expect(after.Checks[evidenceCheckName].Status).To(Equal(health.StatusPass))
	})
})

var _ = Describe("health probe refresh-freshness check", func() {
	It("warns — never fails — while no refresh has succeeded, then passes on a fresh success", func() {
		refresher := newRefresher(nil, nil, time.Hour, 24*time.Hour, 3, "")
		probe := booted(refresher)

		before := probe.Evaluate(context.Background())
		Expect(before.Checks[refreshFreshnessCheckName].Status).To(Equal(health.StatusWarn))
		Expect(before.Checks[refreshFreshnessCheckName].Error).
			To(ContainSubstring(ErrRefreshNeverSucceeded.Error()))
		Expect(before.Status).To(Equal(health.StatusWarn),
			"freshness is non-critical: it must degrade readiness to warn, not fail it")

		refresher.lastRefreshSuccess.Store(time.Now().UnixNano())
		refresher.markDiscoverySuccess()

		after := probe.Evaluate(context.Background())
		Expect(after.Checks[refreshFreshnessCheckName].Status).To(Equal(health.StatusPass))
		Expect(after.Status).To(Equal(health.StatusPass))
	})

	It("warns with the stale age once the last success exceeds the threshold", func() {
		refresher := newRefresher(nil, nil, time.Hour, 24*time.Hour, 3, "")
		refresher.lastRefreshSuccess.Store(time.Now().Add(-refreshFreshnessWarn - time.Minute).UnixNano())
		probe := booted(refresher)

		evaluated := probe.Evaluate(context.Background())
		Expect(evaluated.Checks[refreshFreshnessCheckName].Status).To(Equal(health.StatusWarn))
		Expect(evaluated.Checks[refreshFreshnessCheckName].Error).
			To(ContainSubstring("last successful refresh"))
		Expect(evaluated.Status).To(Equal(health.StatusWarn))
	})
})

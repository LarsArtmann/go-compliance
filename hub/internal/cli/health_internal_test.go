package cli

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-health"

	"github.com/larsartmann/go-compliance/hub/internal/events"
)

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
})

var _ = Describe("health probe evidence gate", func() {
	It("fails readiness until evidence is published, then passes", func() {
		hub := events.NewHub()

		refresher := newRefresher(hub, nil, time.Hour, 24*time.Hour, 3, "")

		probe, err := buildHealthProbe(context.Background(), "test", hub, refresher, "")
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
	// An empty published fleet satisfies the evidence gate (verdict g2);
	// the freshness specs below then control lastRefreshSuccess directly.
	booted := func(refresher *refresher) *health.Probe {
		hub := events.NewHub()
		_, err := hub.PublishSnapshot(nil)
		Expect(err).NotTo(HaveOccurred())

		probe, err := buildHealthProbe(context.Background(), "test", hub, refresher, "")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(probe.Shutdown)

		return probe
	}

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

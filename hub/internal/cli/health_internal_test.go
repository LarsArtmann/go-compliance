package cli

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-health"

	"github.com/larsartmann/go-compliance/hub/internal/events"
)

var _ = Describe("health probe evidence gate", func() {
	It("fails readiness until evidence is published, then passes", func() {
		hub := events.NewHub()

		probe, err := buildHealthProbe(context.Background(), "test", hub, "")
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

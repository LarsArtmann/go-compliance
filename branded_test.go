package compliance_test

import (
	"encoding/json/v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-compliance"
)

var _ = Describe("Branded types", func() {
	It("brand repository names", func() {
		name := compliance.NewRepoName("go-finding")
		Expect(name.Get()).To(Equal("go-finding"))
		Expect(name.IsZero()).To(BeFalse())
	})

	It("brand source names independently", func() {
		name := compliance.NewSourceName("buildflow")
		Expect(name.Get()).To(Equal("buildflow"))
		Expect(name.Equal(compliance.NewSourceName("buildflow"))).To(BeTrue())
	})

	It("treat the zero value as unset", func() {
		var name compliance.RepoName
		Expect(name.IsZero()).To(BeTrue())
	})

	It("serialize as plain JSON strings, null when unset", func() {
		encoded, err := json.Marshal(compliance.NewRepoName("go-finding"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(encoded)).To(Equal(`"go-finding"`))

		var zero compliance.RepoName
		encoded, err = json.Marshal(zero)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(encoded)).To(Equal(`null`))
	})

	It("round-trip through JSON", func() {
		var decoded compliance.RepoName
		Expect(json.Unmarshal([]byte(`"overview"`), &decoded)).To(Succeed())
		Expect(decoded.Get()).To(Equal("overview"))
	})
})

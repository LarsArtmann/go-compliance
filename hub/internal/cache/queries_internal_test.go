package cache

import (
	"errors"

	errorfamily "github.com/larsartmann/go-error-family"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-compliance"
)

var _ = Describe("wrapIterateErr", func() {
	DescribeTable("rows error propagation",
		func(cause error, what string, expectWrap bool) {
			wrapped := wrapIterateErr(cause, what)

			if !expectWrap {
				Expect(wrapped).ToNot(HaveOccurred())

				return
			}

			Expect(errors.Is(wrapped, compliance.ErrIO)).To(BeTrue())
			Expect(errorfamily.Classify(wrapped)).To(Equal(errorfamily.Transient))
			Expect(wrapped.Error()).To(ContainSubstring(what))
			Expect(errors.Unwrap(wrapped)).To(Equal(cause), "the driver cause must stay inspectable")
		},
		Entry("nil stays nil (the healthy-iterate contract)", nil, "iterate window health", false),
		Entry("a mid-iterate failure wraps as IO with context",
			errors.New("driver: interrupted mid-iterate"), "iterate latest records", true),
		Entry("an empty what still wraps", errors.New("boom"), "", true),
	)
})

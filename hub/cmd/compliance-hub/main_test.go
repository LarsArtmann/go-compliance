package main

import (
	"context"
	"errors"

	errorfamily "github.com/larsartmann/go-error-family"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-compliance"
)

// The binary's exit contract: main uses errorfamily.ExitCode (BSD sysexits),
// so compliance errors classify through their family integration instead of
// collapsing to generic 1. Systemd Restart= semantics and operator muscle
// memory depend on these codes staying stable.
var _ = Describe("exit-code contract", func() {
	DescribeTable("compliance errors map to fleet-standard exit codes",
		func(err error, expectedExit int) {
			Expect(errorfamily.ExitCode(err)).To(Equal(expectedExit))
		},
		Entry("nil error succeeds", nil, 0),
		Entry(
			"transient IO (cache unreachable) exits 75 EX_TEMPFAIL — systemd may retry",
			compliance.NewIOError(errors.New("db locked"), "query cache"), 75,
		),
		Entry(
			"validation rejection exits 1 — do not restart",
			compliance.NewValidationError("bad window %q", "nope"), 1,
		),
		Entry(
			"parse rejection exits 1",
			compliance.NewParseError(errors.New("bad json"), "decode fixture"), 1,
		),
		Entry(
			"internal infrastructure exits 69 EX_UNAVAILABLE",
			compliance.NewInternalError(errors.New("boom"), "wiring"), 69,
		),
	)

	It("treats bare invocation as help: no error, exit 0", func() {
		Expect(run(context.Background(), []string{})).NotTo(HaveOccurred())
	})

	It("surfaces run() errors through the family classifier", func() {
		err := run(context.Background(), []string{"report", "--db", "/nonexistent/x.db"})
		Expect(err).To(HaveOccurred())
		Expect(errorfamily.ExitCode(err)).To(Equal(75)) // IO family: cache file unreachable
	})
})

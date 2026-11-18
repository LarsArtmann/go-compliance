package compliance_test

import (
	"errors"
	"io"

	errorfamily "github.com/larsartmann/go-error-family"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/larsartmann/go-compliance"
)

var _ = Describe("Error", func() {
	DescribeTable("category sentinels",
		func(err error, sentinel error) {
			Expect(err).To(MatchError(sentinel))
		},
		Entry("validation", compliance.NewValidationError("bad input"), compliance.ErrValidation),
		Entry("parse", compliance.NewParseError(io.ErrUnexpectedEOF, "bad json"), compliance.ErrParse),
		Entry("io", compliance.NewIOError(io.EOF, "read failed"), compliance.ErrIO),
		Entry("unsupported schema", compliance.NewUnsupportedSchemaError(2, 1), compliance.ErrUnsupportedSchema),
		Entry("internal", compliance.NewInternalError(io.EOF, "boom"), compliance.ErrInternal),
	)

	It("keeps the cause reachable through errors.Is", func() {
		err := compliance.NewParseError(io.ErrUnexpectedEOF, "decode evidence")
		Expect(errors.Is(err, io.ErrUnexpectedEOF)).To(BeTrue())
	})

	It("is matchable as *compliance.Error", func() {
		err := compliance.NewValidationError("no runs")
		typed, ok := errors.AsType[*compliance.Error](err)
		Expect(ok).To(BeTrue())
		Expect(typed.Category).To(Equal(compliance.ErrCategoryValidation))
	})

	DescribeTable("error families",
		func(err error, family errorfamily.Family) {
			typed, ok := errors.AsType[*compliance.Error](err)
			Expect(ok).To(BeTrue())
			Expect(typed.ErrorFamily()).To(Equal(family))
			Expect(errorfamily.Classify(err)).To(Equal(family))
		},
		Entry("validation is rejection", compliance.NewValidationError("x"), errorfamily.Rejection),
		Entry("parse is rejection", compliance.NewParseError(io.EOF, "x"), errorfamily.Rejection),
		Entry("unsupported schema is rejection", compliance.NewUnsupportedSchemaError(3, 1), errorfamily.Rejection),
		Entry("io is transient", compliance.NewIOError(io.EOF, "x"), errorfamily.Transient),
		Entry("internal is infrastructure", compliance.NewInternalError(io.EOF, "x"), errorfamily.Infrastructure),
	)

	DescribeTable("error codes",
		func(err error, code string) {
			typed, ok := errors.AsType[*compliance.Error](err)
			Expect(ok).To(BeTrue())
			Expect(typed.ErrorCode()).To(Equal(code))
		},
		Entry("validation", compliance.NewValidationError("x"), "compliance.validation"),
		Entry("io", compliance.NewIOError(nil, "x"), "compliance.io"),
		Entry("parse", compliance.NewParseError(nil, "x"), "compliance.parse"),
		Entry("unsupported", compliance.NewUnsupportedSchemaError(2, 1), "compliance.unsupported"),
		Entry("internal", compliance.NewInternalError(nil, "x"), "compliance.internal"),
	)

	It("render category, message, and cause", func() {
		err := compliance.NewIOError(io.EOF, "read evidence file")
		Expect(err.Error()).To(Equal("[io] read evidence file: EOF"))
		Expect(compliance.NewValidationError("nope").Error()).To(Equal("[validation] nope"))
	})

	It("name the versions in unsupported-schema errors", func() {
		err := compliance.NewUnsupportedSchemaError(2, 1)
		Expect(err.Error()).To(Equal("[unsupported] unsupported compliance schema version 2 (want 1)"))
	})

	It("validate categories", func() {
		Expect(compliance.ErrCategoryValidation.IsValid()).To(BeTrue())
		Expect(compliance.ErrCategoryIO.IsValid()).To(BeTrue())
		Expect(compliance.ErrCategoryParse.IsValid()).To(BeTrue())
		Expect(compliance.ErrCategoryUnsupported.IsValid()).To(BeTrue())
		Expect(compliance.ErrCategoryInternal.IsValid()).To(BeTrue())
		Expect(compliance.ErrorCategory("bogus").IsValid()).To(BeFalse())
	})
})

var _ = Describe("ParseState error shape", func() {
	It("returns a validation error for unknown states", func() {
		_, err := compliance.ParseState("healthy")
		Expect(err).To(MatchError(compliance.ErrValidation))
		Expect(err.Error()).To(ContainSubstring("healthy"))
	})
})

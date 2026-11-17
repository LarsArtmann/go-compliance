package compliance

import (
	"errors"
	"fmt"

	errorfamily "github.com/larsartmann/go-error-family"
)

// Sentinel errors for use with errors.Is.
var (
	ErrValidation        = errors.New("compliance: validation error")
	ErrIO                = errors.New("compliance: I/O error")
	ErrParse             = errors.New("compliance: parse error")
	ErrUnsupportedSchema = errors.New("compliance: unsupported schema version")
	ErrInternal          = errors.New("compliance: internal error")
)

// ErrorCategory categorizes errors for programmatic handling.
type ErrorCategory string

const (
	// ErrCategoryValidation indicates invalid input or contract misuse.
	ErrCategoryValidation ErrorCategory = "validation"
	// ErrCategoryIO indicates file system or network errors (retryable).
	ErrCategoryIO ErrorCategory = "io"
	// ErrCategoryParse indicates decoding malformed payloads.
	ErrCategoryParse ErrorCategory = "parse"
	// ErrCategoryUnsupported indicates an unsupported contract version.
	ErrCategoryUnsupported ErrorCategory = "unsupported"
	// ErrCategoryInternal indicates internal logic errors.
	ErrCategoryInternal ErrorCategory = "internal"
)

// IsValid reports whether the category is a known lowercase-hyphenated
// category.
func (c ErrorCategory) IsValid() bool {
	switch c {
	case ErrCategoryValidation, ErrCategoryIO, ErrCategoryParse, ErrCategoryUnsupported, ErrCategoryInternal:
		return true
	default:
		return false
	}
}

// Error provides structured error information with context. Every error the
// package returns is a *Error, so callers can errors.Is against the
// sentinels and errorfamily.Classify works on compliance errors.
type Error struct {
	Category ErrorCategory // Behavioral category
	Message  string        // Human-readable message
	Cause    error         // Underlying cause (may be nil)
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Category, e.Message, e.Cause)
	}

	return fmt.Sprintf("[%s] %s", e.Category, e.Message)
}

// Unwrap returns the underlying cause for error inspection.
func (e *Error) Unwrap() error { return e.Cause }

// Is supports errors.Is by matching the sentinel of the error's category.
func (e *Error) Is(target error) bool {
	switch e.Category {
	case ErrCategoryValidation:
		return target == ErrValidation
	case ErrCategoryIO:
		return target == ErrIO
	case ErrCategoryParse:
		return target == ErrParse
	case ErrCategoryUnsupported:
		return target == ErrUnsupportedSchema
	case ErrCategoryInternal:
		return target == ErrInternal
	default:
		return false
	}
}

// ErrorCode implements errorfamily.Coded.
func (e *Error) ErrorCode() string { return "compliance." + string(e.Category) }

// ErrorFamily implements errorfamily.Classified, mapping categories to
// go-error-family families for integration with Classify().
func (e *Error) ErrorFamily() errorfamily.Family {
	switch e.Category {
	case ErrCategoryValidation, ErrCategoryParse, ErrCategoryUnsupported:
		return errorfamily.Rejection
	case ErrCategoryIO:
		return errorfamily.Transient
	case ErrCategoryInternal:
		return errorfamily.Infrastructure
	default:
		return errorfamily.Transient
	}
}

// NewValidationError returns a validation error (formatted message, no cause).
func NewValidationError(format string, args ...any) *Error {
	return &Error{Category: ErrCategoryValidation, Message: fmt.Sprintf(format, args...)}
}

// NewParseError wraps a decoding failure as a parse error.
func NewParseError(cause error, format string, args ...any) *Error {
	return &Error{Category: ErrCategoryParse, Message: fmt.Sprintf(format, args...), Cause: cause}
}

// NewIOError wraps a file system or network failure as an I/O error.
func NewIOError(cause error, format string, args ...any) *Error {
	return &Error{Category: ErrCategoryIO, Message: fmt.Sprintf(format, args...), Cause: cause}
}

// NewUnsupportedSchemaError reports a decoded schema_version the consumer
// does not support: a changed contract needs a conscious bump on both
// sides, never a silent misparse (mirrors the overview consumer).
func NewUnsupportedSchemaError(got, want int) *Error {
	return &Error{
		Category: ErrCategoryUnsupported,
		Message:  fmt.Sprintf("unsupported compliance schema version %d (want %d)", got, want),
	}
}

// NewInternalError wraps an internal failure.
func NewInternalError(cause error, format string, args ...any) *Error {
	return &Error{Category: ErrCategoryInternal, Message: fmt.Sprintf(format, args...), Cause: cause}
}

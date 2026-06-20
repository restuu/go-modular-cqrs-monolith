package errcode

import "errors"

// AppError is a typed error that carries a machine-readable Code and a human-readable Message.
// Domain packages declare sentinel vars of this type; the presentation layer maps Code to HTTP
// status via response.FromError without importing any domain package.
type AppError struct {
	Code    Code
	Message string
	cause   error
}

// New creates a sentinel AppError with no cause. Use for package-level var declarations.
func New(code Code, msg string) *AppError {
	return &AppError{Code: code, Message: msg}
}

// Wrap creates an AppError that preserves an underlying cause for Unwrap traversal.
func Wrap(err error, code Code, msg string) *AppError {
	return &AppError{Code: code, Message: msg, cause: err}
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.cause }

// Is matches by Code so that errors.Is(wrappedErr, sentinelErr) returns true
// even when the error was re-wrapped up the call stack.
func (e *AppError) Is(target error) bool {
	var t *AppError
	if !errors.As(target, &t) {
		return false
	}
	return e.Code == t.Code
}

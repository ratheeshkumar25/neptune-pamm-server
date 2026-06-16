// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/internal/model/errors.go
// Role: Domain models — project-wide error type
// Description: A single AppError type used across every layer. It carries a machine code,
// a category Type, a human Message, an HTTP status, and an optional wrapped cause. The
// JSON shape matches the TFB-parity envelope { "Message", "Type", "Code" } so handlers can
// serialize a domain error straight to the wire (see 05-API-PARITY §3.2).

package model

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode is the machine-readable code in the error envelope. Branch on this, not on text.
type ErrorCode string

const (
	ErrCodeData         ErrorCode = "ERR_DATA"         // validation / business-rule (400)
	ErrCodeUnauthorized ErrorCode = "ERR_UNAUTHORIZED" // missing/invalid auth (401)
	ErrCodeForbidden    ErrorCode = "ERR_FORBIDDEN"    // authenticated but not allowed (403)
	ErrCodeNotFound     ErrorCode = "ERR_NOT_FOUND"    // resource missing (404)
	ErrCodeConflict     ErrorCode = "ERR_CONFLICT"     // duplicate / state conflict (409)
	ErrCodeRateLimited  ErrorCode = "ERR_RATE_LIMITED" // too many requests (429)
	ErrCodeInternal     ErrorCode = "ERR_INTERNAL"     // unexpected server error (500)
	ErrCodeUnavailable  ErrorCode = "ERR_UNAVAILABLE"  // dependency down (503)
)

// AppError is the canonical error carried across services, handlers and repositories.
// The exported fields serialize to the TFB envelope; status and cause stay internal.
type AppError struct {
	Code    ErrorCode `json:"Code"`
	Type    string    `json:"Type"`
	Message string    `json:"Message"`

	status int   // HTTP status — not serialized
	cause  error // wrapped underlying error — not serialized
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap exposes the wrapped cause for errors.Is / errors.As traversal.
func (e *AppError) Unwrap() error { return e.cause }

// Is matches by Code, so errors.Is(err, model.ErrNotFound) works regardless of message/cause.
func (e *AppError) Is(target error) bool {
	var t *AppError
	if errors.As(target, &t) {
		return e.Code == t.Code
	}
	return false
}

// HTTPStatus returns the HTTP status the handler layer should respond with.
func (e *AppError) HTTPStatus() int {
	if e.status != 0 {
		return e.status
	}
	return http.StatusInternalServerError
}

// WithMessage returns a copy with a custom message (preserving code/type/status/cause).
func (e *AppError) WithMessage(msg string) *AppError {
	c := *e
	c.Message = msg
	return &c
}

// Wrap returns a copy that wraps cause, so the original error chain is preserved for logs.
func (e *AppError) Wrap(cause error) *AppError {
	c := *e
	c.cause = cause
	return &c
}

// newBase builds a reusable template error for a given code/status/type.
func newBase(code ErrorCode, status int, typ, msg string) *AppError {
	return &AppError{Code: code, Type: typ, Message: msg, status: status}
}

// Sentinel templates — compare with errors.Is, or use the New* helpers to add context.
var (
	ErrValidation   = newBase(ErrCodeData, http.StatusBadRequest, "Validation", "invalid request data")
	ErrUnauthorized = newBase(ErrCodeUnauthorized, http.StatusUnauthorized, "Unauthorized", "authentication required")
	ErrForbidden    = newBase(ErrCodeForbidden, http.StatusForbidden, "Forbidden", "operation not permitted")
	ErrNotFound     = newBase(ErrCodeNotFound, http.StatusNotFound, "NotFound", "resource not found")
	ErrConflict     = newBase(ErrCodeConflict, http.StatusConflict, "Conflict", "resource conflict")
	ErrRateLimited  = newBase(ErrCodeRateLimited, http.StatusTooManyRequests, "RateLimited", "rate limit exceeded")
	ErrInternal     = newBase(ErrCodeInternal, http.StatusInternalServerError, "Internal", "internal server error")
	ErrUnavailable  = newBase(ErrCodeUnavailable, http.StatusServiceUnavailable, "Unavailable", "service unavailable")
)

// --- constructors (preferred over building AppError literals) --------------------

// NewValidation reports invalid input or a violated business rule (400).
func NewValidation(format string, args ...any) *AppError {
	return ErrValidation.WithMessage(fmt.Sprintf(format, args...))
}

// NewUnauthorized reports missing or invalid authentication (401).
func NewUnauthorized(format string, args ...any) *AppError {
	return ErrUnauthorized.WithMessage(fmt.Sprintf(format, args...))
}

// NewForbidden reports an authenticated principal lacking permission (403).
func NewForbidden(format string, args ...any) *AppError {
	return ErrForbidden.WithMessage(fmt.Sprintf(format, args...))
}

// NewNotFound reports a missing entity, e.g. NewNotFound("investor", id) (404).
func NewNotFound(entity string, id any) *AppError {
	return ErrNotFound.WithMessage(fmt.Sprintf("%s not found: %v", entity, id))
}

// NewConflict reports a duplicate or state conflict (409).
func NewConflict(format string, args ...any) *AppError {
	return ErrConflict.WithMessage(fmt.Sprintf(format, args...))
}

// NewInternal wraps an unexpected error as a 500, keeping the cause for logs.
func NewInternal(cause error) *AppError {
	return ErrInternal.Wrap(cause)
}

// --- inspection helpers ----------------------------------------------------------

// AsAppError extracts an *AppError from anywhere in the chain.
func AsAppError(err error) (*AppError, bool) {
	var e *AppError
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

// FromError returns err as an *AppError, converting unknown errors into a wrapped
// internal error. Handlers can call this to always have an envelope + HTTP status.
func FromError(err error) *AppError {
	if err == nil {
		return nil
	}
	if e, ok := AsAppError(err); ok {
		return e
	}
	return NewInternal(err)
}

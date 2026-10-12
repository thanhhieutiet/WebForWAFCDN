package errs

import (
	"errors"
	"fmt"
)

// ErrorKind identifies the category of domain error.
type ErrorKind int

const (
	KindValidation ErrorKind = iota + 1
	KindUnauthorized
	KindForbidden
	KindNotFound
	KindConflict
	KindBusinessRule
)

// FieldViolation describes a validation error for a specific field.
type FieldViolation struct {
	Field string
	Issue string
}

// DomainError represents a structured error originating from domain logic.
type DomainError struct {
	Kind    ErrorKind
	Code    string
	Message string
	Details []FieldViolation
	Err     error
}

func (e *DomainError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("[%s] %s", e.Code, e.Message)
	}
	return e.Message
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// Constructors
func NewValidation(message string, details ...FieldViolation) *DomainError {
	return &DomainError{
		Kind:    KindValidation,
		Code:    "validation_error",
		Message: message,
		Details: details,
	}
}

func NewUnauthorized(message string) *DomainError {
	if message == "" {
		message = "authentication required"
	}
	return &DomainError{
		Kind:    KindUnauthorized,
		Code:    "unauthorized",
		Message: message,
	}
}

func NewForbidden(message string) *DomainError {
	if message == "" {
		message = "access denied"
	}
	return &DomainError{
		Kind:    KindForbidden,
		Code:    "forbidden",
		Message: message,
	}
}

func NewNotFound(resource, id string) *DomainError {
	msg := fmt.Sprintf("%s not found", resource)
	if id != "" {
		msg = fmt.Sprintf("%s %q not found", resource, id)
	}
	return &DomainError{
		Kind:    KindNotFound,
		Code:    "not_found",
		Message: msg,
	}
}

func NewConflict(message string) *DomainError {
	return &DomainError{
		Kind:    KindConflict,
		Code:    "conflict",
		Message: message,
	}
}

func NewBusinessRule(code, message string) *DomainError {
	return &DomainError{
		Kind:    KindBusinessRule,
		Code:    code,
		Message: message,
	}
}

// Type assertion helpers
func AsDomainError(err error) (*DomainError, bool) {
	var de *DomainError
	if errors.As(err, &de) {
		return de, true
	}
	return nil, false
}

func IsNotFound(err error) bool {
	if de, ok := AsDomainError(err); ok {
		return de.Kind == KindNotFound
	}
	return false
}

func IsConflict(err error) bool {
	if de, ok := AsDomainError(err); ok {
		return de.Kind == KindConflict
	}
	return false
}

func IsValidation(err error) bool {
	if de, ok := AsDomainError(err); ok {
		return de.Kind == KindValidation
	}
	return false
}

func IsBusinessRule(err error) bool {
	if de, ok := AsDomainError(err); ok {
		return de.Kind == KindBusinessRule
	}
	return false
}

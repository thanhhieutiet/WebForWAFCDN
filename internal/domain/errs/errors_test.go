package errs

import (
	"errors"
	"testing"
)

func TestDomainErrors(t *testing.T) {
	errVal := NewValidation("invalid payload", FieldViolation{Field: "email", Issue: "required"})
	if !IsValidation(errVal) {
		t.Errorf("expected IsValidation to be true")
	}
	if errVal.Error() != "[validation_error] invalid payload" {
		t.Errorf("unexpected error string: %s", errVal.Error())
	}

	errNF := NewNotFound("product", "prod-1")
	if !IsNotFound(errNF) {
		t.Errorf("expected IsNotFound to be true")
	}

	errConflict := NewConflict("email already registered")
	if !IsConflict(errConflict) {
		t.Errorf("expected IsConflict to be true")
	}

	errBR := NewBusinessRule("coupon_expired", "coupon has expired")
	if !IsBusinessRule(errBR) {
		t.Errorf("expected IsBusinessRule to be true")
	}

	stdErr := errors.New("some error")
	if IsNotFound(stdErr) {
		t.Errorf("expected IsNotFound for standard error to be false")
	}

	errAuth := NewUnauthorized("")
	if errAuth.Error() != "[unauthorized] authentication required" {
		t.Errorf("unexpected unauthorized message: %s", errAuth.Error())
	}

	errForbid := NewForbidden("")
	if errForbid.Error() != "[forbidden] access denied" {
		t.Errorf("unexpected forbidden message: %s", errForbid.Error())
	}

	wrapped := &DomainError{Kind: KindNotFound, Err: stdErr}
	if wrapped.Unwrap() != stdErr {
		t.Errorf("expected unwrapped error to match")
	}
}

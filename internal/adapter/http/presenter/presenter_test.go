package presenter

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"labshop/internal/domain/errs"
)

func TestPresenterErrorMapping(t *testing.T) {
	tests := []struct {
		err        error
		wantStatus int
		wantCode   string
	}{
		{errs.NewValidation("invalid", errs.FieldViolation{Field: "email", Issue: "required"}), http.StatusBadRequest, "validation_error"},
		{errs.NewUnauthorized("unauthorized"), http.StatusUnauthorized, "unauthorized"},
		{errs.NewForbidden("forbidden"), http.StatusForbidden, "forbidden"},
		{errs.NewNotFound("product", "p-1"), http.StatusNotFound, "not_found"},
		{errs.NewConflict("already exists"), http.StatusConflict, "conflict"},
		{errs.NewBusinessRule("idempotency_conflict", "reused key"), http.StatusUnprocessableEntity, "idempotency_conflict"},
	}

	for _, tt := range tests {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)

		Error(rec, req, tt.err)

		if rec.Code != tt.wantStatus {
			t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
		}

		var resp ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Error.Code != tt.wantCode {
			t.Errorf("expected error code %q, got %q", tt.wantCode, resp.Error.Code)
		}
	}
}

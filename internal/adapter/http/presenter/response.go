package presenter

import (
	"encoding/json"
	"net/http"

	"labshop/internal/domain/errs"
)

type FieldIssue struct {
	Field string `json:"field"`
	Issue string `json:"issue"`
}

type ErrorBody struct {
	Code      string       `json:"code"`
	Message   string       `json:"message"`
	Details   []FieldIssue `json:"details"`
	RequestID string       `json:"request_id"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// JSON writes a JSON payload with standard Content-Type.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// Error maps errors (including DomainError) to standard HTTP error envelopes.
func Error(w http.ResponseWriter, r *http.Request, err error) {
	reqID := w.Header().Get("X-Request-Id")
	if reqID == "" && r != nil {
		reqID = r.Header.Get("X-Request-Id")
	}

	de, ok := errs.AsDomainError(err)
	if !ok {
		// Generic internal server error
		JSON(w, http.StatusInternalServerError, ErrorResponse{
			Error: ErrorBody{
				Code:      "internal_server_error",
				Message:   "an unexpected internal error occurred",
				Details:   []FieldIssue{},
				RequestID: reqID,
			},
		})
		return
	}

	details := make([]FieldIssue, 0, len(de.Details))
	for _, d := range de.Details {
		details = append(details, FieldIssue{
			Field: d.Field,
			Issue: d.Issue,
		})
	}

	statusCode := http.StatusInternalServerError
	switch de.Kind {
	case errs.KindValidation:
		statusCode = http.StatusBadRequest
	case errs.KindUnauthorized:
		statusCode = http.StatusUnauthorized
	case errs.KindForbidden:
		statusCode = http.StatusForbidden
	case errs.KindNotFound:
		statusCode = http.StatusNotFound
	case errs.KindConflict:
		statusCode = http.StatusConflict
	case errs.KindBusinessRule:
		statusCode = http.StatusUnprocessableEntity
	}

	JSON(w, statusCode, ErrorResponse{
		Error: ErrorBody{
			Code:      de.Code,
			Message:   de.Message,
			Details:   details,
			RequestID: reqID,
		},
	})
}

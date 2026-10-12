package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"labshop/internal/domain/errs"
)

// decodeJSON reads and parses JSON with DisallowUnknownFields and a 1MB limit.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var syntaxErr *json.SyntaxError
		var unmarshalTypeErr *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxErr):
			return errs.NewValidation(fmt.Sprintf("malformed JSON at byte %d", syntaxErr.Offset))
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errs.NewValidation("unexpected end of JSON input")
		case errors.As(err, &unmarshalTypeErr):
			return errs.NewValidation(fmt.Sprintf("invalid type for field %s", unmarshalTypeErr.Field), errs.FieldViolation{
				Field: unmarshalTypeErr.Field,
				Issue: "incorrect type",
			})
		case errors.Is(err, io.EOF):
			return errs.NewValidation("request body cannot be empty")
		default:
			return errs.NewValidation(err.Error())
		}
	}

	return nil
}

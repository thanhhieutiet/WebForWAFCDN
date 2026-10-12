package review

import (
	"strings"
	"time"
	"unicode/utf8"

	"labshop/internal/domain/errs"
)

// Review represents a customer rating and feedback on a product.
type Review struct {
	ID        string
	UserID    string
	ProductID string
	UserName  string
	Rating    int
	Comment   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate checks review invariants.
func (r *Review) Validate() error {
	var violations []errs.FieldViolation

	if strings.TrimSpace(r.UserID) == "" {
		violations = append(violations, errs.FieldViolation{Field: "user_id", Issue: "user_id is required"})
	}
	if strings.TrimSpace(r.ProductID) == "" {
		violations = append(violations, errs.FieldViolation{Field: "product_id", Issue: "product_id is required"})
	}
	if r.Rating < 1 || r.Rating > 5 {
		violations = append(violations, errs.FieldViolation{Field: "rating", Issue: "rating must be between 1 and 5"})
	}

	commentLen := utf8.RuneCountInString(strings.TrimSpace(r.Comment))
	if commentLen < 1 || commentLen > 2000 {
		violations = append(violations, errs.FieldViolation{Field: "comment", Issue: "comment must be between 1 and 2000 characters"})
	}

	if len(violations) > 0 {
		return errs.NewValidation("review validation failed", violations...)
	}
	return nil
}

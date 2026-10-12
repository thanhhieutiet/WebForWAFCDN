package review

import (
	"strings"
	"testing"
	"time"
)

func TestReviewValidation(t *testing.T) {
	r := &Review{
		ID:        "rev-1",
		UserID:    "u-1",
		ProductID: "p-1",
		Rating:    5,
		Comment:   "Great product! Highly recommended.",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := r.Validate(); err != nil {
		t.Fatalf("expected review to be valid: %v", err)
	}

	r.Rating = 0
	if err := r.Validate(); err == nil {
		t.Fatalf("expected error for rating 0")
	}

	r.Rating = 6
	if err := r.Validate(); err == nil {
		t.Fatalf("expected error for rating 6")
	}

	r.Rating = 4
	r.Comment = ""
	if err := r.Validate(); err == nil {
		t.Fatalf("expected error for empty comment")
	}

	r.Comment = strings.Repeat("A", 2001)
	if err := r.Validate(); err == nil {
		t.Fatalf("expected error for comment > 2000 chars")
	}
}

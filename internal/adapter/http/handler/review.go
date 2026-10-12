package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"labshop/internal/adapter/http/middleware"
	"labshop/internal/adapter/http/presenter"
	"labshop/internal/domain/review"
	reviewuc "labshop/internal/usecase/review"
)

type ReviewHandler struct {
	reviewUC *reviewuc.Usecase
}

func NewReviewHandler(reviewUC *reviewuc.Usecase) *ReviewHandler {
	return &ReviewHandler{reviewUC: reviewUC}
}

type CreateReviewRequest struct {
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
}

type ReviewDTO struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ProductID string    `json:"product_id"`
	UserName  string    `json:"user_name"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toReviewDTO(r review.Review) ReviewDTO {
	return ReviewDTO{
		ID:        r.ID,
		UserID:    r.UserID,
		ProductID: r.ProductID,
		UserName:  r.UserName,
		Rating:    r.Rating,
		Comment:   r.Comment,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

type ListReviewsResponse struct {
	Items    []ReviewDTO `json:"items"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func (h *ReviewHandler) ListReviews(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	if page <= 0 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	reviews, total, err := h.reviewUC.ListReviews(r.Context(), slug, pageSize, offset)
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	dtos := make([]ReviewDTO, 0, len(reviews))
	for _, rev := range reviews {
		dtos = append(dtos, toReviewDTO(rev))
	}

	presenter.JSON(w, http.StatusOK, ListReviewsResponse{
		Items:    dtos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *ReviewHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	slug := chi.URLParam(r, "slug")

	var req CreateReviewRequest
	if err := decodeJSON(w, r, &req); err != nil {
		presenter.Error(w, r, err)
		return
	}

	rev, err := h.reviewUC.CreateReview(r.Context(), claims.UserID, claims.Email, reviewuc.CreateReviewInput{
		ProductSlug: slug,
		Rating:      req.Rating,
		Comment:     req.Comment,
	})
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.JSON(w, http.StatusCreated, toReviewDTO(*rev))
}

func (h *ReviewHandler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	reviewID := chi.URLParam(r, "id")

	err := h.reviewUC.DeleteReview(r.Context(), claims.UserID, reviewID, claims.Role == "admin")
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.JSON(w, http.StatusOK, map[string]string{"message": "review deleted successfully"})
}

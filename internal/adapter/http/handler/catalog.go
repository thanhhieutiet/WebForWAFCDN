package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"labshop/internal/adapter/http/presenter"
	"labshop/internal/domain/catalog"
	"labshop/internal/domain/money"
	cataloguc "labshop/internal/usecase/catalog"
	"labshop/internal/usecase/port"
)

type CatalogHandler struct {
	catalogUC *cataloguc.Usecase
}

func NewCatalogHandler(catalogUC *cataloguc.Usecase) *CatalogHandler {
	return &CatalogHandler{catalogUC: catalogUC}
}

type CategoryDTO struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type ProductDTO struct {
	ID          string    `json:"id"`
	CategoryID  string    `json:"category_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Price       int64     `json:"price"`
	PriceVND    string    `json:"price_vnd"`
	Stock       int       `json:"stock"`
	ImagePath   string    `json:"image_path"`
	Status      string    `json:"status"`
	RatingAvg   float64   `json:"rating_avg"`
	RatingCount int       `json:"rating_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toCategoryDTO(c catalog.Category) CategoryDTO {
	return CategoryDTO{
		ID:          c.ID,
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
		CreatedAt:   c.CreatedAt,
	}
}

func toProductDTO(p catalog.Product) ProductDTO {
	return ProductDTO{
		ID:          p.ID,
		CategoryID:  p.CategoryID,
		Name:        p.Name,
		Slug:        p.Slug,
		Description: p.Description,
		Price:       p.Price.Int64(),
		PriceVND:    p.Price.Format(),
		Stock:       p.Stock,
		ImagePath:   p.ImagePath,
		Status:      string(p.Status),
		RatingAvg:   p.RatingAvg,
		RatingCount: p.RatingCount,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

type ListProductsResponse struct {
	Items    []ProductDTO `json:"items"`
	Total    int          `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
}

func (h *CatalogHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.catalogUC.ListCategories(r.Context())
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	dtos := make([]CategoryDTO, 0, len(cats))
	for _, c := range cats {
		dtos = append(dtos, toCategoryDTO(c))
	}

	presenter.JSON(w, http.StatusOK, map[string]any{"categories": dtos})
}

func (h *CatalogHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	if page <= 0 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if pageSize <= 0 {
		pageSize = 20
	}

	filter := port.ProductFilter{
		Query:      q.Get("q"),
		CategoryID: q.Get("category"),
		Sort:       q.Get("sort"),
		Page:       page,
		PageSize:   pageSize,
	}

	if minStr := q.Get("min_price"); minStr != "" {
		if minVal, err := strconv.ParseInt(minStr, 10, 64); err == nil {
			v := money.VND(minVal)
			filter.MinPrice = &v
		}
	}

	if maxStr := q.Get("max_price"); maxStr != "" {
		if maxVal, err := strconv.ParseInt(maxStr, 10, 64); err == nil {
			v := money.VND(maxVal)
			filter.MaxPrice = &v
		}
	}

	products, total, err := h.catalogUC.ListProducts(r.Context(), filter)
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	dtos := make([]ProductDTO, 0, len(products))
	for _, p := range products {
		dtos = append(dtos, toProductDTO(p))
	}

	presenter.JSON(w, http.StatusOK, ListProductsResponse{
		Items:    dtos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *CatalogHandler) GetProductBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	prod, err := h.catalogUC.GetProductBySlug(r.Context(), slug)
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.JSON(w, http.StatusOK, toProductDTO(*prod))
}

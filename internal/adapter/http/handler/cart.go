package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"labshop/internal/adapter/http/middleware"
	"labshop/internal/adapter/http/presenter"
	cartuc "labshop/internal/usecase/cart"
)

type CartHandler struct {
	cartUC *cartuc.Usecase
}

func NewCartHandler(cartUC *cartuc.Usecase) *CartHandler {
	return &CartHandler{cartUC: cartUC}
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity"`
}

type ApplyCouponRequest struct {
	Code string `json:"code"`
}

type CartItemDTO struct {
	ProductID   string `json:"product_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	ImagePath   string `json:"image_path"`
	Price       int64  `json:"price"`
	PriceVND    string `json:"price_vnd"`
	Quantity    int    `json:"quantity"`
	Subtotal    int64  `json:"subtotal"`
	SubtotalVND string `json:"subtotal_vnd"`
	Stock       int    `json:"stock"`
}

type CartDTO struct {
	UserID         string        `json:"user_id"`
	Items          []CartItemDTO `json:"items"`
	Subtotal       int64         `json:"subtotal"`
	SubtotalVND    string        `json:"subtotal_vnd"`
	Discount       int64         `json:"discount"`
	DiscountVND    string        `json:"discount_vnd"`
	ShippingFee    int64         `json:"shipping_fee"`
	ShippingFeeVND string        `json:"shipping_fee_vnd"`
	Total          int64         `json:"total"`
	TotalVND       string        `json:"total_vnd"`
	CouponCode     string        `json:"coupon_code"`
	ItemCount      int           `json:"item_count"`
}

func toCartDTO(cv *cartuc.CartView) CartDTO {
	items := make([]CartItemDTO, 0, len(cv.Items))
	for _, it := range cv.Items {
		items = append(items, CartItemDTO{
			ProductID:   it.ProductID,
			Name:        it.Name,
			Slug:        it.Slug,
			ImagePath:   it.ImagePath,
			Price:       it.Price.Int64(),
			PriceVND:    it.Price.Format(),
			Quantity:    it.Quantity,
			Subtotal:    it.Subtotal.Int64(),
			SubtotalVND: it.Subtotal.Format(),
			Stock:       it.Stock,
		})
	}

	return CartDTO{
		UserID:         cv.UserID,
		Items:          items,
		Subtotal:       cv.Subtotal.Int64(),
		SubtotalVND:    cv.Subtotal.Format(),
		Discount:       cv.Discount.Int64(),
		DiscountVND:    cv.Discount.Format(),
		ShippingFee:    cv.ShippingFee.Int64(),
		ShippingFeeVND: cv.ShippingFee.Format(),
		Total:          cv.Total.Int64(),
		TotalVND:       cv.Total.Format(),
		CouponCode:     cv.CouponCode,
		ItemCount:      cv.ItemCount,
	}
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	cv, err := h.cartUC.GetCart(r.Context(), claims.UserID)
	if err != nil {
		presenter.Error(w, r, err)
		return
	}
	presenter.JSON(w, http.StatusOK, toCartDTO(cv))
}

func (h *CartHandler) UpdateItemQuantity(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	productID := chi.URLParam(r, "product_id")

	var req UpdateCartItemRequest
	if err := decodeJSON(w, r, &req); err != nil {
		presenter.Error(w, r, err)
		return
	}

	cv, err := h.cartUC.UpdateItemQuantity(r.Context(), claims.UserID, productID, req.Quantity)
	if err != nil {
		presenter.Error(w, r, err)
		return
	}
	presenter.JSON(w, http.StatusOK, toCartDTO(cv))
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	productID := chi.URLParam(r, "product_id")

	cv, err := h.cartUC.RemoveItem(r.Context(), claims.UserID, productID)
	if err != nil {
		presenter.Error(w, r, err)
		return
	}
	presenter.JSON(w, http.StatusOK, toCartDTO(cv))
}

func (h *CartHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if err := h.cartUC.ClearCart(r.Context(), claims.UserID); err != nil {
		presenter.Error(w, r, err)
		return
	}
	presenter.JSON(w, http.StatusOK, map[string]string{"message": "cart cleared"})
}

func (h *CartHandler) ApplyCoupon(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	var req ApplyCouponRequest
	if err := decodeJSON(w, r, &req); err != nil {
		presenter.Error(w, r, err)
		return
	}

	cv, err := h.cartUC.ApplyCoupon(r.Context(), claims.UserID, req.Code)
	if err != nil {
		presenter.Error(w, r, err)
		return
	}
	presenter.JSON(w, http.StatusOK, toCartDTO(cv))
}

func (h *CartHandler) RemoveCoupon(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	cv, err := h.cartUC.RemoveCoupon(r.Context(), claims.UserID)
	if err != nil {
		presenter.Error(w, r, err)
		return
	}
	presenter.JSON(w, http.StatusOK, toCartDTO(cv))
}

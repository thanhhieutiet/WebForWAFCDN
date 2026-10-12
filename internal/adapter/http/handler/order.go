package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"labshop/internal/adapter/http/middleware"
	"labshop/internal/adapter/http/presenter"
	"labshop/internal/domain/order"
	orderuc "labshop/internal/usecase/order"
)

type OrderHandler struct {
	orderUC *orderuc.Usecase
}

func NewOrderHandler(orderUC *orderuc.Usecase) *OrderHandler {
	return &OrderHandler{orderUC: orderUC}
}

type CreateOrderRequest struct {
	RecipientName   string `json:"recipient_name"`
	RecipientPhone  string `json:"recipient_phone"`
	ShippingAddress string `json:"shipping_address"`
	PaymentMethod   string `json:"payment_method"`
}

type OrderItemDTO struct {
	ID              string `json:"id"`
	ProductID       string `json:"product_id"`
	ProductName     string `json:"product_name"`
	ProductPrice    int64  `json:"product_price"`
	ProductPriceVND string `json:"product_price_vnd"`
	Quantity        int    `json:"quantity"`
	Subtotal        int64  `json:"subtotal"`
	SubtotalVND     string `json:"subtotal_vnd"`
}

type OrderDTO struct {
	ID              string         `json:"id"`
	UserID          string         `json:"user_id"`
	Items           []OrderItemDTO `json:"items"`
	Subtotal        int64          `json:"subtotal"`
	SubtotalVND     string         `json:"subtotal_vnd"`
	Discount        int64          `json:"discount"`
	DiscountVND     string         `json:"discount_vnd"`
	ShippingFee     int64          `json:"shipping_fee"`
	ShippingFeeVND  string         `json:"shipping_fee_vnd"`
	Total           int64          `json:"total"`
	TotalVND        string         `json:"total_vnd"`
	CouponCode      string         `json:"coupon_code"`
	Status          string         `json:"status"`
	PaymentMethod   string         `json:"payment_method"`
	PaymentRef      string         `json:"payment_ref,omitempty"`
	RecipientName   string         `json:"recipient_name"`
	RecipientPhone  string         `json:"recipient_phone"`
	ShippingAddress string         `json:"shipping_address"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

func toOrderDTO(o *order.Order) OrderDTO {
	items := make([]OrderItemDTO, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, OrderItemDTO{
			ID:              it.ID,
			ProductID:       it.ProductID,
			ProductName:     it.ProductName,
			ProductPrice:    it.ProductPrice.Int64(),
			ProductPriceVND: it.ProductPrice.Format(),
			Quantity:        it.Quantity,
			Subtotal:        it.Subtotal.Int64(),
			SubtotalVND:     it.Subtotal.Format(),
		})
	}

	return OrderDTO{
		ID:              o.ID,
		UserID:          o.UserID,
		Items:           items,
		Subtotal:        o.Subtotal.Int64(),
		SubtotalVND:     o.Subtotal.Format(),
		Discount:        o.Discount.Int64(),
		DiscountVND:     o.Discount.Format(),
		ShippingFee:     o.ShippingFee.Int64(),
		ShippingFeeVND:  o.ShippingFee.Format(),
		Total:           o.Total.Int64(),
		TotalVND:        o.Total.Format(),
		CouponCode:      o.CouponCode,
		Status:          string(o.Status),
		PaymentMethod:   o.PaymentMethod,
		PaymentRef:      o.PaymentRef,
		RecipientName:   o.RecipientName,
		RecipientPhone:  o.RecipientPhone,
		ShippingAddress: o.ShippingAddress,
		CreatedAt:       o.CreatedAt,
		UpdatedAt:       o.UpdatedAt,
	}
}

type ListOrdersResponse struct {
	Items    []OrderDTO `json:"items"`
	Total    int        `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	idempotencyKey := r.Header.Get("Idempotency-Key")

	var req CreateOrderRequest
	if err := decodeJSON(w, r, &req); err != nil {
		presenter.Error(w, r, err)
		return
	}

	ord, err := h.orderUC.CreateOrder(r.Context(), claims.UserID, idempotencyKey, orderuc.CreateOrderInput{
		RecipientName:   req.RecipientName,
		RecipientPhone:  req.RecipientPhone,
		ShippingAddress: req.ShippingAddress,
		PaymentMethod:   req.PaymentMethod,
	})
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.JSON(w, http.StatusCreated, toOrderDTO(ord))
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	orderID := chi.URLParam(r, "id")

	ord, err := h.orderUC.GetOrder(r.Context(), claims.UserID, orderID, claims.Role == "admin")
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.JSON(w, http.StatusOK, toOrderDTO(ord))
}

func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
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
	orders, total, err := h.orderUC.ListOrders(r.Context(), claims.UserID, pageSize, offset)
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	dtos := make([]OrderDTO, 0, len(orders))
	for _, o := range orders {
		dtos = append(dtos, toOrderDTO(&o))
	}

	presenter.JSON(w, http.StatusOK, ListOrdersResponse{
		Items:    dtos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *OrderHandler) PayOrder(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	orderID := chi.URLParam(r, "id")

	ord, err := h.orderUC.PayOrder(r.Context(), claims.UserID, orderID)
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.JSON(w, http.StatusOK, toOrderDTO(ord))
}

func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	orderID := chi.URLParam(r, "id")

	ord, err := h.orderUC.CancelOrder(r.Context(), claims.UserID, orderID)
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.JSON(w, http.StatusOK, toOrderDTO(ord))
}

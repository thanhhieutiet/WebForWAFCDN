package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"labshop/internal/adapter/http/middleware"
	"labshop/internal/adapter/http/presenter"
	"labshop/internal/domain/catalog"
	"labshop/internal/domain/money"
	domainorder "labshop/internal/domain/order"
	adminuc "labshop/internal/usecase/admin"
)

type AdminHandler struct {
	adminUC *adminuc.Usecase
}

func NewAdminHandler(adminUC *adminuc.Usecase) *AdminHandler {
	return &AdminHandler{adminUC: adminUC}
}

type CreateProductRequest struct {
	CategoryID  string `json:"category_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Stock       int    `json:"stock"`
	ImagePath   string `json:"image_path"`
}

type UpdateProductRequest struct {
	CategoryID  *string                `json:"category_id,omitempty"`
	Name        *string                `json:"name,omitempty"`
	Slug        *string                `json:"slug,omitempty"`
	Description *string                `json:"description,omitempty"`
	Price       *int64                 `json:"price,omitempty"`
	Stock       *int                   `json:"stock,omitempty"`
	ImagePath   *string                `json:"image_path,omitempty"`
	Status      *catalog.ProductStatus `json:"status,omitempty"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}

type ReplyTicketRequest struct {
	Message string `json:"message"`
	Close   bool   `json:"close"`
}

type DayRevenueDTO struct {
	Date       string `json:"date"`
	Revenue    int64  `json:"revenue"`
	RevenueVND string `json:"revenue_vnd"`
	OrderCount int    `json:"order_count"`
}

type TopProductDTO struct {
	ProductID    string `json:"product_id"`
	ProductName  string `json:"product_name"`
	QuantitySold int    `json:"quantity_sold"`
	Revenue      int64  `json:"revenue"`
	RevenueVND   string `json:"revenue_vnd"`
}

type AdminStatsDTO struct {
	TotalUsers      int             `json:"total_users"`
	TotalOrders     int             `json:"total_orders"`
	TotalRevenue    int64           `json:"total_revenue"`
	TotalRevenueVND string          `json:"total_revenue_vnd"`
	OrdersByStatus  map[string]int  `json:"orders_by_status"`
	RevenueByDay    []DayRevenueDTO `json:"revenue_by_day"`
	TopProducts     []TopProductDTO `json:"top_products"`
}

func (h *AdminHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest
	if err := decodeJSON(w, r, &req); err != nil {
		presenter.Error(w, r, err)
		return
	}

	prod, err := h.adminUC.CreateProduct(r.Context(), adminuc.CreateProductInput{
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Price:       money.VND(req.Price),
		Stock:       req.Stock,
		ImagePath:   req.ImagePath,
	})
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.JSON(w, http.StatusCreated, toProductDTO(*prod))
}

func (h *AdminHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req UpdateProductRequest
	if err := decodeJSON(w, r, &req); err != nil {
		presenter.Error(w, r, err)
		return
	}

	var priceVND *money.VND
	if req.Price != nil {
		p := money.VND(*req.Price)
		priceVND = &p
	}

	prod, err := h.adminUC.UpdateProduct(r.Context(), id, adminuc.UpdateProductInput{
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Price:       priceVND,
		Stock:       req.Stock,
		ImagePath:   req.ImagePath,
		Status:      req.Status,
	})
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.JSON(w, http.StatusOK, toProductDTO(*prod))
}

func (h *AdminHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.adminUC.DeleteProduct(r.Context(), id); err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.JSON(w, http.StatusOK, map[string]string{"message": "product archived successfully"})
}

func (h *AdminHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
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
	orders, total, err := h.adminUC.ListOrders(r.Context(), pageSize, offset)
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

func (h *AdminHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	var req UpdateOrderStatusRequest
	if err := decodeJSON(w, r, &req); err != nil {
		presenter.Error(w, r, err)
		return
	}

	ord, err := h.adminUC.UpdateOrderStatus(r.Context(), orderID, domainorder.Status(req.Status))
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.JSON(w, http.StatusOK, toOrderDTO(ord))
}

func (h *AdminHandler) ListTickets(w http.ResponseWriter, r *http.Request) {
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
	tickets, total, err := h.adminUC.ListTickets(r.Context(), pageSize, offset)
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	dtos := make([]TicketDTO, 0, len(tickets))
	for _, t := range tickets {
		dtos = append(dtos, toTicketDTO(&t))
	}

	presenter.JSON(w, http.StatusOK, ListTicketsResponse{
		Items:    dtos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *AdminHandler) ReplyTicket(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	ticketID := chi.URLParam(r, "id")

	var req ReplyTicketRequest
	if err := decodeJSON(w, r, &req); err != nil {
		presenter.Error(w, r, err)
		return
	}

	ticket, err := h.adminUC.ReplyTicket(r.Context(), claims.UserID, ticketID, req.Message, req.Close)
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.JSON(w, http.StatusOK, toTicketDTO(ticket))
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
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
	users, total, err := h.adminUC.ListUsers(r.Context(), pageSize, offset)
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	dtos := make([]UserDTO, 0, len(users))
	for _, u := range users {
		dtos = append(dtos, toUserDTO(&u))
	}

	presenter.JSON(w, http.StatusOK, map[string]any{
		"items":     dtos,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.adminUC.GetStats(r.Context())
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	revByDay := make([]DayRevenueDTO, 0, len(stats.RevenueByDay))
	for _, d := range stats.RevenueByDay {
		revByDay = append(revByDay, DayRevenueDTO{
			Date:       d.Date,
			Revenue:    d.Revenue.Int64(),
			RevenueVND: d.Revenue.Format(),
			OrderCount: d.OrderCount,
		})
	}

	topProds := make([]TopProductDTO, 0, len(stats.TopProducts))
	for _, tp := range stats.TopProducts {
		topProds = append(topProds, TopProductDTO{
			ProductID:    tp.ProductID,
			ProductName:  tp.ProductName,
			QuantitySold: tp.QuantitySold,
			Revenue:      tp.Revenue.Int64(),
			RevenueVND:   tp.Revenue.Format(),
		})
	}

	presenter.JSON(w, http.StatusOK, AdminStatsDTO{
		TotalUsers:      stats.TotalUsers,
		TotalOrders:     stats.TotalOrders,
		TotalRevenue:    stats.TotalRevenueVND.Int64(),
		TotalRevenueVND: stats.TotalRevenueVND.Format(),
		OrdersByStatus:  stats.OrdersByStatus,
		RevenueByDay:    revByDay,
		TopProducts:     topProds,
	})
}

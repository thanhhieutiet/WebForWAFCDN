package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"labshop/internal/adapter/counter/memory"
	"labshop/internal/adapter/http/middleware"
	mockpayment "labshop/internal/adapter/payment/mock"
	"labshop/internal/adapter/repository/sqlite"
	"labshop/internal/adapter/security/argon2"
	"labshop/internal/adapter/security/jwt"
	infraDB "labshop/internal/infrastructure/db"
	adminuc "labshop/internal/usecase/admin"
	authuc "labshop/internal/usecase/auth"
	cartuc "labshop/internal/usecase/cart"
	cataloguc "labshop/internal/usecase/catalog"
	orderuc "labshop/internal/usecase/order"
	reviewuc "labshop/internal/usecase/review"
	supportuc "labshop/internal/usecase/support"
)

type testEnv struct {
	router      *chi.Mux
	db          *sqlite.DB
	jwtIssuer   *jwt.Issuer
	adminToken  string
	custToken   string
	custID      string
	adminID     string
}

type uuidGen struct{}

func (uuidGen) NewID() string { return uuid.NewString() }

func setupTestApp(t *testing.T) *testEnv {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_shop.db")

	database, err := infraDB.Init(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	hasher := argon2.New()
	jwtSecret := "testing-jwt-secret-key-32-chars-long!"
	jwtIssuer := jwt.New(jwtSecret, 2*time.Hour)
	paymentGw := mockpayment.New(0.0)
	txManager := sqlite.NewTxManager(database)
	idGen := uuidGen{}

	// Seed initial categories, products, coupons, admin and customer
	err = infraDB.Seed(
		context.Background(),
		database,
		hasher,
		"admin@labshop.local",
		"Admin@LabShop2026!",
		"customer@labshop.local",
		"Customer@LabShop2026!",
	)
	require.NoError(t, err)

	userRepo := sqlite.NewUserRepository(database)
	catRepo := sqlite.NewCatalogRepository(database)
	cartRepo := sqlite.NewCartRepository(database)
	coupRepo := sqlite.NewCouponRepository(database)
	orderRepo := sqlite.NewOrderRepository(database)
	revRepo := sqlite.NewReviewRepository(database)
	suppRepo := sqlite.NewSupportRepository(database)
	adminRepo := sqlite.NewAdminRepository(database)

	authUC := authuc.New(userRepo, hasher, jwtIssuer, idGen)
	catUC := cataloguc.New(catRepo)
	cartUC := cartuc.New(cartRepo, catRepo, coupRepo)
	orderUC := orderuc.New(orderRepo, cartRepo, catRepo, coupRepo, paymentGw, txManager, idGen)
	revUC := reviewuc.New(revRepo, orderRepo, catRepo, idGen)
	suppUC := supportuc.New(suppRepo, idGen)
	adminUC := adminuc.New(catRepo, orderRepo, suppRepo, userRepo, adminRepo, txManager, idGen)

	authH := NewAuthHandler(authUC)
	catH := NewCatalogHandler(catUC)
	cartH := NewCartHandler(cartUC)
	orderH := NewOrderHandler(orderUC)
	revH := NewReviewHandler(revUC)
	suppH := NewSupportHandler(suppUC)
	adminH := NewAdminHandler(adminUC)

	// Build router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.HitCounter(memory.New()))
	r.Use(middleware.AuthMiddleware(jwtIssuer))

	// Auth
	r.Post("/api/auth/register", authH.Register)
	r.Post("/api/auth/login", authH.Login)
	r.Group(func(ar chi.Router) {
		ar.Use(middleware.RequireAuth)
		ar.Get("/api/auth/me", authH.GetMe)
		ar.Patch("/api/auth/me", authH.UpdateMe)
		ar.Post("/api/auth/change-password", authH.ChangePassword)
	})

	// Catalog
	r.Get("/api/categories", catH.ListCategories)
	r.Get("/api/products", catH.ListProducts)
	r.Get("/api/products/{slug}", catH.GetProductBySlug)

	// Cart
	r.Group(func(cr chi.Router) {
		cr.Use(middleware.RequireAuth)
		cr.Get("/api/cart", cartH.GetCart)
		cr.Put("/api/cart/items/{product_id}", cartH.UpdateItemQuantity)
		cr.Delete("/api/cart/items/{product_id}", cartH.RemoveItem)
		cr.Delete("/api/cart", cartH.ClearCart)
		cr.Post("/api/cart/coupon", cartH.ApplyCoupon)
		cr.Delete("/api/cart/coupon", cartH.RemoveCoupon)
	})

	// Orders
	r.Group(func(or chi.Router) {
		or.Use(middleware.RequireAuth)
		or.Post("/api/orders", orderH.CreateOrder)
		or.Get("/api/orders", orderH.ListOrders)
		or.Get("/api/orders/{id}", orderH.GetOrder)
		or.Post("/api/orders/{id}/pay", orderH.PayOrder)
		or.Post("/api/orders/{id}/cancel", orderH.CancelOrder)
	})

	// Reviews
	r.Get("/api/products/{slug}/reviews", revH.ListReviews)
	r.Group(func(rr chi.Router) {
		rr.Use(middleware.RequireAuth)
		rr.Post("/api/products/{slug}/reviews", revH.CreateReview)
		rr.Delete("/api/reviews/{id}", revH.DeleteReview)
	})

	// Support
	r.Group(func(sr chi.Router) {
		sr.Use(middleware.RequireAuth)
		sr.Post("/api/support/tickets", suppH.CreateTicket)
		sr.Get("/api/support/tickets", suppH.ListTickets)
		sr.Get("/api/support/tickets/{id}", suppH.GetTicket)
		sr.Post("/api/support/tickets/{id}/messages", suppH.AddMessage)
		sr.Post("/api/support/tickets/{id}/attachments", suppH.AddAttachment)
	})

	// Admin
	r.Group(func(adm chi.Router) {
		adm.Use(middleware.RequireAdmin)
		adm.Get("/api/admin/products", catH.ListProducts)
		adm.Post("/api/admin/products", adminH.CreateProduct)
		adm.Put("/api/admin/products/{id}", adminH.UpdateProduct)
		adm.Patch("/api/admin/products/{id}", adminH.UpdateProduct)
		adm.Delete("/api/admin/products/{id}", adminH.DeleteProduct)
		adm.Get("/api/admin/orders", adminH.ListOrders)
		adm.Patch("/api/admin/orders/{id}/status", adminH.UpdateOrderStatus)
		adm.Get("/api/admin/tickets", adminH.ListTickets)
		adm.Post("/api/admin/tickets/{id}/reply", adminH.ReplyTicket)
		adm.Get("/api/admin/users", adminH.ListUsers)
		adm.Get("/api/admin/stats", adminH.GetStats)
	})

	// Get seeded accounts to issue tokens
	adminObj, err := userRepo.GetByEmail(context.Background(), "admin@labshop.local")
	require.NoError(t, err)
	adminToken, err := jwtIssuer.IssueToken(adminObj)
	require.NoError(t, err)

	custObj, err := userRepo.GetByEmail(context.Background(), "customer@labshop.local")
	require.NoError(t, err)
	custToken, err := jwtIssuer.IssueToken(custObj)
	require.NoError(t, err)

	return &testEnv{
		router:     r,
		db:         database,
		jwtIssuer:  jwtIssuer,
		adminToken: adminToken,
		custToken:  custToken,
		custID:     custObj.ID,
		adminID:    adminObj.ID,
	}
}

func doJSONRequest(router *chi.Mux, method, path, token string, body any) *httptest.ResponseRecorder {
	var bodyReader io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(data)
	}
	req := httptest.NewRequest(method, path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestBusinessFlow_EndToEnd(t *testing.T) {
	env := setupTestApp(t)

	// 1. Auth Flow: Register new user
	newEmail := "alice@example.com"
	wReg := doJSONRequest(env.router, http.MethodPost, "/api/auth/register", "", map[string]string{
		"email":    newEmail,
		"password": "Password123!",
		"name":     "Alice",
	})
	assert.Equal(t, http.StatusCreated, wReg.Code)
	var regResp AuthResponse
	require.NoError(t, json.Unmarshal(wReg.Body.Bytes(), &regResp))
	assert.Equal(t, newEmail, regResp.User.Email)
	assert.NotEmpty(t, regResp.Token)
	aliceToken := regResp.Token

	// Login
	wLogin := doJSONRequest(env.router, http.MethodPost, "/api/auth/login", "", map[string]string{
		"email":    newEmail,
		"password": "Password123!",
	})
	assert.Equal(t, http.StatusOK, wLogin.Code)

	// Get Me & Update Me
	wMe := doJSONRequest(env.router, http.MethodGet, "/api/auth/me", aliceToken, nil)
	assert.Equal(t, http.StatusOK, wMe.Code)

	newName := "Alice Wonderland"
	wUpMe := doJSONRequest(env.router, http.MethodPatch, "/api/auth/me", aliceToken, map[string]string{
		"name": newName,
	})
	assert.Equal(t, http.StatusOK, wUpMe.Code)

	// 2. Catalog: List categories & products
	wCats := doJSONRequest(env.router, http.MethodGet, "/api/categories", "", nil)
	assert.Equal(t, http.StatusOK, wCats.Code)
	var catsBody map[string][]CategoryDTO
	require.NoError(t, json.Unmarshal(wCats.Body.Bytes(), &catsBody))
	assert.Len(t, catsBody["categories"], 5)

	wProds := doJSONRequest(env.router, http.MethodGet, "/api/products?page=1&page_size=10", "", nil)
	assert.Equal(t, http.StatusOK, wProds.Code)
	var prodsBody ListProductsResponse
	require.NoError(t, json.Unmarshal(wProds.Body.Bytes(), &prodsBody))
	assert.Equal(t, 40, prodsBody.Total)
	assert.Len(t, prodsBody.Items, 10)
	targetProd := prodsBody.Items[0]

	wSlug := doJSONRequest(env.router, http.MethodGet, "/api/products/"+targetProd.Slug, "", nil)
	assert.Equal(t, http.StatusOK, wSlug.Code)

	// 3. Cart Flow: Add item, apply coupon, verify discount
	wCartAdd := doJSONRequest(env.router, http.MethodPut, "/api/cart/items/"+targetProd.ID, aliceToken, map[string]int{
		"quantity": 2,
	})
	assert.Equal(t, http.StatusOK, wCartAdd.Code)

	wCoupon := doJSONRequest(env.router, http.MethodPost, "/api/cart/coupon", aliceToken, map[string]string{
		"code": "WELCOME10",
	})
	assert.Equal(t, http.StatusOK, wCoupon.Code)
	var cartBody CartDTO
	require.NoError(t, json.Unmarshal(wCoupon.Body.Bytes(), &cartBody))
	assert.Equal(t, "WELCOME10", cartBody.CouponCode)
	assert.True(t, cartBody.Discount > 0)

	// 4. Order Flow: Checkout with Idempotency-Key
	reqOrder := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader([]byte(`{
		"recipient_name": "Alice",
		"recipient_phone": "0987654321",
		"shipping_address": "456 Main Street, Hanoi",
		"payment_method": "mock_gateway"
	}`)))
	reqOrder.Header.Set("Content-Type", "application/json")
	reqOrder.Header.Set("Authorization", "Bearer "+aliceToken)
	reqOrder.Header.Set("Idempotency-Key", "idem-alice-order-1")
	wOrder := httptest.NewRecorder()
	env.router.ServeHTTP(wOrder, reqOrder)
	assert.Equal(t, http.StatusCreated, wOrder.Code)
	var ord OrderDTO
	require.NoError(t, json.Unmarshal(wOrder.Body.Bytes(), &ord))
	assert.Equal(t, "PENDING", ord.Status)

	// Replay same Idempotency-Key -> returns existing order
	reqOrder2 := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader([]byte(`{
		"recipient_name": "Alice",
		"recipient_phone": "0987654321",
		"shipping_address": "456 Main Street, Hanoi",
		"payment_method": "mock_gateway"
	}`)))
	reqOrder2.Header.Set("Content-Type", "application/json")
	reqOrder2.Header.Set("Authorization", "Bearer "+aliceToken)
	reqOrder2.Header.Set("Idempotency-Key", "idem-alice-order-1")
	wOrder2 := httptest.NewRecorder()
	env.router.ServeHTTP(wOrder2, reqOrder2)
	assert.True(t, wOrder2.Code == http.StatusOK || wOrder2.Code == http.StatusCreated)
	var ordSame OrderDTO
	require.NoError(t, json.Unmarshal(wOrder2.Body.Bytes(), &ordSame))
	assert.Equal(t, ord.ID, ordSame.ID)

	// Replay same Idempotency-Key with different payload -> 422
	reqOrder3 := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader([]byte(`{
		"recipient_name": "Bob",
		"recipient_phone": "0987654321",
		"shipping_address": "456 Main Street, Hanoi",
		"payment_method": "mock_gateway"
	}`)))
	reqOrder3.Header.Set("Content-Type", "application/json")
	reqOrder3.Header.Set("Authorization", "Bearer "+aliceToken)
	reqOrder3.Header.Set("Idempotency-Key", "idem-alice-order-1")
	wOrder3 := httptest.NewRecorder()
	env.router.ServeHTTP(wOrder3, reqOrder3)
	assert.Equal(t, http.StatusUnprocessableEntity, wOrder3.Code)

	// 5. Pay Order
	wPay := doJSONRequest(env.router, http.MethodPost, fmt.Sprintf("/api/orders/%s/pay", ord.ID), aliceToken, nil)
	assert.Equal(t, http.StatusOK, wPay.Code)
	var ordPaid OrderDTO
	require.NoError(t, json.Unmarshal(wPay.Body.Bytes(), &ordPaid))
	assert.Equal(t, "PAID", ordPaid.Status)
	assert.NotEmpty(t, ordPaid.PaymentRef)

	// 6. Review before DELIVERED -> fails (422)
	wRevFail := doJSONRequest(env.router, http.MethodPost, fmt.Sprintf("/api/products/%s/reviews", targetProd.Slug), aliceToken, map[string]any{
		"rating":  5,
		"comment": "Nice phone before delivery",
	})
	assert.Equal(t, http.StatusUnprocessableEntity, wRevFail.Code)

	// Admin delivers order
	wShip := doJSONRequest(env.router, http.MethodPatch, fmt.Sprintf("/api/admin/orders/%s/status", ord.ID), env.adminToken, map[string]string{
		"status": "SHIPPED",
	})
	assert.Equal(t, http.StatusOK, wShip.Code)
	wDeliver := doJSONRequest(env.router, http.MethodPatch, fmt.Sprintf("/api/admin/orders/%s/status", ord.ID), env.adminToken, map[string]string{
		"status": "DELIVERED",
	})
	assert.Equal(t, http.StatusOK, wDeliver.Code)

	// Review after DELIVERED -> succeeds (201)
	wRevOK := doJSONRequest(env.router, http.MethodPost, fmt.Sprintf("/api/products/%s/reviews", targetProd.Slug), aliceToken, map[string]any{
		"rating":  5,
		"comment": "Absolutely fabulous!",
	})
	assert.Equal(t, http.StatusCreated, wRevOK.Code)

	// Duplicate review -> 409
	wRevDup := doJSONRequest(env.router, http.MethodPost, fmt.Sprintf("/api/products/%s/reviews", targetProd.Slug), aliceToken, map[string]any{
		"rating":  4,
		"comment": "Second review",
	})
	assert.Equal(t, http.StatusConflict, wRevDup.Code)

	// 7. Support Ticket
	wTick := doJSONRequest(env.router, http.MethodPost, "/api/support/tickets", aliceToken, map[string]string{
		"subject":         "Need invoice",
		"initial_message": "Please send VAT invoice.",
	})
	assert.Equal(t, http.StatusCreated, wTick.Code)
	var tick TicketDTO
	require.NoError(t, json.Unmarshal(wTick.Body.Bytes(), &tick))

	// Add attachment to ticket
	bodyBuf := &bytes.Buffer{}
	mpWriter := multipart.NewWriter(bodyBuf)
	part, err := mpWriter.CreateFormFile("file", "invoice_request.pdf")
	require.NoError(t, err)
	_, err = part.Write([]byte("%PDF-1.4 dummy pdf content for invoice"))
	require.NoError(t, err)
	require.NoError(t, mpWriter.Close())

	reqAtt := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/support/tickets/%s/attachments", tick.ID), bodyBuf)
	reqAtt.Header.Set("Content-Type", mpWriter.FormDataContentType())
	reqAtt.Header.Set("Authorization", "Bearer "+aliceToken)
	wAtt := httptest.NewRecorder()
	env.router.ServeHTTP(wAtt, reqAtt)
	assert.Equal(t, http.StatusCreated, wAtt.Code)

	// Admin replies to ticket and closes it
	wReply := doJSONRequest(env.router, http.MethodPost, fmt.Sprintf("/api/admin/tickets/%s/reply", tick.ID), env.adminToken, map[string]any{
		"message": "VAT invoice sent to your email.",
		"close":   true,
	})
	assert.Equal(t, http.StatusOK, wReply.Code)
	var tickClosed TicketDTO
	require.NoError(t, json.Unmarshal(wReply.Body.Bytes(), &tickClosed))
	assert.Equal(t, "CLOSED", tickClosed.Status)

	// 8. Admin Stats
	wStats := doJSONRequest(env.router, http.MethodGet, "/api/admin/stats", env.adminToken, nil)
	assert.Equal(t, http.StatusOK, wStats.Code)
	var stats AdminStatsDTO
	require.NoError(t, json.Unmarshal(wStats.Body.Bytes(), &stats))
	assert.True(t, stats.TotalOrders >= 1)
	assert.True(t, stats.TotalUsers >= 2)
}

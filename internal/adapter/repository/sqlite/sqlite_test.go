package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"labshop/internal/domain/cart"
	"labshop/internal/domain/catalog"
	"labshop/internal/domain/coupon"
	"labshop/internal/domain/money"
	"labshop/internal/domain/order"
	"labshop/internal/domain/review"
	"labshop/internal/domain/support"
	"labshop/internal/domain/user"
)

func setupTestDB(t *testing.T) *DB {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_labshop.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test sqlite db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func TestSQLiteRepositories(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// 1. User Repository
	userRepo := NewUserRepository(db)
	now := time.Now().Truncate(time.Second)
	u := &user.User{
		ID:           "u-1",
		Email:        "user@labshop.local",
		PasswordHash: "hashedpassword123",
		Name:         "John Doe",
		Phone:        "0912345678",
		Address:      "Hanoi, Vietnam",
		Role:         user.RoleCustomer,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := userRepo.Create(ctx, u); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	byEmail, err := userRepo.GetByEmail(ctx, "USER@labshop.local")
	if err != nil || byEmail.ID != "u-1" {
		t.Fatalf("failed to get user by email: %v", err)
	}

	byID, err := userRepo.GetByID(ctx, "u-1")
	if err != nil || byID.Email != u.Email {
		t.Fatalf("failed to get user by id: %v", err)
	}

	u.Name = "John Updated"
	if err := userRepo.Update(ctx, u); err != nil {
		t.Fatalf("failed to update user: %v", err)
	}
	users, totalUsers, err := userRepo.List(ctx, 10, 0)
	if err != nil || totalUsers != 1 || users[0].Name != "John Updated" {
		t.Fatalf("failed to list users: %v", err)
	}

	// 2. Catalog Repository
	catRepo := NewCatalogRepository(db)
	category := &catalog.Category{
		ID:          "cat-1",
		Name:        "Smartphones",
		Slug:        "smartphones",
		Description: "Latest mobile phones",
		CreatedAt:   now,
	}
	if err := catRepo.CreateCategory(ctx, category); err != nil {
		t.Fatalf("failed to create category: %v", err)
	}
	cats, err := catRepo.ListCategories(ctx)
	if err != nil || len(cats) != 1 {
		t.Fatalf("failed to list categories: %v", err)
	}

	prod := &catalog.Product{
		ID:          "prod-1",
		CategoryID:  "cat-1",
		Name:        "iPhone 15 Pro",
		Slug:        "iphone-15-pro",
		Description: "Flagship Apple Phone",
		Price:       money.VND(28000000),
		Stock:       10,
		ImagePath:   "/static/img/products/p01.jpg",
		Status:      catalog.StatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := catRepo.CreateProduct(ctx, prod); err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	prodBySlug, err := catRepo.GetProductBySlug(ctx, "iphone-15-pro")
	if err != nil || prodBySlug.ID != "prod-1" {
		t.Fatalf("failed to get product by slug: %v", err)
	}

	// Deduct stock atomically
	ok, err := catRepo.DeductStockIfAvailable(ctx, "prod-1", 4)
	if err != nil || !ok {
		t.Fatalf("expected successful stock deduction: %v", err)
	}
	// Try deducting more than remaining (10 - 4 = 6)
	ok, err = catRepo.DeductStockIfAvailable(ctx, "prod-1", 7)
	if err != nil || ok {
		t.Fatalf("expected stock deduction to fail when exceeding stock")
	}
	// Restore stock
	if err := catRepo.RestoreStock(ctx, "prod-1", 4); err != nil {
		t.Fatalf("failed to restore stock: %v", err)
	}
	pRestored, _ := catRepo.GetProductByID(ctx, "prod-1")
	if pRestored.Stock != 10 {
		t.Fatalf("expected restored stock 10, got %d", pRestored.Stock)
	}

	// 3. Cart Repository
	cartRepo := NewCartRepository(db)
	userCart := &cart.Cart{
		UserID:     "u-1",
		CouponCode: "WELCOME10",
		Items: []cart.CartItem{
			{ProductID: "prod-1", Quantity: 2, UpdatedAt: now},
		},
		UpdatedAt: now,
	}
	if err := cartRepo.Save(ctx, userCart); err != nil {
		t.Fatalf("failed to save cart: %v", err)
	}
	loadedCart, err := cartRepo.GetByUserID(ctx, "u-1")
	if err != nil || len(loadedCart.Items) != 1 || loadedCart.CouponCode != "WELCOME10" {
		t.Fatalf("failed to load cart: %v", err)
	}
	if err := cartRepo.DeleteByUserID(ctx, "u-1"); err != nil {
		t.Fatalf("failed to delete cart: %v", err)
	}

	// 4. Coupon Repository
	couponRepo := NewCouponRepository(db)
	coup := &coupon.Coupon{
		ID:          "coup-1",
		Code:        "WELCOME10",
		Type:        coupon.TypePercent,
		Value:       10,
		MaxDiscount: money.VND(100000),
		MinSubtotal: money.VND(200000),
		ValidFrom:   now.Add(-time.Hour),
		ValidTo:     now.Add(time.Hour),
		MaxUses:     100,
		UsedCount:   0,
		CreatedAt:   now,
	}
	if err := couponRepo.Save(ctx, coup); err != nil {
		t.Fatalf("failed to save coupon: %v", err)
	}
	loadedCoup, err := couponRepo.GetByCode(ctx, "welcome10")
	if err != nil || loadedCoup.ID != "coup-1" {
		t.Fatalf("failed to get coupon: %v", err)
	}
	if err := couponRepo.IncrementUsage(ctx, "WELCOME10"); err != nil {
		t.Fatalf("failed to increment coupon usage: %v", err)
	}

	// 5. Order Repository
	orderRepo := NewOrderRepository(db)
	ord := &order.Order{
		ID:              "ord-1",
		UserID:          "u-1",
		Subtotal:        money.VND(28000000),
		Discount:        money.VND(100000),
		ShippingFee:     money.VND(0),
		Total:           money.VND(27900000),
		CouponCode:      "WELCOME10",
		Status:          order.StatusPending,
		PaymentMethod:   "mock_gateway",
		RecipientName:   "John Doe",
		RecipientPhone:  "0912345678",
		ShippingAddress: "Hanoi, Vietnam",
		IdempotencyKey:  "idem-123",
		PayloadHash:     "hash-123",
		Items: []order.OrderItem{
			{
				ID:           "item-1",
				OrderID:      "ord-1",
				ProductID:    "prod-1",
				ProductName:  "iPhone 15 Pro",
				ProductPrice: money.VND(28000000),
				Quantity:     1,
				Subtotal:     money.VND(28000000),
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := orderRepo.Create(ctx, ord); err != nil {
		t.Fatalf("failed to create order: %v", err)
	}

	byIdem, err := orderRepo.GetByIdempotencyKey(ctx, "u-1", "idem-123")
	if err != nil || byIdem.ID != "ord-1" || len(byIdem.Items) != 1 {
		t.Fatalf("failed to get order by idempotency key: %v", err)
	}

	if err := orderRepo.UpdateStatus(ctx, "ord-1", order.StatusPaid, "mock_pay_ref_1"); err != nil {
		t.Fatalf("failed to update order status: %v", err)
	}

	hasDelivered, err := orderRepo.HasDeliveredOrderWithProduct(ctx, "u-1", "prod-1")
	if err != nil || hasDelivered {
		t.Fatalf("expected hasDelivered to be false for PAID order")
	}
	_ = orderRepo.UpdateStatus(ctx, "ord-1", order.StatusDelivered, "")
	hasDelivered, err = orderRepo.HasDeliveredOrderWithProduct(ctx, "u-1", "prod-1")
	if err != nil || !hasDelivered {
		t.Fatalf("expected hasDelivered to be true for DELIVERED order")
	}

	// 6. Review Repository
	revRepo := NewReviewRepository(db)
	rev := &review.Review{
		ID:        "rev-1",
		UserID:    "u-1",
		ProductID: "prod-1",
		UserName:  "John",
		Rating:    5,
		Comment:   "Excellent phone!",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := revRepo.Create(ctx, rev); err != nil {
		t.Fatalf("failed to create review: %v", err)
	}
	avg, count, err := revRepo.GetProductRatingStats(ctx, "prod-1")
	if err != nil || count != 1 || avg != 5.0 {
		t.Fatalf("failed to get rating stats: avg=%v, count=%d, err=%v", avg, count, err)
	}

	// 7. Support Repository
	suppRepo := NewSupportRepository(db)
	ticket := &support.Ticket{
		ID:        "tick-1",
		UserID:    "u-1",
		UserEmail: "user@labshop.local",
		Subject:   "Warranty enquiry",
		Status:    support.StatusOpen,
		Messages: []support.TicketMessage{
			{
				ID:         "msg-1",
				TicketID:   "tick-1",
				SenderID:   "u-1",
				SenderRole: "customer",
				Message:    "How long is warranty?",
				CreatedAt:  now,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := suppRepo.CreateTicket(ctx, ticket); err != nil {
		t.Fatalf("failed to create ticket: %v", err)
	}
	if err := suppRepo.AddAttachment(ctx, &support.TicketAttachment{
		ID:          "att-1",
		TicketID:    "tick-1",
		FileName:    "invoice.pdf",
		FileSize:    1024,
		SHA256:      "sha256abc",
		ContentType: "application/pdf",
		UploadedAt:  now,
	}); err != nil {
		t.Fatalf("failed to add attachment: %v", err)
	}
	loadedTicket, err := suppRepo.GetTicketByID(ctx, "tick-1")
	if err != nil || len(loadedTicket.Messages) != 1 || len(loadedTicket.Attachments) != 1 {
		t.Fatalf("failed to load ticket: %v", err)
	}

	// 8. Admin Repository Stats
	adminRepo := NewAdminRepository(db)
	stats, err := adminRepo.GetStats(ctx)
	if err != nil || stats.TotalUsers != 1 || stats.TotalOrders != 1 || len(stats.TopProducts) != 1 {
		t.Fatalf("failed to get admin stats: %+v, err: %v", stats, err)
	}

	// 9. TxManager rollback verification
	txManager := NewTxManager(db)
	errRollback := txManager.WithTx(ctx, func(txCtx context.Context) error {
		_ = catRepo.CreateProduct(txCtx, &catalog.Product{
			ID:         "prod-rolled-back",
			CategoryID: "cat-1",
			Name:       "Rolled Back Product",
			Slug:       "rolled-back",
			Price:      money.VND(100),
			Stock:      1,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		})
		return errors.New("abort transaction")
	})
	if errRollback == nil {
		t.Fatalf("expected error from rollback")
	}
	_, errNotFound := catRepo.GetProductByID(ctx, "prod-rolled-back")
	if errNotFound == nil {
		t.Fatalf("expected product to not exist after rollback")
	}
}

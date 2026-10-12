package order

import (
	"context"
	"errors"
	"testing"
	"time"

	domaincart "labshop/internal/domain/cart"
	"labshop/internal/domain/catalog"
	"labshop/internal/domain/coupon"
	"labshop/internal/domain/money"
	domainorder "labshop/internal/domain/order"
	"labshop/internal/usecase/port"
)

type fakeOrderRepo struct {
	orders map[string]*domainorder.Order
}

func newFakeOrderRepo() *fakeOrderRepo {
	return &fakeOrderRepo{orders: make(map[string]*domainorder.Order)}
}

func (f *fakeOrderRepo) Create(ctx context.Context, o *domainorder.Order) error {
	f.orders[o.ID] = o
	return nil
}

func (f *fakeOrderRepo) GetByID(ctx context.Context, id string) (*domainorder.Order, error) {
	o, ok := f.orders[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return o, nil
}

func (f *fakeOrderRepo) GetByIdempotencyKey(ctx context.Context, userID, key string) (*domainorder.Order, error) {
	for _, o := range f.orders {
		if o.UserID == userID && o.IdempotencyKey == key {
			return o, nil
		}
	}
	return nil, errors.New("not found")
}

func (f *fakeOrderRepo) UpdateStatus(ctx context.Context, id string, status domainorder.Status, paymentRef string) error {
	o, ok := f.orders[id]
	if !ok {
		return errors.New("not found")
	}
	o.Status = status
	if paymentRef != "" {
		o.PaymentRef = paymentRef
	}
	return nil
}

func (f *fakeOrderRepo) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]domainorder.Order, int, error) {
	var list []domainorder.Order
	for _, o := range f.orders {
		if o.UserID == userID {
			list = append(list, *o)
		}
	}
	return list, len(list), nil
}

func (f *fakeOrderRepo) ListAll(ctx context.Context, limit, offset int) ([]domainorder.Order, int, error) {
	var list []domainorder.Order
	for _, o := range f.orders {
		list = append(list, *o)
	}
	return list, len(list), nil
}

func (f *fakeOrderRepo) HasDeliveredOrderWithProduct(ctx context.Context, userID, productID string) (bool, error) {
	for _, o := range f.orders {
		if o.UserID == userID && o.Status == domainorder.StatusDelivered {
			for _, it := range o.Items {
				if it.ProductID == productID {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

type fakeTxManager struct{}

func (f fakeTxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type fakeCartRepo struct {
	carts map[string]*domaincart.Cart
}

func (f *fakeCartRepo) GetByUserID(ctx context.Context, userID string) (*domaincart.Cart, error) {
	c, ok := f.carts[userID]
	if !ok {
		return nil, errors.New("not found")
	}
	return c, nil
}

func (f *fakeCartRepo) Save(ctx context.Context, c *domaincart.Cart) error {
	f.carts[c.UserID] = c
	return nil
}

func (f *fakeCartRepo) DeleteByUserID(ctx context.Context, userID string) error {
	delete(f.carts, userID)
	return nil
}

type fakeCatalogRepo struct {
	products map[string]*catalog.Product
}

func (f *fakeCatalogRepo) ListCategories(ctx context.Context) ([]catalog.Category, error) {
	return nil, nil
}
func (f *fakeCatalogRepo) GetCategoryByID(ctx context.Context, id string) (*catalog.Category, error) {
	return nil, nil
}
func (f *fakeCatalogRepo) CreateCategory(ctx context.Context, c *catalog.Category) error {
	return nil
}
func (f *fakeCatalogRepo) ListProducts(ctx context.Context, filter port.ProductFilter) ([]catalog.Product, int, error) {
	return nil, 0, nil
}
func (f *fakeCatalogRepo) GetProductByID(ctx context.Context, id string) (*catalog.Product, error) {
	p, ok := f.products[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return p, nil
}
func (f *fakeCatalogRepo) GetProductBySlug(ctx context.Context, slug string) (*catalog.Product, error) {
	return nil, nil
}
func (f *fakeCatalogRepo) CreateProduct(ctx context.Context, p *catalog.Product) error {
	return nil
}
func (f *fakeCatalogRepo) UpdateProduct(ctx context.Context, p *catalog.Product) error {
	return nil
}
func (f *fakeCatalogRepo) DeductStockIfAvailable(ctx context.Context, id string, quantity int) (bool, error) {
	p, ok := f.products[id]
	if !ok || p.Stock < quantity {
		return false, nil
	}
	p.Stock -= quantity
	return true, nil
}
func (f *fakeCatalogRepo) RestoreStock(ctx context.Context, id string, quantity int) error {
	if p, ok := f.products[id]; ok {
		p.Stock += quantity
	}
	return nil
}
func (f *fakeCatalogRepo) UpdateProductRating(ctx context.Context, id string, ratingAvg float64, ratingCount int) error {
	return nil
}

type fakeCouponRepo struct {
	coupons map[string]*coupon.Coupon
}

func (f *fakeCouponRepo) GetByCode(ctx context.Context, code string) (*coupon.Coupon, error) {
	c, ok := f.coupons[code]
	if !ok {
		return nil, errors.New("not found")
	}
	return c, nil
}
func (f *fakeCouponRepo) IncrementUsage(ctx context.Context, code string) error {
	if c, ok := f.coupons[code]; ok {
		c.UsedCount++
	}
	return nil
}
func (f *fakeCouponRepo) List(ctx context.Context) ([]coupon.Coupon, error) { return nil, nil }
func (f *fakeCouponRepo) Save(ctx context.Context, c *coupon.Coupon) error  { return nil }

type fakePaymentGw struct {
	fail bool
}

func (f fakePaymentGw) Charge(ctx context.Context, orderID string, amount money.VND) (string, error) {
	if f.fail {
		return "", errors.New("charge failed")
	}
	return "mock-ref-" + orderID, nil
}

type fakeIDGen struct {
	counter int
}

func (g *fakeIDGen) NewID() string {
	g.counter++
	return "id-" + string(rune('0'+g.counter))
}

func TestOrderUsecase(t *testing.T) {
	orderRepo := newFakeOrderRepo()
	cartRepo := &fakeCartRepo{carts: make(map[string]*domaincart.Cart)}
	catalogRepo := &fakeCatalogRepo{products: make(map[string]*catalog.Product)}
	couponRepo := &fakeCouponRepo{coupons: make(map[string]*coupon.Coupon)}
	paymentGw := fakePaymentGw{}
	txManager := fakeTxManager{}
	idGen := &fakeIDGen{}

	prod := &catalog.Product{
		ID:     "p-1",
		Name:   "Headphones",
		Price:  money.VND(300000),
		Stock:  5,
		Status: catalog.StatusActive,
	}
	catalogRepo.products[prod.ID] = prod

	cart := domaincart.NewCart("u-1")
	_ = cart.SetItem(prod, 2)
	cartRepo.carts["u-1"] = cart

	uc := New(orderRepo, cartRepo, catalogRepo, couponRepo, paymentGw, txManager, idGen)
	ctx := context.Background()

	input := CreateOrderInput{
		RecipientName:   "Alice",
		RecipientPhone:  "0912345678",
		ShippingAddress: "123 Main St",
		PaymentMethod:   domainorder.PaymentMethodMockGateway,
	}

	// Create order
	ord, err := uc.CreateOrder(ctx, "u-1", "idem-key-1", input)
	if err != nil {
		t.Fatalf("failed to create order: %v", err)
	}
	if ord.Status != domainorder.StatusPending {
		t.Fatalf("expected PENDING status, got %s", ord.Status)
	}
	// Stock should be deducted from 5 to 3
	if catalogRepo.products["p-1"].Stock != 3 {
		t.Fatalf("expected stock 3, got %d", catalogRepo.products["p-1"].Stock)
	}
	// Cart should be deleted
	if _, ok := cartRepo.carts["u-1"]; ok {
		t.Fatalf("expected cart to be deleted after order")
	}

	// Same idempotency key + same payload -> returns existing order
	ordSame, err := uc.CreateOrder(ctx, "u-1", "idem-key-1", input)
	if err != nil || ordSame.ID != ord.ID {
		t.Fatalf("expected same order returned, got %v, err: %v", ordSame, err)
	}

	// Same idempotency key + different payload -> idempotency conflict
	inputDiff := input
	inputDiff.RecipientName = "Bob"
	_, err = uc.CreateOrder(ctx, "u-1", "idem-key-1", inputDiff)
	if err == nil {
		t.Fatalf("expected idempotency conflict error")
	}

	// Pay order
	paidOrd, err := uc.PayOrder(ctx, "u-1", ord.ID)
	if err != nil || paidOrd.Status != domainorder.StatusPaid || paidOrd.PaymentRef == "" {
		t.Fatalf("expected PAID order with ref, got %v, err: %v", paidOrd, err)
	}

	// Cancel order
	cancelledOrd, err := uc.CancelOrder(ctx, "u-1", ord.ID)
	if err != nil || cancelledOrd.Status != domainorder.StatusCancelled {
		t.Fatalf("expected CANCELLED order, got %v, err: %v", cancelledOrd, err)
	}
	// Stock restored from 3 back to 5
	if catalogRepo.products["p-1"].Stock != 5 {
		t.Fatalf("expected stock restored to 5, got %d", catalogRepo.products["p-1"].Stock)
	}

	// GetOrder
	gotOrd, err := uc.GetOrder(ctx, "u-1", ord.ID, false)
	if err != nil || gotOrd.ID != ord.ID {
		t.Fatalf("failed to get order: %v", err)
	}
	_, err = uc.GetOrder(ctx, "u-other", ord.ID, false)
	if err == nil {
		t.Fatalf("expected error getting order for other user")
	}

	// ListOrders
	list, total, err := uc.ListOrders(ctx, "u-1", 10, 0)
	if err != nil || total < 1 || len(list) < 1 {
		t.Fatalf("failed to list orders: %v", err)
	}

	// Empty cart creation
	_, err = uc.CreateOrder(ctx, "u-empty", "idem-empty", input)
	if err == nil {
		t.Fatalf("expected error creating order from empty cart")
	}

	// PayOrder non-existent or wrong user
	_, err = uc.PayOrder(ctx, "u-1", "non-existent")
	if err == nil {
		t.Fatalf("expected error paying non-existent order")
	}
	_, err = uc.PayOrder(ctx, "u-other", ord.ID)
	if err == nil {
		t.Fatalf("expected error paying other user's order")
	}

	// CancelOrder non-existent or wrong user
	_, err = uc.CancelOrder(ctx, "u-1", "non-existent")
	if err == nil {
		t.Fatalf("expected error cancelling non-existent order")
	}
	_, err = uc.CancelOrder(ctx, "u-other", ord.ID)
	if err == nil {
		t.Fatalf("expected error cancelling other user's order")
	}

	// Checkout with coupon
	coup := &coupon.Coupon{
		Code:        "SAVE10",
		Type:        coupon.TypePercent,
		Value:       10,
		MinSubtotal: money.VND(100000),
		ValidFrom:   time.Now().Add(-time.Hour),
		ValidTo:     time.Now().Add(time.Hour),
	}
	couponRepo.coupons[coup.Code] = coup

	cartCoup := domaincart.NewCart("u-coup")
	_ = cartCoup.SetItem(prod, 1)
	cartCoup.ApplyCoupon("SAVE10")
	cartRepo.carts["u-coup"] = cartCoup

	ordCoup, err := uc.CreateOrder(ctx, "u-coup", "idem-coup-1", input)
	if err != nil {
		t.Fatalf("failed to create order with coupon: %v", err)
	}
	if ordCoup.Discount == 0 || ordCoup.CouponCode != "SAVE10" {
		t.Fatalf("expected discount applied from coupon, got %v", ordCoup.Discount)
	}
}

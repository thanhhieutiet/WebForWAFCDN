package admin

import (
	"context"
	"errors"
	"testing"
	"time"

	"labshop/internal/domain/catalog"
	"labshop/internal/domain/money"
	domainorder "labshop/internal/domain/order"
	domainsupport "labshop/internal/domain/support"
	"labshop/internal/domain/user"
	"labshop/internal/usecase/port"
)

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
	for _, p := range f.products {
		if p.Slug == slug {
			return p, nil
		}
	}
	return nil, errors.New("not found")
}
func (f *fakeCatalogRepo) CreateProduct(ctx context.Context, p *catalog.Product) error {
	f.products[p.ID] = p
	return nil
}
func (f *fakeCatalogRepo) UpdateProduct(ctx context.Context, p *catalog.Product) error {
	f.products[p.ID] = p
	return nil
}
func (f *fakeCatalogRepo) DeductStockIfAvailable(ctx context.Context, id string, quantity int) (bool, error) {
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

type fakeOrderRepo struct {
	orders map[string]*domainorder.Order
}

func (f *fakeOrderRepo) Create(ctx context.Context, o *domainorder.Order) error { return nil }
func (f *fakeOrderRepo) GetByID(ctx context.Context, id string) (*domainorder.Order, error) {
	o, ok := f.orders[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return o, nil
}
func (f *fakeOrderRepo) GetByIdempotencyKey(ctx context.Context, userID, key string) (*domainorder.Order, error) {
	return nil, nil
}
func (f *fakeOrderRepo) UpdateStatus(ctx context.Context, id string, status domainorder.Status, paymentRef string) error {
	if o, ok := f.orders[id]; ok {
		o.Status = status
	}
	return nil
}
func (f *fakeOrderRepo) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]domainorder.Order, int, error) {
	return nil, 0, nil
}
func (f *fakeOrderRepo) ListAll(ctx context.Context, limit, offset int) ([]domainorder.Order, int, error) {
	var list []domainorder.Order
	for _, o := range f.orders {
		list = append(list, *o)
	}
	return list, len(list), nil
}
func (f *fakeOrderRepo) HasDeliveredOrderWithProduct(ctx context.Context, userID, productID string) (bool, error) {
	return false, nil
}

type fakeSupportRepo struct {
	tickets map[string]*domainsupport.Ticket
}

func (f *fakeSupportRepo) CreateTicket(ctx context.Context, t *domainsupport.Ticket) error { return nil }
func (f *fakeSupportRepo) GetTicketByID(ctx context.Context, id string) (*domainsupport.Ticket, error) {
	t, ok := f.tickets[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return t, nil
}
func (f *fakeSupportRepo) ListTickets(ctx context.Context, userID string, limit, offset int) ([]domainsupport.Ticket, int, error) {
	var list []domainsupport.Ticket
	for _, t := range f.tickets {
		list = append(list, *t)
	}
	return list, len(list), nil
}
func (f *fakeSupportRepo) AddMessage(ctx context.Context, msg *domainsupport.TicketMessage) error {
	if t, ok := f.tickets[msg.TicketID]; ok {
		t.Messages = append(t.Messages, *msg)
	}
	return nil
}
func (f *fakeSupportRepo) AddAttachment(ctx context.Context, att *domainsupport.TicketAttachment) error {
	return nil
}
func (f *fakeSupportRepo) UpdateTicketStatus(ctx context.Context, ticketID string, status domainsupport.TicketStatus) error {
	if t, ok := f.tickets[ticketID]; ok {
		t.Status = status
	}
	return nil
}

type fakeUserRepo struct {
	users []user.User
}

func (f *fakeUserRepo) Create(ctx context.Context, u *user.User) error                 { return nil }
func (f *fakeUserRepo) GetByID(ctx context.Context, id string) (*user.User, error)     { return nil, nil }
func (f *fakeUserRepo) GetByEmail(ctx context.Context, email string) (*user.User, error) { return nil, nil }
func (f *fakeUserRepo) Update(ctx context.Context, u *user.User) error                 { return nil }
func (f *fakeUserRepo) List(ctx context.Context, limit, offset int) ([]user.User, int, error) {
	return f.users, len(f.users), nil
}

type fakeAdminRepo struct {
	stats *port.AdminStats
}

func (f *fakeAdminRepo) GetStats(ctx context.Context) (*port.AdminStats, error) {
	return f.stats, nil
}

type fakeTxManager struct{}

func (f fakeTxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type fakeIDGen struct{}

func (f fakeIDGen) NewID() string { return "gen-id" }

func TestAdminUsecase(t *testing.T) {
	catRepo := &fakeCatalogRepo{products: make(map[string]*catalog.Product)}
	orderRepo := &fakeOrderRepo{orders: make(map[string]*domainorder.Order)}
	suppRepo := &fakeSupportRepo{tickets: make(map[string]*domainsupport.Ticket)}
	userRepo := &fakeUserRepo{users: []user.User{{ID: "u-1", Email: "test@lab.com"}}}
	adminRepo := &fakeAdminRepo{stats: &port.AdminStats{TotalUsers: 1}}
	txManager := fakeTxManager{}
	idGen := fakeIDGen{}

	uc := New(catRepo, orderRepo, suppRepo, userRepo, adminRepo, txManager, idGen)
	ctx := context.Background()

	// Create Product
	p, err := uc.CreateProduct(ctx, CreateProductInput{
		CategoryID: "c-1",
		Name:       "Phone",
		Slug:       "phone",
		Price:      money.VND(1000000),
		Stock:      5,
	})
	if err != nil || p.ID != "gen-id" {
		t.Fatalf("failed to create product: %v", err)
	}

	// Archive product (delete)
	err = uc.DeleteProduct(ctx, p.ID)
	if err != nil || catRepo.products[p.ID].Status != catalog.StatusArchived {
		t.Fatalf("failed to archive product: %v", err)
	}

	// Update product
	newName := "Updated Phone"
	newPrice := money.VND(1200000)
	newSlug := "updated-phone"
	newDesc := "New description"
	newCat := "cat-2"
	newStock := 20
	newImg := "/static/img/new.jpg"
	newStatus := catalog.StatusActive
	pUp, err := uc.UpdateProduct(ctx, p.ID, UpdateProductInput{
		Name:        &newName,
		Price:       &newPrice,
		Slug:        &newSlug,
		Description: &newDesc,
		CategoryID:  &newCat,
		Stock:       &newStock,
		ImagePath:   &newImg,
		Status:      &newStatus,
	})
	if err != nil || pUp.Name != newName || pUp.Price != newPrice || pUp.Slug != newSlug {
		t.Fatalf("failed to update product: %v", err)
	}

	// Update order status
	orderRepo.orders["ord-1"] = &domainorder.Order{
		ID:        "ord-1",
		Status:    domainorder.StatusPending,
		Items:     []domainorder.OrderItem{{ProductID: p.ID, Quantity: 2}},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	ordUpdated, err := uc.UpdateOrderStatus(ctx, "ord-1", domainorder.StatusPaid)
	if err != nil || ordUpdated.Status != domainorder.StatusPaid {
		t.Fatalf("failed to update order status: %v", err)
	}

	// Cancel order and verify stock restored
	ordCancelled, err := uc.UpdateOrderStatus(ctx, "ord-1", domainorder.StatusCancelled)
	if err != nil || ordCancelled.Status != domainorder.StatusCancelled {
		t.Fatalf("failed to cancel order: %v", err)
	}

	// List orders
	orders, totalOrd, err := uc.ListOrders(ctx, 10, 0)
	if err != nil || totalOrd < 1 || len(orders) < 1 {
		t.Fatalf("failed to list orders: %v", err)
	}

	// Reply ticket
	suppRepo.tickets["tick-1"] = &domainsupport.Ticket{
		ID:        "tick-1",
		Status:    domainsupport.StatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	tick, err := uc.ReplyTicket(ctx, "admin-1", "tick-1", "Resolved", true)
	if err != nil || tick.Status != domainsupport.StatusClosed {
		t.Fatalf("failed to reply and close ticket: %v", err)
	}

	// List tickets
	tickets, totalTick, err := uc.ListTickets(ctx, 10, 0)
	if err != nil || totalTick < 1 || len(tickets) < 1 {
		t.Fatalf("failed to list tickets: %v", err)
	}

	// List users
	users, total, err := uc.ListUsers(ctx, 10, 0)
	if err != nil || total != 1 || len(users) != 1 {
		t.Fatalf("failed to list users: %v", err)
	}

	// Stats
	stats, err := uc.GetStats(ctx)
	if err != nil || stats.TotalUsers != 1 {
		t.Fatalf("failed to get stats: %v", err)
	}

	// Update non-existent product
	_, err = uc.UpdateProduct(ctx, "non-existent", UpdateProductInput{})
	if err == nil {
		t.Fatalf("expected error updating non-existent product")
	}

	// Create product with duplicate slug
	_, err = uc.CreateProduct(ctx, CreateProductInput{
		CategoryID: "c-1",
		Name:       "Phone 2",
		Slug:       "updated-phone",
		Price:      money.VND(1000000),
		Stock:      5,
	})
	if err == nil {
		t.Fatalf("expected error creating product with duplicate slug")
	}

	// Update non-existent order status
	_, err = uc.UpdateOrderStatus(ctx, "non-existent", domainorder.StatusPaid)
	if err == nil {
		t.Fatalf("expected error updating non-existent order status")
	}

	// Invalid order status transition
	_, err = uc.UpdateOrderStatus(ctx, "ord-1", domainorder.StatusDelivered)
	if err == nil {
		t.Fatalf("expected error on invalid state transition")
	}

	// Create product with invalid price
	_, err = uc.CreateProduct(ctx, CreateProductInput{
		CategoryID: "c-1",
		Name:       "Zero Price",
		Slug:       "zero-price",
		Price:      0,
	})
	if err == nil {
		t.Fatalf("expected validation error for zero price")
	}

	// Delete non-existent product
	err = uc.DeleteProduct(ctx, "non-existent")
	if err == nil {
		t.Fatalf("expected error deleting non-existent product")
	}

	// Reply to non-existent ticket
	_, err = uc.ReplyTicket(ctx, "admin-1", "non-existent", "test", false)
	if err == nil {
		t.Fatalf("expected error replying to non-existent ticket")
	}
}

package admin

import (
	"context"
	"strings"
	"time"

	"labshop/internal/domain/catalog"
	"labshop/internal/domain/errs"
	"labshop/internal/domain/money"
	domainorder "labshop/internal/domain/order"
	domainsupport "labshop/internal/domain/support"
	"labshop/internal/domain/user"
	"labshop/internal/usecase/port"
)

type CreateProductInput struct {
	CategoryID  string
	Name        string
	Slug        string
	Description string
	Price       money.VND
	Stock       int
	ImagePath   string
}

type UpdateProductInput struct {
	CategoryID  *string
	Name        *string
	Slug        *string
	Description *string
	Price       *money.VND
	Stock       *int
	ImagePath   *string
	Status      *catalog.ProductStatus
}

type Usecase struct {
	catalogRepo port.CatalogRepository
	orderRepo   port.OrderRepository
	supportRepo port.SupportRepository
	userRepo    port.UserRepository
	adminRepo   port.AdminRepository
	txManager   port.TxManager
	idGen       port.IDGenerator
}

func New(
	catalogRepo port.CatalogRepository,
	orderRepo port.OrderRepository,
	supportRepo port.SupportRepository,
	userRepo port.UserRepository,
	adminRepo port.AdminRepository,
	txManager port.TxManager,
	idGen port.IDGenerator,
) *Usecase {
	return &Usecase{
		catalogRepo: catalogRepo,
		orderRepo:   orderRepo,
		supportRepo: supportRepo,
		userRepo:    userRepo,
		adminRepo:   adminRepo,
		txManager:   txManager,
		idGen:       idGen,
	}
}

// Product CRUD
func (u *Usecase) CreateProduct(ctx context.Context, in CreateProductInput) (*catalog.Product, error) {
	prod := &catalog.Product{
		ID:          u.idGen.NewID(),
		CategoryID:  strings.TrimSpace(in.CategoryID),
		Name:        strings.TrimSpace(in.Name),
		Slug:        strings.TrimSpace(in.Slug),
		Description: strings.TrimSpace(in.Description),
		Price:       in.Price,
		Stock:       in.Stock,
		ImagePath:   strings.TrimSpace(in.ImagePath),
		Status:      catalog.StatusActive,
		RatingAvg:   0,
		RatingCount: 0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := prod.Validate(); err != nil {
		return nil, err
	}

	existing, _ := u.catalogRepo.GetProductBySlug(ctx, prod.Slug)
	if existing != nil {
		return nil, errs.NewConflict("product with this slug already exists")
	}

	if err := u.catalogRepo.CreateProduct(ctx, prod); err != nil {
		return nil, err
	}

	return prod, nil
}

func (u *Usecase) UpdateProduct(ctx context.Context, id string, in UpdateProductInput) (*catalog.Product, error) {
	prod, err := u.catalogRepo.GetProductByID(ctx, id)
	if err != nil || prod == nil {
		return nil, errs.NewNotFound("product", id)
	}

	if in.CategoryID != nil {
		prod.CategoryID = strings.TrimSpace(*in.CategoryID)
	}
	if in.Name != nil {
		prod.Name = strings.TrimSpace(*in.Name)
	}
	if in.Slug != nil {
		slug := strings.TrimSpace(*in.Slug)
		if slug != prod.Slug {
			existing, _ := u.catalogRepo.GetProductBySlug(ctx, slug)
			if existing != nil && existing.ID != prod.ID {
				return nil, errs.NewConflict("product with this slug already exists")
			}
			prod.Slug = slug
		}
	}
	if in.Description != nil {
		prod.Description = strings.TrimSpace(*in.Description)
	}
	if in.Price != nil {
		prod.Price = *in.Price
	}
	if in.Stock != nil {
		prod.Stock = *in.Stock
	}
	if in.ImagePath != nil {
		prod.ImagePath = strings.TrimSpace(*in.ImagePath)
	}
	if in.Status != nil {
		prod.Status = *in.Status
	}
	prod.UpdatedAt = time.Now()

	if err := prod.Validate(); err != nil {
		return nil, err
	}

	if err := u.catalogRepo.UpdateProduct(ctx, prod); err != nil {
		return nil, err
	}

	return prod, nil
}

func (u *Usecase) DeleteProduct(ctx context.Context, id string) error {
	prod, err := u.catalogRepo.GetProductByID(ctx, id)
	if err != nil || prod == nil {
		return errs.NewNotFound("product", id)
	}

	prod.Archive()
	prod.UpdatedAt = time.Now()
	return u.catalogRepo.UpdateProduct(ctx, prod)
}

// Orders
func (u *Usecase) ListOrders(ctx context.Context, limit, offset int) ([]domainorder.Order, int, error) {
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return u.orderRepo.ListAll(ctx, limit, offset)
}

func (u *Usecase) UpdateOrderStatus(ctx context.Context, orderID string, targetStatus domainorder.Status) (*domainorder.Order, error) {
	ord, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil || ord == nil {
		return nil, errs.NewNotFound("order", orderID)
	}

	if err := ord.TransitionTo(targetStatus); err != nil {
		return nil, err
	}

	err = u.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if targetStatus == domainorder.StatusCancelled {
			for _, item := range ord.Items {
				if err := u.catalogRepo.RestoreStock(txCtx, item.ProductID, item.Quantity); err != nil {
					return err
				}
			}
		}
		return u.orderRepo.UpdateStatus(txCtx, ord.ID, ord.Status, ord.PaymentRef)
	})
	if err != nil {
		return nil, err
	}

	return ord, nil
}

// Support Tickets
func (u *Usecase) ListTickets(ctx context.Context, limit, offset int) ([]domainsupport.Ticket, int, error) {
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return u.supportRepo.ListTickets(ctx, "", limit, offset)
}

func (u *Usecase) ReplyTicket(ctx context.Context, adminID, ticketID, message string, closeTicket bool) (*domainsupport.Ticket, error) {
	ticket, err := u.supportRepo.GetTicketByID(ctx, ticketID)
	if err != nil || ticket == nil {
		return nil, errs.NewNotFound("ticket", ticketID)
	}

	msg := domainsupport.TicketMessage{
		ID:         u.idGen.NewID(),
		TicketID:   ticketID,
		SenderID:   adminID,
		SenderRole: "admin",
		Message:    strings.TrimSpace(message),
		CreatedAt:  time.Now(),
	}

	if err := ticket.AddMessage(msg); err != nil {
		return nil, err
	}

	if closeTicket {
		ticket.Close()
	}

	if err := u.supportRepo.AddMessage(ctx, &msg); err != nil {
		return nil, err
	}

	if err := u.supportRepo.UpdateTicketStatus(ctx, ticket.ID, ticket.Status); err != nil {
		return nil, err
	}

	return ticket, nil
}

// Users
func (u *Usecase) ListUsers(ctx context.Context, limit, offset int) ([]user.User, int, error) {
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return u.userRepo.List(ctx, limit, offset)
}

// Stats
func (u *Usecase) GetStats(ctx context.Context) (*port.AdminStats, error) {
	return u.adminRepo.GetStats(ctx)
}

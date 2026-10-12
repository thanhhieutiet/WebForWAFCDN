package order

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"labshop/internal/domain/errs"
	"labshop/internal/domain/money"
	domainorder "labshop/internal/domain/order"
	"labshop/internal/usecase/port"
)

type CreateOrderInput struct {
	RecipientName   string
	RecipientPhone  string
	ShippingAddress string
	PaymentMethod   string
}

type Usecase struct {
	orderRepo   port.OrderRepository
	cartRepo    port.CartRepository
	catalogRepo port.CatalogRepository
	couponRepo  port.CouponRepository
	paymentGw   port.PaymentGateway
	txManager   port.TxManager
	idGen       port.IDGenerator
}

func New(
	orderRepo port.OrderRepository,
	cartRepo port.CartRepository,
	catalogRepo port.CatalogRepository,
	couponRepo port.CouponRepository,
	paymentGw port.PaymentGateway,
	txManager port.TxManager,
	idGen port.IDGenerator,
) *Usecase {
	return &Usecase{
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		catalogRepo: catalogRepo,
		couponRepo:  couponRepo,
		paymentGw:   paymentGw,
		txManager:   txManager,
		idGen:       idGen,
	}
}

func computePayloadHash(in CreateOrderInput) string {
	raw := fmt.Sprintf("%s|%s|%s|%s",
		strings.TrimSpace(in.RecipientName),
		strings.TrimSpace(in.RecipientPhone),
		strings.TrimSpace(in.ShippingAddress),
		strings.TrimSpace(in.PaymentMethod),
	)
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func (u *Usecase) CreateOrder(ctx context.Context, userID, idempotencyKey string, in CreateOrderInput) (*domainorder.Order, error) {
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	payloadHash := computePayloadHash(in)

	if idempotencyKey != "" {
		existing, err := u.orderRepo.GetByIdempotencyKey(ctx, userID, idempotencyKey)
		if err == nil && existing != nil {
			if existing.PayloadHash == payloadHash {
				return existing, nil
			}
			return nil, errs.NewBusinessRule("idempotency_conflict", "idempotency key reused with different request payload")
		}
	}

	cart, err := u.cartRepo.GetByUserID(ctx, userID)
	if err != nil || cart == nil || len(cart.Items) == 0 {
		return nil, errs.NewValidation("cart is empty", errs.FieldViolation{Field: "cart", Issue: "cannot create order with empty cart"})
	}

	var createdOrder *domainorder.Order

	txErr := u.txManager.WithTx(ctx, func(txCtx context.Context) error {
		var (
			orderItems []domainorder.OrderItem
			subtotal   money.VND
		)

		for _, item := range cart.Items {
			ok, err := u.catalogRepo.DeductStockIfAvailable(txCtx, item.ProductID, item.Quantity)
			if err != nil {
				return err
			}
			if !ok {
				return errs.NewConflict(fmt.Sprintf("insufficient stock for product %s", item.ProductID))
			}

			prod, err := u.catalogRepo.GetProductByID(txCtx, item.ProductID)
			if err != nil || prod == nil {
				return errs.NewNotFound("product", item.ProductID)
			}

			lineSubtotal := prod.Price.Mul(int64(item.Quantity))
			subtotal = subtotal.Add(lineSubtotal)

			orderItems = append(orderItems, domainorder.OrderItem{
				ID:           u.idGen.NewID(),
				ProductID:    prod.ID,
				ProductName:  prod.Name,
				ProductPrice: prod.Price,
				Quantity:     item.Quantity,
				Subtotal:     lineSubtotal,
			})
		}

		var (
			discount money.VND
			freeShip bool
		)

		if cart.CouponCode != "" {
			coup, err := u.couponRepo.GetByCode(txCtx, cart.CouponCode)
			if err != nil || coup == nil {
				return errs.NewBusinessRule("invalid_coupon", "applied coupon not found")
			}
			d, fs, errDisc := coup.CalculateDiscount(subtotal, time.Now())
			if errDisc != nil {
				return errDisc
			}
			discount = d
			freeShip = fs
			if err := u.couponRepo.IncrementUsage(txCtx, coup.Code); err != nil {
				return err
			}
		}

		shippingFee := domainorder.CalculateShipping(subtotal, freeShip)
		total := subtotal.Sub(discount).Add(shippingFee)
		if total < 0 {
			total = 0
		}

		now := time.Now()
		ord := &domainorder.Order{
			ID:              u.idGen.NewID(),
			UserID:          userID,
			Items:           orderItems,
			Subtotal:        subtotal,
			Discount:        discount,
			ShippingFee:     shippingFee,
			Total:           total,
			CouponCode:      cart.CouponCode,
			Status:          domainorder.StatusPending,
			PaymentMethod:   in.PaymentMethod,
			RecipientName:   in.RecipientName,
			RecipientPhone:  in.RecipientPhone,
			ShippingAddress: in.ShippingAddress,
			IdempotencyKey:  idempotencyKey,
			PayloadHash:     payloadHash,
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		for i := range ord.Items {
			ord.Items[i].OrderID = ord.ID
		}

		if err := ord.ValidateInvariants(); err != nil {
			return err
		}

		if err := u.orderRepo.Create(txCtx, ord); err != nil {
			return err
		}

		if err := u.cartRepo.DeleteByUserID(txCtx, userID); err != nil {
			return err
		}

		createdOrder = ord
		return nil
	})

	if txErr != nil {
		return nil, txErr
	}

	return createdOrder, nil
}

func (u *Usecase) PayOrder(ctx context.Context, userID, orderID string) (*domainorder.Order, error) {
	ord, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil || ord == nil {
		return nil, errs.NewNotFound("order", orderID)
	}

	if ord.UserID != userID {
		return nil, errs.NewNotFound("order", orderID)
	}

	if err := ord.TransitionTo(domainorder.StatusPaid); err != nil {
		return nil, err
	}

	var paymentRef string
	if ord.PaymentMethod == domainorder.PaymentMethodMockGateway {
		ref, err := u.paymentGw.Charge(ctx, ord.ID, ord.Total)
		if err != nil {
			return nil, errs.NewBusinessRule("payment_failed", "payment failed: "+err.Error())
		}
		paymentRef = ref
		ord.PaymentRef = ref
	}

	if err := u.orderRepo.UpdateStatus(ctx, ord.ID, ord.Status, paymentRef); err != nil {
		return nil, err
	}

	return ord, nil
}

func (u *Usecase) CancelOrder(ctx context.Context, userID, orderID string) (*domainorder.Order, error) {
	ord, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil || ord == nil {
		return nil, errs.NewNotFound("order", orderID)
	}

	if ord.UserID != userID {
		return nil, errs.NewNotFound("order", orderID)
	}

	if err := ord.TransitionTo(domainorder.StatusCancelled); err != nil {
		return nil, err
	}

	err = u.txManager.WithTx(ctx, func(txCtx context.Context) error {
		for _, item := range ord.Items {
			if err := u.catalogRepo.RestoreStock(txCtx, item.ProductID, item.Quantity); err != nil {
				return err
			}
		}
		return u.orderRepo.UpdateStatus(txCtx, ord.ID, ord.Status, "")
	})
	if err != nil {
		return nil, err
	}

	return ord, nil
}

func (u *Usecase) GetOrder(ctx context.Context, userID, orderID string, isAdmin bool) (*domainorder.Order, error) {
	ord, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil || ord == nil {
		return nil, errs.NewNotFound("order", orderID)
	}

	if !isAdmin && ord.UserID != userID {
		return nil, errs.NewNotFound("order", orderID)
	}

	return ord, nil
}

func (u *Usecase) ListOrders(ctx context.Context, userID string, limit, offset int) ([]domainorder.Order, int, error) {
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return u.orderRepo.ListByUserID(ctx, userID, limit, offset)
}

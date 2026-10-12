package cart

import (
	"context"
	"strings"
	"time"

	domaincart "labshop/internal/domain/cart"
	"labshop/internal/domain/errs"
	"labshop/internal/domain/money"
	"labshop/internal/domain/order"
	"labshop/internal/usecase/port"
)

type CartItemView struct {
	ProductID string
	Name      string
	Slug      string
	ImagePath string
	Price     money.VND
	Quantity  int
	Subtotal  money.VND
	Stock     int
}

type CartView struct {
	UserID      string
	Items       []CartItemView
	Subtotal    money.VND
	Discount    money.VND
	ShippingFee money.VND
	Total       money.VND
	CouponCode  string
	ItemCount   int
}

type Usecase struct {
	cartRepo    port.CartRepository
	catalogRepo port.CatalogRepository
	couponRepo  port.CouponRepository
}

func New(cartRepo port.CartRepository, catalogRepo port.CatalogRepository, couponRepo port.CouponRepository) *Usecase {
	return &Usecase{
		cartRepo:    cartRepo,
		catalogRepo: catalogRepo,
		couponRepo:  couponRepo,
	}
}

func (u *Usecase) getOrCreateCart(ctx context.Context, userID string) (*domaincart.Cart, error) {
	c, err := u.cartRepo.GetByUserID(ctx, userID)
	if err != nil || c == nil {
		c = domaincart.NewCart(userID)
		_ = u.cartRepo.Save(ctx, c)
	}
	return c, nil
}

func (u *Usecase) buildCartView(ctx context.Context, c *domaincart.Cart) (*CartView, error) {
	now := time.Now()
	var (
		itemViews []CartItemView
		subtotal  money.VND
		totalQty  int
	)

	for _, item := range c.Items {
		prod, err := u.catalogRepo.GetProductByID(ctx, item.ProductID)
		if err != nil || prod == nil {
			continue
		}
		lineSubtotal := prod.Price.Mul(int64(item.Quantity))
		subtotal = subtotal.Add(lineSubtotal)
		totalQty += item.Quantity

		itemViews = append(itemViews, CartItemView{
			ProductID: prod.ID,
			Name:      prod.Name,
			Slug:      prod.Slug,
			ImagePath: prod.ImagePath,
			Price:     prod.Price,
			Quantity:  item.Quantity,
			Subtotal:  lineSubtotal,
			Stock:     prod.Stock,
		})
	}

	var (
		discount money.VND
		freeShip bool
	)

	if c.CouponCode != "" {
		coup, err := u.couponRepo.GetByCode(ctx, c.CouponCode)
		if err == nil && coup != nil {
			d, fs, errDisc := coup.CalculateDiscount(subtotal, now)
			if errDisc == nil {
				discount = d
				freeShip = fs
			}
		}
	}

	shippingFee := order.CalculateShipping(subtotal, freeShip)
	total := subtotal.Sub(discount).Add(shippingFee)
	if total < 0 {
		total = 0
	}

	return &CartView{
		UserID:      c.UserID,
		Items:       itemViews,
		Subtotal:    subtotal,
		Discount:    discount,
		ShippingFee: shippingFee,
		Total:       total,
		CouponCode:  c.CouponCode,
		ItemCount:   totalQty,
	}, nil
}

func (u *Usecase) GetCart(ctx context.Context, userID string) (*CartView, error) {
	c, err := u.getOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}
	return u.buildCartView(ctx, c)
}

func (u *Usecase) UpdateItemQuantity(ctx context.Context, userID, productID string, quantity int) (*CartView, error) {
	c, err := u.getOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	if quantity == 0 {
		c.RemoveItem(productID)
		if err := u.cartRepo.Save(ctx, c); err != nil {
			return nil, err
		}
		return u.buildCartView(ctx, c)
	}

	prod, err := u.catalogRepo.GetProductByID(ctx, productID)
	if err != nil || prod == nil {
		return nil, errs.NewNotFound("product", productID)
	}

	if err := c.SetItem(prod, quantity); err != nil {
		return nil, err
	}

	if err := u.cartRepo.Save(ctx, c); err != nil {
		return nil, err
	}

	return u.buildCartView(ctx, c)
}

func (u *Usecase) RemoveItem(ctx context.Context, userID, productID string) (*CartView, error) {
	c, err := u.getOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	c.RemoveItem(productID)
	if err := u.cartRepo.Save(ctx, c); err != nil {
		return nil, err
	}

	return u.buildCartView(ctx, c)
}

func (u *Usecase) ClearCart(ctx context.Context, userID string) error {
	return u.cartRepo.DeleteByUserID(ctx, userID)
}

func (u *Usecase) ApplyCoupon(ctx context.Context, userID, couponCode string) (*CartView, error) {
	couponCode = strings.TrimSpace(strings.ToUpper(couponCode))
	if couponCode == "" {
		return nil, errs.NewValidation("coupon code is required", errs.FieldViolation{Field: "code", Issue: "cannot be empty"})
	}

	coup, err := u.couponRepo.GetByCode(ctx, couponCode)
	if err != nil || coup == nil {
		return nil, errs.NewNotFound("coupon", couponCode)
	}

	c, err := u.getOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Calculate subtotal to validate coupon against subtotal
	var subtotal money.VND
	for _, item := range c.Items {
		prod, err := u.catalogRepo.GetProductByID(ctx, item.ProductID)
		if err == nil && prod != nil {
			subtotal = subtotal.Add(prod.Price.Mul(int64(item.Quantity)))
		}
	}

	if _, _, err := coup.CalculateDiscount(subtotal, time.Now()); err != nil {
		return nil, err
	}

	c.ApplyCoupon(couponCode)
	if err := u.cartRepo.Save(ctx, c); err != nil {
		return nil, err
	}

	return u.buildCartView(ctx, c)
}

func (u *Usecase) RemoveCoupon(ctx context.Context, userID string) (*CartView, error) {
	c, err := u.getOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	c.RemoveCoupon()
	if err := u.cartRepo.Save(ctx, c); err != nil {
		return nil, err
	}

	return u.buildCartView(ctx, c)
}

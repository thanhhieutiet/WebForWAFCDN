package db

import (
	"context"
	"fmt"
	"time"

	"labshop/internal/adapter/repository/sqlite"
	"labshop/internal/domain/catalog"
	"labshop/internal/domain/coupon"
	"labshop/internal/domain/money"
	"labshop/internal/domain/user"
	"labshop/internal/usecase/port"
)

// Seed populates the database with initial catalog, coupons, and demo accounts if products table is empty.
func Seed(
	ctx context.Context,
	database *sqlite.DB,
	hasher port.PasswordHasher,
	adminEmail, adminPassword, customerEmail, customerPassword string,
) error {
	var count int
	err := database.SQLDB().QueryRowContext(ctx, "SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing products count: %w", err)
	}
	if count > 0 {
		return nil // Already seeded
	}

	userRepo := sqlite.NewUserRepository(database)
	catRepo := sqlite.NewCatalogRepository(database)
	couponRepo := sqlite.NewCouponRepository(database)

	now := time.Now().UTC()

	// 1. Seed Categories (5)
	categories := []catalog.Category{
		{ID: "cat-smartphones", Name: "Smartphones", Slug: "smartphones", Description: "Flagship and mid-range mobile devices", CreatedAt: now},
		{ID: "cat-laptops", Name: "Laptops", Slug: "laptops", Description: "Ultrabooks, gaming and productivity laptops", CreatedAt: now},
		{ID: "cat-audio", Name: "Audio", Slug: "audio", Description: "Wireless headphones, earphones, and speakers", CreatedAt: now},
		{ID: "cat-accessories", Name: "Accessories", Slug: "accessories", Description: "Keyboards, mice, and desk accessories", CreatedAt: now},
		{ID: "cat-smartwatches", Name: "Smartwatches", Slug: "smartwatches", Description: "Fitness trackers and luxury smartwatches", CreatedAt: now},
	}
	for _, c := range categories {
		if err := catRepo.CreateCategory(ctx, &c); err != nil {
			return fmt.Errorf("failed to seed category %s: %w", c.Name, err)
		}
	}

	// 2. Seed 40 Products (8 per category)
	type seedProduct struct {
		name       string
		slug       string
		categoryID string
		price      int64
		stock      int
		desc       string
	}

	rawProducts := []seedProduct{
		// Smartphones (1-8)
		{"iPhone 15 Pro Max 256GB", "iphone-15-pro-max-256gb", "cat-smartphones", 31990000, 25, "Titanium design, A17 Pro chip, 48MP camera."},
		{"Samsung Galaxy S24 Ultra", "samsung-galaxy-s24-ultra", "cat-smartphones", 29990000, 20, "Galaxy AI, 200MP camera, built-in S Pen."},
		{"Xiaomi 14 Ultra 512GB", "xiaomi-14-ultra-512gb", "cat-smartphones", 27490000, 15, "Leica quad camera system with 1-inch sensor."},
		{"Google Pixel 8 Pro", "google-pixel-8-pro", "cat-smartphones", 22990000, 18, "Google Tensor G3, best-in-class computational photography."},
		{"OnePlus 12 512GB", "oneplus-12-512gb", "cat-smartphones", 19990000, 30, "Snapdragon 8 Gen 3, 5400mAh battery with 100W SuperVOOC."},
		{"Sony Xperia 1 VI", "sony-xperia-1-vi", "cat-smartphones", 28990000, 10, "Continuous optical telephoto zoom, BRAVIA display."},
		{"Asus ROG Phone 8 Pro", "asus-rog-phone-8-pro", "cat-smartphones", 26990000, 12, "Extreme gaming performance with AirTrigger controls."},
		{"Nothing Phone (2) 256GB", "nothing-phone-2-256gb", "cat-smartphones", 14990000, 22, "Unique Glyph Interface, transparent aesthetic, smooth OS."},

		// Laptops (9-16)
		{"MacBook Pro 16 M3 Max", "macbook-pro-16-m3-max", "cat-laptops", 79990000, 10, "M3 Max chip, Liquid Retina XDR, up to 22 hours battery."},
		{"MacBook Air 13 M3", "macbook-air-13-m3", "cat-laptops", 27990000, 35, "Thin and light powerhouse with silent fanless design."},
		{"Dell XPS 15 OLED", "dell-xps-15-oled", "cat-laptops", 45990000, 14, "Intel Core i9, RTX 4070, stunning 3.5K OLED touchscreen."},
		{"Lenovo ThinkPad X1 Carbon Gen 12", "thinkpad-x1-carbon-gen-12", "cat-laptops", 42990000, 16, "Ultralight business laptop with legendary keyboard."},
		{"ASUS ROG Zephyrus G16", "asus-rog-zephyrus-g16", "cat-laptops", 52990000, 12, "OLED 240Hz, Intel Core Ultra 9, RTX 4080 GPU."},
		{"Razer Blade 16", "razer-blade-16", "cat-laptops", 64990000, 8, "Dual-mode Mini-LED display, anodized aluminum chassis."},
		{"HP Spectre x360 14", "hp-spectre-x360-14", "cat-laptops", 36990000, 18, "Convertible 2-in-1 laptop with 2.8K OLED display."},
		{"LG gram 17 Ultra-Light", "lg-gram-17-ultra-light", "cat-laptops", 34990000, 15, "17-inch screen weighing only 1.35kg, huge 80Wh battery."},

		// Audio (17-24)
		{"Sony WH-1000XM5", "sony-wh-1000xm5", "cat-audio", 7990000, 40, "Industry-leading noise cancellation with 30-hour battery."},
		{"Apple AirPods Max", "apple-airpods-max", "cat-audio", 12990000, 25, "Computational audio with Active Noise Cancellation and Spatial Audio."},
		{"Bose QuietComfort Ultra", "bose-quietcomfort-ultra", "cat-audio", 8990000, 30, "Breakthrough spatialized audio with world-class quiet."},
		{"Sennheiser Momentum 4 Wireless", "sennheiser-momentum-4", "cat-audio", 8490000, 28, "Audiophile sound quality with colossal 60-hour battery life."},
		{"Apple AirPods Pro 2 USB-C", "apple-airpods-pro-2-usbc", "cat-audio", 5890000, 50, "Up to 2x more Active Noise Cancellation, Adaptive Audio."},
		{"Sony WF-1000XM5 Earbuds", "sony-wf-1000xm5-earbuds", "cat-audio", 6290000, 45, "High-resolution wireless audio with bone conduction mic."},
		{"B&O Beosound A1 2nd Gen", "bo-beosound-a1-2nd-gen", "cat-audio", 6990000, 20, "Waterproof Bluetooth speaker with punchy True360 sound."},
		{"Marshall Stanmore III Speaker", "marshall-stanmore-iii", "cat-audio", 9490000, 18, "Iconic rock 'n' roll design with re-engineered wider soundstage."},

		// Accessories (25-32)
		{"Keychron Q1 Pro Wireless Keyboard", "keychron-q1-pro-wireless", "cat-accessories", 4490000, 30, "Full aluminum 75% mechanical keyboard with hot-swap switches."},
		{"Logitech MX Master 3S Mouse", "logitech-mx-master-3s", "cat-accessories", 2190000, 50, "Quiet clicks, 8K DPI sensor on glass, ergonomic thumb rest."},
		{"Apple Magic Trackpad Black", "apple-magic-trackpad-black", "cat-accessories", 3590000, 25, "Wireless and rechargeable Multi-Touch and Force Touch surface."},
		{"Anker Prime 20000mAh Power Bank", "anker-prime-20000mah-200w", "cat-accessories", 2890000, 35, "200W total output, smart digital display, multi-device charging."},
		{"NuPhy Air75 V2 Low-Profile Keyboard", "nuphy-air75-v2", "cat-accessories", 2990000, 28, "Ultra-slim wireless mechanical keyboard with 1000Hz polling rate."},
		{"BenQ ScreenBar Halo Light", "benq-screenbar-halo", "cat-accessories", 4290000, 22, "Monitor light bar with wireless controller and rear bias lighting."},
		{"CalDigit TS4 Thunderbolt 4 Dock", "caldigit-ts4-thunderbolt-4", "cat-accessories", 9990000, 15, "18 ports of expansion, 98W charging, dual 6K display support."},
		{"Samsung T9 Shield 2TB Portable SSD", "samsung-t9-shield-2tb-ssd", "cat-accessories", 4990000, 40, "USB 3.2 Gen 2x2 speeds up to 2000MB/s, drop resistant."},

		// Smartwatches (33-40)
		{"Apple Watch Ultra 2", "apple-watch-ultra-2", "cat-smartwatches", 20990000, 20, "Rugged titanium case, precision dual-frequency GPS, 3000 nits display."},
		{"Apple Watch Series 9 45mm", "apple-watch-series-9-45mm", "cat-smartwatches", 10990000, 30, "S9 SiP chip, double tap gesture, edge-to-edge Retina display."},
		{"Samsung Galaxy Watch Ultra", "samsung-galaxy-watch-ultra", "cat-smartwatches", 16990000, 18, "Cushion design, titanium grade 4, dual-frequency GPS."},
		{"Garmin Fenix 7 Pro Sapphire Solar", "garmin-fenix-7-pro-sapphire", "cat-smartwatches", 22490000, 15, "Multisport GPS watch with solar charging lens and LED flashlight."},
		{"Garmin Forerunner 965", "garmin-forerunner-965", "cat-smartwatches", 16490000, 16, "AMOLED running smartwatch with training readiness metrics."},
		{"Google Pixel Watch 2", "google-pixel-watch-2", "cat-smartwatches", 8990000, 22, "Advanced heart rate tracking, Fitbit health insights, Wear OS."},
		{"Huawei Watch Ultimate Expedition", "huawei-watch-ultimate", "cat-smartwatches", 18990000, 12, "Zirconium-based liquid metal case, 100-meter diving mode."},
		{"Coros Pace 3 White", "coros-pace-3-white", "cat-smartwatches", 5990000, 25, "Ultra-lightweight 30g GPS sports watch with 38h battery life."},
	}

	for idx, rp := range rawProducts {
		imgIdx := fmt.Sprintf("%02d", idx+1)
		p := &catalog.Product{
			ID:          fmt.Sprintf("prod-%02d", idx+1),
			CategoryID:  rp.categoryID,
			Name:        rp.name,
			Slug:        rp.slug,
			Description: rp.desc,
			Price:       money.VND(rp.price),
			Stock:       rp.stock,
			ImagePath:   fmt.Sprintf("/static/img/products/p%s.jpg", imgIdx),
			Status:      catalog.StatusActive,
			RatingAvg:   0,
			RatingCount: 0,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := catRepo.CreateProduct(ctx, p); err != nil {
			return fmt.Errorf("failed to seed product %s: %w", p.Name, err)
		}
	}

	// 3. Seed 4 Coupons
	coupons := []coupon.Coupon{
		{
			ID:          "coup-welcome10",
			Code:        "WELCOME10",
			Type:        coupon.TypePercent,
			Value:       10,
			MaxDiscount: money.VND(100000),
			MinSubtotal: money.VND(200000),
			ValidFrom:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			ValidTo:     time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
			MaxUses:     1000,
			UsedCount:   0,
			CreatedAt:   now,
		},
		{
			ID:          "coup-freeship",
			Code:        "FREESHIP",
			Type:        coupon.TypeFreeShip,
			Value:       0,
			MaxDiscount: 0,
			MinSubtotal: 0,
			ValidFrom:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			ValidTo:     time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
			MaxUses:     1000,
			UsedCount:   0,
			CreatedAt:   now,
		},
		{
			ID:          "coup-expired5",
			Code:        "EXPIRED5",
			Type:        coupon.TypeFixed,
			Value:       50000,
			MaxDiscount: 0,
			MinSubtotal: money.VND(100000),
			ValidFrom:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			ValidTo:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			MaxUses:     100,
			UsedCount:   0,
			CreatedAt:   now,
		},
		{
			ID:          "coup-limit1",
			Code:        "LIMIT1",
			Type:        coupon.TypeFixed,
			Value:       100000,
			MaxDiscount: 0,
			MinSubtotal: money.VND(300000),
			ValidFrom:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			ValidTo:     time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
			MaxUses:     1,
			UsedCount:   0,
			CreatedAt:   now,
		},
	}
	for _, cp := range coupons {
		if err := couponRepo.Save(ctx, &cp); err != nil {
			return fmt.Errorf("failed to seed coupon %s: %w", cp.Code, err)
		}
	}

	// 4. Seed Admin User
	adminHash, err := hasher.Hash(adminPassword)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}
	adminUser := &user.User{
		ID:           "u-admin-01",
		Email:        adminEmail,
		PasswordHash: adminHash,
		Name:         "Administrator",
		Phone:        "0900000001",
		Address:      "LabShop Headquarters, Hanoi",
		Role:         user.RoleAdmin,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := userRepo.Create(ctx, adminUser); err != nil {
		return fmt.Errorf("failed to seed admin user: %w", err)
	}

	// 5. Seed Customer User
	custHash, err := hasher.Hash(customerPassword)
	if err != nil {
		return fmt.Errorf("failed to hash customer password: %w", err)
	}
	custUser := &user.User{
		ID:           "u-customer-01",
		Email:        customerEmail,
		PasswordHash: custHash,
		Name:         "Customer Demo",
		Phone:        "0900000002",
		Address:      "123 Hoang Quoc Viet, Cau Giay, Hanoi",
		Role:         user.RoleCustomer,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := userRepo.Create(ctx, custUser); err != nil {
		return fmt.Errorf("failed to seed customer user: %w", err)
	}

	return nil
}

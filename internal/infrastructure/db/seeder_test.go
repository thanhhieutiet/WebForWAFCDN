package db

import (
	"context"
	"path/filepath"
	"testing"

	"labshop/internal/adapter/repository/sqlite"
	"labshop/internal/adapter/security/argon2"
	"labshop/internal/usecase/port"
)

func TestSeeder(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_seed.db")

	database, err := Init(dbPath)
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer database.Close()

	hasher := argon2.New()
	ctx := context.Background()

	err = Seed(ctx, database, hasher, "admin@test.local", "Admin@Test2026!", "cust@test.local", "Cust@Test2026!")
	if err != nil {
		t.Fatalf("seeder failed: %v", err)
	}

	// Verify categories
	catRepo := sqlite.NewCatalogRepository(database)
	cats, err := catRepo.ListCategories(ctx)
	if err != nil || len(cats) != 5 {
		t.Fatalf("expected 5 categories, got %d (err: %v)", len(cats), err)
	}

	// Verify products
	prods, total, err := catRepo.ListProducts(ctx, port.ProductFilter{Page: 1, PageSize: 50})
	if err != nil || total != 40 || len(prods) != 40 {
		t.Fatalf("expected 40 products, got total=%d, len=%d (err: %v)", total, len(prods), err)
	}

	// Verify coupons
	couponRepo := sqlite.NewCouponRepository(database)
	coups, err := couponRepo.List(ctx)
	if err != nil || len(coups) != 4 {
		t.Fatalf("expected 4 coupons, got %d (err: %v)", len(coups), err)
	}

	// Verify users
	userRepo := sqlite.NewUserRepository(database)
	admin, err := userRepo.GetByEmail(ctx, "admin@test.local")
	if err != nil || !admin.IsAdmin() {
		t.Fatalf("expected admin user, got %v (err: %v)", admin, err)
	}

	cust, err := userRepo.GetByEmail(ctx, "cust@test.local")
	if err != nil || cust.IsAdmin() {
		t.Fatalf("expected customer user, got %v (err: %v)", cust, err)
	}

	// Running seed again is idempotent (does nothing)
	err = Seed(ctx, database, hasher, "admin@test.local", "Admin@Test2026!", "cust@test.local", "Cust@Test2026!")
	if err != nil {
		t.Fatalf("second seed failed: %v", err)
	}
}

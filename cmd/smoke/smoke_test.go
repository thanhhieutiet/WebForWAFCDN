package main

import (
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"labshop/internal/adapter/counter/memory"
	httpadapter "labshop/internal/adapter/http"
	"labshop/internal/adapter/http/handler"
	mockpayment "labshop/internal/adapter/payment/mock"
	"labshop/internal/adapter/repository/sqlite"
	"labshop/internal/adapter/scenario/embedded"
	"labshop/internal/adapter/security/argon2"
	"labshop/internal/adapter/security/jwt"
	"labshop/internal/infrastructure/config"
	infraDB "labshop/internal/infrastructure/db"
	adminuc "labshop/internal/usecase/admin"
	authuc "labshop/internal/usecase/auth"
	cartuc "labshop/internal/usecase/cart"
	cataloguc "labshop/internal/usecase/catalog"
	"labshop/internal/usecase/lab"
	orderuc "labshop/internal/usecase/order"
	"labshop/internal/usecase/port"
	reviewuc "labshop/internal/usecase/review"
	supportuc "labshop/internal/usecase/support"
	tutorialuc "labshop/internal/usecase/tutorial"
)

type testUUIDGen struct{}

func (testUUIDGen) NewID() string {
	return uuid.NewString()
}

func TestSmokeSuite(t *testing.T) {
	// Find project root
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find module root")
		}
		dir = parent
	}

	tempDB := filepath.Join(t.TempDir(), "smoke_test.db")
	db, err := infraDB.Init(tempDB)
	require.NoError(t, err)
	defer db.Close()

	hasher := argon2.New()
	adminEmail := "admin@labshop.local"
	adminPass := "Admin@LabShop2026!"
	customerEmail := "customer@labshop.local"
	customerPass := "Customer@LabShop2026!"

	err = infraDB.Seed(context.Background(), db, hasher, adminEmail, adminPass, customerEmail, customerPass)
	require.NoError(t, err)

	idGen := testUUIDGen{}
	jwtSecret := "smoke-secret-key-must-be-32-chars-long-1234"
	jwtIssuer := jwt.New(jwtSecret)
	paymentGw := mockpayment.New(0)
	txManager := sqlite.NewTxManager(db)

	userRepo := sqlite.NewUserRepository(db)
	catalogRepo := sqlite.NewCatalogRepository(db)
	cartRepo := sqlite.NewCartRepository(db)
	couponRepo := sqlite.NewCouponRepository(db)
	orderRepo := sqlite.NewOrderRepository(db)
	reviewRepo := sqlite.NewReviewRepository(db)
	supportRepo := sqlite.NewSupportRepository(db)
	adminRepo := sqlite.NewAdminRepository(db)

	authUC := authuc.New(userRepo, hasher, jwtIssuer, idGen)
	catalogUC := cataloguc.New(catalogRepo)
	cartUC := cartuc.New(cartRepo, catalogRepo, couponRepo)
	orderUC := orderuc.New(orderRepo, cartRepo, catalogRepo, couponRepo, paymentGw, txManager, idGen)
	reviewUC := reviewuc.New(reviewRepo, orderRepo, catalogRepo, idGen)
	supportUC := supportuc.New(supportRepo, idGen)
	adminUC := adminuc.New(catalogRepo, orderRepo, supportRepo, userRepo, adminRepo, txManager, idGen)
	scenarioRepo := embedded.MustNewRepository()
	tutorialUC := tutorialuc.New(scenarioRepo)

	businessHandlers := &httpadapter.BusinessHandlers{
		Auth:        handler.NewAuthHandler(authUC),
		Catalog:     handler.NewCatalogHandler(catalogUC),
		Cart:        handler.NewCartHandler(cartUC),
		Order:       handler.NewOrderHandler(orderUC),
		Review:      reviewHandler(reviewUC),
		Support:     handler.NewSupportHandler(supportUC),
		Admin:       handler.NewAdminHandler(adminUC),
		Tutorial:    handler.NewTutorialHandler(tutorialUC, 60),
		TokenIssuer: jwtIssuer,
	}

	counter := memory.New()
	labUC := lab.New(counter, port.RealClock{}, lab.Config{
		AdminToken:   "admin-token",
		MaxSlowMS:    60000,
		MaxPayloadKB: 5120,
		MaxUploadMB:  10,
	})
	labHandler := handler.NewLabHandler(labUC, nil)

	cfg := &config.Config{
		InstanceID:          "smoke-origin",
		ServeStatic:         true,
		StaticDir:           filepath.Join(dir, "web", "static"),
		CatalogCacheSeconds: 60,
	}

	router := httpadapter.NewRouter(cfg, nil, labHandler, counter, businessHandlers)
	server := httptest.NewServer(router)
	defer server.Close()

	client := &SmokeClient{
		baseURL:    server.URL,
		httpClient: server.Client(),
		adminEmail: adminEmail,
		adminPass:  adminPass,
	}

	results, passed := client.RunAll()
	printSummaryTable(results)

	require.True(t, passed, "all 18 smoke tests should pass")
	require.Equal(t, 18, len(results), "expected 18 tests")
}

func reviewHandler(uc *reviewuc.Usecase) *handler.ReviewHandler {
	return handler.NewReviewHandler(uc)
}

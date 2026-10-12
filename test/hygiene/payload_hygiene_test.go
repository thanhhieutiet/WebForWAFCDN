package hygiene_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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
	"labshop/internal/infrastructure/logger"
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

//go:embed probes.json
var embeddedProbesJSON []byte

type probeCatalog struct {
	SQLiLike      []string `json:"sqli_like"`
	XSSLike       []string `json:"xss_like"`
	TraversalLike []string `json:"traversal_like"`
	CmdLike       []string `json:"cmd_like"`
	ScannerUA     []string `json:"scanner_ua"`
	FalsePositive []string `json:"false_positive"`
}

func loadProbes(t *testing.T) probeCatalog {
	t.Helper()
	var pc probeCatalog
	err := json.Unmarshal(embeddedProbesJSON, &pc)
	require.NoError(t, err, "failed to parse embedded probes.json")
	require.NotEmpty(t, pc.SQLiLike, "sqli_like probes should not be empty")
	require.NotEmpty(t, pc.XSSLike, "xss_like probes should not be empty")
	require.NotEmpty(t, pc.TraversalLike, "traversal_like probes should not be empty")
	require.NotEmpty(t, pc.CmdLike, "cmd_like probes should not be empty")
	require.NotEmpty(t, pc.ScannerUA, "scanner_ua probes should not be empty")
	require.NotEmpty(t, pc.FalsePositive, "false_positive probes should not be empty")
	return pc
}

func (pc probeCatalog) allProbes() []string {
	var all []string
	all = append(all, pc.SQLiLike...)
	all = append(all, pc.XSSLike...)
	all = append(all, pc.TraversalLike...)
	all = append(all, pc.CmdLike...)
	all = append(all, pc.ScannerUA...)
	all = append(all, pc.FalsePositive...)
	return all
}

type uuidGen struct{}

func (uuidGen) NewID() string { return uuid.NewString() }

type hygieneTestApp struct {
	router      *chi.Mux
	db          *sqlite.DB
	catalogRepo port.CatalogRepository
	adminToken  string
	custToken   string
	custID      string
	adminID     string
}

func setupHygieneApp(t *testing.T) *hygieneTestApp {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "hygiene_labshop.db")

	database, err := infraDB.Init(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	hasher := argon2.New()
	jwtSecret := "payload-hygiene-test-secret-at-least-32-chars!"
	jwtIssuer := jwt.New(jwtSecret, 2*time.Hour)
	paymentGw := mockpayment.New(0.0)
	txManager := sqlite.NewTxManager(database)
	idGen := uuidGen{}

	adminEmail := "admin@labshop.local"
	custEmail := "customer@labshop.local"
	err = infraDB.Seed(
		context.Background(),
		database,
		hasher,
		adminEmail,
		"Admin@LabShop2026!",
		custEmail,
		"Customer@LabShop2026!",
	)
	require.NoError(t, err)

	userRepo := sqlite.NewUserRepository(database)
	catalogRepo := sqlite.NewCatalogRepository(database)
	cartRepo := sqlite.NewCartRepository(database)
	couponRepo := sqlite.NewCouponRepository(database)
	orderRepo := sqlite.NewOrderRepository(database)
	reviewRepo := sqlite.NewReviewRepository(database)
	supportRepo := sqlite.NewSupportRepository(database)
	adminRepo := sqlite.NewAdminRepository(database)

	authUC := authuc.New(userRepo, hasher, jwtIssuer, idGen)
	catalogUC := cataloguc.New(catalogRepo)
	cartUC := cartuc.New(cartRepo, catalogRepo, couponRepo)
	orderUC := orderuc.New(orderRepo, cartRepo, catalogRepo, couponRepo, paymentGw, txManager, idGen)
	reviewUC := reviewuc.New(reviewRepo, orderRepo, catalogRepo, idGen)
	supportUC := supportuc.New(supportRepo, idGen)
	adminUC := adminuc.New(catalogRepo, orderRepo, supportRepo, userRepo, adminRepo, txManager, idGen)
	scenarioRepo := embedded.MustNewRepository()
	tutorialUC := tutorialuc.New(scenarioRepo)

	authHandler := handler.NewAuthHandler(authUC)
	catalogHandler := handler.NewCatalogHandler(catalogUC)
	cartHandler := handler.NewCartHandler(cartUC)
	orderHandler := handler.NewOrderHandler(orderUC)
	reviewHandler := handler.NewReviewHandler(reviewUC)
	supportHandler := handler.NewSupportHandler(supportUC)
	adminHandler := handler.NewAdminHandler(adminUC)
	tutorialHandler := handler.NewTutorialHandler(tutorialUC, 60)

	businessHandlers := &httpadapter.BusinessHandlers{
		Auth:        authHandler,
		Catalog:     catalogHandler,
		Cart:        cartHandler,
		Order:       orderHandler,
		Review:      reviewHandler,
		Support:     supportHandler,
		Admin:       adminHandler,
		Tutorial:    tutorialHandler,
		TokenIssuer: jwtIssuer,
	}

	hitCounter := memory.New()
	labUC := lab.New(hitCounter, port.RealClock{}, lab.Config{
		AdminToken:   "hygiene-admin-token-secret",
		MaxSlowMS:    1000,
		MaxPayloadKB: 2048,
		MaxUploadMB:  10,
	})
	labHandler := handler.NewLabHandler(labUC, nil)

	cfg := &config.Config{
		InstanceID:          "origin-hygiene-node",
		CatalogCacheSeconds: 60,
		AdminToken:          "hygiene-admin-token-secret",
		MaxSlowMS:           1000,
		MaxPayloadKB:        2048,
		MaxUploadMB:         10,
		EnableWS:            false,
		ServeStatic:         false,
	}
	log := logger.New("error")
	router := httpadapter.NewRouter(cfg, log, labHandler, hitCounter, businessHandlers)

	adminObj, err := userRepo.GetByEmail(context.Background(), adminEmail)
	require.NoError(t, err)
	adminToken, err := jwtIssuer.IssueToken(adminObj)
	require.NoError(t, err)

	custObj, err := userRepo.GetByEmail(context.Background(), custEmail)
	require.NoError(t, err)
	custToken, err := jwtIssuer.IssueToken(custObj)
	require.NoError(t, err)

	return &hygieneTestApp{
		router:      router,
		db:          database,
		catalogRepo: catalogRepo,
		adminToken:  adminToken,
		custToken:   custToken,
		custID:      custObj.ID,
		adminID:     adminObj.ID,
	}
}

// assertHygiene enforces the 4 cardinal rules of origin payload hygiene:
// 1. NO 5xx Internal Server Errors under any circumstances.
// 2. NO SQL error messages or SQLite syntax errors exposed in responses.
// 3. NO raw unescaped input reflected in HTML.
// 4. If JSON is returned, the JSON structure remains valid and uncorrupted.
func assertHygiene(t *testing.T, rec *httptest.ResponseRecorder, method, path, probe string) {
	t.Helper()

	// Assertion 1: NO 5xx
	require.Lessf(t, rec.Code, 500,
		"HYGIENE VIOLATION (5xx): Status %d for %s %s with probe: %q\nResponse body:\n%s",
		rec.Code, method, path, probe, rec.Body.String(),
	)

	bodyStr := rec.Body.String()
	bodyLower := strings.ToLower(bodyStr)
	cType := strings.ToLower(rec.Header().Get("Content-Type"))

	// Assertion 2: NO SQL/SQLite errors leaked
	sqlErrorSignatures := []string{
		"syntax error",
		"near \"",
		"no such table",
		"no such column",
		"sql logic error",
		"unrecognized token",
		"database is locked",
		"unique constraint failed",
		"foreign key constraint failed",
		"sqlite_master",
		"sqlite_schema",
		"sqlite error",
		"sqlite3 error",
		"table not found",
		"column not found",
	}
	for _, sig := range sqlErrorSignatures {
		assert.NotContainsf(t, bodyLower, sig,
			"HYGIENE VIOLATION (SQL Error Leak): Found signature %q in response for %s %s with probe %q\nBody: %s",
			sig, method, path, probe, bodyStr,
		)
	}

	// Assertion 3: NO raw unescaped input in HTML
	if strings.Contains(cType, "text/html") {
		assert.NotContains(t, bodyStr, "<script>", "HYGIENE VIOLATION (XSS): Raw <script> tag in HTML response")
		assert.NotContains(t, bodyStr, "onerror=", "HYGIENE VIOLATION (XSS): Raw onerror attribute in HTML response")
		assert.NotContains(t, bodyStr, "<img src=", "HYGIENE VIOLATION (XSS): Raw <img tag in HTML response")
	}

	// Assertion 4: Valid JSON when application/json
	if strings.Contains(cType, "application/json") && rec.Body.Len() > 0 {
		var js any
		err := json.Unmarshal(rec.Body.Bytes(), &js)
		assert.NoErrorf(t, err,
			"HYGIENE VIOLATION (Malformed JSON): Corrupted JSON for %s %s with probe %q: %v\nBody: %s",
			method, path, probe, err, bodyStr,
		)
	}
}

// TestPayloadHygiene_ProbesLoaded validates that probes.json exists and contains all required categories.
func TestPayloadHygiene_ProbesLoaded(t *testing.T) {
	pc := loadProbes(t)
	assert.GreaterOrEqual(t, len(pc.SQLiLike), 2)
	assert.GreaterOrEqual(t, len(pc.XSSLike), 2)
	assert.GreaterOrEqual(t, len(pc.TraversalLike), 2)
	assert.GreaterOrEqual(t, len(pc.CmdLike), 2)
	assert.GreaterOrEqual(t, len(pc.ScannerUA), 2)
	assert.GreaterOrEqual(t, len(pc.FalsePositive), 2)
}

// TestPayloadHygiene_QueryParameters tests all probes injected into URL query parameters.
func TestPayloadHygiene_QueryParameters(t *testing.T) {
	app := setupHygieneApp(t)
	pc := loadProbes(t)
	probes := pc.allProbes()

	queryEndpoints := []string{
		"/api/products?q=%s",
		"/api/products?category=%s",
		"/api/products?sort=%s",
		"/api/products?order=%s",
		"/api/products?page=%s",
		"/api/products?limit=%s",
		"/api/cache/public?v=%s",
		"/api/echo?text=%s",
		"/api/slow?delay_ms=%s",
		"/api/payload?size_kb=%s",
		"/api/categories?category=%s",
		"/api/headers?filter=%s",
	}

	for _, ep := range queryEndpoints {
		for _, probe := range probes {
			escaped := url.QueryEscape(probe)
			target := fmt.Sprintf(ep, escaped)
			req := httptest.NewRequest(http.MethodGet, target, nil)
			w := httptest.NewRecorder()

			app.router.ServeHTTP(w, req)
			assertHygiene(t, w, http.MethodGet, target, probe)
		}
	}
}

// TestPayloadHygiene_PathParameters tests all probes injected into URL path parameters.
func TestPayloadHygiene_PathParameters(t *testing.T) {
	app := setupHygieneApp(t)
	pc := loadProbes(t)
	probes := pc.allProbes()

	pathEndpoints := []struct {
		method  string
		format  string
		authTok string
	}{
		{http.MethodGet, "/api/products/%s", ""},
		{http.MethodGet, "/api/products/%s/reviews", ""},
		{http.MethodGet, "/api/echo/%s", ""},
		{http.MethodGet, "/api/cache/%s", ""},
		{http.MethodGet, "/api/status/%s", ""},
		{http.MethodGet, "/api/tutorial/scenarios/%s", ""},
		{http.MethodGet, "/api/orders/%s", app.custToken},
		{http.MethodGet, "/api/support/tickets/%s", app.custToken},
		{http.MethodDelete, "/api/cart/items/%s", app.custToken},
		{http.MethodPut, "/api/cart/items/%s", app.custToken},
	}

	for _, ep := range pathEndpoints {
		for _, probe := range probes {
			// Test URL Path escaped version
			escapedPath := url.PathEscape(probe)
			targetEscaped := fmt.Sprintf(ep.format, escapedPath)

			req := httptest.NewRequest(ep.method, targetEscaped, nil)
			if ep.authTok != "" {
				req.Header.Set("Authorization", "Bearer "+ep.authTok)
			}
			w := httptest.NewRecorder()
			app.router.ServeHTTP(w, req)
			assertHygiene(t, w, ep.method, targetEscaped, probe)

			// Also test raw probe in path when it is a valid URI (e.g. traversal payloads ../../etc/passwd)
			targetRaw := fmt.Sprintf(ep.format, probe)
			if reqRaw, err := http.NewRequest(ep.method, targetRaw, nil); err == nil {
				if ep.authTok != "" {
					reqRaw.Header.Set("Authorization", "Bearer "+ep.authTok)
				}
				wRaw := httptest.NewRecorder()
				app.router.ServeHTTP(wRaw, reqRaw)
				assertHygiene(t, wRaw, ep.method, targetRaw, probe)
			}
		}
	}
}

// TestPayloadHygiene_JSONBodies tests all probes injected into various JSON request payloads.
func TestPayloadHygiene_JSONBodies(t *testing.T) {
	app := setupHygieneApp(t)
	pc := loadProbes(t)
	probes := pc.allProbes()

	for _, probe := range probes {
		// 1. POST /api/auth/login
		{
			body, _ := json.Marshal(map[string]string{
				"email":    probe,
				"password": probe,
			})
			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			app.router.ServeHTTP(w, req)
			assertHygiene(t, w, http.MethodPost, "/api/auth/login", probe)
		}

		// 2. POST /api/auth/register
		{
			body, _ := json.Marshal(map[string]string{
				"name":     probe,
				"email":    "probe-" + uuid.NewString()[:8] + "@probe.local",
				"password": probe,
			})
			req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			app.router.ServeHTTP(w, req)
			assertHygiene(t, w, http.MethodPost, "/api/auth/register", probe)
		}

		// 3. POST /api/products/iphone-15-pro-max-256gb/reviews
		{
			body, _ := json.Marshal(map[string]any{
				"rating":  5,
				"comment": probe,
			})
			req := httptest.NewRequest(http.MethodPost, "/api/products/iphone-15-pro-max-256gb/reviews", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+app.custToken)
			w := httptest.NewRecorder()
			app.router.ServeHTTP(w, req)
			assertHygiene(t, w, http.MethodPost, "/api/products/iphone-15-pro-max-256gb/reviews", probe)
		}

		// 4. POST /api/support/tickets
		{
			body, _ := json.Marshal(map[string]string{
				"subject":         "Inquiry about " + probe,
				"initial_message": probe,
			})
			req := httptest.NewRequest(http.MethodPost, "/api/support/tickets", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+app.custToken)
			w := httptest.NewRecorder()
			app.router.ServeHTTP(w, req)
			assertHygiene(t, w, http.MethodPost, "/api/support/tickets", probe)
		}

		// 5. POST /api/cart/coupon
		{
			body, _ := json.Marshal(map[string]string{
				"code": probe,
			})
			req := httptest.NewRequest(http.MethodPost, "/api/cart/coupon", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+app.custToken)
			w := httptest.NewRecorder()
			app.router.ServeHTTP(w, req)
			assertHygiene(t, w, http.MethodPost, "/api/cart/coupon", probe)
		}

		// 6. POST /api/echo
		{
			body, _ := json.Marshal(map[string]any{
				"probe":   probe,
				"details": map[string]string{"text": probe},
			})
			req := httptest.NewRequest(http.MethodPost, "/api/echo", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			app.router.ServeHTTP(w, req)
			assertHygiene(t, w, http.MethodPost, "/api/echo", probe)
		}

		// 7. PATCH /api/auth/me
		{
			body, _ := json.Marshal(map[string]string{
				"name":    probe,
				"phone":   probe,
				"address": probe,
			})
			req := httptest.NewRequest(http.MethodPatch, "/api/auth/me", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+app.custToken)
			w := httptest.NewRecorder()
			app.router.ServeHTTP(w, req)
			assertHygiene(t, w, http.MethodPatch, "/api/auth/me", probe)
		}

		// 8. POST /api/admin/products
		{
			body, _ := json.Marshal(map[string]any{
				"category_id": "cat-smartphones",
				"name":        "Product " + probe,
				"slug":        "probe-slug-" + uuid.NewString()[:8],
				"description": probe,
				"price":       199000,
				"stock":       10,
				"image_path":  "/static/img/products/p01.jpg",
			})
			req := httptest.NewRequest(http.MethodPost, "/api/admin/products", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+app.adminToken)
			w := httptest.NewRecorder()
			app.router.ServeHTTP(w, req)
			assertHygiene(t, w, http.MethodPost, "/api/admin/products", probe)
		}

		// 9. PATCH /api/admin/orders/fake-order-id/status
		{
			body, _ := json.Marshal(map[string]string{
				"status": probe,
			})
			req := httptest.NewRequest(http.MethodPatch, "/api/admin/orders/fake-order-id/status", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+app.adminToken)
			w := httptest.NewRecorder()
			app.router.ServeHTTP(w, req)
			assertHygiene(t, w, http.MethodPatch, "/api/admin/orders/fake-order-id/status", probe)
		}
	}
}

// TestPayloadHygiene_MultipartUpload tests multipart file and form uploads with malicious filenames and fields.
func TestPayloadHygiene_MultipartUpload(t *testing.T) {
	app := setupHygieneApp(t)
	pc := loadProbes(t)
	probes := pc.allProbes()

	for _, probe := range probes {
		// Test A: Probe as uploaded filename and form field
		var b bytes.Buffer
		wWriter := multipart.NewWriter(&b)

		// Form field containing probe
		_ = wWriter.WriteField("description", probe)
		_ = wWriter.WriteField("category", probe)

		// File part with probe as filename
		part, err := wWriter.CreateFormFile("file", probe)
		require.NoError(t, err)
		_, _ = io.WriteString(part, "file content containing probe: "+probe)
		_ = wWriter.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/upload", &b)
		req.Header.Set("Content-Type", wWriter.FormDataContentType())
		w := httptest.NewRecorder()

		app.router.ServeHTTP(w, req)
		assertHygiene(t, w, http.MethodPost, "/api/upload", probe)

		// Test B: Probe as file Content-Type header in multipart part
		var b2 bytes.Buffer
		wWriter2 := multipart.NewWriter(&b2)
		partHeader := make(map[string][]string)
		partHeader["Content-Disposition"] = []string{`form-data; name="file"; filename="upload_probe.txt"`}
		partHeader["Content-Type"] = []string{probe}
		customPart, err := wWriter2.CreatePart(partHeader)
		require.NoError(t, err)
		_, _ = io.WriteString(customPart, "custom probe body")
		_ = wWriter2.Close()

		req2 := httptest.NewRequest(http.MethodPost, "/api/upload", &b2)
		req2.Header.Set("Content-Type", wWriter2.FormDataContentType())
		w2 := httptest.NewRecorder()

		app.router.ServeHTTP(w2, req2)
		assertHygiene(t, w2, http.MethodPost, "/api/upload", probe)
	}
}

// TestPayloadHygiene_HTTPHeaders tests all probes injected into HTTP headers.
func TestPayloadHygiene_HTTPHeaders(t *testing.T) {
	app := setupHygieneApp(t)
	pc := loadProbes(t)
	probes := pc.allProbes()

	endpoints := []string{
		"/api/headers",
		"/api/products",
		"/api/echo",
		"/api/health",
	}

	for _, ep := range endpoints {
		for _, probe := range probes {
			req := httptest.NewRequest(http.MethodGet, ep, nil)
			req.Header.Set("User-Agent", probe)
			req.Header.Set("X-Custom-Header", probe)
			req.Header.Set("Referer", probe)
			req.Header.Set("Accept", probe)
			req.Header.Set("X-Forwarded-For", probe)
			req.Header.Set("Authorization", "Bearer "+probe)

			w := httptest.NewRecorder()
			app.router.ServeHTTP(w, req)
			assertHygiene(t, w, http.MethodGet, ep, probe)
		}
	}
}

// TestPayloadHygiene_Cookies tests all probes injected into HTTP Cookies.
func TestPayloadHygiene_Cookies(t *testing.T) {
	app := setupHygieneApp(t)
	pc := loadProbes(t)
	probes := pc.allProbes()

	endpoints := []string{
		"/api/echo",
		"/api/health",
		"/api/products",
	}

	for _, ep := range endpoints {
		for _, probe := range probes {
			req := httptest.NewRequest(http.MethodGet, ep, nil)

			// Clean cookie value by replacing chars illegal in HTTP cookie syntax
			cleanCookieVal := strings.ReplaceAll(probe, ";", "%3B")
			cleanCookieVal = strings.ReplaceAll(cleanCookieVal, "\n", "")
			cleanCookieVal = strings.ReplaceAll(cleanCookieVal, "\r", "")

			req.AddCookie(&http.Cookie{Name: "session_id", Value: cleanCookieVal})
			req.AddCookie(&http.Cookie{Name: "auth_token", Value: cleanCookieVal})
			req.AddCookie(&http.Cookie{Name: "user_pref", Value: cleanCookieVal})

			w := httptest.NewRecorder()
			app.router.ServeHTTP(w, req)
			assertHygiene(t, w, http.MethodGet, ep, probe)
		}
	}
}

// TestPayloadHygiene_FalsePositives explicitly verifies that valid strings with SQL or HTML keywords
// are accepted cleanly and stored accurately without false rejection, 500 crashes, or SQL errors.
func TestPayloadHygiene_FalsePositives(t *testing.T) {
	app := setupHygieneApp(t)
	pc := loadProbes(t)

	fp1 := pc.FalsePositive[0] // "Please SELECT the best option FROM the list; my code: if (a < b) { return \"<div>\"; }"
	fp2 := pc.FalsePositive[1] // "O'Brien & Sons -- great service"

	// 1. Submit support ticket with fp1 in message
	{
		body, _ := json.Marshal(map[string]string{
			"subject":         "Legitimate Question about Syntax",
			"initial_message": fp1,
		})
		req := httptest.NewRequest(http.MethodPost, "/api/support/tickets", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+app.custToken)
		w := httptest.NewRecorder()
		app.router.ServeHTTP(w, req)

		assertHygiene(t, w, http.MethodPost, "/api/support/tickets", fp1)
		assert.Equal(t, http.StatusCreated, w.Code, "False positive ticket creation should succeed with 201")

		var createdTicket struct {
			ID string `json:"id"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &createdTicket)
		require.NoError(t, err)
		require.NotEmpty(t, createdTicket.ID)

		// Verify ticket detail retrieval and message fidelity
		getReq := httptest.NewRequest(http.MethodGet, "/api/support/tickets/"+createdTicket.ID, nil)
		getReq.Header.Set("Authorization", "Bearer "+app.custToken)
		getW := httptest.NewRecorder()
		app.router.ServeHTTP(getW, getReq)
		assertHygiene(t, getW, http.MethodGet, "/api/support/tickets/"+createdTicket.ID, fp1)
		assert.Equal(t, http.StatusOK, getW.Code)

		// 2. Add message with fp2 to the same ticket
		replyBody, _ := json.Marshal(map[string]string{
			"message": fp2,
		})
		replyReq := httptest.NewRequest(http.MethodPost, "/api/support/tickets/"+createdTicket.ID+"/messages", bytes.NewReader(replyBody))
		replyReq.Header.Set("Content-Type", "application/json")
		replyReq.Header.Set("Authorization", "Bearer "+app.custToken)
		replyW := httptest.NewRecorder()
		app.router.ServeHTTP(replyW, replyReq)

		assertHygiene(t, replyW, http.MethodPost, "/api/support/tickets/"+createdTicket.ID+"/messages", fp2)
		assert.Equal(t, http.StatusCreated, replyW.Code, "Adding message with O'Brien single quotes should succeed with 201")
	}

	// 3. Search catalog with false positive queries
	for _, fp := range []string{fp1, fp2} {
		target := "/api/products?q=" + url.QueryEscape(fp)
		req := httptest.NewRequest(http.MethodGet, target, nil)
		w := httptest.NewRecorder()
		app.router.ServeHTTP(w, req)

		assertHygiene(t, w, http.MethodGet, target, fp)
		assert.Equal(t, http.StatusOK, w.Code, "Search with false positive query should succeed with 200")
	}
}

// TestPayloadHygiene_DatabaseIntegrity verifies that after processing hundreds of malicious probes,
// the underlying SQLite database remains completely uncorrupted, no tables were dropped, and data is intact.
func TestPayloadHygiene_DatabaseIntegrity(t *testing.T) {
	app := setupHygieneApp(t)

	// 1. Verify PRAGMA integrity_check
	var integrityResult string
	row := app.db.SQLDB().QueryRow("PRAGMA integrity_check")
	err := row.Scan(&integrityResult)
	require.NoError(t, err, "PRAGMA integrity_check query failed")
	assert.Equal(t, "ok", integrityResult, "Database failed SQLite integrity check")

	// 2. Verify seeded categories table is untouched
	var catCount int
	err = app.db.SQLDB().QueryRow("SELECT COUNT(*) FROM categories").Scan(&catCount)
	require.NoError(t, err)
	assert.Equal(t, 5, catCount, "Expected exactly 5 seeded categories in database")

	// 3. Verify seeded products table is untouched
	var prodCount int
	err = app.db.SQLDB().QueryRow("SELECT COUNT(*) FROM products").Scan(&prodCount)
	require.NoError(t, err)
	assert.Equal(t, 40, prodCount, "Expected exactly 40 seeded products in database")

	// 4. Verify seeded coupons table is untouched
	var couponCount int
	err = app.db.SQLDB().QueryRow("SELECT COUNT(*) FROM coupons").Scan(&couponCount)
	require.NoError(t, err)
	assert.Equal(t, 4, couponCount, "Expected exactly 4 seeded coupons in database")

	// 5. Verify seeded users table is intact
	var userCount int
	err = app.db.SQLDB().QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, userCount, 2, "Expected at least 2 seeded users in database")

	// 6. Verify catalog usecase still queries properly
	prods, total, err := app.catalogRepo.ListProducts(context.Background(), port.ProductFilter{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, 40, total)
	assert.Len(t, prods, 10)
}

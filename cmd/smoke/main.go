package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type TestResult struct {
	Step     int
	Name     string
	Endpoint string
	Status   int
	Duration time.Duration
	Passed   bool
	ErrorMsg string
}

type SmokeClient struct {
	baseURL    string
	httpClient *http.Client
	adminEmail string
	adminPass  string
}

func main() {
	var baseURLFlag string
	defaultBaseURL := os.Getenv("BASE_URL")
	if defaultBaseURL == "" {
		defaultBaseURL = "http://localhost"
	}

	flag.StringVar(&baseURLFlag, "base-url", defaultBaseURL, "Base URL of the LabShop target host")
	flag.Parse()

	baseURL := strings.TrimRight(baseURLFlag, "/")

	adminEmail := os.Getenv("SEED_ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "admin@labshop.local"
	}
	adminPass := os.Getenv("SEED_ADMIN_PASSWORD")
	if adminPass == "" {
		adminPass = "Admin@LabShop2026!"
	}

	client := &SmokeClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		adminEmail: adminEmail,
		adminPass:  adminPass,
	}

	fmt.Printf("=========================================================================\n")
	fmt.Printf(" LabShop Smoke Test Suite\n")
	fmt.Printf(" Target Base URL: %s\n", baseURL)
	fmt.Printf(" Started: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("=========================================================================\n\n")

	results, allPassed := client.RunAll()

	printSummaryTable(results)

	if allPassed {
		fmt.Printf("\n[SUCCESS] All %d smoke tests passed successfully.\n", len(results))
		os.Exit(0)
	} else {
		fmt.Printf("\n[FAILURE] One or more smoke tests failed.\n")
		os.Exit(1)
	}
}

func (c *SmokeClient) RunAll() ([]TestResult, bool) {
	var results []TestResult
	allPassed := true

	var firstProductID string
	var customerToken string
	var orderID string

	record := func(res TestResult) {
		results = append(results, res)
		if !res.Passed {
			allPassed = false
		}
	}

	// 1. GET /api/health -> 200, {status: "ok"}
	{
		start := time.Now()
		resp, body, err := c.get("/api/health", nil)
		dur := time.Since(start)
		if err != nil {
			record(TestResult{1, "Health check", "/api/health", 0, dur, false, err.Error()})
		} else if resp.StatusCode != http.StatusOK {
			record(TestResult{1, "Health check", "/api/health", resp.StatusCode, dur, false, fmt.Sprintf("expected 200, got %d", resp.StatusCode)})
		} else {
			var m map[string]any
			if err := json.Unmarshal(body, &m); err != nil || m["status"] != "ok" {
				record(TestResult{1, "Health check", "/api/health", resp.StatusCode, dur, false, "status != ok in json"})
			} else {
				record(TestResult{1, "Health check", "/api/health", resp.StatusCode, dur, true, ""})
			}
		}
	}

	// 2. GET /api/time -> 200, ensure Cache-Control header is NOT present
	{
		start := time.Now()
		resp, _, err := c.get("/api/time", nil)
		dur := time.Since(start)
		if err != nil {
			record(TestResult{2, "Time (no Cache-Control)", "/api/time", 0, dur, false, err.Error()})
		} else if resp.StatusCode != http.StatusOK {
			record(TestResult{2, "Time (no Cache-Control)", "/api/time", resp.StatusCode, dur, false, fmt.Sprintf("expected 200, got %d", resp.StatusCode)})
		} else if cc := resp.Header.Get("Cache-Control"); cc != "" {
			record(TestResult{2, "Time (no Cache-Control)", "/api/time", resp.StatusCode, dur, false, fmt.Sprintf("unexpected Cache-Control header: %s", cc)})
		} else {
			record(TestResult{2, "Time (no Cache-Control)", "/api/time", resp.StatusCode, dur, true, ""})
		}
	}

	// 3. GET /api/cache/public -> 200, Cache-Control: public, max-age=60
	{
		start := time.Now()
		resp, _, err := c.get("/api/cache/public", nil)
		dur := time.Since(start)
		if err != nil {
			record(TestResult{3, "Cache public header", "/api/cache/public", 0, dur, false, err.Error()})
		} else if resp.StatusCode != http.StatusOK {
			record(TestResult{3, "Cache public header", "/api/cache/public", resp.StatusCode, dur, false, fmt.Sprintf("expected 200, got %d", resp.StatusCode)})
		} else {
			cc := resp.Header.Get("Cache-Control")
			if !strings.Contains(cc, "public") || !strings.Contains(cc, "max-age=60") {
				record(TestResult{3, "Cache public header", "/api/cache/public", resp.StatusCode, dur, false, fmt.Sprintf("invalid Cache-Control: %s", cc)})
			} else {
				record(TestResult{3, "Cache public header", "/api/cache/public", resp.StatusCode, dur, true, ""})
			}
		}
	}

	// 4. GET /api/cache/etag -> 200 with ETag, then If-None-Match -> 304
	{
		start := time.Now()
		resp, _, err := c.get("/api/cache/etag", nil)
		dur := time.Since(start)
		if err != nil {
			record(TestResult{4, "Cache ETag + 304", "/api/cache/etag", 0, dur, false, err.Error()})
		} else if resp.StatusCode != http.StatusOK {
			record(TestResult{4, "Cache ETag + 304", "/api/cache/etag", resp.StatusCode, dur, false, fmt.Sprintf("first req expected 200, got %d", resp.StatusCode)})
		} else {
			etag := resp.Header.Get("ETag")
			if etag == "" {
				record(TestResult{4, "Cache ETag + 304", "/api/cache/etag", resp.StatusCode, dur, false, "ETag header missing"})
			} else {
				// Second request with If-None-Match
				start2 := time.Now()
				resp2, _, err2 := c.get("/api/cache/etag", map[string]string{"If-None-Match": etag})
				dur2 := time.Since(start2)
				if err2 != nil {
					record(TestResult{4, "Cache ETag + 304", "/api/cache/etag", 0, dur + dur2, false, err2.Error()})
				} else if resp2.StatusCode != http.StatusNotModified {
					record(TestResult{4, "Cache ETag + 304", "/api/cache/etag", resp2.StatusCode, dur + dur2, false, fmt.Sprintf("expected 304, got %d", resp2.StatusCode)})
				} else {
					record(TestResult{4, "Cache ETag + 304", "/api/cache/etag", resp2.StatusCode, dur + dur2, true, ""})
				}
			}
		}
	}

	// 5. GET /api/headers -> 200, check Authorization and Cookie are redacted
	{
		start := time.Now()
		headers := map[string]string{
			"Authorization": "Bearer secret-token-should-be-redacted",
			"Cookie":        "session_id=secret-cookie-value",
		}
		resp, body, err := c.get("/api/headers", headers)
		dur := time.Since(start)
		if err != nil {
			record(TestResult{5, "Headers redact secret", "/api/headers", 0, dur, false, err.Error()})
		} else if resp.StatusCode != http.StatusOK {
			record(TestResult{5, "Headers redact secret", "/api/headers", resp.StatusCode, dur, false, fmt.Sprintf("expected 200, got %d", resp.StatusCode)})
		} else {
			var m struct {
				Headers map[string]string `json:"headers"`
			}
			if err := json.Unmarshal(body, &m); err != nil {
				record(TestResult{5, "Headers redact secret", "/api/headers", resp.StatusCode, dur, false, "invalid json"})
			} else {
				authVal := ""
				cookieVal := ""
				for k, v := range m.Headers {
					if strings.EqualFold(k, "Authorization") {
						authVal = v
					}
					if strings.EqualFold(k, "Cookie") {
						cookieVal = v
					}
				}
				if authVal != "[REDACTED]" || cookieVal != "[REDACTED]" {
					record(TestResult{5, "Headers redact secret", "/api/headers", resp.StatusCode, dur, false, fmt.Sprintf("auth: %s, cookie: %s", authVal, cookieVal)})
				} else {
					record(TestResult{5, "Headers redact secret", "/api/headers", resp.StatusCode, dur, true, ""})
				}
			}
		}
	}

	// 6. GET /api/categories -> 200, contains 5 categories
	{
		start := time.Now()
		resp, body, err := c.get("/api/categories", nil)
		dur := time.Since(start)
		if err != nil {
			record(TestResult{6, "List categories (5)", "/api/categories", 0, dur, false, err.Error()})
		} else if resp.StatusCode != http.StatusOK {
			record(TestResult{6, "List categories (5)", "/api/categories", resp.StatusCode, dur, false, fmt.Sprintf("expected 200, got %d", resp.StatusCode)})
		} else {
			var m struct {
				Categories []any `json:"categories"`
			}
			if err := json.Unmarshal(body, &m); err != nil {
				record(TestResult{6, "List categories (5)", "/api/categories", resp.StatusCode, dur, false, "invalid json"})
			} else if len(m.Categories) != 5 {
				record(TestResult{6, "List categories (5)", "/api/categories", resp.StatusCode, dur, false, fmt.Sprintf("expected 5 categories, got %d", len(m.Categories))})
			} else {
				record(TestResult{6, "List categories (5)", "/api/categories", resp.StatusCode, dur, true, ""})
			}
		}
	}

	// 7. GET /api/products -> 200, items array present
	{
		start := time.Now()
		resp, body, err := c.get("/api/products", nil)
		dur := time.Since(start)
		if err != nil {
			record(TestResult{7, "List products", "/api/products", 0, dur, false, err.Error()})
		} else if resp.StatusCode != http.StatusOK {
			record(TestResult{7, "List products", "/api/products", resp.StatusCode, dur, false, fmt.Sprintf("expected 200, got %d", resp.StatusCode)})
		} else {
			var m struct {
				Items []struct {
					ID string `json:"id"`
				} `json:"items"`
				Total int `json:"total"`
			}
			if err := json.Unmarshal(body, &m); err != nil {
				record(TestResult{7, "List products", "/api/products", resp.StatusCode, dur, false, "invalid json"})
			} else if len(m.Items) == 0 {
				record(TestResult{7, "List products", "/api/products", resp.StatusCode, dur, false, "items array is empty"})
			} else {
				firstProductID = m.Items[0].ID
				record(TestResult{7, "List products", "/api/products", resp.StatusCode, dur, true, ""})
			}
		}
	}

	// 8. GET /api/tutorial/scenarios -> 200, scenarios present
	{
		start := time.Now()
		resp, body, err := c.get("/api/tutorial/scenarios", nil)
		dur := time.Since(start)
		if err != nil {
			record(TestResult{8, "Tutorial scenarios", "/api/tutorial/scenarios", 0, dur, false, err.Error()})
		} else if resp.StatusCode != http.StatusOK {
			record(TestResult{8, "Tutorial scenarios", "/api/tutorial/scenarios", resp.StatusCode, dur, false, fmt.Sprintf("expected 200, got %d", resp.StatusCode)})
		} else {
			var scenarios []any
			if err := json.Unmarshal(body, &scenarios); err != nil {
				record(TestResult{8, "Tutorial scenarios", "/api/tutorial/scenarios", resp.StatusCode, dur, false, "invalid json"})
			} else if len(scenarios) == 0 {
				record(TestResult{8, "Tutorial scenarios", "/api/tutorial/scenarios", resp.StatusCode, dur, false, "empty scenarios list"})
			} else {
				record(TestResult{8, "Tutorial scenarios", "/api/tutorial/scenarios", resp.StatusCode, dur, true, ""})
			}
		}
	}

	// 9. Customer register & login -> 200 with token
	{
		start := time.Now()
		uniqueEmail := fmt.Sprintf("smoke_%d@labshop.local", time.Now().UnixNano())
		password := "SmokePass123!"

		regPayload := map[string]any{
			"email":    uniqueEmail,
			"password": password,
			"name":     "Smoke Tester",
			"phone":    "0912345678",
			"address":  "123 Smoke Test St",
		}
		regResp, _, regErr := c.postJSON("/api/auth/register", regPayload, nil)
		if regErr != nil {
			record(TestResult{9, "Customer register & login", "/api/auth/register", 0, time.Since(start), false, regErr.Error()})
		} else if regResp.StatusCode != http.StatusCreated && regResp.StatusCode != http.StatusOK {
			record(TestResult{9, "Customer register & login", "/api/auth/register", regResp.StatusCode, time.Since(start), false, fmt.Sprintf("register failed with %d", regResp.StatusCode)})
		} else {
			// Login
			loginPayload := map[string]any{
				"email":    uniqueEmail,
				"password": password,
			}
			loginResp, loginBody, loginErr := c.postJSON("/api/auth/login", loginPayload, nil)
			dur := time.Since(start)
			if loginErr != nil {
				record(TestResult{9, "Customer register & login", "/api/auth/login", 0, dur, false, loginErr.Error()})
			} else if loginResp.StatusCode != http.StatusOK {
				record(TestResult{9, "Customer register & login", "/api/auth/login", loginResp.StatusCode, dur, false, fmt.Sprintf("login failed with %d", loginResp.StatusCode)})
			} else {
				var authRes struct {
					Token string `json:"token"`
				}
				if err := json.Unmarshal(loginBody, &authRes); err != nil || authRes.Token == "" {
					record(TestResult{9, "Customer register & login", "/api/auth/login", loginResp.StatusCode, dur, false, "token missing in login response"})
				} else {
					customerToken = authRes.Token
					record(TestResult{9, "Customer register & login", "/api/auth/login", loginResp.StatusCode, dur, true, ""})
				}
			}
		}
	}

	// 10. GET /api/auth/me -> 200 with user profile
	{
		start := time.Now()
		headers := map[string]string{
			"Authorization": "Bearer " + customerToken,
		}
		resp, body, err := c.get("/api/auth/me", headers)
		dur := time.Since(start)
		if err != nil {
			record(TestResult{10, "Get auth me profile", "/api/auth/me", 0, dur, false, err.Error()})
		} else if resp.StatusCode != http.StatusOK {
			record(TestResult{10, "Get auth me profile", "/api/auth/me", resp.StatusCode, dur, false, fmt.Sprintf("expected 200, got %d", resp.StatusCode)})
		} else {
			var userProf struct {
				ID    string `json:"id"`
				Email string `json:"email"`
				Role  string `json:"role"`
			}
			if err := json.Unmarshal(body, &userProf); err != nil || userProf.ID == "" {
				record(TestResult{10, "Get auth me profile", "/api/auth/me", resp.StatusCode, dur, false, "user id missing"})
			} else {
				record(TestResult{10, "Get auth me profile", "/api/auth/me", resp.StatusCode, dur, true, ""})
			}
		}
	}

	// 11. Add to cart & Checkout with Idempotency-Key -> 201 with order
	{
		start := time.Now()
		if firstProductID == "" {
			firstProductID = "p01"
		}
		authHeader := map[string]string{"Authorization": "Bearer " + customerToken}

		// Update cart item quantity
		putBody, _ := json.Marshal(map[string]any{"quantity": 1})
		putReq, _ := http.NewRequest(http.MethodPut, c.baseURL+"/api/cart/items/"+firstProductID, bytes.NewReader(putBody))
		putReq.Header.Set("Content-Type", "application/json")
		putReq.Header.Set("Authorization", "Bearer "+customerToken)
		putResp, putErr := c.httpClient.Do(putReq)
		if putErr != nil {
			record(TestResult{11, "Cart checkout (idempotent)", "/api/orders", 0, time.Since(start), false, putErr.Error()})
		} else {
			_ = putResp.Body.Close()
			if putResp.StatusCode != http.StatusOK {
				record(TestResult{11, "Cart checkout (idempotent)", "/api/cart/items", putResp.StatusCode, time.Since(start), false, fmt.Sprintf("add to cart failed: %d", putResp.StatusCode)})
			} else {
				// Checkout
				checkoutPayload := map[string]any{
					"recipient_name":   "Smoke Tester",
					"recipient_phone":  "0912345678",
					"shipping_address": "123 Smoke Test St",
					"payment_method":   "mock_gateway",
				}
				checkoutHeaders := map[string]string{
					"Authorization":   authHeader["Authorization"],
					"Idempotency-Key": fmt.Sprintf("smoke-key-%d", time.Now().UnixNano()),
				}
				orderResp, orderBody, orderErr := c.postJSON("/api/orders", checkoutPayload, checkoutHeaders)
				dur := time.Since(start)
				if orderErr != nil {
					record(TestResult{11, "Cart checkout (idempotent)", "/api/orders", 0, dur, false, orderErr.Error()})
				} else if orderResp.StatusCode != http.StatusCreated {
					record(TestResult{11, "Cart checkout (idempotent)", "/api/orders", orderResp.StatusCode, dur, false, fmt.Sprintf("expected 201, got %d", orderResp.StatusCode)})
				} else {
					var ord struct {
						ID string `json:"id"`
					}
					if err := json.Unmarshal(orderBody, &ord); err != nil || ord.ID == "" {
						record(TestResult{11, "Cart checkout (idempotent)", "/api/orders", orderResp.StatusCode, dur, false, "order id missing"})
					} else {
						orderID = ord.ID
						record(TestResult{11, "Cart checkout (idempotent)", "/api/orders", orderResp.StatusCode, dur, true, ""})
					}
				}
			}
		}
	}

	// 12. Pay order with mock gateway -> 200
	{
		start := time.Now()
		authHeader := map[string]string{"Authorization": "Bearer " + customerToken}
		payResp, payBody, payErr := c.postJSON("/api/orders/"+orderID+"/pay", nil, authHeader)
		dur := time.Since(start)
		if payErr != nil {
			record(TestResult{12, "Pay order mock gateway", "/api/orders/:id/pay", 0, dur, false, payErr.Error()})
		} else if payResp.StatusCode != http.StatusOK {
			record(TestResult{12, "Pay order mock gateway", "/api/orders/:id/pay", payResp.StatusCode, dur, false, fmt.Sprintf("expected 200, got %d", payResp.StatusCode)})
		} else {
			var ord struct {
				Status string `json:"status"`
			}
			if err := json.Unmarshal(payBody, &ord); err != nil || !strings.EqualFold(ord.Status, "paid") {
				record(TestResult{12, "Pay order mock gateway", "/api/orders/:id/pay", payResp.StatusCode, dur, false, fmt.Sprintf("expected status 'paid', got '%s'", ord.Status)})
			} else {
				record(TestResult{12, "Pay order mock gateway", "/api/orders/:id/pay", payResp.StatusCode, dur, true, ""})
			}
		}
	}

	// 13. Admin login & GET /api/admin/stats -> 200
	{
		start := time.Now()
		loginPayload := map[string]any{
			"email":    c.adminEmail,
			"password": c.adminPass,
		}
		admLoginResp, admLoginBody, admLoginErr := c.postJSON("/api/auth/login", loginPayload, nil)
		if admLoginErr != nil {
			record(TestResult{13, "Admin stats", "/api/admin/stats", 0, time.Since(start), false, admLoginErr.Error()})
		} else if admLoginResp.StatusCode != http.StatusOK {
			record(TestResult{13, "Admin stats", "/api/admin/stats", admLoginResp.StatusCode, time.Since(start), false, fmt.Sprintf("admin login failed: %d", admLoginResp.StatusCode)})
		} else {
			var admAuth struct {
				Token string `json:"token"`
			}
			if err := json.Unmarshal(admLoginBody, &admAuth); err != nil || admAuth.Token == "" {
				record(TestResult{13, "Admin stats", "/api/admin/stats", admLoginResp.StatusCode, time.Since(start), false, "admin token missing"})
			} else {
				statsResp, statsBody, statsErr := c.get("/api/admin/stats", map[string]string{
					"Authorization": "Bearer " + admAuth.Token,
				})
				dur := time.Since(start)
				if statsErr != nil {
					record(TestResult{13, "Admin stats", "/api/admin/stats", 0, dur, false, statsErr.Error()})
				} else if statsResp.StatusCode != http.StatusOK {
					record(TestResult{13, "Admin stats", "/api/admin/stats", statsResp.StatusCode, dur, false, fmt.Sprintf("expected 200, got %d", statsResp.StatusCode)})
				} else {
					var stats struct {
						TotalUsers  int `json:"total_users"`
						TotalOrders int `json:"total_orders"`
					}
					if err := json.Unmarshal(statsBody, &stats); err != nil {
						record(TestResult{13, "Admin stats", "/api/admin/stats", statsResp.StatusCode, dur, false, "invalid stats json"})
					} else {
						record(TestResult{13, "Admin stats", "/api/admin/stats", statsResp.StatusCode, dur, true, ""})
					}
				}
			}
		}
	}

	// 14. GET /openapi.json (Protected) -> 200 with Bearer token, 401 without
	{
		start := time.Now()
		respUnauth, _, _ := c.get("/openapi.json", nil)
		if respUnauth != nil && respUnauth.StatusCode != http.StatusUnauthorized {
			record(TestResult{14, "OpenAPI spec JSON", "/openapi.json", respUnauth.StatusCode, time.Since(start), false, "expected 401 Unauthorized without token"})
		} else {
			resp, body, err := c.get("/openapi.json", map[string]string{
				"Authorization": "Bearer " + customerToken,
			})
			dur := time.Since(start)
			if err != nil {
				record(TestResult{14, "OpenAPI spec JSON", "/openapi.json", 0, dur, false, err.Error()})
			} else if resp.StatusCode != http.StatusOK {
				record(TestResult{14, "OpenAPI spec JSON", "/openapi.json", resp.StatusCode, dur, false, fmt.Sprintf("expected 200, got %d", resp.StatusCode)})
			} else {
				var raw map[string]any
				if err := json.Unmarshal(body, &raw); err != nil {
					record(TestResult{14, "OpenAPI spec JSON", "/openapi.json", resp.StatusCode, dur, false, "invalid json"})
				} else {
					record(TestResult{14, "OpenAPI spec JSON", "/openapi.json", resp.StatusCode, dur, true, ""})
				}
			}
		}
	}

	// 15. GET /robots.txt -> 200 with Disallow: /
	{
		start := time.Now()
		resp, body, err := c.get("/robots.txt", nil)
		dur := time.Since(start)
		if err != nil {
			record(TestResult{15, "Robots.txt", "/robots.txt", 0, dur, false, err.Error()})
		} else if resp.StatusCode != http.StatusOK {
			record(TestResult{15, "Robots.txt", "/robots.txt", resp.StatusCode, dur, false, fmt.Sprintf("expected 200, got %d", resp.StatusCode)})
		} else if !strings.Contains(string(body), "Disallow: /") {
			record(TestResult{15, "Robots.txt", "/robots.txt", resp.StatusCode, dur, false, "missing Disallow: / in body"})
		} else {
			record(TestResult{15, "Robots.txt", "/robots.txt", resp.StatusCode, dur, true, ""})
		}
	}

	// 16. GET /sitemap.xml -> 404 (Disabled for private enterprise application)
	{
		start := time.Now()
		resp, _, err := c.get("/sitemap.xml", nil)
		dur := time.Since(start)
		if err != nil {
			record(TestResult{16, "Sitemap.xml (Disabled)", "/sitemap.xml", 0, dur, false, err.Error()})
		} else if resp.StatusCode != http.StatusNotFound {
			record(TestResult{16, "Sitemap.xml (Disabled)", "/sitemap.xml", resp.StatusCode, dur, false, fmt.Sprintf("expected 404, got %d", resp.StatusCode)})
		} else {
			record(TestResult{16, "Sitemap.xml (Disabled)", "/sitemap.xml", resp.StatusCode, dur, true, ""})
		}
	}

	// 17. GET /static/docs/sample.pdf -> 200
	{
		start := time.Now()
		resp, body, err := c.get("/static/docs/sample.pdf", nil)
		dur := time.Since(start)
		if err != nil {
			record(TestResult{17, "Static sample PDF", "/static/docs/sample.pdf", 0, dur, false, err.Error()})
		} else if resp.StatusCode != http.StatusOK {
			record(TestResult{17, "Static sample PDF", "/static/docs/sample.pdf", resp.StatusCode, dur, false, fmt.Sprintf("expected 200, got %d", resp.StatusCode)})
		} else if len(body) == 0 {
			record(TestResult{17, "Static sample PDF", "/static/docs/sample.pdf", resp.StatusCode, dur, false, "response body is empty"})
		} else {
			record(TestResult{17, "Static sample PDF", "/static/docs/sample.pdf", resp.StatusCode, dur, true, ""})
		}
	}

	// 18. GET /contents/docs/sample.pdf -> 200
	{
		start := time.Now()
		resp, body, err := c.get("/contents/docs/sample.pdf", nil)
		dur := time.Since(start)
		if err != nil {
			record(TestResult{18, "Contents sample PDF", "/contents/docs/sample.pdf", 0, dur, false, err.Error()})
		} else if resp.StatusCode != http.StatusOK {
			record(TestResult{18, "Contents sample PDF", "/contents/docs/sample.pdf", resp.StatusCode, dur, false, fmt.Sprintf("expected 200, got %d", resp.StatusCode)})
		} else if len(body) == 0 {
			record(TestResult{18, "Contents sample PDF", "/contents/docs/sample.pdf", resp.StatusCode, dur, false, "response body is empty"})
		} else {
			record(TestResult{18, "Contents sample PDF", "/contents/docs/sample.pdf", resp.StatusCode, dur, true, ""})
		}
	}

	return results, allPassed
}

func (c *SmokeClient) get(path string, headers map[string]string) (*http.Response, []byte, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return resp, body, err
}

func (c *SmokeClient) postJSON(path string, payload any, headers map[string]string) (*http.Response, []byte, error) {
	var bodyReader io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, nil, err
		}
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return resp, body, err
}

func printSummaryTable(results []TestResult) {
	fmt.Printf("+------+-------------------------------------+---------------------------------+--------+------------+--------+\n")
	fmt.Printf("| Step | Test Name                           | Endpoint                        | Status | Duration   | Result |\n")
	fmt.Printf("+------+-------------------------------------+---------------------------------+--------+------------+--------+\n")

	var totalDur time.Duration
	passedCount := 0

	for _, r := range results {
		totalDur += r.Duration
		statusStr := fmt.Sprintf("%d", r.Status)
		if r.Status == 0 {
			statusStr = "ERR"
		}

		resStr := "PASS"
		if !r.Passed {
			resStr = "FAIL"
		} else {
			passedCount++
		}

		durStr := fmt.Sprintf("%.2fms", float64(r.Duration.Microseconds())/1000.0)

		fmt.Printf("| %-4d | %-35s | %-31s | %-6s | %-10s | %-6s |\n",
			r.Step,
			truncate(r.Name, 35),
			truncate(r.Endpoint, 31),
			statusStr,
			durStr,
			resStr,
		)

		if !r.Passed && r.ErrorMsg != "" {
			fmt.Printf("|      | ^ ERROR: %-84s |\n", truncate(r.ErrorMsg, 84))
		}
	}

	fmt.Printf("+------+-------------------------------------+---------------------------------+--------+------------+--------+\n")
	fmt.Printf("Summary: Total: %d | Passed: %d | Failed: %d | Total Duration: %.2fms\n",
		len(results),
		passedCount,
		len(results)-passedCount,
		float64(totalDur.Microseconds())/1000.0,
	)
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}

package contract

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	httpadapter "labshop/internal/adapter/http"
	"labshop/internal/adapter/http/handler"
	"labshop/internal/infrastructure/config"
)

func findProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find root directory containing go.mod")
		}
		dir = parent
	}
}

// TestValidateOpenAPISpec ensures the OpenAPI 3.0.3 YAML and JSON specifications are syntactically valid.
func TestValidateOpenAPISpec(t *testing.T) {
	rootDir := findProjectRoot(t)
	yamlPath := filepath.Join(rootDir, "api", "openapi.yaml")
	jsonPath := filepath.Join(rootDir, "api", "openapi.json")
	handlerJSONPath := filepath.Join(rootDir, "internal", "adapter", "http", "handler", "openapi.json")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	// Validate YAML
	yamlDoc, err := loader.LoadFromFile(yamlPath)
	require.NoError(t, err, "failed to load api/openapi.yaml")
	require.NotNil(t, yamlDoc)

	ctx := context.Background()
	err = yamlDoc.Validate(ctx)
	require.NoError(t, err, "openapi.yaml failed kin-openapi validation")

	// Ensure size is < 1MB
	yamlInfo, err := os.Stat(yamlPath)
	require.NoError(t, err)
	assert.Less(t, yamlInfo.Size(), int64(1024*1024), "api/openapi.yaml must be < 1MB")

	// Generate / sync JSON files
	jsonBytes, err := json.MarshalIndent(yamlDoc, "", "  ")
	require.NoError(t, err)
	assert.Less(t, int64(len(jsonBytes)), int64(1024*1024), "generated json must be < 1MB")

	require.NoError(t, os.WriteFile(jsonPath, jsonBytes, 0644))
	require.NoError(t, os.WriteFile(handlerJSONPath, jsonBytes, 0644))

	// Validate generated JSON
	jsonDoc, err := loader.LoadFromFile(jsonPath)
	require.NoError(t, err, "failed to load api/openapi.json")
	err = jsonDoc.Validate(ctx)
	require.NoError(t, err, "openapi.json failed kin-openapi validation")
}

// TestRouterRoutesCoveredInOpenAPI traverses the Chi router and verifies all API routes
// exist in the OpenAPI specification (excluding /ws/*, /static/*, /contents/*, /docs, /openapi.json, HTML).
func TestRouterRoutesCoveredInOpenAPI(t *testing.T) {
	rootDir := findProjectRoot(t)
	yamlPath := filepath.Join(rootDir, "api", "openapi.yaml")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	doc, err := loader.LoadFromFile(yamlPath)
	require.NoError(t, err)
	require.NoError(t, doc.Validate(context.Background()))

	// Construct Chi router with dummy handlers registered so all routes are attached
	cfg := &config.Config{
		InstanceID:          "contract-test",
		EnableWS:            true,
		ServeStatic:         true,
		CatalogCacheSeconds: 60,
	}

	businessHandlers := &httpadapter.BusinessHandlers{
		Auth:     handler.NewAuthHandler(nil),
		Catalog:  handler.NewCatalogHandler(nil),
		Cart:     handler.NewCartHandler(nil),
		Order:    handler.NewOrderHandler(nil),
		Review:   handler.NewReviewHandler(nil),
		Support:  handler.NewSupportHandler(nil),
		Admin:    handler.NewAdminHandler(nil),
		Tutorial: handler.NewTutorialHandler(nil, 60),
	}

	router := httpadapter.NewRouter(cfg, nil, nil, nil, businessHandlers)
	require.NotNil(t, router)

	var missingRoutes []string

	walkErr := chi.Walk(router, func(method string, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		// Filter out excluded routes according to requirements:
		// excluding /ws/*, /static/*, /contents/*, /docs, /openapi.json, HTML
		if strings.HasPrefix(route, "/ws/") || route == "/ws/*" {
			return nil
		}
		if strings.HasPrefix(route, "/static/") || route == "/static/*" {
			return nil
		}
		if strings.HasPrefix(route, "/contents/") || route == "/contents/*" {
			return nil
		}
		if strings.HasPrefix(route, "/docs") || route == "/docs" {
			return nil
		}
		if route == "/openapi.json" || route == "/robots.txt" || route == "/sitemap.xml" {
			return nil
		}

		// Normalize chi wildcard route /* to /{path} to match OpenAPI path template
		openAPIRoute := route
		if strings.HasSuffix(openAPIRoute, "/*") {
			openAPIRoute = strings.TrimSuffix(openAPIRoute, "/*") + "/{path}"
		}

		// Look up in doc.Paths
		pathItem := doc.Paths.Find(openAPIRoute)
		if pathItem == nil {
			missingRoutes = append(missingRoutes, method+" "+route+" (expected "+openAPIRoute+" in spec)")
			return nil
		}

		op := pathItem.GetOperation(method)
		if op == nil {
			missingRoutes = append(missingRoutes, method+" "+route+" (operation not defined on "+openAPIRoute+")")
		}

		return nil
	})

	require.NoError(t, walkErr)
	assert.Empty(t, missingRoutes, "The following API routes are registered in Chi but missing in OpenAPI spec:\n%s", strings.Join(missingRoutes, "\n"))
}

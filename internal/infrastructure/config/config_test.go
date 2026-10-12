package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_LoadDefaults(t *testing.T) {
	// Clear any environment overrides
	os.Unsetenv("HTTP_ADDR")
	os.Unsetenv("APP_ENV")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("ADMIN_TOKEN")
	os.Unsetenv("INSTANCE_ID")
	os.Unsetenv("CATALOG_CACHE_SECONDS")
	os.Unsetenv("MAX_UPLOAD_MB")
	os.Unsetenv("MAX_SLOW_MS")
	os.Unsetenv("MAX_PAYLOAD_KB")
	os.Unsetenv("SERVE_STATIC")
	os.Unsetenv("STATIC_DIR")
	os.Unsetenv("ENABLE_WS")
	os.Unsetenv("ENABLE_DOCS")
	os.Unsetenv("ORIGIN_GZIP")
	os.Unsetenv("LOG_LEVEL")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, ":8080", cfg.HTTPAddr)
	assert.Equal(t, "development", cfg.AppEnv)
	assert.NotEmpty(t, cfg.JWTSecret)
	assert.Equal(t, "labshop-admin-secret-token-key-2026", cfg.AdminToken)
	assert.NotEmpty(t, cfg.InstanceID)
	assert.Equal(t, 60, cfg.CatalogCacheSeconds)
	assert.Equal(t, 10, cfg.MaxUploadMB)
	assert.Equal(t, 60000, cfg.MaxSlowMS)
	assert.Equal(t, 5120, cfg.MaxPayloadKB)
	assert.False(t, cfg.ServeStatic)
	assert.Equal(t, "web/static", cfg.StaticDir)
	assert.True(t, cfg.EnableWS)
	assert.True(t, cfg.EnableDocs)
	assert.Equal(t, "off", cfg.OriginGzip)
	assert.Equal(t, "info", cfg.LogLevel)
}

func TestConfig_ProductionJWTValidation(t *testing.T) {
	t.Setenv("APP_ENV", "production")

	// Short secret should fail
	t.Setenv("JWT_SECRET", "too-short-secret")
	t.Setenv("ADMIN_TOKEN", "valid-admin-token-16-chars")
	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "JWT_SECRET must be at least 32 characters in production")

	// Short admin token should fail
	t.Setenv("JWT_SECRET", "this-is-a-very-secure-secret-key-32-chars")
	t.Setenv("ADMIN_TOKEN", "short")
	cfg, err = Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "ADMIN_TOKEN must be at least 16 characters in production")

	// Valid secrets should succeed
	t.Setenv("ADMIN_TOKEN", "valid-admin-token-16-chars")
	cfg, err = Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "production", cfg.AppEnv)
}

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
	os.Unsetenv("INSTANCE_ID")
	os.Unsetenv("CATALOG_CACHE_SECONDS")
	os.Unsetenv("MAX_UPLOAD_MB")
	os.Unsetenv("MAX_SLOW_MS")
	os.Unsetenv("MAX_PAYLOAD_KB")
	os.Unsetenv("LOG_LEVEL")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, ":8080", cfg.HTTPAddr)
	assert.Equal(t, "development", cfg.AppEnv)
	assert.NotEmpty(t, cfg.JWTSecret)
	assert.NotEmpty(t, cfg.InstanceID)
	assert.Equal(t, 60, cfg.CatalogCacheSeconds)
	assert.Equal(t, 10, cfg.MaxUploadMB)
	assert.Equal(t, 60000, cfg.MaxSlowMS)
	assert.Equal(t, 5120, cfg.MaxPayloadKB)
	assert.Equal(t, "info", cfg.LogLevel)
}

func TestConfig_ProductionJWTValidation(t *testing.T) {
	t.Setenv("APP_ENV", "production")

	// Short secret should fail
	t.Setenv("JWT_SECRET", "too-short-secret")
	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "JWT_SECRET must be at least 32 characters in production")

	// Valid 32+ character secret should succeed
	t.Setenv("JWT_SECRET", "this-is-a-very-secure-secret-key-32-chars")
	cfg, err = Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "production", cfg.AppEnv)
}

package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration parameters.
type Config struct {
	HTTPAddr            string
	AppEnv              string
	JWTSecret           string
	InstanceID          string
	CatalogCacheSeconds int
	MaxUploadMB         int
	MaxSlowMS           int
	MaxPayloadKB        int
	LogLevel            string
}

// Load reads configuration from environment variables with fallbacks.
// It performs validation and fails fast if critical rules are violated.
func Load() (*Config, error) {
	httpAddr := getEnv("HTTP_ADDR", ":8080")
	appEnv := strings.ToLower(getEnv("APP_ENV", "development"))
	jwtSecret := os.Getenv("JWT_SECRET")
	logLevel := strings.ToLower(getEnv("LOG_LEVEL", "info"))

	instanceID := os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		if host, err := os.Hostname(); err == nil && host != "" {
			instanceID = host
		} else {
			instanceID = "labshop-origin"
		}
	}

	catalogCacheSeconds := getEnvInt("CATALOG_CACHE_SECONDS", 60)
	maxUploadMB := getEnvInt("MAX_UPLOAD_MB", 10)
	maxSlowMS := getEnvInt("MAX_SLOW_MS", 60000)
	maxPayloadKB := getEnvInt("MAX_PAYLOAD_KB", 5120)

	// Validate production JWT secret requirement
	if appEnv == "production" {
		if len(jwtSecret) < 32 {
			return nil, errors.New("JWT_SECRET must be at least 32 characters in production")
		}
	} else if jwtSecret == "" {
		// Provide a secure default for development/testing
		jwtSecret = "development-secret-key-must-be-at-least-32-chars-long"
	}

	cfg := &Config{
		HTTPAddr:            httpAddr,
		AppEnv:              appEnv,
		JWTSecret:           jwtSecret,
		InstanceID:          instanceID,
		CatalogCacheSeconds: catalogCacheSeconds,
		MaxUploadMB:         maxUploadMB,
		MaxSlowMS:           maxSlowMS,
		MaxPayloadKB:        maxPayloadKB,
		LogLevel:            logLevel,
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// Validate verifies config bounds and invariants.
func (c *Config) Validate() error {
	if c.HTTPAddr == "" {
		return errors.New("HTTP_ADDR cannot be empty")
	}
	if c.CatalogCacheSeconds < 0 {
		return errors.New("CATALOG_CACHE_SECONDS must be non-negative")
	}
	if c.MaxUploadMB <= 0 {
		return errors.New("MAX_UPLOAD_MB must be positive")
	}
	if c.MaxSlowMS <= 0 {
		return errors.New("MAX_SLOW_MS must be positive")
	}
	if c.MaxPayloadKB <= 0 {
		return errors.New("MAX_PAYLOAD_KB must be positive")
	}
	return nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

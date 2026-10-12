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
	AdminToken          string
	InstanceID          string
	CatalogCacheSeconds int
	MaxUploadMB         int
	MaxSlowMS           int
	MaxPayloadKB        int
	ServeStatic         bool
	StaticDir           string
	EnableWS            bool
	EnableDocs          bool
	OriginGzip          string
	LogLevel            string
	DBPath              string
	Seed                bool
	SeedAdminEmail      string
	SeedAdminPassword   string
	SeedCustomerEmail   string
	SeedCustomerPassword string
	MockPaymentFailRate float64
}

// Load reads configuration from environment variables with fallbacks.
// It performs validation and fails fast if critical rules are violated.
func Load() (*Config, error) {
	httpAddr := getEnv("HTTP_ADDR", ":8080")
	appEnv := strings.ToLower(getEnv("APP_ENV", "development"))
	jwtSecret := os.Getenv("JWT_SECRET")
	adminToken := os.Getenv("ADMIN_TOKEN")
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
	serveStatic := getEnvBool("SERVE_STATIC", false)
	staticDir := getEnv("STATIC_DIR", "web/static")
	enableWS := getEnvBool("ENABLE_WS", true)
	enableDocs := getEnvBool("ENABLE_DOCS", true)
	originGzip := getEnv("ORIGIN_GZIP", "off")

	dbPath := getEnv("DB_PATH", "labshop.db")
	seed := getEnvBool("SEED", false)
	seedAdminEmail := getEnv("SEED_ADMIN_EMAIL", "admin@labshop.local")
	seedAdminPassword := getEnv("SEED_ADMIN_PASSWORD", "Admin@LabShop2026!")
	seedCustomerEmail := getEnv("SEED_CUSTOMER_EMAIL", "customer@labshop.local")
	seedCustomerPassword := getEnv("SEED_CUSTOMER_PASSWORD", "Customer@LabShop2026!")
	mockPaymentFailRate := getEnvFloat("MOCK_PAYMENT_FAIL_RATE", 0.0)

	// Validate production JWT secret and admin token requirement
	if appEnv == "production" {
		if len(jwtSecret) < 32 {
			return nil, errors.New("JWT_SECRET must be at least 32 characters in production")
		}
		if len(adminToken) < 16 {
			return nil, errors.New("ADMIN_TOKEN must be at least 16 characters in production")
		}
	} else {
		if jwtSecret == "" {
			jwtSecret = "development-secret-key-must-be-at-least-32-chars-long"
		}
		if adminToken == "" {
			adminToken = "labshop-admin-secret-token-key-2026"
		}
	}

	cfg := &Config{
		HTTPAddr:            httpAddr,
		AppEnv:              appEnv,
		JWTSecret:           jwtSecret,
		AdminToken:          adminToken,
		InstanceID:          instanceID,
		CatalogCacheSeconds: catalogCacheSeconds,
		MaxUploadMB:         maxUploadMB,
		MaxSlowMS:           maxSlowMS,
		MaxPayloadKB:        maxPayloadKB,
		ServeStatic:         serveStatic,
		StaticDir:           staticDir,
		EnableWS:            enableWS,
		EnableDocs:          enableDocs,
		OriginGzip:          originGzip,
		LogLevel:            logLevel,
		DBPath:              dbPath,
		Seed:                seed,
		SeedAdminEmail:      seedAdminEmail,
		SeedAdminPassword:   seedAdminPassword,
		SeedCustomerEmail:   seedCustomerEmail,
		SeedCustomerPassword: seedCustomerPassword,
		MockPaymentFailRate: mockPaymentFailRate,
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

func getEnvBool(key string, defaultVal bool) bool {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(valStr)
	if err != nil {
		return defaultVal
	}
	return b
}

func getEnvFloat(key string, defaultVal float64) float64 {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	f, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return defaultVal
	}
	return f
}


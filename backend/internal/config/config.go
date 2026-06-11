// Package config loads and validates runtime configuration from the
// environment (and an optional .env file). It only reads configuration —
// connecting to external systems is the job of other packages.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration for the backend service.
type Config struct {
	// AppEnv is "development" or "production".
	AppEnv string
	// ServerPort is the HTTP listen port (without colon), e.g. "8080".
	ServerPort string

	// MySQL connection settings.
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// MySQL connection pool settings.
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration

	// CORSAllowedOrigins is the list of origins allowed by the CORS
	// middleware. In development this is the frontend dev server.
	CORSAllowedOrigins []string
}

// IsProduction reports whether the service runs in production mode.
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

// DBDSN builds the go-sql-driver/mysql DSN. parseTime + UTC keep all time
// columns consistent with the API contract (UTC ISO-8601).
func (c *Config) DBDSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=UTC",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

// Load reads configuration from the environment. A .env file, if present in
// the working directory, is loaded first but never overrides variables that
// are already set in the real environment. Every field has a safe default for
// local development; only clearly invalid combinations cause an error.
func Load() (*Config, error) {
	// .env is optional and developer-local. Missing file is not an error.
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:     getEnv("APP_ENV", "development"),
		ServerPort: getEnv("SERVER_PORT", "8080"),

		DBHost:     getEnv("DB_HOST", "127.0.0.1"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "iclassroom"),

		DBMaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 25),
		DBConnMaxLifetime: time.Duration(getEnvInt("DB_CONN_MAX_LIFETIME_MINUTES", 5)) * time.Minute,

		CORSAllowedOrigins: getEnvList("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173"}),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// validate guards against configuration that would silently break the service.
func (c *Config) validate() error {
	if c.ServerPort == "" {
		return fmt.Errorf("config: SERVER_PORT must not be empty")
	}
	if c.DBUser == "" {
		return fmt.Errorf("config: DB_USER must not be empty")
	}
	if c.DBName == "" {
		return fmt.Errorf("config: DB_NAME must not be empty")
	}
	if c.DBMaxOpenConns <= 0 {
		return fmt.Errorf("config: DB_MAX_OPEN_CONNS must be > 0, got %d", c.DBMaxOpenConns)
	}
	if c.DBMaxIdleConns < 0 {
		return fmt.Errorf("config: DB_MAX_IDLE_CONNS must be >= 0, got %d", c.DBMaxIdleConns)
	}
	if len(c.CORSAllowedOrigins) == 0 {
		return fmt.Errorf("config: CORS_ALLOWED_ORIGINS must list at least one origin")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvList(key string, fallback []string) []string {
	v, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(v) == "" {
		return fallback
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}

package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL        string
	CTLogBaseURL       string
	CTRequestTimeout   time.Duration
	MonitorInterval    time.Duration
	BatchSize          int64
	HTTPPort           string
	MigrationsDir      string
	CORSAllowedOrigins []string
}

func Load() Config {
	_ = godotenv.Load()

	intervalSeconds := getInt("MONITOR_INTERVAL_SECONDS", 60)
	ctRequestTimeoutSeconds := getInt("CT_REQUEST_TIMEOUT_SECONDS", 30)
	batchSize := getInt64("BATCH_SIZE", 100)

	return Config{
		DatabaseURL:        getString("DATABASE_URL", "postgres://postgres:admin@localhost:5432/brand_protection?sslmode=disable"),
		CTLogBaseURL:       getString("CT_LOG_BASE_URL", "https://ct.googleapis.com/logs/us1/argon2026h1/ct/v1"),
		CTRequestTimeout:   time.Duration(ctRequestTimeoutSeconds) * time.Second,
		MonitorInterval:    time.Duration(intervalSeconds) * time.Second,
		BatchSize:          batchSize,
		HTTPPort:           getString("HTTP_PORT", "8080"),
		MigrationsDir:      getString("MIGRATIONS_DIR", "../migrations"),
		CORSAllowedOrigins: getCSV("CORS_ALLOWED_ORIGINS", "http://localhost:5173"),
	}
}

func getString(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func getCSV(key string, fallback string) []string {
	value := getString(key, fallback)
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	if len(result) == 0 {
		return []string{fallback}
	}
	return result
}

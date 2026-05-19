package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL     string
	CTLogBaseURL    string
	MonitorInterval time.Duration
	BatchSize       int64
	HTTPPort        string
	MigrationsDir   string
}

func Load() Config {
	_ = godotenv.Load()

	intervalSeconds := getInt("MONITOR_INTERVAL_SECONDS", 60)
	batchSize := getInt64("BATCH_SIZE", 100)

	return Config{
		DatabaseURL:     getString("DATABASE_URL", "postgres://postgres:admin@localhost:5432/brand_protection?sslmode=disable"),
		CTLogBaseURL:    getString("CT_LOG_BASE_URL", "https://ct.googleapis.com/logs/us1/argon2026h1/ct/v1"),
		MonitorInterval: time.Duration(intervalSeconds) * time.Second,
		BatchSize:       batchSize,
		HTTPPort:        getString("HTTP_PORT", "8080"),
		MigrationsDir:   getString("MIGRATIONS_DIR", "../migrations"),
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

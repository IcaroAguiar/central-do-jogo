package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds process configuration loaded from the environment.
type Config struct {
	HTTPAddr          string
	ShutdownTimeout   time.Duration
	DatabaseURL       string
	StaticDir         string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

// Load reads configuration from environment variables and returns explicit errors.
func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:          envOr("HTTP_ADDR", ":8080"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		StaticDir:         envOr("STATIC_DIR", "web/dist"),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	shutdownMS, err := envInt("SHUTDOWN_TIMEOUT_MS", 15000)
	if err != nil {
		return Config{}, err
	}
	if shutdownMS <= 0 {
		return Config{}, fmt.Errorf("SHUTDOWN_TIMEOUT_MS must be positive, got %d", shutdownMS)
	}
	cfg.ShutdownTimeout = time.Duration(shutdownMS) * time.Millisecond

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return value, nil
}

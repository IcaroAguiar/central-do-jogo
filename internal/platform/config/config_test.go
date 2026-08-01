package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("SHUTDOWN_TIMEOUT_MS", "")
	t.Setenv("STATIC_DIR", "")
	t.Setenv("DATABASE_URL", "postgres://central:central_dev_only@127.0.0.1:5433/central_do_jogo?sslmode=disable")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}
	if cfg.ShutdownTimeout != 15*time.Second {
		t.Fatalf("ShutdownTimeout = %v, want 15s", cfg.ShutdownTimeout)
	}
	if cfg.StaticDir != "web/dist" {
		t.Fatalf("StaticDir = %q, want web/dist", cfg.StaticDir)
	}
	if cfg.DatabaseURL == "" {
		t.Fatal("DatabaseURL unexpectedly empty")
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error when DATABASE_URL is empty")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("error = %v, want DATABASE_URL mention", err)
	}
}

func TestLoadInvalidShutdownTimeout(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://central:central_dev_only@127.0.0.1:5433/central_do_jogo?sslmode=disable")
	t.Setenv("SHUTDOWN_TIMEOUT_MS", "not-a-number")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for invalid SHUTDOWN_TIMEOUT_MS")
	}
}

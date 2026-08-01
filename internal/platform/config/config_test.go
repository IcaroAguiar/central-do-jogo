package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("SHUTDOWN_TIMEOUT_MS", "")
	t.Setenv("STATIC_DIR", "")
	t.Setenv("DATABASE_URL", "")

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
}

func TestLoadInvalidShutdownTimeout(t *testing.T) {
	t.Setenv("SHUTDOWN_TIMEOUT_MS", "not-a-number")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for invalid SHUTDOWN_TIMEOUT_MS")
	}
}

func TestLoadRejectsNonPositiveShutdownTimeout(t *testing.T) {
	t.Setenv("SHUTDOWN_TIMEOUT_MS", "0")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for non-positive SHUTDOWN_TIMEOUT_MS")
	}
}

func TestLoadCustomValues(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("SHUTDOWN_TIMEOUT_MS", "2500")
	t.Setenv("STATIC_DIR", "/srv/static")
	t.Setenv("DATABASE_URL", "postgres://example")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("HTTPAddr = %q, want :9090", cfg.HTTPAddr)
	}
	if cfg.ShutdownTimeout != 2500*time.Millisecond {
		t.Fatalf("ShutdownTimeout = %v, want 2500ms", cfg.ShutdownTimeout)
	}
	if cfg.StaticDir != "/srv/static" {
		t.Fatalf("StaticDir = %q, want /srv/static", cfg.StaticDir)
	}
	if cfg.DatabaseURL != "postgres://example" {
		t.Fatalf("DatabaseURL = %q, want postgres://example", cfg.DatabaseURL)
	}
}

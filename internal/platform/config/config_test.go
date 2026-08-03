package config

import (
	"strings"
	"testing"
	"time"
)

const testDatabaseURL = "postgres://central:central_dev_only@127.0.0.1:5433/central_do_jogo?sslmode=disable"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("SHUTDOWN_TIMEOUT_MS", "")
	t.Setenv("STATIC_DIR", "")
	t.Setenv("DATABASE_URL", testDatabaseURL)
	clearAuthEnv(t)

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
	if !cfg.SSREnabled {
		t.Fatal("SSREnabled default should be true")
	}
	if cfg.SearchRateLimitPerSecond != 2 {
		t.Fatalf("SearchRateLimitPerSecond = %v, want 2", cfg.SearchRateLimitPerSecond)
	}
	if cfg.SearchRateLimitBurst != 10 {
		t.Fatalf("SearchRateLimitBurst = %d, want 10", cfg.SearchRateLimitBurst)
	}
	if cfg.PublicBaseURL != "" {
		t.Fatalf("PublicBaseURL = %q, want empty", cfg.PublicBaseURL)
	}
	if cfg.AuthEnabled {
		t.Fatal("AuthEnabled default should be false")
	}
}

func clearAuthEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"GOOGLE_OAUTH_CLIENT_ID",
		"GOOGLE_OAUTH_CLIENT_SECRET",
		"GOOGLE_OAUTH_REDIRECT_URL",
		"SESSION_COOKIE_SECRET",
		"MAINTAINER_ALLOWLIST",
		"AUTH_COOKIE_SECURE",
		"AUTH_POST_LOGIN_REDIRECT",
		"SESSION_TTL_HOURS",
		"AUTH_RATE_LIMIT_PER_SECOND",
		"AUTH_RATE_LIMIT_BURST",
	} {
		t.Setenv(key, "")
	}
}

func TestLoadAuthEnabled(t *testing.T) {
	t.Setenv("DATABASE_URL", testDatabaseURL)
	clearAuthEnv(t)
	t.Setenv("PUBLIC_BASE_URL", "https://example.org")
	t.Setenv("GOOGLE_OAUTH_CLIENT_ID", "client-id")
	t.Setenv("GOOGLE_OAUTH_CLIENT_SECRET", "client-secret")
	t.Setenv("SESSION_COOKIE_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("MAINTAINER_ALLOWLIST", "Owner@Example.com, other@example.com, owner@example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if !cfg.AuthEnabled {
		t.Fatal("AuthEnabled should be true")
	}
	if cfg.GoogleOAuthRedirectURL != "https://example.org/api/v1/auth/google/callback" {
		t.Fatalf("redirect = %q", cfg.GoogleOAuthRedirectURL)
	}
	if !cfg.AuthCookieSecure {
		t.Fatal("AuthCookieSecure should default true for https PublicBaseURL")
	}
	if len(cfg.MaintainerAllowlistEmails) != 2 {
		t.Fatalf("allowlist = %#v", cfg.MaintainerAllowlistEmails)
	}
}

func TestLoadRejectsAuthWithoutPublicBaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", testDatabaseURL)
	clearAuthEnv(t)
	t.Setenv("PUBLIC_BASE_URL", "")
	t.Setenv("GOOGLE_OAUTH_CLIENT_ID", "client-id")
	t.Setenv("GOOGLE_OAUTH_CLIENT_SECRET", "client-secret")
	t.Setenv("SESSION_COOKIE_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("GOOGLE_OAUTH_REDIRECT_URL", "https://example.org/api/v1/auth/google/callback")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when PUBLIC_BASE_URL is empty with auth enabled")
	}
	if !strings.Contains(err.Error(), "PUBLIC_BASE_URL") {
		t.Fatalf("error = %v", err)
	}
}

func TestLoadRejectsIncompleteAuth(t *testing.T) {
	t.Setenv("DATABASE_URL", testDatabaseURL)
	clearAuthEnv(t)
	t.Setenv("GOOGLE_OAUTH_CLIENT_ID", "only-id")

	_, err := Load()
	if err == nil {
		t.Fatal("expected incomplete auth error")
	}
	if !strings.Contains(err.Error(), "incomplete Google OAuth") {
		t.Fatalf("error = %v", err)
	}
}

func TestLoadCustomValues(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("SHUTDOWN_TIMEOUT_MS", "2500")
	t.Setenv("STATIC_DIR", "/srv/static")
	t.Setenv("DATABASE_URL", testDatabaseURL)
	clearAuthEnv(t)

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
}

func TestLoadSSRAndRateLimitOverrides(t *testing.T) {
	t.Setenv("DATABASE_URL", testDatabaseURL)
	clearAuthEnv(t)
	t.Setenv("SSR_ENABLED", "false")
	t.Setenv("SEARCH_RATE_LIMIT_PER_SECOND", "5.5")
	t.Setenv("SEARCH_RATE_LIMIT_BURST", "20")
	t.Setenv("PUBLIC_BASE_URL", "https://example.org/")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.SSREnabled {
		t.Fatal("SSREnabled should be false")
	}
	if cfg.SearchRateLimitPerSecond != 5.5 {
		t.Fatalf("SearchRateLimitPerSecond = %v, want 5.5", cfg.SearchRateLimitPerSecond)
	}
	if cfg.SearchRateLimitBurst != 20 {
		t.Fatalf("SearchRateLimitBurst = %d, want 20", cfg.SearchRateLimitBurst)
	}
	if cfg.PublicBaseURL != "https://example.org" {
		t.Fatalf("PublicBaseURL = %q, want trimmed trailing slash", cfg.PublicBaseURL)
	}
}

func TestLoadRejectsInvalidRateLimit(t *testing.T) {
	t.Setenv("DATABASE_URL", testDatabaseURL)
	clearAuthEnv(t)
	t.Setenv("SEARCH_RATE_LIMIT_PER_SECOND", "0")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for non-positive rate limit")
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	clearAuthEnv(t)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error when DATABASE_URL is empty")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("error = %v, want DATABASE_URL mention", err)
	}
}

func TestLoadInvalidShutdownTimeout(t *testing.T) {
	t.Setenv("DATABASE_URL", testDatabaseURL)
	clearAuthEnv(t)
	t.Setenv("SHUTDOWN_TIMEOUT_MS", "not-a-number")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for invalid SHUTDOWN_TIMEOUT_MS")
	}
}

func TestLoadRejectsNonPositiveShutdownTimeout(t *testing.T) {
	t.Setenv("DATABASE_URL", testDatabaseURL)
	clearAuthEnv(t)
	t.Setenv("SHUTDOWN_TIMEOUT_MS", "0")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for non-positive SHUTDOWN_TIMEOUT_MS")
	}
}

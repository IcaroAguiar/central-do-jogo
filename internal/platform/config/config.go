package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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
	// PublicBaseURL is the absolute origin (scheme+host, no trailing slash)
	// used to build canonical/OG URLs in SSR pages (REQ-015). Empty means
	// root-relative canonical URLs are emitted instead.
	PublicBaseURL string
	// SSREnabled toggles Go server-side rendering for "/", "/clubes/{slug}",
	// and "/jogos/{slug}" (PAT-004). When false, those routes fall back to
	// the SPA shell like any other client route.
	SSREnabled bool
	// SearchRateLimitPerSecond and SearchRateLimitBurst configure the
	// in-process token bucket applied to GET /api/v1/search (SEC-001).
	SearchRateLimitPerSecond float64
	SearchRateLimitBurst     int
	// TrustedProxyCIDRs lists CIDRs (or bare IPs) whose RemoteAddr may set
	// X-Forwarded-For for search rate limiting. Empty means proxy headers
	// are ignored (correct for direct exposure; required behind nginx/CDN).
	TrustedProxyCIDRs []string
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
		PublicBaseURL:     strings.TrimSuffix(os.Getenv("PUBLIC_BASE_URL"), "/"),
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

	ssrEnabled, err := envBool("SSR_ENABLED", true)
	if err != nil {
		return Config{}, err
	}
	cfg.SSREnabled = ssrEnabled

	rateLimitPerSecond, err := envFloat("SEARCH_RATE_LIMIT_PER_SECOND", 2)
	if err != nil {
		return Config{}, err
	}
	if rateLimitPerSecond <= 0 {
		return Config{}, fmt.Errorf("SEARCH_RATE_LIMIT_PER_SECOND must be positive, got %v", rateLimitPerSecond)
	}
	cfg.SearchRateLimitPerSecond = rateLimitPerSecond

	rateLimitBurst, err := envInt("SEARCH_RATE_LIMIT_BURST", 10)
	if err != nil {
		return Config{}, err
	}
	if rateLimitBurst <= 0 {
		return Config{}, fmt.Errorf("SEARCH_RATE_LIMIT_BURST must be positive, got %d", rateLimitBurst)
	}
	cfg.SearchRateLimitBurst = rateLimitBurst

	cfg.TrustedProxyCIDRs = envCSV("TRUSTED_PROXY_CIDRS")

	return cfg, nil
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envCSV(key string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
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

func envFloat(key string, fallback float64) (float64, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return value, nil
}

func envBool(key string, fallback bool) (bool, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("parse %s: %w", key, err)
	}
	return value, nil
}

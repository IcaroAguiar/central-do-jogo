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

	// AuthEnabled is true when Google OAuth + session secret are fully set.
	// When false, public content still works and /api/v1/auth/me reports
	// authEnabled=false (RISK-008).
	AuthEnabled bool
	// GoogleOAuthClientID / GoogleOAuthClientSecret / GoogleOAuthRedirectURL
	// configure the initial public OAuth provider (REQ-017).
	GoogleOAuthClientID     string
	GoogleOAuthClientSecret string
	GoogleOAuthRedirectURL  string
	// SessionCookieSecret HMAC-signs OAuth state and is mixed into session
	// operations. Required when AuthEnabled.
	SessionCookieSecret string
	// SessionTTL is how long a login session cookie remains valid.
	SessionTTL time.Duration
	// AuthCookieSecure forces Secure cookies. When unset, Secure follows
	// whether PublicBaseURL is https.
	AuthCookieSecure bool
	// MaintainerAllowlistEmails grants RoleMaintainer on login (REQ-018).
	// First login never promotes unless the email is present here.
	MaintainerAllowlistEmails []string
	// AuthRateLimitPerSecond / AuthRateLimitBurst guard OAuth start+callback.
	AuthRateLimitPerSecond float64
	AuthRateLimitBurst     int
	// AuthPostLoginRedirect is the relative path after successful OAuth.
	AuthPostLoginRedirect string
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

	if err := loadAuth(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func loadAuth(cfg *Config) error {
	cfg.GoogleOAuthClientID = strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_CLIENT_ID"))
	cfg.GoogleOAuthClientSecret = strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"))
	cfg.GoogleOAuthRedirectURL = strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_REDIRECT_URL"))
	cfg.SessionCookieSecret = strings.TrimSpace(os.Getenv("SESSION_COOKIE_SECRET"))
	cfg.MaintainerAllowlistEmails = normalizeEmails(envCSV("MAINTAINER_ALLOWLIST"))
	cfg.AuthPostLoginRedirect = strings.TrimSpace(os.Getenv("AUTH_POST_LOGIN_REDIRECT"))
	if cfg.AuthPostLoginRedirect == "" {
		cfg.AuthPostLoginRedirect = "/"
	}

	sessionTTLHours, err := envInt("SESSION_TTL_HOURS", 720) // 30 days
	if err != nil {
		return err
	}
	if sessionTTLHours <= 0 {
		return fmt.Errorf("SESSION_TTL_HOURS must be positive, got %d", sessionTTLHours)
	}
	cfg.SessionTTL = time.Duration(sessionTTLHours) * time.Hour

	cookieSecure, err := envBool("AUTH_COOKIE_SECURE", strings.HasPrefix(cfg.PublicBaseURL, "https://"))
	if err != nil {
		return err
	}
	cfg.AuthCookieSecure = cookieSecure

	authRate, err := envFloat("AUTH_RATE_LIMIT_PER_SECOND", 1)
	if err != nil {
		return err
	}
	if authRate <= 0 {
		return fmt.Errorf("AUTH_RATE_LIMIT_PER_SECOND must be positive, got %v", authRate)
	}
	cfg.AuthRateLimitPerSecond = authRate

	authBurst, err := envInt("AUTH_RATE_LIMIT_BURST", 5)
	if err != nil {
		return err
	}
	if authBurst <= 0 {
		return fmt.Errorf("AUTH_RATE_LIMIT_BURST must be positive, got %d", authBurst)
	}
	cfg.AuthRateLimitBurst = authBurst

	anyOAuth := cfg.GoogleOAuthClientID != "" || cfg.GoogleOAuthClientSecret != "" ||
		cfg.GoogleOAuthRedirectURL != "" || cfg.SessionCookieSecret != ""
	allOAuth := cfg.GoogleOAuthClientID != "" && cfg.GoogleOAuthClientSecret != "" &&
		cfg.SessionCookieSecret != ""

	if anyOAuth && !allOAuth {
		return fmt.Errorf("incomplete Google OAuth config: set GOOGLE_OAUTH_CLIENT_ID, GOOGLE_OAUTH_CLIENT_SECRET, and SESSION_COOKIE_SECRET together (or leave all empty)")
	}
	if allOAuth && len(cfg.SessionCookieSecret) < 32 {
		return fmt.Errorf("SESSION_COOKIE_SECRET must be at least 32 characters")
	}
	if allOAuth {
		if cfg.GoogleOAuthRedirectURL == "" {
			if cfg.PublicBaseURL == "" {
				return fmt.Errorf("GOOGLE_OAUTH_REDIRECT_URL is required when PUBLIC_BASE_URL is empty")
			}
			cfg.GoogleOAuthRedirectURL = cfg.PublicBaseURL + "/api/v1/auth/google/callback"
		}
		cfg.AuthEnabled = true
	}
	return nil
}

func normalizeEmails(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, 0, len(in))
	seen := map[string]struct{}{}
	for _, e := range in {
		e = strings.ToLower(strings.TrimSpace(e))
		if e == "" {
			continue
		}
		if _, ok := seen[e]; ok {
			continue
		}
		seen[e] = struct{}{}
		out = append(out, e)
	}
	return out
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

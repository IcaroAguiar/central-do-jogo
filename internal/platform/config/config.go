package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// AuthConfig groups Google OAuth and session settings (env names unchanged).
type AuthConfig struct {
	Enabled                   bool
	GoogleOAuthClientID       string
	GoogleOAuthClientSecret   string
	GoogleOAuthRedirectURL    string
	SessionCookieSecret       string
	SessionTTL                time.Duration
	CookieSecure              bool
	MaintainerAllowlistEmails []string
	RateLimitPerSecond        float64
	RateLimitBurst            int
	PostLoginRedirect         string
}

// PushConfig groups Web Push / VAPID settings (env names unchanged).
type PushConfig struct {
	Enabled         bool
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDSubject    string
}

// PrivacyConfig groups first-party analytics retention (TASK-030).
type PrivacyConfig struct {
	AnalyticsRetentionDays int
}

// AdminConfig groups maintainer panel limits (TASK-031).
type AdminConfig struct {
	RateLimitPerSecond float64
	RateLimitBurst     int
}

// ReportsConfig groups anonymous report limits (TASK-032).
type ReportsConfig struct {
	RateLimitPerSecond float64
	RateLimitBurst     int
}

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

	Auth    AuthConfig
	Push    PushConfig
	Privacy PrivacyConfig
	Admin   AdminConfig
	Reports ReportsConfig
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
	if err := loadPush(&cfg); err != nil {
		return Config{}, err
	}
	if err := loadPrivacy(&cfg); err != nil {
		return Config{}, err
	}
	if err := loadAdmin(&cfg); err != nil {
		return Config{}, err
	}
	if err := loadReports(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func loadAuth(cfg *Config) error {
	cfg.Auth.GoogleOAuthClientID = strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_CLIENT_ID"))
	cfg.Auth.GoogleOAuthClientSecret = strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"))
	cfg.Auth.GoogleOAuthRedirectURL = strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_REDIRECT_URL"))
	cfg.Auth.SessionCookieSecret = strings.TrimSpace(os.Getenv("SESSION_COOKIE_SECRET"))
	cfg.Auth.MaintainerAllowlistEmails = normalizeEmails(envCSV("MAINTAINER_ALLOWLIST"))
	cfg.Auth.PostLoginRedirect = strings.TrimSpace(os.Getenv("AUTH_POST_LOGIN_REDIRECT"))
	if cfg.Auth.PostLoginRedirect == "" {
		cfg.Auth.PostLoginRedirect = "/"
	}

	sessionTTLHours, err := envInt("SESSION_TTL_HOURS", 720) // 30 days
	if err != nil {
		return err
	}
	if sessionTTLHours <= 0 {
		return fmt.Errorf("SESSION_TTL_HOURS must be positive, got %d", sessionTTLHours)
	}
	cfg.Auth.SessionTTL = time.Duration(sessionTTLHours) * time.Hour

	cookieSecure, err := envBool("AUTH_COOKIE_SECURE", strings.HasPrefix(cfg.PublicBaseURL, "https://"))
	if err != nil {
		return err
	}
	cfg.Auth.CookieSecure = cookieSecure

	authRate, err := envFloat("AUTH_RATE_LIMIT_PER_SECOND", 1)
	if err != nil {
		return err
	}
	if authRate <= 0 {
		return fmt.Errorf("AUTH_RATE_LIMIT_PER_SECOND must be positive, got %v", authRate)
	}
	cfg.Auth.RateLimitPerSecond = authRate

	authBurst, err := envInt("AUTH_RATE_LIMIT_BURST", 5)
	if err != nil {
		return err
	}
	if authBurst <= 0 {
		return fmt.Errorf("AUTH_RATE_LIMIT_BURST must be positive, got %d", authBurst)
	}
	cfg.Auth.RateLimitBurst = authBurst

	anyOAuth := cfg.Auth.GoogleOAuthClientID != "" || cfg.Auth.GoogleOAuthClientSecret != "" ||
		cfg.Auth.GoogleOAuthRedirectURL != "" || cfg.Auth.SessionCookieSecret != ""
	allOAuth := cfg.Auth.GoogleOAuthClientID != "" && cfg.Auth.GoogleOAuthClientSecret != "" &&
		cfg.Auth.SessionCookieSecret != ""

	if anyOAuth && !allOAuth {
		return fmt.Errorf("incomplete Google OAuth config: set GOOGLE_OAUTH_CLIENT_ID, GOOGLE_OAUTH_CLIENT_SECRET, and SESSION_COOKIE_SECRET together (or leave all empty)")
	}
	if allOAuth && len(cfg.Auth.SessionCookieSecret) < 32 {
		return fmt.Errorf("SESSION_COOKIE_SECRET must be at least 32 characters")
	}
	if allOAuth {
		if cfg.PublicBaseURL == "" {
			return fmt.Errorf("PUBLIC_BASE_URL is required when Google OAuth is enabled (CSRF origin binding)")
		}
		if cfg.Auth.GoogleOAuthRedirectURL == "" {
			cfg.Auth.GoogleOAuthRedirectURL = cfg.PublicBaseURL + "/api/v1/auth/google/callback"
		}
		cfg.Auth.Enabled = true
	}
	return nil
}

func loadPush(cfg *Config) error {
	cfg.Push.VAPIDPublicKey = strings.TrimSpace(os.Getenv("VAPID_PUBLIC_KEY"))
	cfg.Push.VAPIDPrivateKey = strings.TrimSpace(os.Getenv("VAPID_PRIVATE_KEY"))
	cfg.Push.VAPIDSubject = strings.TrimSpace(os.Getenv("VAPID_SUBJECT"))
	if cfg.Push.VAPIDSubject == "" {
		cfg.Push.VAPIDSubject = "mailto:ops@centraldojogo.local"
	}

	any := cfg.Push.VAPIDPublicKey != "" || cfg.Push.VAPIDPrivateKey != ""
	all := cfg.Push.VAPIDPublicKey != "" && cfg.Push.VAPIDPrivateKey != ""
	if any && !all {
		return fmt.Errorf("incomplete Web Push config: set VAPID_PUBLIC_KEY and VAPID_PRIVATE_KEY together (or leave both empty)")
	}
	if all {
		cfg.Push.Enabled = true
	}
	return nil
}

func loadPrivacy(cfg *Config) error {
	days, err := envInt("ANALYTICS_RETENTION_DAYS", 90)
	if err != nil {
		return err
	}
	if days <= 0 {
		return fmt.Errorf("ANALYTICS_RETENTION_DAYS must be positive, got %d", days)
	}
	cfg.Privacy.AnalyticsRetentionDays = days
	return nil
}

func loadAdmin(cfg *Config) error {
	rate, err := envFloat("ADMIN_RATE_LIMIT_PER_SECOND", 2)
	if err != nil {
		return err
	}
	if rate <= 0 {
		return fmt.Errorf("ADMIN_RATE_LIMIT_PER_SECOND must be positive, got %v", rate)
	}
	cfg.Admin.RateLimitPerSecond = rate

	burst, err := envInt("ADMIN_RATE_LIMIT_BURST", 10)
	if err != nil {
		return err
	}
	if burst <= 0 {
		return fmt.Errorf("ADMIN_RATE_LIMIT_BURST must be positive, got %d", burst)
	}
	cfg.Admin.RateLimitBurst = burst
	return nil
}

func loadReports(cfg *Config) error {
	rate, err := envFloat("REPORTS_RATE_LIMIT_PER_SECOND", 0.5)
	if err != nil {
		return err
	}
	if rate <= 0 {
		return fmt.Errorf("REPORTS_RATE_LIMIT_PER_SECOND must be positive, got %v", rate)
	}
	cfg.Reports.RateLimitPerSecond = rate

	burst, err := envInt("REPORTS_RATE_LIMIT_BURST", 3)
	if err != nil {
		return err
	}
	if burst <= 0 {
		return fmt.Errorf("REPORTS_RATE_LIMIT_BURST must be positive, got %d", burst)
	}
	cfg.Reports.RateLimitBurst = burst
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

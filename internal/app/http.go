// Package app is the HTTP composition root: it wires stores, feature services,
// and Register() mounts without growing the platform HTTP kernel (ADR 0002).
package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/features/admin"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/auth"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/clubs"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/matches"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/preferences"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/privacy"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/push"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/reports"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/search"
	"github.com/IcaroAguiar/central-do-jogo/internal/jobs"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/config"
	httpplatform "github.com/IcaroAguiar/central-do-jogo/internal/platform/http"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/http/ratelimit"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/render"
	"github.com/IcaroAguiar/central-do-jogo/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewHTTPHandler wires persistence and features into the application router.
func NewHTTPHandler(cfg config.Config, pool *pgxpool.Pool) (http.Handler, error) {
	clubStore := store.NewClubStore(pool)
	matchStore := store.NewMatchStore(pool)
	broadcastStore := store.NewBroadcastStore(pool)
	lineupStore := store.NewLineupStore(pool)
	newsStore := store.NewNewsStore(pool)
	userStore := store.NewUserStore(pool)
	prefsStore := store.NewPreferencesStore(pool)
	pushStore := store.NewPushStore(pool)
	analyticsStore := store.NewAnalyticsStore(pool)
	auditStore := store.NewAuditStore(pool)
	reportStore := store.NewReportStore(pool)
	healthStore := jobs.NewHealthStore(pool)
	matchActionStore := store.NewMatchActionStore(pool)

	searchSvc := search.NewService(clubStore, matchStore)
	clubsSvc := clubs.NewService(clubStore, matchStore, time.Now)
	matchesSvc := matches.NewService(matchStore, broadcastStore, lineupStore, newsStore)
	authSvc := newAuthService(cfg, userStore)
	prefsSvc := preferences.NewService(authSvc, prefsStore, clubStore, time.Now)
	pushSvc := push.NewService(authSvc, pushStore, cfg.Push.Enabled, cfg.Push.VAPIDPublicKey, time.Now)
	privacySvc := privacy.NewService(authSvc, userStore, prefsStore, analyticsStore, cfg.Privacy.AnalyticsRetentionDays, time.Now)
	adminSvc := admin.NewService(authSvc, healthStore, matchStore, matchActionStore, auditStore, time.Now)
	reportsSvc := reports.NewService(authSvc, reportStore, time.Now)

	searchLimiter := ratelimit.New(cfg.SearchRateLimitPerSecond, cfg.SearchRateLimitBurst)
	if err := searchLimiter.SetTrustedProxies(cfg.TrustedProxyCIDRs); err != nil {
		return nil, fmt.Errorf("trusted proxies: %w", err)
	}
	authLimiter := ratelimit.New(cfg.Auth.RateLimitPerSecond, cfg.Auth.RateLimitBurst)
	if err := authLimiter.SetTrustedProxies(cfg.TrustedProxyCIDRs); err != nil {
		return nil, fmt.Errorf("auth trusted proxies: %w", err)
	}
	adminLimiter := ratelimit.New(cfg.Admin.RateLimitPerSecond, cfg.Admin.RateLimitBurst)
	if err := adminLimiter.SetTrustedProxies(cfg.TrustedProxyCIDRs); err != nil {
		return nil, fmt.Errorf("admin trusted proxies: %w", err)
	}
	reportsLimiter := ratelimit.New(cfg.Reports.RateLimitPerSecond, cfg.Reports.RateLimitBurst)
	if err := reportsLimiter.SetTrustedProxies(cfg.TrustedProxyCIDRs); err != nil {
		return nil, fmt.Errorf("reports trusted proxies: %w", err)
	}
	privacyLimiter := ratelimit.New(cfg.Privacy.EventsRateLimitPerSecond, cfg.Privacy.EventsRateLimitBurst)
	if err := privacyLimiter.SetTrustedProxies(cfg.TrustedProxyCIDRs); err != nil {
		return nil, fmt.Errorf("privacy trusted proxies: %w", err)
	}

	var ssr httpplatform.SSRHandlers
	if cfg.SSREnabled {
		renderer, err := render.New()
		if err != nil {
			return nil, fmt.Errorf("build renderer: %w", err)
		}
		ssr.Home = clubs.NewHomeSSRHandler(clubsSvc, renderer, cfg.PublicBaseURL)
		ssr.Club = clubs.NewClubSSRHandler(clubsSvc, renderer, cfg.PublicBaseURL, time.Now)
		ssr.Match = matches.NewSSRHandler(matchesSvc, renderer, cfg.PublicBaseURL)
	}

	return httpplatform.NewRouter(httpplatform.Options{
		StaticDir: cfg.StaticDir,
		SSR:       ssr,
		RegisterAPI: func(mux *http.ServeMux) {
			search.Register(mux, searchSvc, searchLimiter)
			clubs.Register(mux, clubsSvc)
			matches.Register(mux, matchesSvc)
			auth.Register(mux, authSvc, authLimiter)
			preferences.Register(mux, prefsSvc)
			push.Register(mux, pushSvc)
			privacy.Register(mux, privacySvc, privacyLimiter)
			admin.Register(mux, adminSvc, adminLimiter)
			reports.Register(mux, reportsSvc, reportsLimiter)
		},
	}), nil
}

func newAuthService(cfg config.Config, users *store.UserStore) *auth.Service {
	allow := make(map[string]struct{}, len(cfg.Auth.MaintainerAllowlistEmails))
	for _, email := range cfg.Auth.MaintainerAllowlistEmails {
		allow[email] = struct{}{}
	}
	authCfg := auth.Config{
		Enabled:           cfg.Auth.Enabled,
		SessionSecret:     []byte(cfg.Auth.SessionCookieSecret),
		SessionTTL:        cfg.Auth.SessionTTL,
		CookieSecure:      cfg.Auth.CookieSecure,
		PublicBaseURL:     cfg.PublicBaseURL,
		MaintainerEmails:  allow,
		PostLoginRedirect: auth.SafeRelativePath(cfg.Auth.PostLoginRedirect),
	}
	var provider auth.Provider
	if cfg.Auth.Enabled {
		provider = &auth.GoogleProvider{
			ClientID:     cfg.Auth.GoogleOAuthClientID,
			ClientSecret: cfg.Auth.GoogleOAuthClientSecret,
			RedirectURL:  cfg.Auth.GoogleOAuthRedirectURL,
		}
	}
	return auth.NewService(users, provider, authCfg, time.Now)
}

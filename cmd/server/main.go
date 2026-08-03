package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/features/auth"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/clubs"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/matches"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/preferences"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/push"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/search"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/config"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/database"
	httpplatform "github.com/IcaroAguiar/central-do-jogo/internal/platform/http"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/http/ratelimit"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/logging"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/render"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server exited", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := logging.NewJSON(slog.LevelInfo)
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.OpenPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}
	logger.Info("database migrations applied")

	router, err := buildRouter(cfg, pool)
	if err != nil {
		return fmt.Errorf("build router: %w", err)
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("listening", "addr", cfg.HTTPAddr, "auth_enabled", cfg.AuthEnabled)
		if serveErr := server.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errCh <- serveErr
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
			return fmt.Errorf("shutdown: %w", shutdownErr)
		}
		return <-errCh
	case serveErr := <-errCh:
		return serveErr
	}
}

// buildRouter wires the pgx pool through the read stores, feature services,
// and HTTP handlers into the top-level router (GOAL-004 / GOAL-005).
func buildRouter(cfg config.Config, pool *pgxpool.Pool) (http.Handler, error) {
	clubStore := store.NewClubStore(pool)
	matchStore := store.NewMatchStore(pool)
	broadcastStore := store.NewBroadcastStore(pool)
	lineupStore := store.NewLineupStore(pool)
	newsStore := store.NewNewsStore(pool)
	userStore := store.NewUserStore(pool)
	prefsStore := store.NewPreferencesStore(pool)

	searchSvc := search.NewService(clubStore, matchStore)
	clubsSvc := clubs.NewService(clubStore, matchStore, time.Now)
	matchesSvc := matches.NewService(matchStore, broadcastStore, lineupStore, newsStore)

	deps := httpplatform.Dependencies{
		Search:      search.NewHandler(searchSvc),
		Club:        clubs.NewDetailHandler(clubsSvc),
		ClubMatches: clubs.NewMatchesHandler(clubsSvc),
		Match:       matches.NewDetailHandler(matchesSvc),
	}

	authSvc := buildAuthService(cfg, userStore)
	authHandlers := auth.NewHandlers(authSvc)
	deps.AuthGoogleStart = authHandlers.StartGoogle()
	deps.AuthGoogleCallback = authHandlers.CallbackGoogle()
	deps.AuthMe = authHandlers.Me()
	deps.AuthLogout = authHandlers.Logout()

	prefsSvc := preferences.NewService(authSvc, prefsStore, clubStore, time.Now)
	prefsHandlers := preferences.NewHandlers(prefsSvc)
	deps.PreferencesGet = prefsHandlers.Get()
	deps.PreferencesPut = prefsHandlers.Put()

	pushStore := store.NewPushStore(pool)
	pushSvc := push.NewService(authSvc, pushStore, cfg.PushEnabled, cfg.VAPIDPublicKey, time.Now)
	pushHandlers := push.NewHandlers(pushSvc)
	deps.PushVAPIDPublicKey = pushHandlers.VAPIDPublicKey()
	deps.PushSubscriptionsList = pushHandlers.List()
	deps.PushSubscribe = pushHandlers.Subscribe()
	deps.PushUnsubscribe = pushHandlers.Unsubscribe()

	if cfg.SSREnabled {
		renderer, err := render.New()
		if err != nil {
			return nil, fmt.Errorf("build renderer: %w", err)
		}
		deps.HomeSSR = clubs.NewHomeSSRHandler(clubsSvc, renderer, cfg.PublicBaseURL)
		deps.ClubSSR = clubs.NewClubSSRHandler(clubsSvc, renderer, cfg.PublicBaseURL, time.Now)
		deps.MatchSSR = matches.NewSSRHandler(matchesSvc, renderer, cfg.PublicBaseURL)
	}

	limiter := ratelimit.New(cfg.SearchRateLimitPerSecond, cfg.SearchRateLimitBurst)
	if err := limiter.SetTrustedProxies(cfg.TrustedProxyCIDRs); err != nil {
		return nil, fmt.Errorf("trusted proxies: %w", err)
	}
	authLimiter := ratelimit.New(cfg.AuthRateLimitPerSecond, cfg.AuthRateLimitBurst)
	if err := authLimiter.SetTrustedProxies(cfg.TrustedProxyCIDRs); err != nil {
		return nil, fmt.Errorf("auth trusted proxies: %w", err)
	}

	return httpplatform.NewRouter(httpplatform.Options{
		StaticDir:       cfg.StaticDir,
		Deps:            deps,
		RateLimiter:     limiter,
		AuthRateLimiter: authLimiter,
	}), nil
}

func buildAuthService(cfg config.Config, users *store.UserStore) *auth.Service {
	allow := make(map[string]struct{}, len(cfg.MaintainerAllowlistEmails))
	for _, email := range cfg.MaintainerAllowlistEmails {
		allow[email] = struct{}{}
	}
	authCfg := auth.Config{
		Enabled:           cfg.AuthEnabled,
		SessionSecret:     []byte(cfg.SessionCookieSecret),
		SessionTTL:        cfg.SessionTTL,
		CookieSecure:      cfg.AuthCookieSecure,
		PublicBaseURL:     cfg.PublicBaseURL,
		MaintainerEmails:  allow,
		PostLoginRedirect: auth.SafeRelativePath(cfg.AuthPostLoginRedirect),
	}
	var provider auth.Provider
	if cfg.AuthEnabled {
		provider = &auth.GoogleProvider{
			ClientID:     cfg.GoogleOAuthClientID,
			ClientSecret: cfg.GoogleOAuthClientSecret,
			RedirectURL:  cfg.GoogleOAuthRedirectURL,
		}
	}
	return auth.NewService(users, provider, authCfg, time.Now)
}

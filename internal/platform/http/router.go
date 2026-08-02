package httpplatform

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/platform/http/ratelimit"
)

// Dependencies wires feature HTTP handlers into the router without the
// router needing to know about domain/store/feature packages directly.
// Any handler left nil is simply not registered.
type Dependencies struct {
	// Search handles GET /api/v1/search. Wrapped with RateLimiter when set.
	Search http.Handler
	// Club handles GET /api/v1/clubs/{slug}.
	Club http.Handler
	// ClubMatches handles GET /api/v1/clubs/{slug}/matches.
	ClubMatches http.Handler
	// Match handles GET /api/v1/matches/{slug}.
	Match http.Handler

	// AuthGoogleStart handles GET /api/v1/auth/google/start.
	AuthGoogleStart http.Handler
	// AuthGoogleCallback handles GET /api/v1/auth/google/callback.
	AuthGoogleCallback http.Handler
	// AuthMe handles GET /api/v1/auth/me.
	AuthMe http.Handler
	// AuthLogout handles POST /api/v1/auth/logout.
	AuthLogout http.Handler

	// HomeSSR, ClubSSR, and MatchSSR render server-side HTML for "/",
	// "/clubes/{slug}", and "/jogos/{slug}" respectively (PAT-004). When a
	// handler is nil, that route falls through to the SPA shell.
	HomeSSR  http.Handler
	ClubSSR  http.Handler
	MatchSSR http.Handler
}

// Options configures the HTTP router.
type Options struct {
	StaticDir string
	Now       func() time.Time
	Deps      Dependencies
	// RateLimiter, when set, guards GET /api/v1/search (SEC-001).
	RateLimiter *ratelimit.Limiter
	// AuthRateLimiter, when set, guards OAuth start and callback (SEC-001).
	AuthRateLimiter *ratelimit.Limiter
}

// NewRouter builds the application ServeMux. API routes are registered as
// exact/one-segment patterns, which net/http's ServeMux always prefers over
// the "/" catch-all used for the SPA shell, so API routes take priority.
func NewRouter(opts Options) http.Handler {
	if opts.Now == nil {
		opts.Now = time.Now
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":    "ok",
			"checkedAt": opts.Now().UTC().Format(time.RFC3339),
		})
	})

	registerAPIRoutes(mux, opts)
	registerSSRRoutes(mux, opts.Deps)

	if opts.StaticDir != "" {
		fileServer := http.FileServer(http.Dir(opts.StaticDir))
		mux.Handle("GET /", spaHandler(opts.StaticDir, fileServer))
	}

	return mux
}

func registerAPIRoutes(mux *http.ServeMux, opts Options) {
	if opts.Deps.Search != nil {
		search := opts.Deps.Search
		if opts.RateLimiter != nil {
			search = ratelimit.Middleware(opts.RateLimiter, "GET /api/v1/search", writeRateLimitedError)(search)
		}
		mux.Handle("GET /api/v1/search", search)
	}
	if opts.Deps.Club != nil {
		mux.Handle("GET /api/v1/clubs/{slug}", opts.Deps.Club)
	}
	if opts.Deps.ClubMatches != nil {
		mux.Handle("GET /api/v1/clubs/{slug}/matches", opts.Deps.ClubMatches)
	}
	if opts.Deps.Match != nil {
		mux.Handle("GET /api/v1/matches/{slug}", opts.Deps.Match)
	}
	if opts.Deps.AuthGoogleStart != nil {
		h := opts.Deps.AuthGoogleStart
		if opts.AuthRateLimiter != nil {
			h = ratelimit.Middleware(opts.AuthRateLimiter, "GET /api/v1/auth/google/start", writeRateLimitedError)(h)
		}
		mux.Handle("GET /api/v1/auth/google/start", h)
	}
	if opts.Deps.AuthGoogleCallback != nil {
		h := opts.Deps.AuthGoogleCallback
		if opts.AuthRateLimiter != nil {
			h = ratelimit.Middleware(opts.AuthRateLimiter, "GET /api/v1/auth/google/callback", writeRateLimitedError)(h)
		}
		mux.Handle("GET /api/v1/auth/google/callback", h)
	}
	if opts.Deps.AuthMe != nil {
		mux.Handle("GET /api/v1/auth/me", opts.Deps.AuthMe)
	}
	if opts.Deps.AuthLogout != nil {
		mux.Handle("POST /api/v1/auth/logout", opts.Deps.AuthLogout)
	}
}

func registerSSRRoutes(mux *http.ServeMux, deps Dependencies) {
	if deps.HomeSSR != nil {
		// "/{$}" matches only the exact root path, letting the SPA "/"
		// catch-all continue to serve every other client-side route.
		mux.Handle("GET /{$}", deps.HomeSSR)
	}
	if deps.ClubSSR != nil {
		mux.Handle("GET /clubes/{slug}", deps.ClubSSR)
	}
	if deps.MatchSSR != nil {
		mux.Handle("GET /jogos/{slug}", deps.MatchSSR)
	}
}

func writeRateLimitedError(w http.ResponseWriter, _ *http.Request) {
	WriteError(w, http.StatusTooManyRequests, "rate_limited", "too many requests, try again shortly")
}

func spaHandler(staticDir string, fileServer http.Handler) http.Handler {
	root := http.Dir(staticDir)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && !strings.Contains(filepath.Base(r.URL.Path), ".") {
			rel := strings.TrimPrefix(filepath.Clean("/"+r.URL.Path), "/")
			if rel != "" && rel != "." {
				f, err := root.Open(rel)
				if err != nil {
					http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
					return
				}
				_ = f.Close()
			}
		}
		fileServer.ServeHTTP(w, r)
	})
}

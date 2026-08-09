package auth

import (
	"net/http"

	httpplatform "github.com/IcaroAguiar/central-do-jogo/internal/platform/http"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/http/ratelimit"
)

// Register mounts auth API routes onto mux.
func Register(mux *http.ServeMux, svc *Service, limiter *ratelimit.Limiter) {
	h := NewHandlers(svc)
	start := h.StartGoogle()
	callback := h.CallbackGoogle()
	if limiter != nil {
		start = ratelimit.Middleware(limiter, "GET /api/v1/auth/google/start", httpplatform.WriteRateLimited)(start)
		callback = ratelimit.Middleware(limiter, "GET /api/v1/auth/google/callback", httpplatform.WriteRateLimited)(callback)
	}
	mux.Handle("GET /api/v1/auth/google/start", start)
	mux.Handle("GET /api/v1/auth/google/callback", callback)
	mux.Handle("GET /api/v1/auth/me", h.Me())
	mux.Handle("POST /api/v1/auth/logout", h.Logout())
}

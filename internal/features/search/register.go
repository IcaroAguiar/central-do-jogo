package search

import (
	"net/http"

	httpplatform "github.com/IcaroAguiar/central-do-jogo/internal/platform/http"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/http/ratelimit"
)

// Register mounts search API routes onto mux.
func Register(mux *http.ServeMux, svc *Service, limiter *ratelimit.Limiter) {
	h := NewHandler(svc)
	if limiter != nil {
		h = ratelimit.Middleware(limiter, "GET /api/v1/search", httpplatform.WriteRateLimited)(h)
	}
	mux.Handle("GET /api/v1/search", h)
}

package admin

import (
	"net/http"

	httpplatform "github.com/IcaroAguiar/central-do-jogo/internal/platform/http"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/http/ratelimit"
)

// Register mounts admin API routes onto mux.
func Register(mux *http.ServeMux, svc *Service, limiter *ratelimit.Limiter) {
	h := NewHandlers(svc)
	action := h.MatchAction()
	if limiter != nil {
		action = ratelimit.Middleware(limiter, "POST /api/v1/admin/matches/{slug}/actions", httpplatform.WriteRateLimited)(action)
	}
	mux.Handle("GET /api/v1/admin/sources/health", h.SourceHealth())
	mux.Handle("GET /api/v1/admin/matches/at-risk", h.AtRiskMatches())
	mux.Handle("GET /api/v1/admin/audit", h.AuditList())
	mux.Handle("POST /api/v1/admin/matches/{slug}/actions", action)
}

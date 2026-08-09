package reports

import (
	"net/http"

	httpplatform "github.com/IcaroAguiar/central-do-jogo/internal/platform/http"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/http/ratelimit"
)

// Register mounts public report intake and maintainer queue routes.
func Register(mux *http.ServeMux, svc *Service, limiter *ratelimit.Limiter) {
	h := NewHandlers(svc)
	create := h.Create()
	if limiter != nil {
		create = ratelimit.Middleware(limiter, "POST /api/v1/reports", httpplatform.WriteRateLimited)(create)
	}
	mux.Handle("POST /api/v1/reports", create)
	mux.Handle("GET /api/v1/admin/reports", h.ListOpen())
	mux.Handle("POST /api/v1/admin/reports/{id}/review", h.Review())
}

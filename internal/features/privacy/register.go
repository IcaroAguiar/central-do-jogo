package privacy

import (
	"net/http"

	httpplatform "github.com/IcaroAguiar/central-do-jogo/internal/platform/http"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/http/ratelimit"
)

// Register mounts privacy API routes onto mux.
func Register(mux *http.ServeMux, svc *Service, limiter *ratelimit.Limiter) {
	h := NewHandlers(svc)
	mux.Handle("GET /api/v1/privacy/export", h.Export())
	mux.Handle("DELETE /api/v1/privacy/account", h.DeleteAccount())
	events := h.RecordEvent()
	if limiter != nil {
		events = ratelimit.Middleware(limiter, "POST /api/v1/privacy/events", httpplatform.WriteRateLimited)(events)
	}
	mux.Handle("POST /api/v1/privacy/events", events)
}

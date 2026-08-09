package privacy

import "net/http"

// Register mounts privacy API routes onto mux.
func Register(mux *http.ServeMux, svc *Service) {
	h := NewHandlers(svc)
	mux.Handle("GET /api/v1/privacy/export", h.Export())
	mux.Handle("DELETE /api/v1/privacy/account", h.DeleteAccount())
	mux.Handle("POST /api/v1/privacy/events", h.RecordEvent())
}

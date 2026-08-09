package preferences

import "net/http"

// Register mounts preferences API routes onto mux.
func Register(mux *http.ServeMux, svc *Service) {
	h := NewHandlers(svc)
	mux.Handle("GET /api/v1/preferences", h.Get())
	mux.Handle("PUT /api/v1/preferences", h.Put())
}

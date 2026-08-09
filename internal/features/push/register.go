package push

import "net/http"

// Register mounts push API routes onto mux.
func Register(mux *http.ServeMux, svc *Service) {
	h := NewHandlers(svc)
	mux.Handle("GET /api/v1/push/vapid-public-key", h.VAPIDPublicKey())
	mux.Handle("GET /api/v1/push/subscriptions", h.List())
	mux.Handle("POST /api/v1/push/subscriptions", h.Subscribe())
	mux.Handle("DELETE /api/v1/push/subscriptions", h.Unsubscribe())
}

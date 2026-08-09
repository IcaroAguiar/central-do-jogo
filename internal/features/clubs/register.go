package clubs

import "net/http"

// Register mounts club API routes onto mux.
func Register(mux *http.ServeMux, svc *Service) {
	mux.Handle("GET /api/v1/clubs/{slug}", NewDetailHandler(svc))
	mux.Handle("GET /api/v1/clubs/{slug}/matches", NewMatchesHandler(svc))
}

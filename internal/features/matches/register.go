package matches

import "net/http"

// Register mounts match API routes onto mux.
func Register(mux *http.ServeMux, svc *Service) {
	mux.Handle("GET /api/v1/matches/{slug}", NewDetailHandler(svc))
}

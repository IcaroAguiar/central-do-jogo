package matches

import (
	"net/http"

	httpplatform "github.com/IcaroAguiar/central-do-jogo/internal/platform/http"
)

// NewDetailHandler builds the GET /api/v1/matches/{slug} handler.
func NewDetailHandler(svc *Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		if slug == "" {
			httpplatform.WriteError(w, http.StatusBadRequest, "invalid_slug", "match slug is required")
			return
		}

		detail, err := svc.GetDetail(r.Context(), slug)
		if err != nil {
			httpplatform.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to load match")
			return
		}
		if detail == nil {
			httpplatform.WriteError(w, http.StatusNotFound, "match_not_found", "match not found")
			return
		}
		httpplatform.WriteJSON(w, http.StatusOK, detail)
	})
}

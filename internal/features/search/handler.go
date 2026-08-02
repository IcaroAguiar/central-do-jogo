package search

import (
	"net/http"
	"strings"

	httpplatform "github.com/IcaroAguiar/central-do-jogo/internal/platform/http"
)

// NewHandler builds the GET /api/v1/search handler.
func NewHandler(svc *Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := strings.TrimSpace(r.URL.Query().Get("q"))
		if len(query) < MinQueryLength {
			httpplatform.WriteError(w, http.StatusBadRequest, "invalid_query", "query parameter q is required")
			return
		}
		if len(query) > MaxQueryLength {
			query = query[:MaxQueryLength]
		}

		resp, err := svc.Search(r.Context(), query)
		if err != nil {
			httpplatform.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to search")
			return
		}
		httpplatform.WriteJSON(w, http.StatusOK, resp)
	})
}

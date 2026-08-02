package clubs

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	httpplatform "github.com/IcaroAguiar/central-do-jogo/internal/platform/http"
)

var errInvalidSeason = errors.New("season must be a valid year")

// NewDetailHandler builds the GET /api/v1/clubs/{slug} handler.
func NewDetailHandler(svc *Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		if slug == "" {
			httpplatform.WriteError(w, http.StatusBadRequest, "invalid_slug", "club slug is required")
			return
		}

		detail, err := svc.GetDetail(r.Context(), slug)
		if err != nil {
			httpplatform.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to load club")
			return
		}
		if detail == nil {
			httpplatform.WriteError(w, http.StatusNotFound, "club_not_found", "club not found")
			return
		}
		httpplatform.WriteJSON(w, http.StatusOK, detail)
	})
}

// NewMatchesHandler builds the GET /api/v1/clubs/{slug}/matches handler.
func NewMatchesHandler(svc *Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		if slug == "" {
			httpplatform.WriteError(w, http.StatusBadRequest, "invalid_slug", "club slug is required")
			return
		}

		rng, err := ParseRange(r.URL.Query().Get("range"))
		if err != nil {
			httpplatform.WriteError(w, http.StatusBadRequest, "invalid_range", err.Error())
			return
		}

		season, err := parseSeason(r.URL.Query().Get("season"), time.Now().Year())
		if err != nil {
			httpplatform.WriteError(w, http.StatusBadRequest, "invalid_season", err.Error())
			return
		}

		resp, err := svc.GetMatches(r.Context(), slug, rng, season)
		if err != nil {
			httpplatform.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to load club matches")
			return
		}
		if resp == nil {
			httpplatform.WriteError(w, http.StatusNotFound, "club_not_found", "club not found")
			return
		}
		httpplatform.WriteJSON(w, http.StatusOK, resp)
	})
}

func parseSeason(raw string, defaultSeason int) (int, error) {
	if raw == "" {
		return defaultSeason, nil
	}
	season, err := strconv.Atoi(raw)
	if err != nil || season < 1900 {
		return 0, errInvalidSeason
	}
	return season, nil
}

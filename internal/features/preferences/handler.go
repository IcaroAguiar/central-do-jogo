package preferences

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/features/auth"
	httpplatform "github.com/IcaroAguiar/central-do-jogo/internal/platform/http"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/logging"
)

// Response is the JSON payload for preference reads/writes.
type Response struct {
	PrimaryClubSlug   *string  `json:"primaryClubSlug"`
	FavoriteClubSlugs []string `json:"favoriteClubSlugs"`
	UpdatedAt         *string  `json:"updatedAt,omitempty"`
}

// UpdateRequest is the JSON body for PUT /api/v1/preferences.
type UpdateRequest struct {
	PrimaryClubSlug   *string  `json:"primaryClubSlug"`
	FavoriteClubSlugs []string `json:"favoriteClubSlugs"`
}

// Handlers exposes the preferences HTTP surface.
type Handlers struct {
	svc *Service
}

// NewHandlers builds preferences HTTP handlers.
func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Get handles GET /api/v1/preferences.
func (h *Handlers) Get() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		view, err := h.svc.Get(r.Context(), cookieValue(r, auth.SessionCookieName))
		if writePrefsError(w, r, err) {
			return
		}
		httpplatform.WriteJSON(w, http.StatusOK, toResponse(view))
	})
}

// Put handles PUT /api/v1/preferences.
func (h *Handlers) Put() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !httpplatform.OriginAllowed(r, h.svc.PublicBaseURL()) {
			httpplatform.WriteError(w, http.StatusForbidden, "csrf_rejected", "preferences origin not allowed")
			return
		}
		var body UpdateRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&body); err != nil {
			httpplatform.WriteError(w, http.StatusBadRequest, "invalid_body", "invalid preferences payload")
			return
		}
		if body.FavoriteClubSlugs == nil {
			body.FavoriteClubSlugs = []string{}
		}
		view, err := h.svc.Put(r.Context(), cookieValue(r, auth.SessionCookieName), Update(body))
		if writePrefsError(w, r, err) {
			return
		}
		httpplatform.WriteJSON(w, http.StatusOK, toResponse(view))
	})
}

func writePrefsError(w http.ResponseWriter, r *http.Request, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, ErrAuthDisabled):
		httpplatform.WriteError(w, http.StatusServiceUnavailable, "auth_disabled", "google oauth is not configured")
	case errors.Is(err, ErrUnauthorized):
		httpplatform.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
	case errors.Is(err, ErrInvalidClub):
		httpplatform.WriteError(w, http.StatusBadRequest, "invalid_club", "unknown club slug in preferences")
	case errors.Is(err, ErrTooManyFavorites):
		httpplatform.WriteError(w, http.StatusBadRequest, "too_many_favorites", "favorite club limit exceeded")
	default:
		logging.FromContext(r.Context()).Error("preferences", slog.String("error", err.Error()))
		httpplatform.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to load preferences")
	}
	return true
}

func toResponse(view View) Response {
	favorites := view.FavoriteClubSlugs
	if favorites == nil {
		favorites = []string{}
	}
	resp := Response{
		PrimaryClubSlug:   view.PrimaryClubSlug,
		FavoriteClubSlugs: favorites,
	}
	if view.UpdatedAt != nil {
		s := view.UpdatedAt.UTC().Format(time.RFC3339)
		resp.UpdatedAt = &s
	}
	return resp
}

func cookieValue(r *http.Request, name string) string {
	return httpplatform.CookieValue(r, name)
}

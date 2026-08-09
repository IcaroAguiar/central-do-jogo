package privacy

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	httpplatform "github.com/IcaroAguiar/central-do-jogo/internal/platform/http"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/logging"
)

// Handlers exposes the privacy HTTP surface.
type Handlers struct {
	svc *Service
}

// NewHandlers builds privacy HTTP handlers.
func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

type exportResponse struct {
	ExportedAt      string                    `json:"exportedAt"`
	User            exportUserResponse        `json:"user"`
	Preferences     exportPreferencesResponse `json:"preferences"`
	AnalyticsEvents []exportEventResponse     `json:"analyticsEvents"`
}

type exportUserResponse struct {
	ID          string  `json:"id"`
	Provider    string  `json:"provider"`
	Email       string  `json:"email"`
	DisplayName string  `json:"displayName"`
	Role        string  `json:"role"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
	LastLoginAt *string `json:"lastLoginAt,omitempty"`
}

type exportPreferencesResponse struct {
	PrimaryClubSlug   *string  `json:"primaryClubSlug"`
	FavoriteClubSlugs []string `json:"favoriteClubSlugs"`
	UpdatedAt         *string  `json:"updatedAt,omitempty"`
}

type exportEventResponse struct {
	ID         string         `json:"id"`
	EventType  string         `json:"eventType"`
	CreatedAt  string         `json:"createdAt"`
	Properties map[string]any `json:"properties"`
}

type analyticsCreateRequest struct {
	AnonymousID   string         `json:"anonymousId"`
	EventType     string         `json:"eventType"`
	ConsentToLink bool           `json:"consentToLink"`
	Properties    map[string]any `json:"properties"`
}

// Export handles GET /api/v1/privacy/export.
func (h *Handlers) Export() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		exp, err := h.svc.ExportAccount(r.Context(), httpplatform.CookieValue(r, httpplatform.SessionCookieName))
		if writePrivacyError(w, r, err) {
			return
		}
		httpplatform.WriteJSON(w, http.StatusOK, toExportResponse(exp))
	})
}

// DeleteAccount handles DELETE /api/v1/privacy/account.
func (h *Handlers) DeleteAccount() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !httpplatform.OriginAllowed(r, h.svc.PublicBaseURL()) {
			httpplatform.WriteError(w, http.StatusForbidden, "csrf_rejected", "privacy origin not allowed")
			return
		}
		err := h.svc.DeleteAccount(r.Context(), httpplatform.CookieValue(r, httpplatform.SessionCookieName))
		if writePrivacyError(w, r, err) {
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     httpplatform.SessionCookieName,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   -1,
			Expires:  time.Unix(0, 0).UTC(),
		})
		w.WriteHeader(http.StatusNoContent)
	})
}

// RecordEvent handles POST /api/v1/privacy/events.
func (h *Handlers) RecordEvent() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		var body analyticsCreateRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&body); err != nil {
			httpplatform.WriteError(w, http.StatusBadRequest, "invalid_body", "invalid analytics payload")
			return
		}
		err := h.svc.RecordEvent(r.Context(), httpplatform.CookieValue(r, httpplatform.SessionCookieName), AnalyticsInput(body))
		if writePrivacyError(w, r, err) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

func writePrivacyError(w http.ResponseWriter, r *http.Request, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, ErrAuthDisabled):
		httpplatform.WriteError(w, http.StatusServiceUnavailable, "auth_disabled", "google oauth is not configured")
	case errors.Is(err, ErrUnauthorized):
		httpplatform.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
	case errors.Is(err, ErrInvalidEvent):
		httpplatform.WriteError(w, http.StatusBadRequest, "invalid_event", "invalid analytics event")
	default:
		logging.FromContext(r.Context()).Error("privacy", slog.String("error", err.Error()))
		httpplatform.WriteError(w, http.StatusInternalServerError, "internal_error", "privacy request failed")
	}
	return true
}

func toExportResponse(exp Export) exportResponse {
	resp := exportResponse{
		ExportedAt: exp.ExportedAt.UTC().Format(time.RFC3339),
		User: exportUserResponse{
			ID:          exp.User.ID,
			Provider:    exp.User.Provider,
			Email:       exp.User.Email,
			DisplayName: exp.User.DisplayName,
			Role:        exp.User.Role,
			CreatedAt:   exp.User.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:   exp.User.UpdatedAt.UTC().Format(time.RFC3339),
		},
		Preferences: exportPreferencesResponse{
			PrimaryClubSlug:   exp.Preferences.PrimaryClubSlug,
			FavoriteClubSlugs: exp.Preferences.FavoriteClubSlugs,
		},
		AnalyticsEvents: make([]exportEventResponse, 0, len(exp.AnalyticsEvents)),
	}
	if exp.User.LastLoginAt != nil {
		s := exp.User.LastLoginAt.UTC().Format(time.RFC3339)
		resp.User.LastLoginAt = &s
	}
	if exp.Preferences.UpdatedAt != nil {
		s := exp.Preferences.UpdatedAt.UTC().Format(time.RFC3339)
		resp.Preferences.UpdatedAt = &s
	}
	if resp.Preferences.FavoriteClubSlugs == nil {
		resp.Preferences.FavoriteClubSlugs = []string{}
	}
	for _, ev := range exp.AnalyticsEvents {
		props := ev.Properties
		if props == nil {
			props = map[string]any{}
		}
		resp.AnalyticsEvents = append(resp.AnalyticsEvents, exportEventResponse{
			ID:         ev.ID,
			EventType:  ev.EventType,
			CreatedAt:  ev.CreatedAt.UTC().Format(time.RFC3339),
			Properties: props,
		})
	}
	return resp
}

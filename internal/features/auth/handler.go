package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	httpplatform "github.com/IcaroAguiar/central-do-jogo/internal/platform/http"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/logging"
)

// MeResponse is the JSON payload for GET /api/v1/auth/me.
type MeResponse struct {
	Authenticated bool   `json:"authenticated"`
	AuthEnabled   bool   `json:"authEnabled"`
	Email         string `json:"email,omitempty"`
	DisplayName   string `json:"displayName,omitempty"`
	Role          string `json:"role,omitempty"`
}

// Handlers exposes the auth HTTP surface.
type Handlers struct {
	svc *Service
}

// NewHandlers builds auth HTTP handlers.
func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// StartGoogle handles GET /api/v1/auth/google/start.
func (h *Handlers) StartGoogle() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		redirectTo, err := h.svc.StartURL(w)
		if errors.Is(err, ErrAuthDisabled) {
			httpplatform.WriteError(w, http.StatusServiceUnavailable, "auth_disabled", "google oauth is not configured")
			return
		}
		if err != nil {
			logging.FromContext(r.Context()).Error("auth start", slog.String("error", err.Error()))
			httpplatform.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to start login")
			return
		}
		http.Redirect(w, r, redirectTo, http.StatusFound)
	})
}

// CallbackGoogle handles GET /api/v1/auth/google/callback.
func (h *Handlers) CallbackGoogle() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if errParam := strings.TrimSpace(r.URL.Query().Get("error")); errParam != "" {
			httpplatform.WriteError(w, http.StatusBadRequest, "oauth_error", "google login was denied or failed")
			return
		}
		code := strings.TrimSpace(r.URL.Query().Get("code"))
		state := strings.TrimSpace(r.URL.Query().Get("state"))
		if code == "" || state == "" {
			httpplatform.WriteError(w, http.StatusBadRequest, "invalid_callback", "missing code or state")
			return
		}
		_, err := h.svc.CompleteLogin(r.Context(), w, r, code, state)
		if errors.Is(err, ErrAuthDisabled) {
			httpplatform.WriteError(w, http.StatusServiceUnavailable, "auth_disabled", "google oauth is not configured")
			return
		}
		if errors.Is(err, ErrInvalidState) {
			httpplatform.WriteError(w, http.StatusBadRequest, "invalid_state", "oauth state mismatch")
			return
		}
		if errors.Is(err, ErrEmailUnverified) {
			httpplatform.WriteError(w, http.StatusForbidden, "email_unverified", "verified email is required")
			return
		}
		if err != nil {
			logging.FromContext(r.Context()).Error("auth callback", slog.String("error", err.Error()))
			httpplatform.WriteError(w, http.StatusBadGateway, "oauth_exchange_failed", "failed to complete google login")
			return
		}
		http.Redirect(w, r, h.svc.PostLoginRedirect(), http.StatusFound)
	})
}

// Me handles GET /api/v1/auth/me.
func (h *Handlers) Me() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		resp := MeResponse{AuthEnabled: h.svc.Enabled()}
		user, err := h.svc.CurrentUser(r.Context(), r)
		if err != nil {
			logging.FromContext(r.Context()).Error("auth me", slog.String("error", err.Error()))
			httpplatform.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to resolve session")
			return
		}
		if user != nil {
			resp.Authenticated = true
			resp.Email = user.Email
			resp.DisplayName = user.DisplayName
			resp.Role = string(user.Role)
		}
		httpplatform.WriteJSON(w, http.StatusOK, resp)
	})
}

// Logout handles POST /api/v1/auth/logout.
func (h *Handlers) Logout() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !originAllowed(r, h.svc.cfg.PublicBaseURL) {
			httpplatform.WriteError(w, http.StatusForbidden, "csrf_rejected", "logout origin not allowed")
			return
		}
		if err := h.svc.Logout(r.Context(), w, r); err != nil {
			logging.FromContext(r.Context()).Error("auth logout", slog.String("error", err.Error()))
			httpplatform.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to logout")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

func originAllowed(r *http.Request, publicBaseURL string) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		// Same-origin navigational POSTs may omit Origin; require a matching Referer.
		referer := strings.TrimSpace(r.Header.Get("Referer"))
		if referer == "" {
			return publicBaseURL == ""
		}
		origin = referer
	}
	if publicBaseURL == "" {
		return true
	}
	base, err := url.Parse(publicBaseURL)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return false
	}
	got, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return strings.EqualFold(got.Scheme, base.Scheme) && strings.EqualFold(got.Host, base.Host)
}

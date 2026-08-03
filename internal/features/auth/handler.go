package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

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

// Handlers exposes the auth HTTP surface and owns cookie transport.
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
		result, err := h.svc.StartLogin()
		if errors.Is(err, ErrAuthDisabled) {
			httpplatform.WriteError(w, http.StatusServiceUnavailable, "auth_disabled", "google oauth is not configured")
			return
		}
		if err != nil {
			logging.FromContext(r.Context()).Error("auth start", slog.String("error", err.Error()))
			httpplatform.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to start login")
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     OAuthStateCookieName,
			Value:    result.SignedStateCookie,
			Path:     "/",
			HttpOnly: true,
			Secure:   h.svc.CookieSecure(),
			SameSite: http.SameSiteLaxMode,
			MaxAge:   int(h.svc.OAuthStateTTL().Seconds()),
		})
		http.Redirect(w, r, result.AuthURL, http.StatusFound)
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
		var signedState string
		if c, err := r.Cookie(OAuthStateCookieName); err == nil {
			signedState = c.Value
		}
		result, err := h.svc.CompleteLogin(r.Context(), code, state, signedState)
		clearCookie(w, OAuthStateCookieName, h.svc.CookieSecure())
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
		http.SetCookie(w, &http.Cookie{
			Name:     SessionCookieName,
			Value:    result.SessionToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   h.svc.CookieSecure(),
			SameSite: http.SameSiteLaxMode,
			Expires:  result.ExpiresAt,
			MaxAge:   int(h.svc.SessionTTL().Seconds()),
		})
		http.Redirect(w, r, h.svc.PostLoginRedirect(), http.StatusFound)
	})
}

// Me handles GET /api/v1/auth/me.
func (h *Handlers) Me() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		resp := MeResponse{AuthEnabled: h.svc.Enabled()}
		user, err := h.svc.CurrentUser(r.Context(), cookieValue(r, SessionCookieName))
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
		if !httpplatform.OriginAllowed(r, h.svc.PublicBaseURL()) {
			httpplatform.WriteError(w, http.StatusForbidden, "csrf_rejected", "logout origin not allowed")
			return
		}
		token := cookieValue(r, SessionCookieName)
		clearCookie(w, SessionCookieName, h.svc.CookieSecure())
		if err := h.svc.Logout(r.Context(), token); err != nil {
			logging.FromContext(r.Context()).Error("auth logout", slog.String("error", err.Error()))
			httpplatform.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to logout")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

func cookieValue(r *http.Request, name string) string {
	return httpplatform.CookieValue(r, name)
}

func clearCookie(w http.ResponseWriter, name string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0).UTC(),
	})
}

package push

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

// Handlers exposes the push HTTP surface.
type Handlers struct {
	svc *Service
}

// NewHandlers builds push HTTP handlers.
func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// VAPIDPublicKey handles GET /api/v1/push/vapid-public-key.
func (h *Handlers) VAPIDPublicKey() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		key, err := h.svc.PublicKey()
		if writePushError(w, r, err) {
			return
		}
		httpplatform.WriteJSON(w, http.StatusOK, map[string]any{
			"publicKey": key,
			"enabled":   true,
		})
	})
}

// List handles GET /api/v1/push/subscriptions.
func (h *Handlers) List() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		subs, err := h.svc.ListSubscriptions(r.Context(), cookieValue(r, auth.SessionCookieName))
		if writePushError(w, r, err) {
			return
		}
		items := make([]map[string]string, 0, len(subs))
		for _, sub := range subs {
			items = append(items, map[string]string{
				"endpoint":  sub.Endpoint,
				"createdAt": sub.CreatedAt.UTC().Format(time.RFC3339),
			})
		}
		httpplatform.WriteJSON(w, http.StatusOK, map[string]any{"subscriptions": items})
	})
}

type subscribeBody struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

// Subscribe handles POST /api/v1/push/subscriptions.
func (h *Handlers) Subscribe() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !httpplatform.OriginAllowed(r, h.svc.PublicBaseURL()) {
			httpplatform.WriteError(w, http.StatusForbidden, "csrf_rejected", "push origin not allowed")
			return
		}
		var body subscribeBody
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&body); err != nil {
			httpplatform.WriteError(w, http.StatusBadRequest, "invalid_body", "invalid push subscription payload")
			return
		}
		sub, err := h.svc.Subscribe(r.Context(), cookieValue(r, auth.SessionCookieName), SubscribeInput{
			Endpoint:  body.Endpoint,
			P256dh:    body.Keys.P256dh,
			Auth:      body.Keys.Auth,
			UserAgent: r.UserAgent(),
		})
		if writePushError(w, r, err) {
			return
		}
		httpplatform.WriteJSON(w, http.StatusOK, map[string]any{
			"endpoint":  sub.Endpoint,
			"createdAt": sub.CreatedAt.UTC().Format(time.RFC3339),
		})
	})
}

type unsubscribeBody struct {
	Endpoint string `json:"endpoint"`
}

// Unsubscribe handles DELETE /api/v1/push/subscriptions.
func (h *Handlers) Unsubscribe() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !httpplatform.OriginAllowed(r, h.svc.PublicBaseURL()) {
			httpplatform.WriteError(w, http.StatusForbidden, "csrf_rejected", "push origin not allowed")
			return
		}
		var body unsubscribeBody
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&body); err != nil {
			httpplatform.WriteError(w, http.StatusBadRequest, "invalid_body", "invalid unsubscribe payload")
			return
		}
		err := h.svc.Unsubscribe(r.Context(), cookieValue(r, auth.SessionCookieName), body.Endpoint)
		if writePushError(w, r, err) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

func writePushError(w http.ResponseWriter, r *http.Request, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, ErrPushDisabled):
		httpplatform.WriteError(w, http.StatusServiceUnavailable, "push_disabled", "web push is not configured")
	case errors.Is(err, ErrAuthDisabled):
		httpplatform.WriteError(w, http.StatusServiceUnavailable, "auth_disabled", "google oauth is not configured")
	case errors.Is(err, ErrUnauthorized):
		httpplatform.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
	case errors.Is(err, ErrInvalidSubscription):
		httpplatform.WriteError(w, http.StatusBadRequest, "invalid_subscription", "invalid push subscription")
	default:
		logging.FromContext(r.Context()).Error("push", slog.String("error", err.Error()))
		httpplatform.WriteError(w, http.StatusInternalServerError, "internal_error", "push request failed")
	}
	return true
}

func cookieValue(r *http.Request, name string) string {
	c, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return c.Value
}

package push

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

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

// vapidPublicKeyResponse is GET /api/v1/push/vapid-public-key.
type vapidPublicKeyResponse struct {
	PublicKey string `json:"publicKey"`
	Enabled   bool   `json:"enabled"`
}

// subscriptionItem is one public subscription summary (no secrets).
type subscriptionItem struct {
	Endpoint  string `json:"endpoint"`
	CreatedAt string `json:"createdAt"`
}

// listSubscriptionsResponse is GET /api/v1/push/subscriptions.
type listSubscriptionsResponse struct {
	Subscriptions []subscriptionItem `json:"subscriptions"`
}

// subscribeRequest is POST /api/v1/push/subscriptions.
type subscribeRequest struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

// subscribeResponse is the created/updated subscription summary.
type subscribeResponse struct {
	Endpoint  string `json:"endpoint"`
	CreatedAt string `json:"createdAt"`
}

// unsubscribeRequest is DELETE /api/v1/push/subscriptions.
type unsubscribeRequest struct {
	Endpoint string `json:"endpoint"`
}

// VAPIDPublicKey handles GET /api/v1/push/vapid-public-key.
func (h *Handlers) VAPIDPublicKey() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		key, err := h.svc.PublicKey()
		if writePushError(w, r, err) {
			return
		}
		httpplatform.WriteJSON(w, http.StatusOK, vapidPublicKeyResponse{
			PublicKey: key,
			Enabled:   true,
		})
	})
}

// List handles GET /api/v1/push/subscriptions.
func (h *Handlers) List() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		subs, err := h.svc.ListSubscriptions(r.Context(), httpplatform.CookieValue(r, httpplatform.SessionCookieName))
		if writePushError(w, r, err) {
			return
		}
		items := make([]subscriptionItem, 0, len(subs))
		for _, sub := range subs {
			items = append(items, subscriptionItem{
				Endpoint:  sub.Endpoint,
				CreatedAt: sub.CreatedAt.UTC().Format(time.RFC3339),
			})
		}
		httpplatform.WriteJSON(w, http.StatusOK, listSubscriptionsResponse{Subscriptions: items})
	})
}

// Subscribe handles POST /api/v1/push/subscriptions.
func (h *Handlers) Subscribe() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !httpplatform.OriginAllowed(r, h.svc.PublicBaseURL()) {
			httpplatform.WriteError(w, http.StatusForbidden, "csrf_rejected", "push origin not allowed")
			return
		}
		var body subscribeRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&body); err != nil {
			httpplatform.WriteError(w, http.StatusBadRequest, "invalid_body", "invalid push subscription payload")
			return
		}
		sub, err := h.svc.Subscribe(r.Context(), httpplatform.CookieValue(r, httpplatform.SessionCookieName), SubscribeInput{
			Endpoint:  body.Endpoint,
			P256dh:    body.Keys.P256dh,
			Auth:      body.Keys.Auth,
			UserAgent: r.UserAgent(),
		})
		if writePushError(w, r, err) {
			return
		}
		httpplatform.WriteJSON(w, http.StatusOK, subscribeResponse{
			Endpoint:  sub.Endpoint,
			CreatedAt: sub.CreatedAt.UTC().Format(time.RFC3339),
		})
	})
}

// Unsubscribe handles DELETE /api/v1/push/subscriptions.
func (h *Handlers) Unsubscribe() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !httpplatform.OriginAllowed(r, h.svc.PublicBaseURL()) {
			httpplatform.WriteError(w, http.StatusForbidden, "csrf_rejected", "push origin not allowed")
			return
		}
		var body unsubscribeRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&body); err != nil {
			httpplatform.WriteError(w, http.StatusBadRequest, "invalid_body", "invalid unsubscribe payload")
			return
		}
		err := h.svc.Unsubscribe(r.Context(), httpplatform.CookieValue(r, httpplatform.SessionCookieName), body.Endpoint)
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
	case errors.Is(err, ErrEndpointOwned):
		httpplatform.WriteError(w, http.StatusConflict, "endpoint_owned", "push endpoint owned by another user")
	default:
		logging.FromContext(r.Context()).Error("push", slog.String("error", err.Error()))
		httpplatform.WriteError(w, http.StatusInternalServerError, "internal_error", "push request failed")
	}
	return true
}

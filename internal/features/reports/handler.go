package reports

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	httpplatform "github.com/IcaroAguiar/central-do-jogo/internal/platform/http"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/logging"
)

// Handlers exposes the reports HTTP surface.
type Handlers struct {
	svc *Service
}

// NewHandlers builds reports HTTP handlers.
func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

type createRequest struct {
	ContextType string `json:"contextType"`
	ContextSlug string `json:"contextSlug"`
	Message     string `json:"message"`
}

type reportItem struct {
	ID          string `json:"id"`
	ContextType string `json:"contextType"`
	ContextSlug string `json:"contextSlug"`
	Message     string `json:"message"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
}

type reviewRequest struct {
	Status string `json:"status"`
}

// Create handles POST /api/v1/reports.
func (h *Handlers) Create() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		var body createRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&body); err != nil {
			httpplatform.WriteError(w, http.StatusBadRequest, "invalid_body", "invalid report payload")
			return
		}
		err := h.svc.Create(r.Context(), CreateInput{
			ContextType: body.ContextType,
			ContextSlug: body.ContextSlug,
			Message:     body.Message,
			ClientIP:    clientIP(r),
			UserAgent:   r.UserAgent(),
		})
		if writeReportError(w, r, err) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

// ListOpen handles GET /api/v1/admin/reports.
func (h *Handlers) ListOpen() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		rows, err := h.svc.ListOpen(r.Context(), httpplatform.CookieValue(r, httpplatform.SessionCookieName))
		if writeReportError(w, r, err) {
			return
		}
		items := make([]reportItem, 0, len(rows))
		for _, row := range rows {
			items = append(items, reportItem{
				ID: row.ID.String(), ContextType: row.ContextType, ContextSlug: row.ContextSlug,
				Message: row.Message, Status: row.Status, CreatedAt: row.CreatedAt.UTC().Format(time.RFC3339),
			})
		}
		httpplatform.WriteJSON(w, http.StatusOK, map[string]any{"reports": items})
	})
}

// Review handles POST /api/v1/admin/reports/{id}/review.
func (h *Handlers) Review() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !httpplatform.OriginAllowed(r, h.svc.PublicBaseURL()) {
			httpplatform.WriteError(w, http.StatusForbidden, "csrf_rejected", "reports origin not allowed")
			return
		}
		var body reviewRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&body); err != nil {
			httpplatform.WriteError(w, http.StatusBadRequest, "invalid_body", "invalid review payload")
			return
		}
		err := h.svc.Review(r.Context(), httpplatform.CookieValue(r, httpplatform.SessionCookieName), r.PathValue("id"), body.Status)
		if writeReportError(w, r, err) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

func writeReportError(w http.ResponseWriter, r *http.Request, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, ErrAuthDisabled):
		httpplatform.WriteError(w, http.StatusServiceUnavailable, "auth_disabled", "google oauth is not configured")
	case errors.Is(err, ErrUnauthorized):
		httpplatform.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
	case errors.Is(err, ErrForbidden):
		httpplatform.WriteError(w, http.StatusForbidden, "forbidden", "maintainer access required")
	case errors.Is(err, ErrInvalidInput):
		httpplatform.WriteError(w, http.StatusBadRequest, "invalid_input", "invalid report request")
	default:
		logging.FromContext(r.Context()).Error("reports", slog.String("error", err.Error()))
		httpplatform.WriteError(w, http.StatusInternalServerError, "internal_error", "report request failed")
	}
	return true
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return host
}

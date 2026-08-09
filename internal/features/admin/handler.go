package admin

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/IcaroAguiar/central-do-jogo/internal/jobs"
	httpplatform "github.com/IcaroAguiar/central-do-jogo/internal/platform/http"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/logging"
)

// Handlers exposes the admin HTTP surface.
type Handlers struct {
	svc *Service
}

// NewHandlers builds admin HTTP handlers.
func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

type healthItem struct {
	SourceID            string  `json:"sourceId"`
	LastSuccessAt       *string `json:"lastSuccessAt"`
	LastErrorAt         *string `json:"lastErrorAt"`
	LastError           string  `json:"lastError"`
	NextRunAt           string  `json:"nextRunAt"`
	ConsecutiveFailures int     `json:"consecutiveFailures"`
	UpdatedAt           string  `json:"updatedAt"`
}

type atRiskItem struct {
	Slug           string  `json:"slug"`
	Round          string  `json:"round"`
	HomeClub       string  `json:"homeClub"`
	AwayClub       string  `json:"awayClub"`
	KickoffAt      *string `json:"kickoffAt"`
	BroadcastState string  `json:"broadcastState"`
	LineupState    string  `json:"lineupState"`
	NewsState      string  `json:"newsState"`
}

type auditItem struct {
	ID         int64  `json:"id"`
	Actor      string `json:"actor"`
	Action     string `json:"action"`
	EntityType string `json:"entityType"`
	EntityID   string `json:"entityId"`
	Reason     string `json:"reason"`
	CreatedAt  string `json:"createdAt"`
}

type matchActionRequest struct {
	Action  string `json:"action"`
	Surface string `json:"surface"`
	Reason  string `json:"reason"`
	Value   string `json:"value"`
}

// SourceHealth handles GET /api/v1/admin/sources/health.
func (h *Handlers) SourceHealth() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		rows, err := h.svc.ListSourceHealth(r.Context(), httpplatform.CookieValue(r, httpplatform.SessionCookieName))
		if writeAdminError(w, r, err) {
			return
		}
		items := make([]healthItem, 0, len(rows))
		for _, row := range rows {
			items = append(items, toHealthItem(row))
		}
		httpplatform.WriteJSON(w, http.StatusOK, map[string]any{"sources": items})
	})
}

// AtRiskMatches handles GET /api/v1/admin/matches/at-risk.
func (h *Handlers) AtRiskMatches() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		rows, err := h.svc.ListAtRiskMatches(r.Context(), httpplatform.CookieValue(r, httpplatform.SessionCookieName))
		if writeAdminError(w, r, err) {
			return
		}
		items := make([]atRiskItem, 0, len(rows))
		for _, row := range rows {
			items = append(items, toAtRiskItem(row))
		}
		httpplatform.WriteJSON(w, http.StatusOK, map[string]any{"matches": items})
	})
}

// AuditList handles GET /api/v1/admin/audit.
func (h *Handlers) AuditList() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		rows, err := h.svc.ListAudit(
			r.Context(),
			httpplatform.CookieValue(r, httpplatform.SessionCookieName),
			r.URL.Query().Get("entityType"),
			r.URL.Query().Get("entityId"),
		)
		if writeAdminError(w, r, err) {
			return
		}
		items := make([]auditItem, 0, len(rows))
		for _, row := range rows {
			items = append(items, auditItem{
				ID: row.ID, Actor: row.Actor, Action: row.Action,
				EntityType: row.EntityType, EntityID: row.EntityID, Reason: row.Reason,
				CreatedAt: row.CreatedAt.UTC().Format(time.RFC3339),
			})
		}
		httpplatform.WriteJSON(w, http.StatusOK, map[string]any{"events": items})
	})
}

// MatchAction handles POST /api/v1/admin/matches/{slug}/actions.
func (h *Handlers) MatchAction() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !httpplatform.OriginAllowed(r, h.svc.PublicBaseURL()) {
			httpplatform.WriteError(w, http.StatusForbidden, "csrf_rejected", "admin origin not allowed")
			return
		}
		slug := r.PathValue("slug")
		var body matchActionRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&body); err != nil {
			httpplatform.WriteError(w, http.StatusBadRequest, "invalid_body", "invalid admin action payload")
			return
		}
		err := h.svc.ApplyMatchAction(r.Context(), httpplatform.CookieValue(r, httpplatform.SessionCookieName), slug, MatchAction(body))
		if writeAdminError(w, r, err) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

func writeAdminError(w http.ResponseWriter, r *http.Request, err error) bool {
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
	case errors.Is(err, ErrNotFound):
		httpplatform.WriteError(w, http.StatusNotFound, "not_found", "resource not found")
	case errors.Is(err, ErrInvalidInput):
		httpplatform.WriteError(w, http.StatusBadRequest, "invalid_input", "invalid admin request")
	default:
		logging.FromContext(r.Context()).Error("admin", slog.String("error", err.Error()))
		httpplatform.WriteError(w, http.StatusInternalServerError, "internal_error", "admin request failed")
	}
	return true
}

func toHealthItem(row jobs.SourceHealth) healthItem {
	item := healthItem{
		SourceID:            row.SourceID,
		LastError:           row.LastError,
		NextRunAt:           row.NextRunAt.UTC().Format(time.RFC3339),
		ConsecutiveFailures: row.ConsecutiveFailures,
		UpdatedAt:           row.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if row.LastSuccessAt != nil {
		s := row.LastSuccessAt.UTC().Format(time.RFC3339)
		item.LastSuccessAt = &s
	}
	if row.LastErrorAt != nil {
		s := row.LastErrorAt.UTC().Format(time.RFC3339)
		item.LastErrorAt = &s
	}
	return item
}

func toAtRiskItem(row domain.MatchRecord) atRiskItem {
	item := atRiskItem{
		Slug: row.Slug, Round: row.Round,
		HomeClub: row.HomeClub.Name, AwayClub: row.AwayClub.Name,
		BroadcastState: string(row.BroadcastState),
		LineupState:    string(row.LineupState),
		NewsState:      string(row.NewsState),
	}
	if row.KickoffAt != nil {
		s := row.KickoffAt.UTC().Format(time.RFC3339)
		item.KickoffAt = &s
	}
	return item
}

// Package admin implements the maintainer panel APIs (REQ-013, REQ-018, CON-006, TASK-031).
package admin

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
)

var (
	ErrAuthDisabled = domain.ErrAuthDisabled
	ErrUnauthorized = domain.ErrUnauthorized
	ErrForbidden    = domain.ErrForbidden
	ErrNotFound     = domain.ErrNotFound
	ErrInvalidInput = errors.New("invalid input")
)

// HealthLister lists source health rows.
type HealthLister interface {
	List(ctx context.Context) ([]domain.SourceHealth, error)
}

// MatchAdminStore loads matches for admin review.
type MatchAdminStore interface {
	GetBySlug(ctx context.Context, slug string) (*domain.MatchRecord, error)
	ListAtRisk(ctx context.Context, limit int) ([]domain.MatchRecord, error)
}

// MatchActionWriter applies override + surface + audit atomically.
type MatchActionWriter interface {
	ApplyMatchAction(ctx context.Context, in domain.MatchActionApply) error
}

// AuditRepository lists audit events.
type AuditRepository interface {
	ListRecent(ctx context.Context, entityType, entityID string, limit int) ([]domain.AuditEvent, error)
}

// Action kinds for match corrections.
const (
	ActionConfirm       = "confirm"
	ActionCorrect       = "correct"
	ActionMarkDivergent = "mark_divergent"
)

// MatchAction is a maintainer correction request.
type MatchAction struct {
	Action  string
	Surface string
	Reason  string
	Value   string
}

// Service orchestrates maintainer panel use cases.
type Service struct {
	gate    domain.MaintainerGate
	health  HealthLister
	matches MatchAdminStore
	writer  MatchActionWriter
	audit   AuditRepository
	now     func() time.Time
}

// NewService builds an admin service.
func NewService(
	gate domain.MaintainerGate,
	health HealthLister,
	matches MatchAdminStore,
	writer MatchActionWriter,
	audit AuditRepository,
	now func() time.Time,
) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{gate: gate, health: health, matches: matches, writer: writer, audit: audit, now: now}
}

// PublicBaseURL exposes the configured origin for CSRF checks.
func (s *Service) PublicBaseURL() string { return s.gate.PublicBaseURL() }

// ListSourceHealth returns operational source health (CON-006).
func (s *Service) ListSourceHealth(ctx context.Context, sessionToken string) ([]domain.SourceHealth, error) {
	if _, err := s.gate.RequireMaintainer(ctx, sessionToken); err != nil {
		return nil, err
	}
	return s.health.List(ctx)
}

// ListAtRiskMatches returns matches needing maintainer attention.
func (s *Service) ListAtRiskMatches(ctx context.Context, sessionToken string) ([]domain.MatchRecord, error) {
	if _, err := s.gate.RequireMaintainer(ctx, sessionToken); err != nil {
		return nil, err
	}
	return s.matches.ListAtRisk(ctx, 50)
}

// ListAudit returns recent audit events, optionally filtered.
func (s *Service) ListAudit(ctx context.Context, sessionToken, entityType, entityID string) ([]domain.AuditEvent, error) {
	if _, err := s.gate.RequireMaintainer(ctx, sessionToken); err != nil {
		return nil, err
	}
	return s.audit.ListRecent(ctx, entityType, entityID, 50)
}

// ApplyMatchAction confirms, corrects, or marks divergence on a match surface.
func (s *Service) ApplyMatchAction(ctx context.Context, sessionToken, slug string, action MatchAction) error {
	user, err := s.gate.RequireMaintainer(ctx, sessionToken)
	if err != nil {
		return err
	}
	action.Action = strings.TrimSpace(action.Action)
	action.Surface = strings.TrimSpace(action.Surface)
	action.Reason = strings.TrimSpace(action.Reason)
	action.Value = strings.TrimSpace(action.Value)
	if action.Reason == "" {
		return ErrInvalidInput
	}
	switch action.Surface {
	case "broadcast", "lineup", "news":
	default:
		return ErrInvalidInput
	}
	rec, err := s.matches.GetBySlug(ctx, slug)
	if err != nil {
		return err
	}
	if rec == nil {
		return ErrNotFound
	}

	beforeState := surfaceState(rec, action.Surface)
	var afterState domain.AvailabilityState
	var override *domain.MatchOverride
	switch action.Action {
	case ActionConfirm:
		afterState = domain.AvailabilityAvailable
	case ActionMarkDivergent:
		afterState = domain.AvailabilityDivergent
	case ActionCorrect:
		afterState = domain.AvailabilityAvailable
		if action.Value == "" {
			action.Value = string(domain.AvailabilityAvailable)
		}
		id, err := domain.NewID("ovr_")
		if err != nil {
			return err
		}
		override = &domain.MatchOverride{
			ID: id, MatchID: rec.ID, DataType: action.Surface, Field: "state",
			Value: action.Value, Justification: action.Reason, Actor: user.Email, AppliedAt: s.now().UTC(),
		}
	default:
		return ErrInvalidInput
	}

	beforeJSON, _ := json.Marshal(map[string]string{"state": string(beforeState)})
	afterJSON, _ := json.Marshal(map[string]string{"state": string(afterState), "value": action.Value})
	return s.writer.ApplyMatchAction(ctx, domain.MatchActionApply{
		MatchID: rec.ID, Surface: action.Surface, AfterState: afterState,
		Override: override,
		Audit: domain.AuditEvent{
			Actor: user.Email, Action: action.Action, EntityType: "match", EntityID: rec.ID.String(),
			Reason: action.Reason, BeforeJSON: beforeJSON, AfterJSON: afterJSON,
		},
		Now: s.now().UTC(),
	})
}

func surfaceState(rec *domain.MatchRecord, surface string) domain.AvailabilityState {
	switch surface {
	case "broadcast":
		return rec.BroadcastState
	case "lineup":
		return rec.LineupState
	case "news":
		return rec.NewsState
	default:
		return ""
	}
}

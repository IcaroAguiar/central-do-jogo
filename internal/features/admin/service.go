// Package admin implements the maintainer panel APIs (REQ-013, REQ-018, CON-006, TASK-031).
package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/IcaroAguiar/central-do-jogo/internal/jobs"
)

var (
	ErrAuthDisabled = errors.New("auth disabled")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
)

// SessionResolver looks up the current user from an opaque session token.
type SessionResolver interface {
	Enabled() bool
	CurrentUser(ctx context.Context, sessionToken string) (*domain.User, error)
	PublicBaseURL() string
}

// HealthLister lists source health rows.
type HealthLister interface {
	List(ctx context.Context) ([]jobs.SourceHealth, error)
}

// MatchAdminStore loads matches and updates surface states.
type MatchAdminStore interface {
	GetBySlug(ctx context.Context, slug string) (*domain.MatchRecord, error)
	ListAtRisk(ctx context.Context, limit int) ([]domain.MatchRecord, error)
	UpdateSurfaceState(ctx context.Context, matchID domain.ID, surface string, state domain.AvailabilityState, now time.Time) error
}

// OverrideRepository persists match overrides.
type OverrideRepository interface {
	Save(ctx context.Context, override domain.MatchOverride) (*domain.MatchOverride, error)
}

// AuditRepository appends and lists audit events.
type AuditRepository interface {
	Append(ctx context.Context, event domain.AuditEvent) (*domain.AuditEvent, error)
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
	sessions  SessionResolver
	health    HealthLister
	matches   MatchAdminStore
	overrides OverrideRepository
	audit     AuditRepository
	now       func() time.Time
}

// NewService builds an admin service.
func NewService(
	sessions SessionResolver,
	health HealthLister,
	matches MatchAdminStore,
	overrides OverrideRepository,
	audit AuditRepository,
	now func() time.Time,
) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{sessions: sessions, health: health, matches: matches, overrides: overrides, audit: audit, now: now}
}

// PublicBaseURL exposes the configured origin for CSRF checks.
func (s *Service) PublicBaseURL() string { return s.sessions.PublicBaseURL() }

// ListSourceHealth returns operational source health (CON-006).
func (s *Service) ListSourceHealth(ctx context.Context, sessionToken string) ([]jobs.SourceHealth, error) {
	if _, err := s.requireMaintainer(ctx, sessionToken); err != nil {
		return nil, err
	}
	return s.health.List(ctx)
}

// ListAtRiskMatches returns matches needing maintainer attention.
func (s *Service) ListAtRiskMatches(ctx context.Context, sessionToken string) ([]domain.MatchRecord, error) {
	if _, err := s.requireMaintainer(ctx, sessionToken); err != nil {
		return nil, err
	}
	return s.matches.ListAtRisk(ctx, 50)
}

// ListAudit returns recent audit events, optionally filtered.
func (s *Service) ListAudit(ctx context.Context, sessionToken, entityType, entityID string) ([]domain.AuditEvent, error) {
	if _, err := s.requireMaintainer(ctx, sessionToken); err != nil {
		return nil, err
	}
	return s.audit.ListRecent(ctx, entityType, entityID, 50)
}

// ApplyMatchAction confirms, corrects, or marks divergence on a match surface.
func (s *Service) ApplyMatchAction(ctx context.Context, sessionToken, slug string, action MatchAction) error {
	user, err := s.requireMaintainer(ctx, sessionToken)
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
		if _, err := s.overrides.Save(ctx, domain.MatchOverride{
			ID: id, MatchID: rec.ID, DataType: action.Surface, Field: "state",
			Value: action.Value, Justification: action.Reason, Actor: user.Email, AppliedAt: s.now().UTC(),
		}); err != nil {
			return err
		}
	default:
		return ErrInvalidInput
	}

	if err := s.matches.UpdateSurfaceState(ctx, rec.ID, action.Surface, afterState, s.now()); err != nil {
		return err
	}
	beforeJSON, _ := json.Marshal(map[string]string{"state": string(beforeState)})
	afterJSON, _ := json.Marshal(map[string]string{"state": string(afterState), "value": action.Value})
	_, err = s.audit.Append(ctx, domain.AuditEvent{
		Actor: user.Email, Action: action.Action, EntityType: "match", EntityID: rec.ID.String(),
		Reason: action.Reason, BeforeJSON: beforeJSON, AfterJSON: afterJSON,
	})
	return err
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

func (s *Service) requireMaintainer(ctx context.Context, sessionToken string) (*domain.User, error) {
	if !s.sessions.Enabled() {
		return nil, ErrAuthDisabled
	}
	user, err := s.sessions.CurrentUser(ctx, sessionToken)
	if err != nil {
		return nil, fmt.Errorf("resolve session: %w", err)
	}
	if user == nil {
		return nil, ErrUnauthorized
	}
	if user.Role != domain.RoleMaintainer {
		return nil, ErrForbidden
	}
	return user, nil
}

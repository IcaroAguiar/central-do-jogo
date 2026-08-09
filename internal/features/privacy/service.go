// Package privacy implements account export/delete and first-party analytics
// retention (REQ-019, REQ-020, TASK-030).
package privacy

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
)

// ErrAuthDisabled is returned when OAuth is not configured on this instance.
var ErrAuthDisabled = errors.New("auth disabled")

// ErrUnauthorized is returned when the request has no valid session.
var ErrUnauthorized = errors.New("unauthorized")

// ErrInvalidEvent is returned when an analytics payload fails validation.
var ErrInvalidEvent = errors.New("invalid analytics event")

// SessionResolver looks up the current user from an opaque session token.
type SessionResolver interface {
	Enabled() bool
	CurrentUser(ctx context.Context, sessionToken string) (*domain.User, error)
	PublicBaseURL() string
}

// UserRepository loads and deletes account rows.
type UserRepository interface {
	GetByID(ctx context.Context, id domain.ID) (*domain.User, error)
	Delete(ctx context.Context, id domain.ID) error
}

// PreferencesRepository loads preference rows for export.
type PreferencesRepository interface {
	GetByUserID(ctx context.Context, userID domain.ID) (*domain.UserPreferences, error)
}

// AnalyticsRepository persists and queries first-party events.
type AnalyticsRepository interface {
	Insert(ctx context.Context, event domain.AnalyticsEvent) error
	ListByUserID(ctx context.Context, userID domain.ID, limit int) ([]domain.AnalyticsEvent, error)
	DeleteBefore(ctx context.Context, cutoff time.Time) (int64, error)
}

// Export is the portable account snapshot (REQ-019). It never includes
// session tokens, cookies, or Push endpoint/auth material.
type Export struct {
	ExportedAt      time.Time
	User            ExportUser
	Preferences     ExportPreferences
	AnalyticsEvents []ExportAnalyticsEvent
}

// ExportUser is the safe account projection for export.
type ExportUser struct {
	ID          string
	Provider    string
	Email       string
	DisplayName string
	Role        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	LastLoginAt *time.Time
}

// ExportPreferences is the preference snapshot for export.
type ExportPreferences struct {
	PrimaryClubSlug   *string
	FavoriteClubSlugs []string
	UpdatedAt         *time.Time
}

// ExportAnalyticsEvent is a linked analytics row for export (no anonymous id).
type ExportAnalyticsEvent struct {
	ID         string
	EventType  string
	CreatedAt  time.Time
	Properties map[string]any
}

// AnalyticsInput is the validated write payload for POST /privacy/events.
type AnalyticsInput struct {
	AnonymousID   string
	EventType     string
	ConsentToLink bool
	Properties    map[string]any
}

// Service orchestrates privacy use cases.
type Service struct {
	sessions  SessionResolver
	users     UserRepository
	prefs     PreferencesRepository
	analytics AnalyticsRepository
	retention time.Duration
	now       func() time.Time
}

// NewService builds a privacy service.
func NewService(
	sessions SessionResolver,
	users UserRepository,
	prefs PreferencesRepository,
	analytics AnalyticsRepository,
	retentionDays int,
	now func() time.Time,
) *Service {
	if now == nil {
		now = time.Now
	}
	if retentionDays <= 0 {
		retentionDays = 90
	}
	return &Service{
		sessions:  sessions,
		users:     users,
		prefs:     prefs,
		analytics: analytics,
		retention: time.Duration(retentionDays) * 24 * time.Hour,
		now:       now,
	}
}

// PublicBaseURL exposes the configured origin for CSRF checks in handlers.
func (s *Service) PublicBaseURL() string {
	return s.sessions.PublicBaseURL()
}

// ExportAccount builds a JSON-serializable account snapshot for the session.
func (s *Service) ExportAccount(ctx context.Context, sessionToken string) (Export, error) {
	user, err := s.requireUser(ctx, sessionToken)
	if err != nil {
		return Export{}, err
	}
	prefs, err := s.prefs.GetByUserID(ctx, user.ID)
	if err != nil {
		return Export{}, err
	}
	events, err := s.analytics.ListByUserID(ctx, user.ID, 200)
	if err != nil {
		return Export{}, err
	}

	exp := Export{
		ExportedAt: s.now().UTC(),
		User: ExportUser{
			ID:          user.ID.String(),
			Provider:    user.Provider,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			Role:        string(user.Role),
			CreatedAt:   user.CreatedAt.UTC(),
			UpdatedAt:   user.UpdatedAt.UTC(),
		},
		Preferences: ExportPreferences{
			FavoriteClubSlugs: []string{},
		},
		AnalyticsEvents: make([]ExportAnalyticsEvent, 0, len(events)),
	}
	if user.LastLoginAt != nil {
		t := user.LastLoginAt.UTC()
		exp.User.LastLoginAt = &t
	}
	if prefs != nil {
		exp.Preferences.PrimaryClubSlug = prefs.PrimaryClubSlug
		if prefs.FavoriteClubSlugs != nil {
			exp.Preferences.FavoriteClubSlugs = prefs.FavoriteClubSlugs
		}
		t := prefs.UpdatedAt.UTC()
		exp.Preferences.UpdatedAt = &t
	}
	for _, ev := range events {
		props := ev.Properties
		if props == nil {
			props = map[string]any{}
		}
		exp.AnalyticsEvents = append(exp.AnalyticsEvents, ExportAnalyticsEvent{
			ID:         ev.ID.String(),
			EventType:  ev.EventType,
			CreatedAt:  ev.CreatedAt.UTC(),
			Properties: props,
		})
	}
	return exp, nil
}

// DeleteAccount permanently removes the authenticated account (REQ-019).
func (s *Service) DeleteAccount(ctx context.Context, sessionToken string) error {
	user, err := s.requireUser(ctx, sessionToken)
	if err != nil {
		return err
	}
	return s.users.Delete(ctx, user.ID)
}

// RecordEvent stores a first-party analytics event (REQ-020).
func (s *Service) RecordEvent(ctx context.Context, sessionToken string, input AnalyticsInput) error {
	anonymousID := strings.TrimSpace(input.AnonymousID)
	eventType := strings.TrimSpace(input.EventType)
	if len(anonymousID) < 8 || len(anonymousID) > 128 {
		return ErrInvalidEvent
	}
	if eventType == "" || len(eventType) > 64 {
		return ErrInvalidEvent
	}
	props := input.Properties
	if props == nil {
		props = map[string]any{}
	}

	var userID *domain.ID
	if input.ConsentToLink && sessionToken != "" && s.sessions.Enabled() {
		user, err := s.sessions.CurrentUser(ctx, sessionToken)
		if err != nil {
			return fmt.Errorf("resolve session: %w", err)
		}
		if user != nil {
			id := user.ID
			userID = &id
		}
	}

	id, err := domain.NewID("aev_")
	if err != nil {
		return err
	}
	return s.analytics.Insert(ctx, domain.AnalyticsEvent{
		ID:          id,
		AnonymousID: anonymousID,
		UserID:      userID,
		EventType:   eventType,
		Properties:  props,
		CreatedAt:   s.now().UTC(),
	})
}

// PurgeExpired deletes analytics rows older than the configured retention.
func (s *Service) PurgeExpired(ctx context.Context) (int64, error) {
	cutoff := s.now().UTC().Add(-s.retention)
	return s.analytics.DeleteBefore(ctx, cutoff)
}

func (s *Service) requireUser(ctx context.Context, sessionToken string) (*domain.User, error) {
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
	return user, nil
}

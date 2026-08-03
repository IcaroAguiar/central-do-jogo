// Package preferences implements account-backed club preference sync (REQ-006, TASK-028).
package preferences

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
)

const maxFavoriteClubs = 32

// ErrAuthDisabled is returned when OAuth is not configured on this instance.
var ErrAuthDisabled = errors.New("auth disabled")

// ErrUnauthorized is returned when the request has no valid session.
var ErrUnauthorized = errors.New("unauthorized")

// ErrInvalidClub is returned when a preference references an unknown club slug.
var ErrInvalidClub = errors.New("invalid club slug")

// ErrTooManyFavorites is returned when favoriteClubSlugs exceeds the allowed cap.
var ErrTooManyFavorites = errors.New("too many favorite clubs")

// SessionResolver looks up the current user from an opaque session token.
type SessionResolver interface {
	Enabled() bool
	CurrentUser(ctx context.Context, sessionToken string) (*domain.User, error)
	PublicBaseURL() string
}

// PreferencesRepository persists preference rows.
type PreferencesRepository interface {
	GetByUserID(ctx context.Context, userID domain.ID) (*domain.UserPreferences, error)
	Upsert(ctx context.Context, prefs domain.UserPreferences, now time.Time) (*domain.UserPreferences, error)
}

// ClubLookup validates that a club slug exists in the allowlisted catalog.
type ClubLookup interface {
	GetBySlug(ctx context.Context, slug string) (*domain.Club, error)
}

// Update is the validated write payload for PUT /api/v1/preferences.
type Update struct {
	PrimaryClubSlug   *string
	FavoriteClubSlugs []string
}

// View is the API-facing preference snapshot.
type View struct {
	PrimaryClubSlug   *string
	FavoriteClubSlugs []string
	UpdatedAt         *time.Time
}

// Service orchestrates authenticated preference reads and writes.
type Service struct {
	sessions SessionResolver
	prefs    PreferencesRepository
	clubs    ClubLookup
	now      func() time.Time
}

// NewService builds a preferences service.
func NewService(sessions SessionResolver, prefs PreferencesRepository, clubs ClubLookup, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{sessions: sessions, prefs: prefs, clubs: clubs, now: now}
}

// PublicBaseURL exposes the configured origin for CSRF checks in handlers.
func (s *Service) PublicBaseURL() string {
	return s.sessions.PublicBaseURL()
}

// Get returns the authenticated user's preferences (empty defaults when unset).
func (s *Service) Get(ctx context.Context, sessionToken string) (View, error) {
	user, err := s.requireUser(ctx, sessionToken)
	if err != nil {
		return View{}, err
	}
	row, err := s.prefs.GetByUserID(ctx, user.ID)
	if err != nil {
		return View{}, err
	}
	return viewFromRow(row), nil
}

// Put replaces the authenticated user's preferences after club validation.
func (s *Service) Put(ctx context.Context, sessionToken string, update Update) (View, error) {
	user, err := s.requireUser(ctx, sessionToken)
	if err != nil {
		return View{}, err
	}
	normalized, err := s.normalizeUpdate(ctx, update)
	if err != nil {
		return View{}, err
	}
	row, err := s.prefs.Upsert(ctx, domain.UserPreferences{
		UserID:            user.ID,
		PrimaryClubSlug:   normalized.PrimaryClubSlug,
		FavoriteClubSlugs: normalized.FavoriteClubSlugs,
	}, s.now())
	if err != nil {
		return View{}, err
	}
	return viewFromRow(row), nil
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

func (s *Service) normalizeUpdate(ctx context.Context, update Update) (Update, error) {
	var primary *string
	if update.PrimaryClubSlug != nil {
		slug := strings.TrimSpace(*update.PrimaryClubSlug)
		if slug == "" {
			primary = nil
		} else {
			if err := s.requireClub(ctx, slug); err != nil {
				return Update{}, err
			}
			primary = &slug
		}
	}

	favorites := make([]string, 0, len(update.FavoriteClubSlugs))
	seen := make(map[string]struct{}, len(update.FavoriteClubSlugs))
	for _, raw := range update.FavoriteClubSlugs {
		slug := strings.TrimSpace(raw)
		if slug == "" {
			continue
		}
		if _, ok := seen[slug]; ok {
			continue
		}
		if err := s.requireClub(ctx, slug); err != nil {
			return Update{}, err
		}
		seen[slug] = struct{}{}
		favorites = append(favorites, slug)
	}
	if len(favorites) > maxFavoriteClubs {
		return Update{}, ErrTooManyFavorites
	}
	return Update{PrimaryClubSlug: primary, FavoriteClubSlugs: favorites}, nil
}

func (s *Service) requireClub(ctx context.Context, slug string) error {
	club, err := s.clubs.GetBySlug(ctx, slug)
	if err != nil {
		return fmt.Errorf("lookup club %q: %w", slug, err)
	}
	if club == nil {
		return fmt.Errorf("%w: %s", ErrInvalidClub, slug)
	}
	return nil
}

func viewFromRow(row *domain.UserPreferences) View {
	if row == nil {
		return View{FavoriteClubSlugs: []string{}}
	}
	favorites := row.FavoriteClubSlugs
	if favorites == nil {
		favorites = []string{}
	}
	updated := row.UpdatedAt.UTC()
	return View{
		PrimaryClubSlug:   row.PrimaryClubSlug,
		FavoriteClubSlugs: favorites,
		UpdatedAt:         &updated,
	}
}

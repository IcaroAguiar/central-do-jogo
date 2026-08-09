package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PreferencesStore persists account-backed club preferences.
type PreferencesStore struct {
	pool *pgxpool.Pool
}

// NewPreferencesStore creates a preferences store backed by the provided pool.
func NewPreferencesStore(pool *pgxpool.Pool) *PreferencesStore {
	return &PreferencesStore{pool: pool}
}

const preferencesColumns = `user_id, primary_club_slug, favorite_club_slugs, updated_at`

func scanPreferences(row pgx.Row) (*domain.UserPreferences, error) {
	var p domain.UserPreferences
	var userID string
	var primary *string
	var favorites []string
	if err := row.Scan(&userID, &primary, &favorites, &p.UpdatedAt); err != nil {
		return nil, err
	}
	p.UserID = domain.ID(userID)
	p.PrimaryClubSlug = primary
	if favorites == nil {
		favorites = []string{}
	}
	p.FavoriteClubSlugs = favorites
	p.UpdatedAt = utc(p.UpdatedAt)
	return &p, nil
}

// GetByUserID returns preferences for the user, or nil when none are stored yet.
func (s *PreferencesStore) GetByUserID(ctx context.Context, userID domain.ID) (*domain.UserPreferences, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+preferencesColumns+` FROM user_preferences WHERE user_id = $1`, userID.String())
	prefs, err := scanPreferences(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get preferences: %w", err)
	}
	return prefs, nil
}

// Upsert replaces the user's preferences row.
func (s *PreferencesStore) Upsert(ctx context.Context, prefs domain.UserPreferences, now time.Time) (*domain.UserPreferences, error) {
	favorites := prefs.FavoriteClubSlugs
	if favorites == nil {
		favorites = []string{}
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO user_preferences (user_id, primary_club_slug, favorite_club_slugs, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET
			primary_club_slug = EXCLUDED.primary_club_slug,
			favorite_club_slugs = EXCLUDED.favorite_club_slugs,
			updated_at = EXCLUDED.updated_at
		RETURNING `+preferencesColumns,
		prefs.UserID.String(), prefs.PrimaryClubSlug, favorites, now.UTC(),
	)
	out, err := scanPreferences(row)
	if err != nil {
		return nil, fmt.Errorf("upsert preferences: %w", err)
	}
	return out, nil
}

// Package store implements read repositories over PostgreSQL for the public
// search, clubs, and matches features. Repositories return domain types (or
// small read-only records composed from them) and never leak SQL to callers.
package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ClubStore provides read access to clubs.
type ClubStore struct {
	pool *pgxpool.Pool
}

// NewClubStore creates a club store backed by the provided pgx pool.
func NewClubStore(pool *pgxpool.Pool) *ClubStore {
	return &ClubStore{pool: pool}
}

const clubColumns = `id, slug, name, short_name, aliases, created_at, updated_at`

func scanClub(row pgx.Row) (*domain.Club, error) {
	var c domain.Club
	var id, slug string
	if err := row.Scan(&id, &slug, &c.Name, &c.ShortName, &c.Aliases, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, err
	}
	c.ID = domain.ID(id)
	c.Slug = slug
	c.CreatedAt = utc(c.CreatedAt)
	c.UpdatedAt = utc(c.UpdatedAt)
	return &c, nil
}

// GetBySlug returns the club with the given slug, or nil if none exists.
func (s *ClubStore) GetBySlug(ctx context.Context, slug string) (*domain.Club, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+clubColumns+` FROM clubs WHERE slug = $1`, slug)
	club, err := scanClub(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get club by slug %q: %w", slug, err)
	}
	return club, nil
}

// Search returns clubs whose name, short name, slug, or aliases match the
// query (case-insensitive substring), ordered by name, limited to limit rows.
func (s *ClubStore) Search(ctx context.Context, query string, limit int) ([]domain.Club, error) {
	if limit <= 0 {
		limit = 10
	}
	pattern := "%" + query + "%"
	rows, err := s.pool.Query(ctx, `
		SELECT `+clubColumns+`
		FROM clubs
		WHERE name ILIKE $1
		   OR short_name ILIKE $1
		   OR slug ILIKE $1
		   OR EXISTS (SELECT 1 FROM unnest(aliases) AS alias WHERE alias ILIKE $1)
		ORDER BY name
		LIMIT $2
	`, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("search clubs: %w", err)
	}
	defer rows.Close()

	var clubs []domain.Club
	for rows.Next() {
		club, err := scanClub(rows)
		if err != nil {
			return nil, fmt.Errorf("scan club: %w", err)
		}
		clubs = append(clubs, *club)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("search clubs rows: %w", err)
	}
	return clubs, nil
}

// List returns all clubs ordered by name. Used for indexable listings (e.g. the home page).
func (s *ClubStore) List(ctx context.Context) ([]domain.Club, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+clubColumns+` FROM clubs ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list clubs: %w", err)
	}
	defer rows.Close()

	var clubs []domain.Club
	for rows.Next() {
		club, err := scanClub(rows)
		if err != nil {
			return nil, fmt.Errorf("scan club: %w", err)
		}
		clubs = append(clubs, *club)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list clubs rows: %w", err)
	}
	return clubs, nil
}

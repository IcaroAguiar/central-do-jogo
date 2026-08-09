package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OverrideStore persists versioned match overrides.
type OverrideStore struct {
	pool *pgxpool.Pool
}

// NewOverrideStore creates an override store backed by the provided pool.
func NewOverrideStore(pool *pgxpool.Pool) *OverrideStore {
	return &OverrideStore{pool: pool}
}

// Save inserts a new override version for the match/data_type/field triple.
// Version allocation runs inside a transaction with an advisory lock to avoid races.
func (s *OverrideStore) Save(ctx context.Context, override domain.MatchOverride) (*domain.MatchOverride, error) {
	var saved domain.MatchOverride
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		out, err := saveOverrideTx(ctx, tx, override)
		if err != nil {
			return err
		}
		saved = out
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &saved, nil
}

// FindActive returns the latest override for the match/data_type/field, or nil.
func (s *OverrideStore) FindActive(ctx context.Context, matchID domain.ID, dataType, field string) (*domain.MatchOverride, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, match_id, data_type, field, value, justification, actor, version, applied_at
		FROM match_overrides
		WHERE match_id = $1 AND data_type = $2 AND field = $3
		ORDER BY version DESC
		LIMIT 1
	`, matchID.String(), dataType, field)
	var o domain.MatchOverride
	var id, matchCol string
	if err := row.Scan(
		&id, &matchCol, &o.DataType, &o.Field, &o.Value, &o.Justification, &o.Actor, &o.Version, &o.AppliedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find active override: %w", err)
	}
	o.ID = domain.ID(id)
	o.MatchID = domain.ID(matchCol)
	o.AppliedAt = utc(o.AppliedAt)
	return &o, nil
}

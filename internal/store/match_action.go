package store

import (
	"context"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MatchActionStore applies maintainer match actions in one transaction.
type MatchActionStore struct {
	pool *pgxpool.Pool
}

// NewMatchActionStore creates a transactional match-action writer.
func NewMatchActionStore(pool *pgxpool.Pool) *MatchActionStore {
	return &MatchActionStore{pool: pool}
}

// ApplyMatchAction persists override (optional), surface state, and audit atomically.
func (s *MatchActionStore) ApplyMatchAction(ctx context.Context, in domain.MatchActionApply) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if in.Override != nil {
			if _, err := saveOverrideTx(ctx, tx, *in.Override); err != nil {
				return err
			}
		}
		if err := updateSurfaceStateTx(ctx, tx, in.MatchID, in.Surface, in.AfterState, in.Now); err != nil {
			return err
		}
		_, err := appendAuditTx(ctx, tx, in.Audit)
		return err
	})
}

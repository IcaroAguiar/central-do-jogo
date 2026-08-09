package store

import (
	"context"
	"fmt"
	"hash/fnv"
	"time"

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
			if err := saveOverrideTx(ctx, tx, *in.Override); err != nil {
				return err
			}
		}
		if err := updateSurfaceStateTx(ctx, tx, in.MatchID, in.Surface, in.AfterState, in.Now); err != nil {
			return err
		}
		return appendAuditTx(ctx, tx, in.Audit)
	})
}

func saveOverrideTx(ctx context.Context, tx pgx.Tx, override domain.MatchOverride) error {
	if override.Justification == "" {
		return fmt.Errorf("override must have a justification")
	}
	if override.Actor == "" {
		return fmt.Errorf("override must have an actor")
	}
	if err := lockOverrideKey(ctx, tx, override.MatchID.String(), override.DataType, override.Field); err != nil {
		return err
	}
	var nextVersion int
	if err := tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(version), 0) + 1
		FROM match_overrides
		WHERE match_id = $1 AND data_type = $2 AND field = $3
	`, override.MatchID.String(), override.DataType, override.Field).Scan(&nextVersion); err != nil {
		return fmt.Errorf("next override version: %w", err)
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO match_overrides (
			id, match_id, data_type, field, value, justification, actor, version, applied_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, override.ID.String(), override.MatchID.String(), override.DataType, override.Field,
		override.Value, override.Justification, override.Actor, nextVersion, override.AppliedAt.UTC())
	if err != nil {
		return fmt.Errorf("save override: %w", err)
	}
	return nil
}

func updateSurfaceStateTx(ctx context.Context, tx pgx.Tx, matchID domain.ID, surface string, state domain.AvailabilityState, now time.Time) error {
	var query string
	switch surface {
	case "broadcast":
		query = `UPDATE matches SET broadcast_state = $2, updated_at = $3 WHERE id = $1`
	case "lineup":
		query = `UPDATE matches SET lineup_state = $2, updated_at = $3 WHERE id = $1`
	case "news":
		query = `UPDATE matches SET news_state = $2, updated_at = $3 WHERE id = $1`
	default:
		return fmt.Errorf("unknown surface %q", surface)
	}
	tag, err := tx.Exec(ctx, query, matchID.String(), string(state), now.UTC())
	if err != nil {
		return fmt.Errorf("update surface state: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func appendAuditTx(ctx context.Context, tx pgx.Tx, event domain.AuditEvent) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO audit_events (actor, action, entity_type, entity_id, reason, before_json, after_json)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, event.Actor, event.Action, event.EntityType, event.EntityID, event.Reason, event.BeforeJSON, event.AfterJSON)
	if err != nil {
		return fmt.Errorf("append audit event: %w", err)
	}
	return nil
}

func lockOverrideKey(ctx context.Context, tx pgx.Tx, matchID, dataType, field string) error {
	h := fnv.New64a()
	_, _ = h.Write([]byte(matchID + "\x00" + dataType + "\x00" + field))
	key := int64(h.Sum64())
	_, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, key)
	if err != nil {
		return fmt.Errorf("lock override key: %w", err)
	}
	return nil
}

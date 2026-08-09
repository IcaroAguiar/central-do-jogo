package store

import (
	"context"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// writeDB is satisfied by pgx.Tx and *pgxpool.Pool for shared write helpers.
type writeDB interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func saveOverrideTx(ctx context.Context, db writeDB, override domain.MatchOverride) (domain.MatchOverride, error) {
	if override.Justification == "" {
		return domain.MatchOverride{}, fmt.Errorf("override must have a justification")
	}
	if override.Actor == "" {
		return domain.MatchOverride{}, fmt.Errorf("override must have an actor")
	}
	if err := lockOverrideKey(ctx, db, override.MatchID.String(), override.DataType, override.Field); err != nil {
		return domain.MatchOverride{}, err
	}
	var nextVersion int
	if err := db.QueryRow(ctx, `
		SELECT COALESCE(MAX(version), 0) + 1
		FROM match_overrides
		WHERE match_id = $1 AND data_type = $2 AND field = $3
	`, override.MatchID.String(), override.DataType, override.Field).Scan(&nextVersion); err != nil {
		return domain.MatchOverride{}, fmt.Errorf("next override version: %w", err)
	}
	override.Version = nextVersion
	_, err := db.Exec(ctx, `
		INSERT INTO match_overrides (
			id, match_id, data_type, field, value, justification, actor, version, applied_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, override.ID.String(), override.MatchID.String(), override.DataType, override.Field,
		override.Value, override.Justification, override.Actor, override.Version, override.AppliedAt.UTC())
	if err != nil {
		return domain.MatchOverride{}, fmt.Errorf("save override: %w", err)
	}
	return override, nil
}

func updateSurfaceStateTx(ctx context.Context, db writeDB, matchID domain.ID, surface string, state domain.AvailabilityState, now time.Time) error {
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
	tag, err := db.Exec(ctx, query, matchID.String(), string(state), now.UTC())
	if err != nil {
		return fmt.Errorf("update surface state: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func appendAuditTx(ctx context.Context, db writeDB, event domain.AuditEvent) (domain.AuditEvent, error) {
	var id int64
	var createdAt time.Time
	err := db.QueryRow(ctx, `
		INSERT INTO audit_events (actor, action, entity_type, entity_id, reason, before_json, after_json)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`, event.Actor, event.Action, event.EntityType, event.EntityID, event.Reason, event.BeforeJSON, event.AfterJSON).Scan(&id, &createdAt)
	if err != nil {
		return domain.AuditEvent{}, fmt.Errorf("append audit event: %w", err)
	}
	event.ID = id
	event.CreatedAt = utc(createdAt)
	return event, nil
}

func lockOverrideKey(ctx context.Context, db writeDB, matchID, dataType, field string) error {
	h := fnv.New64a()
	_, _ = h.Write([]byte(matchID + "\x00" + dataType + "\x00" + field))
	key := int64(h.Sum64())
	_, err := db.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, key)
	if err != nil {
		return fmt.Errorf("lock override key: %w", err)
	}
	return nil
}

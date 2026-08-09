package store

import (
	"context"
	"fmt"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuditStore persists maintainer audit events.
type AuditStore struct {
	pool *pgxpool.Pool
}

// NewAuditStore creates an audit store backed by the provided pool.
func NewAuditStore(pool *pgxpool.Pool) *AuditStore {
	return &AuditStore{pool: pool}
}

// Append inserts an audit event and returns it with generated id/timestamp.
func (s *AuditStore) Append(ctx context.Context, event domain.AuditEvent) (*domain.AuditEvent, error) {
	saved, err := appendAuditTx(ctx, s.pool, event)
	if err != nil {
		return nil, err
	}
	return &saved, nil
}

// ListRecent returns recent audit events, optionally filtered by entity.
func (s *AuditStore) ListRecent(ctx context.Context, entityType, entityID string, limit int) ([]domain.AuditEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, actor, action, entity_type, entity_id, reason, before_json, after_json, created_at
		FROM audit_events
		WHERE ($1 = '' OR entity_type = $1)
		  AND ($2 = '' OR entity_id = $2)
		ORDER BY created_at DESC, id DESC
		LIMIT $3
	`, entityType, entityID, limit)
	if err != nil {
		return nil, fmt.Errorf("list audit events: %w", err)
	}
	defer rows.Close()

	var out []domain.AuditEvent
	for rows.Next() {
		var ev domain.AuditEvent
		if err := rows.Scan(
			&ev.ID, &ev.Actor, &ev.Action, &ev.EntityType, &ev.EntityID, &ev.Reason,
			&ev.BeforeJSON, &ev.AfterJSON, &ev.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan audit event: %w", err)
		}
		ev.CreatedAt = utc(ev.CreatedAt)
		out = append(out, ev)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("audit rows: %w", err)
	}
	return out, nil
}

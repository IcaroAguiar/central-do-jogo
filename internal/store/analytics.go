package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AnalyticsStore persists first-party analytics events (REQ-020).
type AnalyticsStore struct {
	pool *pgxpool.Pool
}

// NewAnalyticsStore creates an analytics store backed by the provided pool.
func NewAnalyticsStore(pool *pgxpool.Pool) *AnalyticsStore {
	return &AnalyticsStore{pool: pool}
}

// Insert writes a new analytics event.
func (s *AnalyticsStore) Insert(ctx context.Context, event domain.AnalyticsEvent) error {
	props, err := json.Marshal(event.Properties)
	if err != nil {
		return fmt.Errorf("marshal analytics properties: %w", err)
	}
	if props == nil {
		props = []byte("{}")
	}
	var userID *string
	if event.UserID != nil {
		v := event.UserID.String()
		userID = &v
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO analytics_events (id, anonymous_id, user_id, event_type, properties, created_at)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6)
	`, event.ID.String(), event.AnonymousID, userID, event.EventType, string(props), event.CreatedAt.UTC())
	if err != nil {
		return fmt.Errorf("insert analytics event: %w", err)
	}
	return nil
}

// ListByUserID returns events currently linked to the account, newest first.
func (s *AnalyticsStore) ListByUserID(ctx context.Context, userID domain.ID, limit int) ([]domain.AnalyticsEvent, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, anonymous_id, user_id, event_type, properties, created_at
		FROM analytics_events
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID.String(), limit)
	if err != nil {
		return nil, fmt.Errorf("list analytics by user: %w", err)
	}
	defer rows.Close()

	var out []domain.AnalyticsEvent
	for rows.Next() {
		var ev domain.AnalyticsEvent
		var id, anonymousID, eventType string
		var userCol *string
		var props []byte
		if err := rows.Scan(&id, &anonymousID, &userCol, &eventType, &props, &ev.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan analytics event: %w", err)
		}
		ev.ID = domain.ID(id)
		ev.AnonymousID = anonymousID
		ev.EventType = eventType
		ev.CreatedAt = utc(ev.CreatedAt)
		if userCol != nil {
			uid := domain.ID(*userCol)
			ev.UserID = &uid
		}
		if len(props) > 0 {
			if err := json.Unmarshal(props, &ev.Properties); err != nil {
				return nil, fmt.Errorf("unmarshal analytics properties: %w", err)
			}
		}
		if ev.Properties == nil {
			ev.Properties = map[string]any{}
		}
		out = append(out, ev)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("analytics rows: %w", err)
	}
	return out, nil
}

// DeleteBefore hard-deletes events older than cutoff (REQ-020 retention).
func (s *AnalyticsStore) DeleteBefore(ctx context.Context, cutoff time.Time) (int64, error) {
	tag, err := s.pool.Exec(ctx, `DELETE FROM analytics_events WHERE created_at < $1`, cutoff.UTC())
	if err != nil {
		return 0, fmt.Errorf("delete analytics before: %w", err)
	}
	return tag.RowsAffected(), nil
}

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

// PushStore persists Web Push subscriptions and outbox rows.
type PushStore struct {
	pool *pgxpool.Pool
}

// NewPushStore creates a push store backed by the provided pool.
func NewPushStore(pool *pgxpool.Pool) *PushStore {
	return &PushStore{pool: pool}
}

const pushSubscriptionColumns = `id, user_id, endpoint, p256dh, auth, user_agent, created_at, last_seen_at, disabled_at`

func scanPushSubscription(row pgx.Row) (*domain.PushSubscription, error) {
	var s domain.PushSubscription
	var id, userID string
	var disabledAt *time.Time
	if err := row.Scan(
		&id, &userID, &s.Endpoint, &s.P256dh, &s.Auth, &s.UserAgent,
		&s.CreatedAt, &s.LastSeenAt, &disabledAt,
	); err != nil {
		return nil, err
	}
	s.ID = domain.ID(id)
	s.UserID = domain.ID(userID)
	s.CreatedAt = utc(s.CreatedAt)
	s.LastSeenAt = utc(s.LastSeenAt)
	if disabledAt != nil {
		t := utc(*disabledAt)
		s.DisabledAt = &t
	}
	return &s, nil
}

// GetByEndpoint returns a subscription by endpoint URL, or nil.
func (s *PushStore) GetByEndpoint(ctx context.Context, endpoint string) (*domain.PushSubscription, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+pushSubscriptionColumns+` FROM push_subscriptions WHERE endpoint = $1`, endpoint)
	sub, err := scanPushSubscription(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get push subscription by endpoint: %w", err)
	}
	return sub, nil
}

// UpsertSubscription inserts or reactivates a subscription keyed by endpoint.
// Callers must ensure the endpoint is not owned by a different user.
func (s *PushStore) UpsertSubscription(ctx context.Context, sub domain.PushSubscription, now time.Time) (*domain.PushSubscription, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO push_subscriptions (
			id, user_id, endpoint, p256dh, auth, user_agent, created_at, last_seen_at, disabled_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $7, NULL)
		ON CONFLICT (endpoint) DO UPDATE SET
			p256dh = EXCLUDED.p256dh,
			auth = EXCLUDED.auth,
			user_agent = EXCLUDED.user_agent,
			last_seen_at = EXCLUDED.last_seen_at,
			disabled_at = NULL
		WHERE push_subscriptions.user_id = EXCLUDED.user_id
		RETURNING `+pushSubscriptionColumns,
		sub.ID.String(), sub.UserID.String(), sub.Endpoint, sub.P256dh, sub.Auth,
		sub.UserAgent, now.UTC(),
	)
	out, err := scanPushSubscription(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("upsert push subscription: endpoint owned by another user")
		}
		return nil, fmt.Errorf("upsert push subscription: %w", err)
	}
	return out, nil
}

// ListActiveByUser returns non-disabled subscriptions for a user.
func (s *PushStore) ListActiveByUser(ctx context.Context, userID domain.ID) ([]domain.PushSubscription, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT `+pushSubscriptionColumns+`
		FROM push_subscriptions
		WHERE user_id = $1 AND disabled_at IS NULL
		ORDER BY created_at ASC`, userID.String())
	if err != nil {
		return nil, fmt.Errorf("list push subscriptions: %w", err)
	}
	defer rows.Close()
	var out []domain.PushSubscription
	for rows.Next() {
		sub, err := scanPushSubscription(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *sub)
	}
	return out, rows.Err()
}

// DeleteByEndpoint removes a subscription for the owning user.
func (s *PushStore) DeleteByEndpoint(ctx context.Context, userID domain.ID, endpoint string) (bool, error) {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM push_subscriptions WHERE user_id = $1 AND endpoint = $2`,
		userID.String(), endpoint)
	if err != nil {
		return false, fmt.Errorf("delete push subscription: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

// DisableByEndpoint marks an endpoint expired/gone so cleanup can purge it.
func (s *PushStore) DisableByEndpoint(ctx context.Context, endpoint string, now time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE push_subscriptions SET disabled_at = $2
		WHERE endpoint = $1 AND disabled_at IS NULL`, endpoint, now.UTC())
	if err != nil {
		return fmt.Errorf("disable push subscription: %w", err)
	}
	return nil
}

// DeleteDisabledBefore purges endpoints disabled before the cutoff.
func (s *PushStore) DeleteDisabledBefore(ctx context.Context, cutoff time.Time) (int64, error) {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM push_subscriptions WHERE disabled_at IS NOT NULL AND disabled_at < $1`,
		cutoff.UTC())
	if err != nil {
		return 0, fmt.Errorf("cleanup push subscriptions: %w", err)
	}
	return tag.RowsAffected(), nil
}

const pushOutboxColumns = `id, idempotency_key, alert_type, match_id, version, payload, status, attempts, max_attempts, last_error, created_at, updated_at, accepted_at`

func scanPushOutbox(row pgx.Row) (*domain.PushOutboxEntry, error) {
	var e domain.PushOutboxEntry
	var id, status string
	var matchID *string
	var acceptedAt *time.Time
	if err := row.Scan(
		&id, &e.IdempotencyKey, &e.AlertType, &matchID, &e.Version, &e.Payload,
		&status, &e.Attempts, &e.MaxAttempts, &e.LastError,
		&e.CreatedAt, &e.UpdatedAt, &acceptedAt,
	); err != nil {
		return nil, err
	}
	e.ID = domain.ID(id)
	e.Status = domain.PushOutboxStatus(status)
	e.CreatedAt = utc(e.CreatedAt)
	e.UpdatedAt = utc(e.UpdatedAt)
	if matchID != nil {
		mid := domain.ID(*matchID)
		e.MatchID = &mid
	}
	if acceptedAt != nil {
		t := utc(*acceptedAt)
		e.AcceptedAt = &t
	}
	return &e, nil
}

// EnqueueOutbox inserts an idempotent outbox row (REQ-012).
func (s *PushStore) EnqueueOutbox(ctx context.Context, entry domain.PushOutboxEntry, now time.Time) (*domain.PushOutboxEntry, error) {
	var matchID *string
	if entry.MatchID != nil {
		s := entry.MatchID.String()
		matchID = &s
	}
	if entry.MaxAttempts <= 0 {
		entry.MaxAttempts = 5
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO push_outbox (
			id, idempotency_key, alert_type, match_id, version, payload,
			status, attempts, max_attempts, last_error, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, 'pending', 0, $7, '', $8, $8)
		ON CONFLICT (idempotency_key) DO UPDATE SET id = push_outbox.id
		RETURNING `+pushOutboxColumns,
		entry.ID.String(), entry.IdempotencyKey, entry.AlertType, matchID,
		entry.Version, entry.Payload, entry.MaxAttempts, now.UTC(),
	)
	out, err := scanPushOutbox(row)
	if err != nil {
		return nil, fmt.Errorf("enqueue push outbox: %w", err)
	}
	return out, nil
}

// GetOutboxByIdempotencyKey returns an outbox row or nil.
func (s *PushStore) GetOutboxByIdempotencyKey(ctx context.Context, key string) (*domain.PushOutboxEntry, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT `+pushOutboxColumns+` FROM push_outbox WHERE idempotency_key = $1`, key)
	out, err := scanPushOutbox(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get push outbox: %w", err)
	}
	return out, nil
}

// MarkOutboxAccepted records push-service acceptance (REQ-025).
func (s *PushStore) MarkOutboxAccepted(ctx context.Context, id domain.ID, now time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE push_outbox SET status = 'accepted', accepted_at = $2, updated_at = $2, last_error = ''
		WHERE id = $1`, id.String(), now.UTC())
	if err != nil {
		return fmt.Errorf("mark push outbox accepted: %w", err)
	}
	return nil
}

// UpdateOutboxPayload persists progress fields (e.g. deliveredEndpoints) without changing status.
func (s *PushStore) UpdateOutboxPayload(ctx context.Context, id domain.ID, payload []byte, now time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE push_outbox SET payload = $2, updated_at = $3 WHERE id = $1`,
		id.String(), payload, now.UTC())
	if err != nil {
		return fmt.Errorf("update push outbox payload: %w", err)
	}
	return nil
}

// MarkOutboxFailure increments attempts and sets failed/dead.
func (s *PushStore) MarkOutboxFailure(ctx context.Context, id domain.ID, lastError string, now time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE push_outbox SET
			attempts = attempts + 1,
			last_error = $2,
			updated_at = $3,
			status = CASE WHEN attempts + 1 >= max_attempts THEN 'dead' ELSE 'failed' END
		WHERE id = $1`, id.String(), lastError, now.UTC())
	if err != nil {
		return fmt.Errorf("mark push outbox failure: %w", err)
	}
	return nil
}

package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SourceHealth tracks operational health for a source adapter.
type SourceHealth struct {
	SourceID            string
	LastSuccessAt       *time.Time
	LastErrorAt         *time.Time
	LastError           string
	NextRunAt           time.Time
	ConsecutiveFailures int
	UpdatedAt           time.Time
}

// HealthStore provides source_health operations.
type HealthStore struct {
	pool *pgxpool.Pool
}

// NewHealthStore creates a health store backed by the provided pgx pool.
func NewHealthStore(pool *pgxpool.Pool) *HealthStore {
	return &HealthStore{pool: pool}
}

// RecordSuccess marks a source as healthy and resets failure count.
func (h *HealthStore) RecordSuccess(ctx context.Context, sourceID string, nextRunAt time.Time) error {
	_, err := h.pool.Exec(ctx, `
		INSERT INTO source_health (source_id, last_success_at, consecutive_failures, next_run_at, updated_at)
		VALUES ($1, now(), 0, $2, now())
		ON CONFLICT (source_id) DO UPDATE SET
			last_success_at = now(),
			consecutive_failures = 0,
			next_run_at = EXCLUDED.next_run_at,
			updated_at = now()
	`, sourceID, nextRunAt)
	if err != nil {
		return fmt.Errorf("record health success for %s: %w", sourceID, err)
	}
	return nil
}

// RecordFailure marks a source failure and increments consecutive failure count.
func (h *HealthStore) RecordFailure(ctx context.Context, sourceID string, errMsg string, nextRunAt time.Time) error {
	_, err := h.pool.Exec(ctx, `
		INSERT INTO source_health (source_id, last_error_at, last_error, consecutive_failures, next_run_at, updated_at)
		VALUES ($1, now(), $2, 1, $3, now())
		ON CONFLICT (source_id) DO UPDATE SET
			last_error_at = now(),
			last_error = EXCLUDED.last_error,
			consecutive_failures = source_health.consecutive_failures + 1,
			next_run_at = EXCLUDED.next_run_at,
			updated_at = now()
	`, sourceID, errMsg, nextRunAt)
	if err != nil {
		return fmt.Errorf("record health failure for %s: %w", sourceID, err)
	}
	return nil
}

// Get returns the current health state for a source, or nil if not tracked.
func (h *HealthStore) Get(ctx context.Context, sourceID string) (*SourceHealth, error) {
	var sh SourceHealth
	err := h.pool.QueryRow(ctx, `
		SELECT source_id, last_success_at, last_error_at, last_error,
		       next_run_at, consecutive_failures, updated_at
		FROM source_health
		WHERE source_id = $1
	`, sourceID).Scan(
		&sh.SourceID, &sh.LastSuccessAt, &sh.LastErrorAt, &sh.LastError,
		&sh.NextRunAt, &sh.ConsecutiveFailures, &sh.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("get health for %s: %w", sourceID, err)
	}
	return &sh, nil
}

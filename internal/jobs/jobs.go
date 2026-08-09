// Package jobs implements Postgres-backed job leases, retries, idempotency,
// and source health tracking for the Central do Jogo worker.
package jobs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Job statuses.
const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusDead      = "dead"
)

// Job represents a persisted job record.
type Job struct {
	ID             string
	JobType        string
	Payload        json.RawMessage
	Status         string
	LeaseOwner     string
	LeaseExpiresAt *time.Time
	RunAfter       time.Time
	Attempts       int
	MaxAttempts    int
	IdempotencyKey string
	LastError      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Enqueuer schedules jobs with idempotency keys.
type Enqueuer interface {
	Enqueue(ctx context.Context, jobType string, payload json.RawMessage, idempotencyKey string, runAfter time.Time, maxAttempts int) (*Job, error)
}

// Store provides Postgres-backed job operations.
type Store struct {
	pool *pgxpool.Pool
}

// Ensure Store satisfies Enqueuer.
var _ Enqueuer = (*Store)(nil)

// NewStore creates a job store backed by the provided pgx pool.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Enqueue inserts a new job or returns the existing job if the idempotency key already exists.
func (s *Store) Enqueue(ctx context.Context, jobType string, payload json.RawMessage, idempotencyKey string, runAfter time.Time, maxAttempts int) (*Job, error) {
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	id, err := generateJobID()
	if err != nil {
		return nil, fmt.Errorf("generate job id: %w", err)
	}

	var job Job
	err = s.pool.QueryRow(ctx, `
		INSERT INTO jobs (id, job_type, payload, idempotency_key, run_after, max_attempts)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (idempotency_key) DO UPDATE SET id = jobs.id
		RETURNING id, job_type, payload, status, lease_owner, lease_expires_at,
		          run_after, attempts, max_attempts, idempotency_key, last_error,
		          created_at, updated_at
	`, id, jobType, payload, idempotencyKey, runAfter, maxAttempts).Scan(
		&job.ID, &job.JobType, &job.Payload, &job.Status, &job.LeaseOwner,
		&job.LeaseExpiresAt, &job.RunAfter, &job.Attempts, &job.MaxAttempts,
		&job.IdempotencyKey, &job.LastError, &job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("enqueue job: %w", err)
	}
	return &job, nil
}

// LeaseDuration is the default time a worker holds a job lease.
const LeaseDuration = 5 * time.Minute

// Claim attempts to acquire a lease on the next available job.
// Uses a CTE with FOR UPDATE SKIP LOCKED for concurrency-safe claiming.
func (s *Store) Claim(ctx context.Context, owner string) (*Job, error) {
	leaseExpires := time.Now().Add(LeaseDuration)

	var job Job
	err := s.pool.QueryRow(ctx, `
		WITH next_job AS (
			SELECT id FROM jobs
			WHERE (status = 'pending' OR (status = 'failed' AND attempts < max_attempts))
			  AND run_after <= now()
			  AND (lease_expires_at IS NULL OR lease_expires_at < now())
			ORDER BY run_after ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		UPDATE jobs SET
			status = $1,
			lease_owner = $2,
			lease_expires_at = $3,
			attempts = attempts + 1,
			updated_at = now()
		FROM next_job
		WHERE jobs.id = next_job.id
		RETURNING jobs.id, jobs.job_type, jobs.payload, jobs.status, jobs.lease_owner,
		          jobs.lease_expires_at, jobs.run_after, jobs.attempts, jobs.max_attempts,
		          jobs.idempotency_key, jobs.last_error, jobs.created_at, jobs.updated_at
	`, StatusRunning, owner, leaseExpires).Scan(
		&job.ID, &job.JobType, &job.Payload, &job.Status, &job.LeaseOwner,
		&job.LeaseExpiresAt, &job.RunAfter, &job.Attempts, &job.MaxAttempts,
		&job.IdempotencyKey, &job.LastError, &job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("claim job: %w", err)
	}
	return &job, nil
}

// Complete marks a job as completed.
func (s *Store) Complete(ctx context.Context, jobID string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE jobs SET
			status = $1,
			lease_owner = '',
			lease_expires_at = NULL,
			last_error = '',
			updated_at = now()
		WHERE id = $2
	`, StatusCompleted, jobID)
	if err != nil {
		return fmt.Errorf("complete job %s: %w", jobID, err)
	}
	return nil
}

// Fail marks a job as failed with an error message. If max attempts are exhausted,
// the job transitions to dead.
func (s *Store) Fail(ctx context.Context, jobID string, jobErr string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE jobs SET
			status = CASE
				WHEN attempts >= max_attempts THEN 'dead'
				ELSE 'failed'
			END,
			lease_owner = '',
			lease_expires_at = NULL,
			last_error = $1,
			run_after = now() + (interval '1 second' * power(2, LEAST(attempts, 10))),
			updated_at = now()
		WHERE id = $2
	`, jobErr, jobID)
	if err != nil {
		return fmt.Errorf("fail job %s: %w", jobID, err)
	}
	return nil
}

// RecordAttempt inserts a job_attempts row.
func (s *Store) RecordAttempt(ctx context.Context, jobID string, attemptNo int, outcome string, errMsg string) error {
	finishedAt := time.Now()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO job_attempts (job_id, attempt_no, finished_at, error, outcome)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (job_id, attempt_no) DO UPDATE SET
			finished_at = EXCLUDED.finished_at,
			error = EXCLUDED.error,
			outcome = EXCLUDED.outcome
	`, jobID, attemptNo, finishedAt, errMsg, outcome)
	if err != nil {
		return fmt.Errorf("record attempt: %w", err)
	}
	return nil
}

func generateJobID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "job_" + hex.EncodeToString(b), nil
}

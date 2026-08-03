package push

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/jobs"
)

const (
	JobTypeDeliver = "push.deliver"
	JobTypeCleanup = "push.cleanup_expired"
)

// EnqueueDeliverJob schedules outbox delivery with the same idempotency key (REQ-012).
func EnqueueDeliverJob(ctx context.Context, store *jobs.Store, idempotencyKey string, now time.Time) (*jobs.Job, error) {
	payload, err := json.Marshal(map[string]string{"idempotencyKey": idempotencyKey})
	if err != nil {
		return nil, err
	}
	return store.Enqueue(ctx, JobTypeDeliver, payload, idempotencyKey, now, 5)
}

// EnqueueCleanupJob schedules expired-endpoint purge (unique per calendar day).
func EnqueueCleanupJob(ctx context.Context, store *jobs.Store, now time.Time) (*jobs.Job, error) {
	day := now.UTC().Format("2006-01-02")
	key := "push:cleanup:" + day
	payload, err := json.Marshal(map[string]string{"day": day})
	if err != nil {
		return nil, err
	}
	return store.Enqueue(ctx, JobTypeCleanup, payload, key, now, 3)
}

// DeliverHandler processes push.deliver jobs.
func DeliverHandler(runner *OutboxRunner) jobs.Handler {
	return func(ctx context.Context, job *jobs.Job) error {
		var body struct {
			IdempotencyKey string `json:"idempotencyKey"`
		}
		if err := json.Unmarshal(job.Payload, &body); err != nil {
			return fmt.Errorf("decode push.deliver payload: %w", err)
		}
		if body.IdempotencyKey == "" {
			return fmt.Errorf("push.deliver missing idempotencyKey")
		}
		return runner.DeliverOutbox(ctx, body.IdempotencyKey)
	}
}

// CleanupHandler processes push.cleanup_expired jobs.
func CleanupHandler(runner *OutboxRunner) jobs.Handler {
	return func(ctx context.Context, _ *jobs.Job) error {
		_, err := runner.CleanupExpiredEndpoints(ctx, 30*24*time.Hour)
		return err
	}
}

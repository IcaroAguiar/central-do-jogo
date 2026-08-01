package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Handler processes a claimed job. Implementations should be idempotent.
type Handler func(ctx context.Context, job *Job) error

// HandlerRegistry maps job types to handlers.
type HandlerRegistry map[string]Handler

// Worker is the main job processing loop.
type Worker struct {
	store        *Store
	health       *HealthStore
	handlers     HandlerRegistry
	owner        string
	pollInterval time.Duration
}

// NewWorker creates a worker that claims and executes jobs.
func NewWorker(store *Store, health *HealthStore, handlers HandlerRegistry, owner string) *Worker {
	return &Worker{
		store:        store,
		health:       health,
		handlers:     handlers,
		owner:        owner,
		pollInterval: 5 * time.Second,
	}
}

// Run starts the claim-execute loop until ctx is cancelled. Returns nil on clean shutdown.
func (w *Worker) Run(ctx context.Context, logger *slog.Logger) error {
	logger.Info("worker loop started", "owner", w.owner)
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("worker loop shutting down")
			return nil
		case <-ticker.C:
			if err := w.tick(ctx, logger); err != nil {
				logger.Error("worker tick error", "error", err)
			}
		}
	}
}

func (w *Worker) tick(ctx context.Context, logger *slog.Logger) error {
	job, err := w.store.Claim(ctx, w.owner)
	if err != nil {
		return fmt.Errorf("claim: %w", err)
	}
	if job == nil {
		return nil
	}

	jobLogger := logger.With(
		slog.String("job_id", job.ID),
		slog.String("job_type", job.JobType),
		slog.Int("attempt", job.Attempts),
	)
	jobLogger.Info("job claimed")

	handler, ok := w.handlers[job.JobType]
	if !ok {
		errMsg := fmt.Sprintf("no handler registered for job type %q", job.JobType)
		jobLogger.Error(errMsg)
		_ = w.store.RecordAttempt(ctx, job.ID, job.Attempts, "error", errMsg)
		return w.store.Fail(ctx, job.ID, errMsg)
	}

	execErr := handler(ctx, job)
	if execErr != nil {
		errMsg := execErr.Error()
		jobLogger.Error("job failed", "error", errMsg)
		_ = w.store.RecordAttempt(ctx, job.ID, job.Attempts, "error", errMsg)
		return w.store.Fail(ctx, job.ID, errMsg)
	}

	jobLogger.Info("job completed")
	_ = w.store.RecordAttempt(ctx, job.ID, job.Attempts, "success", "")
	return w.store.Complete(ctx, job.ID)
}

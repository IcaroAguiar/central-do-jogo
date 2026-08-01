package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// Handler processes a claimed job. Implementations should be idempotent.
type Handler func(ctx context.Context, job *Job) error

// HandlerRegistry maps job types to handlers.
type HandlerRegistry map[string]Handler

// handlerTimeout is the maximum time a handler can run before the context is cancelled.
// Set to LeaseDuration minus a safety buffer so the handler cannot outlive the lease.
const handlerTimeout = 4 * time.Minute

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
		if err := w.store.RecordAttempt(ctx, job.ID, job.Attempts, "error", errMsg); err != nil {
			return fmt.Errorf("record attempt: %w", err)
		}
		return w.store.Fail(ctx, job.ID, errMsg)
	}

	handlerCtx, cancel := context.WithTimeout(ctx, handlerTimeout)
	defer cancel()

	execErr := handler(handlerCtx, job)

	sourceID := extractSourceID(job)

	if execErr != nil {
		errMsg := execErr.Error()
		jobLogger.Error("job failed", "error", errMsg)
		if err := w.store.RecordAttempt(ctx, job.ID, job.Attempts, "error", errMsg); err != nil {
			return fmt.Errorf("record attempt: %w", err)
		}
		if sourceID != "" && w.health != nil {
			nextRun := NextRunAt(time.Now(), dataTypeFromJobType(job.JobType), nil)
			if hErr := w.health.RecordFailure(ctx, sourceID, errMsg, nextRun); hErr != nil {
				return fmt.Errorf("record health failure: %w", hErr)
			}
		}
		return w.store.Fail(ctx, job.ID, errMsg)
	}

	jobLogger.Info("job completed")
	if err := w.store.RecordAttempt(ctx, job.ID, job.Attempts, "success", ""); err != nil {
		return fmt.Errorf("record attempt: %w", err)
	}
	if sourceID != "" && w.health != nil {
		nextRun := NextRunAt(time.Now(), dataTypeFromJobType(job.JobType), nil)
		if hErr := w.health.RecordSuccess(ctx, sourceID, nextRun); hErr != nil {
			return fmt.Errorf("record health success: %w", hErr)
		}
	}
	return w.store.Complete(ctx, job.ID)
}

// extractSourceID attempts to read the source field from the job payload JSON,
// falling back to the suffix after "ingest." in the job type.
func extractSourceID(job *Job) string {
	if job.Payload != nil {
		var p struct {
			Source string `json:"source"`
		}
		if json.Unmarshal(job.Payload, &p) == nil && p.Source != "" {
			return p.Source
		}
	}
	if strings.HasPrefix(job.JobType, "ingest.") {
		return strings.TrimPrefix(job.JobType, "ingest.")
	}
	return ""
}

// dataTypeFromJobType maps a job type to the cadence data type.
// Ingest jobs default to "schedule" cadence unless a more specific mapping exists.
func dataTypeFromJobType(jobType string) string {
	switch {
	case strings.Contains(jobType, "lineup"):
		return "lineup"
	case strings.Contains(jobType, "news"):
		return "news"
	default:
		return "schedule"
	}
}

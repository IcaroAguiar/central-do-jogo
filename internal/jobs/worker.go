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

// ScheduleFunc runs opportunistic maintenance (e.g. daily purge enqueue).
type ScheduleFunc func(ctx context.Context) error

// Worker is the main job processing loop.
type Worker struct {
	store         *Store
	health        *HealthStore
	handlers      HandlerRegistry
	owner         string
	pollInterval  time.Duration
	schedule      ScheduleFunc
	scheduleEvery time.Duration
	lastSchedule  time.Time
}

// WorkerOption configures optional worker behavior.
type WorkerOption func(*Worker)

// WithSchedule registers a hook invoked on worker start and about every interval.
func WithSchedule(fn ScheduleFunc, every time.Duration) WorkerOption {
	return func(w *Worker) {
		w.schedule = fn
		if every <= 0 {
			every = time.Hour
		}
		w.scheduleEvery = every
	}
}

// NewWorker creates a worker that claims and executes jobs.
func NewWorker(store *Store, health *HealthStore, handlers HandlerRegistry, owner string, opts ...WorkerOption) *Worker {
	w := &Worker{
		store:        store,
		health:       health,
		handlers:     handlers,
		owner:        owner,
		pollInterval: 5 * time.Second,
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

// Run starts the claim-execute loop until ctx is cancelled. Returns nil on clean shutdown.
func (w *Worker) Run(ctx context.Context, logger *slog.Logger) error {
	logger.Info("worker loop started", "owner", w.owner)
	w.maybeSchedule(ctx, logger)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("worker loop shutting down")
			return nil
		case <-ticker.C:
			w.maybeSchedule(ctx, logger)
			if err := w.tick(ctx, logger); err != nil {
				logger.Error("worker tick error", "error", err)
			}
		}
	}
}

// maybeSchedule runs the schedule hook at most once per scheduleEvery.
// Exported behavior is covered via MaybeScheduleForTest.
func (w *Worker) maybeSchedule(ctx context.Context, logger *slog.Logger) {
	if w.schedule == nil {
		return
	}
	now := time.Now()
	if !w.lastSchedule.IsZero() && now.Sub(w.lastSchedule) < w.scheduleEvery {
		return
	}
	if err := w.schedule(ctx); err != nil {
		if logger != nil {
			logger.Error("worker schedule hook failed", "error", err)
		}
		return
	}
	w.lastSchedule = now
}

// MaybeScheduleForTest exposes maybeSchedule for unit tests.
func (w *Worker) MaybeScheduleForTest(ctx context.Context) {
	w.maybeSchedule(ctx, nil)
}

// LastScheduleForTest returns the last successful schedule timestamp.
func (w *Worker) LastScheduleForTest() time.Time { return w.lastSchedule }

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

package jobs_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/jobs"
)

func TestWorkerScheduleRunsOnFirstCallAndRespectsInterval(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	w := jobs.NewWorker(nil, nil, nil, "test-owner",
		jobs.WithSchedule(func(context.Context) error {
			calls.Add(1)
			return nil
		}, time.Hour),
	)

	w.MaybeScheduleForTest(context.Background())
	w.MaybeScheduleForTest(context.Background())
	if calls.Load() != 1 {
		t.Fatalf("calls = %d, want 1 within interval", calls.Load())
	}
	if w.LastScheduleForTest().IsZero() {
		t.Fatal("expected last schedule timestamp")
	}
}

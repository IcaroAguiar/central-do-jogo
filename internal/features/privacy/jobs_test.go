package privacy_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/features/privacy"
	"github.com/IcaroAguiar/central-do-jogo/internal/jobs"
)

type memEnqueuer struct {
	calls []enqueueCall
}

type enqueueCall struct {
	jobType        string
	idempotencyKey string
	payload        json.RawMessage
}

func (m *memEnqueuer) Enqueue(_ context.Context, jobType string, payload json.RawMessage, idempotencyKey string, _ time.Time, _ int) (*jobs.Job, error) {
	m.calls = append(m.calls, enqueueCall{jobType: jobType, idempotencyKey: idempotencyKey, payload: payload})
	return &jobs.Job{ID: "job_1", JobType: jobType, IdempotencyKey: idempotencyKey, Payload: payload}, nil
}

type memPurger struct{ called int }

func (m *memPurger) PurgeExpired(context.Context) (int64, error) {
	m.called++
	return 3, nil
}

func TestEnqueuePurgeJobIdempotencyKey(t *testing.T) {
	t.Parallel()
	enq := &memEnqueuer{}
	now := time.Date(2026, 8, 9, 15, 0, 0, 0, time.UTC)
	job, err := privacy.EnqueuePurgeJob(context.Background(), enq, now)
	if err != nil {
		t.Fatal(err)
	}
	if job.JobType != privacy.JobTypePurgeAnalytics {
		t.Fatalf("type = %s", job.JobType)
	}
	if job.IdempotencyKey != "privacy:purge:2026-08-09" {
		t.Fatalf("key = %s", job.IdempotencyKey)
	}
	if len(enq.calls) != 1 {
		t.Fatalf("calls = %d", len(enq.calls))
	}
}

func TestPurgeHandlerUsesPurgeExpired(t *testing.T) {
	t.Parallel()
	p := &memPurger{}
	handler := privacy.PurgeHandler(p)
	if err := handler(context.Background(), &jobs.Job{ID: "job_1"}); err != nil {
		t.Fatal(err)
	}
	if p.called != 1 {
		t.Fatalf("called = %d", p.called)
	}
}

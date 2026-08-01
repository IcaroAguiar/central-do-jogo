package jobs_test

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/jobs"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

func openTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := database.OpenPool(ctx, url)
	if err != nil {
		t.Fatalf("OpenPool: %v", err)
	}
	if err := database.Migrate(ctx, pool); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	_, _ = pool.Exec(ctx, "DELETE FROM job_attempts")
	_, _ = pool.Exec(ctx, "DELETE FROM jobs")
	_, _ = pool.Exec(ctx, "DELETE FROM source_health")

	t.Cleanup(func() { pool.Close() })
	return pool
}

func TestEnqueueAndClaim(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	store := jobs.NewStore(pool)

	payload := json.RawMessage(`{"source":"openfootball_brazil"}`)
	job, err := store.Enqueue(ctx, "ingest.openfootball_brazil", payload, "key-1", time.Now(), 3)
	if err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	if job.Status != jobs.StatusPending {
		t.Errorf("status = %q, want pending", job.Status)
	}

	// Idempotency: same key returns existing
	job2, err := store.Enqueue(ctx, "ingest.openfootball_brazil", payload, "key-1", time.Now(), 3)
	if err != nil {
		t.Fatalf("Enqueue idempotent: %v", err)
	}
	if job2.ID != job.ID {
		t.Errorf("idempotency failed: got different ID %s vs %s", job2.ID, job.ID)
	}

	// Claim
	claimed, err := store.Claim(ctx, "worker-1")
	if err != nil {
		t.Fatalf("Claim: %v", err)
	}
	if claimed == nil {
		t.Fatal("expected a job, got nil")
	}
	if claimed.ID != job.ID {
		t.Errorf("claimed wrong job: %s", claimed.ID)
	}
	if claimed.Status != jobs.StatusRunning {
		t.Errorf("claimed status = %q, want running", claimed.Status)
	}

	// No more jobs available
	empty, err := store.Claim(ctx, "worker-2")
	if err != nil {
		t.Fatalf("Claim empty: %v", err)
	}
	if empty != nil {
		t.Error("expected nil when no jobs available")
	}
}

func TestCompleteAndFail(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	store := jobs.NewStore(pool)

	payload := json.RawMessage(`{}`)
	job, _ := store.Enqueue(ctx, "test.complete", payload, "complete-1", time.Now(), 3)
	_, _ = store.Claim(ctx, "w")
	if err := store.Complete(ctx, job.ID); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	// Test Fail with retry
	job2, _ := store.Enqueue(ctx, "test.fail", payload, "fail-1", time.Now(), 3)
	_, _ = store.Claim(ctx, "w")
	if err := store.Fail(ctx, job2.ID, "temporary error"); err != nil {
		t.Fatalf("Fail: %v", err)
	}

	// After fail, job should be reclaimable (after run_after passes)
	_, _ = pool.Exec(ctx, "UPDATE jobs SET run_after = now() WHERE id = $1", job2.ID)
	reclaimed, err := store.Claim(ctx, "w")
	if err != nil {
		t.Fatalf("Reclaim: %v", err)
	}
	if reclaimed == nil || reclaimed.ID != job2.ID {
		t.Error("expected failed job to be reclaimable")
	}
}

func TestFailExhaustsToDeadStatus(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	store := jobs.NewStore(pool)

	payload := json.RawMessage(`{}`)
	job, _ := store.Enqueue(ctx, "test.dead", payload, "dead-1", time.Now(), 1)
	_, _ = store.Claim(ctx, "w")
	_ = store.Fail(ctx, job.ID, "fatal")

	// Should not be claimable (dead)
	_, _ = pool.Exec(ctx, "UPDATE jobs SET run_after = now() WHERE id = $1", job.ID)
	got, err := store.Claim(ctx, "w")
	if err != nil {
		t.Fatalf("Claim dead: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for dead job, got %+v", got)
	}
}

func TestConcurrentClaim(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	store := jobs.NewStore(pool)

	payload := json.RawMessage(`{}`)
	_, _ = store.Enqueue(ctx, "test.concurrent", payload, "conc-1", time.Now(), 3)

	var claimed int64
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			job, err := store.Claim(ctx, "worker-"+string(rune('A'+id)))
			if err == nil && job != nil {
				atomic.AddInt64(&claimed, 1)
			}
		}(i)
	}
	wg.Wait()

	if claimed != 1 {
		t.Errorf("expected exactly 1 claim, got %d", claimed)
	}
}

func TestSourceHealth(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	_, _ = pool.Exec(ctx, `
		INSERT INTO sources (id, display_name) VALUES ('test_source', 'Test Source')
		ON CONFLICT (id) DO NOTHING
	`)

	hs := jobs.NewHealthStore(pool)

	if err := hs.RecordSuccess(ctx, "test_source", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("RecordSuccess: %v", err)
	}

	health, err := hs.Get(ctx, "test_source")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if health == nil {
		t.Fatal("expected health record")
	}
	if health.ConsecutiveFailures != 0 {
		t.Errorf("failures = %d, want 0", health.ConsecutiveFailures)
	}

	if err := hs.RecordFailure(ctx, "test_source", "timeout", time.Now().Add(2*time.Hour)); err != nil {
		t.Fatalf("RecordFailure: %v", err)
	}
	health, _ = hs.Get(ctx, "test_source")
	if health.ConsecutiveFailures != 1 {
		t.Errorf("failures = %d, want 1", health.ConsecutiveFailures)
	}

	if err := hs.RecordSuccess(ctx, "test_source", time.Now().Add(3*time.Hour)); err != nil {
		t.Fatalf("RecordSuccess reset: %v", err)
	}
	health, _ = hs.Get(ctx, "test_source")
	if health.ConsecutiveFailures != 0 {
		t.Errorf("failures after success = %d, want 0", health.ConsecutiveFailures)
	}
}

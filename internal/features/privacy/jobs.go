package privacy

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/jobs"
)

// JobTypePurgeAnalytics deletes analytics older than the retention window.
const JobTypePurgeAnalytics = "privacy.purge_analytics"

// EnqueuePurgeJob schedules analytics retention purge (unique per calendar day).
func EnqueuePurgeJob(ctx context.Context, store *jobs.Store, now time.Time) (*jobs.Job, error) {
	day := now.UTC().Format("2006-01-02")
	key := "privacy:purge:" + day
	payload, err := json.Marshal(map[string]string{"day": day})
	if err != nil {
		return nil, err
	}
	return store.Enqueue(ctx, JobTypePurgeAnalytics, payload, key, now, 3)
}

// PurgeHandler processes privacy.purge_analytics jobs using the retention window.
func PurgeHandler(analytics AnalyticsRepository, retentionDays int, now func() time.Time) jobs.Handler {
	if now == nil {
		now = time.Now
	}
	if retentionDays <= 0 {
		retentionDays = 90
	}
	retention := time.Duration(retentionDays) * 24 * time.Hour
	return func(ctx context.Context, _ *jobs.Job) error {
		cutoff := now().UTC().Add(-retention)
		n, err := analytics.DeleteBefore(ctx, cutoff)
		if err != nil {
			return fmt.Errorf("purge analytics: %w", err)
		}
		_ = n
		return nil
	}
}

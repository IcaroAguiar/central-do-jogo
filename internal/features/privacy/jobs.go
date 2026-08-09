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
func EnqueuePurgeJob(ctx context.Context, enq jobs.Enqueuer, now time.Time) (*jobs.Job, error) {
	day := now.UTC().Format("2006-01-02")
	key := "privacy:purge:" + day
	payload, err := json.Marshal(map[string]string{"day": day})
	if err != nil {
		return nil, err
	}
	return enq.Enqueue(ctx, JobTypePurgeAnalytics, payload, key, now, 3)
}

// Purger deletes expired analytics rows.
type Purger interface {
	PurgeExpired(ctx context.Context) (int64, error)
}

// PurgeHandler processes privacy.purge_analytics jobs via the privacy retention owner.
func PurgeHandler(purger Purger) jobs.Handler {
	return func(ctx context.Context, _ *jobs.Job) error {
		if _, err := purger.PurgeExpired(ctx); err != nil {
			return fmt.Errorf("purge analytics: %w", err)
		}
		return nil
	}
}

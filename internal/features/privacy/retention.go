package privacy

import (
	"context"
	"time"
)

// Retention owns analytics purge cutoff math for HTTP and worker paths.
type Retention struct {
	analytics AnalyticsRepository
	window    time.Duration
	now       func() time.Time
}

// NewRetention builds a retention owner with a default 90-day window.
func NewRetention(analytics AnalyticsRepository, retentionDays int, now func() time.Time) *Retention {
	if now == nil {
		now = time.Now
	}
	if retentionDays <= 0 {
		retentionDays = 90
	}
	return &Retention{
		analytics: analytics,
		window:    time.Duration(retentionDays) * 24 * time.Hour,
		now:       now,
	}
}

// PurgeExpired deletes analytics rows older than the configured retention window.
func (r *Retention) PurgeExpired(ctx context.Context) (int64, error) {
	cutoff := r.now().UTC().Add(-r.window)
	return r.analytics.DeleteBefore(ctx, cutoff)
}

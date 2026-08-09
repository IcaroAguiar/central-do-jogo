package domain

import "time"

// AnalyticsEvent is a first-party product analytics row (REQ-020).
// anonymous_id is a local browser identifier; user_id is set only when the
// visitor is logged in and consents to link events to their account.
type AnalyticsEvent struct {
	ID          ID
	AnonymousID string
	UserID      *ID
	EventType   string
	Properties  map[string]any
	CreatedAt   time.Time
}

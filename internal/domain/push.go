package domain

import "time"

// Push alert types covered by REQ-011 (opt-in Web Push).
const (
	PushAlertBroadcastConfirmed = "broadcast_confirmed"
	PushAlertBroadcastChanged   = "broadcast_changed"
	PushAlertLineupOfficial     = "lineup_official"
	PushAlertSmokeTest          = "smoke_test"
)

// PushSubscription is a browser Push subscription bound to a user account.
// Endpoint/auth material must never appear in account exports (REQ-019).
type PushSubscription struct {
	ID         ID
	UserID     ID
	Endpoint   string
	P256dh     string
	Auth       string
	UserAgent  string
	CreatedAt  time.Time
	LastSeenAt time.Time
	DisabledAt *time.Time
}

// PushOutboxStatus tracks acceptance by the push service (REQ-025), not
// device delivery (RISK-005).
type PushOutboxStatus string

const (
	PushOutboxPending  PushOutboxStatus = "pending"
	PushOutboxAccepted PushOutboxStatus = "accepted"
	PushOutboxFailed   PushOutboxStatus = "failed"
	PushOutboxDead     PushOutboxStatus = "dead"
)

// PushOutboxEntry is an idempotent alert fan-out record (REQ-012).
type PushOutboxEntry struct {
	ID             ID
	IdempotencyKey string
	AlertType      string
	MatchID        *ID
	Version        string
	Payload        []byte
	Status         PushOutboxStatus
	Attempts       int
	MaxAttempts    int
	LastError      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	AcceptedAt     *time.Time
}

// PushIdempotencyKey builds the stable key for match+type+version grouping.
func PushIdempotencyKey(matchID, alertType, version string) string {
	return "push:" + matchID + ":" + alertType + ":" + version
}

package domain

import "time"

// AuditEvent is a maintainer action trail row (REQ-013).
type AuditEvent struct {
	ID         int64
	Actor      string
	Action     string
	EntityType string
	EntityID   string
	Reason     string
	BeforeJSON []byte
	AfterJSON  []byte
	CreatedAt  time.Time
}

// MatchOverride is a versioned human correction for a match surface (REQ-013).
type MatchOverride struct {
	ID            ID
	MatchID       ID
	DataType      string
	Field         string
	Value         string
	Justification string
	Actor         string
	Version       int
	AppliedAt     time.Time
}

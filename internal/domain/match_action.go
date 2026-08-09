package domain

import "time"

// MatchActionApply is the transactional write bundle for a maintainer match action.
type MatchActionApply struct {
	MatchID    ID
	Surface    string
	AfterState AvailabilityState
	Override   *MatchOverride
	Audit      AuditEvent
	Now        time.Time
}

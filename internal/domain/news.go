package domain

import "time"

// NewsLink is an allowlisted related article card (title + URL only).
type NewsLink struct {
	ID           ID
	MatchID      ID
	EvidenceID   ID
	SourceID     string
	Title        string
	URL          string
	PublishedAt  time.Time
	Availability AvailabilityState
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

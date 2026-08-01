package domain

import "time"

// SeasonClubMembership ties a club to an official season snapshot (REQ-002).
type SeasonClubMembership struct {
	Season    int
	ClubID    ID
	CreatedAt time.Time
}

package domain

import "time"

// LineupSide identifies which club a lineup belongs to.
type LineupSide string

const (
	LineupHome LineupSide = "home"
	LineupAway LineupSide = "away"
)

// Valid reports whether the side is known.
func (s LineupSide) Valid() bool {
	switch s {
	case LineupHome, LineupAway:
		return true
	default:
		return false
	}
}

// LineupPlayer is one named player entry in a lineup.
type LineupPlayer struct {
	ShirtNumber string
	Name        string
	IsStarter   bool
}

// Lineup is an official or candidate lineup observation for one side.
type Lineup struct {
	ID           ID
	MatchID      ID
	ClubID       ID
	EvidenceID   ID
	Side         LineupSide
	Formation    string
	Coach        string
	Players      []LineupPlayer
	Official     bool
	Availability AvailabilityState
	PublishedAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

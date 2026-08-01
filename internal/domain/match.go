package domain

import "time"

// KickoffState describes whether a kickoff instant is known and stable.
type KickoffState string

const (
	KickoffPublished  KickoffState = "published"
	KickoffIndefinite KickoffState = "indefinite"
	KickoffChanged    KickoffState = "changed"
)

// Valid reports whether the kickoff state is known.
func (s KickoffState) Valid() bool {
	switch s {
	case KickoffPublished, KickoffIndefinite, KickoffChanged:
		return true
	default:
		return false
	}
}

// Match is a scheduled fixture between two clubs in a competition.
type Match struct {
	ID            ID
	CompetitionID ID
	HomeClubID    ID
	AwayClubID    ID
	Slug          string
	Round         string
	Venue         string
	// KickoffAt is UTC when known; nil when KickoffState is indefinite.
	KickoffAt      *time.Time
	KickoffState   KickoffState
	BroadcastState AvailabilityState
	LineupState    AvailabilityState
	NewsState      AvailabilityState
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

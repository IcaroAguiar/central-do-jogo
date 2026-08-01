package domain

import "time"

// AccessType describes whether a broadcast requires a subscription.
type AccessType string

const (
	AccessFree         AccessType = "free"
	AccessSubscription AccessType = "subscription"
	AccessUnknown      AccessType = "unknown"
)

// Valid reports whether the access type is known.
func (a AccessType) Valid() bool {
	switch a {
	case AccessFree, AccessSubscription, AccessUnknown:
		return true
	default:
		return false
	}
}

// ConfidenceLevel is a deterministic confidence band for a claim.
type ConfidenceLevel string

const (
	ConfidenceHigh   ConfidenceLevel = "high"
	ConfidenceMedium ConfidenceLevel = "medium"
	ConfidenceLow    ConfidenceLevel = "low"
)

// Valid reports whether the confidence level is known.
func (c ConfidenceLevel) Valid() bool {
	switch c {
	case ConfidenceHigh, ConfidenceMedium, ConfidenceLow:
		return true
	default:
		return false
	}
}

// Broadcast is a transmission claim for a match, with provenance.
type Broadcast struct {
	ID           ID
	MatchID      ID
	EvidenceID   ID
	Channel      string
	Platform     string
	Access       AccessType
	Region       string
	OfficialURL  string
	Confidence   ConfidenceLevel
	VerifiedAt   time.Time
	Availability AvailabilityState
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

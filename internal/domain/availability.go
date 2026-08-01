package domain

// AvailabilityState models REQ-010 gap states for match-scoped data surfaces.
type AvailabilityState string

const (
	AvailabilityAvailable           AvailabilityState = "available"
	AvailabilityAwaitingPublication AvailabilityState = "awaiting_publication"
	AvailabilityNotFound            AvailabilityState = "not_found"
	AvailabilityDivergent           AvailabilityState = "divergent"
	AvailabilityNoCoverage          AvailabilityState = "no_coverage"
)

// Valid reports whether the state is one of the known REQ-010 values.
func (s AvailabilityState) Valid() bool {
	switch s {
	case AvailabilityAvailable,
		AvailabilityAwaitingPublication,
		AvailabilityNotFound,
		AvailabilityDivergent,
		AvailabilityNoCoverage:
		return true
	default:
		return false
	}
}

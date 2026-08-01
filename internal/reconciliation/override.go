package reconciliation

import (
	"fmt"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
)

// Override is a versioned human correction applied to a reconciliation result.
type Override struct {
	ID            string
	MatchID       string
	DataType      string
	Field         string
	Value         string
	Justification string
	Actor         string
	Version       int
	AppliedAt     time.Time
}

// OverrideStore is a minimal interface for persisting overrides.
type OverrideStore interface {
	Save(override Override) error
	FindActive(matchID, dataType, field string) (*Override, error)
}

// ApplyOverride applies a human override to a reconciliation result if one exists
// for the given match/data/field. Returns the possibly-modified result and whether
// an override was applied.
func ApplyOverride(result Result, override *Override) (Result, bool) {
	if override == nil {
		return result, false
	}
	return Result{
		Value:        override.Value,
		Confidence:   domain.ConfidenceHigh,
		Availability: domain.AvailabilityAvailable,
		WinnerSource: fmt.Sprintf("override:%s", override.Actor),
		Divergent:    false,
	}, true
}

// InMemoryOverrideStore is a simple in-memory implementation of OverrideStore
// suitable for tests and non-persistent contexts.
type InMemoryOverrideStore struct {
	overrides []Override
}

// NewInMemoryOverrideStore creates a new in-memory override store.
func NewInMemoryOverrideStore() *InMemoryOverrideStore {
	return &InMemoryOverrideStore{}
}

func (s *InMemoryOverrideStore) Save(override Override) error {
	if override.Justification == "" {
		return fmt.Errorf("override must have a justification")
	}
	if override.Actor == "" {
		return fmt.Errorf("override must have an actor")
	}
	s.overrides = append(s.overrides, override)
	return nil
}

func (s *InMemoryOverrideStore) FindActive(matchID, dataType, field string) (*Override, error) {
	var latest *Override
	for i := range s.overrides {
		o := &s.overrides[i]
		if o.MatchID == matchID && o.DataType == dataType && o.Field == field {
			if latest == nil || o.Version > latest.Version {
				latest = o
			}
		}
	}
	return latest, nil
}

// Package reconciliation implements PAT-002 deterministic rules for deriving
// confidence and availability state from multiple source observations.
package reconciliation

import (
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
)

// SourcePriority maps source IDs to integer priority (lower = higher priority).
// Sources not in the map get the lowest priority (MaxInt).
type SourcePriority map[string]int

// Claim represents a single observation's claim about a data surface.
type Claim struct {
	SourceID   string
	ObservedAt time.Time
	Value      string
	EvidenceID string
}

// Result is the reconciliation output for a data surface.
type Result struct {
	Value        string
	Confidence   domain.ConfidenceLevel
	Availability domain.AvailabilityState
	WinnerSource string
	Divergent    bool
}

// FreshnessWindow is the maximum age of an observation before it loses priority.
const FreshnessWindow = 6 * time.Hour

// Reconcile evaluates a set of claims using priority, freshness, and concordance.
// Returns the reconciled result or a divergent state when sources disagree.
func Reconcile(claims []Claim, priorities SourcePriority) Result {
	if len(claims) == 0 {
		return Result{
			Availability: domain.AvailabilityNotFound,
			Confidence:   domain.ConfidenceLow,
		}
	}

	now := time.Now()
	fresh := filterFresh(claims, now)
	if len(fresh) == 0 {
		fresh = claims
	}

	deduped := deduplicateBySource(fresh)
	sorted := sortByPriority(deduped, priorities)

	if concordant(sorted) {
		return Result{
			Value:        sorted[0].Value,
			Confidence:   confidenceFromCount(len(sorted)),
			Availability: domain.AvailabilityAvailable,
			WinnerSource: sorted[0].SourceID,
		}
	}

	top := highestPriorityClaims(sorted, priorities)
	if len(top) == 1 {
		return Result{
			Value:        top[0].Value,
			Confidence:   domain.ConfidenceMedium,
			Availability: domain.AvailabilityAvailable,
			WinnerSource: top[0].SourceID,
			Divergent:    true,
		}
	}

	return Result{
		Availability: domain.AvailabilityDivergent,
		Confidence:   domain.ConfidenceLow,
		Divergent:    true,
	}
}

// deduplicateBySource keeps only the most recent claim per source ID.
func deduplicateBySource(claims []Claim) []Claim {
	best := make(map[string]Claim)
	for _, c := range claims {
		if existing, ok := best[c.SourceID]; !ok || c.ObservedAt.After(existing.ObservedAt) {
			best[c.SourceID] = c
		}
	}
	result := make([]Claim, 0, len(best))
	for _, c := range best {
		result = append(result, c)
	}
	return result
}

func filterFresh(claims []Claim, now time.Time) []Claim {
	cutoff := now.Add(-FreshnessWindow)
	var result []Claim
	for _, c := range claims {
		if c.ObservedAt.After(cutoff) {
			result = append(result, c)
		}
	}
	return result
}

func sortByPriority(claims []Claim, priorities SourcePriority) []Claim {
	sorted := make([]Claim, len(claims))
	copy(sorted, claims)
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0; j-- {
			pi := priority(sorted[j].SourceID, priorities)
			pj := priority(sorted[j-1].SourceID, priorities)
			if pi < pj || (pi == pj && sorted[j].ObservedAt.After(sorted[j-1].ObservedAt)) {
				sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
			} else {
				break
			}
		}
	}
	return sorted
}

func priority(sourceID string, priorities SourcePriority) int {
	if p, ok := priorities[sourceID]; ok {
		return p
	}
	return 1<<31 - 1
}

func concordant(sorted []Claim) bool {
	if len(sorted) <= 1 {
		return true
	}
	v := sorted[0].Value
	for _, c := range sorted[1:] {
		if c.Value != v {
			return false
		}
	}
	return true
}

func highestPriorityClaims(sorted []Claim, priorities SourcePriority) []Claim {
	if len(sorted) == 0 {
		return nil
	}
	topPriority := priority(sorted[0].SourceID, priorities)
	var result []Claim
	for _, c := range sorted {
		if priority(c.SourceID, priorities) == topPriority {
			result = append(result, c)
		} else {
			break
		}
	}
	return result
}

func confidenceFromCount(n int) domain.ConfidenceLevel {
	switch {
	case n >= 3:
		return domain.ConfidenceHigh
	case n == 2:
		return domain.ConfidenceMedium
	default:
		return domain.ConfidenceLow
	}
}

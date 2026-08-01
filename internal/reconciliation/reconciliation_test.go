package reconciliation

import (
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
)

func TestReconcile(t *testing.T) {
	now := time.Now()
	priorities := SourcePriority{
		"cbf_official": 1,
		"cbf_match":    2,
		"openfootball": 3,
		"gazeta_rss":   4,
	}

	tests := []struct {
		name      string
		claims    []Claim
		wantAvail domain.AvailabilityState
		wantConf  domain.ConfidenceLevel
		wantValue string
		wantDiv   bool
	}{
		{
			name:      "no claims returns not_found",
			claims:    nil,
			wantAvail: domain.AvailabilityNotFound,
			wantConf:  domain.ConfidenceLow,
		},
		{
			name: "single claim low confidence",
			claims: []Claim{
				{SourceID: "cbf_official", ObservedAt: now, Value: "Globo"},
			},
			wantAvail: domain.AvailabilityAvailable,
			wantConf:  domain.ConfidenceLow,
			wantValue: "Globo",
		},
		{
			name: "two concordant claims medium confidence",
			claims: []Claim{
				{SourceID: "cbf_official", ObservedAt: now, Value: "Globo"},
				{SourceID: "cbf_match", ObservedAt: now.Add(-time.Hour), Value: "Globo"},
			},
			wantAvail: domain.AvailabilityAvailable,
			wantConf:  domain.ConfidenceMedium,
			wantValue: "Globo",
		},
		{
			name: "three concordant claims high confidence",
			claims: []Claim{
				{SourceID: "cbf_official", ObservedAt: now, Value: "Globo"},
				{SourceID: "cbf_match", ObservedAt: now.Add(-time.Hour), Value: "Globo"},
				{SourceID: "openfootball", ObservedAt: now.Add(-2 * time.Hour), Value: "Globo"},
			},
			wantAvail: domain.AvailabilityAvailable,
			wantConf:  domain.ConfidenceHigh,
			wantValue: "Globo",
		},
		{
			name: "divergent with clear priority winner",
			claims: []Claim{
				{SourceID: "cbf_official", ObservedAt: now, Value: "Globo"},
				{SourceID: "openfootball", ObservedAt: now, Value: "SporTV"},
			},
			wantAvail: domain.AvailabilityAvailable,
			wantConf:  domain.ConfidenceMedium,
			wantValue: "Globo",
			wantDiv:   true,
		},
		{
			name: "divergent with no clear winner returns divergent state",
			claims: []Claim{
				{SourceID: "unknown_a", ObservedAt: now, Value: "Globo"},
				{SourceID: "unknown_b", ObservedAt: now, Value: "SporTV"},
			},
			wantAvail: domain.AvailabilityDivergent,
			wantConf:  domain.ConfidenceLow,
			wantDiv:   true,
		},
		{
			name: "freshness filters stale claims",
			claims: []Claim{
				{SourceID: "cbf_official", ObservedAt: now.Add(-7 * time.Hour), Value: "old"},
				{SourceID: "openfootball", ObservedAt: now, Value: "fresh"},
			},
			wantAvail: domain.AvailabilityAvailable,
			wantConf:  domain.ConfidenceLow,
			wantValue: "fresh",
		},
		{
			name: "all stale falls back to all claims",
			claims: []Claim{
				{SourceID: "cbf_official", ObservedAt: now.Add(-8 * time.Hour), Value: "A"},
				{SourceID: "openfootball", ObservedAt: now.Add(-9 * time.Hour), Value: "A"},
			},
			wantAvail: domain.AvailabilityAvailable,
			wantConf:  domain.ConfidenceMedium,
			wantValue: "A",
		},
		{
			name: "same source deduplicates to most recent",
			claims: []Claim{
				{SourceID: "cbf_official", ObservedAt: now.Add(-time.Minute), Value: "early"},
				{SourceID: "cbf_official", ObservedAt: now, Value: "latest"},
			},
			wantAvail: domain.AvailabilityAvailable,
			wantConf:  domain.ConfidenceLow,
			wantValue: "latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Reconcile(tt.claims, priorities)
			if got.Availability != tt.wantAvail {
				t.Errorf("availability = %q, want %q", got.Availability, tt.wantAvail)
			}
			if got.Confidence != tt.wantConf {
				t.Errorf("confidence = %q, want %q", got.Confidence, tt.wantConf)
			}
			if tt.wantValue != "" && got.Value != tt.wantValue {
				t.Errorf("value = %q, want %q", got.Value, tt.wantValue)
			}
			if got.Divergent != tt.wantDiv {
				t.Errorf("divergent = %v, want %v", got.Divergent, tt.wantDiv)
			}
		})
	}
}

func TestApplyOverride(t *testing.T) {
	base := Result{
		Value:        "SporTV",
		Confidence:   domain.ConfidenceLow,
		Availability: domain.AvailabilityDivergent,
		Divergent:    true,
	}

	t.Run("nil override is no-op", func(t *testing.T) {
		got, applied := ApplyOverride(base, nil)
		if applied {
			t.Fatal("expected applied=false")
		}
		if got.Value != base.Value {
			t.Errorf("value changed unexpectedly")
		}
	})

	t.Run("override replaces result with high confidence", func(t *testing.T) {
		override := &Override{
			ID:            "ov-1",
			MatchID:       "m-1",
			DataType:      "broadcast",
			Field:         "channel",
			Value:         "Globo",
			Justification: "confirmed via official announcement",
			Actor:         "admin@example.com",
			Version:       1,
		}
		got, applied := ApplyOverride(base, override)
		if !applied {
			t.Fatal("expected applied=true")
		}
		if got.Value != "Globo" {
			t.Errorf("value = %q, want Globo", got.Value)
		}
		if got.Confidence != domain.ConfidenceHigh {
			t.Errorf("confidence = %q, want high", got.Confidence)
		}
		if got.Divergent {
			t.Error("divergent should be false after override")
		}
	})
}

func TestInMemoryOverrideStore(t *testing.T) {
	store := NewInMemoryOverrideStore()

	t.Run("save requires justification", func(t *testing.T) {
		err := store.Save(Override{Actor: "a"})
		if err == nil {
			t.Fatal("expected error for missing justification")
		}
	})

	t.Run("save requires actor", func(t *testing.T) {
		err := store.Save(Override{Justification: "j"})
		if err == nil {
			t.Fatal("expected error for missing actor")
		}
	})

	t.Run("find returns latest version", func(t *testing.T) {
		_ = store.Save(Override{
			MatchID: "m1", DataType: "broadcast", Field: "channel",
			Value: "v1", Version: 1, Actor: "a", Justification: "j",
		})
		_ = store.Save(Override{
			MatchID: "m1", DataType: "broadcast", Field: "channel",
			Value: "v2", Version: 2, Actor: "b", Justification: "j2",
		})

		got, err := store.FindActive("m1", "broadcast", "channel")
		if err != nil {
			t.Fatal(err)
		}
		if got == nil || got.Value != "v2" {
			t.Errorf("expected v2, got %+v", got)
		}
	})

	t.Run("find returns nil for unmatched", func(t *testing.T) {
		got, err := store.FindActive("m999", "broadcast", "channel")
		if err != nil {
			t.Fatal(err)
		}
		if got != nil {
			t.Errorf("expected nil, got %+v", got)
		}
	})
}

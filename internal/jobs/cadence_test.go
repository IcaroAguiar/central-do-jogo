package jobs

import (
	"testing"
	"time"
)

func TestNextRunAt(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		dataType  string
		kickoffAt *time.Time
		wantMin   time.Duration
		wantMax   time.Duration
	}{
		{
			name:     "schedule default",
			dataType: "schedule",
			wantMin:  CadenceScheduleDefault - time.Second,
			wantMax:  CadenceScheduleDefault + time.Second,
		},
		{
			name:     "lineup default",
			dataType: "lineup",
			wantMin:  CadenceLineupDefault - time.Second,
			wantMax:  CadenceLineupDefault + time.Second,
		},
		{
			name:     "news default",
			dataType: "news",
			wantMin:  CadenceNewsDefault - time.Second,
			wantMax:  CadenceNewsDefault + time.Second,
		},
		{
			name:     "schedule near kickoff",
			dataType: "schedule",
			kickoffAt: func() *time.Time {
				t := now.Add(2 * time.Hour)
				return &t
			}(),
			wantMin: CadenceScheduleNearKickoff - time.Second,
			wantMax: CadenceScheduleNearKickoff + time.Second,
		},
		{
			name:     "lineup near kickoff",
			dataType: "lineup",
			kickoffAt: func() *time.Time {
				t := now.Add(1 * time.Hour)
				return &t
			}(),
			wantMin: CadenceLineupNearKickoff - time.Second,
			wantMax: CadenceLineupNearKickoff + time.Second,
		},
		{
			name:     "news near kickoff",
			dataType: "news",
			kickoffAt: func() *time.Time {
				t := now.Add(2 * time.Hour)
				return &t
			}(),
			wantMin: CadenceNewsNearKickoff - time.Second,
			wantMax: CadenceNewsNearKickoff + time.Second,
		},
		{
			name:     "kickoff in past uses default",
			dataType: "lineup",
			kickoffAt: func() *time.Time {
				t := now.Add(-1 * time.Hour)
				return &t
			}(),
			wantMin: CadenceLineupDefault - time.Second,
			wantMax: CadenceLineupDefault + time.Second,
		},
		{
			name:     "kickoff far away uses default",
			dataType: "lineup",
			kickoffAt: func() *time.Time {
				t := now.Add(24 * time.Hour)
				return &t
			}(),
			wantMin: CadenceLineupDefault - time.Second,
			wantMax: CadenceLineupDefault + time.Second,
		},
		{
			name:     "unknown data type uses schedule default",
			dataType: "unknown",
			wantMin:  CadenceScheduleDefault - time.Second,
			wantMax:  CadenceScheduleDefault + time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NextRunAt(now, tt.dataType, tt.kickoffAt)
			diff := got.Sub(now)
			if diff < tt.wantMin || diff > tt.wantMax {
				t.Errorf("NextRunAt diff = %v, want between %v and %v", diff, tt.wantMin, tt.wantMax)
			}
		})
	}
}

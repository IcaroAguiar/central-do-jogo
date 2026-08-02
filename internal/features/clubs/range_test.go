package clubs

import (
	"testing"
	"time"
)

func TestParseRange(t *testing.T) {
	t.Parallel()
	cases := []struct {
		raw     string
		want    Range
		wantErr bool
	}{
		{"", RangeWeek, false},
		{"week", RangeWeek, false},
		{"month", RangeMonth, false},
		{"season", RangeSeason, false},
		{"bogus", "", true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.raw, func(t *testing.T) {
			t.Parallel()
			got, err := ParseRange(tc.raw)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ParseRange(%q) error = %v, wantErr %v", tc.raw, err, tc.wantErr)
			}
			if !tc.wantErr && got != tc.want {
				t.Fatalf("ParseRange(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestBoundsSeasonIsUnbounded(t *testing.T) {
	t.Parallel()
	start, end := bounds(RangeSeason, time.Now())
	if start != nil || end != nil {
		t.Fatalf("season bounds = (%v, %v), want (nil, nil)", start, end)
	}
}

func TestBoundsWeekAnchorsOnBrasiliaMidnight(t *testing.T) {
	t.Parallel()
	// 2026-03-10 02:30 UTC is 2026-03-09 23:30 in Brasilia (-03:00).
	now := time.Date(2026, 3, 10, 2, 30, 0, 0, time.UTC)
	start, end := bounds(RangeWeek, now)
	if start == nil || end == nil {
		t.Fatal("week bounds should not be nil")
	}

	wantStart := time.Date(2026, 3, 9, 3, 0, 0, 0, time.UTC) // 2026-03-09 00:00 BRT
	wantEnd := wantStart.AddDate(0, 0, 7)
	if !start.Equal(wantStart) {
		t.Fatalf("start = %v, want %v", start, wantStart)
	}
	if !end.Equal(wantEnd) {
		t.Fatalf("end = %v, want %v", end, wantEnd)
	}
}

func TestBoundsMonthCoversCalendarMonthInBrasilia(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 2, 15, 12, 0, 0, 0, time.UTC)
	start, end := bounds(RangeMonth, now)
	if start == nil || end == nil {
		t.Fatal("month bounds should not be nil")
	}

	wantStart := time.Date(2026, 2, 1, 3, 0, 0, 0, time.UTC) // 2026-02-01 00:00 BRT
	wantEnd := time.Date(2026, 3, 1, 3, 0, 0, 0, time.UTC)   // 2026-03-01 00:00 BRT
	if !start.Equal(wantStart) {
		t.Fatalf("start = %v, want %v", start, wantStart)
	}
	if !end.Equal(wantEnd) {
		t.Fatalf("end = %v, want %v", end, wantEnd)
	}
}

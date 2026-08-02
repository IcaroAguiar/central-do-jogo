package clubs

import (
	"fmt"
	"time"
)

// Range identifies an agenda window for a club's matches (REQ-004).
type Range string

// Supported agenda ranges.
const (
	RangeWeek   Range = "week"
	RangeMonth  Range = "month"
	RangeSeason Range = "season"
)

// ParseRange validates a range query parameter, defaulting to week when empty.
func ParseRange(raw string) (Range, error) {
	if raw == "" {
		return RangeWeek, nil
	}
	switch Range(raw) {
	case RangeWeek, RangeMonth, RangeSeason:
		return Range(raw), nil
	default:
		return "", fmt.Errorf("invalid range %q: must be week, month, or season", raw)
	}
}

// brasiliaOffset is the fixed UTC-3 offset used for Brasilia time. Brazil has
// not observed daylight saving time since 2019, so a fixed offset is
// deterministic and avoids depending on tzdata being present in the runtime
// image (CON-008 requires explicit Brasilia time).
var brasiliaZone = time.FixedZone("-03:00", -3*60*60)

// bounds computes the [start, end) UTC kickoff window for a range, anchored
// on "today" in Brasilia time as observed at `now`. Season returns (nil, nil)
// since it is not time-bounded.
func bounds(rng Range, now time.Time) (start, end *time.Time) {
	if rng == RangeSeason {
		return nil, nil
	}

	local := now.In(brasiliaZone)

	var startLocal, endLocal time.Time
	switch rng {
	case RangeMonth:
		startLocal = time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, brasiliaZone)
		endLocal = startLocal.AddDate(0, 1, 0)
	default: // RangeWeek
		startLocal = time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, brasiliaZone)
		endLocal = startLocal.AddDate(0, 0, 7)
	}

	startUTC := startLocal.UTC()
	endUTC := endLocal.UTC()
	return &startUTC, &endUTC
}

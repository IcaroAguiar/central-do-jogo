package jobs

import "time"

// Cadence intervals by data type.
const (
	CadenceScheduleDefault = 6 * time.Hour
	CadenceLineupDefault   = 2 * time.Hour
	CadenceNewsDefault     = 1 * time.Hour

	CadenceScheduleNearKickoff = 1 * time.Hour
	CadenceLineupNearKickoff   = 15 * time.Minute
	CadenceNewsNearKickoff     = 30 * time.Minute
)

// NearKickoffThreshold is the proximity window where adaptive cadence kicks in.
const NearKickoffThreshold = 3 * time.Hour

// NextRunAt computes the next run time for a given data type and optional
// kickoff proximity. When kickoffAt is non-nil and within NearKickoffThreshold,
// a shorter interval is used.
func NextRunAt(dataType string, kickoffAt *time.Time) time.Time {
	now := time.Now()
	near := kickoffAt != nil && kickoffAt.Sub(now) <= NearKickoffThreshold && kickoffAt.After(now)

	switch dataType {
	case "schedule":
		if near {
			return now.Add(CadenceScheduleNearKickoff)
		}
		return now.Add(CadenceScheduleDefault)
	case "lineup":
		if near {
			return now.Add(CadenceLineupNearKickoff)
		}
		return now.Add(CadenceLineupDefault)
	case "news":
		if near {
			return now.Add(CadenceNewsNearKickoff)
		}
		return now.Add(CadenceNewsDefault)
	default:
		return now.Add(CadenceScheduleDefault)
	}
}

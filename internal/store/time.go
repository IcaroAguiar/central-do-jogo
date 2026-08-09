package store

import "time"

// pgx decodes timestamptz using time.Unix internally, which yields a
// time.Time in the process's local Location even though the instant is
// correct. Public API responses must wire times in UTC, so every store
// normalizes scanned timestamps through these helpers before returning them.

func utc(t time.Time) time.Time {
	return t.UTC()
}

func utcPtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	u := t.UTC()
	return &u
}

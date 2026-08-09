package domain

import "time"

// SourceHealth tracks operational health for a source adapter (CON-006).
type SourceHealth struct {
	SourceID            string
	LastSuccessAt       *time.Time
	LastErrorAt         *time.Time
	LastError           string
	NextRunAt           time.Time
	ConsecutiveFailures int
	UpdatedAt           time.Time
}

package domain

import "time"

// Evidence is an immutable provenance record for an observed source value.
type Evidence struct {
	ID            ID
	SourceID      ID
	MatchID       *ID
	DataType      string
	ObservedAt    time.Time
	FetchedAt     time.Time
	ParserVersion string
	RunID         string
	ContentHash   string
	RawRef        string
	CreatedAt     time.Time
}

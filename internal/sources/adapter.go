// Package sources defines the source adapter contract for Central do Jogo.
// Every source adapter must produce domain observations with provenance metadata.
package sources

import (
	"context"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
)

// DataType classifies what a source adapter produces.
type DataType string

const (
	DataTypeSchedule DataType = "schedule"
	DataTypeLineup   DataType = "lineup"
	DataTypeNews     DataType = "news"
)

// Observation is a raw parsed result from a source, carrying evidence metadata.
type Observation struct {
	SourceID      string
	DataType      DataType
	ObservedAt    time.Time
	ParserVersion string
	ContentHash   string
	RawRef        string

	Schedules []ScheduleEntry
	Lineups   []LineupEntry
	NewsLinks []NewsLinkEntry
}

// ScheduleEntry is a parsed schedule fixture from a source.
type ScheduleEntry struct {
	HomeTeam    string
	AwayTeam    string
	Round       string
	Venue       string
	KickoffAt   *time.Time
	Competition string
}

// LineupEntry is a parsed lineup from a source.
type LineupEntry struct {
	HomeTeam  string
	AwayTeam  string
	Side      domain.LineupSide
	Formation string
	Coach     string
	Players   []domain.LineupPlayer
	Official  bool
}

// NewsLinkEntry is a parsed news link from a source.
type NewsLinkEntry struct {
	Title       string
	URL         string
	PublishedAt time.Time
}

// Adapter is the contract that every source must implement.
// Adapters parse pre-fetched data (fixtures in tests, fetched bytes in production)
// and produce typed observations with evidence metadata.
type Adapter interface {
	// SourceID returns the unique identifier matching the manifest.
	SourceID() string

	// Parse processes raw content and returns observations.
	// It must be deterministic given the same input.
	Parse(ctx context.Context, raw []byte) (*Observation, error)
}

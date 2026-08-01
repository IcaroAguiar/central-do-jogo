package domain

import "time"

// Club is a supported football club in an official season snapshot.
type Club struct {
	ID        ID
	Slug      string
	Name      string
	ShortName string
	Aliases   []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

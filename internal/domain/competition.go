package domain

import "time"

// Competition identifies a tournament season (for example Serie A 2026).
type Competition struct {
	ID        ID
	Slug      string
	Name      string
	Season    int
	CreatedAt time.Time
	UpdatedAt time.Time
}

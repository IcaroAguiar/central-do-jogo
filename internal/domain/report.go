package domain

import "time"

// Report is an anonymous visitor error report (REQ-014).
// Reports never auto-mutate product data.
type Report struct {
	ID          ID
	ContextType string
	ContextSlug string
	Message     string
	Status      string
	IPHash      string
	UserAgent   string
	CreatedAt   time.Time
	ReviewedAt  *time.Time
}

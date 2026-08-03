package domain

import "time"

// UserPreferences stores the account-backed club preferences for a user
// (REQ-006). Club identity uses public slugs so the contract matches the
// localStorage-backed visitor prefs and the public club API.
type UserPreferences struct {
	UserID            ID
	PrimaryClubSlug   *string
	FavoriteClubSlugs []string
	UpdatedAt         time.Time
}

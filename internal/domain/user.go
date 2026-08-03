package domain

import "time"

// UserRole is the authorization role for an authenticated account (REQ-018).
type UserRole string

const (
	RoleUser       UserRole = "user"
	RoleMaintainer UserRole = "maintainer"
)

// Valid reports whether the role is a known UserRole.
func (r UserRole) Valid() bool {
	switch r {
	case RoleUser, RoleMaintainer:
		return true
	default:
		return false
	}
}

// User is an account authenticated via an external OAuth provider (REQ-017).
// Passwords are never stored.
type User struct {
	ID              ID
	Provider        string
	ProviderSubject string
	Email           string
	DisplayName     string
	Role            UserRole
	CreatedAt       time.Time
	UpdatedAt       time.Time
	LastLoginAt     *time.Time
}

// Session is a server-side login session. The raw token is never persisted;
// only a one-way hash is stored.
type Session struct {
	ID        ID
	UserID    ID
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
}

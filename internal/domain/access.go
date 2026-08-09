package domain

import (
	"context"
	"errors"
)

// Shared access errors used across features via MaintainerGate (ADR 0002).
var (
	ErrAuthDisabled = errors.New("auth disabled")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("not found")
)

// MaintainerGate authorizes maintainer sessions without importing features/auth.
type MaintainerGate interface {
	Enabled() bool
	RequireMaintainer(ctx context.Context, sessionToken string) (*User, error)
	PublicBaseURL() string
}

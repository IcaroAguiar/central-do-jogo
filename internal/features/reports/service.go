// Package reports implements anonymous rate-limited error reports (REQ-014, TASK-032).
package reports

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
)

var (
	ErrAuthDisabled = errors.New("auth disabled")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrInvalidInput = errors.New("invalid input")
)

const (
	maxMessageRunes = 1000
	StatusOpen      = "open"
	StatusReviewed  = "reviewed"
	StatusDismissed = "dismissed"
)

// SessionResolver looks up the current user from an opaque session token.
type SessionResolver interface {
	Enabled() bool
	CurrentUser(ctx context.Context, sessionToken string) (*domain.User, error)
	PublicBaseURL() string
}

// ReportRepository persists reports.
type ReportRepository interface {
	Insert(ctx context.Context, report domain.Report) error
	ListOpen(ctx context.Context, limit int) ([]domain.Report, error)
	MarkReviewed(ctx context.Context, id domain.ID, status string, now time.Time) error
}

// CreateInput is a sanitized anonymous report payload.
type CreateInput struct {
	ContextType string
	ContextSlug string
	Message     string
	ClientIP    string
	UserAgent   string
}

// Service orchestrates report intake and maintainer queue reads.
type Service struct {
	sessions SessionResolver
	reports  ReportRepository
	now      func() time.Time
}

// NewService builds a reports service.
func NewService(sessions SessionResolver, reports ReportRepository, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{sessions: sessions, reports: reports, now: now}
}

// PublicBaseURL exposes the configured origin for CSRF checks on maintainer mutations.
func (s *Service) PublicBaseURL() string { return s.sessions.PublicBaseURL() }

// Create stores a sanitized anonymous report. It never mutates match/club data.
func (s *Service) Create(ctx context.Context, input CreateInput) error {
	contextType := strings.TrimSpace(input.ContextType)
	contextSlug := strings.TrimSpace(input.ContextSlug)
	message := strings.TrimSpace(input.Message)
	switch contextType {
	case "match", "club", "other":
	default:
		return ErrInvalidInput
	}
	if message == "" || utf8.RuneCountInString(message) > maxMessageRunes {
		return ErrInvalidInput
	}
	if len(contextSlug) > 128 {
		return ErrInvalidInput
	}
	id, err := domain.NewID("rpt_")
	if err != nil {
		return err
	}
	return s.reports.Insert(ctx, domain.Report{
		ID: id, ContextType: contextType, ContextSlug: contextSlug, Message: message,
		Status: StatusOpen, IPHash: hashIP(input.ClientIP), UserAgent: truncate(input.UserAgent, 256),
		CreatedAt: s.now().UTC(),
	})
}

// ListOpen returns the maintainer report queue.
func (s *Service) ListOpen(ctx context.Context, sessionToken string) ([]domain.Report, error) {
	if _, err := s.requireMaintainer(ctx, sessionToken); err != nil {
		return nil, err
	}
	return s.reports.ListOpen(ctx, 50)
}

// Review marks a report reviewed/dismissed without changing product data.
func (s *Service) Review(ctx context.Context, sessionToken, reportID, status string) error {
	if _, err := s.requireMaintainer(ctx, sessionToken); err != nil {
		return err
	}
	status = strings.TrimSpace(status)
	if status != StatusReviewed && status != StatusDismissed {
		return ErrInvalidInput
	}
	id := domain.ID(strings.TrimSpace(reportID))
	if id == "" {
		return ErrInvalidInput
	}
	return s.reports.MarkReviewed(ctx, id, status, s.now())
}

func (s *Service) requireMaintainer(ctx context.Context, sessionToken string) (*domain.User, error) {
	if !s.sessions.Enabled() {
		return nil, ErrAuthDisabled
	}
	user, err := s.sessions.CurrentUser(ctx, sessionToken)
	if err != nil {
		return nil, fmt.Errorf("resolve session: %w", err)
	}
	if user == nil {
		return nil, ErrUnauthorized
	}
	if user.Role != domain.RoleMaintainer {
		return nil, ErrForbidden
	}
	return user, nil
}

func hashIP(ip string) string {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(sum[:])
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

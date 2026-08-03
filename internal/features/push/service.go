// Package push implements Web Push consent, subscriptions, idempotent outbox,
// and expired-endpoint cleanup (REQ-011, REQ-012, TASK-029).
package push

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
)

// ErrPushDisabled is returned when VAPID keys are not configured.
var ErrPushDisabled = errors.New("push disabled")

// ErrUnauthorized is returned when the request has no valid session.
var ErrUnauthorized = errors.New("unauthorized")

// ErrAuthDisabled is returned when OAuth is not configured.
var ErrAuthDisabled = errors.New("auth disabled")

// ErrInvalidSubscription is returned for malformed subscription payloads.
var ErrInvalidSubscription = errors.New("invalid subscription")

// SessionResolver looks up the current user from a session token.
type SessionResolver interface {
	Enabled() bool
	CurrentUser(ctx context.Context, sessionToken string) (*domain.User, error)
	PublicBaseURL() string
}

// SubscriptionRepository persists browser push subscriptions.
type SubscriptionRepository interface {
	UpsertSubscription(ctx context.Context, sub domain.PushSubscription, now time.Time) (*domain.PushSubscription, error)
	ListActiveByUser(ctx context.Context, userID domain.ID) ([]domain.PushSubscription, error)
	ListActive(ctx context.Context, limit int) ([]domain.PushSubscription, error)
	DeleteByEndpoint(ctx context.Context, userID domain.ID, endpoint string) (bool, error)
	DisableByEndpoint(ctx context.Context, endpoint string, now time.Time) error
	DeleteDisabledBefore(ctx context.Context, cutoff time.Time) (int64, error)
}

// OutboxRepository persists idempotent alert rows.
type OutboxRepository interface {
	EnqueueOutbox(ctx context.Context, entry domain.PushOutboxEntry, now time.Time) (*domain.PushOutboxEntry, error)
	GetOutboxByIdempotencyKey(ctx context.Context, key string) (*domain.PushOutboxEntry, error)
	MarkOutboxAccepted(ctx context.Context, id domain.ID, now time.Time) error
	MarkOutboxFailure(ctx context.Context, id domain.ID, lastError string, now time.Time) error
}

// Deliverer sends a payload to one subscription (stub or real webpush).
type Deliverer interface {
	Deliver(ctx context.Context, sub domain.PushSubscription, payload []byte) DeliveryResult
}

// DeliveryResult is the push-service outcome for one endpoint (REQ-025).
type DeliveryResult struct {
	Accepted bool
	Gone     bool
	Err      error
}

// StubDeliverer always reports acceptance (local/dev without a push network).
type StubDeliverer struct{}

// Deliver implements Deliverer.
func (StubDeliverer) Deliver(context.Context, domain.PushSubscription, []byte) DeliveryResult {
	return DeliveryResult{Accepted: true}
}

// Config holds runtime push settings.
type Config struct {
	Enabled    bool
	PublicKey  string
	PrivateKey string
	Subject    string
}

// SubscribeInput is the browser PushSubscriptionJSON payload.
type SubscribeInput struct {
	Endpoint  string
	P256dh    string
	Auth      string
	UserAgent string
}

// Service orchestrates authenticated subscription APIs and outbox delivery.
type Service struct {
	sessions SessionResolver
	subs     SubscriptionRepository
	outbox   OutboxRepository
	deliver  Deliverer
	cfg      Config
	now      func() time.Time
}

// NewService builds a push service.
func NewService(
	sessions SessionResolver,
	subs SubscriptionRepository,
	outbox OutboxRepository,
	deliver Deliverer,
	cfg Config,
	now func() time.Time,
) *Service {
	if now == nil {
		now = time.Now
	}
	if deliver == nil {
		deliver = StubDeliverer{}
	}
	return &Service{
		sessions: sessions,
		subs:     subs,
		outbox:   outbox,
		deliver:  deliver,
		cfg:      cfg,
		now:      now,
	}
}

// PublicBaseURL exposes the CSRF bind origin for handlers.
func (s *Service) PublicBaseURL() string { return s.sessions.PublicBaseURL() }

// Enabled reports whether VAPID is configured.
func (s *Service) Enabled() bool { return s.cfg.Enabled }

// PublicKey returns the VAPID applicationServerKey for the browser.
func (s *Service) PublicKey() (string, error) {
	if !s.cfg.Enabled {
		return "", ErrPushDisabled
	}
	return s.cfg.PublicKey, nil
}

// ListSubscriptions returns the caller's active endpoints (no secrets).
func (s *Service) ListSubscriptions(ctx context.Context, sessionToken string) ([]domain.PushSubscription, error) {
	user, err := s.requireUser(ctx, sessionToken)
	if err != nil {
		return nil, err
	}
	if !s.cfg.Enabled {
		return nil, ErrPushDisabled
	}
	return s.subs.ListActiveByUser(ctx, user.ID)
}

// Subscribe upserts a browser subscription for the authenticated user.
func (s *Service) Subscribe(ctx context.Context, sessionToken string, in SubscribeInput) (*domain.PushSubscription, error) {
	user, err := s.requireUser(ctx, sessionToken)
	if err != nil {
		return nil, err
	}
	if !s.cfg.Enabled {
		return nil, ErrPushDisabled
	}
	in, err = normalizeSubscribe(in)
	if err != nil {
		return nil, err
	}
	id, err := domain.NewID("psub_")
	if err != nil {
		return nil, err
	}
	return s.subs.UpsertSubscription(ctx, domain.PushSubscription{
		ID: id, UserID: user.ID, Endpoint: in.Endpoint,
		P256dh: in.P256dh, Auth: in.Auth, UserAgent: in.UserAgent,
	}, s.now())
}

// Unsubscribe deletes the caller's subscription by endpoint URL.
func (s *Service) Unsubscribe(ctx context.Context, sessionToken, endpoint string) error {
	user, err := s.requireUser(ctx, sessionToken)
	if err != nil {
		return err
	}
	if !s.cfg.Enabled {
		return ErrPushDisabled
	}
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return ErrInvalidSubscription
	}
	_, err = s.subs.DeleteByEndpoint(ctx, user.ID, endpoint)
	return err
}

// EnqueueAlert records an idempotent outbox row and returns it (REQ-012).
func (s *Service) EnqueueAlert(ctx context.Context, matchID, alertType, version string, payload map[string]any) (*domain.PushOutboxEntry, error) {
	if !s.cfg.Enabled {
		return nil, ErrPushDisabled
	}
	alertType = strings.TrimSpace(alertType)
	version = strings.TrimSpace(version)
	matchID = strings.TrimSpace(matchID)
	if alertType == "" || version == "" || matchID == "" {
		return nil, fmt.Errorf("%w: matchId, alertType, and version are required", ErrInvalidSubscription)
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	id, err := domain.NewID("pout_")
	if err != nil {
		return nil, err
	}
	mid := domain.ID(matchID)
	return s.outbox.EnqueueOutbox(ctx, domain.PushOutboxEntry{
		ID:             id,
		IdempotencyKey: domain.PushIdempotencyKey(matchID, alertType, version),
		AlertType:      alertType,
		MatchID:        &mid,
		Version:        version,
		Payload:        raw,
	}, s.now())
}

// DeliverOutbox fans out one outbox payload to active subscriptions.
// Acceptance is measured at the push service boundary (REQ-025), not the device.
func (s *Service) DeliverOutbox(ctx context.Context, idempotencyKey string) error {
	entry, err := s.outbox.GetOutboxByIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		return err
	}
	if entry == nil {
		return fmt.Errorf("push outbox not found: %s", idempotencyKey)
	}
	if entry.Status == domain.PushOutboxAccepted {
		return nil
	}
	subs, err := s.subs.ListActive(ctx, 1000)
	if err != nil {
		return err
	}
	var firstErr error
	acceptedAny := false
	for _, sub := range subs {
		result := s.deliver.Deliver(ctx, sub, entry.Payload)
		if result.Gone {
			_ = s.subs.DisableByEndpoint(ctx, sub.Endpoint, s.now())
			continue
		}
		if result.Err != nil {
			if firstErr == nil {
				firstErr = result.Err
			}
			continue
		}
		if result.Accepted {
			acceptedAny = true
		}
	}
	if firstErr != nil && !acceptedAny {
		_ = s.outbox.MarkOutboxFailure(ctx, entry.ID, firstErr.Error(), s.now())
		return firstErr
	}
	return s.outbox.MarkOutboxAccepted(ctx, entry.ID, s.now())
}

// CleanupExpiredEndpoints deletes subscriptions disabled before the retention window.
func (s *Service) CleanupExpiredEndpoints(ctx context.Context, retention time.Duration) (int64, error) {
	if retention <= 0 {
		retention = 30 * 24 * time.Hour
	}
	return s.subs.DeleteDisabledBefore(ctx, s.now().Add(-retention))
}

func (s *Service) requireUser(ctx context.Context, sessionToken string) (*domain.User, error) {
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
	return user, nil
}

func normalizeSubscribe(in SubscribeInput) (SubscribeInput, error) {
	in.Endpoint = strings.TrimSpace(in.Endpoint)
	in.P256dh = strings.TrimSpace(in.P256dh)
	in.Auth = strings.TrimSpace(in.Auth)
	in.UserAgent = strings.TrimSpace(in.UserAgent)
	if in.Endpoint == "" || in.P256dh == "" || in.Auth == "" {
		return SubscribeInput{}, ErrInvalidSubscription
	}
	u, err := url.Parse(in.Endpoint)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return SubscribeInput{}, fmt.Errorf("%w: endpoint must be an absolute http(s) URL", ErrInvalidSubscription)
	}
	if len(in.Endpoint) > 2048 || len(in.P256dh) > 512 || len(in.Auth) > 512 {
		return SubscribeInput{}, fmt.Errorf("%w: field too long", ErrInvalidSubscription)
	}
	return in, nil
}

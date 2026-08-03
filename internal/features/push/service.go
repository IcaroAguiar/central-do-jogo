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

// ErrInvalidAlert is returned for malformed outbox enqueue inputs.
var ErrInvalidAlert = errors.New("invalid push alert")

// ErrEndpointOwned is returned when another account already owns the endpoint.
var ErrEndpointOwned = errors.New("push endpoint owned by another user")

// SessionResolver looks up the current user from a session token.
type SessionResolver interface {
	Enabled() bool
	CurrentUser(ctx context.Context, sessionToken string) (*domain.User, error)
	PublicBaseURL() string
}

// SubscriptionRepository persists browser push subscriptions.
type SubscriptionRepository interface {
	GetByEndpoint(ctx context.Context, endpoint string) (*domain.PushSubscription, error)
	UpsertSubscription(ctx context.Context, sub domain.PushSubscription, now time.Time) (*domain.PushSubscription, error)
	ListActiveByUser(ctx context.Context, userID domain.ID) ([]domain.PushSubscription, error)
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

// SubscribeInput is the browser PushSubscriptionJSON payload.
type SubscribeInput struct {
	Endpoint  string
	P256dh    string
	Auth      string
	UserAgent string
}

// alertPayload is the JSON shape stored in push_outbox.payload.
type alertPayload struct {
	UserIDs []string       `json:"userIds"`
	Title   string         `json:"title,omitempty"`
	Body    string         `json:"body,omitempty"`
	URL     string         `json:"url,omitempty"`
	Extra   map[string]any `json:"-"`
}

// Service handles authenticated subscription APIs (HTTP-facing).
type Service struct {
	sessions SessionResolver
	subs     SubscriptionRepository
	cfg      subscriptionConfig
	now      func() time.Time
}

type subscriptionConfig struct {
	Enabled   bool
	PublicKey string
}

// NewService builds the HTTP subscription service.
func NewService(sessions SessionResolver, subs SubscriptionRepository, enabled bool, publicKey string, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{
		sessions: sessions,
		subs:     subs,
		cfg:      subscriptionConfig{Enabled: enabled, PublicKey: publicKey},
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
	existing, err := s.subs.GetByEndpoint(ctx, in.Endpoint)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.UserID != user.ID {
		return nil, ErrEndpointOwned
	}
	id, err := domain.NewID("psub_")
	if err != nil {
		return nil, err
	}
	if existing != nil {
		id = existing.ID
	}
	out, err := s.subs.UpsertSubscription(ctx, domain.PushSubscription{
		ID: id, UserID: user.ID, Endpoint: in.Endpoint,
		P256dh: in.P256dh, Auth: in.Auth, UserAgent: in.UserAgent,
	}, s.now())
	if err != nil {
		// Concurrent claim race: store refuses to rewrite another user's row.
		if again, getErr := s.subs.GetByEndpoint(ctx, in.Endpoint); getErr == nil && again != nil && again.UserID != user.ID {
			return nil, ErrEndpointOwned
		}
		return nil, err
	}
	return out, nil
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

// OutboxRunner handles enqueue, audience-scoped delivery, and cleanup (worker-facing).
type OutboxRunner struct {
	subs    SubscriptionRepository
	outbox  OutboxRepository
	deliver Deliverer
	enabled bool
	now     func() time.Time
}

// NewOutboxRunner builds the worker-facing outbox processor (no session dependency).
func NewOutboxRunner(subs SubscriptionRepository, outbox OutboxRepository, deliver Deliverer, enabled bool, now func() time.Time) *OutboxRunner {
	if now == nil {
		now = time.Now
	}
	if deliver == nil {
		deliver = StubDeliverer{}
	}
	return &OutboxRunner{subs: subs, outbox: outbox, deliver: deliver, enabled: enabled, now: now}
}

// EnqueueAlert records an idempotent outbox row for explicit recipient user IDs (REQ-012).
func (r *OutboxRunner) EnqueueAlert(ctx context.Context, matchID, alertType, version string, userIDs []string, payload map[string]any) (*domain.PushOutboxEntry, error) {
	if !r.enabled {
		return nil, ErrPushDisabled
	}
	alertType = strings.TrimSpace(alertType)
	version = strings.TrimSpace(version)
	matchID = strings.TrimSpace(matchID)
	if alertType == "" || version == "" || matchID == "" {
		return nil, fmt.Errorf("%w: matchId, alertType, and version are required", ErrInvalidAlert)
	}
	if len(userIDs) == 0 {
		return nil, fmt.Errorf("%w: userIds audience is required (no global fan-out)", ErrInvalidAlert)
	}
	normalized := make([]string, 0, len(userIDs))
	seen := map[string]struct{}{}
	for _, id := range userIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		normalized = append(normalized, id)
	}
	if len(normalized) == 0 {
		return nil, fmt.Errorf("%w: userIds audience is required", ErrInvalidAlert)
	}
	body := map[string]any{}
	for k, v := range payload {
		body[k] = v
	}
	body["userIds"] = normalized
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	id, err := domain.NewID("pout_")
	if err != nil {
		return nil, err
	}
	mid := domain.ID(matchID)
	return r.outbox.EnqueueOutbox(ctx, domain.PushOutboxEntry{
		ID:             id,
		IdempotencyKey: domain.PushIdempotencyKey(matchID, alertType, version),
		AlertType:      alertType,
		MatchID:        &mid,
		Version:        version,
		Payload:        raw,
	}, r.now())
}

// DeliverOutbox fans out one outbox payload only to the listed recipient users.
func (r *OutboxRunner) DeliverOutbox(ctx context.Context, idempotencyKey string) error {
	entry, err := r.outbox.GetOutboxByIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		return err
	}
	if entry == nil {
		return fmt.Errorf("push outbox not found: %s", idempotencyKey)
	}
	if entry.Status == domain.PushOutboxAccepted {
		return nil
	}
	var parsed alertPayload
	if err := json.Unmarshal(entry.Payload, &parsed); err != nil {
		_ = r.outbox.MarkOutboxFailure(ctx, entry.ID, "invalid payload", r.now())
		return fmt.Errorf("decode push payload: %w", err)
	}
	if len(parsed.UserIDs) == 0 {
		_ = r.outbox.MarkOutboxFailure(ctx, entry.ID, "missing audience", r.now())
		return fmt.Errorf("%w: missing userIds audience", ErrInvalidAlert)
	}

	var targets []domain.PushSubscription
	for _, uid := range parsed.UserIDs {
		list, err := r.subs.ListActiveByUser(ctx, domain.ID(uid))
		if err != nil {
			_ = r.outbox.MarkOutboxFailure(ctx, entry.ID, err.Error(), r.now())
			return err
		}
		targets = append(targets, list...)
	}

	var firstErr error
	for _, sub := range targets {
		result := r.deliver.Deliver(ctx, sub, entry.Payload)
		if result.Gone {
			_ = r.subs.DisableByEndpoint(ctx, sub.Endpoint, r.now())
			continue
		}
		if result.Err != nil || !result.Accepted {
			if firstErr == nil {
				firstErr = result.Err
				if firstErr == nil {
					firstErr = errors.New("push delivery not accepted")
				}
			}
			continue
		}
	}
	if firstErr != nil {
		_ = r.outbox.MarkOutboxFailure(ctx, entry.ID, firstErr.Error(), r.now())
		return firstErr
	}
	return r.outbox.MarkOutboxAccepted(ctx, entry.ID, r.now())
}

// CleanupExpiredEndpoints deletes subscriptions disabled before the retention window.
func (r *OutboxRunner) CleanupExpiredEndpoints(ctx context.Context, retention time.Duration) (int64, error) {
	if retention <= 0 {
		retention = 30 * 24 * time.Hour
	}
	return r.subs.DeleteDisabledBefore(ctx, r.now().Add(-retention))
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

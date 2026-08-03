package push_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/push"
)

type memSessions struct {
	enabled bool
	user    *domain.User
	baseURL string
}

func (m *memSessions) Enabled() bool                         { return m.enabled }
func (m *memSessions) PublicBaseURL() string                 { return m.baseURL }
func (m *memSessions) CurrentUser(context.Context, string) (*domain.User, error) {
	return m.user, nil
}

type memPush struct {
	subs   map[string]domain.PushSubscription
	outbox map[string]domain.PushOutboxEntry
}

func newMemPush() *memPush {
	return &memPush{subs: map[string]domain.PushSubscription{}, outbox: map[string]domain.PushOutboxEntry{}}
}

func (m *memPush) UpsertSubscription(_ context.Context, sub domain.PushSubscription, now time.Time) (*domain.PushSubscription, error) {
	sub.CreatedAt = now
	sub.LastSeenAt = now
	m.subs[sub.Endpoint] = sub
	cp := sub
	return &cp, nil
}
func (m *memPush) ListActiveByUser(_ context.Context, userID domain.ID) ([]domain.PushSubscription, error) {
	var out []domain.PushSubscription
	for _, s := range m.subs {
		if s.UserID == userID && s.DisabledAt == nil {
			out = append(out, s)
		}
	}
	return out, nil
}
func (m *memPush) ListActive(context.Context, int) ([]domain.PushSubscription, error) {
	var out []domain.PushSubscription
	for _, s := range m.subs {
		if s.DisabledAt == nil {
			out = append(out, s)
		}
	}
	return out, nil
}
func (m *memPush) DeleteByEndpoint(_ context.Context, userID domain.ID, endpoint string) (bool, error) {
	s, ok := m.subs[endpoint]
	if !ok || s.UserID != userID {
		return false, nil
	}
	delete(m.subs, endpoint)
	return true, nil
}
func (m *memPush) DisableByEndpoint(_ context.Context, endpoint string, now time.Time) error {
	s, ok := m.subs[endpoint]
	if !ok {
		return nil
	}
	s.DisabledAt = &now
	m.subs[endpoint] = s
	return nil
}
func (m *memPush) DeleteDisabledBefore(_ context.Context, cutoff time.Time) (int64, error) {
	var n int64
	for k, s := range m.subs {
		if s.DisabledAt != nil && s.DisabledAt.Before(cutoff) {
			delete(m.subs, k)
			n++
		}
	}
	return n, nil
}
func (m *memPush) EnqueueOutbox(_ context.Context, entry domain.PushOutboxEntry, now time.Time) (*domain.PushOutboxEntry, error) {
	if existing, ok := m.outbox[entry.IdempotencyKey]; ok {
		cp := existing
		return &cp, nil
	}
	entry.Status = domain.PushOutboxPending
	entry.CreatedAt = now
	entry.UpdatedAt = now
	m.outbox[entry.IdempotencyKey] = entry
	cp := entry
	return &cp, nil
}
func (m *memPush) GetOutboxByIdempotencyKey(_ context.Context, key string) (*domain.PushOutboxEntry, error) {
	e, ok := m.outbox[key]
	if !ok {
		return nil, nil
	}
	cp := e
	return &cp, nil
}
func (m *memPush) MarkOutboxAccepted(_ context.Context, id domain.ID, now time.Time) error {
	for k, e := range m.outbox {
		if e.ID == id {
			e.Status = domain.PushOutboxAccepted
			e.AcceptedAt = &now
			e.UpdatedAt = now
			m.outbox[k] = e
		}
	}
	return nil
}
func (m *memPush) MarkOutboxFailure(_ context.Context, id domain.ID, lastError string, now time.Time) error {
	for k, e := range m.outbox {
		if e.ID == id {
			e.Attempts++
			e.LastError = lastError
			e.UpdatedAt = now
			if e.Attempts >= e.MaxAttempts {
				e.Status = domain.PushOutboxDead
			} else {
				e.Status = domain.PushOutboxFailed
			}
			m.outbox[k] = e
		}
	}
	return nil
}

type goneDeliverer struct{}

func (goneDeliverer) Deliver(context.Context, domain.PushSubscription, []byte) push.DeliveryResult {
	return push.DeliveryResult{Gone: true}
}

func TestSubscribeRequiresAuthAndValidEndpoint(t *testing.T) {
	t.Parallel()
	repo := newMemPush()
	svc := push.NewService(
		&memSessions{enabled: true, baseURL: "http://localhost"},
		repo, repo, nil,
		push.Config{Enabled: true, PublicKey: "pub"},
		time.Now,
	)
	_, err := svc.Subscribe(context.Background(), "tok", push.SubscribeInput{
		Endpoint: "https://push.example/1", P256dh: "k", Auth: "a",
	})
	if !errors.Is(err, push.ErrUnauthorized) {
		t.Fatalf("err = %v", err)
	}

	svc = push.NewService(
		&memSessions{enabled: true, user: &domain.User{ID: "u1"}, baseURL: "http://localhost"},
		repo, repo, nil,
		push.Config{Enabled: true, PublicKey: "pub"},
		time.Now,
	)
	_, err = svc.Subscribe(context.Background(), "tok", push.SubscribeInput{
		Endpoint: "not-a-url", P256dh: "k", Auth: "a",
	})
	if !errors.Is(err, push.ErrInvalidSubscription) {
		t.Fatalf("err = %v", err)
	}
}

func TestEnqueueIdempotentAndDeliverDisablesGone(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 3, 15, 0, 0, 0, time.UTC)
	repo := newMemPush()
	user := &domain.User{ID: "u1"}
	svc := push.NewService(
		&memSessions{enabled: true, user: user, baseURL: "http://localhost"},
		repo, repo, goneDeliverer{},
		push.Config{Enabled: true, PublicKey: "pub"},
		func() time.Time { return now },
	)
	_, err := svc.Subscribe(context.Background(), "tok", push.SubscribeInput{
		Endpoint: "https://push.example/gone", P256dh: "k", Auth: "a",
	})
	if err != nil {
		t.Fatal(err)
	}
	first, err := svc.EnqueueAlert(context.Background(), "mtc_1", domain.PushAlertLineupOfficial, "v1", map[string]any{"t": "1"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.EnqueueAlert(context.Background(), "mtc_1", domain.PushAlertLineupOfficial, "v1", map[string]any{"t": "2"})
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID {
		t.Fatalf("idempotency failed")
	}
	if err := svc.DeliverOutbox(context.Background(), first.IdempotencyKey); err != nil {
		t.Fatal(err)
	}
	active, _ := repo.ListActiveByUser(context.Background(), user.ID)
	if len(active) != 0 {
		t.Fatalf("expected gone endpoint disabled, got %d", len(active))
	}
}

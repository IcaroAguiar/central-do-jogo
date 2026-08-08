package push_test

import (
	"context"
	"encoding/json"
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

func (m *memSessions) Enabled() bool         { return m.enabled }
func (m *memSessions) PublicBaseURL() string { return m.baseURL }
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

func (m *memPush) GetByEndpoint(_ context.Context, endpoint string) (*domain.PushSubscription, error) {
	s, ok := m.subs[endpoint]
	if !ok {
		return nil, nil
	}
	cp := s
	return &cp, nil
}

func (m *memPush) UpsertSubscription(_ context.Context, sub domain.PushSubscription, now time.Time) (*domain.PushSubscription, error) {
	if existing, ok := m.subs[sub.Endpoint]; ok && existing.UserID != sub.UserID {
		return nil, errors.New("upsert push subscription: endpoint owned by another user")
	}
	if existing, ok := m.subs[sub.Endpoint]; ok {
		sub.ID = existing.ID
		sub.UserID = existing.UserID
		sub.CreatedAt = existing.CreatedAt
	} else {
		sub.CreatedAt = now
	}
	sub.LastSeenAt = now
	sub.DisabledAt = nil
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

func (m *memPush) UpdateOutboxPayload(_ context.Context, id domain.ID, payload []byte, now time.Time) error {
	for k, e := range m.outbox {
		if e.ID == id {
			e.Payload = append([]byte(nil), payload...)
			e.UpdatedAt = now
			m.outbox[k] = e
		}
	}
	return nil
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

type partialFailDeliverer struct {
	failEndpoint string
}

func (p partialFailDeliverer) Deliver(_ context.Context, sub domain.PushSubscription, _ []byte) push.DeliveryResult {
	if sub.Endpoint == p.failEndpoint {
		return push.DeliveryResult{Err: errors.New("temporary failure")}
	}
	return push.DeliveryResult{Accepted: true}
}

func TestSubscribeRequiresAuthAndValidEndpoint(t *testing.T) {
	t.Parallel()
	repo := newMemPush()
	svc := push.NewService(
		&memSessions{enabled: true, baseURL: "http://localhost"},
		repo,
		true,
		"pub",
		time.Now,
	)
	_, err := svc.Subscribe(context.Background(), "tok", push.SubscribeInput{
		Endpoint: "https://fcm.googleapis.com/fcm/send/1", P256dh: "k", Auth: "a",
	})
	if !errors.Is(err, push.ErrUnauthorized) {
		t.Fatalf("err = %v", err)
	}

	svc = push.NewService(
		&memSessions{enabled: true, user: &domain.User{ID: "u1"}, baseURL: "http://localhost"},
		repo,
		true,
		"pub",
		time.Now,
	)
	_, err = svc.Subscribe(context.Background(), "tok", push.SubscribeInput{
		Endpoint: "not-a-url", P256dh: "k", Auth: "a",
	})
	if !errors.Is(err, push.ErrInvalidSubscription) {
		t.Fatalf("err = %v", err)
	}
}

func TestSubscribeRejectsEndpointOwnedByAnotherUser(t *testing.T) {
	t.Parallel()
	repo := newMemPush()
	now := time.Date(2026, 8, 3, 15, 0, 0, 0, time.UTC)
	repo.subs["https://fcm.googleapis.com/fcm/send/shared"] = domain.PushSubscription{
		ID: "psub_a", UserID: "u1", Endpoint: "https://fcm.googleapis.com/fcm/send/shared",
		P256dh: "k", Auth: "a", CreatedAt: now, LastSeenAt: now,
	}
	svc := push.NewService(
		&memSessions{enabled: true, user: &domain.User{ID: "u2"}, baseURL: "http://localhost"},
		repo,
		true,
		"pub",
		func() time.Time { return now },
	)
	_, err := svc.Subscribe(context.Background(), "tok", push.SubscribeInput{
		Endpoint: "https://fcm.googleapis.com/fcm/send/shared", P256dh: "k2", Auth: "a2",
	})
	if !errors.Is(err, push.ErrEndpointOwned) {
		t.Fatalf("err = %v", err)
	}
	if repo.subs["https://fcm.googleapis.com/fcm/send/shared"].UserID != "u1" {
		t.Fatalf("owner changed")
	}
}

func TestEnqueueIdempotentAndDeliverDisablesGone(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 3, 15, 0, 0, 0, time.UTC)
	repo := newMemPush()
	user := &domain.User{ID: "u1"}
	svc := push.NewService(
		&memSessions{enabled: true, user: user, baseURL: "http://localhost"},
		repo,
		true,
		"pub",
		func() time.Time { return now },
	)
	runner := push.NewOutboxRunner(repo, repo, goneDeliverer{}, true, func() time.Time { return now })
	_, err := svc.Subscribe(context.Background(), "tok", push.SubscribeInput{
		Endpoint: "https://fcm.googleapis.com/fcm/send/gone", P256dh: "k", Auth: "a",
	})
	if err != nil {
		t.Fatal(err)
	}
	first, err := runner.EnqueueAlert(context.Background(), "mtc_1", domain.PushAlertLineupOfficial, "v1", []string{"u1"}, push.AlertContent{Title: "1"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := runner.EnqueueAlert(context.Background(), "mtc_1", domain.PushAlertLineupOfficial, "v1", []string{"u1"}, push.AlertContent{Title: "2"})
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID {
		t.Fatalf("idempotency failed")
	}
	if err := runner.DeliverOutbox(context.Background(), first.IdempotencyKey); err != nil {
		t.Fatal(err)
	}
	active, _ := repo.ListActiveByUser(context.Background(), user.ID)
	if len(active) != 0 {
		t.Fatalf("expected gone endpoint disabled, got %d", len(active))
	}
}

func TestDeliverRequiresAudienceAndDoesNotGlobalFanOut(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 3, 16, 0, 0, 0, time.UTC)
	repo := newMemPush()
	repo.subs["https://fcm.googleapis.com/fcm/send/a"] = domain.PushSubscription{
		ID: "psub_a", UserID: "u1", Endpoint: "https://fcm.googleapis.com/fcm/send/a",
		P256dh: "k", Auth: "a", CreatedAt: now, LastSeenAt: now,
	}
	repo.subs["https://fcm.googleapis.com/fcm/send/b"] = domain.PushSubscription{
		ID: "psub_b", UserID: "u2", Endpoint: "https://fcm.googleapis.com/fcm/send/b",
		P256dh: "k", Auth: "a", CreatedAt: now, LastSeenAt: now,
	}
	runner := push.NewOutboxRunner(repo, repo, push.StubDeliverer{}, true, func() time.Time { return now })
	_, err := runner.EnqueueAlert(context.Background(), "mtc_1", domain.PushAlertLineupOfficial, "v1", nil, push.AlertContent{})
	if !errors.Is(err, push.ErrInvalidAlert) {
		t.Fatalf("err = %v", err)
	}
	entry, err := runner.EnqueueAlert(context.Background(), "mtc_1", domain.PushAlertLineupOfficial, "v1", []string{"u1"}, push.AlertContent{Title: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.DeliverOutbox(context.Background(), entry.IdempotencyKey); err != nil {
		t.Fatal(err)
	}
	got := repo.outbox[entry.IdempotencyKey]
	if got.Status != domain.PushOutboxAccepted {
		t.Fatalf("status = %s", got.Status)
	}
}

func TestDeliverPartialFailureDoesNotAccept(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 3, 17, 0, 0, 0, time.UTC)
	repo := newMemPush()
	repo.subs["https://fcm.googleapis.com/fcm/send/ok"] = domain.PushSubscription{
		ID: "psub_ok", UserID: "u1", Endpoint: "https://fcm.googleapis.com/fcm/send/ok",
		P256dh: "k", Auth: "a", CreatedAt: now, LastSeenAt: now,
	}
	repo.subs["https://fcm.googleapis.com/fcm/send/bad"] = domain.PushSubscription{
		ID: "psub_bad", UserID: "u1", Endpoint: "https://fcm.googleapis.com/fcm/send/bad",
		P256dh: "k", Auth: "a", CreatedAt: now, LastSeenAt: now,
	}
	runner := push.NewOutboxRunner(repo, repo, partialFailDeliverer{failEndpoint: "https://fcm.googleapis.com/fcm/send/bad"}, true, func() time.Time { return now })
	entry, err := runner.EnqueueAlert(context.Background(), "mtc_2", domain.PushAlertLineupOfficial, "v1", []string{"u1"}, push.AlertContent{})
	if err != nil {
		t.Fatal(err)
	}
	err = runner.DeliverOutbox(context.Background(), entry.IdempotencyKey)
	if err == nil {
		t.Fatal("expected delivery error")
	}
	got := repo.outbox[entry.IdempotencyKey]
	if got.Status == domain.PushOutboxAccepted {
		t.Fatalf("partial failure must not accept")
	}
	var body map[string]any
	if err := json.Unmarshal(got.Payload, &body); err != nil {
		t.Fatal(err)
	}
	delivered, _ := body["deliveredEndpoints"].([]any)
	if len(delivered) != 1 || delivered[0] != "https://fcm.googleapis.com/fcm/send/ok" {
		t.Fatalf("deliveredEndpoints = %#v", body["deliveredEndpoints"])
	}

	// Retry must not re-deliver the already-accepted endpoint.
	counting := &countingDeliverer{failEndpoint: "https://fcm.googleapis.com/fcm/send/bad"}
	runner = push.NewOutboxRunner(repo, repo, counting, true, func() time.Time { return now })
	_ = runner.DeliverOutbox(context.Background(), entry.IdempotencyKey)
	if counting.hits["https://fcm.googleapis.com/fcm/send/ok"] != 0 {
		t.Fatalf("retry re-delivered ok endpoint: %#v", counting.hits)
	}
	if counting.hits["https://fcm.googleapis.com/fcm/send/bad"] != 1 {
		t.Fatalf("retry should attempt failed endpoint once: %#v", counting.hits)
	}
}

type countingDeliverer struct {
	failEndpoint string
	hits         map[string]int
}

func (c *countingDeliverer) Deliver(_ context.Context, sub domain.PushSubscription, _ []byte) push.DeliveryResult {
	if c.hits == nil {
		c.hits = map[string]int{}
	}
	c.hits[sub.Endpoint]++
	if sub.Endpoint == c.failEndpoint {
		return push.DeliveryResult{Err: errors.New("temporary failure")}
	}
	return push.DeliveryResult{Accepted: true}
}

type captureDeliverer struct {
	payloads [][]byte
}

func (c *captureDeliverer) Deliver(_ context.Context, _ domain.PushSubscription, payload []byte) push.DeliveryResult {
	cp := append([]byte(nil), payload...)
	c.payloads = append(c.payloads, cp)
	return push.DeliveryResult{Accepted: true}
}

func TestDeliverOmitsAudienceFromClientPayload(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 3, 18, 0, 0, 0, time.UTC)
	repo := newMemPush()
	repo.subs["https://fcm.googleapis.com/fcm/send/a"] = domain.PushSubscription{
		ID: "psub_a", UserID: "u1", Endpoint: "https://fcm.googleapis.com/fcm/send/a",
		P256dh: "k", Auth: "a", CreatedAt: now, LastSeenAt: now,
	}
	cap := &captureDeliverer{}
	runner := push.NewOutboxRunner(repo, repo, cap, true, func() time.Time { return now })
	entry, err := runner.EnqueueAlert(context.Background(), "mtc_3", domain.PushAlertLineupOfficial, "v1", []string{"u1", "u2"}, push.AlertContent{
		Title: "Escalação", Body: "Oficial", URL: "/matches/x",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.DeliverOutbox(context.Background(), entry.IdempotencyKey); err != nil {
		t.Fatal(err)
	}
	if len(cap.payloads) != 1 {
		t.Fatalf("deliveries = %d", len(cap.payloads))
	}
	var body map[string]any
	if err := json.Unmarshal(cap.payloads[0], &body); err != nil {
		t.Fatal(err)
	}
	if _, ok := body["userIds"]; ok {
		t.Fatalf("client payload leaked userIds: %s", cap.payloads[0])
	}
	if body["title"] != "Escalação" || body["url"] != "/matches/x" {
		t.Fatalf("client payload = %s", cap.payloads[0])
	}
}

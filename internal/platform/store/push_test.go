package store_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/store"
)

func TestPushStoreSubscriptionAndOutbox(t *testing.T) {
	pool := openTestPool(t)
	users := store.NewUserStore(pool)
	push := store.NewPushStore(pool)
	ctx := context.Background()
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)

	user, err := users.UpsertByProviderSubject(ctx, domain.User{
		ID: "usr_push_1", Provider: "google", ProviderSubject: "sub-push",
		Email: "push@example.com", DisplayName: "Push", Role: domain.RoleUser,
	}, now)
	if err != nil {
		t.Fatal(err)
	}

	sub, err := push.UpsertSubscription(ctx, domain.PushSubscription{
		ID: "psub_1", UserID: user.ID, Endpoint: "https://push.example/1",
		P256dh: "p256", Auth: "auth", UserAgent: "test",
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	if sub.Endpoint != "https://push.example/1" || sub.DisabledAt != nil {
		t.Fatalf("sub = %+v", sub)
	}

	listed, err := push.ListActiveByUser(ctx, user.ID)
	if err != nil || len(listed) != 1 {
		t.Fatalf("list = %v %v", listed, err)
	}

	matchID := domain.ID("mtc_1")
	key := domain.PushIdempotencyKey("mtc_1", domain.PushAlertLineupOfficial, "v1")
	payload, _ := json.Marshal(map[string]string{"title": "Escalação"})
	first, err := push.EnqueueOutbox(ctx, domain.PushOutboxEntry{
		ID: "pout_1", IdempotencyKey: key, AlertType: domain.PushAlertLineupOfficial,
		MatchID: &matchID, Version: "v1", Payload: payload,
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	second, err := push.EnqueueOutbox(ctx, domain.PushOutboxEntry{
		ID: "pout_ignored", IdempotencyKey: key, AlertType: domain.PushAlertLineupOfficial,
		MatchID: &matchID, Version: "v1", Payload: payload,
	}, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if second.ID != first.ID {
		t.Fatalf("idempotency broken: %s vs %s", second.ID, first.ID)
	}

	if err := push.DisableByEndpoint(ctx, sub.Endpoint, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	active, err := push.ListActiveByUser(ctx, user.ID)
	if err != nil || len(active) != 0 {
		t.Fatalf("active after disable = %v %v", active, err)
	}
	n, err := push.DeleteDisabledBefore(ctx, now.Add(2*time.Hour))
	if err != nil || n != 1 {
		t.Fatalf("cleanup = %d %v", n, err)
	}
}

package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/store"
)

func TestUserStoreUpsertAndSession(t *testing.T) {
	pool := openTestPool(t)
	s := store.NewUserStore(pool)
	ctx := context.Background()
	now := time.Date(2026, 8, 2, 15, 0, 0, 0, time.UTC)

	user, err := s.UpsertByProviderSubject(ctx, domain.User{
		ID:              "usr_test_1",
		Provider:        "google",
		ProviderSubject: "sub-abc",
		Email:           "fan@example.com",
		DisplayName:     "Fan",
		Role:            domain.RoleUser,
	}, now)
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if user.Email != "fan@example.com" || user.Role != domain.RoleUser {
		t.Fatalf("user = %+v", user)
	}

	promoted, err := s.UpsertByProviderSubject(ctx, domain.User{
		ID:              "usr_ignored",
		Provider:        "google",
		ProviderSubject: "sub-abc",
		Email:           "fan@example.com",
		DisplayName:     "Fan",
		Role:            domain.RoleMaintainer,
	}, now.Add(time.Minute))
	if err != nil {
		t.Fatalf("Upsert promote: %v", err)
	}
	if promoted.ID != user.ID {
		t.Fatalf("id changed on upsert: %s vs %s", promoted.ID, user.ID)
	}
	if promoted.Role != domain.RoleMaintainer {
		t.Fatalf("role = %s", promoted.Role)
	}

	if err := s.CreateSession(ctx, domain.Session{
		ID:        "ses_test_1",
		UserID:    user.ID,
		TokenHash: "hash-1",
		ExpiresAt: now.Add(time.Hour),
		CreatedAt: now,
	}); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	sess, err := s.GetSessionByTokenHash(ctx, "hash-1", now)
	if err != nil || sess == nil {
		t.Fatalf("GetSession: %v %+v", err, sess)
	}
	if err := s.RevokeSession(ctx, "hash-1", now.Add(time.Second)); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
	gone, err := s.GetSessionByTokenHash(ctx, "hash-1", now.Add(2*time.Second))
	if err != nil {
		t.Fatalf("GetSession after revoke: %v", err)
	}
	if gone != nil {
		t.Fatalf("expected nil session, got %+v", gone)
	}
}

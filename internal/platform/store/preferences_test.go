package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/store"
)

func TestPreferencesStoreUpsertAndGet(t *testing.T) {
	pool := openTestPool(t)
	users := store.NewUserStore(pool)
	prefs := store.NewPreferencesStore(pool)
	ctx := context.Background()
	now := time.Date(2026, 8, 2, 18, 0, 0, 0, time.UTC)

	user, err := users.UpsertByProviderSubject(ctx, domain.User{
		ID:              "usr_prefs_1",
		Provider:        "google",
		ProviderSubject: "sub-prefs",
		Email:           "prefs@example.com",
		DisplayName:     "Prefs",
		Role:            domain.RoleUser,
	}, now)
	if err != nil {
		t.Fatalf("Upsert user: %v", err)
	}

	primary := "flamengo"
	saved, err := prefs.Upsert(ctx, domain.UserPreferences{
		UserID:            user.ID,
		PrimaryClubSlug:   &primary,
		FavoriteClubSlugs: []string{"vasco", "flamengo"},
	}, now)
	if err != nil {
		t.Fatalf("Upsert prefs: %v", err)
	}
	if saved.PrimaryClubSlug == nil || *saved.PrimaryClubSlug != "flamengo" {
		t.Fatalf("primary = %v", saved.PrimaryClubSlug)
	}
	if len(saved.FavoriteClubSlugs) != 2 {
		t.Fatalf("favorites = %v", saved.FavoriteClubSlugs)
	}

	got, err := prefs.GetByUserID(ctx, user.ID)
	if err != nil || got == nil {
		t.Fatalf("GetByUserID: %v %+v", err, got)
	}
	if got.PrimaryClubSlug == nil || *got.PrimaryClubSlug != primary {
		t.Fatalf("got primary = %v", got.PrimaryClubSlug)
	}

	cleared, err := prefs.Upsert(ctx, domain.UserPreferences{
		UserID:            user.ID,
		PrimaryClubSlug:   nil,
		FavoriteClubSlugs: []string{},
	}, now.Add(time.Minute))
	if err != nil {
		t.Fatalf("clear: %v", err)
	}
	if cleared.PrimaryClubSlug != nil || len(cleared.FavoriteClubSlugs) != 0 {
		t.Fatalf("cleared = %+v", cleared)
	}
}

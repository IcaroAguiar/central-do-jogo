package preferences_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/preferences"
)

type memSessions struct {
	enabled bool
	user    *domain.User
	baseURL string
}

func (m *memSessions) Enabled() bool { return m.enabled }
func (m *memSessions) PublicBaseURL() string {
	return m.baseURL
}
func (m *memSessions) CurrentUser(context.Context, string) (*domain.User, error) {
	return m.user, nil
}

type memPrefs struct {
	row *domain.UserPreferences
}

func (m *memPrefs) GetByUserID(context.Context, domain.ID) (*domain.UserPreferences, error) {
	return m.row, nil
}
func (m *memPrefs) Upsert(_ context.Context, prefs domain.UserPreferences, now time.Time) (*domain.UserPreferences, error) {
	cp := prefs
	cp.UpdatedAt = now
	if cp.FavoriteClubSlugs == nil {
		cp.FavoriteClubSlugs = []string{}
	}
	m.row = &cp
	return &cp, nil
}

type memClubs struct {
	slugs map[string]struct{}
}

func (m *memClubs) GetBySlug(_ context.Context, slug string) (*domain.Club, error) {
	if _, ok := m.slugs[slug]; !ok {
		return nil, nil
	}
	return &domain.Club{Slug: slug}, nil
}

func TestServiceGetUnauthorized(t *testing.T) {
	t.Parallel()
	svc := preferences.NewService(&memSessions{enabled: true, baseURL: "http://localhost"}, &memPrefs{}, &memClubs{}, time.Now)
	_, err := svc.Get(context.Background(), "")
	if !errors.Is(err, preferences.ErrUnauthorized) {
		t.Fatalf("err = %v", err)
	}
}

func TestServicePutValidatesClubsAndPersists(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 2, 20, 0, 0, 0, time.UTC)
	user := &domain.User{ID: "usr_1", Email: "a@b.co", Role: domain.RoleUser}
	repo := &memPrefs{}
	svc := preferences.NewService(
		&memSessions{enabled: true, user: user, baseURL: "http://localhost"},
		repo,
		&memClubs{slugs: map[string]struct{}{"flamengo": {}, "vasco": {}}},
		func() time.Time { return now },
	)
	primary := "flamengo"
	view, err := svc.Put(context.Background(), "tok", preferences.Update{
		PrimaryClubSlug:   &primary,
		FavoriteClubSlugs: []string{"vasco", "flamengo", "vasco", " "},
	})
	if err != nil {
		t.Fatal(err)
	}
	if view.PrimaryClubSlug == nil || *view.PrimaryClubSlug != "flamengo" {
		t.Fatalf("primary = %v", view.PrimaryClubSlug)
	}
	if len(view.FavoriteClubSlugs) != 2 {
		t.Fatalf("favorites = %v", view.FavoriteClubSlugs)
	}
	unknown := "nao-existe"
	_, err = svc.Put(context.Background(), "tok", preferences.Update{
		PrimaryClubSlug:   &unknown,
		FavoriteClubSlugs: []string{},
	})
	if !errors.Is(err, preferences.ErrInvalidClub) {
		t.Fatalf("err = %v", err)
	}
}

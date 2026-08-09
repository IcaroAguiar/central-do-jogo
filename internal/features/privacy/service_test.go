package privacy_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/privacy"
)

type memSessions struct {
	enabled bool
	user    *domain.User
	baseURL string
	secure  bool
}

func (m *memSessions) Enabled() bool { return m.enabled }
func (m *memSessions) CurrentUser(context.Context, string) (*domain.User, error) {
	return m.user, nil
}
func (m *memSessions) PublicBaseURL() string { return m.baseURL }
func (m *memSessions) CookieSecure() bool    { return m.secure }

type memUsers struct {
	byID    map[domain.ID]*domain.User
	deleted []domain.ID
}

func (m *memUsers) GetByID(_ context.Context, id domain.ID) (*domain.User, error) {
	return m.byID[id], nil
}
func (m *memUsers) Delete(_ context.Context, id domain.ID) error {
	if _, ok := m.byID[id]; !ok {
		return errors.New("not found")
	}
	delete(m.byID, id)
	m.deleted = append(m.deleted, id)
	return nil
}

type memPrefs struct {
	row *domain.UserPreferences
}

func (m *memPrefs) GetByUserID(context.Context, domain.ID) (*domain.UserPreferences, error) {
	return m.row, nil
}

type memAnalytics struct {
	events []domain.AnalyticsEvent
}

func (m *memAnalytics) Insert(_ context.Context, event domain.AnalyticsEvent) error {
	m.events = append(m.events, event)
	return nil
}
func (m *memAnalytics) ListByUserID(_ context.Context, userID domain.ID, _ int) ([]domain.AnalyticsEvent, error) {
	var out []domain.AnalyticsEvent
	for _, ev := range m.events {
		if ev.UserID != nil && *ev.UserID == userID {
			out = append(out, ev)
		}
	}
	return out, nil
}
func (m *memAnalytics) DeleteBefore(_ context.Context, cutoff time.Time) (int64, error) {
	kept := m.events[:0]
	var n int64
	for _, ev := range m.events {
		if ev.CreatedAt.Before(cutoff) {
			n++
			continue
		}
		kept = append(kept, ev)
	}
	m.events = kept
	return n, nil
}

func TestExportOmitsSecretsAndIncludesPrefs(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 9, 12, 0, 0, 0, time.UTC)
	user := &domain.User{
		ID: "usr_1", Provider: "google", Email: "a@example.com", DisplayName: "A",
		Role: domain.RoleUser, CreatedAt: now.Add(-time.Hour), UpdatedAt: now,
	}
	primary := "flamengo"
	analytics := &memAnalytics{events: []domain.AnalyticsEvent{{
		ID: "aev_1", AnonymousID: "anon-secret-id", UserID: &user.ID,
		EventType: "page_view", Properties: map[string]any{"path": "/jogos/x"}, CreatedAt: now,
	}}}
	svc := privacy.NewService(
		&memSessions{enabled: true, user: user, baseURL: "http://127.0.0.1:8080"},
		&memUsers{byID: map[domain.ID]*domain.User{user.ID: user}},
		&memPrefs{row: &domain.UserPreferences{UserID: user.ID, PrimaryClubSlug: &primary, FavoriteClubSlugs: []string{"flamengo"}, UpdatedAt: now}},
		analytics,
		90,
		func() time.Time { return now },
	)

	exp, err := svc.ExportAccount(context.Background(), "tok")
	if err != nil {
		t.Fatal(err)
	}
	if exp.User.Email != "a@example.com" || exp.User.Provider != "google" {
		t.Fatalf("user = %+v", exp.User)
	}
	if exp.Preferences.PrimaryClubSlug == nil || *exp.Preferences.PrimaryClubSlug != "flamengo" {
		t.Fatalf("prefs = %+v", exp.Preferences)
	}
	if len(exp.AnalyticsEvents) != 1 || exp.AnalyticsEvents[0].EventType != "page_view" {
		t.Fatalf("events = %+v", exp.AnalyticsEvents)
	}
	// Export must not surface the anonymous local id (REQ-019/020).
	raw := exp.AnalyticsEvents[0]
	if raw.ID == "" || raw.Properties["path"] != "/jogos/x" {
		t.Fatalf("event = %+v", raw)
	}
}

func TestDeleteAccount(t *testing.T) {
	t.Parallel()
	user := &domain.User{ID: "usr_1", Role: domain.RoleUser}
	users := &memUsers{byID: map[domain.ID]*domain.User{user.ID: user}}
	svc := privacy.NewService(
		&memSessions{enabled: true, user: user},
		users,
		&memPrefs{},
		&memAnalytics{},
		90,
		time.Now,
	)
	if err := svc.DeleteAccount(context.Background(), "tok"); err != nil {
		t.Fatal(err)
	}
	if len(users.deleted) != 1 || users.deleted[0] != user.ID {
		t.Fatalf("deleted = %#v", users.deleted)
	}
}

func TestRecordEventLinksOnlyWithConsent(t *testing.T) {
	t.Parallel()
	user := &domain.User{ID: "usr_1", Role: domain.RoleUser}
	analytics := &memAnalytics{}
	svc := privacy.NewService(
		&memSessions{enabled: true, user: user},
		&memUsers{byID: map[domain.ID]*domain.User{user.ID: user}},
		&memPrefs{},
		analytics,
		90,
		time.Now,
	)

	if err := svc.RecordEvent(context.Background(), "tok", privacy.AnalyticsInput{
		AnonymousID: "anon-12345678", EventType: "view", ConsentToLink: false,
	}); err != nil {
		t.Fatal(err)
	}
	if analytics.events[0].UserID != nil {
		t.Fatal("expected no user link without consent")
	}

	if err := svc.RecordEvent(context.Background(), "tok", privacy.AnalyticsInput{
		AnonymousID: "anon-12345678", EventType: "view", ConsentToLink: true,
	}); err != nil {
		t.Fatal(err)
	}
	if analytics.events[1].UserID == nil || *analytics.events[1].UserID != user.ID {
		t.Fatalf("expected linked user, got %+v", analytics.events[1])
	}
}

func TestPurgeExpired(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 9, 0, 0, 0, 0, time.UTC)
	analytics := &memAnalytics{events: []domain.AnalyticsEvent{
		{ID: "old", AnonymousID: "anon-12345678", EventType: "x", CreatedAt: now.Add(-100 * 24 * time.Hour)},
		{ID: "new", AnonymousID: "anon-12345678", EventType: "y", CreatedAt: now.Add(-10 * 24 * time.Hour)},
	}}
	svc := privacy.NewService(
		&memSessions{enabled: true},
		&memUsers{},
		&memPrefs{},
		analytics,
		90,
		func() time.Time { return now },
	)
	n, err := svc.PurgeExpired(context.Background())
	if err != nil || n != 1 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	if len(analytics.events) != 1 || analytics.events[0].ID != "new" {
		t.Fatalf("events = %+v", analytics.events)
	}
}

func TestRecordEventRejectsOversizedProperties(t *testing.T) {
	t.Parallel()
	svc := privacy.NewService(
		&memSessions{enabled: false},
		&memUsers{},
		&memPrefs{},
		&memAnalytics{},
		90,
		time.Now,
	)
	props := map[string]any{}
	for i := 0; i < 21; i++ {
		props[fmt.Sprintf("k%d", i)] = "x"
	}
	err := svc.RecordEvent(context.Background(), "", privacy.AnalyticsInput{
		AnonymousID: "anon-12345678", EventType: "view", Properties: props,
	})
	if err != privacy.ErrInvalidEvent {
		t.Fatalf("err = %v, want ErrInvalidEvent", err)
	}

	deep := map[string]any{"a": map[string]any{"b": map[string]any{"c": map[string]any{"d": "too deep"}}}}
	err = svc.RecordEvent(context.Background(), "", privacy.AnalyticsInput{
		AnonymousID: "anon-12345678", EventType: "view", Properties: deep,
	})
	if err != privacy.ErrInvalidEvent {
		t.Fatalf("depth err = %v, want ErrInvalidEvent", err)
	}
}

func TestCookieSecureDelegatesToSessions(t *testing.T) {
	t.Parallel()
	svc := privacy.NewService(&memSessions{secure: true}, &memUsers{}, &memPrefs{}, &memAnalytics{}, 90, time.Now)
	if !svc.CookieSecure() {
		t.Fatal("expected CookieSecure true")
	}
}

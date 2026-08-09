package auth_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/auth"
)

type memUser struct {
	users    map[string]*domain.User // provider|subject
	byID     map[domain.ID]*domain.User
	sessions map[string]*domain.Session // token hash
}

func newMem() *memUser {
	return &memUser{
		users:    map[string]*domain.User{},
		byID:     map[domain.ID]*domain.User{},
		sessions: map[string]*domain.Session{},
	}
}

func (m *memUser) GetByID(_ context.Context, id domain.ID) (*domain.User, error) {
	u := m.byID[id]
	if u == nil {
		return nil, nil
	}
	cp := *u
	return &cp, nil
}

func (m *memUser) UpsertByProviderSubject(_ context.Context, user domain.User, now time.Time) (*domain.User, error) {
	key := user.Provider + "|" + user.ProviderSubject
	if existing, ok := m.users[key]; ok {
		existing.Email = user.Email
		existing.DisplayName = user.DisplayName
		existing.Role = user.Role
		existing.UpdatedAt = now
		existing.LastLoginAt = &now
		cp := *existing
		return &cp, nil
	}
	user.CreatedAt = now
	user.UpdatedAt = now
	user.LastLoginAt = &now
	cp := user
	m.users[key] = &cp
	m.byID[user.ID] = &cp
	out := cp
	return &out, nil
}

func (m *memUser) CreateSession(_ context.Context, sess domain.Session) error {
	cp := sess
	m.sessions[sess.TokenHash] = &cp
	return nil
}

func (m *memUser) GetSessionByTokenHash(_ context.Context, tokenHash string, now time.Time) (*domain.Session, error) {
	s := m.sessions[tokenHash]
	if s == nil || s.RevokedAt != nil || !s.ExpiresAt.After(now) {
		return nil, nil
	}
	cp := *s
	return &cp, nil
}

func (m *memUser) RevokeSession(_ context.Context, tokenHash string, now time.Time) error {
	if s := m.sessions[tokenHash]; s != nil && s.RevokedAt == nil {
		s.RevokedAt = &now
	}
	return nil
}

type fakeProvider struct {
	identity auth.Identity
	err      error
}

func (f *fakeProvider) Name() string { return auth.ProviderGoogle }

func (f *fakeProvider) AuthCodeURL(state string) string {
	return "https://accounts.example/auth?state=" + state
}

func (f *fakeProvider) Exchange(_ context.Context, _ string) (auth.Identity, error) {
	if f.err != nil {
		return auth.Identity{}, f.err
	}
	return f.identity, nil
}

func testCfg(emails ...string) auth.Config {
	allow := map[string]struct{}{}
	for _, e := range emails {
		allow[strings.ToLower(e)] = struct{}{}
	}
	return auth.Config{
		Enabled:          true,
		SessionSecret:    []byte("0123456789abcdef0123456789abcdef"),
		SessionTTL:       24 * time.Hour,
		CookieSecure:     false,
		PublicBaseURL:    "http://127.0.0.1:8080",
		MaintainerEmails: allow,
	}
}

func stateFromAuthURL(authURL string) string {
	return authURL[strings.LastIndex(authURL, "state=")+6:]
}

func TestFirstLoginNeverPromotesWithoutAllowlist(t *testing.T) {
	t.Parallel()
	mem := newMem()
	svc := auth.NewService(mem, &fakeProvider{identity: auth.Identity{
		Provider: auth.ProviderGoogle, Subject: "sub-1", Email: "fan@example.com", EmailVerified: true, DisplayName: "Fan",
	}}, testCfg(), func() time.Time { return time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC) })

	start, err := svc.StartLogin()
	if err != nil {
		t.Fatalf("StartLogin: %v", err)
	}
	user, err := svc.CompleteLogin(context.Background(), "code", stateFromAuthURL(start.AuthURL), start.SignedStateCookie)
	if err != nil {
		t.Fatalf("CompleteLogin: %v", err)
	}
	if user.User.Role != domain.RoleUser {
		t.Fatalf("role = %q, want user", user.User.Role)
	}
}

func TestAllowlistGrantsMaintainerOnLogin(t *testing.T) {
	t.Parallel()
	mem := newMem()
	now := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	svc := auth.NewService(mem, &fakeProvider{identity: auth.Identity{
		Provider: auth.ProviderGoogle, Subject: "sub-2", Email: "Owner@Example.com", EmailVerified: true,
	}}, testCfg("owner@example.com"), func() time.Time { return now })

	start, err := svc.StartLogin()
	if err != nil {
		t.Fatalf("StartLogin: %v", err)
	}
	result, err := svc.CompleteLogin(context.Background(), "code", stateFromAuthURL(start.AuthURL), start.SignedStateCookie)
	if err != nil {
		t.Fatalf("CompleteLogin: %v", err)
	}
	if result.User.Role != domain.RoleMaintainer {
		t.Fatalf("role = %q, want maintainer", result.User.Role)
	}
	if result.User.Email != "owner@example.com" {
		t.Fatalf("email = %q", result.User.Email)
	}
	if result.SessionToken == "" {
		t.Fatal("missing session token")
	}
	sum := sha256.Sum256([]byte(result.SessionToken))
	hash := hex.EncodeToString(sum[:])
	if _, ok := mem.sessions[hash]; !ok {
		t.Fatal("session hash not stored")
	}

	got, err := svc.CurrentUser(context.Background(), result.SessionToken)
	if err != nil || got == nil || got.Role != domain.RoleMaintainer {
		t.Fatalf("CurrentUser = %+v err=%v", got, err)
	}
}

func TestInvalidStateRejected(t *testing.T) {
	t.Parallel()
	svc := auth.NewService(newMem(), &fakeProvider{identity: auth.Identity{
		Provider: auth.ProviderGoogle, Subject: "x", Email: "a@b.co", EmailVerified: true,
	}}, testCfg(), time.Now)

	start, _ := svc.StartLogin()
	_, err := svc.CompleteLogin(context.Background(), "code", "wrong-state", start.SignedStateCookie)
	if err != auth.ErrInvalidState {
		t.Fatalf("err = %v, want ErrInvalidState", err)
	}
}

func TestUnverifiedEmailRejected(t *testing.T) {
	t.Parallel()
	svc := auth.NewService(newMem(), &fakeProvider{identity: auth.Identity{
		Provider: auth.ProviderGoogle, Subject: "x", Email: "a@b.co", EmailVerified: false,
	}}, testCfg(), time.Now)

	start, _ := svc.StartLogin()
	_, err := svc.CompleteLogin(context.Background(), "code", stateFromAuthURL(start.AuthURL), start.SignedStateCookie)
	if err != auth.ErrEmailUnverified {
		t.Fatalf("err = %v, want ErrEmailUnverified", err)
	}
}

func TestAllowlistDemotesActiveSession(t *testing.T) {
	t.Parallel()
	mem := newMem()
	now := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	cfg := testCfg("owner@example.com")
	svc := auth.NewService(mem, &fakeProvider{identity: auth.Identity{
		Provider: auth.ProviderGoogle, Subject: "sub-demote", Email: "owner@example.com", EmailVerified: true,
	}}, cfg, func() time.Time { return now })

	start, _ := svc.StartLogin()
	result, err := svc.CompleteLogin(context.Background(), "code", stateFromAuthURL(start.AuthURL), start.SignedStateCookie)
	if err != nil {
		t.Fatal(err)
	}
	delete(cfg.MaintainerEmails, "owner@example.com")
	got, err := svc.CurrentUser(context.Background(), result.SessionToken)
	if err != nil || got == nil {
		t.Fatalf("CurrentUser = %+v err=%v", got, err)
	}
	if got.Role != domain.RoleUser {
		t.Fatalf("role after allowlist removal = %q, want user", got.Role)
	}
}

func TestLogoutRevokesSession(t *testing.T) {
	t.Parallel()
	mem := newMem()
	now := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	svc := auth.NewService(mem, &fakeProvider{identity: auth.Identity{
		Provider: auth.ProviderGoogle, Subject: "sub-3", Email: "u@example.com", EmailVerified: true,
	}}, testCfg(), func() time.Time { return now })

	start, _ := svc.StartLogin()
	result, err := svc.CompleteLogin(context.Background(), "code", stateFromAuthURL(start.AuthURL), start.SignedStateCookie)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Logout(context.Background(), result.SessionToken); err != nil {
		t.Fatal(err)
	}
	got, err := svc.CurrentUser(context.Background(), result.SessionToken)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("expected anonymous after logout, got %+v", got)
	}
}

func TestSafeRelativePath(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"":             "/",
		"/":            "/",
		"/clubes/foo":  "/clubes/foo",
		"//evil":       "/",
		"https://evil": "/",
		"/ok?x=1":      "/ok?x=1",
		"relative":     "/",
		"/a\nb":        "/",
	}
	for in, want := range cases {
		if got := auth.SafeRelativePath(in); got != want {
			t.Fatalf("SafeRelativePath(%q)=%q want %q", in, got, want)
		}
	}
}

func TestHandlersMeAnonymousWhenDisabled(t *testing.T) {
	t.Parallel()
	svc := auth.NewService(newMem(), nil, auth.Config{Enabled: false}, time.Now)
	h := auth.NewHandlers(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	w := httptest.NewRecorder()
	h.Me().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"authenticated":false`) {
		t.Fatalf("body = %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"authEnabled":false`) {
		t.Fatalf("body = %s", w.Body.String())
	}
}

func TestLogoutRejectsMissingOriginWhenBaseConfigured(t *testing.T) {
	t.Parallel()
	svc := auth.NewService(newMem(), &fakeProvider{identity: auth.Identity{
		Provider: auth.ProviderGoogle, Subject: "x", Email: "a@b.co", EmailVerified: true,
	}}, testCfg(), time.Now)
	h := auth.NewHandlers(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	w := httptest.NewRecorder()
	h.Logout().ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", w.Code)
	}
}

func TestServiceRequireMaintainer(t *testing.T) {
	t.Parallel()
	mem := newMem()
	now := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	svc := auth.NewService(mem, &fakeProvider{identity: auth.Identity{
		Provider: auth.ProviderGoogle, Subject: "sub-gate", Email: "owner@example.com", EmailVerified: true,
	}}, testCfg("owner@example.com"), func() time.Time { return now })

	start, _ := svc.StartLogin()
	result, err := svc.CompleteLogin(context.Background(), "code", stateFromAuthURL(start.AuthURL), start.SignedStateCookie)
	if err != nil {
		t.Fatal(err)
	}
	user, err := svc.RequireMaintainer(context.Background(), result.SessionToken)
	if err != nil || user == nil || user.Role != domain.RoleMaintainer {
		t.Fatalf("user=%+v err=%v", user, err)
	}

	userSvc := auth.NewService(mem, &fakeProvider{identity: auth.Identity{
		Provider: auth.ProviderGoogle, Subject: "sub-user", Email: "user@example.com", EmailVerified: true,
	}}, testCfg("owner@example.com"), func() time.Time { return now })
	start2, _ := userSvc.StartLogin()
	result2, err := userSvc.CompleteLogin(context.Background(), "code", stateFromAuthURL(start2.AuthURL), start2.SignedStateCookie)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := userSvc.RequireMaintainer(context.Background(), result2.SessionToken); err != auth.ErrForbidden {
		t.Fatalf("err = %v, want forbidden", err)
	}
}

func TestLogoutRejectsForeignOrigin(t *testing.T) {
	t.Parallel()
	svc := auth.NewService(newMem(), &fakeProvider{identity: auth.Identity{
		Provider: auth.ProviderGoogle, Subject: "x", Email: "a@b.co", EmailVerified: true,
	}}, testCfg(), time.Now)
	h := auth.NewHandlers(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Origin", "https://evil.example")
	w := httptest.NewRecorder()
	h.Logout().ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", w.Code)
	}
}

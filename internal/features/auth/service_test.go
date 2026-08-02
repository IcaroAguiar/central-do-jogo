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

func TestFirstLoginNeverPromotesWithoutAllowlist(t *testing.T) {
	t.Parallel()
	mem := newMem()
	svc := auth.NewService(mem, &fakeProvider{identity: auth.Identity{
		Provider: auth.ProviderGoogle, Subject: "sub-1", Email: "fan@example.com", EmailVerified: true, DisplayName: "Fan",
	}}, testCfg(), func() time.Time { return time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC) })

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/callback?code=x&state=abc", nil)
	startURL, err := svc.StartURL(w)
	if err != nil {
		t.Fatalf("StartURL: %v", err)
	}
	state := startURL[strings.LastIndex(startURL, "state=")+6:]
	r.AddCookie(w.Result().Cookies()[0])

	user, err := svc.CompleteLogin(context.Background(), w, r, "code", state)
	if err != nil {
		t.Fatalf("CompleteLogin: %v", err)
	}
	if user.Role != domain.RoleUser {
		t.Fatalf("role = %q, want user", user.Role)
	}
}

func TestAllowlistGrantsMaintainerOnLogin(t *testing.T) {
	t.Parallel()
	mem := newMem()
	now := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	svc := auth.NewService(mem, &fakeProvider{identity: auth.Identity{
		Provider: auth.ProviderGoogle, Subject: "sub-2", Email: "Owner@Example.com", EmailVerified: true,
	}}, testCfg("owner@example.com"), func() time.Time { return now })

	sw := httptest.NewRecorder()
	startURL, err := svc.StartURL(sw)
	if err != nil {
		t.Fatalf("StartURL: %v", err)
	}
	state := startURL[strings.LastIndex(startURL, "state=")+6:]
	rw := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(sw.Result().Cookies()[0])

	user, err := svc.CompleteLogin(context.Background(), rw, req, "code", state)
	if err != nil {
		t.Fatalf("CompleteLogin: %v", err)
	}
	if user.Role != domain.RoleMaintainer {
		t.Fatalf("role = %q, want maintainer", user.Role)
	}
	if user.Email != "owner@example.com" {
		t.Fatalf("email = %q", user.Email)
	}

	cookies := rw.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "cdj_session" {
			sessionCookie = c
		}
	}
	if sessionCookie == nil || sessionCookie.Value == "" {
		t.Fatal("missing session cookie")
	}
	sum := sha256.Sum256([]byte(sessionCookie.Value))
	hash := hex.EncodeToString(sum[:])
	if _, ok := mem.sessions[hash]; !ok {
		t.Fatal("session hash not stored")
	}

	meReq := httptest.NewRequest(http.MethodGet, "/me", nil)
	meReq.AddCookie(sessionCookie)
	got, err := svc.CurrentUser(context.Background(), meReq)
	if err != nil || got == nil || got.Role != domain.RoleMaintainer {
		t.Fatalf("CurrentUser = %+v err=%v", got, err)
	}
}

func TestInvalidStateRejected(t *testing.T) {
	t.Parallel()
	svc := auth.NewService(newMem(), &fakeProvider{identity: auth.Identity{
		Provider: auth.ProviderGoogle, Subject: "x", Email: "a@b.co", EmailVerified: true,
	}}, testCfg(), time.Now)

	w := httptest.NewRecorder()
	_, _ = svc.StartURL(w)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(w.Result().Cookies()[0])
	_, err := svc.CompleteLogin(context.Background(), httptest.NewRecorder(), req, "code", "wrong-state")
	if err != auth.ErrInvalidState {
		t.Fatalf("err = %v, want ErrInvalidState", err)
	}
}

func TestUnverifiedEmailRejected(t *testing.T) {
	t.Parallel()
	svc := auth.NewService(newMem(), &fakeProvider{identity: auth.Identity{
		Provider: auth.ProviderGoogle, Subject: "x", Email: "a@b.co", EmailVerified: false,
	}}, testCfg(), time.Now)

	sw := httptest.NewRecorder()
	startURL, _ := svc.StartURL(sw)
	state := startURL[strings.LastIndex(startURL, "state=")+6:]
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(sw.Result().Cookies()[0])
	_, err := svc.CompleteLogin(context.Background(), httptest.NewRecorder(), req, "code", state)
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

	sw := httptest.NewRecorder()
	startURL, _ := svc.StartURL(sw)
	state := startURL[strings.LastIndex(startURL, "state=")+6:]
	rw := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(sw.Result().Cookies()[0])
	if _, err := svc.CompleteLogin(context.Background(), rw, req, "code", state); err != nil {
		t.Fatal(err)
	}
	var sessionCookie *http.Cookie
	for _, c := range rw.Result().Cookies() {
		if c.Name == "cdj_session" {
			sessionCookie = c
		}
	}
	delete(cfg.MaintainerEmails, "owner@example.com")
	meReq := httptest.NewRequest(http.MethodGet, "/me", nil)
	meReq.AddCookie(sessionCookie)
	got, err := svc.CurrentUser(context.Background(), meReq)
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

	sw := httptest.NewRecorder()
	startURL, _ := svc.StartURL(sw)
	state := startURL[strings.LastIndex(startURL, "state=")+6:]
	rw := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(sw.Result().Cookies()[0])
	_, err := svc.CompleteLogin(context.Background(), rw, req, "code", state)
	if err != nil {
		t.Fatal(err)
	}
	var sessionCookie *http.Cookie
	for _, c := range rw.Result().Cookies() {
		if c.Name == "cdj_session" {
			sessionCookie = c
		}
	}
	logoutW := httptest.NewRecorder()
	logoutR := httptest.NewRequest(http.MethodPost, "/logout", nil)
	logoutR.AddCookie(sessionCookie)
	if err := svc.Logout(context.Background(), logoutW, logoutR); err != nil {
		t.Fatal(err)
	}
	meReq := httptest.NewRequest(http.MethodGet, "/me", nil)
	meReq.AddCookie(sessionCookie)
	got, err := svc.CurrentUser(context.Background(), meReq)
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

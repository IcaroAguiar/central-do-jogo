package preferences_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/auth"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/preferences"
)

func TestPutRejectsMissingOrigin(t *testing.T) {
	t.Parallel()
	user := &domain.User{ID: "usr_1", Role: domain.RoleUser}
	svc := preferences.NewService(
		&memSessions{enabled: true, user: user, baseURL: "http://127.0.0.1:8080"},
		&memPrefs{},
		&memClubs{slugs: map[string]struct{}{"flamengo": {}}},
		time.Now,
	)
	h := preferences.NewHandlers(svc)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/preferences", strings.NewReader(`{"favoriteClubSlugs":[]}`))
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "tok"})
	w := httptest.NewRecorder()
	h.Put().ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestGetRequiresAuth(t *testing.T) {
	t.Parallel()
	svc := preferences.NewService(
		&memSessions{enabled: true, baseURL: "http://127.0.0.1:8080"},
		&memPrefs{},
		&memClubs{},
		time.Now,
	)
	h := preferences.NewHandlers(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/preferences", nil)
	w := httptest.NewRecorder()
	h.Get().ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestPutHappyPath(t *testing.T) {
	t.Parallel()
	user := &domain.User{ID: "usr_1", Role: domain.RoleUser}
	svc := preferences.NewService(
		&memSessions{enabled: true, user: user, baseURL: "http://127.0.0.1:8080"},
		&memPrefs{},
		&memClubs{slugs: map[string]struct{}{"flamengo": {}, "vasco": {}}},
		func() time.Time { return time.Date(2026, 8, 2, 21, 0, 0, 0, time.UTC) },
	)
	h := preferences.NewHandlers(svc)
	body := `{"primaryClubSlug":"flamengo","favoriteClubSlugs":["vasco"]}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/preferences", strings.NewReader(body))
	req.Header.Set("Origin", "http://127.0.0.1:8080")
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "tok"})
	w := httptest.NewRecorder()
	h.Put().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", w.Code, w.Body.String())
	}
	var resp preferences.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.PrimaryClubSlug == nil || *resp.PrimaryClubSlug != "flamengo" {
		t.Fatalf("resp = %+v", resp)
	}
}

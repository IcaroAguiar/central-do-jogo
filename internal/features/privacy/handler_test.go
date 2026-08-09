package privacy_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/privacy"
	httpplatform "github.com/IcaroAguiar/central-do-jogo/internal/platform/http"
)

func TestDeleteAccountClearsSecureCookie(t *testing.T) {
	t.Parallel()
	user := &domain.User{ID: "usr_1", Role: domain.RoleUser}
	users := &memUsers{byID: map[domain.ID]*domain.User{user.ID: user}}
	svc := privacy.NewService(
		&memSessions{enabled: true, user: user, baseURL: "http://127.0.0.1:8080", secure: true},
		users,
		&memPrefs{},
		&memAnalytics{},
		90,
		time.Now,
	)
	h := privacy.NewHandlers(svc)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/privacy/account", nil)
	req.Header.Set("Origin", "http://127.0.0.1:8080")
	req.AddCookie(&http.Cookie{Name: httpplatform.SessionCookieName, Value: "tok"})
	w := httptest.NewRecorder()
	h.DeleteAccount().ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d body=%s", w.Code, w.Body.String())
	}
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected cleared session cookie")
	}
	var cleared *http.Cookie
	for _, c := range cookies {
		if c.Name == httpplatform.SessionCookieName {
			cleared = c
			break
		}
	}
	if cleared == nil || !cleared.Secure {
		t.Fatalf("cookie = %#v, want Secure=true", cleared)
	}
}

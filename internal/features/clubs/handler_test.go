package clubs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
)

func withSlug(req *http.Request, slug string) *http.Request {
	req.SetPathValue("slug", slug)
	return req
}

func TestDetailHandlerNotFound(t *testing.T) {
	t.Parallel()
	svc := NewService(&fakeClubReader{bySlug: map[string]*domain.Club{}}, &fakeMatchLister{}, nil)
	handler := NewDetailHandler(svc)

	rec := httptest.NewRecorder()
	req := withSlug(httptest.NewRequest(http.MethodGet, "/api/v1/clubs/missing", nil), "missing")
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"club_not_found"`) {
		t.Fatalf("body = %q, want club_not_found error code", rec.Body.String())
	}
}

func TestDetailHandlerOK(t *testing.T) {
	t.Parallel()
	club := &domain.Club{Slug: "flamengo", Name: "Flamengo"}
	svc := NewService(&fakeClubReader{bySlug: map[string]*domain.Club{"flamengo": club}}, &fakeMatchLister{}, nil)
	handler := NewDetailHandler(svc)

	rec := httptest.NewRecorder()
	req := withSlug(httptest.NewRequest(http.MethodGet, "/api/v1/clubs/flamengo", nil), "flamengo")
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"Flamengo"`) {
		t.Fatalf("body = %q, want club name", rec.Body.String())
	}
}

func TestMatchesHandlerRejectsInvalidRange(t *testing.T) {
	t.Parallel()
	svc := NewService(&fakeClubReader{}, &fakeMatchLister{}, nil)
	handler := NewMatchesHandler(svc)

	rec := httptest.NewRecorder()
	req := withSlug(httptest.NewRequest(http.MethodGet, "/api/v1/clubs/flamengo/matches?range=decade", nil), "flamengo")
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"invalid_range"`) {
		t.Fatalf("body = %q, want invalid_range error code", rec.Body.String())
	}
}

func TestMatchesHandlerRejectsInvalidSeason(t *testing.T) {
	t.Parallel()
	svc := NewService(&fakeClubReader{}, &fakeMatchLister{}, nil)
	handler := NewMatchesHandler(svc)

	rec := httptest.NewRecorder()
	req := withSlug(httptest.NewRequest(http.MethodGet, "/api/v1/clubs/flamengo/matches?season=abc", nil), "flamengo")
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"invalid_season"`) {
		t.Fatalf("body = %q, want invalid_season error code", rec.Body.String())
	}
}

func TestMatchesHandlerNotFound(t *testing.T) {
	t.Parallel()
	svc := NewService(&fakeClubReader{bySlug: map[string]*domain.Club{}}, &fakeMatchLister{}, nil)
	handler := NewMatchesHandler(svc)

	rec := httptest.NewRecorder()
	req := withSlug(httptest.NewRequest(http.MethodGet, "/api/v1/clubs/missing/matches", nil), "missing")
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestMatchesHandlerOKWithDefaults(t *testing.T) {
	t.Parallel()
	club := &domain.Club{ID: "club_flamengo", Slug: "flamengo", Name: "Flamengo"}
	svc := NewService(&fakeClubReader{bySlug: map[string]*domain.Club{"flamengo": club}}, &fakeMatchLister{}, nil)
	handler := NewMatchesHandler(svc)

	rec := httptest.NewRecorder()
	req := withSlug(httptest.NewRequest(http.MethodGet, "/api/v1/clubs/flamengo/matches", nil), "flamengo")
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"range":"week"`) {
		t.Fatalf("body = %q, want default range week", rec.Body.String())
	}
}

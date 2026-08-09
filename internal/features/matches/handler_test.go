package matches

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
	svc := NewService(&fakeMatchGetter{bySlug: map[string]*domain.MatchRecord{}}, &fakeBroadcastLister{}, &fakeLineupLister{}, &fakeNewsLister{})
	handler := NewDetailHandler(svc)

	rec := httptest.NewRecorder()
	req := withSlug(httptest.NewRequest(http.MethodGet, "/api/v1/matches/missing", nil), "missing")
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"match_not_found"`) {
		t.Fatalf("body = %q, want match_not_found error code", rec.Body.String())
	}
}

func TestDetailHandlerOK(t *testing.T) {
	t.Parallel()
	rec := &domain.MatchRecord{Match: domain.Match{ID: "match_1", Slug: "flamengo-x-vasco"}}
	svc := NewService(&fakeMatchGetter{bySlug: map[string]*domain.MatchRecord{"flamengo-x-vasco": rec}}, &fakeBroadcastLister{}, &fakeLineupLister{}, &fakeNewsLister{})
	handler := NewDetailHandler(svc)

	recorder := httptest.NewRecorder()
	req := withSlug(httptest.NewRequest(http.MethodGet, "/api/v1/matches/flamengo-x-vasco", nil), "flamengo-x-vasco")
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"flamengo-x-vasco"`) {
		t.Fatalf("body = %q, want match slug", recorder.Body.String())
	}
}

func TestDetailHandlerReturns500OnServiceError(t *testing.T) {
	t.Parallel()
	svc := NewService(&fakeMatchGetter{err: errBoom}, &fakeBroadcastLister{}, &fakeLineupLister{}, &fakeNewsLister{})
	handler := NewDetailHandler(svc)

	recorder := httptest.NewRecorder()
	req := withSlug(httptest.NewRequest(http.MethodGet, "/api/v1/matches/x", nil), "x")
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", recorder.Code)
	}
}

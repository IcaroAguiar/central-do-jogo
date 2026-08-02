package search

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
)

func TestHandlerRequiresQuery(t *testing.T) {
	t.Parallel()
	svc := NewService(&fakeClubSearcher{}, &fakeMatchSearcher{})
	handler := NewHandler(svc)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/search", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if got := rec.Body.String(); !containsAll(got, `"error"`, `"code":"invalid_query"`) {
		t.Fatalf("body = %q, want error envelope with invalid_query code", got)
	}
}

func TestHandlerTrimsWhitespaceOnlyQuery(t *testing.T) {
	t.Parallel()
	svc := NewService(&fakeClubSearcher{}, &fakeMatchSearcher{})
	handler := NewHandler(svc)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/search?q=%20%20%20", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 for whitespace-only query", rec.Code)
	}
}

func TestHandlerReturnsResultsOnOK(t *testing.T) {
	t.Parallel()
	svc := NewService(&fakeClubSearcher{clubs: []domain.Club{{Slug: "flamengo", Name: "Flamengo"}}}, &fakeMatchSearcher{})
	handler := NewHandler(svc)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/search?q=fla", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}
	if got := rec.Body.String(); !containsAll(got, `"flamengo"`) {
		t.Fatalf("body = %q, want flamengo hit", got)
	}
}

func TestHandlerReturns500OnServiceError(t *testing.T) {
	t.Parallel()
	svc := NewService(&fakeClubSearcher{err: errBoom}, &fakeMatchSearcher{})
	handler := NewHandler(svc)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/search?q=x", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}

func containsAll(haystack string, needles ...string) bool {
	for _, n := range needles {
		if !strings.Contains(haystack, n) {
			return false
		}
	}
	return true
}

var errBoom = errors.New("boom")

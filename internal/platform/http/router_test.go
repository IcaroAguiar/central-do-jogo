package httpplatform

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHealthz(t *testing.T) {
	fixed := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
	router := NewRouter(Options{
		Now: func() time.Time { return fixed },
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status field = %v, want ok", body["status"])
	}
	if body["checkedAt"] != "2026-07-31T12:00:00Z" {
		t.Fatalf("checkedAt = %v, want 2026-07-31T12:00:00Z", body["checkedAt"])
	}
}

func TestSPAFallbackAndAssets(t *testing.T) {
	dir := t.TempDir()
	indexHTML := "<!doctype html><title>shell</title>"
	assetJS := "console.log('ok')"
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(indexHTML), 0o644); err != nil {
		t.Fatalf("write index.html: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "app.js"), []byte(assetJS), 0o644); err != nil {
		t.Fatalf("write app.js: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatalf("mkdir assets: %v", err)
	}

	router := NewRouter(Options{StaticDir: dir})

	t.Run("root serves index", func(t *testing.T) {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "shell") {
			t.Fatalf("body = %q, want index html", rec.Body.String())
		}
	})

	t.Run("missing client route falls back to index", func(t *testing.T) {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/jogos/1", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "shell") {
			t.Fatalf("body = %q, want index html fallback", rec.Body.String())
		}
	})

	t.Run("path colliding with host root still falls back", func(t *testing.T) {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/tmp", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "shell") {
			t.Fatalf("body = %q, want index html fallback for /tmp", rec.Body.String())
		}
	})

	t.Run("existing asset is served", func(t *testing.T) {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/app.js", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if rec.Body.String() != assetJS {
			t.Fatalf("body = %q, want %q", rec.Body.String(), assetJS)
		}
	})
}

package httpplatform_test

import (
	"net/http"
	"testing"

	httpplatform "github.com/IcaroAguiar/central-do-jogo/internal/platform/http"
)

func TestOriginAllowed(t *testing.T) {
	t.Parallel()
	base := "http://127.0.0.1:8080"

	t.Run("empty base fails closed", func(t *testing.T) {
		t.Parallel()
		req, _ := http.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Origin", base)
		if httpplatform.OriginAllowed(req, "") {
			t.Fatal("expected false")
		}
	})

	t.Run("matching origin", func(t *testing.T) {
		t.Parallel()
		req, _ := http.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Origin", base)
		if !httpplatform.OriginAllowed(req, base) {
			t.Fatal("expected true")
		}
	})

	t.Run("matching referer when origin missing", func(t *testing.T) {
		t.Parallel()
		req, _ := http.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Referer", base+"/clubes/flamengo")
		if !httpplatform.OriginAllowed(req, base) {
			t.Fatal("expected true")
		}
	})

	t.Run("foreign origin rejected", func(t *testing.T) {
		t.Parallel()
		req, _ := http.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Origin", "https://evil.example")
		if httpplatform.OriginAllowed(req, base) {
			t.Fatal("expected false")
		}
	})
}

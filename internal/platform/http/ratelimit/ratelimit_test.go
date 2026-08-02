package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAllowBurstThenBlocks(t *testing.T) {
	t.Parallel()
	clock := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	l := New(1, 3)
	l.now = func() time.Time { return clock }

	for i := 0; i < 3; i++ {
		if !l.Allow("ip-a") {
			t.Fatalf("request %d should be allowed within burst", i)
		}
	}
	if l.Allow("ip-a") {
		t.Fatal("4th immediate request should be blocked")
	}
}

func TestAllowRefillsOverTime(t *testing.T) {
	t.Parallel()
	clock := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	l := New(1, 1) // 1 token/sec, burst 1
	l.now = func() time.Time { return clock }

	if !l.Allow("ip-b") {
		t.Fatal("first request should be allowed")
	}
	if l.Allow("ip-b") {
		t.Fatal("second immediate request should be blocked")
	}

	clock = clock.Add(1100 * time.Millisecond)
	if !l.Allow("ip-b") {
		t.Fatal("request after refill window should be allowed")
	}
}

func TestAllowKeysAreIndependent(t *testing.T) {
	t.Parallel()
	l := New(1, 1)

	if !l.Allow("ip-1") {
		t.Fatal("ip-1 first request should be allowed")
	}
	if !l.Allow("ip-2") {
		t.Fatal("ip-2 should have an independent bucket")
	}
	if l.Allow("ip-1") {
		t.Fatal("ip-1 second immediate request should be blocked")
	}
}

func TestMiddlewareReturns429WithErrorBody(t *testing.T) {
	t.Parallel()
	l := New(0.0001, 1)
	handlerCalls := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalls++
		w.WriteHeader(http.StatusOK)
	})

	onLimited := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"code":"rate_limited","message":"too many requests"}}`))
	}
	wrapped := Middleware(l, "GET /api/v1/search", onLimited)(next)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=a", nil)
	req.RemoteAddr = "203.0.113.9:12345"

	rec1 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec1, req)
	if rec1.Code != http.StatusOK {
		t.Fatalf("first request status = %d, want 200", rec1.Code)
	}

	rec2 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request status = %d, want 429", rec2.Code)
	}
	if handlerCalls != 1 {
		t.Fatalf("handler called %d times, want 1 (second request should be blocked before reaching it)", handlerCalls)
	}
}

func TestMiddlewareKeysByIPAndRoute(t *testing.T) {
	t.Parallel()
	l := New(0.0001, 1)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	wrapped := Middleware(l, "GET /api/v1/search", nil)(next)

	reqA := httptest.NewRequest(http.MethodGet, "/api/v1/search", nil)
	reqA.RemoteAddr = "198.51.100.1:1"
	reqB := httptest.NewRequest(http.MethodGet, "/api/v1/search", nil)
	reqB.RemoteAddr = "198.51.100.2:2"

	recA := httptest.NewRecorder()
	wrapped.ServeHTTP(recA, reqA)
	if recA.Code != http.StatusOK {
		t.Fatalf("ip A first request status = %d, want 200", recA.Code)
	}

	recB := httptest.NewRecorder()
	wrapped.ServeHTTP(recB, reqB)
	if recB.Code != http.StatusOK {
		t.Fatalf("ip B first request should have its own bucket, status = %d, want 200", recB.Code)
	}
}

func TestClientIPStripsPort(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.10:54321"
	if ip := ClientIP(req); ip != "192.0.2.10" {
		t.Fatalf("ClientIP = %q, want 192.0.2.10", ip)
	}
}

func TestClientIPFallsBackToRawWhenNoPort(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "not-a-host-port"
	if ip := ClientIP(req); ip != "not-a-host-port" {
		t.Fatalf("ClientIP = %q, want raw RemoteAddr fallback", ip)
	}
}

func TestClientIPIgnoresXFFWithoutTrustedProxy(t *testing.T) {
	t.Parallel()
	l := New(1, 1)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.1:443"
	req.Header.Set("X-Forwarded-For", "198.51.100.50")
	if ip := l.ClientIP(req); ip != "203.0.113.1" {
		t.Fatalf("ClientIP = %q, want peer address when proxies are untrusted", ip)
	}
}

func TestClientIPHonorsXFFFromTrustedProxy(t *testing.T) {
	t.Parallel()
	l := New(1, 1)
	if err := l.SetTrustedProxies([]string{"203.0.113.0/24"}); err != nil {
		t.Fatalf("SetTrustedProxies: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.10:443"
	req.Header.Set("X-Forwarded-For", "198.51.100.50, 203.0.113.10")
	if ip := l.ClientIP(req); ip != "198.51.100.50" {
		t.Fatalf("ClientIP = %q, want original client from XFF", ip)
	}
}


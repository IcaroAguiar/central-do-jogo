// Package ratelimit implements an in-process token bucket limiter keyed by
// client IP and route, suitable for a single-process deployment (SEC-001).
// It intentionally avoids external dependencies (no Redis) per CON-003.
package ratelimit

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// bucket tracks token state for one IP+route key.
type bucket struct {
	tokens     float64
	lastRefill time.Time
}

// Limiter is a per-key token bucket rate limiter.
type Limiter struct {
	mu             sync.Mutex
	buckets        map[string]*bucket
	rate           float64 // tokens added per second
	burst          float64 // max tokens (and initial allowance)
	now            func() time.Time
	lastSwept      time.Time
	idleTTL        time.Duration
	trustedProxies []*net.IPNet
}

// New creates a limiter that allows `burst` requests immediately and refills
// at `ratePerSecond` tokens per second thereafter.
func New(ratePerSecond float64, burst int) *Limiter {
	if burst <= 0 {
		burst = 1
	}
	return &Limiter{
		buckets: make(map[string]*bucket),
		rate:    ratePerSecond,
		burst:   float64(burst),
		now:     time.Now,
		idleTTL: 10 * time.Minute,
	}
}

// SetTrustedProxies configures CIDRs whose RemoteAddr may set X-Forwarded-For.
// When empty (default), proxy headers are ignored and RemoteAddr is used.
func (l *Limiter) SetTrustedProxies(cidrs []string) error {
	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, raw := range cidrs {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if !strings.Contains(raw, "/") {
			if ip := net.ParseIP(raw); ip != nil {
				if ip.To4() != nil {
					raw = raw + "/32"
				} else {
					raw = raw + "/128"
				}
			}
		}
		_, network, err := net.ParseCIDR(raw)
		if err != nil {
			return fmt.Errorf("parse trusted proxy CIDR %q: %w", raw, err)
		}
		nets = append(nets, network)
	}
	l.mu.Lock()
	l.trustedProxies = nets
	l.mu.Unlock()
	return nil
}

// Allow reports whether a request identified by key may proceed, consuming a
// token if so.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	l.sweepLocked(now)

	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.burst - 1, lastRefill: now}
		l.buckets[key] = b
		return true
	}

	elapsed := now.Sub(b.lastRefill).Seconds()
	if elapsed > 0 {
		b.tokens += elapsed * l.rate
		if b.tokens > l.burst {
			b.tokens = l.burst
		}
		b.lastRefill = now
	}

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// sweepLocked evicts buckets that have been idle long enough to be full
// again, bounding memory use. Caller must hold l.mu.
func (l *Limiter) sweepLocked(now time.Time) {
	if now.Sub(l.lastSwept) < l.idleTTL {
		return
	}
	l.lastSwept = now
	for key, b := range l.buckets {
		if now.Sub(b.lastRefill) >= l.idleTTL {
			delete(l.buckets, key)
		}
	}
}

// Middleware wraps next, rejecting requests over the limit with HTTP 429.
// The rate limit key combines the client IP and routeLabel so different
// routes get independent budgets. onLimited, if non-nil, writes the response
// body for a limited request; otherwise a minimal text body is written.
func Middleware(limiter *Limiter, routeLabel string, onLimited http.HandlerFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := limiter.ClientIP(r) + "|" + routeLabel
			if !limiter.Allow(key) {
				if onLimited != nil {
					onLimited(w, r)
					return
				}
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ClientIP extracts the client IP. When the immediate peer is in the trusted
// proxy list and X-Forwarded-For is present, the left-most (original client)
// address is used; otherwise RemoteAddr is used. Untrusted peers cannot spoof
// XFF to bypass the per-IP budget.
func (l *Limiter) ClientIP(r *http.Request) string {
	remote := remoteHost(r)
	l.mu.Lock()
	trusted := l.trustedProxies
	l.mu.Unlock()
	if len(trusted) > 0 && ipInNets(remote, trusted) {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			candidate := strings.TrimSpace(parts[0])
			if host, _, err := net.SplitHostPort(candidate); err == nil {
				candidate = host
			}
			if net.ParseIP(candidate) != nil {
				return candidate
			}
		}
	}
	return remote
}

// ClientIP is a package helper for tests that use an empty trusted list.
func ClientIP(r *http.Request) string {
	return New(1, 1).ClientIP(r)
}

func remoteHost(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func ipInNets(host string, nets []*net.IPNet) bool {
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	for _, network := range nets {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

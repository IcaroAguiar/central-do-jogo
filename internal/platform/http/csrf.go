package httpplatform

import (
	"net/http"
	"net/url"
	"strings"
)

// OriginAllowed reports whether the request Origin (or Referer when Origin is
// absent) matches publicBaseURL. Empty publicBaseURL fails closed so CSRF
// checks never become a no-op when a mutation endpoint is registered (SEC-003).
func OriginAllowed(r *http.Request, publicBaseURL string) bool {
	if publicBaseURL == "" {
		return false
	}
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		referer := strings.TrimSpace(r.Header.Get("Referer"))
		if referer == "" {
			return false
		}
		origin = referer
	}
	base, err := url.Parse(publicBaseURL)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return false
	}
	got, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return strings.EqualFold(got.Scheme, base.Scheme) && strings.EqualFold(got.Host, base.Host)
}

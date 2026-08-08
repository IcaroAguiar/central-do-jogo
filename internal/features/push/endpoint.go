package push

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// allowedPushHost reports whether host is a known browser push service.
// Keep this tight: delivery POSTs to the endpoint, so unknown hosts are SSRF risk.
func allowedPushHost(host string) bool {
	h := strings.ToLower(strings.TrimSpace(host))
	if h == "" {
		return false
	}
	if host, _, err := net.SplitHostPort(h); err == nil {
		h = host
	}
	switch h {
	case "fcm.googleapis.com",
		"android.googleapis.com",
		"updates.push.services.mozilla.com",
		"updates-autopush.stage.mozaws.net",
		"updates-autopush.dev.mozaws.net",
		"web.push.apple.com":
		return true
	}
	if strings.HasSuffix(h, ".notify.windows.com") {
		return true
	}
	if strings.HasSuffix(h, ".push.apple.com") {
		return true
	}
	if strings.HasSuffix(h, ".push.services.mozilla.com") {
		return true
	}
	return false
}

func validatePushEndpointURL(raw string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		return nil, fmt.Errorf("%w: endpoint must be an absolute https URL", ErrInvalidSubscription)
	}
	if u.Scheme != "https" {
		return nil, fmt.Errorf("%w: endpoint must use https", ErrInvalidSubscription)
	}
	if u.User != nil {
		return nil, fmt.Errorf("%w: endpoint must not include userinfo", ErrInvalidSubscription)
	}
	host := u.Hostname()
	if ip := net.ParseIP(host); ip != nil {
		return nil, fmt.Errorf("%w: endpoint host must not be an IP address", ErrInvalidSubscription)
	}
	if !allowedPushHost(host) {
		return nil, fmt.Errorf("%w: endpoint host is not an allowed push service", ErrInvalidSubscription)
	}
	return u, nil
}

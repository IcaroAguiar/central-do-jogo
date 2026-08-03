package httpplatform

import "net/http"

// CookieValue returns the named cookie value, or "" when missing.
func CookieValue(r *http.Request, name string) string {
	c, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return c.Value
}

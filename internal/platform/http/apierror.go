package httpplatform

import (
	"encoding/json"
	"net/http"
)

// ErrorBody is the public JSON error envelope used by all /api/v1 routes.
type ErrorBody struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail carries a stable machine-readable code and a human message.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteJSON writes status and body as a JSON response.
func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// WriteError writes the standard {"error":{"code","message"}} envelope.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, ErrorBody{Error: ErrorDetail{Code: code, Message: message}})
}

// WriteRateLimited writes the standard SEC-001 rate-limit error envelope.
func WriteRateLimited(w http.ResponseWriter, _ *http.Request) {
	WriteError(w, http.StatusTooManyRequests, "rate_limited", "too many requests, try again shortly")
}

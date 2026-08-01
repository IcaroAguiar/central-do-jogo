package logging

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type ctxKey struct{}

// Fields commonly attached to structured logs.
const (
	FieldCorrelationID = "correlation_id"
	FieldSourceID      = "source_id"
	FieldJobID         = "job_id"
	FieldMatchID       = "match_id"
)

// NewJSON returns a JSON slog logger writing to stdout with secret redaction.
func NewJSON(level slog.Level) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:       level,
		ReplaceAttr: redactAttr,
	})
	return slog.New(handler)
}

// WithContext stores logger in context.
func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, logger)
}

// FromContext returns the logger from context or a default JSON logger.
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok && logger != nil {
		return logger
	}
	return NewJSON(slog.LevelInfo)
}

// WithCorrelation returns a child logger enriched with common IDs.
func WithCorrelation(logger *slog.Logger, correlationID, sourceID, jobID, matchID string) *slog.Logger {
	attrs := make([]any, 0, 8)
	if correlationID != "" {
		attrs = append(attrs, FieldCorrelationID, correlationID)
	}
	if sourceID != "" {
		attrs = append(attrs, FieldSourceID, sourceID)
	}
	if jobID != "" {
		attrs = append(attrs, FieldJobID, jobID)
	}
	if matchID != "" {
		attrs = append(attrs, FieldMatchID, matchID)
	}
	if len(attrs) == 0 {
		return logger
	}
	return logger.With(attrs...)
}

func redactAttr(_ []string, a slog.Attr) slog.Attr {
	key := strings.ToLower(a.Key)
	if looksSecretKey(key) {
		return slog.String(a.Key, "[REDACTED]")
	}
	if a.Value.Kind() == slog.KindString && looksSecretValue(a.Value.String()) {
		return slog.String(a.Key, "[REDACTED]")
	}
	return a
}

func looksSecretKey(key string) bool {
	switch {
	case strings.Contains(key, "password"),
		strings.Contains(key, "secret"),
		strings.Contains(key, "token"),
		strings.Contains(key, "authorization"),
		strings.Contains(key, "cookie"),
		strings.Contains(key, "api_key"),
		strings.Contains(key, "apikey"),
		strings.Contains(key, "database_url"),
		key == "dsn":
		return true
	default:
		return false
	}
}

func looksSecretValue(v string) bool {
	lower := strings.ToLower(v)
	return strings.Contains(lower, "password=") ||
		strings.HasPrefix(lower, "postgres://") ||
		strings.HasPrefix(lower, "postgresql://") ||
		strings.HasPrefix(lower, "bearer ")
}

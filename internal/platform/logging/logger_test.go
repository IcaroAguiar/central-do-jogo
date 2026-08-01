package logging

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestRedactSecretKeysAndDatabaseURL(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		ReplaceAttr: redactAttr,
	}))
	logger.Info("probe",
		"password", "super-secret",
		"dsn", "postgres://central:central_dev_only@127.0.0.1:5433/central_do_jogo?sslmode=disable",
		"ok", "visible",
	)

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal log: %v", err)
	}
	if payload["password"] != "[REDACTED]" {
		t.Fatalf("password not redacted: %#v", payload["password"])
	}
	if payload["dsn"] != "[REDACTED]" {
		t.Fatalf("dsn not redacted: %#v", payload["dsn"])
	}
	if payload["ok"] != "visible" {
		t.Fatalf("non-secret attr changed: %#v", payload["ok"])
	}
	if strings.Contains(buf.String(), "central_dev_only") {
		t.Fatal("raw secret leaked into log payload")
	}
}

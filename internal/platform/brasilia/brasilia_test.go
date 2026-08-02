package brasilia_test

import (
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/platform/brasilia"
)

func TestFormatBrasilia(t *testing.T) {
	t.Parallel()
	utc := time.Date(2026, 8, 2, 21, 0, 0, 0, time.UTC) // 18:00 BRT
	got := brasilia.Format(&utc)
	want := "02/08/2026 18:00"
	if got != want {
		t.Fatalf("Format = %q, want %q", got, want)
	}
	if brasilia.Format(nil) != "" {
		t.Fatal("nil should format as empty")
	}
	labeled := brasilia.FormatWithLabel(&utc)
	if labeled != want+" (Horário de Brasília)" {
		t.Fatalf("FormatWithLabel = %q", labeled)
	}
}

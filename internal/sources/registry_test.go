package sources_test

import (
	"context"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/sources"
)

type stubAdapter struct{ id string }

func (s *stubAdapter) SourceID() string { return s.id }
func (s *stubAdapter) Parse(_ context.Context, _ []byte, _ time.Time) (*sources.Observation, error) {
	return nil, nil
}

func validManifest(id string) *sources.Manifest {
	return &sources.Manifest{
		SourceID:     id,
		DisplayName:  "Test Source",
		Purpose:      "testing",
		Access:       "public",
		TermsNotes:   "none",
		RobotsNotes:  "allowed",
		RateLimit:    "none",
		Attribution:  "test",
		Stability:    "stable",
		DataTypes:    []string{"schedule"},
		RemovalNotes: "delete adapter directory",
	}
}

func TestRegistry_Register_Success(t *testing.T) {
	reg := sources.NewRegistry()
	adapter := &stubAdapter{id: "test_source"}
	manifest := validManifest("test_source")

	if err := reg.Register(adapter, manifest); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, ok := reg.Lookup("test_source")
	if !ok {
		t.Fatal("expected adapter to be found")
	}
	if got.SourceID() != "test_source" {
		t.Errorf("got source ID %q, want %q", got.SourceID(), "test_source")
	}
}

func TestRegistry_Register_NilManifest(t *testing.T) {
	reg := sources.NewRegistry()
	adapter := &stubAdapter{id: "test"}

	err := reg.Register(adapter, nil)
	if err == nil {
		t.Fatal("expected error for nil manifest")
	}
}

func TestRegistry_Register_InvalidManifest(t *testing.T) {
	reg := sources.NewRegistry()
	adapter := &stubAdapter{id: "test"}
	manifest := &sources.Manifest{SourceID: "test"}

	err := reg.Register(adapter, manifest)
	if err == nil {
		t.Fatal("expected error for invalid manifest")
	}
}

func TestRegistry_Register_IDMismatch(t *testing.T) {
	reg := sources.NewRegistry()
	adapter := &stubAdapter{id: "adapter_id"}
	manifest := validManifest("different_id")

	err := reg.Register(adapter, manifest)
	if err == nil {
		t.Fatal("expected error for ID mismatch")
	}
}

func TestRegistry_Register_Duplicate(t *testing.T) {
	reg := sources.NewRegistry()
	adapter := &stubAdapter{id: "dup"}
	manifest := validManifest("dup")

	if err := reg.Register(adapter, manifest); err != nil {
		t.Fatalf("first register: %v", err)
	}
	err := reg.Register(adapter, manifest)
	if err == nil {
		t.Fatal("expected error for duplicate registration")
	}
}

func TestRegistry_Lookup_NotFound(t *testing.T) {
	reg := sources.NewRegistry()
	_, ok := reg.Lookup("nonexistent")
	if ok {
		t.Fatal("expected adapter not found")
	}
}

func TestRegistry_All(t *testing.T) {
	reg := sources.NewRegistry()
	_ = reg.Register(&stubAdapter{id: "a"}, validManifest("a"))
	_ = reg.Register(&stubAdapter{id: "b"}, validManifest("b"))

	all := reg.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 adapters, got %d", len(all))
	}
}

package sources_test

import (
	"testing"

	"github.com/IcaroAguiar/central-do-jogo/internal/sources"
)

func TestManifest_Validate_Complete(t *testing.T) {
	m := validManifest("test")
	if err := m.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManifest_Validate_MissingFields(t *testing.T) {
	m := &sources.Manifest{}
	err := m.Validate()
	if err == nil {
		t.Fatal("expected validation error for empty manifest")
	}
}

func TestParseManifest_ValidYAML(t *testing.T) {
	yaml := []byte(`
source_id: test
display_name: Test
purpose: testing
access: public
terms_notes: none
robots_notes: allowed
rate_limit: none
attribution: test
stability: stable
data_types: [schedule]
removal_notes: remove directory
`)
	m, err := sources.ParseManifest(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.SourceID != "test" {
		t.Errorf("got source_id %q, want %q", m.SourceID, "test")
	}
}

func TestParseManifest_InvalidYAML(t *testing.T) {
	_, err := sources.ParseManifest([]byte("{{invalid"))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestParseManifest_IncompleteManifest(t *testing.T) {
	yaml := []byte(`source_id: test`)
	_, err := sources.ParseManifest(yaml)
	if err == nil {
		t.Fatal("expected validation error for incomplete manifest")
	}
}

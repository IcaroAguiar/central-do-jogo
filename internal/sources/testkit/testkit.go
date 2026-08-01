// Package testkit provides helpers for source adapter tests.
package testkit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/IcaroAguiar/central-do-jogo/internal/sources"
)

// LoadFixture reads a fixture file relative to the calling test's directory.
func LoadFixture(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("loading fixture %s: %v", path, err)
	}
	return data
}

// LoadManifestFromDir loads and validates a manifest.yaml from the given directory.
func LoadManifestFromDir(t *testing.T, dir string) *sources.Manifest {
	t.Helper()
	m, err := sources.LoadManifest(filepath.Join(dir, "manifest.yaml"))
	if err != nil {
		t.Fatalf("loading manifest from %s: %v", dir, err)
	}
	return m
}

// AssertEvidencePreserved verifies that an observation carries all required
// evidence metadata fields.
func AssertEvidencePreserved(t *testing.T, obs *sources.Observation) {
	t.Helper()
	if obs == nil {
		t.Fatal("observation is nil")
	}
	if obs.SourceID == "" {
		t.Error("observation missing SourceID")
	}
	if obs.DataType == "" {
		t.Error("observation missing DataType")
	}
	if obs.ObservedAt.IsZero() {
		t.Error("observation missing ObservedAt")
	}
	if obs.ParserVersion == "" {
		t.Error("observation missing ParserVersion")
	}
	if obs.ContentHash == "" {
		t.Error("observation missing ContentHash")
	}
}

// AssertScheduleNotEmpty checks that at least one schedule entry exists.
func AssertScheduleNotEmpty(t *testing.T, obs *sources.Observation) {
	t.Helper()
	if len(obs.Schedules) == 0 {
		t.Error("expected at least one schedule entry")
	}
}

// AssertLineupsNotEmpty checks that at least one lineup entry exists.
func AssertLineupsNotEmpty(t *testing.T, obs *sources.Observation) {
	t.Helper()
	if len(obs.Lineups) == 0 {
		t.Error("expected at least one lineup entry")
	}
}

// AssertNewsNotEmpty checks that at least one news entry exists.
func AssertNewsNotEmpty(t *testing.T, obs *sources.Observation) {
	t.Helper()
	if len(obs.NewsLinks) == 0 {
		t.Error("expected at least one news link entry")
	}
}

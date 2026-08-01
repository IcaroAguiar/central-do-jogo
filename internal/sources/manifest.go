package sources

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Manifest describes a source adapter per SEC-004 / PAT-003.
type Manifest struct {
	SourceID     string   `yaml:"source_id"`
	DisplayName  string   `yaml:"display_name"`
	Purpose      string   `yaml:"purpose"`
	Access       string   `yaml:"access"`
	TermsNotes   string   `yaml:"terms_notes"`
	RobotsNotes  string   `yaml:"robots_notes"`
	RateLimit    string   `yaml:"rate_limit"`
	Attribution  string   `yaml:"attribution"`
	Stability    string   `yaml:"stability"`
	DataTypes    []string `yaml:"data_types"`
	RemovalNotes string   `yaml:"removal_notes"`
}

// Validate rejects incomplete manifests per SEC-004.
func (m *Manifest) Validate() error {
	var errs []string
	if m.SourceID == "" {
		errs = append(errs, "source_id is required")
	}
	if m.DisplayName == "" {
		errs = append(errs, "display_name is required")
	}
	if m.Purpose == "" {
		errs = append(errs, "purpose is required")
	}
	if m.Access == "" {
		errs = append(errs, "access is required")
	}
	if m.TermsNotes == "" {
		errs = append(errs, "terms_notes is required")
	}
	if m.RobotsNotes == "" {
		errs = append(errs, "robots_notes is required")
	}
	if m.RateLimit == "" {
		errs = append(errs, "rate_limit is required")
	}
	if m.Attribution == "" {
		errs = append(errs, "attribution is required")
	}
	if m.Stability == "" {
		errs = append(errs, "stability is required")
	}
	if len(m.DataTypes) == 0 {
		errs = append(errs, "data_types must have at least one entry")
	}
	if m.RemovalNotes == "" {
		errs = append(errs, "removal_notes is required")
	}
	if len(errs) > 0 {
		return fmt.Errorf("manifest validation failed: %s", strings.Join(errs, "; "))
	}
	return nil
}

// LoadManifest reads and validates a manifest YAML file.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest %s: %w", path, err)
	}
	return ParseManifest(data)
}

// ParseManifest parses and validates manifest YAML bytes.
func ParseManifest(data []byte) (*Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing manifest YAML: %w", err)
	}
	if err := m.Validate(); err != nil {
		return nil, errors.Join(fmt.Errorf("invalid manifest"), err)
	}
	return &m, nil
}

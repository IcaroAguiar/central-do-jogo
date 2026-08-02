package api_test

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"

	"github.com/IcaroAguiar/central-do-jogo/internal/api"
	"gopkg.in/yaml.v3"
)

func TestSSRPagesMatchOpenAPI(t *testing.T) {
	t.Parallel()
	openapiPath := filepath.Join(repoRoot(t), "api", "openapi.yaml")
	raw, err := os.ReadFile(openapiPath)
	if err != nil {
		t.Fatalf("read openapi: %v", err)
	}
	var doc struct {
		Components struct {
			Schemas map[string]struct {
				Enum []string `yaml:"enum"`
			} `yaml:"schemas"`
		} `yaml:"components"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse openapi: %v", err)
	}
	schema, ok := doc.Components.Schemas["SSRPage"]
	if !ok {
		t.Fatal("OpenAPI missing components.schemas.SSRPage")
	}
	want := api.AllSSRPages()
	if len(schema.Enum) != len(want) {
		t.Fatalf("SSRPage enum len = %d, want %d (%v vs %v)", len(schema.Enum), len(want), schema.Enum, want)
	}
	for i, page := range want {
		if schema.Enum[i] != page {
			t.Fatalf("SSRPage enum[%d] = %q, want %q", i, schema.Enum[i], page)
		}
	}
}

func TestSSRPagesMatchTypeScriptContract(t *testing.T) {
	t.Parallel()
	tsPath := filepath.Join(repoRoot(t), "web", "src", "lib", "pages.ts")
	raw, err := os.ReadFile(tsPath)
	if err != nil {
		t.Fatalf("read pages.ts: %v", err)
	}
	src := string(raw)
	// Require exact key→value mappings so a swapped SSR_PAGE entry fails.
	for _, page := range api.AllSSRPages() {
		pattern := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(page) + `:\s*"` + regexp.QuoteMeta(page) + `"\s*,?\s*$`)
		if !pattern.MatchString(src) {
			t.Fatalf("web/src/lib/pages.ts missing exact SSR_PAGE mapping %s: %q", page, page)
		}
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

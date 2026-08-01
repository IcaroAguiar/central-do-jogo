package gazetaesportiva_rss_test

import (
	"context"
	"testing"

	"github.com/IcaroAguiar/central-do-jogo/internal/sources/gazetaesportiva_rss"
	"github.com/IcaroAguiar/central-do-jogo/internal/sources/testkit"
)

func TestAdapter_SourceID(t *testing.T) {
	a := &gazetaesportiva_rss.Adapter{}
	if a.SourceID() != "gazetaesportiva_rss" {
		t.Errorf("got %q, want %q", a.SourceID(), "gazetaesportiva_rss")
	}
}

func TestAdapter_Parse_Fixture(t *testing.T) {
	a := &gazetaesportiva_rss.Adapter{}
	fixture := testkit.LoadFixture(t, "fixtures/gazetaesportiva-sample.xml")

	obs, err := a.Parse(context.Background(), fixture)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	testkit.AssertEvidencePreserved(t, obs)
	testkit.AssertNewsNotEmpty(t, obs)

	if len(obs.NewsLinks) != 3 {
		t.Errorf("expected 3 news links, got %d", len(obs.NewsLinks))
	}

	first := obs.NewsLinks[0]
	if first.Title == "" {
		t.Error("expected non-empty title")
	}
	if first.URL == "" {
		t.Error("expected non-empty URL")
	}
	if first.PublishedAt.IsZero() {
		t.Error("expected non-zero publish time")
	}
}

func TestAdapter_Parse_InvalidXML(t *testing.T) {
	a := &gazetaesportiva_rss.Adapter{}
	_, err := a.Parse(context.Background(), []byte("not xml at all <<<"))
	if err == nil {
		t.Fatal("expected error for invalid XML")
	}
}

func TestAdapter_Parse_EmptyFeed(t *testing.T) {
	a := &gazetaesportiva_rss.Adapter{}
	fixture := []byte(`<?xml version="1.0"?><rss><channel></channel></rss>`)

	obs, err := a.Parse(context.Background(), fixture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(obs.NewsLinks) != 0 {
		t.Errorf("expected 0 news links for empty feed, got %d", len(obs.NewsLinks))
	}
}

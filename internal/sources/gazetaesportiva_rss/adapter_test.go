package gazetaesportiva_rss_test

import (
	"context"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/sources/gazetaesportiva_rss"
	"github.com/IcaroAguiar/central-do-jogo/internal/sources/testkit"
)

var fixedTime = time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

func TestAdapter_SourceID(t *testing.T) {
	a := &gazetaesportiva_rss.Adapter{}
	if a.SourceID() != "gazetaesportiva_rss" {
		t.Errorf("got %q, want %q", a.SourceID(), "gazetaesportiva_rss")
	}
}

func TestAdapter_Parse_Fixture(t *testing.T) {
	a := &gazetaesportiva_rss.Adapter{}
	fixture := testkit.LoadFixture(t, "fixtures/gazetaesportiva-sample.xml")

	obs, err := a.Parse(context.Background(), fixture, fixedTime)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	testkit.AssertEvidencePreserved(t, obs)
	testkit.AssertNewsNotEmpty(t, obs)

	if obs.ObservedAt != fixedTime {
		t.Errorf("observedAt: got %v, want %v", obs.ObservedAt, fixedTime)
	}

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
	_, err := a.Parse(context.Background(), []byte("not xml at all <<<"), fixedTime)
	if err == nil {
		t.Fatal("expected error for invalid XML")
	}
}

func TestAdapter_Parse_EmptyInput(t *testing.T) {
	a := &gazetaesportiva_rss.Adapter{}
	_, err := a.Parse(context.Background(), []byte(""), fixedTime)
	if err == nil {
		t.Fatal("expected fail-closed error for empty input")
	}
}

func TestAdapter_Parse_EmptyFeed(t *testing.T) {
	a := &gazetaesportiva_rss.Adapter{}
	fixture := []byte(`<?xml version="1.0"?><rss><channel></channel></rss>`)

	_, err := a.Parse(context.Background(), fixture, fixedTime)
	if err == nil {
		t.Fatal("expected fail-closed error for feed with zero usable items")
	}
}

func TestAdapter_Parse_ZeroObservedAt(t *testing.T) {
	a := &gazetaesportiva_rss.Adapter{}
	_, err := a.Parse(context.Background(), []byte("<rss/>"), time.Time{})
	if err == nil {
		t.Fatal("expected error for zero observedAt")
	}
}

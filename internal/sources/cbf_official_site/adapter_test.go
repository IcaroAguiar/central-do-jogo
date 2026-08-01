package cbf_official_site_test

import (
	"context"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/sources/cbf_official_site"
	"github.com/IcaroAguiar/central-do-jogo/internal/sources/testkit"
)

var fixedTime = time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

func TestAdapter_SourceID(t *testing.T) {
	a := &cbf_official_site.Adapter{}
	if a.SourceID() != "cbf_official_site" {
		t.Errorf("got %q, want %q", a.SourceID(), "cbf_official_site")
	}
}

func TestAdapter_Parse_Fixture(t *testing.T) {
	a := &cbf_official_site.Adapter{}
	fixture := testkit.LoadFixture(t, "fixtures/cbf-tabela-basica-2026.meta.json")

	obs, err := a.Parse(context.Background(), fixture, fixedTime)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	testkit.AssertEvidencePreserved(t, obs)
	testkit.AssertScheduleNotEmpty(t, obs)

	if obs.ObservedAt != fixedTime {
		t.Errorf("observedAt: got %v, want %v", obs.ObservedAt, fixedTime)
	}

	if len(obs.Schedules) != 2 {
		t.Errorf("expected 2 schedule entries (round1 dates), got %d", len(obs.Schedules))
	}

	for _, s := range obs.Schedules {
		if s.Competition != "Brasileiro Serie A" {
			t.Errorf("competition: got %q, want %q", s.Competition, "Brasileiro Serie A")
		}
		if s.KickoffAt == nil {
			t.Error("expected kickoff date to be parsed")
		}
	}
}

func TestAdapter_Parse_InvalidJSON(t *testing.T) {
	a := &cbf_official_site.Adapter{}
	_, err := a.Parse(context.Background(), []byte("{invalid"), fixedTime)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestAdapter_Parse_EmptyInput(t *testing.T) {
	a := &cbf_official_site.Adapter{}
	_, err := a.Parse(context.Background(), []byte(""), fixedTime)
	if err == nil {
		t.Fatal("expected fail-closed error for empty input")
	}
}

func TestAdapter_Parse_FailClosed_NoData(t *testing.T) {
	a := &cbf_official_site.Adapter{}
	fixture := []byte(`{
		"source_id": "cbf_official_site",
		"source_url": "https://cbf.com.br/empty",
		"observed": {
			"announces_basic_table_pdf": false,
			"round1_dates": []
		}
	}`)
	_, err := a.Parse(context.Background(), fixture, fixedTime)
	if err == nil {
		t.Fatal("expected fail-closed error when no data found")
	}
}

func TestAdapter_Parse_ZeroObservedAt(t *testing.T) {
	a := &cbf_official_site.Adapter{}
	_, err := a.Parse(context.Background(), []byte(`{}`), time.Time{})
	if err == nil {
		t.Fatal("expected error for zero observedAt")
	}
}

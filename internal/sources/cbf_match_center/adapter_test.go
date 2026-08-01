package cbf_match_center_test

import (
	"context"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/IcaroAguiar/central-do-jogo/internal/sources/cbf_match_center"
	"github.com/IcaroAguiar/central-do-jogo/internal/sources/testkit"
)

var fixedTime = time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

func TestAdapter_SourceID(t *testing.T) {
	a := &cbf_match_center.Adapter{}
	if a.SourceID() != "cbf_match_center" {
		t.Errorf("got %q, want %q", a.SourceID(), "cbf_match_center")
	}
}

func TestAdapter_Parse_Fixture(t *testing.T) {
	a := &cbf_match_center.Adapter{}
	fixture := testkit.LoadFixture(t, "fixtures/cbf-fluminense-gremio-2025.redacted.json")

	obs, err := a.Parse(context.Background(), fixture, fixedTime)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	testkit.AssertEvidencePreserved(t, obs)
	testkit.AssertLineupsNotEmpty(t, obs)

	if obs.ObservedAt != fixedTime {
		t.Errorf("observedAt: got %v, want %v", obs.ObservedAt, fixedTime)
	}

	if len(obs.Lineups) != 2 {
		t.Fatalf("expected 2 lineup entries (home+away), got %d", len(obs.Lineups))
	}

	home := obs.Lineups[0]
	if home.Side != domain.LineupHome {
		t.Errorf("first lineup side: got %q, want %q", home.Side, domain.LineupHome)
	}
	if home.HomeTeam != "Fluminense" {
		t.Errorf("home team: got %q, want %q", home.HomeTeam, "Fluminense")
	}
	if len(home.Players) != 11 {
		t.Errorf("home players: got %d, want 11", len(home.Players))
	}
	if !home.Official {
		t.Error("expected lineup to be marked as official")
	}

	away := obs.Lineups[1]
	if away.Side != domain.LineupAway {
		t.Errorf("second lineup side: got %q, want %q", away.Side, domain.LineupAway)
	}
	if len(away.Players) != 11 {
		t.Errorf("away players: got %d, want 11", len(away.Players))
	}
}

func TestAdapter_Parse_InvalidJSON(t *testing.T) {
	a := &cbf_match_center.Adapter{}
	_, err := a.Parse(context.Background(), []byte("not json"), fixedTime)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestAdapter_Parse_EmptyInput(t *testing.T) {
	a := &cbf_match_center.Adapter{}
	_, err := a.Parse(context.Background(), []byte(""), fixedTime)
	if err == nil {
		t.Fatal("expected fail-closed error for empty input")
	}
}

func TestAdapter_Parse_FailClosed_Empty(t *testing.T) {
	a := &cbf_match_center.Adapter{}
	fixture := []byte(`{
		"source_id": "cbf_match_center",
		"sample_url": "https://cbf.com.br/test",
		"home_team": "A",
		"away_team": "B",
		"home_starting_sample": [],
		"away_starting_sample": []
	}`)
	_, err := a.Parse(context.Background(), fixture, fixedTime)
	if err == nil {
		t.Fatal("expected fail-closed error for empty lineups")
	}
}

func TestAdapter_Parse_ZeroObservedAt(t *testing.T) {
	a := &cbf_match_center.Adapter{}
	_, err := a.Parse(context.Background(), []byte(`{}`), time.Time{})
	if err == nil {
		t.Fatal("expected error for zero observedAt")
	}
}

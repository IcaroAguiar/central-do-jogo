package openfootball_brazil_test

import (
	"context"
	"testing"

	"github.com/IcaroAguiar/central-do-jogo/internal/sources/openfootball_brazil"
	"github.com/IcaroAguiar/central-do-jogo/internal/sources/testkit"
)

func TestAdapter_SourceID(t *testing.T) {
	a := &openfootball_brazil.Adapter{}
	if a.SourceID() != "openfootball_brazil" {
		t.Errorf("got %q, want %q", a.SourceID(), "openfootball_brazil")
	}
}

func TestAdapter_Parse_Fixture(t *testing.T) {
	a := &openfootball_brazil.Adapter{}
	fixture := testkit.LoadFixture(t, "fixtures/matchday1.txt")

	obs, err := a.Parse(context.Background(), fixture)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	testkit.AssertEvidencePreserved(t, obs)
	testkit.AssertScheduleNotEmpty(t, obs)

	if len(obs.Schedules) != 10 {
		t.Errorf("expected 10 matches, got %d", len(obs.Schedules))
	}

	first := obs.Schedules[0]
	if first.HomeTeam != "CA Mineiro" {
		t.Errorf("first home team: got %q, want %q", first.HomeTeam, "CA Mineiro")
	}
	if first.AwayTeam != "SE Palmeiras" {
		t.Errorf("first away team: got %q, want %q", first.AwayTeam, "SE Palmeiras")
	}
	if first.Round != "Matchday 1" {
		t.Errorf("round: got %q, want %q", first.Round, "Matchday 1")
	}
	if first.KickoffAt == nil {
		t.Error("expected kickoff time to be parsed")
	}
}

func TestAdapter_Parse_Empty(t *testing.T) {
	a := &openfootball_brazil.Adapter{}
	obs, err := a.Parse(context.Background(), []byte(""))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(obs.Schedules) != 0 {
		t.Errorf("expected 0 schedules for empty input, got %d", len(obs.Schedules))
	}
}

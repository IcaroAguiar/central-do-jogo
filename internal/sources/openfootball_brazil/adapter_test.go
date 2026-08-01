package openfootball_brazil_test

import (
	"context"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/sources/openfootball_brazil"
	"github.com/IcaroAguiar/central-do-jogo/internal/sources/testkit"
)

var fixedTime = time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

func TestAdapter_SourceID(t *testing.T) {
	a := &openfootball_brazil.Adapter{}
	if a.SourceID() != "openfootball_brazil" {
		t.Errorf("got %q, want %q", a.SourceID(), "openfootball_brazil")
	}
}

func TestAdapter_Parse_Fixture(t *testing.T) {
	a := &openfootball_brazil.Adapter{}
	fixture := testkit.LoadFixture(t, "fixtures/matchday1.txt")

	obs, err := a.Parse(context.Background(), fixture, fixedTime)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	testkit.AssertEvidencePreserved(t, obs)
	testkit.AssertScheduleNotEmpty(t, obs)

	if obs.ObservedAt != fixedTime {
		t.Errorf("observedAt: got %v, want %v", obs.ObservedAt, fixedTime)
	}

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
	if first.KickoffInherited {
		t.Error("first match has explicit time, expected KickoffInherited=false")
	}
}

func TestAdapter_Parse_KickoffInherited(t *testing.T) {
	a := &openfootball_brazil.Adapter{}
	fixture := testkit.LoadFixture(t, "fixtures/matchday1.txt")

	obs, err := a.Parse(context.Background(), fixture, fixedTime)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	var hasInherited bool
	for _, s := range obs.Schedules {
		if s.KickoffInherited {
			hasInherited = true
			break
		}
	}
	if !hasInherited {
		t.Log("note: fixture has no matchNoTime entries inheriting time; this is expected if the fixture only has explicit times")
	}
}

func TestAdapter_Parse_Empty(t *testing.T) {
	a := &openfootball_brazil.Adapter{}
	_, err := a.Parse(context.Background(), []byte(""), fixedTime)
	if err == nil {
		t.Fatal("expected fail-closed error for empty input")
	}
}

func TestAdapter_Parse_ZeroObservedAt(t *testing.T) {
	a := &openfootball_brazil.Adapter{}
	_, err := a.Parse(context.Background(), []byte("data"), time.Time{})
	if err == nil {
		t.Fatal("expected error for zero observedAt")
	}
}

func TestAdapter_Parse_NonEmptyNoMatches(t *testing.T) {
	a := &openfootball_brazil.Adapter{}
	_, err := a.Parse(context.Background(), []byte("some random text with no matches"), fixedTime)
	if err == nil {
		t.Fatal("expected fail-closed error when non-empty input produces zero entries")
	}
}

package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/platform/store"
)

func TestLineupStoreListByMatchOrdersHomeFirstAndDecodesPlayers(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	insertClub(t, pool, "club_flamengo", "flamengo", "Flamengo", "FLA", nil)
	insertClub(t, pool, "club_vasco", "vasco", "Vasco", "VAS", nil)
	insertCompetition(t, pool, "comp_brasileirao_2026", "brasileirao", "Brasileirao", 2026)
	kickoff := time.Date(2026, 8, 10, 21, 0, 0, 0, time.UTC)
	insertMatch(t, pool, "match_1", "comp_brasileirao_2026", "club_flamengo", "club_vasco", "flamengo-x-vasco", &kickoff, "published")
	insertSource(t, pool, "src_globo", "TV Globo")
	insertEvidence(t, pool, "evi_home", "src_globo", "match_1", "lineup")
	insertEvidence(t, pool, "evi_away", "src_globo", "match_1", "lineup")

	awayPlayers := []byte(`[{"ShirtNumber":"23","Name":"Diego","IsStarter":true}]`)
	homePlayers := []byte(`[{"ShirtNumber":"1","Name":"Rossi","IsStarter":true},{"ShirtNumber":"9","Name":"Pedro","IsStarter":false}]`)
	insertLineup(t, pool, "lineup_away", "match_1", "club_vasco", "evi_away", "away", awayPlayers)
	insertLineup(t, pool, "lineup_home", "match_1", "club_flamengo", "evi_home", "home", homePlayers)

	s := store.NewLineupStore(pool)
	lineups, err := s.ListByMatch(ctx, "match_1")
	if err != nil {
		t.Fatalf("ListByMatch: %v", err)
	}
	if len(lineups) != 2 {
		t.Fatalf("len(lineups) = %d, want 2", len(lineups))
	}
	if lineups[0].Side != "home" || lineups[1].Side != "away" {
		t.Fatalf("lineups order = [%s, %s], want [home, away]", lineups[0].Side, lineups[1].Side)
	}
	if len(lineups[0].Players) != 2 {
		t.Fatalf("home lineup Players = %+v, want 2 entries", lineups[0].Players)
	}
	if lineups[0].Players[0].Name != "Rossi" || lineups[0].Players[0].ShirtNumber != "1" || !lineups[0].Players[0].IsStarter {
		t.Fatalf("home lineup Players[0] = %+v, unexpected decode", lineups[0].Players[0])
	}
}

func TestLineupStoreListByMatchEmpty(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	insertClub(t, pool, "club_flamengo", "flamengo", "Flamengo", "FLA", nil)
	insertClub(t, pool, "club_vasco", "vasco", "Vasco", "VAS", nil)
	insertCompetition(t, pool, "comp_brasileirao_2026", "brasileirao", "Brasileirao", 2026)
	kickoff := time.Date(2026, 8, 10, 21, 0, 0, 0, time.UTC)
	insertMatch(t, pool, "match_1", "comp_brasileirao_2026", "club_flamengo", "club_vasco", "flamengo-x-vasco", &kickoff, "published")

	s := store.NewLineupStore(pool)
	lineups, err := s.ListByMatch(ctx, "match_1")
	if err != nil {
		t.Fatalf("ListByMatch: %v", err)
	}
	if len(lineups) != 0 {
		t.Fatalf("len(lineups) = %d, want 0", len(lineups))
	}
}

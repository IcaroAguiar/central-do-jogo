package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/platform/store"
)

func TestMatchStoreGetBySlug(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	insertClub(t, pool, "club_flamengo", "flamengo", "Flamengo", "FLA", nil)
	insertClub(t, pool, "club_vasco", "vasco", "Vasco", "VAS", nil)
	insertCompetition(t, pool, "comp_brasileirao_2026", "brasileirao", "Brasileirao", 2026)
	kickoff := time.Date(2026, 8, 10, 21, 0, 0, 0, time.UTC)
	insertMatch(t, pool, "match_1", "comp_brasileirao_2026", "club_flamengo", "club_vasco", "flamengo-x-vasco", &kickoff, "published")

	s := store.NewMatchStore(pool)
	rec, err := s.GetBySlug(ctx, "flamengo-x-vasco")
	if err != nil {
		t.Fatalf("GetBySlug: %v", err)
	}
	if rec == nil {
		t.Fatal("expected match, got nil")
	}
	if rec.HomeClub.Slug != "flamengo" || rec.AwayClub.Slug != "vasco" {
		t.Fatalf("rec = %+v, unexpected club summaries", rec)
	}
	if rec.Competition.Season != 2026 {
		t.Fatalf("Competition.Season = %d, want 2026", rec.Competition.Season)
	}
	if rec.KickoffAt == nil || !rec.KickoffAt.Equal(kickoff) {
		t.Fatalf("KickoffAt = %v, want %v", rec.KickoffAt, kickoff)
	}
	if rec.KickoffAt.Location() != time.UTC {
		t.Fatalf("KickoffAt should be normalized to UTC, got %v", rec.KickoffAt.Location())
	}
}

func TestMatchStoreGetBySlugNotFound(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	s := store.NewMatchStore(pool)

	rec, err := s.GetBySlug(ctx, "does-not-exist")
	if err != nil {
		t.Fatalf("GetBySlug: %v", err)
	}
	if rec != nil {
		t.Fatalf("expected nil, got %+v", rec)
	}
}

func TestMatchStoreSearch(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	insertClub(t, pool, "club_flamengo", "flamengo", "Flamengo", "FLA", nil)
	insertClub(t, pool, "club_vasco", "vasco", "Vasco", "VAS", nil)
	insertCompetition(t, pool, "comp_brasileirao_2026", "brasileirao", "Brasileirao", 2026)
	kickoff := time.Date(2026, 8, 10, 21, 0, 0, 0, time.UTC)
	insertMatch(t, pool, "match_1", "comp_brasileirao_2026", "club_flamengo", "club_vasco", "flamengo-x-vasco", &kickoff, "published")

	s := store.NewMatchStore(pool)
	records, err := s.Search(ctx, "vasco", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(records) != 1 || records[0].Slug != "flamengo-x-vasco" {
		t.Fatalf("records = %+v, want one flamengo-x-vasco hit", records)
	}
}

func TestMatchStoreListByClubSeasonIncludesIndefiniteKickoff(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	insertClub(t, pool, "club_flamengo", "flamengo", "Flamengo", "FLA", nil)
	insertClub(t, pool, "club_vasco", "vasco", "Vasco", "VAS", nil)
	insertClub(t, pool, "club_palmeiras", "palmeiras", "Palmeiras", "PAL", nil)
	insertCompetition(t, pool, "comp_brasileirao_2026", "brasileirao", "Brasileirao", 2026)

	scheduled := time.Date(2026, 8, 10, 21, 0, 0, 0, time.UTC)
	insertMatch(t, pool, "match_1", "comp_brasileirao_2026", "club_flamengo", "club_vasco", "flamengo-x-vasco", &scheduled, "published")
	insertMatch(t, pool, "match_2", "comp_brasileirao_2026", "club_flamengo", "club_palmeiras", "flamengo-x-palmeiras", nil, "indefinite")

	s := store.NewMatchStore(pool)
	records, err := s.ListByClub(ctx, "club_flamengo", 2026, nil, nil)
	if err != nil {
		t.Fatalf("ListByClub: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("len(records) = %d, want 2 (season range includes indefinite kickoff)", len(records))
	}
}

func TestMatchStoreListByClubRangeExcludesIndefiniteKickoff(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	insertClub(t, pool, "club_flamengo", "flamengo", "Flamengo", "FLA", nil)
	insertClub(t, pool, "club_vasco", "vasco", "Vasco", "VAS", nil)
	insertClub(t, pool, "club_palmeiras", "palmeiras", "Palmeiras", "PAL", nil)
	insertCompetition(t, pool, "comp_brasileirao_2026", "brasileirao", "Brasileirao", 2026)

	scheduled := time.Date(2026, 8, 10, 21, 0, 0, 0, time.UTC)
	insertMatch(t, pool, "match_1", "comp_brasileirao_2026", "club_flamengo", "club_vasco", "flamengo-x-vasco", &scheduled, "published")
	insertMatch(t, pool, "match_2", "comp_brasileirao_2026", "club_flamengo", "club_palmeiras", "flamengo-x-palmeiras", nil, "indefinite")

	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

	s := store.NewMatchStore(pool)
	records, err := s.ListByClub(ctx, "club_flamengo", 2026, &start, &end)
	if err != nil {
		t.Fatalf("ListByClub: %v", err)
	}
	if len(records) != 1 || records[0].Slug != "flamengo-x-vasco" {
		t.Fatalf("records = %+v, want only the scheduled match within range", records)
	}
}

func TestMatchStoreListByClubFiltersBySeason(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	insertClub(t, pool, "club_flamengo", "flamengo", "Flamengo", "FLA", nil)
	insertClub(t, pool, "club_vasco", "vasco", "Vasco", "VAS", nil)
	insertCompetition(t, pool, "comp_brasileirao_2025", "brasileirao", "Brasileirao", 2025)
	insertCompetition(t, pool, "comp_brasileirao_2026", "brasileirao", "Brasileirao", 2026)

	kickoff := time.Date(2025, 8, 10, 21, 0, 0, 0, time.UTC)
	insertMatch(t, pool, "match_2025", "comp_brasileirao_2025", "club_flamengo", "club_vasco", "flamengo-x-vasco-2025", &kickoff, "published")
	kickoff2026 := time.Date(2026, 8, 10, 21, 0, 0, 0, time.UTC)
	insertMatch(t, pool, "match_2026", "comp_brasileirao_2026", "club_flamengo", "club_vasco", "flamengo-x-vasco-2026", &kickoff2026, "published")

	s := store.NewMatchStore(pool)
	records, err := s.ListByClub(ctx, "club_flamengo", 2026, nil, nil)
	if err != nil {
		t.Fatalf("ListByClub: %v", err)
	}
	if len(records) != 1 || records[0].Slug != "flamengo-x-vasco-2026" {
		t.Fatalf("records = %+v, want only the 2026 season match", records)
	}
}

func TestMatchStoreListByClubIncludesHomeAndAway(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	insertClub(t, pool, "club_flamengo", "flamengo", "Flamengo", "FLA", nil)
	insertClub(t, pool, "club_vasco", "vasco", "Vasco", "VAS", nil)
	insertCompetition(t, pool, "comp_brasileirao_2026", "brasileirao", "Brasileirao", 2026)

	k1 := time.Date(2026, 8, 10, 21, 0, 0, 0, time.UTC)
	k2 := time.Date(2026, 9, 10, 21, 0, 0, 0, time.UTC)
	insertMatch(t, pool, "match_home", "comp_brasileirao_2026", "club_flamengo", "club_vasco", "flamengo-x-vasco", &k1, "published")
	insertMatch(t, pool, "match_away", "comp_brasileirao_2026", "club_vasco", "club_flamengo", "vasco-x-flamengo", &k2, "published")

	s := store.NewMatchStore(pool)
	records, err := s.ListByClub(ctx, "club_flamengo", 2026, nil, nil)
	if err != nil {
		t.Fatalf("ListByClub: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("len(records) = %d, want 2 (home and away)", len(records))
	}
}

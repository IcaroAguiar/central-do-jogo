package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/platform/store"
)

func TestBroadcastStoreListByMatchOrdersByConfidence(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	insertClub(t, pool, "club_flamengo", "flamengo", "Flamengo", "FLA", nil)
	insertClub(t, pool, "club_vasco", "vasco", "Vasco", "VAS", nil)
	insertCompetition(t, pool, "comp_brasileirao_2026", "brasileirao", "Brasileirao", 2026)
	kickoff := time.Date(2026, 8, 10, 21, 0, 0, 0, time.UTC)
	insertMatch(t, pool, "match_1", "comp_brasileirao_2026", "club_flamengo", "club_vasco", "flamengo-x-vasco", &kickoff, "published")
	insertSource(t, pool, "src_globo", "TV Globo")
	insertEvidence(t, pool, "evi_1", "src_globo", "match_1", "broadcast")
	insertEvidence(t, pool, "evi_2", "src_globo", "match_1", "broadcast")

	verifiedAt := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	insertBroadcast(t, pool, "bc_low", "match_1", "evi_1", "Premiere", "low", verifiedAt)
	insertBroadcast(t, pool, "bc_high", "match_1", "evi_2", "Globo", "high", verifiedAt)

	s := store.NewBroadcastStore(pool)
	records, err := s.ListByMatch(ctx, "match_1")
	if err != nil {
		t.Fatalf("ListByMatch: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("len(records) = %d, want 2", len(records))
	}
	if records[0].Channel != "Globo" || records[0].Confidence != "high" {
		t.Fatalf("records[0] = %+v, want high-confidence Globo first", records[0])
	}
	if records[0].SourceDisplayName != "TV Globo" {
		t.Fatalf("SourceDisplayName = %q, want %q", records[0].SourceDisplayName, "TV Globo")
	}
	if records[0].VerifiedAt.Location() != time.UTC {
		t.Fatalf("VerifiedAt should be normalized to UTC, got %v", records[0].VerifiedAt.Location())
	}
}

func TestBroadcastStoreListByMatchEmpty(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	insertClub(t, pool, "club_flamengo", "flamengo", "Flamengo", "FLA", nil)
	insertClub(t, pool, "club_vasco", "vasco", "Vasco", "VAS", nil)
	insertCompetition(t, pool, "comp_brasileirao_2026", "brasileirao", "Brasileirao", 2026)
	kickoff := time.Date(2026, 8, 10, 21, 0, 0, 0, time.UTC)
	insertMatch(t, pool, "match_1", "comp_brasileirao_2026", "club_flamengo", "club_vasco", "flamengo-x-vasco", &kickoff, "published")

	s := store.NewBroadcastStore(pool)
	records, err := s.ListByMatch(ctx, "match_1")
	if err != nil {
		t.Fatalf("ListByMatch: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("len(records) = %d, want 0", len(records))
	}
}

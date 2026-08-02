package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/platform/store"
)

func TestNewsStoreListByMatchOrdersByPublishedAtDesc(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	insertClub(t, pool, "club_flamengo", "flamengo", "Flamengo", "FLA", nil)
	insertClub(t, pool, "club_vasco", "vasco", "Vasco", "VAS", nil)
	insertCompetition(t, pool, "comp_brasileirao_2026", "brasileirao", "Brasileirao", 2026)
	kickoff := time.Date(2026, 8, 10, 21, 0, 0, 0, time.UTC)
	insertMatch(t, pool, "match_1", "comp_brasileirao_2026", "club_flamengo", "club_vasco", "flamengo-x-vasco", &kickoff, "published")
	insertSource(t, pool, "src_ge", "ge.globo.com")
	insertEvidence(t, pool, "evi_1", "src_ge", "match_1", "news")
	insertEvidence(t, pool, "evi_2", "src_ge", "match_1", "news")

	older := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	insertNewsLink(t, pool, "news_old", "match_1", "evi_1", "src_ge", "Old news", "https://ge.globo.com/old", older)
	insertNewsLink(t, pool, "news_new", "match_1", "evi_2", "src_ge", "New news", "https://ge.globo.com/new", newer)

	s := store.NewNewsStore(pool)
	records, err := s.ListByMatch(ctx, "match_1")
	if err != nil {
		t.Fatalf("ListByMatch: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("len(records) = %d, want 2", len(records))
	}
	if records[0].Title != "New news" {
		t.Fatalf("records[0].Title = %q, want most recently published first", records[0].Title)
	}
	if records[0].SourceDisplayName != "ge.globo.com" {
		t.Fatalf("SourceDisplayName = %q, want %q", records[0].SourceDisplayName, "ge.globo.com")
	}
}

func TestNewsStoreListByMatchCapsAtMax(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	insertClub(t, pool, "club_flamengo", "flamengo", "Flamengo", "FLA", nil)
	insertClub(t, pool, "club_vasco", "vasco", "Vasco", "VAS", nil)
	insertCompetition(t, pool, "comp_brasileirao_2026", "brasileirao", "Brasileirao", 2026)
	kickoff := time.Date(2026, 8, 10, 21, 0, 0, 0, time.UTC)
	insertMatch(t, pool, "match_1", "comp_brasileirao_2026", "club_flamengo", "club_vasco", "flamengo-x-vasco", &kickoff, "published")
	insertSource(t, pool, "src_ge", "ge.globo.com")

	for i := 0; i < store.MaxNewsPerMatch+3; i++ {
		evidenceID := fmt.Sprintf("evi_%d", i)
		newsID := fmt.Sprintf("news_%d", i)
		insertEvidence(t, pool, evidenceID, "src_ge", "match_1", "news")
		publishedAt := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(i) * time.Hour)
		insertNewsLink(t, pool, newsID, "match_1", evidenceID, "src_ge", fmt.Sprintf("News %d", i), fmt.Sprintf("https://ge.globo.com/%d", i), publishedAt)
	}

	s := store.NewNewsStore(pool)
	records, err := s.ListByMatch(ctx, "match_1")
	if err != nil {
		t.Fatalf("ListByMatch: %v", err)
	}
	if len(records) != store.MaxNewsPerMatch {
		t.Fatalf("len(records) = %d, want %d (capped)", len(records), store.MaxNewsPerMatch)
	}
}

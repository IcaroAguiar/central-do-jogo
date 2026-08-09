package store_test

import (
	"context"
	"testing"

	"github.com/IcaroAguiar/central-do-jogo/internal/store"
)

func TestClubStoreGetBySlug(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	insertClub(t, pool, "club_flamengo", "flamengo", "Clube de Regatas do Flamengo", "Flamengo", []string{"Mengão", "Fla"})

	s := store.NewClubStore(pool)

	club, err := s.GetBySlug(ctx, "flamengo")
	if err != nil {
		t.Fatalf("GetBySlug: %v", err)
	}
	if club == nil {
		t.Fatal("expected club, got nil")
	}
	if club.Name != "Clube de Regatas do Flamengo" || club.ShortName != "Flamengo" {
		t.Fatalf("club = %+v, unexpected fields", club)
	}
	if len(club.Aliases) != 2 {
		t.Fatalf("Aliases = %v, want 2 entries", club.Aliases)
	}
	if club.CreatedAt.Location() != club.CreatedAt.UTC().Location() {
		t.Fatalf("CreatedAt should be normalized to UTC, got location %v", club.CreatedAt.Location())
	}
}

func TestClubStoreGetBySlugNotFound(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	s := store.NewClubStore(pool)

	club, err := s.GetBySlug(ctx, "does-not-exist")
	if err != nil {
		t.Fatalf("GetBySlug: %v", err)
	}
	if club != nil {
		t.Fatalf("expected nil for missing club, got %+v", club)
	}
}

func TestClubStoreSearchMatchesNameShortNameSlugAndAliases(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	insertClub(t, pool, "club_flamengo", "flamengo", "Clube de Regatas do Flamengo", "Flamengo", []string{"Mengão"})
	insertClub(t, pool, "club_corinthians", "corinthians", "Sport Club Corinthians Paulista", "Corinthians", []string{"Timão"})
	insertClub(t, pool, "club_palmeiras", "palmeiras", "Sociedade Esportiva Palmeiras", "Palmeiras", []string{})
	s := store.NewClubStore(pool)

	cases := []struct {
		name     string
		query    string
		wantSlug string
	}{
		{"by name substring", "regatas do flamengo", "flamengo"},
		{"by short name", "corinthians", "corinthians"},
		{"by slug", "palmeiras", "palmeiras"},
		{"by alias", "mengão", "flamengo"},
		{"case insensitive", "TIMÃO", "corinthians"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			clubs, err := s.Search(ctx, tc.query, 10)
			if err != nil {
				t.Fatalf("Search: %v", err)
			}
			found := false
			for _, c := range clubs {
				if c.Slug == tc.wantSlug {
					found = true
				}
			}
			if !found {
				t.Fatalf("Search(%q) = %+v, want a hit for slug %q", tc.query, clubs, tc.wantSlug)
			}
		})
	}
}

func TestClubStoreSearchRespectsLimit(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		slug := string(rune('a' + i))
		insertClub(t, pool, "club_"+slug, slug, "Clube "+slug, slug, nil)
	}
	s := store.NewClubStore(pool)

	clubs, err := s.Search(ctx, "clube", 3)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(clubs) != 3 {
		t.Fatalf("len(clubs) = %d, want 3", len(clubs))
	}
}

func TestClubStoreList(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	insertClub(t, pool, "club_b", "b", "B Clube", "B", nil)
	insertClub(t, pool, "club_a", "a", "A Clube", "A", nil)
	s := store.NewClubStore(pool)

	clubs, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(clubs) != 2 {
		t.Fatalf("len(clubs) = %d, want 2", len(clubs))
	}
	if clubs[0].Slug != "a" || clubs[1].Slug != "b" {
		t.Fatalf("clubs = %+v, want ordered by name", clubs)
	}
}

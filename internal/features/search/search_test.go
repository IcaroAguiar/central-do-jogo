package search

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
)

type fakeClubSearcher struct {
	clubs []domain.Club
	err   error
}

func (f *fakeClubSearcher) Search(ctx context.Context, query string, limit int) ([]domain.Club, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.clubs, nil
}

type fakeMatchSearcher struct {
	matches []domain.MatchRecord
	err     error
}

func (f *fakeMatchSearcher) Search(ctx context.Context, query string, limit int) ([]domain.MatchRecord, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.matches, nil
}

func TestServiceSearchMapsClubsAndMatches(t *testing.T) {
	t.Parallel()
	kickoff := time.Date(2026, 8, 10, 21, 0, 0, 0, time.UTC)
	clubs := &fakeClubSearcher{clubs: []domain.Club{{Slug: "flamengo", Name: "Flamengo", ShortName: "FLA"}}}
	matches := &fakeMatchSearcher{matches: []domain.MatchRecord{
		{
			Match:    domain.Match{Slug: "flamengo-x-vasco", Round: "R1", KickoffAt: &kickoff, KickoffState: domain.KickoffPublished},
			HomeClub: domain.ClubSummary{Slug: "flamengo", Name: "Flamengo", ShortName: "FLA"},
			AwayClub: domain.ClubSummary{Slug: "vasco", Name: "Vasco", ShortName: "VAS"},
		},
	}}

	svc := NewService(clubs, matches)
	resp, err := svc.Search(context.Background(), "fla")
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if resp.Query != "fla" {
		t.Fatalf("Query = %q, want fla", resp.Query)
	}
	if len(resp.Clubs) != 1 || resp.Clubs[0].Slug != "flamengo" {
		t.Fatalf("Clubs = %+v, want one flamengo hit", resp.Clubs)
	}
	if len(resp.Matches) != 1 || resp.Matches[0].Slug != "flamengo-x-vasco" {
		t.Fatalf("Matches = %+v, want one flamengo-x-vasco hit", resp.Matches)
	}
	if resp.Matches[0].KickoffAt == nil || !resp.Matches[0].KickoffAt.Equal(kickoff) {
		t.Fatalf("KickoffAt = %v, want %v", resp.Matches[0].KickoffAt, kickoff)
	}
}

func TestServiceSearchPropagatesClubStoreError(t *testing.T) {
	t.Parallel()
	svc := NewService(&fakeClubSearcher{err: errors.New("boom")}, &fakeMatchSearcher{})
	if _, err := svc.Search(context.Background(), "x"); err == nil {
		t.Fatal("expected error when club search fails")
	}
}

func TestServiceSearchPropagatesMatchStoreError(t *testing.T) {
	t.Parallel()
	svc := NewService(&fakeClubSearcher{}, &fakeMatchSearcher{err: errors.New("boom")})
	if _, err := svc.Search(context.Background(), "x"); err == nil {
		t.Fatal("expected error when match search fails")
	}
}

func TestServiceSearchReturnsEmptySlicesNotNil(t *testing.T) {
	t.Parallel()
	svc := NewService(&fakeClubSearcher{}, &fakeMatchSearcher{})
	resp, err := svc.Search(context.Background(), "nothing")
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if resp.Clubs == nil || resp.Matches == nil {
		t.Fatal("expected empty (non-nil) slices for no results, to serialize as [] not null")
	}
}

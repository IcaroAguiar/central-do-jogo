package clubs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/store"
)

type fakeClubReader struct {
	bySlug map[string]*domain.Club
	all    []domain.Club
	err    error
}

func (f *fakeClubReader) GetBySlug(ctx context.Context, slug string) (*domain.Club, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.bySlug[slug], nil
}

func (f *fakeClubReader) List(ctx context.Context) ([]domain.Club, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.all, nil
}

type fakeMatchLister struct {
	byClub map[domain.ID][]store.MatchRecord
	err    error
	// gotSeason/gotStart/gotEnd capture the last call's arguments for assertions.
	gotSeason int
	gotStart  *time.Time
	gotEnd    *time.Time
}

func (f *fakeMatchLister) ListByClub(ctx context.Context, clubID domain.ID, season int, start, end *time.Time) ([]store.MatchRecord, error) {
	f.gotSeason = season
	f.gotStart = start
	f.gotEnd = end
	if f.err != nil {
		return nil, f.err
	}
	return f.byClub[clubID], nil
}

func TestGetDetailReturnsNilWhenClubMissing(t *testing.T) {
	t.Parallel()
	svc := NewService(&fakeClubReader{bySlug: map[string]*domain.Club{}}, &fakeMatchLister{}, nil)
	detail, err := svc.GetDetail(context.Background(), "missing")
	if err != nil {
		t.Fatalf("GetDetail() error: %v", err)
	}
	if detail != nil {
		t.Fatalf("detail = %+v, want nil for missing club", detail)
	}
}

func TestGetDetailMapsClubFieldsAndDefaultsAliases(t *testing.T) {
	t.Parallel()
	club := &domain.Club{ID: "club_x", Slug: "x", Name: "Clube X", ShortName: "X"}
	svc := NewService(&fakeClubReader{bySlug: map[string]*domain.Club{"x": club}}, &fakeMatchLister{}, nil)

	detail, err := svc.GetDetail(context.Background(), "x")
	if err != nil {
		t.Fatalf("GetDetail() error: %v", err)
	}
	if detail == nil {
		t.Fatal("expected non-nil detail")
	}
	if detail.Name != "Clube X" || detail.ShortName != "X" {
		t.Fatalf("detail = %+v, unexpected fields", detail)
	}
	if detail.Aliases == nil {
		t.Fatal("Aliases should default to an empty slice, not nil, so it serializes as []")
	}
}

func TestGetDetailPropagatesStoreError(t *testing.T) {
	t.Parallel()
	svc := NewService(&fakeClubReader{err: errors.New("boom")}, &fakeMatchLister{}, nil)
	if _, err := svc.GetDetail(context.Background(), "x"); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetMatchesReturnsNilWhenClubMissing(t *testing.T) {
	t.Parallel()
	svc := NewService(&fakeClubReader{bySlug: map[string]*domain.Club{}}, &fakeMatchLister{}, nil)
	resp, err := svc.GetMatches(context.Background(), "missing", RangeWeek, 2026)
	if err != nil {
		t.Fatalf("GetMatches() error: %v", err)
	}
	if resp != nil {
		t.Fatalf("resp = %+v, want nil for missing club", resp)
	}
}

func TestGetMatchesPassesSeasonRangeAndMapsSummaries(t *testing.T) {
	t.Parallel()
	fixedNow := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	club := &domain.Club{ID: "club_flamengo", Slug: "flamengo", Name: "Flamengo"}
	kickoff := time.Date(2026, 3, 12, 20, 0, 0, 0, time.UTC)
	matches := &fakeMatchLister{byClub: map[domain.ID][]store.MatchRecord{
		"club_flamengo": {
			{
				Match:    domain.Match{Slug: "flamengo-x-vasco", Round: "R1", Venue: "Maracanã", KickoffAt: &kickoff, KickoffState: domain.KickoffPublished, BroadcastState: domain.AvailabilityAvailable, LineupState: domain.AvailabilityAvailable, NewsState: domain.AvailabilityAvailable},
				HomeClub: store.ClubSummary{Slug: "flamengo", Name: "Flamengo"},
				AwayClub: store.ClubSummary{Slug: "vasco", Name: "Vasco"},
			},
		},
	}}
	svc := NewService(&fakeClubReader{bySlug: map[string]*domain.Club{"flamengo": club}}, matches, func() time.Time { return fixedNow })

	resp, err := svc.GetMatches(context.Background(), "flamengo", RangeWeek, 2026)
	if err != nil {
		t.Fatalf("GetMatches() error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.Range != "week" || resp.Season != 2026 {
		t.Fatalf("resp = %+v, unexpected range/season", resp)
	}
	if len(resp.Matches) != 1 || resp.Matches[0].Slug != "flamengo-x-vasco" {
		t.Fatalf("Matches = %+v, unexpected content", resp.Matches)
	}
	if matches.gotSeason != 2026 {
		t.Fatalf("season passed to store = %d, want 2026", matches.gotSeason)
	}
	if matches.gotStart == nil || matches.gotEnd == nil {
		t.Fatal("week range should pass non-nil bounds to the store")
	}
}

func TestGetMatchesSeasonRangePassesNilBounds(t *testing.T) {
	t.Parallel()
	club := &domain.Club{ID: "club_flamengo", Slug: "flamengo", Name: "Flamengo"}
	matches := &fakeMatchLister{byClub: map[domain.ID][]store.MatchRecord{}}
	svc := NewService(&fakeClubReader{bySlug: map[string]*domain.Club{"flamengo": club}}, matches, nil)

	if _, err := svc.GetMatches(context.Background(), "flamengo", RangeSeason, 2026); err != nil {
		t.Fatalf("GetMatches() error: %v", err)
	}
	if matches.gotStart != nil || matches.gotEnd != nil {
		t.Fatalf("season range should pass nil bounds, got start=%v end=%v", matches.gotStart, matches.gotEnd)
	}
}

func TestListClubsMapsToListItems(t *testing.T) {
	t.Parallel()
	svc := NewService(&fakeClubReader{all: []domain.Club{{Slug: "a", Name: "A"}, {Slug: "b", Name: "B"}}}, &fakeMatchLister{}, nil)
	items, err := svc.ListClubs(context.Background())
	if err != nil {
		t.Fatalf("ListClubs() error: %v", err)
	}
	if len(items) != 2 || items[0].Slug != "a" || items[1].Slug != "b" {
		t.Fatalf("items = %+v, unexpected", items)
	}
}

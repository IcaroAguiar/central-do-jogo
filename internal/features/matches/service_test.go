package matches

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/store"
)

var errBoom = errors.New("boom")

type fakeMatchGetter struct {
	bySlug map[string]*store.MatchRecord
	err    error
}

func (f *fakeMatchGetter) GetBySlug(ctx context.Context, slug string) (*store.MatchRecord, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.bySlug[slug], nil
}

type fakeBroadcastLister struct {
	byMatch map[domain.ID][]store.BroadcastRecord
	err     error
}

func (f *fakeBroadcastLister) ListByMatch(ctx context.Context, matchID domain.ID) ([]store.BroadcastRecord, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.byMatch[matchID], nil
}

type fakeLineupLister struct {
	byMatch map[domain.ID][]domain.Lineup
	err     error
}

func (f *fakeLineupLister) ListByMatch(ctx context.Context, matchID domain.ID) ([]domain.Lineup, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.byMatch[matchID], nil
}

type fakeNewsLister struct {
	byMatch map[domain.ID][]store.NewsRecord
	err     error
}

func (f *fakeNewsLister) ListByMatch(ctx context.Context, matchID domain.ID) ([]store.NewsRecord, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.byMatch[matchID], nil
}

func TestGetDetailReturnsNilWhenMatchMissing(t *testing.T) {
	t.Parallel()
	svc := NewService(&fakeMatchGetter{bySlug: map[string]*store.MatchRecord{}}, &fakeBroadcastLister{}, &fakeLineupLister{}, &fakeNewsLister{})
	detail, err := svc.GetDetail(context.Background(), "missing")
	if err != nil {
		t.Fatalf("GetDetail() error: %v", err)
	}
	if detail != nil {
		t.Fatalf("detail = %+v, want nil", detail)
	}
}

func TestGetDetailAggregatesAllSurfaces(t *testing.T) {
	t.Parallel()
	attempt := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	rec := &store.MatchRecord{
		Match: domain.Match{
			ID: "match_1", Slug: "flamengo-x-vasco", Round: "R1", Venue: "Maracanã",
			KickoffState: domain.KickoffPublished, BroadcastState: domain.AvailabilityAvailable,
			LineupState: domain.AvailabilityAwaitingPublication, NewsState: domain.AvailabilityAvailable,
			LineupLastAttemptAt: &attempt,
		},
		HomeClub:    store.ClubSummary{Slug: "flamengo", Name: "Flamengo"},
		AwayClub:    store.ClubSummary{Slug: "vasco", Name: "Vasco"},
		Competition: store.CompetitionSummary{Slug: "brasileirao", Name: "Brasileirao", Season: 2026},
	}
	matches := &fakeMatchGetter{bySlug: map[string]*store.MatchRecord{"flamengo-x-vasco": rec}}
	broadcasts := &fakeBroadcastLister{byMatch: map[domain.ID][]store.BroadcastRecord{
		"match_1": {{Broadcast: domain.Broadcast{Channel: "TV Globo", Access: domain.AccessFree, Confidence: domain.ConfidenceHigh}, SourceDisplayName: "Fonte X"}},
	}}
	lineups := &fakeLineupLister{byMatch: map[domain.ID][]domain.Lineup{
		"match_1": {{Side: domain.LineupHome, Formation: "4-3-3", Players: []domain.LineupPlayer{{ShirtNumber: "9", Name: "Atacante", IsStarter: true}}}},
	}}
	news := &fakeNewsLister{byMatch: map[domain.ID][]store.NewsRecord{
		"match_1": {{NewsLink: domain.NewsLink{Title: "Notícia", URL: "https://example.com"}, SourceDisplayName: "Fonte Y"}},
	}}

	svc := NewService(matches, broadcasts, lineups, news)
	detail, err := svc.GetDetail(context.Background(), "flamengo-x-vasco")
	if err != nil {
		t.Fatalf("GetDetail() error: %v", err)
	}
	if detail == nil {
		t.Fatal("expected non-nil detail")
	}
	if len(detail.Broadcasts) != 1 || detail.Broadcasts[0].Channel != "TV Globo" || detail.Broadcasts[0].Source != "Fonte X" {
		t.Fatalf("Broadcasts = %+v, unexpected", detail.Broadcasts)
	}
	if len(detail.Lineups) != 1 || len(detail.Lineups[0].Players) != 1 || detail.Lineups[0].Players[0].Name != "Atacante" {
		t.Fatalf("Lineups = %+v, unexpected", detail.Lineups)
	}
	if len(detail.News) != 1 || detail.News[0].Source != "Fonte Y" {
		t.Fatalf("News = %+v, unexpected", detail.News)
	}
	if detail.LineupLastAttemptAt == nil || !detail.LineupLastAttemptAt.Equal(attempt) {
		t.Fatalf("LineupLastAttemptAt = %v, want %v", detail.LineupLastAttemptAt, attempt)
	}
	if detail.Competition.Season != 2026 {
		t.Fatalf("Competition.Season = %d, want 2026", detail.Competition.Season)
	}
}

func TestGetDetailPropagatesErrorsFromEachSurface(t *testing.T) {
	t.Parallel()
	rec := &store.MatchRecord{Match: domain.Match{ID: "match_1", Slug: "m"}}
	baseMatches := &fakeMatchGetter{bySlug: map[string]*store.MatchRecord{"m": rec}}

	t.Run("broadcasts", func(t *testing.T) {
		t.Parallel()
		svc := NewService(baseMatches, &fakeBroadcastLister{err: errors.New("boom")}, &fakeLineupLister{}, &fakeNewsLister{})
		if _, err := svc.GetDetail(context.Background(), "m"); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("lineups", func(t *testing.T) {
		t.Parallel()
		svc := NewService(baseMatches, &fakeBroadcastLister{}, &fakeLineupLister{err: errors.New("boom")}, &fakeNewsLister{})
		if _, err := svc.GetDetail(context.Background(), "m"); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("news", func(t *testing.T) {
		t.Parallel()
		svc := NewService(baseMatches, &fakeBroadcastLister{}, &fakeLineupLister{}, &fakeNewsLister{err: errors.New("boom")})
		if _, err := svc.GetDetail(context.Background(), "m"); err == nil {
			t.Fatal("expected error")
		}
	})
}

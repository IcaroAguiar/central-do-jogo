// Package search implements REQ-005 direct search over clubs and matches.
package search

import (
	"context"
	"fmt"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/store"
)

// MaxClubResults and MaxMatchResults bound the result set returned per query.
const (
	MaxClubResults  = 10
	MaxMatchResults = 10
	MinQueryLength  = 1
	MaxQueryLength  = 100
)

// ClubSearcher is the subset of store.ClubStore consumed by this package,
// declared here so tests can supply a fake without a database.
type ClubSearcher interface {
	Search(ctx context.Context, query string, limit int) ([]domain.Club, error)
}

// MatchSearcher is the subset of store.MatchStore consumed by this package.
type MatchSearcher interface {
	Search(ctx context.Context, query string, limit int) ([]store.MatchRecord, error)
}

// Service resolves search queries against the club and match read stores.
type Service struct {
	clubs   ClubSearcher
	matches MatchSearcher
}

// NewService creates a search service.
func NewService(clubs ClubSearcher, matches MatchSearcher) *Service {
	return &Service{clubs: clubs, matches: matches}
}

// ClubResult is a search hit describing a club.
type ClubResult struct {
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
}

// MatchResult is a search hit describing a match.
type MatchResult struct {
	Slug         string     `json:"slug"`
	Round        string     `json:"round"`
	HomeClub     ClubResult `json:"homeClub"`
	AwayClub     ClubResult `json:"awayClub"`
	KickoffAt    *time.Time `json:"kickoffAt"`
	KickoffState string     `json:"kickoffState"`
}

// Response is the JSON payload for a search request.
type Response struct {
	Query   string        `json:"query"`
	Clubs   []ClubResult  `json:"clubs"`
	Matches []MatchResult `json:"matches"`
}

// Search resolves a query into club and match hits. The query is expected to
// already be trimmed and length-validated by the caller (see handler.go).
func (s *Service) Search(ctx context.Context, query string) (Response, error) {
	clubs, err := s.clubs.Search(ctx, query, MaxClubResults)
	if err != nil {
		return Response{}, fmt.Errorf("search clubs: %w", err)
	}
	matches, err := s.matches.Search(ctx, query, MaxMatchResults)
	if err != nil {
		return Response{}, fmt.Errorf("search matches: %w", err)
	}

	resp := Response{
		Query:   query,
		Clubs:   make([]ClubResult, 0, len(clubs)),
		Matches: make([]MatchResult, 0, len(matches)),
	}
	for _, c := range clubs {
		resp.Clubs = append(resp.Clubs, ClubResult{Slug: c.Slug, Name: c.Name, ShortName: c.ShortName})
	}
	for _, m := range matches {
		resp.Matches = append(resp.Matches, MatchResult{
			Slug:         m.Slug,
			Round:        m.Round,
			HomeClub:     ClubResult{Slug: m.HomeClub.Slug, Name: m.HomeClub.Name, ShortName: m.HomeClub.ShortName},
			AwayClub:     ClubResult{Slug: m.AwayClub.Slug, Name: m.AwayClub.Name, ShortName: m.AwayClub.ShortName},
			KickoffAt:    m.KickoffAt,
			KickoffState: string(m.KickoffState),
		})
	}
	return resp, nil
}

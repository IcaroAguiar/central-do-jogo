// Package clubs implements the club detail and agenda journeys (REQ-002, REQ-004).
package clubs

import (
	"context"
	"fmt"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/IcaroAguiar/central-do-jogo/internal/httpapi"
)

// ClubReader is the club read port consumed by this package,
// declared here so tests can supply a fake without a database.
type ClubReader interface {
	GetBySlug(ctx context.Context, slug string) (*domain.Club, error)
	List(ctx context.Context) ([]domain.Club, error)
}

// MatchLister is the match read port consumed by this package.
type MatchLister interface {
	ListByClub(ctx context.Context, clubID domain.ID, season int, start, end *time.Time) ([]domain.MatchRecord, error)
}

// Service resolves club detail and agenda queries.
type Service struct {
	clubs   ClubReader
	matches MatchLister
	now     func() time.Time
}

// NewService creates a clubs service. now defaults to time.Now when nil.
func NewService(clubs ClubReader, matches MatchLister, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{clubs: clubs, matches: matches, now: now}
}

// Detail is the JSON payload for GET /api/v1/clubs/{slug}.
type Detail struct {
	Slug      string   `json:"slug"`
	Name      string   `json:"name"`
	ShortName string   `json:"shortName"`
	Aliases   []string `json:"aliases"`
}

// GetDetail returns club detail for slug, or nil if the club does not exist.
func (s *Service) GetDetail(ctx context.Context, slug string) (*Detail, error) {
	club, err := s.clubs.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("get club detail: %w", err)
	}
	if club == nil {
		return nil, nil
	}
	return detailFromDomain(club), nil
}

// ClubListItem is a minimal club reference used for indexable listings.
type ClubListItem struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// ListClubs returns all clubs ordered by name, for indexable listings such as
// the home page.
func (s *Service) ListClubs(ctx context.Context) ([]ClubListItem, error) {
	all, err := s.clubs.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list clubs: %w", err)
	}
	items := make([]ClubListItem, 0, len(all))
	for _, c := range all {
		items = append(items, ClubListItem{Slug: c.Slug, Name: c.Name})
	}
	return items, nil
}

func detailFromDomain(club *domain.Club) *Detail {
	aliases := club.Aliases
	if aliases == nil {
		aliases = []string{}
	}
	return &Detail{Slug: club.Slug, Name: club.Name, ShortName: club.ShortName, Aliases: aliases}
}

// MatchSummary is one match entry in a club's agenda.
type MatchSummary struct {
	Slug           string          `json:"slug"`
	Round          string          `json:"round"`
	Venue          string          `json:"venue"`
	HomeClub       httpapi.ClubRef `json:"homeClub"`
	AwayClub       httpapi.ClubRef `json:"awayClub"`
	KickoffAt      *time.Time      `json:"kickoffAt"`
	KickoffState   string          `json:"kickoffState"`
	BroadcastState string          `json:"broadcastState"`
	LineupState    string          `json:"lineupState"`
	NewsState      string          `json:"newsState"`
}

// MatchesResponse is the JSON payload for GET /api/v1/clubs/{slug}/matches.
type MatchesResponse struct {
	Range   string         `json:"range"`
	Season  int            `json:"season"`
	Matches []MatchSummary `json:"matches"`
}

// GetMatches returns a club's agenda for the given range and season, or nil
// if the club does not exist.
func (s *Service) GetMatches(ctx context.Context, slug string, rng Range, season int) (*MatchesResponse, error) {
	club, err := s.clubs.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("get club for matches: %w", err)
	}
	if club == nil {
		return nil, nil
	}

	start, end := bounds(rng, s.now())
	records, err := s.matches.ListByClub(ctx, club.ID, season, start, end)
	if err != nil {
		return nil, fmt.Errorf("list club matches: %w", err)
	}

	resp := &MatchesResponse{Range: string(rng), Season: season, Matches: make([]MatchSummary, 0, len(records))}
	for _, m := range records {
		resp.Matches = append(resp.Matches, MatchSummary{
			Slug:           m.Slug,
			Round:          m.Round,
			Venue:          m.Venue,
			HomeClub:       httpapi.ClubRefFromParts(m.HomeClub.Slug, m.HomeClub.Name, m.HomeClub.ShortName),
			AwayClub:       httpapi.ClubRefFromParts(m.AwayClub.Slug, m.AwayClub.Name, m.AwayClub.ShortName),
			KickoffAt:      m.KickoffAt,
			KickoffState:   string(m.KickoffState),
			BroadcastState: string(m.BroadcastState),
			LineupState:    string(m.LineupState),
			NewsState:      string(m.NewsState),
		})
	}
	return resp, nil
}

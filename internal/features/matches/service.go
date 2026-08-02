// Package matches implements the match detail journey: broadcasts,
// confidence, lineups, news, and REQ-010 gap states.
package matches

import (
	"context"
	"fmt"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/store"
)

// MatchGetter is the subset of store.MatchStore consumed by this package,
// declared here so tests can supply a fake without a database.
type MatchGetter interface {
	GetBySlug(ctx context.Context, slug string) (*store.MatchRecord, error)
}

// BroadcastLister is the subset of store.BroadcastStore consumed by this package.
type BroadcastLister interface {
	ListByMatch(ctx context.Context, matchID domain.ID) ([]store.BroadcastRecord, error)
}

// LineupLister is the subset of store.LineupStore consumed by this package.
type LineupLister interface {
	ListByMatch(ctx context.Context, matchID domain.ID) ([]domain.Lineup, error)
}

// NewsLister is the subset of store.NewsStore consumed by this package.
type NewsLister interface {
	ListByMatch(ctx context.Context, matchID domain.ID) ([]store.NewsRecord, error)
}

// Service resolves match detail queries.
type Service struct {
	matches    MatchGetter
	broadcasts BroadcastLister
	lineups    LineupLister
	news       NewsLister
}

// NewService creates a matches service.
func NewService(matches MatchGetter, broadcasts BroadcastLister, lineups LineupLister, news NewsLister) *Service {
	return &Service{matches: matches, broadcasts: broadcasts, lineups: lineups, news: news}
}

// ClubRef is a minimal club reference embedded in a match detail.
type ClubRef struct {
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
}

// CompetitionRef is a minimal competition reference embedded in a match detail.
type CompetitionRef struct {
	Slug   string `json:"slug"`
	Name   string `json:"name"`
	Season int    `json:"season"`
}

// BroadcastView is one known broadcast for a match (REQ-007).
type BroadcastView struct {
	Channel     string    `json:"channel"`
	Platform    string    `json:"platform"`
	Access      string    `json:"access"`
	Region      string    `json:"region"`
	OfficialURL string    `json:"officialUrl"`
	Confidence  string    `json:"confidence"`
	VerifiedAt  time.Time `json:"verifiedAt"`
	Source      string    `json:"source"`
}

// LineupPlayerView is one player entry in a lineup (REQ-008).
type LineupPlayerView struct {
	ShirtNumber string `json:"shirtNumber"`
	Name        string `json:"name"`
	IsStarter   bool   `json:"isStarter"`
}

// LineupView is one club's lineup for a match (REQ-008).
type LineupView struct {
	Side        string             `json:"side"`
	Formation   string             `json:"formation"`
	Coach       string             `json:"coach"`
	Players     []LineupPlayerView `json:"players"`
	Official    bool               `json:"official"`
	PublishedAt *time.Time         `json:"publishedAt"`
}

// NewsLinkView is one related news card (REQ-009).
type NewsLinkView struct {
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Source      string    `json:"source"`
	PublishedAt time.Time `json:"publishedAt"`
}

// Detail is the JSON payload for GET /api/v1/matches/{slug}.
type Detail struct {
	Slug           string         `json:"slug"`
	Round          string         `json:"round"`
	Venue          string         `json:"venue"`
	HomeClub       ClubRef        `json:"homeClub"`
	AwayClub       ClubRef        `json:"awayClub"`
	Competition    CompetitionRef `json:"competition"`
	KickoffAt      *time.Time     `json:"kickoffAt"`
	KickoffState   string         `json:"kickoffState"`
	BroadcastState string         `json:"broadcastState"`
	LineupState    string         `json:"lineupState"`
	NewsState      string         `json:"newsState"`
	// *LastAttemptAt surface the most recent refresh attempt even when it
	// found nothing, per REQ-010 ("a ultima tentativa fica visivel").
	BroadcastLastAttemptAt *time.Time      `json:"broadcastLastAttemptAt"`
	LineupLastAttemptAt    *time.Time      `json:"lineupLastAttemptAt"`
	NewsLastAttemptAt      *time.Time      `json:"newsLastAttemptAt"`
	Broadcasts             []BroadcastView `json:"broadcasts"`
	Lineups                []LineupView    `json:"lineups"`
	News                   []NewsLinkView  `json:"news"`
}

// GetDetail returns match detail for slug, or nil if the match does not exist.
func (s *Service) GetDetail(ctx context.Context, slug string) (*Detail, error) {
	rec, err := s.matches.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("get match: %w", err)
	}
	if rec == nil {
		return nil, nil
	}

	broadcasts, err := s.broadcasts.ListByMatch(ctx, rec.ID)
	if err != nil {
		return nil, fmt.Errorf("list broadcasts: %w", err)
	}
	lineups, err := s.lineups.ListByMatch(ctx, rec.ID)
	if err != nil {
		return nil, fmt.Errorf("list lineups: %w", err)
	}
	news, err := s.news.ListByMatch(ctx, rec.ID)
	if err != nil {
		return nil, fmt.Errorf("list news: %w", err)
	}

	return &Detail{
		Slug:                   rec.Slug,
		Round:                  rec.Round,
		Venue:                  rec.Venue,
		HomeClub:               ClubRef{Slug: rec.HomeClub.Slug, Name: rec.HomeClub.Name, ShortName: rec.HomeClub.ShortName},
		AwayClub:               ClubRef{Slug: rec.AwayClub.Slug, Name: rec.AwayClub.Name, ShortName: rec.AwayClub.ShortName},
		Competition:            CompetitionRef{Slug: rec.Competition.Slug, Name: rec.Competition.Name, Season: rec.Competition.Season},
		KickoffAt:              rec.KickoffAt,
		KickoffState:           string(rec.KickoffState),
		BroadcastState:         string(rec.BroadcastState),
		LineupState:            string(rec.LineupState),
		NewsState:              string(rec.NewsState),
		BroadcastLastAttemptAt: rec.BroadcastLastAttemptAt,
		LineupLastAttemptAt:    rec.LineupLastAttemptAt,
		NewsLastAttemptAt:      rec.NewsLastAttemptAt,
		Broadcasts:             broadcastViews(broadcasts),
		Lineups:                lineupViews(lineups),
		News:                   newsViews(news),
	}, nil
}

func broadcastViews(records []store.BroadcastRecord) []BroadcastView {
	views := make([]BroadcastView, 0, len(records))
	for _, b := range records {
		views = append(views, BroadcastView{
			Channel:     b.Channel,
			Platform:    b.Platform,
			Access:      string(b.Access),
			Region:      b.Region,
			OfficialURL: b.OfficialURL,
			Confidence:  string(b.Confidence),
			VerifiedAt:  b.VerifiedAt,
			Source:      b.SourceDisplayName,
		})
	}
	return views
}

func lineupViews(lineups []domain.Lineup) []LineupView {
	views := make([]LineupView, 0, len(lineups))
	for _, l := range lineups {
		players := make([]LineupPlayerView, 0, len(l.Players))
		for _, p := range l.Players {
			players = append(players, LineupPlayerView{ShirtNumber: p.ShirtNumber, Name: p.Name, IsStarter: p.IsStarter})
		}
		views = append(views, LineupView{
			Side:        string(l.Side),
			Formation:   l.Formation,
			Coach:       l.Coach,
			Players:     players,
			Official:    l.Official,
			PublishedAt: l.PublishedAt,
		})
	}
	return views
}

func newsViews(records []store.NewsRecord) []NewsLinkView {
	views := make([]NewsLinkView, 0, len(records))
	for _, n := range records {
		views = append(views, NewsLinkView{
			Title:       n.Title,
			URL:         n.URL,
			Source:      n.SourceDisplayName,
			PublishedAt: n.PublishedAt,
		})
	}
	return views
}

package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MatchStore provides read access to matches.
type MatchStore struct {
	pool *pgxpool.Pool
}

// NewMatchStore creates a match store backed by the provided pgx pool.
func NewMatchStore(pool *pgxpool.Pool) *MatchStore {
	return &MatchStore{pool: pool}
}

const matchSelectColumns = `
	m.id, m.competition_id, m.home_club_id, m.away_club_id, m.slug, m.round, m.venue,
	m.kickoff_at, m.kickoff_state, m.broadcast_state, m.lineup_state, m.news_state,
	m.broadcast_last_attempt_at, m.lineup_last_attempt_at, m.news_last_attempt_at,
	m.created_at, m.updated_at,
	hc.id, hc.slug, hc.name, hc.short_name,
	ac.id, ac.slug, ac.name, ac.short_name,
	c.id, c.slug, c.name, c.season
`

const matchFromJoin = `
	FROM matches m
	JOIN clubs hc ON hc.id = m.home_club_id
	JOIN clubs ac ON ac.id = m.away_club_id
	JOIN competitions c ON c.id = m.competition_id
`

func scanMatchRecord(row pgx.Row) (*domain.MatchRecord, error) {
	var rec domain.MatchRecord
	var id, competitionID, homeClubID, awayClubID string
	var homeID, homeSlug, homeName, homeShort string
	var awayID, awaySlug, awayName, awayShort string
	var compID, compSlug, compName string
	if err := row.Scan(
		&id, &competitionID, &homeClubID, &awayClubID, &rec.Slug, &rec.Round, &rec.Venue,
		&rec.KickoffAt, &rec.KickoffState, &rec.BroadcastState, &rec.LineupState, &rec.NewsState,
		&rec.BroadcastLastAttemptAt, &rec.LineupLastAttemptAt, &rec.NewsLastAttemptAt,
		&rec.CreatedAt, &rec.UpdatedAt,
		&homeID, &homeSlug, &homeName, &homeShort,
		&awayID, &awaySlug, &awayName, &awayShort,
		&compID, &compSlug, &compName, &rec.Competition.Season,
	); err != nil {
		return nil, err
	}
	rec.ID = domain.ID(id)
	rec.CompetitionID = domain.ID(competitionID)
	rec.HomeClubID = domain.ID(homeClubID)
	rec.AwayClubID = domain.ID(awayClubID)
	rec.HomeClub = domain.ClubSummary{ID: domain.ID(homeID), Slug: homeSlug, Name: homeName, ShortName: homeShort}
	rec.AwayClub = domain.ClubSummary{ID: domain.ID(awayID), Slug: awaySlug, Name: awayName, ShortName: awayShort}
	rec.Competition.ID = domain.ID(compID)
	rec.Competition.Slug = compSlug
	rec.Competition.Name = compName

	rec.KickoffAt = utcPtr(rec.KickoffAt)
	rec.BroadcastLastAttemptAt = utcPtr(rec.BroadcastLastAttemptAt)
	rec.LineupLastAttemptAt = utcPtr(rec.LineupLastAttemptAt)
	rec.NewsLastAttemptAt = utcPtr(rec.NewsLastAttemptAt)
	rec.CreatedAt = utc(rec.CreatedAt)
	rec.UpdatedAt = utc(rec.UpdatedAt)
	return &rec, nil
}

// GetBySlug returns the match with the given slug joined with club and
// competition summaries, or nil if none exists.
func (s *MatchStore) GetBySlug(ctx context.Context, slug string) (*domain.MatchRecord, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+matchSelectColumns+matchFromJoin+` WHERE m.slug = $1`, slug)
	rec, err := scanMatchRecord(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get match by slug %q: %w", slug, err)
	}
	return rec, nil
}

// Search returns matches whose home/away club name, short name, round, or
// slug match the query (case-insensitive substring), ordered by kickoff time
// (soonest first, unscheduled last), limited to limit rows.
func (s *MatchStore) Search(ctx context.Context, query string, limit int) ([]domain.MatchRecord, error) {
	if limit <= 0 {
		limit = 10
	}
	pattern := "%" + query + "%"
	rows, err := s.pool.Query(ctx, `
		SELECT `+matchSelectColumns+matchFromJoin+`
		WHERE hc.name ILIKE $1 OR ac.name ILIKE $1
		   OR hc.short_name ILIKE $1 OR ac.short_name ILIKE $1
		   OR m.round ILIKE $1 OR m.slug ILIKE $1
		ORDER BY m.kickoff_at ASC NULLS LAST, m.slug
		LIMIT $2
	`, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("search matches: %w", err)
	}
	defer rows.Close()
	return collectMatchRecords(rows)
}

// ListByClub returns matches for a club, optionally bounded by [start, end)
// UTC kickoff range, filtered to the given competition season. When start
// and end are both nil, all matches in the season are returned regardless of
// whether the kickoff time is known; otherwise matches with an unknown
// kickoff time are excluded since they cannot be placed in the range.
func (s *MatchStore) ListByClub(ctx context.Context, clubID domain.ID, season int, start, end *time.Time) ([]domain.MatchRecord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT `+matchSelectColumns+matchFromJoin+`
		WHERE (m.home_club_id = $1 OR m.away_club_id = $1)
		  AND c.season = $2
		  AND (
		        ($3::timestamptz IS NULL AND $4::timestamptz IS NULL)
		        OR (m.kickoff_at IS NOT NULL AND m.kickoff_at >= $3 AND m.kickoff_at < $4)
		      )
		ORDER BY m.kickoff_at ASC NULLS LAST, m.slug
	`, string(clubID), season, start, end)
	if err != nil {
		return nil, fmt.Errorf("list matches by club: %w", err)
	}
	defer rows.Close()
	return collectMatchRecords(rows)
}

func collectMatchRecords(rows pgx.Rows) ([]domain.MatchRecord, error) {
	var records []domain.MatchRecord
	for rows.Next() {
		rec, err := scanMatchRecord(rows)
		if err != nil {
			return nil, fmt.Errorf("scan match: %w", err)
		}
		records = append(records, *rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("match rows: %w", err)
	}
	return records, nil
}

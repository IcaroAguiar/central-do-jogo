package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Summary reports how many rows of each kind the seed touched.
type Summary struct {
	Clubs        int
	Competitions int
	Matches      int
	Broadcasts   int
	Lineups      int
	News         int
}

// attemptedAgo is how far in the past a "found nothing" refresh attempt is
// recorded, for surfaces seeded as awaiting_publication or not_found.
const attemptedAgo = 1 * time.Hour

// Run upserts the deterministic Serie A 2026 demo dataset. now anchors
// relative kickoff offsets so seeded matches always look "upcoming"
// regardless of when the seed runs.
func Run(ctx context.Context, pool *pgxpool.Pool, now time.Time) (Summary, error) {
	var summary Summary

	if err := upsertSource(ctx, pool); err != nil {
		return summary, err
	}

	competitionID, err := upsertCompetition(ctx, pool)
	if err != nil {
		return summary, err
	}
	summary.Competitions = 1

	clubIDs := make(map[string]string, len(serieAClubs))
	for _, c := range serieAClubs {
		id, err := upsertClub(ctx, pool, c)
		if err != nil {
			return summary, err
		}
		clubIDs[c.Slug] = id
		if err := upsertSeasonClub(ctx, pool, seedSeason, id); err != nil {
			return summary, err
		}
		summary.Clubs++
	}

	for _, m := range seedMatches {
		homeID, ok := clubIDs[m.HomeSlug]
		if !ok {
			return summary, fmt.Errorf("match %s: unknown home club slug %q", m.Slug, m.HomeSlug)
		}
		awayID, ok := clubIDs[m.AwaySlug]
		if !ok {
			return summary, fmt.Errorf("match %s: unknown away club slug %q", m.Slug, m.AwaySlug)
		}

		matchID, err := upsertMatch(ctx, pool, m, competitionID, homeID, awayID, now)
		if err != nil {
			return summary, err
		}
		summary.Matches++

		for i, b := range m.Broadcasts {
			evidenceID := fmt.Sprintf("evd_%s_broadcast_%d", m.Slug, i+1)
			if err := upsertEvidence(ctx, pool, evidenceID, "schedule", matchID, now); err != nil {
				return summary, err
			}
			if err := upsertBroadcast(ctx, pool, fmt.Sprintf("bcast_%s_%d", m.Slug, i+1), matchID, evidenceID, b, now); err != nil {
				return summary, err
			}
			summary.Broadcasts++
		}

		for _, l := range m.Lineups {
			suffix := l.Suffix
			if suffix == "" {
				suffix = "1"
			}
			evidenceID := fmt.Sprintf("evd_%s_lineup_%s_%s", m.Slug, l.Side, suffix)
			if err := upsertEvidence(ctx, pool, evidenceID, "lineup", matchID, now); err != nil {
				return summary, err
			}
			lineupID := fmt.Sprintf("lineup_%s_%s_%s", m.Slug, l.Side, suffix)
			if err := upsertLineup(ctx, pool, lineupID, matchID, homeAwayClubID(l.Side, homeID, awayID), evidenceID, l, now); err != nil {
				return summary, err
			}
			summary.Lineups++
		}

		for i, n := range m.News {
			evidenceID := fmt.Sprintf("evd_%s_news_%d", m.Slug, i+1)
			if err := upsertEvidence(ctx, pool, evidenceID, "news", matchID, now); err != nil {
				return summary, err
			}
			newsID := fmt.Sprintf("news_%s_%d", m.Slug, i+1)
			publishedAt := kickoffTime(m.KickoffOffset, now, -time.Duration(n.HoursBefore)*time.Hour)
			if err := upsertNews(ctx, pool, newsID, matchID, evidenceID, n, publishedAt); err != nil {
				return summary, err
			}
			summary.News++
		}
	}

	return summary, nil
}

func homeAwayClubID(side domain.LineupSide, homeID, awayID string) string {
	if side == domain.LineupHome {
		return homeID
	}
	return awayID
}

func upsertSource(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO sources (id, display_name, home_url)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET display_name = EXCLUDED.display_name, home_url = EXCLUDED.home_url
	`, seedSourceID, seedSourceDisplayName, seedSourceHomeURL)
	if err != nil {
		return fmt.Errorf("upsert seed source: %w", err)
	}
	return nil
}

func upsertCompetition(ctx context.Context, pool *pgxpool.Pool) (string, error) {
	id := fmt.Sprintf("comp_%s_%d", seedCompetitionSlug, seedSeason)
	_, err := pool.Exec(ctx, `
		INSERT INTO competitions (id, slug, name, season)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, updated_at = now()
	`, id, seedCompetitionSlug, seedCompetitionName, seedSeason)
	if err != nil {
		return "", fmt.Errorf("upsert competition: %w", err)
	}
	return id, nil
}

func upsertClub(ctx context.Context, pool *pgxpool.Pool, c clubSeed) (string, error) {
	id := "club_" + c.Slug
	aliases := c.Aliases
	if aliases == nil {
		aliases = []string{}
	}
	_, err := pool.Exec(ctx, `
		INSERT INTO clubs (id, slug, name, short_name, aliases)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name, short_name = EXCLUDED.short_name, aliases = EXCLUDED.aliases, updated_at = now()
	`, id, c.Slug, c.Name, c.ShortName, aliases)
	if err != nil {
		return "", fmt.Errorf("upsert club %s: %w", c.Slug, err)
	}
	return id, nil
}

func upsertSeasonClub(ctx context.Context, pool *pgxpool.Pool, season int, clubID string) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO season_clubs (season, club_id)
		VALUES ($1, $2)
		ON CONFLICT (season, club_id) DO NOTHING
	`, season, clubID)
	if err != nil {
		return fmt.Errorf("upsert season_clubs for %s: %w", clubID, err)
	}
	return nil
}

// kickoffTime resolves a match's kickoff offset (or nil for indefinite) plus
// an extra delta (e.g. "3 hours before kickoff" for a news timestamp).
func kickoffTime(offset *time.Duration, now time.Time, delta time.Duration) time.Time {
	if offset == nil {
		return now.Add(delta)
	}
	return now.Add(*offset).Add(delta)
}

func upsertMatch(ctx context.Context, pool *pgxpool.Pool, m matchSeed, competitionID, homeID, awayID string, now time.Time) (string, error) {
	id := "match_" + m.Slug

	var kickoffAt *time.Time
	if m.KickoffOffset != nil {
		t := now.Add(*m.KickoffOffset)
		kickoffAt = &t
	}

	var broadcastAttempt, lineupAttempt, newsAttempt *time.Time
	if m.BroadcastAttempted {
		t := now.Add(-attemptedAgo)
		broadcastAttempt = &t
	}
	if m.LineupAttempted {
		t := now.Add(-attemptedAgo)
		lineupAttempt = &t
	}
	if m.NewsAttempted {
		t := now.Add(-attemptedAgo)
		newsAttempt = &t
	}

	_, err := pool.Exec(ctx, `
		INSERT INTO matches (
			id, competition_id, home_club_id, away_club_id, slug, round, venue,
			kickoff_at, kickoff_state, broadcast_state, lineup_state, news_state,
			broadcast_last_attempt_at, lineup_last_attempt_at, news_last_attempt_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (id) DO UPDATE SET
			round = EXCLUDED.round, venue = EXCLUDED.venue,
			kickoff_at = EXCLUDED.kickoff_at, kickoff_state = EXCLUDED.kickoff_state,
			broadcast_state = EXCLUDED.broadcast_state, lineup_state = EXCLUDED.lineup_state,
			news_state = EXCLUDED.news_state,
			broadcast_last_attempt_at = EXCLUDED.broadcast_last_attempt_at,
			lineup_last_attempt_at = EXCLUDED.lineup_last_attempt_at,
			news_last_attempt_at = EXCLUDED.news_last_attempt_at,
			updated_at = now()
	`, id, competitionID, homeID, awayID, m.Slug, m.Round, m.Venue,
		kickoffAt, string(m.KickoffState), string(m.BroadcastState), string(m.LineupState), string(m.NewsState),
		broadcastAttempt, lineupAttempt, newsAttempt,
	)
	if err != nil {
		return "", fmt.Errorf("upsert match %s: %w", m.Slug, err)
	}
	return id, nil
}

func upsertEvidence(ctx context.Context, pool *pgxpool.Pool, id, dataType, matchID string, now time.Time) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO evidence (id, source_id, match_id, data_type, observed_at, fetched_at, parser_version, run_id, content_hash, raw_ref)
		VALUES ($1, $2, $3, $4, $5, $5, 'seed-v1', 'seed-run', $1, '')
		ON CONFLICT (id) DO NOTHING
	`, id, seedSourceID, matchID, dataType, now)
	if err != nil {
		return fmt.Errorf("upsert evidence %s: %w", id, err)
	}
	return nil
}

func upsertBroadcast(ctx context.Context, pool *pgxpool.Pool, id, matchID, evidenceID string, b broadcastSeed, now time.Time) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO broadcasts (id, match_id, evidence_id, channel, platform, access, region, official_url, confidence, verified_at, availability)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'available')
		ON CONFLICT (id) DO UPDATE SET
			channel = EXCLUDED.channel, platform = EXCLUDED.platform, access = EXCLUDED.access,
			region = EXCLUDED.region, official_url = EXCLUDED.official_url, confidence = EXCLUDED.confidence,
			verified_at = EXCLUDED.verified_at, updated_at = now()
	`, id, matchID, evidenceID, b.Channel, b.Platform, string(b.Access), b.Region, b.OfficialURL, string(b.Confidence), now)
	if err != nil {
		return fmt.Errorf("upsert broadcast %s: %w", id, err)
	}
	return nil
}

func upsertLineup(ctx context.Context, pool *pgxpool.Pool, id, matchID, clubID, evidenceID string, l lineupSeed, now time.Time) error {
	players := make([]lineupPlayerSeed, len(l.Players))
	copy(players, l.Players)
	playersJSON, err := json.Marshal(players)
	if err != nil {
		return fmt.Errorf("marshal players for lineup %s: %w", id, err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO lineups (id, match_id, club_id, evidence_id, side, formation, coach, players, official, availability, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'available', $10)
		ON CONFLICT (id) DO UPDATE SET
			formation = EXCLUDED.formation, coach = EXCLUDED.coach, players = EXCLUDED.players,
			official = EXCLUDED.official, published_at = EXCLUDED.published_at, updated_at = now()
	`, id, matchID, clubID, evidenceID, string(l.Side), l.Formation, l.Coach, playersJSON, l.Official, now)
	if err != nil {
		return fmt.Errorf("upsert lineup %s: %w", id, err)
	}
	return nil
}

func upsertNews(ctx context.Context, pool *pgxpool.Pool, id, matchID, evidenceID string, n newsSeed, publishedAt time.Time) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO news_links (id, match_id, evidence_id, source_id, title, url, published_at, availability)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'available')
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title, url = EXCLUDED.url, published_at = EXCLUDED.published_at, updated_at = now()
	`, id, matchID, evidenceID, seedSourceID, n.Title, n.URL, publishedAt)
	if err != nil {
		return fmt.Errorf("upsert news %s: %w", id, err)
	}
	return nil
}

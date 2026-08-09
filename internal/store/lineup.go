package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LineupStore provides read access to match lineups.
type LineupStore struct {
	pool *pgxpool.Pool
}

// NewLineupStore creates a lineup store backed by the provided pgx pool.
func NewLineupStore(pool *pgxpool.Pool) *LineupStore {
	return &LineupStore{pool: pool}
}

// ListByMatch returns lineups for a match, home side first.
func (s *LineupStore) ListByMatch(ctx context.Context, matchID domain.ID) ([]domain.Lineup, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, match_id, club_id, evidence_id, side, formation, coach, players,
		       official, availability, published_at, created_at, updated_at
		FROM lineups
		WHERE match_id = $1
		ORDER BY CASE side WHEN 'home' THEN 0 ELSE 1 END
	`, string(matchID))
	if err != nil {
		return nil, fmt.Errorf("list lineups by match: %w", err)
	}
	defer rows.Close()

	var lineups []domain.Lineup
	for rows.Next() {
		var l domain.Lineup
		var id, matchIDCol, clubID, evidenceID string
		var playersJSON []byte
		if err := rows.Scan(
			&id, &matchIDCol, &clubID, &evidenceID, &l.Side, &l.Formation, &l.Coach, &playersJSON,
			&l.Official, &l.Availability, &l.PublishedAt, &l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan lineup: %w", err)
		}
		l.ID = domain.ID(id)
		l.MatchID = domain.ID(matchIDCol)
		l.ClubID = domain.ID(clubID)
		l.EvidenceID = domain.ID(evidenceID)
		if len(playersJSON) > 0 {
			if err := json.Unmarshal(playersJSON, &l.Players); err != nil {
				return nil, fmt.Errorf("unmarshal lineup players: %w", err)
			}
		}
		l.PublishedAt = utcPtr(l.PublishedAt)
		l.CreatedAt = utc(l.CreatedAt)
		l.UpdatedAt = utc(l.UpdatedAt)
		lineups = append(lineups, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("lineup rows: %w", err)
	}
	return lineups, nil
}

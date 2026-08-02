package store

import (
	"context"
	"fmt"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BroadcastRecord is a broadcast joined with its evidence source display name.
type BroadcastRecord struct {
	domain.Broadcast
	SourceDisplayName string
}

// BroadcastStore provides read access to match broadcasts.
type BroadcastStore struct {
	pool *pgxpool.Pool
}

// NewBroadcastStore creates a broadcast store backed by the provided pgx pool.
func NewBroadcastStore(pool *pgxpool.Pool) *BroadcastStore {
	return &BroadcastStore{pool: pool}
}

// ListByMatch returns all broadcasts for a match, ordered by confidence (high
// first) then channel, with the source display name resolved via evidence.
func (s *BroadcastStore) ListByMatch(ctx context.Context, matchID domain.ID) ([]BroadcastRecord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			b.id, b.match_id, b.evidence_id, b.channel, b.platform, b.access, b.region,
			b.official_url, b.confidence, b.verified_at, b.availability, b.created_at, b.updated_at,
			src.display_name
		FROM broadcasts b
		JOIN evidence e ON e.id = b.evidence_id
		JOIN sources src ON src.id = e.source_id
		WHERE b.match_id = $1
		ORDER BY
			CASE b.confidence WHEN 'high' THEN 0 WHEN 'medium' THEN 1 ELSE 2 END,
			b.channel
	`, string(matchID))
	if err != nil {
		return nil, fmt.Errorf("list broadcasts by match: %w", err)
	}
	defer rows.Close()

	var records []BroadcastRecord
	for rows.Next() {
		var rec BroadcastRecord
		var id, matchIDCol, evidenceID string
		if err := rows.Scan(
			&id, &matchIDCol, &evidenceID, &rec.Channel, &rec.Platform, &rec.Access, &rec.Region,
			&rec.OfficialURL, &rec.Confidence, &rec.VerifiedAt, &rec.Availability, &rec.CreatedAt, &rec.UpdatedAt,
			&rec.SourceDisplayName,
		); err != nil {
			return nil, fmt.Errorf("scan broadcast: %w", err)
		}
		rec.ID = domain.ID(id)
		rec.MatchID = domain.ID(matchIDCol)
		rec.EvidenceID = domain.ID(evidenceID)
		rec.VerifiedAt = utc(rec.VerifiedAt)
		rec.CreatedAt = utc(rec.CreatedAt)
		rec.UpdatedAt = utc(rec.UpdatedAt)
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("broadcast rows: %w", err)
	}
	return records, nil
}

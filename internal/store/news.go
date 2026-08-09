package store

import (
	"context"
	"fmt"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewsStore provides read access to match news links.
type NewsStore struct {
	pool *pgxpool.Pool
}

// NewNewsStore creates a news store backed by the provided pgx pool.
func NewNewsStore(pool *pgxpool.Pool) *NewsStore {
	return &NewsStore{pool: pool}
}

// MaxNewsPerMatch caps the number of news links surfaced per match (REQ-009).
const MaxNewsPerMatch = 5

// ListByMatch returns up to MaxNewsPerMatch news links for a match, most
// recently published first.
func (s *NewsStore) ListByMatch(ctx context.Context, matchID domain.ID) ([]domain.NewsRecord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT n.id, n.match_id, n.evidence_id, n.source_id, n.title, n.url,
		       n.published_at, n.availability, n.created_at, n.updated_at, src.display_name
		FROM news_links n
		JOIN sources src ON src.id = n.source_id
		WHERE n.match_id = $1
		ORDER BY n.published_at DESC
		LIMIT $2
	`, string(matchID), MaxNewsPerMatch)
	if err != nil {
		return nil, fmt.Errorf("list news by match: %w", err)
	}
	defer rows.Close()

	var records []domain.NewsRecord
	for rows.Next() {
		var rec domain.NewsRecord
		var id, matchIDCol, evidenceID, sourceID string
		if err := rows.Scan(
			&id, &matchIDCol, &evidenceID, &sourceID, &rec.Title, &rec.URL,
			&rec.PublishedAt, &rec.Availability, &rec.CreatedAt, &rec.UpdatedAt, &rec.SourceDisplayName,
		); err != nil {
			return nil, fmt.Errorf("scan news link: %w", err)
		}
		rec.ID = domain.ID(id)
		rec.MatchID = domain.ID(matchIDCol)
		rec.EvidenceID = domain.ID(evidenceID)
		rec.SourceID = domain.ID(sourceID)
		rec.PublishedAt = utc(rec.PublishedAt)
		rec.CreatedAt = utc(rec.CreatedAt)
		rec.UpdatedAt = utc(rec.UpdatedAt)
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("news rows: %w", err)
	}
	return records, nil
}

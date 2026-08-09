package store

import (
	"context"
	"fmt"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ReportStore persists anonymous visitor reports.
type ReportStore struct {
	pool *pgxpool.Pool
}

// NewReportStore creates a report store backed by the provided pool.
func NewReportStore(pool *pgxpool.Pool) *ReportStore {
	return &ReportStore{pool: pool}
}

// Insert creates a new open report.
func (s *ReportStore) Insert(ctx context.Context, report domain.Report) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO reports (
			id, context_type, context_slug, message, status, ip_hash, user_agent, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, report.ID.String(), report.ContextType, report.ContextSlug, report.Message,
		report.Status, report.IPHash, report.UserAgent, report.CreatedAt.UTC())
	if err != nil {
		return fmt.Errorf("insert report: %w", err)
	}
	return nil
}

// ListOpen returns open reports newest first.
func (s *ReportStore) ListOpen(ctx context.Context, limit int) ([]domain.Report, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, context_type, context_slug, message, status, ip_hash, user_agent, created_at, reviewed_at
		FROM reports
		WHERE status = 'open'
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("list open reports: %w", err)
	}
	defer rows.Close()

	var out []domain.Report
	for rows.Next() {
		var r domain.Report
		var id string
		if err := rows.Scan(
			&id, &r.ContextType, &r.ContextSlug, &r.Message, &r.Status, &r.IPHash, &r.UserAgent,
			&r.CreatedAt, &r.ReviewedAt,
		); err != nil {
			return nil, fmt.Errorf("scan report: %w", err)
		}
		r.ID = domain.ID(id)
		r.CreatedAt = utc(r.CreatedAt)
		if r.ReviewedAt != nil {
			t := utc(*r.ReviewedAt)
			r.ReviewedAt = &t
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("report rows: %w", err)
	}
	return out, nil
}

// MarkReviewed sets status for a report without mutating match data.
func (s *ReportStore) MarkReviewed(ctx context.Context, id domain.ID, status string, now time.Time) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE reports
		SET status = $2, reviewed_at = $3
		WHERE id = $1
	`, id.String(), status, now.UTC())
	if err != nil {
		return fmt.Errorf("mark report reviewed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("mark report reviewed: not found")
	}
	return nil
}

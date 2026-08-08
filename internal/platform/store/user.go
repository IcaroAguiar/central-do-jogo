package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserStore provides persistence for OAuth users and server-side sessions.
type UserStore struct {
	pool *pgxpool.Pool
}

// NewUserStore creates a user/session store backed by the provided pgx pool.
func NewUserStore(pool *pgxpool.Pool) *UserStore {
	return &UserStore{pool: pool}
}

const userColumns = `id, provider, provider_subject, email, display_name, role, created_at, updated_at, last_login_at`

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	var id, provider, subject, email, displayName, role string
	var lastLogin *time.Time
	if err := row.Scan(&id, &provider, &subject, &email, &displayName, &role, &u.CreatedAt, &u.UpdatedAt, &lastLogin); err != nil {
		return nil, err
	}
	u.ID = domain.ID(id)
	u.Provider = provider
	u.ProviderSubject = subject
	u.Email = email
	u.DisplayName = displayName
	u.Role = domain.UserRole(role)
	u.CreatedAt = utc(u.CreatedAt)
	u.UpdatedAt = utc(u.UpdatedAt)
	if lastLogin != nil {
		t := utc(*lastLogin)
		u.LastLoginAt = &t
	}
	return &u, nil
}

// GetByID returns the user with the given id, or nil if none exists.
func (s *UserStore) GetByID(ctx context.Context, id domain.ID) (*domain.User, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id.String())
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

// GetByEmail returns the user with the given email (case-insensitive), or nil.
func (s *UserStore) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, nil
	}
	row := s.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE lower(email) = lower($1)`, email)
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

// UpsertByProviderSubject inserts or updates a user keyed by provider+subject.
// role is applied on every login so allowlist membership stays authoritative.
func (s *UserStore) UpsertByProviderSubject(ctx context.Context, user domain.User, now time.Time) (*domain.User, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO users (
			id, provider, provider_subject, email, display_name, role,
			created_at, updated_at, last_login_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $7, $7)
		ON CONFLICT (provider, provider_subject) DO UPDATE SET
			email = EXCLUDED.email,
			display_name = EXCLUDED.display_name,
			role = EXCLUDED.role,
			updated_at = EXCLUDED.updated_at,
			last_login_at = EXCLUDED.last_login_at
		RETURNING `+userColumns,
		user.ID.String(), user.Provider, user.ProviderSubject, user.Email,
		user.DisplayName, string(user.Role), now.UTC(),
	)
	out, err := scanUser(row)
	if err != nil {
		return nil, fmt.Errorf("upsert user: %w", err)
	}
	return out, nil
}

// CreateSession inserts a new session row.
func (s *UserStore) CreateSession(ctx context.Context, sess domain.Session) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO sessions (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, sess.ID.String(), sess.UserID.String(), sess.TokenHash, sess.ExpiresAt.UTC(), sess.CreatedAt.UTC())
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

// GetSessionByTokenHash returns the non-revoked, unexpired session for hash.
func (s *UserStore) GetSessionByTokenHash(ctx context.Context, tokenHash string, now time.Time) (*domain.Session, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, created_at, revoked_at
		FROM sessions
		WHERE token_hash = $1
		  AND revoked_at IS NULL
		  AND expires_at > $2
	`, tokenHash, now.UTC())
	var sess domain.Session
	var id, userID string
	var revoked *time.Time
	if err := row.Scan(&id, &userID, &sess.TokenHash, &sess.ExpiresAt, &sess.CreatedAt, &revoked); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get session by token hash: %w", err)
	}
	sess.ID = domain.ID(id)
	sess.UserID = domain.ID(userID)
	sess.ExpiresAt = utc(sess.ExpiresAt)
	sess.CreatedAt = utc(sess.CreatedAt)
	if revoked != nil {
		t := utc(*revoked)
		sess.RevokedAt = &t
	}
	return &sess, nil
}

// RevokeSession marks a session revoked by token hash. Missing rows are ignored.
func (s *UserStore) RevokeSession(ctx context.Context, tokenHash string, now time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE sessions
		SET revoked_at = $2
		WHERE token_hash = $1 AND revoked_at IS NULL
	`, tokenHash, now.UTC())
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

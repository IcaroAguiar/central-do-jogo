package store_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/platform/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

// openTestPool skips the test when DATABASE_URL is unset, mirroring the
// pattern used by internal/jobs/jobs_test.go so CI never depends on a live
// database unless one is explicitly provided.
func openTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := database.OpenPool(ctx, url)
	if err != nil {
		t.Fatalf("OpenPool: %v", err)
	}
	if err := database.Migrate(ctx, pool); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	// Clean slate for every test, respecting FK order. source_health is owned
	// by the jobs package but references sources, so it must be cleared here
	// too or deleting sources fails when jobs data is present.
	for _, stmt := range []string{
		"DELETE FROM push_outbox",
		"DELETE FROM push_subscriptions",
		"DELETE FROM user_preferences",
		"DELETE FROM sessions",
		"DELETE FROM users",
		"DELETE FROM news_links",
		"DELETE FROM lineups",
		"DELETE FROM broadcasts",
		"DELETE FROM evidence",
		"DELETE FROM matches",
		"DELETE FROM competitions",
		"DELETE FROM season_clubs",
		"DELETE FROM clubs",
		"DELETE FROM source_health",
		"DELETE FROM sources",
	} {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			t.Fatalf("cleanup %q: %v", stmt, err)
		}
	}

	t.Cleanup(func() { pool.Close() })
	return pool
}

func mustExec(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := pool.Exec(ctx, sql, args...); err != nil {
		t.Fatalf("exec %q: %v", sql, err)
	}
}

func insertSource(t *testing.T, pool *pgxpool.Pool, id, displayName string) {
	t.Helper()
	mustExec(t, pool, `INSERT INTO sources (id, display_name) VALUES ($1, $2)`, id, displayName)
}

func insertClub(t *testing.T, pool *pgxpool.Pool, id, slug, name, shortName string, aliases []string) {
	t.Helper()
	if aliases == nil {
		aliases = []string{}
	}
	mustExec(t, pool, `INSERT INTO clubs (id, slug, name, short_name, aliases) VALUES ($1, $2, $3, $4, $5)`,
		id, slug, name, shortName, aliases)
}

func insertCompetition(t *testing.T, pool *pgxpool.Pool, id, slug, name string, season int) {
	t.Helper()
	mustExec(t, pool, `INSERT INTO competitions (id, slug, name, season) VALUES ($1, $2, $3, $4)`, id, slug, name, season)
}

func insertMatch(t *testing.T, pool *pgxpool.Pool, id, competitionID, homeClubID, awayClubID, slug string, kickoffAt *time.Time, kickoffState string) {
	t.Helper()
	mustExec(t, pool, `
		INSERT INTO matches (id, competition_id, home_club_id, away_club_id, slug, kickoff_at, kickoff_state,
			broadcast_state, lineup_state, news_state)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'available', 'available', 'available')
	`, id, competitionID, homeClubID, awayClubID, slug, kickoffAt, kickoffState)
}

func insertEvidence(t *testing.T, pool *pgxpool.Pool, id, sourceID, matchID, dataType string) {
	t.Helper()
	now := time.Now()
	mustExec(t, pool, `
		INSERT INTO evidence (id, source_id, match_id, data_type, observed_at, fetched_at, parser_version, run_id, content_hash)
		VALUES ($1, $2, $3, $4, $5, $5, 'test-v1', 'test-run', $1)
	`, id, sourceID, matchID, dataType, now)
}

func insertBroadcast(t *testing.T, pool *pgxpool.Pool, id, matchID, evidenceID, channel, confidence string, verifiedAt time.Time) {
	t.Helper()
	mustExec(t, pool, `
		INSERT INTO broadcasts (id, match_id, evidence_id, channel, access, confidence, verified_at, availability)
		VALUES ($1, $2, $3, $4, 'free', $5, $6, 'available')
	`, id, matchID, evidenceID, channel, confidence, verifiedAt)
}

func insertLineup(t *testing.T, pool *pgxpool.Pool, id, matchID, clubID, evidenceID, side string, playersJSON []byte) {
	t.Helper()
	mustExec(t, pool, `
		INSERT INTO lineups (id, match_id, club_id, evidence_id, side, players, availability)
		VALUES ($1, $2, $3, $4, $5, $6, 'available')
	`, id, matchID, clubID, evidenceID, side, playersJSON)
}

func insertNewsLink(t *testing.T, pool *pgxpool.Pool, id, matchID, evidenceID, sourceID, title, url string, publishedAt time.Time) {
	t.Helper()
	mustExec(t, pool, `
		INSERT INTO news_links (id, match_id, evidence_id, source_id, title, url, published_at, availability)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'available')
	`, id, matchID, evidenceID, sourceID, title, url, publishedAt)
}

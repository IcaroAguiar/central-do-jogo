// Command seed loads deterministic Serie A 2026 demo data (20 clubs, the
// Brasileirao competition, and a handful of matches covering the REQ-004
// kickoff states and REQ-010 availability states) so the public read
// journeys (GOAL-004) have something real to render. It is idempotent: each
// row uses a deterministic ID and every insert upserts on conflict, so
// running it repeatedly against the same database is a no-op past the first
// run. It never touches ingest/worker adapters, which remain noop per scope.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/platform/config"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/database"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/logging"
)

func main() {
	if err := run(); err != nil {
		slog.Error("seed failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := logging.NewJSON(slog.LevelInfo)
	slog.SetDefault(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err := database.OpenPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	summary, err := Run(ctx, pool, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("seed: %w", err)
	}

	logger.Info("seed complete",
		"clubs", summary.Clubs,
		"competitions", summary.Competitions,
		"matches", summary.Matches,
		"broadcasts", summary.Broadcasts,
		"lineups", summary.Lineups,
		"news", summary.News,
	)
	return nil
}

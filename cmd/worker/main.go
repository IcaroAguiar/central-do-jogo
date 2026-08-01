package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/IcaroAguiar/central-do-jogo/internal/jobs"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/config"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/database"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/logging"
)

func main() {
	if err := run(); err != nil {
		slog.Error("worker exited", "error", err)
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.OpenPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}
	logger.Info("database migrations applied")

	store := jobs.NewStore(pool)
	healthStore := jobs.NewHealthStore(pool)

	handlers := jobs.HandlerRegistry{
		"ingest.openfootball_brazil": noopHandler("openfootball_brazil"),
		"ingest.cbf_match_center":    noopHandler("cbf_match_center"),
		"ingest.cbf_official_site":   noopHandler("cbf_official_site"),
		"ingest.gazetaesportiva":     noopHandler("gazetaesportiva"),
	}

	hostname, _ := os.Hostname()
	owner := fmt.Sprintf("worker-%s-%d", hostname, os.Getpid())

	worker := jobs.NewWorker(store, healthStore, handlers, owner)
	logger.Info("worker started", "owner", owner, "handlers", len(handlers))

	return worker.Run(ctx, logger)
}

func noopHandler(sourceID string) jobs.Handler {
	return func(ctx context.Context, job *jobs.Job) error {
		logger := logging.FromContext(ctx)
		logger.Info("noop handler executed", "source_id", sourceID, "job_id", job.ID)
		return nil
	}
}

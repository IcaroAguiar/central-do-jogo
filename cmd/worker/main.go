package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/IcaroAguiar/central-do-jogo/internal/features/push"
	"github.com/IcaroAguiar/central-do-jogo/internal/jobs"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/config"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/database"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/logging"
	"github.com/IcaroAguiar/central-do-jogo/internal/store"
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

	jobStore := jobs.NewStore(pool)
	healthStore := jobs.NewHealthStore(pool)
	pushStore := store.NewPushStore(pool)
	deliverer, err := push.DelivererForConfig(cfg.Push.Enabled, cfg.Push.VAPIDPublicKey, cfg.Push.VAPIDPrivateKey, cfg.Push.VAPIDSubject, nil)
	if err != nil {
		return fmt.Errorf("configure push deliverer: %w", err)
	}
	pushRunner := push.NewOutboxRunner(pushStore, pushStore, deliverer, cfg.Push.Enabled, nil)

	handlers := jobs.HandlerRegistry{
		"ingest.openfootball_brazil": noopHandler("openfootball_brazil"),
		"ingest.cbf_match_center":    noopHandler("cbf_match_center"),
		"ingest.cbf_official_site":   noopHandler("cbf_official_site"),
		"ingest.gazetaesportiva":     noopHandler("gazetaesportiva"),
		push.JobTypeDeliver:          push.DeliverHandler(pushRunner),
		push.JobTypeCleanup:          push.CleanupHandler(pushRunner),
	}

	hostname, _ := os.Hostname()
	owner := fmt.Sprintf("worker-%s-%d", hostname, os.Getpid())

	worker := jobs.NewWorker(jobStore, healthStore, handlers, owner)
	logger.Info("worker started",
		"owner", owner,
		"handlers", len(handlers),
		"push_enabled", cfg.Push.Enabled,
	)

	return worker.Run(ctx, logger)
}

func noopHandler(sourceID string) jobs.Handler {
	return func(ctx context.Context, job *jobs.Job) error {
		logger := logging.FromContext(ctx)
		logger.Info("noop handler executed", "source_id", sourceID, "job_id", job.ID)
		return nil
	}
}

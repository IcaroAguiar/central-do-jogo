// Command pushsmoke enqueues and optionally delivers a one-off Web Push alert
// for a single user (local/operator smoke for REQ-011).
//
// Prerequisites: VAPID keys configured, user already subscribed via the PWA,
// and DATABASE_URL pointing at the same database as the API.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/domain"
	"github.com/IcaroAguiar/central-do-jogo/internal/features/push"
	"github.com/IcaroAguiar/central-do-jogo/internal/jobs"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/config"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/database"
	"github.com/IcaroAguiar/central-do-jogo/internal/platform/logging"
	"github.com/IcaroAguiar/central-do-jogo/internal/store"
)

func main() {
	if err := run(); err != nil {
		slog.Error("pushsmoke failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	userID := flag.String("user-id", "", "target user id (usr_...)")
	email := flag.String("email", "", "target user email (alternative to -user-id)")
	title := flag.String("title", "Central do Jogo", "notification title")
	body := flag.String("body", "Smoke test de Web Push.", "notification body")
	urlPath := flag.String("url", "/", "same-origin path opened on click")
	matchID := flag.String("match-id", "mtc_smoke", "synthetic match id used in the idempotency key")
	version := flag.String("version", "", "alert version (default: UTC timestamp)")
	viaJob := flag.Bool("via-job", false, "enqueue push.deliver for the worker instead of delivering in-process")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if !cfg.Push.Enabled {
		return fmt.Errorf("web push is disabled: set VAPID_PUBLIC_KEY and VAPID_PRIVATE_KEY")
	}

	logger := logging.NewJSON(slog.LevelInfo)
	slog.SetDefault(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := database.OpenPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer pool.Close()

	users := store.NewUserStore(pool)
	target, err := resolveUser(ctx, users, strings.TrimSpace(*userID), strings.TrimSpace(*email))
	if err != nil {
		return err
	}

	pushStore := store.NewPushStore(pool)
	subs, err := pushStore.ListActiveByUser(ctx, target.ID)
	if err != nil {
		return err
	}
	if len(subs) == 0 {
		return fmt.Errorf("user %s has no active push subscriptions; subscribe via the PWA first", target.ID)
	}

	ver := strings.TrimSpace(*version)
	if ver == "" {
		ver = time.Now().UTC().Format("20060102T150405")
	}

	deliverer, err := push.DelivererForConfig(true, cfg.Push.VAPIDPublicKey, cfg.Push.VAPIDPrivateKey, cfg.Push.VAPIDSubject, nil)
	if err != nil {
		return fmt.Errorf("configure push deliverer: %w", err)
	}
	runner := push.NewOutboxRunner(pushStore, pushStore, deliverer, cfg.Push.Enabled, time.Now)

	entry, err := runner.EnqueueAlert(ctx, *matchID, domain.PushAlertSmokeTest, ver, []string{target.ID.String()}, push.AlertContent{
		Title: strings.TrimSpace(*title),
		Body:  strings.TrimSpace(*body),
		URL:   strings.TrimSpace(*urlPath),
	})
	if err != nil {
		return fmt.Errorf("enqueue alert: %w", err)
	}
	logger.Info("push outbox enqueued",
		"idempotency_key", entry.IdempotencyKey,
		"user_id", target.ID.String(),
		"subscriptions", len(subs),
	)

	if *viaJob {
		jobStore := jobs.NewStore(pool)
		job, err := push.EnqueueDeliverJob(ctx, jobStore, entry.IdempotencyKey, time.Now().UTC())
		if err != nil {
			return fmt.Errorf("enqueue deliver job: %w", err)
		}
		logger.Info("push.deliver job enqueued; ensure the worker is running", "job_id", job.ID)
		return nil
	}

	if err := runner.DeliverOutbox(ctx, entry.IdempotencyKey); err != nil {
		return fmt.Errorf("deliver outbox: %w", err)
	}
	logger.Info("push smoke delivered", "idempotency_key", entry.IdempotencyKey)
	return nil
}

func resolveUser(ctx context.Context, users *store.UserStore, userID, email string) (*domain.User, error) {
	switch {
	case userID != "" && email != "":
		return nil, fmt.Errorf("pass only one of -user-id or -email")
	case userID != "":
		user, err := users.GetByID(ctx, domain.ID(userID))
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, fmt.Errorf("user not found: %s", userID)
		}
		return user, nil
	case email != "":
		user, err := users.GetByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, fmt.Errorf("user not found for email %q", email)
		}
		return user, nil
	default:
		return nil, fmt.Errorf("required: -user-id or -email")
	}
}

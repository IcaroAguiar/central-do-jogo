package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/platform/config"
)

func main() {
	if err := run(); err != nil {
		log.Printf("worker exited: %v", err)
		os.Exit(1)
	}
}

func run() error {
	if _, err := config.Load(); err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("worker started (stub); waiting for shutdown signal")
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("worker shutting down")
			return nil
		case <-ticker.C:
			log.Printf("worker heartbeat (no jobs scheduled yet)")
		}
	}
}

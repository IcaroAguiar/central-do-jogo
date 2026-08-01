package db_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/IcaroAguiar/central-do-jogo/internal/platform/database"
)

func TestMigrateUp(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := database.OpenPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("OpenPool: %v", err)
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	if err := database.Migrate(ctx, pool); err != nil {
		t.Fatalf("Migrate second pass: %v", err)
	}
}

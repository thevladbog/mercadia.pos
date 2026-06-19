package migrate_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	platformmigrate "mercadia.dev/pos/platform/migrate"
)

func TestUpAppliesMigrationsWhenPostgresAvailable(t *testing.T) {
	databaseURL := os.Getenv("MERCADIA_STORE_EDGE_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("MERCADIA_STORE_EDGE_DATABASE_URL is not set")
	}

	migrationsDir := platformmigrate.FindMigrationsDir("MERCADIA_STORE_EDGE_MIGRATIONS_DIR", "infra/migrations/store-edge")
	if _, err := os.Stat(migrationsDir); err != nil {
		t.Skipf("migrations dir unavailable: %v", err)
	}
	migrationsDir, _ = filepath.Abs(migrationsDir)

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	result, err := platformmigrate.Up(ctx, db, "store-edge-test", migrationsDir)
	if err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	if result.AfterVersion < result.BeforeVersion {
		t.Fatalf("unexpected version regression: before=%d after=%d", result.BeforeVersion, result.AfterVersion)
	}

	second, err := platformmigrate.Up(ctx, db, "store-edge-test", migrationsDir)
	if err != nil {
		t.Fatalf("run migrations second time: %v", err)
	}
	if second.BeforeVersion != second.AfterVersion {
		t.Fatalf("expected idempotent second run, before=%d after=%d", second.BeforeVersion, second.AfterVersion)
	}
}

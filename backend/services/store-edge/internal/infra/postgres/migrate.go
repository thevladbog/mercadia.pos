package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"mercadia.dev/pos/platform/migrate"
)

const defaultMigrationsDir = "infra/migrations/store-edge"

func DefaultMigrationsDir() string {
	return migrate.FindMigrationsDir("MERCADIA_STORE_EDGE_MIGRATIONS_DIR", defaultMigrationsDir)
}

func RunMigrations(ctx context.Context, pool *pgxpool.Pool, dir string) (migrate.Result, error) {
	if dir == "" {
		dir = DefaultMigrationsDir()
	}
	return migrate.UpPool(ctx, pool, "store-edge", dir)
}

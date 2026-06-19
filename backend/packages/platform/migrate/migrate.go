package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type Result struct {
	Service       string
	Directory     string
	BeforeVersion int64
	AfterVersion  int64
}

func (r Result) Applied() bool {
	return r.AfterVersion > r.BeforeVersion
}

func FindMigrationsDir(envKey, relativePath string) string {
	if dir := os.Getenv(envKey); dir != "" {
		return dir
	}

	wd, err := os.Getwd()
	if err != nil {
		return relativePath
	}

	dir := wd
	for i := 0; i < 8; i++ {
		candidate := filepath.Join(dir, relativePath)
		if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
			abs, absErr := filepath.Abs(candidate)
			if absErr == nil {
				return abs
			}
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return relativePath
}

func UpPool(ctx context.Context, pool *pgxpool.Pool, service, dir string) (Result, error) {
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()
	return Up(ctx, db, service, dir)
}

func Up(ctx context.Context, db *sql.DB, service, dir string) (Result, error) {
	result := Result{
		Service:   service,
		Directory: dir,
	}

	if dir == "" {
		return result, fmt.Errorf("migrations directory is empty")
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return result, fmt.Errorf("set goose dialect: %w", err)
	}

	before, err := goose.GetDBVersion(db)
	if err != nil {
		return result, fmt.Errorf("read migration version before up: %w", err)
	}
	result.BeforeVersion = before

	if err := goose.UpContext(ctx, db, dir); err != nil {
		return result, fmt.Errorf("run migrations from %s: %w", dir, err)
	}

	after, err := goose.GetDBVersion(db)
	if err != nil {
		return result, fmt.Errorf("read migration version after up: %w", err)
	}
	result.AfterVersion = after

	return result, nil
}

func LogResult(result Result) {
	if result.Applied() {
		slog.Info("✅ migrations applied",
			"service", result.Service,
			"directory", result.Directory,
			"from_version", result.BeforeVersion,
			"to_version", result.AfterVersion,
		)
		return
	}

	slog.Info("⏭️ migrations already up to date",
		"service", result.Service,
		"directory", result.Directory,
		"version", result.AfterVersion,
	)
}

func LogError(service, dir string, err error) {
	slog.Error("❌ migration failed",
		"service", service,
		"directory", dir,
		"error", err,
	)
}

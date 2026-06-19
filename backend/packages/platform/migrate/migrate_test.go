package migrate_test

import (
	"testing"

	"mercadia.dev/pos/platform/migrate"
)

func TestResultApplied(t *testing.T) {
	if !(migrate.Result{BeforeVersion: 1, AfterVersion: 2}).Applied() {
		t.Fatal("expected applied when version increased")
	}
	if (migrate.Result{BeforeVersion: 2, AfterVersion: 2}).Applied() {
		t.Fatal("expected not applied when version unchanged")
	}
}

func TestFindMigrationsDirUsesEnv(t *testing.T) {
	t.Setenv("MERCADIA_TEST_MIGRATIONS_DIR", "/tmp/custom-migrations")
	if got := migrate.FindMigrationsDir("MERCADIA_TEST_MIGRATIONS_DIR", "infra/migrations/store-edge"); got != "/tmp/custom-migrations" {
		t.Fatalf("expected env override, got %q", got)
	}
}

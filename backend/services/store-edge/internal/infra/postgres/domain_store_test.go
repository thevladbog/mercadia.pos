package postgres_test

import (
	"context"
	"testing"

	"mercadia.dev/pos/services/store-edge/internal/domain"
)

// TestSeedDemoCashBalancesSeedsSafeOpeningBalance guards against a fresh
// Postgres-backed demo instance booting with zero cash balances, mirroring
// the same gap fixed for the in-memory store
// (memory.WithDemoCashBalances): ListCashBalances derives every balance
// from posted cash movements, so a demo store with none has no balances at
// all.
func TestSeedDemoCashBalancesSeedsSafeOpeningBalance(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	if err := store.SeedDemoCashBalances(ctx); err != nil {
		t.Fatalf("seed demo cash balances: %v", err)
	}

	movements, err := store.ListCashMovements(ctx, "store-1")
	if err != nil {
		t.Fatalf("list cash movements: %v", err)
	}

	var safeMovement *domain.CashMovement
	for i := range movements {
		if movements[i].ToContainerType == domain.CashContainerTypeSafe {
			safeMovement = &movements[i]
			break
		}
	}
	if safeMovement == nil {
		t.Fatalf("expected a posted movement into the safe, got none in %+v", movements)
	}
	if safeMovement.AmountMinor <= 0 {
		t.Fatalf("expected a positive demo opening amount, got %d", safeMovement.AmountMinor)
	}
	if safeMovement.Status != domain.CashMovementStatusPosted {
		t.Fatalf("expected posted status, got %s", safeMovement.Status)
	}
}

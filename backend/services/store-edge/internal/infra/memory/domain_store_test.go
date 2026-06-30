package memory

import (
	"context"
	"testing"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/domain"
)

func TestCountFailedAuthAttemptsUsesExclusiveCutoffAfterSuccess(t *testing.T) {
	store := NewStore()
	ctx := context.Background()
	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)

	if err := store.SaveAuthAttempt(ctx, domain.AuthAttempt{
		ID:            "auth_attempt_failed",
		StoreID:       "store-1",
		ActorID:       "cashier-1",
		Successful:    false,
		FailureReason: "invalid_pin",
		CreatedAt:     now,
	}); err != nil {
		t.Fatalf("save failed attempt: %v", err)
	}
	if err := store.SaveAuthAttempt(ctx, domain.AuthAttempt{
		ID:         "auth_attempt_success",
		StoreID:    "store-1",
		ActorID:    "cashier-1",
		Successful: true,
		CreatedAt:  now,
	}); err != nil {
		t.Fatalf("save success attempt: %v", err)
	}

	count, err := store.CountFailedAuthAttemptsSinceLastSuccess(ctx, "store-1", "cashier-1", now.Add(-time.Minute))
	if err != nil {
		t.Fatalf("count failed attempts: %v", err)
	}
	if count != 0 {
		t.Fatalf("failed attempt count = %d, want 0", count)
	}
}

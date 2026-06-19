package app_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"mercadia.dev/pos/services/central-backend/internal/app"
	"mercadia.dev/pos/services/central-backend/internal/infra/memory"
)

func TestAcceptSyncEventsPersistsCashMovement(t *testing.T) {
	store := memory.NewStore()
	registry := app.NewStoreRegistryService(store, store)
	if _, err := registry.RegisterStore(context.Background(), app.RegisterStoreCommand{
		StoreID:        "store-1",
		Name:           "Main Street",
		IdempotencyKey: "register-1",
	}); err != nil {
		t.Fatalf("register store: %v", err)
	}

	syncService := app.NewSyncService(store, store, store, store, store, store, store, store)
	cashService := app.NewCashMovementsService(store, store)

	postedAt := time.Date(2026, 6, 19, 15, 0, 0, 0, time.UTC)
	payload, err := json.Marshal(map[string]any{
		"storeId":           "store-1",
		"cashMovementId":    "cash-1",
		"type":              "safe_to_bank",
		"fromContainerId":   "safe-1",
		"fromContainerType": "safe",
		"toContainerId":     "bank-1",
		"toContainerType":   "bank",
		"amountMinor":       int64(200000),
		"currency":          "RUB",
		"actorId":           "senior-1",
		"postedAt":          postedAt,
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	result, err := syncService.AcceptEvents(context.Background(), app.AcceptSyncEventsCommand{
		StoreID:        "store-1",
		IdempotencyKey: "sync-cash-1",
		Events: []app.SyncEventInput{{
			EventID:    "obx-cash-1",
			EventType:  "cash.movement.posted",
			OccurredAt: postedAt,
			Payload:    payload,
		}},
	})
	if err != nil {
		t.Fatalf("accept events: %v", err)
	}
	if result.Accepted != 1 {
		t.Fatalf("accepted = %d", result.Accepted)
	}

	movement, err := cashService.GetCashMovement(context.Background(), "store-1", "cash-1")
	if err != nil {
		t.Fatalf("get cash movement: %v", err)
	}
	if movement.Type != "safe_to_bank" || movement.AmountMinor != 200000 {
		t.Fatalf("unexpected movement: %+v", movement)
	}
}

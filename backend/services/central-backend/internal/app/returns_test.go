package app_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"mercadia.dev/pos/services/central-backend/internal/app"
	"mercadia.dev/pos/services/central-backend/internal/infra/memory"
)

func TestAcceptSyncEventsPersistsReturn(t *testing.T) {
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
	returnsService := app.NewReturnsService(store, store)

	settledAt := time.Date(2026, 6, 19, 17, 0, 0, 0, time.UTC)
	payload, err := json.Marshal(map[string]any{
		"storeId":        "store-1",
		"returnId":       "ret-1",
		"receiptId":      "rcpt-1",
		"totalMinor":     int64(50000),
		"paymentIds":     []string{"pay-1", "pay-2"},
		"cashMovementId": "cash-1",
		"settledAt":      settledAt,
		"actorId":        "cashier-1",
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	result, err := syncService.AcceptEvents(context.Background(), app.AcceptSyncEventsCommand{
		StoreID:        "store-1",
		IdempotencyKey: "sync-return-1",
		Events: []app.SyncEventInput{{
			EventID:    "obx-ret-1",
			EventType:  "return.settled",
			OccurredAt: settledAt,
			Payload:    payload,
		}},
	})
	if err != nil {
		t.Fatalf("accept events: %v", err)
	}
	if result.Accepted != 1 {
		t.Fatalf("accepted = %d", result.Accepted)
	}

	ret, err := returnsService.GetReturn(context.Background(), "store-1", "ret-1")
	if err != nil {
		t.Fatalf("get return: %v", err)
	}
	if ret.ReceiptID != "rcpt-1" || ret.TotalMinor != 50000 || ret.CashMovementID != "cash-1" {
		t.Fatalf("unexpected return: %+v", ret)
	}
	if len(ret.PaymentIDs) != 2 || ret.PaymentIDs[0] != "pay-1" {
		t.Fatalf("payment ids = %+v", ret.PaymentIDs)
	}

	listed, err := returnsService.ListReturns(context.Background(), "store-1", app.PageParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("list returns: %v", err)
	}
	if listed.TotalCount != 1 || len(listed.Items) != 1 || listed.Items[0].ID != "ret-1" {
		t.Fatalf("listed returns = %+v", listed)
	}
}

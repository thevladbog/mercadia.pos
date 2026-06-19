package app_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"mercadia.dev/pos/services/central-backend/internal/app"
	"mercadia.dev/pos/services/central-backend/internal/infra/memory"
)

func TestAcceptSyncEventsPersistsPayment(t *testing.T) {
	store := memory.NewStore()
	registry := app.NewStoreRegistryService(store, store)
	if _, err := registry.RegisterStore(context.Background(), app.RegisterStoreCommand{
		StoreID:        "store-1",
		Name:           "Main Street",
		IdempotencyKey: "register-1",
	}); err != nil {
		t.Fatalf("register store: %v", err)
	}

	syncService := app.NewSyncService(store, store, store, store, store)
	paymentsService := app.NewPaymentsService(store, store)

	capturedAt := time.Date(2026, 6, 19, 14, 30, 0, 0, time.UTC)
	payload, err := json.Marshal(map[string]any{
		"storeId":     "store-1",
		"paymentId":   "pay-1",
		"receiptId":   "rcpt-1",
		"method":      "card",
		"amountMinor": int64(150000),
		"capturedAt":  capturedAt,
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	result, err := syncService.AcceptEvents(context.Background(), app.AcceptSyncEventsCommand{
		StoreID:        "store-1",
		IdempotencyKey: "sync-payment-1",
		Events: []app.SyncEventInput{{
			EventID:    "obx-pay-1",
			EventType:  "payment.captured",
			OccurredAt: capturedAt,
			Payload:    payload,
		}},
	})
	if err != nil {
		t.Fatalf("accept events: %v", err)
	}
	if result.Accepted != 1 {
		t.Fatalf("accepted = %d", result.Accepted)
	}

	payment, err := paymentsService.GetPayment(context.Background(), "store-1", "pay-1")
	if err != nil {
		t.Fatalf("get payment: %v", err)
	}
	if payment.ReceiptID != "rcpt-1" || payment.Method != "card" || payment.AmountMinor != 150000 {
		t.Fatalf("unexpected payment: %+v", payment)
	}
	if !payment.CapturedAt.Equal(capturedAt) {
		t.Fatalf("capturedAt = %v", payment.CapturedAt)
	}

	listed, err := paymentsService.ListPayments(context.Background(), "store-1", app.PageParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("list payments: %v", err)
	}
	if listed.TotalCount != 1 || len(listed.Items) != 1 || listed.Items[0].ID != "pay-1" {
		t.Fatalf("listed payments = %+v", listed)
	}
}

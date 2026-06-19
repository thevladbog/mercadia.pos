package app_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"mercadia.dev/pos/services/central-backend/internal/app"
	"mercadia.dev/pos/services/central-backend/internal/domain"
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

	syncService := app.NewSyncService(store, store, store, store, store, store)
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
	if payment.Status != domain.SyncedPaymentStatusCaptured {
		t.Fatalf("payment status = %s", payment.Status)
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

func TestAcceptSyncEventsUpdatesPaymentOnCancelAndRefund(t *testing.T) {
	store := memory.NewStore()
	registry := app.NewStoreRegistryService(store, store)
	if _, err := registry.RegisterStore(context.Background(), app.RegisterStoreCommand{
		StoreID:        "store-1",
		Name:           "Main Street",
		IdempotencyKey: "register-lifecycle",
	}); err != nil {
		t.Fatalf("register store: %v", err)
	}

	syncService := app.NewSyncService(store, store, store, store, store, store)
	paymentsService := app.NewPaymentsService(store, store)

	capturedAt := time.Date(2026, 6, 19, 14, 30, 0, 0, time.UTC)
	cancelledAt := time.Date(2026, 6, 19, 15, 0, 0, 0, time.UTC)
	refundedAt := time.Date(2026, 6, 19, 16, 0, 0, 0, time.UTC)

	capturePayload, err := json.Marshal(map[string]any{
		"storeId":     "store-1",
		"paymentId":   "pay-1",
		"receiptId":   "rcpt-1",
		"method":      "card",
		"amountMinor": int64(150000),
		"capturedAt":  capturedAt,
	})
	if err != nil {
		t.Fatalf("marshal capture payload: %v", err)
	}
	if _, err := syncService.AcceptEvents(context.Background(), app.AcceptSyncEventsCommand{
		StoreID:        "store-1",
		IdempotencyKey: "sync-capture",
		Events: []app.SyncEventInput{{
			EventID:    "obx-pay-1",
			EventType:  "payment.captured",
			OccurredAt: capturedAt,
			Payload:    capturePayload,
		}},
	}); err != nil {
		t.Fatalf("accept capture: %v", err)
	}

	cancelPayload, err := json.Marshal(map[string]any{
		"storeId":     "store-1",
		"paymentId":   "pay-2",
		"receiptId":   "rcpt-2",
		"method":      "cash",
		"amountMinor": int64(50000),
		"cancelledAt": cancelledAt,
		"actorId":     "cashier-1",
		"reason":      "void",
	})
	if err != nil {
		t.Fatalf("marshal cancel payload: %v", err)
	}
	if _, err := syncService.AcceptEvents(context.Background(), app.AcceptSyncEventsCommand{
		StoreID:        "store-1",
		IdempotencyKey: "sync-cancel-only",
		Events: []app.SyncEventInput{{
			EventID:    "obx-pay-cancel-only",
			EventType:  "payment.cancelled",
			OccurredAt: cancelledAt,
			Payload:    cancelPayload,
		}},
	}); err != nil {
		t.Fatalf("accept cancel-only: %v", err)
	}

	cancelExistingPayload, err := json.Marshal(map[string]any{
		"storeId":     "store-1",
		"paymentId":   "pay-1",
		"receiptId":   "rcpt-1",
		"method":      "card",
		"amountMinor": int64(150000),
		"cancelledAt": cancelledAt,
		"actorId":     "manager-1",
		"reason":      "customer request",
	})
	if err != nil {
		t.Fatalf("marshal cancel existing payload: %v", err)
	}
	if _, err := syncService.AcceptEvents(context.Background(), app.AcceptSyncEventsCommand{
		StoreID:        "store-1",
		IdempotencyKey: "sync-cancel-existing",
		Events: []app.SyncEventInput{{
			EventID:    "obx-pay-cancel-1",
			EventType:  "payment.cancelled",
			OccurredAt: cancelledAt,
			Payload:    cancelExistingPayload,
		}},
	}); err != nil {
		t.Fatalf("accept cancel existing: %v", err)
	}

	partialRefundPayload, err := json.Marshal(map[string]any{
		"storeId":              "store-1",
		"paymentId":            "pay-3",
		"receiptId":            "rcpt-3",
		"method":               "card",
		"amountMinor":          int64(100000),
		"refundedAmountMinor":  int64(40000),
		"remainingAmountMinor": int64(60000),
		"refundedAt":           refundedAt,
		"actorId":              "manager-1",
		"reason":               "partial return",
	})
	if err != nil {
		t.Fatalf("marshal partial refund payload: %v", err)
	}
	if _, err := syncService.AcceptEvents(context.Background(), app.AcceptSyncEventsCommand{
		StoreID:        "store-1",
		IdempotencyKey: "sync-partial-refund",
		Events: []app.SyncEventInput{{
			EventID:    "obx-pay-refund-partial",
			EventType:  "payment.refunded",
			OccurredAt: refundedAt,
			Payload:    partialRefundPayload,
		}},
	}); err != nil {
		t.Fatalf("accept partial refund: %v", err)
	}

	fullRefundPayload, err := json.Marshal(map[string]any{
		"storeId":              "store-1",
		"paymentId":            "pay-4",
		"receiptId":            "rcpt-4",
		"method":               "card",
		"amountMinor":          int64(80000),
		"refundedAmountMinor":  int64(80000),
		"remainingAmountMinor": int64(0),
		"refundedAt":           refundedAt,
		"actorId":              "manager-1",
		"reason":               "full return",
	})
	if err != nil {
		t.Fatalf("marshal full refund payload: %v", err)
	}
	if _, err := syncService.AcceptEvents(context.Background(), app.AcceptSyncEventsCommand{
		StoreID:        "store-1",
		IdempotencyKey: "sync-full-refund",
		Events: []app.SyncEventInput{{
			EventID:    "obx-pay-refund-full",
			EventType:  "payment.refunded",
			OccurredAt: refundedAt,
			Payload:    fullRefundPayload,
		}},
	}); err != nil {
		t.Fatalf("accept full refund: %v", err)
	}

	cancelledExisting, err := paymentsService.GetPayment(context.Background(), "store-1", "pay-1")
	if err != nil {
		t.Fatalf("get cancelled existing payment: %v", err)
	}
	if cancelledExisting.Status != domain.SyncedPaymentStatusCancelled || cancelledExisting.LastEventID != "obx-pay-cancel-1" {
		t.Fatalf("cancelled existing payment = %+v", cancelledExisting)
	}

	cancelOnly, err := paymentsService.GetPayment(context.Background(), "store-1", "pay-2")
	if err != nil {
		t.Fatalf("get cancel-only payment: %v", err)
	}
	if cancelOnly.Status != domain.SyncedPaymentStatusCancelled {
		t.Fatalf("cancel-only payment = %+v", cancelOnly)
	}

	partial, err := paymentsService.GetPayment(context.Background(), "store-1", "pay-3")
	if err != nil {
		t.Fatalf("get partial refund payment: %v", err)
	}
	if partial.Status != domain.SyncedPaymentStatusPartiallyRefunded || partial.RemainingAmountMinor != 60000 {
		t.Fatalf("partial refund payment = %+v", partial)
	}

	full, err := paymentsService.GetPayment(context.Background(), "store-1", "pay-4")
	if err != nil {
		t.Fatalf("get full refund payment: %v", err)
	}
	if full.Status != domain.SyncedPaymentStatusRefunded || full.RemainingAmountMinor != 0 {
		t.Fatalf("full refund payment = %+v", full)
	}
}

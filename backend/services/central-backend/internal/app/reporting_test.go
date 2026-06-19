package app_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"mercadia.dev/pos/services/central-backend/internal/app"
	"mercadia.dev/pos/services/central-backend/internal/infra/memory"
)

func TestStoreReportingSummaryAggregatesSyncedProjections(t *testing.T) {
	store := memory.NewStore()
	registry := app.NewStoreRegistryService(store, store)
	if _, err := registry.RegisterStore(context.Background(), app.RegisterStoreCommand{
		StoreID:        "store-1",
		Name:           "Main Street",
		Region:         "west",
		IdempotencyKey: "register-reporting",
	}); err != nil {
		t.Fatalf("register store: %v", err)
	}

	syncService := app.NewSyncService(store, store, store, store, store, store, store, store, store)
	reportingService := app.NewReportingService(store, store)

	windowStart := time.Date(2026, 6, 19, 0, 0, 0, 0, time.UTC)
	windowEnd := time.Date(2026, 6, 19, 23, 59, 59, 0, time.UTC)
	capturedAt := time.Date(2026, 6, 19, 14, 30, 0, 0, time.UTC)
	fiscalizedAt := time.Date(2026, 6, 19, 15, 0, 0, 0, time.UTC)
	postedAt := time.Date(2026, 6, 19, 16, 0, 0, 0, time.UTC)
	settledAt := time.Date(2026, 6, 19, 17, 0, 0, 0, time.UTC)
	closedAt := time.Date(2026, 6, 19, 22, 0, 0, 0, time.UTC)

	paymentPayload, err := json.Marshal(map[string]any{
		"storeId":     "store-1",
		"paymentId":   "pay-1",
		"receiptId":   "rcpt-1",
		"method":      "card",
		"amountMinor": int64(150000),
		"capturedAt":  capturedAt,
	})
	if err != nil {
		t.Fatalf("marshal payment payload: %v", err)
	}
	fiscalPayload, err := json.Marshal(map[string]any{
		"storeId":          "store-1",
		"fiscalDocumentId": "fisc-1",
		"receiptId":        "rcpt-1",
		"kind":             "receipt",
		"amountMinor":      int64(150000),
		"deviceId":         "kkt-1",
		"fiscalSign":       "sign-abc",
		"fiscalizedAt":     fiscalizedAt,
	})
	if err != nil {
		t.Fatalf("marshal fiscal payload: %v", err)
	}
	cashPayload, err := json.Marshal(map[string]any{
		"storeId":           "store-1",
		"cashMovementId":    "cm-1",
		"type":              "cash_in",
		"fromContainerId":   "external",
		"fromContainerType": "external",
		"toContainerId":     "drawer-1",
		"toContainerType":   "drawer",
		"amountMinor":       int64(50000),
		"currency":          "RUB",
		"actorId":           "cashier-1",
		"postedAt":          postedAt,
	})
	if err != nil {
		t.Fatalf("marshal cash payload: %v", err)
	}
	returnPayload, err := json.Marshal(map[string]any{
		"storeId":    "store-1",
		"returnId":   "ret-1",
		"receiptId":  "rcpt-1",
		"totalMinor": int64(25000),
		"paymentIds": []string{"pay-1"},
		"actorId":    "cashier-1",
		"settledAt":  settledAt,
	})
	if err != nil {
		t.Fatalf("marshal return payload: %v", err)
	}
	operationalDayPayload, err := json.Marshal(map[string]any{
		"storeId":          "store-1",
		"operationalDayId": "od-1",
		"businessDate":     "2026-06-19",
		"closedById":       "manager-1",
		"closedAt":         closedAt,
	})
	if err != nil {
		t.Fatalf("marshal operational day payload: %v", err)
	}

	events := []app.SyncEventInput{
		{EventID: "obx-pay-1", EventType: "payment.captured", OccurredAt: capturedAt, Payload: paymentPayload},
		{EventID: "obx-fisc-1", EventType: "fiscal.document.created", OccurredAt: fiscalizedAt, Payload: fiscalPayload},
		{EventID: "obx-cash-1", EventType: "cash.movement.posted", OccurredAt: postedAt, Payload: cashPayload},
		{EventID: "obx-ret-1", EventType: "return.settled", OccurredAt: settledAt, Payload: returnPayload},
		{EventID: "obx-od-1", EventType: "operational_day.closed", OccurredAt: closedAt, Payload: operationalDayPayload},
	}
	if _, err := syncService.AcceptEvents(context.Background(), app.AcceptSyncEventsCommand{
		StoreID:        "store-1",
		IdempotencyKey: "sync-reporting",
		Events:         events,
	}); err != nil {
		t.Fatalf("accept events: %v", err)
	}

	summary, err := reportingService.GetStoreSummary(context.Background(), "store-1", app.ReportingWindow{
		Since: windowStart,
		Until: windowEnd,
	})
	if err != nil {
		t.Fatalf("get store summary: %v", err)
	}
	if summary.FiscalReceiptCount != 1 || summary.FiscalReceiptAmountMinor != 150000 {
		t.Fatalf("fiscal receipt summary = %+v", summary)
	}
	if summary.PaymentsCapturedAmountMinor != 150000 {
		t.Fatalf("payments captured = %+v", summary)
	}
	if summary.CashMovementsPostedCount != 1 {
		t.Fatalf("cash movements = %+v", summary)
	}
	if summary.ReturnsSettledCount != 1 || summary.ReturnsSettledAmountMinor != 25000 {
		t.Fatalf("returns summary = %+v", summary)
	}
	if summary.OperationalDaysClosedCount != 1 {
		t.Fatalf("operational days = %+v", summary)
	}
}

func TestCentralReportingSummaryAggregatesAcrossStores(t *testing.T) {
	store := memory.NewStore()
	registry := app.NewStoreRegistryService(store, store)
	for _, registration := range []app.RegisterStoreCommand{
		{StoreID: "store-west", Name: "West", Region: "west", IdempotencyKey: "register-west"},
		{StoreID: "store-east", Name: "East", Region: "east", IdempotencyKey: "register-east"},
	} {
		if _, err := registry.RegisterStore(context.Background(), registration); err != nil {
			t.Fatalf("register store %s: %v", registration.StoreID, err)
		}
	}

	syncService := app.NewSyncService(store, store, store, store, store, store, store, store, store)
	reportingService := app.NewReportingService(store, store)

	windowStart := time.Date(2026, 6, 19, 0, 0, 0, 0, time.UTC)
	windowEnd := time.Date(2026, 6, 19, 23, 59, 59, 0, time.UTC)
	capturedAt := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)

	for _, fixture := range []struct {
		storeID   string
		paymentID string
		eventID   string
		amount    int64
	}{
		{storeID: "store-west", paymentID: "pay-west", eventID: "obx-west", amount: 100000},
		{storeID: "store-east", paymentID: "pay-east", eventID: "obx-east", amount: 200000},
	} {
		payload, err := json.Marshal(map[string]any{
			"storeId":     fixture.storeID,
			"paymentId":   fixture.paymentID,
			"receiptId":   "rcpt-1",
			"method":      "card",
			"amountMinor": fixture.amount,
			"capturedAt":  capturedAt,
		})
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		if _, err := syncService.AcceptEvents(context.Background(), app.AcceptSyncEventsCommand{
			StoreID:        fixture.storeID,
			IdempotencyKey: "sync-" + fixture.storeID,
			Events: []app.SyncEventInput{{
				EventID:    fixture.eventID,
				EventType:  "payment.captured",
				OccurredAt: capturedAt,
				Payload:    payload,
			}},
		}); err != nil {
			t.Fatalf("accept events for %s: %v", fixture.storeID, err)
		}
	}

	window := app.ReportingWindow{Since: windowStart, Until: windowEnd}

	central, err := reportingService.GetCentralSummary(context.Background(), window, "")
	if err != nil {
		t.Fatalf("get central summary: %v", err)
	}
	if central.StoreCount != 2 || central.PaymentsCapturedAmountMinor != 300000 {
		t.Fatalf("central summary = %+v", central)
	}

	westOnly, err := reportingService.GetCentralSummary(context.Background(), window, "west")
	if err != nil {
		t.Fatalf("get west summary: %v", err)
	}
	if westOnly.StoreCount != 1 || westOnly.PaymentsCapturedAmountMinor != 100000 {
		t.Fatalf("west summary = %+v", westOnly)
	}

	listed, err := reportingService.ListStoreSummaries(context.Background(), window, "", app.PageParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("list store summaries: %v", err)
	}
	if listed.TotalCount != 2 || len(listed.Items) != 2 {
		t.Fatalf("listed summaries = %+v", listed)
	}
}

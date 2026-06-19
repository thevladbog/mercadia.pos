package app_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"mercadia.dev/pos/services/central-backend/internal/app"
	"mercadia.dev/pos/services/central-backend/internal/infra/memory"
)

func TestAcceptSyncEventsPersistsFiscalDocument(t *testing.T) {
	store := memory.NewStore()
	registry := app.NewStoreRegistryService(store, store)
	if _, err := registry.RegisterStore(context.Background(), app.RegisterStoreCommand{
		StoreID:        "store-1",
		Name:           "Main Street",
		IdempotencyKey: "register-1",
	}); err != nil {
		t.Fatalf("register store: %v", err)
	}

	syncService := app.NewSyncService(store, store, store, store, store, store, store, store, store)
	fiscalService := app.NewFiscalDocumentsService(store, store)

	fiscalizedAt := time.Date(2026, 6, 19, 16, 0, 0, 0, time.UTC)
	payload, err := json.Marshal(map[string]any{
		"storeId":          "store-1",
		"fiscalDocumentId": "fisc-1",
		"receiptId":        "rcpt-1",
		"kind":             "sale",
		"amountMinor":      int64(150000),
		"deviceId":         "kkt-1",
		"fiscalSign":       "sign-abc",
		"fiscalizedAt":     fiscalizedAt,
		"returnId":         "ret-1",
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	result, err := syncService.AcceptEvents(context.Background(), app.AcceptSyncEventsCommand{
		StoreID:        "store-1",
		IdempotencyKey: "sync-fiscal-1",
		Events: []app.SyncEventInput{{
			EventID:    "obx-fisc-1",
			EventType:  "fiscal.document.created",
			OccurredAt: fiscalizedAt,
			Payload:    payload,
		}},
	})
	if err != nil {
		t.Fatalf("accept events: %v", err)
	}
	if result.Accepted != 1 {
		t.Fatalf("accepted = %d", result.Accepted)
	}

	document, err := fiscalService.GetFiscalDocument(context.Background(), "store-1", "fisc-1")
	if err != nil {
		t.Fatalf("get fiscal document: %v", err)
	}
	if document.Kind != "sale" || document.FiscalSign != "sign-abc" || document.ReturnID != "ret-1" {
		t.Fatalf("unexpected document: %+v", document)
	}

	listed, err := fiscalService.ListFiscalDocuments(context.Background(), "store-1", app.PageParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("list fiscal documents: %v", err)
	}
	if listed.TotalCount != 1 || len(listed.Items) != 1 || listed.Items[0].ID != "fisc-1" {
		t.Fatalf("listed fiscal documents = %+v", listed)
	}
}

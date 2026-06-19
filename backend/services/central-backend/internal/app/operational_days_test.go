package app_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"mercadia.dev/pos/services/central-backend/internal/app"
	"mercadia.dev/pos/services/central-backend/internal/infra/memory"
)

func TestAcceptSyncEventsPersistsOperationalDay(t *testing.T) {
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
	operationalDaysService := app.NewOperationalDaysService(store, store)

	closedAt := time.Date(2026, 6, 19, 23, 0, 0, 0, time.UTC)
	payload, err := json.Marshal(map[string]any{
		"storeId":          "store-1",
		"operationalDayId": "od-1",
		"businessDate":     "2026-06-19",
		"closedById":       "manager-1",
		"closedAt":         closedAt,
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	result, err := syncService.AcceptEvents(context.Background(), app.AcceptSyncEventsCommand{
		StoreID:        "store-1",
		IdempotencyKey: "sync-od-1",
		Events: []app.SyncEventInput{{
			EventID:    "obx-od-1",
			EventType:  "operational_day.closed",
			OccurredAt: closedAt,
			Payload:    payload,
		}},
	})
	if err != nil {
		t.Fatalf("accept events: %v", err)
	}
	if result.Accepted != 1 {
		t.Fatalf("accepted = %d", result.Accepted)
	}

	day, err := operationalDaysService.GetOperationalDay(context.Background(), "store-1", "od-1")
	if err != nil {
		t.Fatalf("get operational day: %v", err)
	}
	if day.BusinessDate != "2026-06-19" || day.ClosedByID != "manager-1" {
		t.Fatalf("unexpected operational day: %+v", day)
	}

	listed, err := operationalDaysService.ListOperationalDays(context.Background(), "store-1", app.PageParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("list operational days: %v", err)
	}
	if listed.TotalCount != 1 || len(listed.Items) != 1 || listed.Items[0].ID != "od-1" {
		t.Fatalf("listed operational days = %+v", listed)
	}
}

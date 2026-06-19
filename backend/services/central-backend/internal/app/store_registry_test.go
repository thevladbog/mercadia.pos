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

func TestRegisterStoreIsIdempotent(t *testing.T) {
	store := memory.NewStore()
	service := app.NewStoreRegistryService(store, store)

	command := app.RegisterStoreCommand{
		StoreID:        "store-1",
		Name:           "Main Street",
		Region:         "west",
		IdempotencyKey: "register-1",
	}

	first, err := service.RegisterStore(context.Background(), command)
	if err != nil {
		t.Fatalf("register store: %v", err)
	}
	second, err := service.RegisterStore(context.Background(), command)
	if err != nil {
		t.Fatalf("register store again: %v", err)
	}
	if first.Store.ID != second.Store.ID {
		t.Fatalf("store ids differ: %s vs %s", first.Store.ID, second.Store.ID)
	}

	count, err := service.CountStores(context.Background())
	if err != nil {
		t.Fatalf("count stores: %v", err)
	}
	if count != 1 {
		t.Fatalf("store count = %d", count)
	}
}

func TestAcceptSyncEventsPersistsCatalogProduct(t *testing.T) {
	store := memory.NewStore()
	registry := app.NewStoreRegistryService(store, store)
	syncService := app.NewSyncService(store, store, store, store, store)
	catalogService := app.NewCatalogService(store, store)

	_, err := registry.RegisterStore(context.Background(), app.RegisterStoreCommand{
		StoreID:        "store-1",
		Name:           "Main Street",
		IdempotencyKey: "register-1",
	})
	if err != nil {
		t.Fatalf("register store: %v", err)
	}

	payload, err := json.Marshal(map[string]any{
		"productId":      "sku-1",
		"name":           "Milk",
		"barcodes":       []string{"4600000000000"},
		"unitPriceMinor": int64(19999),
		"taxCategoryId":  "vat_20",
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	result, err := syncService.AcceptEvents(context.Background(), app.AcceptSyncEventsCommand{
		StoreID:        "store-1",
		IdempotencyKey: "sync-1",
		Events: []app.SyncEventInput{{
			EventID:   "evt-1",
			EventType: "catalog.product.upserted",
			Payload:   payload,
		}},
	})
	if err != nil {
		t.Fatalf("accept events: %v", err)
	}
	if result.Accepted != 1 {
		t.Fatalf("accepted = %d", result.Accepted)
	}

	catalog, err := catalogService.ListProducts(context.Background(), "store-1")
	if err != nil {
		t.Fatalf("list products: %v", err)
	}
	if len(catalog.Products) != 1 || catalog.Products[0].ID != "sku-1" {
		t.Fatalf("unexpected catalog: %+v", catalog.Products)
	}
}

func TestCatalogDeltaReturnsUpdatedProducts(t *testing.T) {
	now := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
	store := memory.NewStore(memory.WithStores(domain.Store{
		ID:           "store-1",
		Name:         "Main Street",
		Region:       "default",
		RegisteredAt: now,
		UpdatedAt:    now,
	}), memory.WithProducts(domain.CatalogProduct{
		ID:             "sku-old",
		StoreID:        "store-1",
		Name:           "Old Product",
		Barcodes:       []string{"111"},
		UnitPriceMinor: 100,
		Active:         true,
		Version:        1,
		UpdatedAt:      now.Add(-time.Hour),
	}, domain.CatalogProduct{
		ID:             "sku-new",
		StoreID:        "store-1",
		Name:           "New Product",
		Barcodes:       []string{"222"},
		UnitPriceMinor: 200,
		Active:         true,
		Version:        1,
		UpdatedAt:      now,
	}))

	service := app.NewCatalogService(store, store)
	result, err := service.CatalogDelta(context.Background(), "store-1", now.Add(-30*time.Minute))
	if err != nil {
		t.Fatalf("catalog delta: %v", err)
	}
	if len(result.Products) != 1 || result.Products[0].ID != "sku-new" {
		t.Fatalf("unexpected delta: %+v", result.Products)
	}
}

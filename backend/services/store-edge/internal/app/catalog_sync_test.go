package app_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/app"
	"mercadia.dev/pos/services/store-edge/internal/domain"
	"mercadia.dev/pos/services/store-edge/internal/infra/central"
	"mercadia.dev/pos/services/store-edge/internal/infra/memory"
)

func TestCatalogSyncPullsProductsFromCentral(t *testing.T) {
	updatedAt := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
	centralServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/stores/store-1/catalog/delta" {
			http.NotFound(w, r)
			return
		}
		sinceRaw := r.URL.Query().Get("since")
		since, err := time.Parse(time.RFC3339, sinceRaw)
		if err != nil {
			http.Error(w, "invalid since", http.StatusBadRequest)
			return
		}

		products := []map[string]any{}
		if since.Before(updatedAt) {
			products = append(products, map[string]any{
				"id":             "remote-sku-1",
				"storeId":        "store-1",
				"name":           "Remote Milk",
				"barcodes":       []string{"7770001112223"},
				"unitPriceMinor": 25000,
				"taxCategoryId":  "vat_20",
				"active":         true,
				"version":        1,
				"updatedAt":      updatedAt,
			})
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"since":    since,
			"products": products,
		})
	}))
	defer centralServer.Close()

	store := memory.NewStore()
	service := app.NewCatalogSyncService(store, store, central.NewClient(centralServer.URL, centralServer.Client()))

	result, err := service.Sync(context.Background(), app.SyncCatalogCommand{StoreID: "store-1"})
	if err != nil {
		t.Fatalf("sync catalog: %v", err)
	}
	if result.ProductsCount != 1 {
		t.Fatalf("products count = %d", result.ProductsCount)
	}

	product, err := store.FindProductByBarcode(context.Background(), "7770001112223")
	if err != nil {
		t.Fatalf("find synced product: %v", err)
	}
	if product.ID != "remote-sku-1" || product.Name != "Remote Milk" {
		t.Fatalf("unexpected product: %+v", product)
	}

	second, err := service.Sync(context.Background(), app.SyncCatalogCommand{StoreID: "store-1"})
	if err != nil {
		t.Fatalf("second sync: %v", err)
	}
	if second.ProductsCount != 0 {
		t.Fatalf("expected empty delta on second sync, got %d", second.ProductsCount)
	}
}

func TestCatalogSyncRequiresStoreID(t *testing.T) {
	store := memory.NewStore()
	service := app.NewCatalogSyncService(store, store, central.NewClient("http://example.invalid", nil))

	_, err := service.Sync(context.Background(), app.SyncCatalogCommand{})
	if err != app.ErrInvalidCatalogSyncCommand {
		t.Fatalf("expected invalid command error, got %v", err)
	}
}

func TestSaveProductUpdatesBarcodeIndex(t *testing.T) {
	store := memory.NewStore()
	product, err := domain.NewProduct(domain.Product{
		ID:             "sku-sync-1",
		Name:           "Synced Bread",
		Barcodes:       []string{"8880001112223"},
		UnitPriceMinor: 9900,
		TaxCategoryID:  "vat_10",
	})
	if err != nil {
		t.Fatalf("new product: %v", err)
	}
	if err := store.SaveProduct(context.Background(), product); err != nil {
		t.Fatalf("save product: %v", err)
	}

	found, err := store.FindProductByBarcode(context.Background(), "8880001112223")
	if err != nil {
		t.Fatalf("find product: %v", err)
	}
	if found.ID != product.ID {
		t.Fatalf("product id = %s", found.ID)
	}
}

package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mercadia.dev/pos/services/central-backend/internal/api"
	"mercadia.dev/pos/services/central-backend/internal/app"
	"mercadia.dev/pos/services/central-backend/internal/infra/memory"
)

func newTestServer() http.Handler {
	store := memory.NewStore()
	repo := store
	return api.NewServerWithServices(api.Services{
		StoreRegistry: app.NewStoreRegistryService(repo, repo),
		Sync:          app.NewSyncService(repo, repo, repo, repo, repo),
		Catalog:       app.NewCatalogService(repo, repo),
		Payments:      app.NewPaymentsService(repo, repo),
	})
}

func TestOpenAPIExposesCentralOperations(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)

	newTestServer().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}

	var document map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &document); err != nil {
		t.Fatalf("decode OpenAPI: %v", err)
	}

	paths := document["paths"].(map[string]any)
	for _, path := range []string{
		"/v1/central/status",
		"/v1/stores",
		"/v1/stores/{storeId}/sync-events",
		"/v1/stores/{storeId}/catalog/products",
		"/v1/stores/{storeId}/catalog/delta",
	} {
		if _, ok := paths[path]; !ok {
			t.Fatalf("expected %s path", path)
		}
	}
}

func TestRegisterStoreAndStatus(t *testing.T) {
	server := newTestServer()

	registerBody := bytes.NewBufferString(`{"storeId":"store-1","name":"Main Street","region":"west"}`)
	registerRequest := httptest.NewRequest(http.MethodPost, "/v1/stores", registerBody)
	registerRequest.Header.Set("Content-Type", "application/json")
	registerRequest.Header.Set("Idempotency-Key", "register-1")
	registerResponse := httptest.NewRecorder()
	server.ServeHTTP(registerResponse, registerRequest)
	if registerResponse.Code != http.StatusAccepted {
		t.Fatalf("register status = %d body=%s", registerResponse.Code, registerResponse.Body.String())
	}

	statusResponse := httptest.NewRecorder()
	statusRequest := httptest.NewRequest(http.MethodGet, "/v1/central/status", nil)
	server.ServeHTTP(statusResponse, statusRequest)
	if statusResponse.Code != http.StatusOK {
		t.Fatalf("status = %d", statusResponse.Code)
	}

	var status map[string]any
	if err := json.Unmarshal(statusResponse.Body.Bytes(), &status); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if int(status["storeCount"].(float64)) != 1 {
		t.Fatalf("storeCount = %v", status["storeCount"])
	}
}

func TestSyncEventsAndCatalogEndpoints(t *testing.T) {
	server := newTestServer()

	registerRequest := httptest.NewRequest(http.MethodPost, "/v1/stores", bytes.NewBufferString(`{"storeId":"store-1","name":"Main Street"}`))
	registerRequest.Header.Set("Content-Type", "application/json")
	registerRequest.Header.Set("Idempotency-Key", "register-1")
	registerResponse := httptest.NewRecorder()
	server.ServeHTTP(registerResponse, registerRequest)
	if registerResponse.Code != http.StatusAccepted {
		t.Fatalf("register status = %d", registerResponse.Code)
	}

	since := time.Now().UTC().Add(-time.Hour).Format(time.RFC3339)
	syncBody := bytes.NewBufferString(`{"events":[{"eventId":"evt-1","eventType":"catalog.product.upserted","payload":{"productId":"sku-1","name":"Milk","barcodes":["4600000000000"],"unitPriceMinor":19999,"taxCategoryId":"vat_20"}}]}`)
	syncRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/sync-events", syncBody)
	syncRequest.Header.Set("Content-Type", "application/json")
	syncRequest.Header.Set("Idempotency-Key", "sync-1")
	syncResponse := httptest.NewRecorder()
	server.ServeHTTP(syncResponse, syncRequest)
	if syncResponse.Code != http.StatusAccepted {
		t.Fatalf("sync status = %d body=%s", syncResponse.Code, syncResponse.Body.String())
	}

	productsResponse := httptest.NewRecorder()
	productsRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/catalog/products", nil)
	server.ServeHTTP(productsResponse, productsRequest)
	if productsResponse.Code != http.StatusOK {
		t.Fatalf("products status = %d body=%s", productsResponse.Code, productsResponse.Body.String())
	}

	deltaResponse := httptest.NewRecorder()
	deltaRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/catalog/delta?since="+since, nil)
	server.ServeHTTP(deltaResponse, deltaRequest)
	if deltaResponse.Code != http.StatusOK {
		t.Fatalf("delta status = %d body=%s", deltaResponse.Code, deltaResponse.Body.String())
	}

	var delta map[string]any
	if err := json.Unmarshal(deltaResponse.Body.Bytes(), &delta); err != nil {
		t.Fatalf("decode delta: %v", err)
	}
	products := delta["products"].([]any)
	if len(products) != 1 {
		t.Fatalf("delta products = %d", len(products))
	}
}

func TestListStoreSyncEventsReturnsAcceptedEvents(t *testing.T) {
	server := newTestServer()

	registerRequest := httptest.NewRequest(http.MethodPost, "/v1/stores", bytes.NewBufferString(`{"storeId":"store-1","name":"Main Street"}`))
	registerRequest.Header.Set("Content-Type", "application/json")
	registerRequest.Header.Set("Idempotency-Key", "register-list-sync")
	registerResponse := httptest.NewRecorder()
	server.ServeHTTP(registerResponse, registerRequest)
	if registerResponse.Code != http.StatusAccepted {
		t.Fatalf("register status = %d", registerResponse.Code)
	}

	syncBody := bytes.NewBufferString(`{"events":[{"eventId":"evt-list-1","eventType":"catalog.product.upserted","payload":{"productId":"sku-1","name":"Milk","barcodes":["4600000000000"],"unitPriceMinor":19999,"taxCategoryId":"vat_20"}}]}`)
	syncRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/sync-events", syncBody)
	syncRequest.Header.Set("Content-Type", "application/json")
	syncRequest.Header.Set("Idempotency-Key", "sync-list-1")
	syncResponse := httptest.NewRecorder()
	server.ServeHTTP(syncResponse, syncRequest)
	if syncResponse.Code != http.StatusAccepted {
		t.Fatalf("sync status = %d body=%s", syncResponse.Code, syncResponse.Body.String())
	}

	listResponse := httptest.NewRecorder()
	listRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/sync-events", nil)
	server.ServeHTTP(listResponse, listRequest)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("list sync events status = %d body=%s", listResponse.Code, listResponse.Body.String())
	}

	var listed api.PaginatedSyncEventsResponse
	if err := json.Unmarshal(listResponse.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode sync events list: %v", err)
	}
	if listed.TotalCount != 1 || len(listed.Items) != 1 {
		t.Fatalf("listed sync events = %+v", listed)
	}
	if listed.Items[0].SourceEventID != "evt-list-1" || listed.Items[0].EventType != "catalog.product.upserted" {
		t.Fatalf("sync event = %+v", listed.Items[0])
	}
}

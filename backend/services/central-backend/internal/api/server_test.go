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
		Sync:            app.NewSyncService(repo, repo, repo, repo, repo, repo, repo),
		Catalog:         app.NewCatalogService(repo, repo),
		Payments:        app.NewPaymentsService(repo, repo),
		CashMovements:   app.NewCashMovementsService(repo, repo),
		FiscalDocuments: app.NewFiscalDocumentsService(repo, repo),
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
		"/v1/stores/{storeId}/payments",
		"/v1/stores/{storeId}/payments/{paymentId}",
		"/v1/stores/{storeId}/cash-movements",
		"/v1/stores/{storeId}/cash-movements/{cashMovementId}",
		"/v1/stores/{storeId}/fiscal-documents",
		"/v1/stores/{storeId}/fiscal-documents/{fiscalDocumentId}",
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

func TestSyncEventsProjectPaymentsAndCashMovements(t *testing.T) {
	server := newTestServer()

	registerRequest := httptest.NewRequest(http.MethodPost, "/v1/stores", bytes.NewBufferString(`{"storeId":"store-1","name":"Main Street"}`))
	registerRequest.Header.Set("Content-Type", "application/json")
	registerRequest.Header.Set("Idempotency-Key", "register-projection-1")
	registerResponse := httptest.NewRecorder()
	server.ServeHTTP(registerResponse, registerRequest)
	if registerResponse.Code != http.StatusAccepted {
		t.Fatalf("register status = %d", registerResponse.Code)
	}

	capturedAt := time.Date(2026, 6, 19, 14, 30, 0, 0, time.UTC).Format(time.RFC3339)
	postedAt := time.Date(2026, 6, 19, 15, 0, 0, 0, time.UTC).Format(time.RFC3339)
	syncBody := bytes.NewBufferString(`{"events":[` +
		`{"eventId":"obx-pay-1","eventType":"payment.captured","payload":{"storeId":"store-1","paymentId":"pay-1","receiptId":"rcpt-1","method":"card","amountMinor":150000,"capturedAt":"` + capturedAt + `"}},` +
		`{"eventId":"obx-cash-1","eventType":"cash.movement.posted","payload":{"storeId":"store-1","cashMovementId":"cash-1","type":"safe_to_bank","fromContainerId":"safe-1","fromContainerType":"safe","toContainerId":"bank-1","toContainerType":"bank","amountMinor":200000,"currency":"RUB","actorId":"senior-1","postedAt":"` + postedAt + `"}}` +
		`]}`)
	syncRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/sync-events", syncBody)
	syncRequest.Header.Set("Content-Type", "application/json")
	syncRequest.Header.Set("Idempotency-Key", "sync-projection-1")
	syncResponse := httptest.NewRecorder()
	server.ServeHTTP(syncResponse, syncRequest)
	if syncResponse.Code != http.StatusAccepted {
		t.Fatalf("sync status = %d body=%s", syncResponse.Code, syncResponse.Body.String())
	}

	paymentResponse := httptest.NewRecorder()
	paymentRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/payments/pay-1", nil)
	server.ServeHTTP(paymentResponse, paymentRequest)
	if paymentResponse.Code != http.StatusOK {
		t.Fatalf("get payment status = %d body=%s", paymentResponse.Code, paymentResponse.Body.String())
	}

	movementResponse := httptest.NewRecorder()
	movementRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/cash-movements/cash-1", nil)
	server.ServeHTTP(movementResponse, movementRequest)
	if movementResponse.Code != http.StatusOK {
		t.Fatalf("get cash movement status = %d body=%s", movementResponse.Code, movementResponse.Body.String())
	}

	listPaymentsResponse := httptest.NewRecorder()
	listPaymentsRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/payments", nil)
	server.ServeHTTP(listPaymentsResponse, listPaymentsRequest)
	if listPaymentsResponse.Code != http.StatusOK {
		t.Fatalf("list payments status = %d body=%s", listPaymentsResponse.Code, listPaymentsResponse.Body.String())
	}

	var listedPayments api.PaginatedSyncedPaymentsResponse
	if err := json.Unmarshal(listPaymentsResponse.Body.Bytes(), &listedPayments); err != nil {
		t.Fatalf("decode payments list: %v", err)
	}
	if listedPayments.TotalCount != 1 || len(listedPayments.Items) != 1 {
		t.Fatalf("listed payments = %+v", listedPayments)
	}
}

func TestSyncEventsUpdatePaymentLifecycle(t *testing.T) {
	server := newTestServer()

	registerRequest := httptest.NewRequest(http.MethodPost, "/v1/stores", bytes.NewBufferString(`{"storeId":"store-1","name":"Main Street"}`))
	registerRequest.Header.Set("Content-Type", "application/json")
	registerRequest.Header.Set("Idempotency-Key", "register-lifecycle-http")
	registerResponse := httptest.NewRecorder()
	server.ServeHTTP(registerResponse, registerRequest)
	if registerResponse.Code != http.StatusAccepted {
		t.Fatalf("register status = %d", registerResponse.Code)
	}

	capturedAt := time.Date(2026, 6, 19, 14, 30, 0, 0, time.UTC).Format(time.RFC3339)
	cancelledAt := time.Date(2026, 6, 19, 15, 0, 0, 0, time.UTC).Format(time.RFC3339)
	syncBody := bytes.NewBufferString(`{"events":[` +
		`{"eventId":"obx-pay-cap","eventType":"payment.captured","payload":{"storeId":"store-1","paymentId":"pay-1","receiptId":"rcpt-1","method":"card","amountMinor":150000,"capturedAt":"` + capturedAt + `"}},` +
		`{"eventId":"obx-pay-cancel","eventType":"payment.cancelled","payload":{"storeId":"store-1","paymentId":"pay-1","receiptId":"rcpt-1","method":"card","amountMinor":150000,"cancelledAt":"` + cancelledAt + `","actorId":"manager-1","reason":"void"}}` +
		`]}`)
	syncRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/sync-events", syncBody)
	syncRequest.Header.Set("Content-Type", "application/json")
	syncRequest.Header.Set("Idempotency-Key", "sync-lifecycle-http")
	syncResponse := httptest.NewRecorder()
	server.ServeHTTP(syncResponse, syncRequest)
	if syncResponse.Code != http.StatusAccepted {
		t.Fatalf("sync status = %d body=%s", syncResponse.Code, syncResponse.Body.String())
	}

	paymentResponse := httptest.NewRecorder()
	paymentRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/payments/pay-1", nil)
	server.ServeHTTP(paymentResponse, paymentRequest)
	if paymentResponse.Code != http.StatusOK {
		t.Fatalf("get payment status = %d body=%s", paymentResponse.Code, paymentResponse.Body.String())
	}

	var payment api.SyncedPaymentResponse
	if err := json.Unmarshal(paymentResponse.Body.Bytes(), &payment); err != nil {
		t.Fatalf("decode payment: %v", err)
	}
	if payment.Status != "cancelled" || payment.LastEventID != "obx-pay-cancel" {
		t.Fatalf("payment = %+v", payment)
	}
}

func TestSyncEventsProjectFiscalDocument(t *testing.T) {
	server := newTestServer()

	registerRequest := httptest.NewRequest(http.MethodPost, "/v1/stores", bytes.NewBufferString(`{"storeId":"store-1","name":"Main Street"}`))
	registerRequest.Header.Set("Content-Type", "application/json")
	registerRequest.Header.Set("Idempotency-Key", "register-fiscal-http")
	registerResponse := httptest.NewRecorder()
	server.ServeHTTP(registerResponse, registerRequest)
	if registerResponse.Code != http.StatusAccepted {
		t.Fatalf("register status = %d", registerResponse.Code)
	}

	fiscalizedAt := time.Date(2026, 6, 19, 16, 0, 0, 0, time.UTC).Format(time.RFC3339)
	syncBody := bytes.NewBufferString(`{"events":[` +
		`{"eventId":"obx-fisc-1","eventType":"fiscal.document.created","payload":{"storeId":"store-1","fiscalDocumentId":"fisc-1","receiptId":"rcpt-1","kind":"sale","amountMinor":150000,"deviceId":"kkt-1","fiscalSign":"sign-abc","fiscalizedAt":"` + fiscalizedAt + `"}}` +
		`]}`)
	syncRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/sync-events", syncBody)
	syncRequest.Header.Set("Content-Type", "application/json")
	syncRequest.Header.Set("Idempotency-Key", "sync-fiscal-http")
	syncResponse := httptest.NewRecorder()
	server.ServeHTTP(syncResponse, syncRequest)
	if syncResponse.Code != http.StatusAccepted {
		t.Fatalf("sync status = %d body=%s", syncResponse.Code, syncResponse.Body.String())
	}

	documentResponse := httptest.NewRecorder()
	documentRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/fiscal-documents/fisc-1", nil)
	server.ServeHTTP(documentResponse, documentRequest)
	if documentResponse.Code != http.StatusOK {
		t.Fatalf("get fiscal document status = %d body=%s", documentResponse.Code, documentResponse.Body.String())
	}

	var document api.SyncedFiscalDocumentResponse
	if err := json.Unmarshal(documentResponse.Body.Bytes(), &document); err != nil {
		t.Fatalf("decode fiscal document: %v", err)
	}
	if document.Kind != "sale" || document.FiscalSign != "sign-abc" {
		t.Fatalf("fiscal document = %+v", document)
	}
}

package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mercadia.dev/pos/services/central-backend/internal/api"
	"mercadia.dev/pos/services/central-backend/internal/app"
	"mercadia.dev/pos/services/central-backend/internal/domain"
	"mercadia.dev/pos/services/central-backend/internal/infra/memory"
)

func newTestServer() http.Handler {
	store := memory.NewStore()
	if err := seedHTTPTestAdmin(store); err != nil {
		panic(err)
	}
	return api.NewServerWithServices(newTestServices(store))
}

func newTestServices(store *memory.Store) api.Services {
	repo := store
	return api.Services{
		StoreRegistry:   app.NewStoreRegistryService(repo, repo),
		Sync:            app.NewSyncService(repo, repo, repo, repo, repo, repo, repo, repo, repo),
		Catalog:         app.NewCatalogService(repo, repo),
		Payments:        app.NewPaymentsService(repo, repo),
		CashMovements:   app.NewCashMovementsService(repo, repo),
		FiscalDocuments: app.NewFiscalDocumentsService(repo, repo),
		Returns:         app.NewReturnsService(repo, repo),
		OperationalDays: app.NewOperationalDaysService(repo, repo),
		Reporting:       app.NewReportingService(repo, repo),
		Auth:            app.NewAuthService(repo, repo),
		CentralUsers:    app.NewCentralUsersService(repo),
		SyncAPIKey:      app.NewSyncAPIKeyService(""),
	}
}

func newTestServicesWithSyncAPIKey(store *memory.Store, syncAPIKey string) api.Services {
	services := newTestServices(store)
	services.SyncAPIKey = app.NewSyncAPIKeyService(syncAPIKey)
	return services
}

func seedHTTPTestAdmin(store *memory.Store) error {
	passwordHash, err := app.HashPassword("admin-pass")
	if err != nil {
		return err
	}
	user, err := domain.NewCentralUser(domain.CentralUser{
		ID:           "admin-1",
		Email:        "admin@example.com",
		DisplayName:  "Admin",
		PasswordHash: passwordHash,
		Roles:        []domain.CentralRole{domain.CentralRoleAdmin},
		CreatedAt:    time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		return err
	}
	return store.SaveUser(context.Background(), user)
}

func seedHTTPTestViewer(store *memory.Store) error {
	passwordHash, err := app.HashPassword("viewer-pass")
	if err != nil {
		return err
	}
	user, err := domain.NewCentralUser(domain.CentralUser{
		ID:           "viewer-1",
		Email:        "viewer@example.com",
		DisplayName:  "Viewer",
		PasswordHash: passwordHash,
		Roles:        []domain.CentralRole{domain.CentralRoleViewer},
		CreatedAt:    time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		return err
	}
	return store.SaveUser(context.Background(), user)
}

func loginTestSession(t *testing.T, server http.Handler, email, password string) string {
	t.Helper()
	body := bytes.NewBufferString(`{"email":"` + email + `","password":"` + password + `"}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/auth/sessions", body)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("login status = %d body=%s", response.Code, response.Body.String())
	}
	var payload api.CentralSessionAcceptedResponse
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	return payload.Session.Token
}

func newTestServerWithoutSeed() http.Handler {
	store := memory.NewStore()
	return api.NewServerWithServices(newTestServices(store))
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
		"/v1/stores/{storeId}/returns",
		"/v1/stores/{storeId}/returns/{returnId}",
		"/v1/stores/{storeId}/operational-days",
		"/v1/stores/{storeId}/operational-days/{operationalDayId}",
		"/v1/stores/{storeId}/reporting/summary",
		"/v1/central/reporting/summary",
		"/v1/central/reporting/stores",
		"/v1/auth/sessions",
		"/v1/central/users",
		"/v1/central/users/{userId}",
	} {
		if _, ok := paths[path]; !ok {
			t.Fatalf("expected %s path", path)
		}
	}
}

func loginAdminAndRegisterStore(t *testing.T, server http.Handler, body, idempotencyKey string) string {
	t.Helper()
	token := loginTestSession(t, server, "admin@example.com", "admin-pass")
	registerTestStore(t, server, body, idempotencyKey, registerTestStoreAuth{sessionToken: token})
	return token
}

func TestRegisterStoreAndStatus(t *testing.T) {
	server := newTestServer()

	loginAdminAndRegisterStore(t, server, `{"storeId":"store-1","name":"Main Street","region":"west"}`, "register-1")

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

	loginAdminAndRegisterStore(t, server, `{"storeId":"store-1","name":"Main Street"}`, "register-1")

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

	loginAdminAndRegisterStore(t, server, `{"storeId":"store-1","name":"Main Street"}`, "register-list-sync")

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

	loginAdminAndRegisterStore(t, server, `{"storeId":"store-1","name":"Main Street"}`, "register-projection-1")

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

	loginAdminAndRegisterStore(t, server, `{"storeId":"store-1","name":"Main Street"}`, "register-lifecycle-http")

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

	loginAdminAndRegisterStore(t, server, `{"storeId":"store-1","name":"Main Street"}`, "register-fiscal-http")

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

func TestSyncEventsProjectReturn(t *testing.T) {
	server := newTestServer()

	loginAdminAndRegisterStore(t, server, `{"storeId":"store-1","name":"Main Street"}`, "register-return-http")

	settledAt := time.Date(2026, 6, 19, 17, 0, 0, 0, time.UTC).Format(time.RFC3339)
	syncBody := bytes.NewBufferString(`{"events":[` +
		`{"eventId":"obx-ret-1","eventType":"return.settled","payload":{"storeId":"store-1","returnId":"ret-1","receiptId":"rcpt-1","totalMinor":50000,"paymentIds":["pay-1"],"cashMovementId":"cash-1","settledAt":"` + settledAt + `","actorId":"cashier-1"}}` +
		`]}`)
	syncRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/sync-events", syncBody)
	syncRequest.Header.Set("Content-Type", "application/json")
	syncRequest.Header.Set("Idempotency-Key", "sync-return-http")
	syncResponse := httptest.NewRecorder()
	server.ServeHTTP(syncResponse, syncRequest)
	if syncResponse.Code != http.StatusAccepted {
		t.Fatalf("sync status = %d body=%s", syncResponse.Code, syncResponse.Body.String())
	}

	returnResponse := httptest.NewRecorder()
	returnRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/returns/ret-1", nil)
	server.ServeHTTP(returnResponse, returnRequest)
	if returnResponse.Code != http.StatusOK {
		t.Fatalf("get return status = %d body=%s", returnResponse.Code, returnResponse.Body.String())
	}

	var ret api.SyncedReturnResponse
	if err := json.Unmarshal(returnResponse.Body.Bytes(), &ret); err != nil {
		t.Fatalf("decode return: %v", err)
	}
	if ret.TotalMinor != 50000 || ret.CashMovementID != "cash-1" || len(ret.PaymentIDs) != 1 {
		t.Fatalf("return = %+v", ret)
	}
}

func TestSyncEventsProjectOperationalDay(t *testing.T) {
	server := newTestServer()

	loginAdminAndRegisterStore(t, server, `{"storeId":"store-1","name":"Main Street"}`, "register-od-http")

	closedAt := time.Date(2026, 6, 19, 23, 0, 0, 0, time.UTC).Format(time.RFC3339)
	syncBody := bytes.NewBufferString(`{"events":[` +
		`{"eventId":"obx-od-1","eventType":"operational_day.closed","payload":{"storeId":"store-1","operationalDayId":"od-1","businessDate":"2026-06-19","closedById":"manager-1","closedAt":"` + closedAt + `"}}` +
		`]}`)
	syncRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/sync-events", syncBody)
	syncRequest.Header.Set("Content-Type", "application/json")
	syncRequest.Header.Set("Idempotency-Key", "sync-od-http")
	syncResponse := httptest.NewRecorder()
	server.ServeHTTP(syncResponse, syncRequest)
	if syncResponse.Code != http.StatusAccepted {
		t.Fatalf("sync status = %d body=%s", syncResponse.Code, syncResponse.Body.String())
	}

	dayResponse := httptest.NewRecorder()
	dayRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/operational-days/od-1", nil)
	server.ServeHTTP(dayResponse, dayRequest)
	if dayResponse.Code != http.StatusOK {
		t.Fatalf("get operational day status = %d body=%s", dayResponse.Code, dayResponse.Body.String())
	}

	var day api.SyncedOperationalDayResponse
	if err := json.Unmarshal(dayResponse.Body.Bytes(), &day); err != nil {
		t.Fatalf("decode operational day: %v", err)
	}
	if day.BusinessDate != "2026-06-19" || day.ClosedByID != "manager-1" {
		t.Fatalf("operational day = %+v", day)
	}
}

func TestStoreReportingSummaryEndpoint(t *testing.T) {
	server := newTestServer()

	loginAdminAndRegisterStore(t, server, `{"storeId":"store-1","name":"Main Street"}`, "register-reporting-http")

	capturedAt := time.Date(2026, 6, 19, 14, 30, 0, 0, time.UTC).Format(time.RFC3339)
	fiscalizedAt := time.Date(2026, 6, 19, 15, 0, 0, 0, time.UTC).Format(time.RFC3339)
	syncBody := bytes.NewBufferString(`{"events":[` +
		`{"eventId":"obx-pay-1","eventType":"payment.captured","payload":{"storeId":"store-1","paymentId":"pay-1","receiptId":"rcpt-1","method":"card","amountMinor":150000,"capturedAt":"` + capturedAt + `"}},` +
		`{"eventId":"obx-fisc-1","eventType":"fiscal.document.created","payload":{"storeId":"store-1","fiscalDocumentId":"fisc-1","receiptId":"rcpt-1","kind":"receipt","amountMinor":150000,"deviceId":"kkt-1","fiscalSign":"sign-abc","fiscalizedAt":"` + fiscalizedAt + `"}}` +
		`]}`)
	syncRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/sync-events", syncBody)
	syncRequest.Header.Set("Content-Type", "application/json")
	syncRequest.Header.Set("Idempotency-Key", "sync-reporting-http")
	syncResponse := httptest.NewRecorder()
	server.ServeHTTP(syncResponse, syncRequest)
	if syncResponse.Code != http.StatusAccepted {
		t.Fatalf("sync status = %d body=%s", syncResponse.Code, syncResponse.Body.String())
	}

	since := time.Date(2026, 6, 19, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	until := time.Date(2026, 6, 19, 23, 59, 59, 0, time.UTC).Format(time.RFC3339)
	token := loginTestSession(t, server, "admin@example.com", "admin-pass")
	summaryResponse := httptest.NewRecorder()
	summaryRequest := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/reporting/summary?since="+since+"&until="+until, nil)
	summaryRequest.Header.Set("X-Session-Token", token)
	server.ServeHTTP(summaryResponse, summaryRequest)
	if summaryResponse.Code != http.StatusOK {
		t.Fatalf("reporting summary status = %d body=%s", summaryResponse.Code, summaryResponse.Body.String())
	}

	var summary api.StoreReportingSummaryResponse
	if err := json.Unmarshal(summaryResponse.Body.Bytes(), &summary); err != nil {
		t.Fatalf("decode reporting summary: %v", err)
	}
	if summary.FiscalReceiptCount != 1 || summary.FiscalReceiptAmountMinor != 150000 {
		t.Fatalf("fiscal summary = %+v", summary)
	}
	if summary.PaymentsCapturedAmountMinor != 150000 {
		t.Fatalf("payment summary = %+v", summary)
	}
}

func TestCentralReportingSummaryEndpoint(t *testing.T) {
	server := newTestServer()
	adminToken := loginTestSession(t, server, "admin@example.com", "admin-pass")

	for _, fixture := range []struct {
		body           string
		idempotencyKey string
	}{
		{`{"storeId":"store-west","name":"West","region":"west"}`, "register-west-http"},
		{`{"storeId":"store-east","name":"East","region":"east"}`, "register-east-http"},
	} {
		registerTestStore(t, server, fixture.body, fixture.idempotencyKey, registerTestStoreAuth{sessionToken: adminToken})
	}

	capturedAt := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC).Format(time.RFC3339)
	for _, fixture := range []struct {
		storeID        string
		paymentID      string
		eventID        string
		amount         int
		idempotencyKey string
	}{
		{storeID: "store-west", paymentID: "pay-west", eventID: "obx-west-http", amount: 100000, idempotencyKey: "sync-west-http"},
		{storeID: "store-east", paymentID: "pay-east", eventID: "obx-east-http", amount: 200000, idempotencyKey: "sync-east-http"},
	} {
		syncBody := bytes.NewBufferString(`{"events":[` +
			`{"eventId":"` + fixture.eventID + `","eventType":"payment.captured","payload":{"storeId":"` + fixture.storeID + `","paymentId":"` + fixture.paymentID + `","receiptId":"rcpt-1","method":"card","amountMinor":` + fmt.Sprintf("%d", fixture.amount) + `,"capturedAt":"` + capturedAt + `"}}` +
			`]}`)
		syncRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/"+fixture.storeID+"/sync-events", syncBody)
		syncRequest.Header.Set("Content-Type", "application/json")
		syncRequest.Header.Set("Idempotency-Key", fixture.idempotencyKey)
		syncResponse := httptest.NewRecorder()
		server.ServeHTTP(syncResponse, syncRequest)
		if syncResponse.Code != http.StatusAccepted {
			t.Fatalf("sync status = %d body=%s", syncResponse.Code, syncResponse.Body.String())
		}
	}

	since := time.Date(2026, 6, 19, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	until := time.Date(2026, 6, 19, 23, 59, 59, 0, time.UTC).Format(time.RFC3339)
	token := adminToken

	centralResponse := httptest.NewRecorder()
	centralRequest := httptest.NewRequest(http.MethodGet, "/v1/central/reporting/summary?since="+since+"&until="+until, nil)
	centralRequest.Header.Set("X-Session-Token", token)
	server.ServeHTTP(centralResponse, centralRequest)
	if centralResponse.Code != http.StatusOK {
		t.Fatalf("central summary status = %d body=%s", centralResponse.Code, centralResponse.Body.String())
	}

	var central api.CentralReportingSummaryResponse
	if err := json.Unmarshal(centralResponse.Body.Bytes(), &central); err != nil {
		t.Fatalf("decode central summary: %v", err)
	}
	if central.StoreCount != 2 || central.PaymentsCapturedAmountMinor != 300000 {
		t.Fatalf("central summary = %+v", central)
	}

	westResponse := httptest.NewRecorder()
	westRequest := httptest.NewRequest(http.MethodGet, "/v1/central/reporting/summary?since="+since+"&until="+until+"&region=west", nil)
	westRequest.Header.Set("X-Session-Token", token)
	server.ServeHTTP(westResponse, westRequest)
	if westResponse.Code != http.StatusOK {
		t.Fatalf("west summary status = %d body=%s", westResponse.Code, westResponse.Body.String())
	}

	var west api.CentralReportingSummaryResponse
	if err := json.Unmarshal(westResponse.Body.Bytes(), &west); err != nil {
		t.Fatalf("decode west summary: %v", err)
	}
	if west.StoreCount != 1 || west.PaymentsCapturedAmountMinor != 100000 {
		t.Fatalf("west summary = %+v", west)
	}

	storesResponse := httptest.NewRecorder()
	storesRequest := httptest.NewRequest(http.MethodGet, "/v1/central/reporting/stores?since="+since+"&until="+until, nil)
	storesRequest.Header.Set("X-Session-Token", token)
	server.ServeHTTP(storesResponse, storesRequest)
	if storesResponse.Code != http.StatusOK {
		t.Fatalf("stores summary status = %d body=%s", storesResponse.Code, storesResponse.Body.String())
	}

	var stores api.PaginatedStoreReportingSummariesResponse
	if err := json.Unmarshal(storesResponse.Body.Bytes(), &stores); err != nil {
		t.Fatalf("decode stores summaries: %v", err)
	}
	if stores.TotalCount != 2 || len(stores.Items) != 2 {
		t.Fatalf("stores summaries = %+v", stores)
	}
}

func TestReportingRequiresSession(t *testing.T) {
	server := newTestServerWithoutSeed()

	since := time.Date(2026, 6, 19, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	until := time.Date(2026, 6, 19, 23, 59, 59, 0, time.UTC).Format(time.RFC3339)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/v1/central/reporting/summary?since="+since+"&until="+until, nil)
	server.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestCentralAuthAndUserManagementHTTP(t *testing.T) {
	store := memory.NewStore()
	if err := seedHTTPTestAdmin(store); err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	if err := seedHTTPTestViewer(store); err != nil {
		t.Fatalf("seed viewer: %v", err)
	}
	server := api.NewServerWithServices(newTestServices(store))

	adminToken := loginTestSession(t, server, "admin@example.com", "admin-pass")
	viewerToken := loginTestSession(t, server, "viewer@example.com", "viewer-pass")

	createBody := bytes.NewBufferString(`{"userId":"manager-1","email":"manager@example.com","password":"manager-pass","roles":["central_admin"]}`)
	createRequest := httptest.NewRequest(http.MethodPost, "/v1/central/users", createBody)
	createRequest.Header.Set("Content-Type", "application/json")
	createRequest.Header.Set("X-Session-Token", viewerToken)
	createResponse := httptest.NewRecorder()
	server.ServeHTTP(createResponse, createRequest)
	if createResponse.Code != http.StatusForbidden {
		t.Fatalf("viewer create status = %d body=%s", createResponse.Code, createResponse.Body.String())
	}

	createRequest.Header.Set("X-Session-Token", adminToken)
	createResponse = httptest.NewRecorder()
	server.ServeHTTP(createResponse, createRequest)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("admin create status = %d body=%s", createResponse.Code, createResponse.Body.String())
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/v1/central/users", nil)
	listRequest.Header.Set("X-Session-Token", adminToken)
	listResponse := httptest.NewRecorder()
	server.ServeHTTP(listResponse, listRequest)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("list users status = %d body=%s", listResponse.Code, listResponse.Body.String())
	}

	since := time.Date(2026, 6, 19, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	until := time.Date(2026, 6, 19, 23, 59, 59, 0, time.UTC).Format(time.RFC3339)
	reportingRequest := httptest.NewRequest(http.MethodGet, "/v1/central/reporting/summary?since="+since+"&until="+until, nil)
	reportingRequest.Header.Set("X-Session-Token", viewerToken)
	reportingResponse := httptest.NewRecorder()
	server.ServeHTTP(reportingResponse, reportingRequest)
	if reportingResponse.Code != http.StatusOK {
		t.Fatalf("viewer reporting status = %d body=%s", reportingResponse.Code, reportingResponse.Body.String())
	}
}

func TestSyncEventsRequireAPIKeyWhenConfigured(t *testing.T) {
	store := memory.NewStore()
	server := api.NewServerWithServices(newTestServicesWithSyncAPIKey(store, "test-key"))

	registerTestStore(t, server, `{"storeId":"store-1","name":"Main Street"}`, "register-sync-key", registerTestStoreAuth{syncAPIKey: "test-key"})

	syncBody := bytes.NewBufferString(`{"events":[{"eventId":"evt-key-1","eventType":"catalog.product.upserted","payload":{"productId":"sku-1","name":"Milk","barcodes":["4600000000000"],"unitPriceMinor":19999,"taxCategoryId":"vat_20"}}]}`)

	unauthorizedRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/sync-events", syncBody)
	unauthorizedRequest.Header.Set("Content-Type", "application/json")
	unauthorizedRequest.Header.Set("Idempotency-Key", "sync-key-unauthorized")
	unauthorizedResponse := httptest.NewRecorder()
	server.ServeHTTP(unauthorizedResponse, unauthorizedRequest)
	if unauthorizedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("sync without key status = %d body=%s", unauthorizedResponse.Code, unauthorizedResponse.Body.String())
	}

	authorizedRequest := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/sync-events", bytes.NewBufferString(`{"events":[{"eventId":"evt-key-1","eventType":"catalog.product.upserted","payload":{"productId":"sku-1","name":"Milk","barcodes":["4600000000000"],"unitPriceMinor":19999,"taxCategoryId":"vat_20"}}]}`))
	authorizedRequest.Header.Set("Content-Type", "application/json")
	authorizedRequest.Header.Set("Idempotency-Key", "sync-key-authorized")
	authorizedRequest.Header.Set("X-Sync-Api-Key", "test-key")
	authorizedResponse := httptest.NewRecorder()
	server.ServeHTTP(authorizedResponse, authorizedRequest)
	if authorizedResponse.Code != http.StatusAccepted {
		t.Fatalf("sync with key status = %d body=%s", authorizedResponse.Code, authorizedResponse.Body.String())
	}
}

type registerTestStoreAuth struct {
	sessionToken string
	syncAPIKey   string
}

func registerTestStore(t *testing.T, server http.Handler, body, idempotencyKey string, auth registerTestStoreAuth) {
	t.Helper()

	registerRequest := httptest.NewRequest(http.MethodPost, "/v1/stores", bytes.NewBufferString(body))
	registerRequest.Header.Set("Content-Type", "application/json")
	registerRequest.Header.Set("Idempotency-Key", idempotencyKey)
	if auth.sessionToken != "" {
		registerRequest.Header.Set("X-Session-Token", auth.sessionToken)
	}
	if auth.syncAPIKey != "" {
		registerRequest.Header.Set("X-Sync-Api-Key", auth.syncAPIKey)
	}
	registerResponse := httptest.NewRecorder()
	server.ServeHTTP(registerResponse, registerRequest)
	if registerResponse.Code != http.StatusAccepted {
		t.Fatalf("register status = %d body=%s", registerResponse.Code, registerResponse.Body.String())
	}
}

func TestRegisterStoreRequiresAdminSessionWhenSyncKeyUnset(t *testing.T) {
	store := memory.NewStore()
	if err := seedHTTPTestAdmin(store); err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	if err := seedHTTPTestViewer(store); err != nil {
		t.Fatalf("seed viewer: %v", err)
	}
	server := api.NewServerWithServices(newTestServices(store))
	body := `{"storeId":"store-1","name":"Main Street","region":"west"}`

	unauthorizedRequest := httptest.NewRequest(http.MethodPost, "/v1/stores", bytes.NewBufferString(body))
	unauthorizedRequest.Header.Set("Content-Type", "application/json")
	unauthorizedRequest.Header.Set("Idempotency-Key", "register-unauthorized")
	unauthorizedResponse := httptest.NewRecorder()
	server.ServeHTTP(unauthorizedResponse, unauthorizedRequest)
	if unauthorizedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("register without auth status = %d body=%s", unauthorizedResponse.Code, unauthorizedResponse.Body.String())
	}

	viewerToken := loginTestSession(t, server, "viewer@example.com", "viewer-pass")
	forbiddenRequest := httptest.NewRequest(http.MethodPost, "/v1/stores", bytes.NewBufferString(body))
	forbiddenRequest.Header.Set("Content-Type", "application/json")
	forbiddenRequest.Header.Set("Idempotency-Key", "register-forbidden")
	forbiddenRequest.Header.Set("X-Session-Token", viewerToken)
	forbiddenResponse := httptest.NewRecorder()
	server.ServeHTTP(forbiddenResponse, forbiddenRequest)
	if forbiddenResponse.Code != http.StatusForbidden {
		t.Fatalf("register viewer status = %d body=%s", forbiddenResponse.Code, forbiddenResponse.Body.String())
	}

	adminToken := loginTestSession(t, server, "admin@example.com", "admin-pass")
	registerTestStore(t, server, body, "register-admin", registerTestStoreAuth{sessionToken: adminToken})
}

func TestRegisterStoreAcceptsSyncAPIKeyWhenConfigured(t *testing.T) {
	store := memory.NewStore()
	server := api.NewServerWithServices(newTestServicesWithSyncAPIKey(store, "test-key"))
	body := `{"storeId":"store-1","name":"Main Street"}`

	unauthorizedRequest := httptest.NewRequest(http.MethodPost, "/v1/stores", bytes.NewBufferString(body))
	unauthorizedRequest.Header.Set("Content-Type", "application/json")
	unauthorizedRequest.Header.Set("Idempotency-Key", "register-no-key")
	unauthorizedResponse := httptest.NewRecorder()
	server.ServeHTTP(unauthorizedResponse, unauthorizedRequest)
	if unauthorizedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("register without key status = %d body=%s", unauthorizedResponse.Code, unauthorizedResponse.Body.String())
	}

	registerTestStore(t, server, body, "register-with-key", registerTestStoreAuth{syncAPIKey: "test-key"})
}

func TestRegisterStoreRejectsAdminSessionWhenSyncKeyConfigured(t *testing.T) {
	store := memory.NewStore()
	if err := seedHTTPTestAdmin(store); err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	server := api.NewServerWithServices(newTestServicesWithSyncAPIKey(store, "test-key"))
	adminToken := loginTestSession(t, server, "admin@example.com", "admin-pass")

	request := httptest.NewRequest(http.MethodPost, "/v1/stores", bytes.NewBufferString(`{"storeId":"store-1","name":"Main Street"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "register-admin-without-key")
	request.Header.Set("X-Session-Token", adminToken)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("register with session only status = %d body=%s", response.Code, response.Body.String())
	}
}

func TestListStoresRequiresSession(t *testing.T) {
	store := memory.NewStore()
	if err := seedHTTPTestAdmin(store); err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	if err := seedHTTPTestViewer(store); err != nil {
		t.Fatalf("seed viewer: %v", err)
	}
	server := api.NewServerWithServices(newTestServices(store))
	adminToken := loginAdminAndRegisterStore(t, server, `{"storeId":"store-1","name":"Main Street"}`, "register-list-stores")

	unauthorizedRequest := httptest.NewRequest(http.MethodGet, "/v1/stores", nil)
	unauthorizedResponse := httptest.NewRecorder()
	server.ServeHTTP(unauthorizedResponse, unauthorizedRequest)
	if unauthorizedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("list without auth status = %d body=%s", unauthorizedResponse.Code, unauthorizedResponse.Body.String())
	}

	viewerToken := loginTestSession(t, server, "viewer@example.com", "viewer-pass")
	viewerRequest := httptest.NewRequest(http.MethodGet, "/v1/stores", nil)
	viewerRequest.Header.Set("X-Session-Token", viewerToken)
	viewerResponse := httptest.NewRecorder()
	server.ServeHTTP(viewerResponse, viewerRequest)
	if viewerResponse.Code != http.StatusOK {
		t.Fatalf("list viewer status = %d body=%s", viewerResponse.Code, viewerResponse.Body.String())
	}

	adminRequest := httptest.NewRequest(http.MethodGet, "/v1/stores", nil)
	adminRequest.Header.Set("X-Session-Token", adminToken)
	adminResponse := httptest.NewRecorder()
	server.ServeHTTP(adminResponse, adminRequest)
	if adminResponse.Code != http.StatusOK {
		t.Fatalf("list admin status = %d body=%s", adminResponse.Code, adminResponse.Body.String())
	}

	var listed api.StoresResponse
	if err := json.Unmarshal(adminResponse.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode stores list: %v", err)
	}
	if len(listed.Stores) != 1 || listed.Stores[0].ID != "store-1" {
		t.Fatalf("listed stores = %+v", listed.Stores)
	}
}

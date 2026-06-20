package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"mercadia.dev/pos/platform/httpapi"
	"mercadia.dev/pos/services/central-backend/internal/app"
	"mercadia.dev/pos/services/central-backend/internal/domain"
	"mercadia.dev/pos/services/central-backend/internal/infra"
	"mercadia.dev/pos/services/central-backend/internal/infra/memory"
)

const version = "0.1.0"

type Services struct {
	StoreRegistry   *app.StoreRegistryService
	Sync            *app.SyncService
	Catalog         *app.CatalogService
	Payments        *app.PaymentsService
	CashMovements   *app.CashMovementsService
	FiscalDocuments *app.FiscalDocumentsService
	Returns         *app.ReturnsService
	OperationalDays *app.OperationalDaysService
	Reporting       *app.ReportingService
	Auth            *app.AuthService
	CentralUsers    *app.CentralUsersService
	SyncAPIKey      *app.SyncAPIKeyService
}

type StatusResponse struct {
	Region      string    `json:"region"`
	Status      string    `json:"status"`
	StoreCount  int       `json:"storeCount"`
	GeneratedAt time.Time `json:"generatedAt"`
}

type StoreResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Region       string    `json:"region"`
	RegisteredAt time.Time `json:"registeredAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type StoresResponse struct {
	Stores []StoreResponse `json:"stores"`
}

type RegisterStoreRequest struct {
	StoreID string `json:"storeId"`
	Name    string `json:"name"`
	Region  string `json:"region,omitempty"`
}

type StoreAcceptedResponse struct {
	Store StoreResponse `json:"store"`
}

type SyncEventRequest struct {
	EventID    string          `json:"eventId"`
	EventType  string          `json:"eventType"`
	OccurredAt time.Time       `json:"occurredAt,omitempty"`
	Payload    json.RawMessage `json:"payload,omitempty"`
}

type AcceptSyncEventsRequest struct {
	Events []SyncEventRequest `json:"events"`
}

type SyncEventsAcceptedResponse struct {
	StoreID  string `json:"storeId"`
	Status   string `json:"status"`
	Accepted int    `json:"accepted"`
}

type SyncEventResponse struct {
	EventID       string          `json:"eventId"`
	EventType     string          `json:"eventType"`
	SourceEventID string          `json:"sourceEventId"`
	OccurredAt    time.Time       `json:"occurredAt"`
	ReceivedAt    time.Time       `json:"receivedAt"`
	Payload       json.RawMessage `json:"payload"`
}

type PaginatedSyncEventsResponse struct {
	Items      []SyncEventResponse `json:"items"`
	TotalCount int                 `json:"totalCount"`
}

type CatalogProductResponse struct {
	ID             string    `json:"id"`
	StoreID        string    `json:"storeId"`
	Name           string    `json:"name"`
	Barcodes       []string  `json:"barcodes"`
	UnitPriceMinor int64     `json:"unitPriceMinor"`
	TaxCategoryID  string    `json:"taxCategoryId"`
	Active         bool      `json:"active"`
	Version        int64     `json:"version"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type CatalogProductsResponse struct {
	Products []CatalogProductResponse `json:"products"`
}

type CatalogDeltaResponse struct {
	Since    time.Time                `json:"since"`
	Products []CatalogProductResponse `json:"products"`
}

type SyncedPaymentResponse struct {
	ID                   string     `json:"id"`
	StoreID              string     `json:"storeId"`
	ReceiptID            string     `json:"receiptId"`
	Method               string     `json:"method"`
	AmountMinor          int64      `json:"amountMinor"`
	Status               string     `json:"status"`
	CapturedAt           time.Time  `json:"capturedAt"`
	CancelledAt          *time.Time `json:"cancelledAt,omitempty"`
	RefundedAmountMinor  int64      `json:"refundedAmountMinor"`
	RemainingAmountMinor int64      `json:"remainingAmountMinor"`
	SourceEventID        string     `json:"sourceEventId"`
	LastEventID          string     `json:"lastEventId"`
	SyncedAt             time.Time  `json:"syncedAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

type PaginatedSyncedPaymentsResponse struct {
	Items      []SyncedPaymentResponse `json:"items"`
	TotalCount int                     `json:"totalCount"`
}

type SyncedCashMovementResponse struct {
	ID                string    `json:"id"`
	StoreID           string    `json:"storeId"`
	Type              string    `json:"type"`
	FromContainerID   string    `json:"fromContainerId"`
	FromContainerType string    `json:"fromContainerType"`
	ToContainerID     string    `json:"toContainerId"`
	ToContainerType   string    `json:"toContainerType"`
	AmountMinor       int64     `json:"amountMinor"`
	Currency          string    `json:"currency"`
	ActorID           string    `json:"actorId"`
	PostedAt          time.Time `json:"postedAt"`
	SourceEventID     string    `json:"sourceEventId"`
	SyncedAt          time.Time `json:"syncedAt"`
}

type PaginatedSyncedCashMovementsResponse struct {
	Items      []SyncedCashMovementResponse `json:"items"`
	TotalCount int                          `json:"totalCount"`
}

type SyncedFiscalDocumentResponse struct {
	ID            string    `json:"id"`
	StoreID       string    `json:"storeId"`
	ReceiptID     string    `json:"receiptId"`
	Kind          string    `json:"kind"`
	AmountMinor   int64     `json:"amountMinor"`
	DeviceID      string    `json:"deviceId"`
	FiscalSign    string    `json:"fiscalSign"`
	FiscalizedAt  time.Time `json:"fiscalizedAt"`
	ReturnID      string    `json:"returnId,omitempty"`
	SourceEventID string    `json:"sourceEventId"`
	SyncedAt      time.Time `json:"syncedAt"`
}

type PaginatedSyncedFiscalDocumentsResponse struct {
	Items      []SyncedFiscalDocumentResponse `json:"items"`
	TotalCount int                            `json:"totalCount"`
}

type SyncedReturnResponse struct {
	ID             string    `json:"id"`
	StoreID        string    `json:"storeId"`
	ReceiptID      string    `json:"receiptId"`
	TotalMinor     int64     `json:"totalMinor"`
	PaymentIDs     []string  `json:"paymentIds"`
	CashMovementID string    `json:"cashMovementId,omitempty"`
	ActorID        string    `json:"actorId"`
	SettledAt      time.Time `json:"settledAt"`
	SourceEventID  string    `json:"sourceEventId"`
	SyncedAt       time.Time `json:"syncedAt"`
}

type PaginatedSyncedReturnsResponse struct {
	Items      []SyncedReturnResponse `json:"items"`
	TotalCount int                    `json:"totalCount"`
}

type SyncedOperationalDayResponse struct {
	ID            string    `json:"id"`
	StoreID       string    `json:"storeId"`
	BusinessDate  string    `json:"businessDate"`
	ClosedByID    string    `json:"closedById"`
	ClosedAt      time.Time `json:"closedAt"`
	SourceEventID string    `json:"sourceEventId"`
	SyncedAt      time.Time `json:"syncedAt"`
}

type PaginatedSyncedOperationalDaysResponse struct {
	Items      []SyncedOperationalDayResponse `json:"items"`
	TotalCount int                            `json:"totalCount"`
}

type StoreReportingSummaryResponse struct {
	StoreID                     string    `json:"storeId"`
	Since                       time.Time `json:"since"`
	Until                       time.Time `json:"until"`
	FiscalReceiptCount          int       `json:"fiscalReceiptCount"`
	FiscalReceiptAmountMinor    int64     `json:"fiscalReceiptAmountMinor"`
	FiscalReturnCount           int       `json:"fiscalReturnCount"`
	FiscalReturnAmountMinor     int64     `json:"fiscalReturnAmountMinor"`
	PaymentsCapturedAmountMinor int64     `json:"paymentsCapturedAmountMinor"`
	PaymentsCancelledCount      int       `json:"paymentsCancelledCount"`
	PaymentsRefundedAmountMinor int64     `json:"paymentsRefundedAmountMinor"`
	ReturnsSettledCount         int       `json:"returnsSettledCount"`
	ReturnsSettledAmountMinor   int64     `json:"returnsSettledAmountMinor"`
	CashMovementsPostedCount    int       `json:"cashMovementsPostedCount"`
	OperationalDaysClosedCount  int       `json:"operationalDaysClosedCount"`
}

type CentralReportingSummaryResponse struct {
	Region                      string    `json:"region,omitempty"`
	Since                       time.Time `json:"since"`
	Until                       time.Time `json:"until"`
	StoreCount                  int       `json:"storeCount"`
	FiscalReceiptCount          int       `json:"fiscalReceiptCount"`
	FiscalReceiptAmountMinor    int64     `json:"fiscalReceiptAmountMinor"`
	FiscalReturnCount           int       `json:"fiscalReturnCount"`
	FiscalReturnAmountMinor     int64     `json:"fiscalReturnAmountMinor"`
	PaymentsCapturedAmountMinor int64     `json:"paymentsCapturedAmountMinor"`
	PaymentsCancelledCount      int       `json:"paymentsCancelledCount"`
	PaymentsRefundedAmountMinor int64     `json:"paymentsRefundedAmountMinor"`
	ReturnsSettledCount         int       `json:"returnsSettledCount"`
	ReturnsSettledAmountMinor   int64     `json:"returnsSettledAmountMinor"`
	CashMovementsPostedCount    int       `json:"cashMovementsPostedCount"`
	OperationalDaysClosedCount  int       `json:"operationalDaysClosedCount"`
}

type PaginatedStoreReportingSummariesResponse struct {
	Items      []StoreReportingSummaryResponse `json:"items"`
	TotalCount int                             `json:"totalCount"`
}

type ServerOptions struct {
	ReadinessChecks []func(context.Context) error
}

func NewServer() http.Handler {
	handle, err := infra.NewHandle(context.Background())
	if err != nil {
		panic(err)
	}
	_ = handle
	return NewServerWithServices(newServices(handle.Repository()))
}

type ServerBundle struct {
	Handler  http.Handler
	Services Services
	Handle   infra.Handle
}

func NewServerBundle(opts ServerOptions) (*ServerBundle, error) {
	handle, err := infra.NewHandle(context.Background())
	if err != nil {
		return nil, err
	}
	services := newServices(handle.Repository())
	if config, ok := app.SeedCentralAdminConfigFromEnv(); ok {
		if err := app.BootstrapSeedCentralAdmin(context.Background(), handle.Repository(), config); err != nil {
			return nil, err
		}
	}
	return &ServerBundle{
		Handler:  newHandler(services, opts),
		Services: services,
		Handle:   handle,
	}, nil
}

func newServices(repo infra.Repository) Services {
	return Services{
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
		SyncAPIKey:      app.NewSyncAPIKeyServiceFromEnv(),
	}
}

func NewServerWithServices(services Services) http.Handler {
	return NewHandler(services, ServerOptions{})
}

func NewHandler(services Services, opts ServerOptions) http.Handler {
	return newHandler(services, opts)
}

func OpenAPI() map[string]any {
	_, spec := newMuxAndSpec(newServices(memory.NewStore()), nil)
	return spec.OpenAPI()
}

func newHandler(services Services, opts ServerOptions) http.Handler {
	mux, _ := newMuxAndSpec(services, opts.ReadinessChecks)
	return mux
}

func newMuxAndSpec(services Services, readinessChecks []func(context.Context) error) (*http.ServeMux, *httpapi.Spec) {
	info := httpapi.ServiceInfo{
		Name:        "central-backend",
		Title:       "Mercadia Central Backend",
		Description: "Central API for global administration, cross-store reporting, integrations, and Store Edge synchronization.",
		Version:     version,
	}

	mux := http.NewServeMux()
	spec := httpapi.NewSpec(info)
	var systemOptions []httpapi.SystemRoutesOption
	if len(readinessChecks) > 0 {
		systemOptions = append(systemOptions, httpapi.WithReadinessCheck(combineReadinessChecks(readinessChecks)))
	}
	httpapi.MountSystemRoutes(mux, spec, info, systemOptions...)
	mountAuthRoutes(mux, spec, services)
	mountRoutes(mux, spec, services)

	return mux, spec
}

func combineReadinessChecks(checks []func(context.Context) error) func(context.Context) error {
	return func(ctx context.Context) error {
		for _, check := range checks {
			if check == nil {
				continue
			}
			if err := check(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}

func mountRoutes(mux *http.ServeMux, spec *httpapi.Spec, services Services) {
	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/central/status",
		OperationID: "getCentralStatus",
		Summary:     "Get central backend status",
		Description: sessionProtectedDescription("Returns region status and registered store count."),
		Tags:        []string{"system"},
		Responses:   protectedResponseSpecs("200", "Central backend status", statusResponseSchema()),
	}, RequireSession(services.Auth, app.PermissionReportingRead, func(w http.ResponseWriter, r *http.Request) {
		count, err := services.StoreRegistry.CountStores(r.Context())
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, StatusResponse{
			Region:      regionName(),
			Status:      "ok",
			StoreCount:  count,
			GeneratedAt: time.Now().UTC(),
		})
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/stores",
		OperationID: "listStores",
		Summary:     "List registered stores",
		Description: sessionProtectedDescription("Lists stores registered in the central registry."),
		Tags:        []string{"stores"},
		Responses:   protectedResponseSpecs("200", "Registered stores", storesResponseSchema()),
	}, RequireSession(services.Auth, app.PermissionReportingRead, func(w http.ResponseWriter, r *http.Request) {
		stores, err := services.StoreRegistry.ListStores(r.Context())
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, StoresResponse{Stores: storeResponses(stores)})
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/stores",
		OperationID:         "registerStore",
		Summary:             "Register a store",
		Description:         storeRegistrationProtectedDescription("Registers a store in the central registry."),
		Tags:                []string{"stores"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Store registration command",
			Required:    true,
			Schema:      registerStoreRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Store registered", Schema: storeAcceptedResponseSchema()},
			"400": {Description: "Invalid store command", Schema: httpapi.ProblemSchema()},
			"401": {Description: "Sync API key or session is missing or invalid", Schema: httpapi.ProblemSchema()},
			"403": {Description: "Permission denied", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Idempotency conflict", Schema: httpapi.ProblemSchema()},
		},
	}, RequireSyncAPIKeyOrSession(services.Auth, services.SyncAPIKey, app.PermissionUsersManage, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request RegisterStoreRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_store_command", "Invalid store command", err.Error())
			return
		}
		result, err := services.StoreRegistry.RegisterStore(r.Context(), app.RegisterStoreCommand{
			StoreID:        request.StoreID,
			Name:           request.Name,
			Region:         request.Region,
			IdempotencyKey: r.Header.Get("Idempotency-Key"),
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, StoreAcceptedResponse{Store: storeResponse(result.Store)})
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/stores/{storeId}/sync-events",
		OperationID:         "acceptStoreSyncEvents",
		Summary:             "Accept synchronized Store Edge events",
		Description:         syncAPIKeyProtectedDescription("Accepts synchronized Store Edge event batches over HTTP."),
		Tags:                []string{"sync"},
		RequiresIdempotency: true,
		RequestBody: &httpapi.BodySpec{
			Description: "Synchronized event batch",
			Required:    true,
			Schema:      acceptSyncEventsRequestSchema(),
		},
		Responses: map[string]httpapi.ResponseSpec{
			"202": {Description: "Sync batch accepted", Schema: syncEventsAcceptedResponseSchema()},
			"400": {Description: "Invalid sync batch", Schema: httpapi.ProblemSchema()},
			"401": {Description: "Sync API key is missing or invalid", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Store not found", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Idempotency conflict", Schema: httpapi.ProblemSchema()},
		},
	}, RequireSyncAPIKey(services.SyncAPIKey, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}
		var request AcceptSyncEventsRequest
		if err := httpapi.DecodeJSON(r, &request); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_sync_command", "Invalid sync batch", err.Error())
			return
		}
		events := make([]app.SyncEventInput, 0, len(request.Events))
		for _, event := range request.Events {
			events = append(events, app.SyncEventInput{
				EventID:    event.EventID,
				EventType:  event.EventType,
				OccurredAt: event.OccurredAt,
				Payload:    event.Payload,
			})
		}
		result, err := services.Sync.AcceptEvents(r.Context(), app.AcceptSyncEventsCommand{
			StoreID:        r.PathValue("storeId"),
			IdempotencyKey: r.Header.Get("Idempotency-Key"),
			Events:         events,
		})
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusAccepted, SyncEventsAcceptedResponse{
			StoreID:  result.StoreID,
			Status:   result.Status,
			Accepted: result.Accepted,
		})
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:          http.MethodGet,
		Path:            "/v1/stores/{storeId}/sync-events",
		OperationID:     "listStoreSyncEvents",
		Summary:         "List synchronized Store Edge events",
		Description:     sessionProtectedDescription("Lists synchronized Store Edge events accepted for a store."),
		Tags:            []string{"sync"},
		QueryParameters: paginationQueryParams(),
		Responses: mergeResponseSpecs(
			protectedResponseSpecs("200", "Synchronized events", paginatedSyncEventsResponseSchema()),
			map[string]httpapi.ResponseSpec{
				"400": {Description: "Invalid list query", Schema: httpapi.ProblemSchema()},
				"404": {Description: "Store not found", Schema: httpapi.ProblemSchema()},
			},
		),
	}, RequireSession(services.Auth, app.PermissionReportingRead, func(w http.ResponseWriter, r *http.Request) {
		params := app.ParsePageParams(r.URL.Query().Get("limit"), r.URL.Query().Get("offset"))
		result, err := services.Sync.ListEvents(r.Context(), r.PathValue("storeId"), params)
		if err != nil {
			writeAppError(w, err)
			return
		}
		items := make([]SyncEventResponse, 0, len(result.Items))
		for _, event := range result.Items {
			items = append(items, syncEventResponse(event))
		}
		httpapi.WriteJSON(w, http.StatusOK, PaginatedSyncEventsResponse{
			Items:      items,
			TotalCount: result.TotalCount,
		})
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/stores/{storeId}/catalog/products",
		OperationID: "listStoreCatalogProducts",
		Summary:     "List catalog products for a store",
		Description: catalogListProtectedDescription("Lists catalog products for a store."),
		Tags:        []string{"catalog"},
		Responses: mergeResponseSpecs(
			protectedResponseSpecs("200", "Catalog products", catalogProductsResponseSchema()),
			map[string]httpapi.ResponseSpec{
				"404": {Description: "Store not found", Schema: httpapi.ProblemSchema()},
			},
		),
	}, RequireSyncAPIKeyOrSessionAuth(services.Auth, services.SyncAPIKey, app.PermissionReportingRead, func(w http.ResponseWriter, r *http.Request) {
		result, err := services.Catalog.ListProducts(r.Context(), r.PathValue("storeId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, CatalogProductsResponse{Products: catalogProductResponses(result.Products)})
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/stores/{storeId}/catalog/delta",
		OperationID: "getStoreCatalogDelta",
		Summary:     "Get catalog changes since a timestamp",
		Description: syncAPIKeyProtectedDescription("Query parameter `since` must be an RFC3339 timestamp."),
		Tags:        []string{"catalog"},
		Responses: mergeResponseSpecs(
			map[string]httpapi.ResponseSpec{
				"200": {Description: "Catalog delta", Schema: catalogDeltaResponseSchema()},
				"400": {Description: "Invalid catalog query", Schema: httpapi.ProblemSchema()},
				"404": {Description: "Store not found", Schema: httpapi.ProblemSchema()},
			},
			map[string]httpapi.ResponseSpec{
				"401": {Description: "Sync API key is missing or invalid", Schema: httpapi.ProblemSchema()},
			},
		),
	}, RequireSyncAPIKey(services.SyncAPIKey, func(w http.ResponseWriter, r *http.Request) {
		sinceRaw := r.URL.Query().Get("since")
		since, err := time.Parse(time.RFC3339, sinceRaw)
		if err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_catalog_query", "Invalid catalog query", "since must be RFC3339")
			return
		}
		result, err := services.Catalog.CatalogDelta(r.Context(), r.PathValue("storeId"), since)
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, CatalogDeltaResponse{
			Since:    result.Since,
			Products: catalogProductResponses(result.Products),
		})
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:          http.MethodGet,
		Path:            "/v1/stores/{storeId}/payments",
		OperationID:     "listStorePayments",
		Summary:         "List synchronized payments for a store",
		Description:     sessionProtectedDescription("Lists synchronized payments projected from Store Edge events."),
		Tags:            []string{"sync"},
		QueryParameters: paginationQueryParams(),
		Responses: mergeResponseSpecs(
			protectedResponseSpecs("200", "Synchronized payments", paginatedSyncedPaymentsResponseSchema()),
			map[string]httpapi.ResponseSpec{
				"400": {Description: "Invalid list query", Schema: httpapi.ProblemSchema()},
				"404": {Description: "Store not found", Schema: httpapi.ProblemSchema()},
			},
		),
	}, RequireSession(services.Auth, app.PermissionReportingRead, func(w http.ResponseWriter, r *http.Request) {
		params := app.ParsePageParams(r.URL.Query().Get("limit"), r.URL.Query().Get("offset"))
		result, err := services.Payments.ListPayments(r.Context(), r.PathValue("storeId"), params)
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, PaginatedSyncedPaymentsResponse{
			Items:      syncedPaymentResponses(result.Items),
			TotalCount: result.TotalCount,
		})
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/stores/{storeId}/payments/{paymentId}",
		OperationID: "getStorePayment",
		Summary:     "Get a synchronized payment",
		Description: sessionProtectedDescription("Returns a synchronized payment projected from Store Edge events."),
		Tags:        []string{"sync"},
		Responses: mergeResponseSpecs(
			protectedResponseSpecs("200", "Synchronized payment", syncedPaymentResponseSchema()),
			map[string]httpapi.ResponseSpec{
				"404": {Description: "Store or payment not found", Schema: httpapi.ProblemSchema()},
			},
		),
	}, RequireSession(services.Auth, app.PermissionReportingRead, func(w http.ResponseWriter, r *http.Request) {
		payment, err := services.Payments.GetPayment(r.Context(), r.PathValue("storeId"), r.PathValue("paymentId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, syncedPaymentResponse(payment))
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:          http.MethodGet,
		Path:            "/v1/stores/{storeId}/cash-movements",
		OperationID:     "listStoreCashMovements",
		Summary:         "List synchronized cash movements for a store",
		Description:     sessionProtectedDescription("Lists synchronized cash movements projected from Store Edge events."),
		Tags:            []string{"sync"},
		QueryParameters: paginationQueryParams(),
		Responses: mergeResponseSpecs(
			protectedResponseSpecs("200", "Synchronized cash movements", paginatedSyncedCashMovementsResponseSchema()),
			map[string]httpapi.ResponseSpec{
				"400": {Description: "Invalid list query", Schema: httpapi.ProblemSchema()},
				"404": {Description: "Store not found", Schema: httpapi.ProblemSchema()},
			},
		),
	}, RequireSession(services.Auth, app.PermissionReportingRead, func(w http.ResponseWriter, r *http.Request) {
		params := app.ParsePageParams(r.URL.Query().Get("limit"), r.URL.Query().Get("offset"))
		result, err := services.CashMovements.ListCashMovements(r.Context(), r.PathValue("storeId"), params)
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, PaginatedSyncedCashMovementsResponse{
			Items:      syncedCashMovementResponses(result.Items),
			TotalCount: result.TotalCount,
		})
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/stores/{storeId}/cash-movements/{cashMovementId}",
		OperationID: "getStoreCashMovement",
		Summary:     "Get a synchronized cash movement",
		Description: sessionProtectedDescription("Returns a synchronized cash movement projected from Store Edge events."),
		Tags:        []string{"sync"},
		Responses: mergeResponseSpecs(
			protectedResponseSpecs("200", "Synchronized cash movement", syncedCashMovementResponseSchema()),
			map[string]httpapi.ResponseSpec{
				"404": {Description: "Store or cash movement not found", Schema: httpapi.ProblemSchema()},
			},
		),
	}, RequireSession(services.Auth, app.PermissionReportingRead, func(w http.ResponseWriter, r *http.Request) {
		movement, err := services.CashMovements.GetCashMovement(r.Context(), r.PathValue("storeId"), r.PathValue("cashMovementId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, syncedCashMovementResponse(movement))
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:          http.MethodGet,
		Path:            "/v1/stores/{storeId}/fiscal-documents",
		OperationID:     "listStoreFiscalDocuments",
		Summary:         "List synchronized fiscal documents for a store",
		Description:     sessionProtectedDescription("Lists synchronized fiscal documents projected from Store Edge events."),
		Tags:            []string{"sync"},
		QueryParameters: paginationQueryParams(),
		Responses: mergeResponseSpecs(
			protectedResponseSpecs("200", "Synchronized fiscal documents", paginatedSyncedFiscalDocumentsResponseSchema()),
			map[string]httpapi.ResponseSpec{
				"400": {Description: "Invalid list query", Schema: httpapi.ProblemSchema()},
				"404": {Description: "Store not found", Schema: httpapi.ProblemSchema()},
			},
		),
	}, RequireSession(services.Auth, app.PermissionReportingRead, func(w http.ResponseWriter, r *http.Request) {
		params := app.ParsePageParams(r.URL.Query().Get("limit"), r.URL.Query().Get("offset"))
		result, err := services.FiscalDocuments.ListFiscalDocuments(r.Context(), r.PathValue("storeId"), params)
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, PaginatedSyncedFiscalDocumentsResponse{
			Items:      syncedFiscalDocumentResponses(result.Items),
			TotalCount: result.TotalCount,
		})
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/stores/{storeId}/fiscal-documents/{fiscalDocumentId}",
		OperationID: "getStoreFiscalDocument",
		Summary:     "Get a synchronized fiscal document",
		Description: sessionProtectedDescription("Returns a synchronized fiscal document projected from Store Edge events."),
		Tags:        []string{"sync"},
		Responses: mergeResponseSpecs(
			protectedResponseSpecs("200", "Synchronized fiscal document", syncedFiscalDocumentResponseSchema()),
			map[string]httpapi.ResponseSpec{
				"404": {Description: "Store or fiscal document not found", Schema: httpapi.ProblemSchema()},
			},
		),
	}, RequireSession(services.Auth, app.PermissionReportingRead, func(w http.ResponseWriter, r *http.Request) {
		document, err := services.FiscalDocuments.GetFiscalDocument(r.Context(), r.PathValue("storeId"), r.PathValue("fiscalDocumentId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, syncedFiscalDocumentResponse(document))
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:          http.MethodGet,
		Path:            "/v1/stores/{storeId}/returns",
		OperationID:     "listStoreReturns",
		Summary:         "List synchronized returns for a store",
		Description:     sessionProtectedDescription("Lists synchronized returns projected from Store Edge events."),
		Tags:            []string{"sync"},
		QueryParameters: paginationQueryParams(),
		Responses: mergeResponseSpecs(
			protectedResponseSpecs("200", "Synchronized returns", paginatedSyncedReturnsResponseSchema()),
			map[string]httpapi.ResponseSpec{
				"400": {Description: "Invalid list query", Schema: httpapi.ProblemSchema()},
				"404": {Description: "Store not found", Schema: httpapi.ProblemSchema()},
			},
		),
	}, RequireSession(services.Auth, app.PermissionReportingRead, func(w http.ResponseWriter, r *http.Request) {
		params := app.ParsePageParams(r.URL.Query().Get("limit"), r.URL.Query().Get("offset"))
		result, err := services.Returns.ListReturns(r.Context(), r.PathValue("storeId"), params)
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, PaginatedSyncedReturnsResponse{
			Items:      syncedReturnResponses(result.Items),
			TotalCount: result.TotalCount,
		})
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/stores/{storeId}/returns/{returnId}",
		OperationID: "getStoreReturn",
		Summary:     "Get a synchronized return",
		Description: sessionProtectedDescription("Returns a synchronized return projected from Store Edge events."),
		Tags:        []string{"sync"},
		Responses: mergeResponseSpecs(
			protectedResponseSpecs("200", "Synchronized return", syncedReturnResponseSchema()),
			map[string]httpapi.ResponseSpec{
				"404": {Description: "Store or return not found", Schema: httpapi.ProblemSchema()},
			},
		),
	}, RequireSession(services.Auth, app.PermissionReportingRead, func(w http.ResponseWriter, r *http.Request) {
		ret, err := services.Returns.GetReturn(r.Context(), r.PathValue("storeId"), r.PathValue("returnId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, syncedReturnResponse(ret))
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:          http.MethodGet,
		Path:            "/v1/stores/{storeId}/operational-days",
		OperationID:     "listStoreOperationalDays",
		Summary:         "List synchronized closed operational days for a store",
		Description:     sessionProtectedDescription("Lists synchronized closed operational days projected from Store Edge events."),
		Tags:            []string{"sync"},
		QueryParameters: paginationQueryParams(),
		Responses: mergeResponseSpecs(
			protectedResponseSpecs("200", "Synchronized operational days", paginatedSyncedOperationalDaysResponseSchema()),
			map[string]httpapi.ResponseSpec{
				"400": {Description: "Invalid list query", Schema: httpapi.ProblemSchema()},
				"404": {Description: "Store not found", Schema: httpapi.ProblemSchema()},
			},
		),
	}, RequireSession(services.Auth, app.PermissionReportingRead, func(w http.ResponseWriter, r *http.Request) {
		params := app.ParsePageParams(r.URL.Query().Get("limit"), r.URL.Query().Get("offset"))
		result, err := services.OperationalDays.ListOperationalDays(r.Context(), r.PathValue("storeId"), params)
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, PaginatedSyncedOperationalDaysResponse{
			Items:      syncedOperationalDayResponses(result.Items),
			TotalCount: result.TotalCount,
		})
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/stores/{storeId}/operational-days/{operationalDayId}",
		OperationID: "getStoreOperationalDay",
		Summary:     "Get a synchronized closed operational day",
		Description: sessionProtectedDescription("Returns a synchronized closed operational day projected from Store Edge events."),
		Tags:        []string{"sync"},
		Responses: mergeResponseSpecs(
			protectedResponseSpecs("200", "Synchronized operational day", syncedOperationalDayResponseSchema()),
			map[string]httpapi.ResponseSpec{
				"404": {Description: "Store or operational day not found", Schema: httpapi.ProblemSchema()},
			},
		),
	}, RequireSession(services.Auth, app.PermissionReportingRead, func(w http.ResponseWriter, r *http.Request) {
		day, err := services.OperationalDays.GetOperationalDay(r.Context(), r.PathValue("storeId"), r.PathValue("operationalDayId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, syncedOperationalDayResponse(day))
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:          http.MethodGet,
		Path:            "/v1/stores/{storeId}/reporting/summary",
		OperationID:     "getStoreReportingSummary",
		Summary:         "Get store reporting summary for a time window",
		Description:     sessionProtectedDescription("Query parameters `since` and `until` must be RFC3339 timestamps (inclusive window)."),
		Tags:            []string{"reporting"},
		QueryParameters: reportingWindowQueryParams(),
		Responses: mergeResponseSpecs(
			protectedResponseSpecs("200", "Store reporting summary", storeReportingSummaryResponseSchema()),
			map[string]httpapi.ResponseSpec{
				"400": {Description: "Invalid reporting query", Schema: httpapi.ProblemSchema()},
				"404": {Description: "Store not found", Schema: httpapi.ProblemSchema()},
			},
		),
	}, RequireSession(services.Auth, app.PermissionReportingRead, func(w http.ResponseWriter, r *http.Request) {
		window, err := app.ParseReportingWindow(r.URL.Query().Get("since"), r.URL.Query().Get("until"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		summary, err := services.Reporting.GetStoreSummary(r.Context(), r.PathValue("storeId"), window)
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, storeReportingSummaryResponse(summary))
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/central/reporting/summary",
		OperationID: "getCentralReportingSummary",
		Summary:     "Get cross-store reporting summary for a time window",
		Description: sessionProtectedDescription("Query parameters `since` and `until` must be RFC3339 timestamps (inclusive window). Optional `region` filters registered stores."),
		Tags:        []string{"reporting"},
		QueryParameters: append(reportingWindowQueryParams(), httpapi.QueryParamSpec{
			Name:        "region",
			Description: "Optional store region filter",
			Schema:      httpapi.StringSchema(),
		}),
		Responses: mergeResponseSpecs(
			protectedResponseSpecs("200", "Central reporting summary", centralReportingSummaryResponseSchema()),
			map[string]httpapi.ResponseSpec{
				"400": {Description: "Invalid reporting query", Schema: httpapi.ProblemSchema()},
			},
		),
	}, RequireSession(services.Auth, app.PermissionReportingCentralRead, func(w http.ResponseWriter, r *http.Request) {
		window, err := app.ParseReportingWindow(r.URL.Query().Get("since"), r.URL.Query().Get("until"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		summary, err := services.Reporting.GetCentralSummary(r.Context(), window, r.URL.Query().Get("region"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, centralReportingSummaryResponse(summary))
	}))

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/central/reporting/stores",
		OperationID: "listCentralStoreReportingSummaries",
		Summary:     "List per-store reporting summaries for a time window",
		Description: sessionProtectedDescription("Query parameters `since` and `until` must be RFC3339 timestamps (inclusive window). Optional `region` filters registered stores."),
		Tags:        []string{"reporting"},
		QueryParameters: append(append(reportingWindowQueryParams(), httpapi.QueryParamSpec{
			Name:        "region",
			Description: "Optional store region filter",
			Schema:      httpapi.StringSchema(),
		}), paginationQueryParams()...),
		Responses: mergeResponseSpecs(
			protectedResponseSpecs("200", "Per-store reporting summaries", paginatedStoreReportingSummariesResponseSchema()),
			map[string]httpapi.ResponseSpec{
				"400": {Description: "Invalid reporting query", Schema: httpapi.ProblemSchema()},
			},
		),
	}, RequireSession(services.Auth, app.PermissionReportingCentralRead, func(w http.ResponseWriter, r *http.Request) {
		window, err := app.ParseReportingWindow(r.URL.Query().Get("since"), r.URL.Query().Get("until"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		params := app.ParsePageParams(r.URL.Query().Get("limit"), r.URL.Query().Get("offset"))
		result, err := services.Reporting.ListStoreSummaries(r.Context(), window, r.URL.Query().Get("region"), params)
		if err != nil {
			writeAppError(w, err)
			return
		}
		items := make([]StoreReportingSummaryResponse, 0, len(result.Items))
		for _, summary := range result.Items {
			items = append(items, storeReportingSummaryResponse(summary))
		}
		httpapi.WriteJSON(w, http.StatusOK, PaginatedStoreReportingSummariesResponse{
			Items:      items,
			TotalCount: result.TotalCount,
		})
	}))
}

func writeAppError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, app.ErrStoreNotFound):
		httpapi.WriteProblem(w, http.StatusNotFound, "store_not_found", "Store was not found", err.Error())
	case errors.Is(err, app.ErrCatalogProductNotFound):
		httpapi.WriteProblem(w, http.StatusNotFound, "catalog_product_not_found", "Catalog product was not found", err.Error())
	case errors.Is(err, app.ErrPaymentNotFound):
		httpapi.WriteProblem(w, http.StatusNotFound, "payment_not_found", "Payment was not found", err.Error())
	case errors.Is(err, app.ErrCashMovementNotFound):
		httpapi.WriteProblem(w, http.StatusNotFound, "cash_movement_not_found", "Cash movement was not found", err.Error())
	case errors.Is(err, app.ErrFiscalDocumentNotFound):
		httpapi.WriteProblem(w, http.StatusNotFound, "fiscal_document_not_found", "Fiscal document was not found", err.Error())
	case errors.Is(err, app.ErrReturnNotFound):
		httpapi.WriteProblem(w, http.StatusNotFound, "return_not_found", "Return was not found", err.Error())
	case errors.Is(err, app.ErrOperationalDayNotFound):
		httpapi.WriteProblem(w, http.StatusNotFound, "operational_day_not_found", "Operational day was not found", err.Error())
	case errors.Is(err, app.ErrIdempotencyKeyRequired):
		httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
	case errors.Is(err, app.ErrIdempotencyKeyReused):
		httpapi.WriteProblem(w, http.StatusConflict, "idempotency_key_reused", "Idempotency key was reused", err.Error())
	case errors.Is(err, app.ErrInvalidStoreCommand), errors.Is(err, app.ErrInvalidStoreRegistryCmd), errors.Is(err, domain.ErrInvalidStoreInput):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_store_command", "Invalid store command", err.Error())
	case errors.Is(err, app.ErrInvalidSyncCommand), errors.Is(err, domain.ErrInvalidSyncEventInput):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_sync_command", "Invalid sync batch", err.Error())
	case errors.Is(err, app.ErrInvalidCatalogQuery), errors.Is(err, domain.ErrInvalidCatalogProductInput):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_catalog_query", "Invalid catalog query", err.Error())
	case errors.Is(err, app.ErrInvalidPaymentQuery), errors.Is(err, domain.ErrInvalidSyncedPaymentInput):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_payment_query", "Invalid payment query", err.Error())
	case errors.Is(err, app.ErrInvalidCashMovementQuery), errors.Is(err, domain.ErrInvalidSyncedCashMovementInput):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_cash_movement_query", "Invalid cash movement query", err.Error())
	case errors.Is(err, app.ErrInvalidFiscalDocumentQuery), errors.Is(err, domain.ErrInvalidSyncedFiscalDocumentInput):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_fiscal_document_query", "Invalid fiscal document query", err.Error())
	case errors.Is(err, app.ErrInvalidReturnQuery), errors.Is(err, domain.ErrInvalidSyncedReturnInput):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_return_query", "Invalid return query", err.Error())
	case errors.Is(err, app.ErrInvalidOperationalDayQuery), errors.Is(err, domain.ErrInvalidSyncedOperationalDayInput):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_operational_day_query", "Invalid operational day query", err.Error())
	case errors.Is(err, app.ErrInvalidReportingQuery):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_reporting_query", "Invalid reporting query", err.Error())
	case errors.Is(err, app.ErrInvalidAuthCommand):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_auth_command", "Invalid session command", err.Error())
	case errors.Is(err, app.ErrInvalidCredentials):
		httpapi.WriteProblem(w, http.StatusUnauthorized, "invalid_credentials", "Invalid credentials", err.Error())
	case errors.Is(err, app.ErrSessionNotFound), errors.Is(err, app.ErrSessionExpired):
		httpapi.WriteProblem(w, http.StatusUnauthorized, "session_invalid", "Session is missing or invalid", err.Error())
	case errors.Is(err, app.ErrPermissionDenied):
		httpapi.WriteProblem(w, http.StatusForbidden, "permission_denied", "Permission denied", err.Error())
	case errors.Is(err, app.ErrCentralUserNotFound):
		httpapi.WriteProblem(w, http.StatusNotFound, "central_user_not_found", "Central user was not found", err.Error())
	case errors.Is(err, app.ErrInvalidCentralUserCommand):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_central_user_command", "Invalid central user command", err.Error())
	case errors.Is(err, app.ErrCentralUserConflict):
		httpapi.WriteProblem(w, http.StatusConflict, "central_user_conflict", "Central user already exists", err.Error())
	case errors.Is(err, domain.ErrInvalidCentralUserInput):
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_central_user_command", "Invalid central user command", err.Error())
	case errors.Is(err, app.ErrSyncAPIKeyRequired):
		httpapi.WriteProblem(w, http.StatusUnauthorized, "sync_api_key_required", "Sync API key is required", err.Error())
	case errors.Is(err, app.ErrSyncAPIKeyInvalid):
		httpapi.WriteProblem(w, http.StatusUnauthorized, "sync_api_key_invalid", "Sync API key is invalid", err.Error())
	default:
		httpapi.WriteProblem(w, http.StatusInternalServerError, "internal_error", "Unexpected server error", err.Error())
	}
}

func regionName() string {
	if region := os.Getenv("MERCADIA_CENTRAL_BACKEND_REGION"); region != "" {
		return region
	}
	return "default"
}

func storeResponse(store domain.Store) StoreResponse {
	return StoreResponse{
		ID:           store.ID,
		Name:         store.Name,
		Region:       store.Region,
		RegisteredAt: store.RegisteredAt,
		UpdatedAt:    store.UpdatedAt,
	}
}

func storeResponses(stores []domain.Store) []StoreResponse {
	responses := make([]StoreResponse, 0, len(stores))
	for _, store := range stores {
		responses = append(responses, storeResponse(store))
	}
	return responses
}

func catalogProductResponse(product domain.CatalogProduct) CatalogProductResponse {
	return CatalogProductResponse{
		ID:             product.ID,
		StoreID:        product.StoreID,
		Name:           product.Name,
		Barcodes:       append([]string(nil), product.Barcodes...),
		UnitPriceMinor: product.UnitPriceMinor,
		TaxCategoryID:  product.TaxCategoryID,
		Active:         product.Active,
		Version:        product.Version,
		UpdatedAt:      product.UpdatedAt,
	}
}

func catalogProductResponses(products []domain.CatalogProduct) []CatalogProductResponse {
	responses := make([]CatalogProductResponse, 0, len(products))
	for _, product := range products {
		responses = append(responses, catalogProductResponse(product))
	}
	return responses
}

func syncEventResponse(event domain.SyncEvent) SyncEventResponse {
	payload := event.Payload
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}
	return SyncEventResponse{
		EventID:       event.ID,
		EventType:     event.EventType,
		SourceEventID: event.SourceEventID,
		OccurredAt:    event.OccurredAt,
		ReceivedAt:    event.ReceivedAt,
		Payload:       append(json.RawMessage(nil), payload...),
	}
}

func syncedPaymentResponse(payment domain.SyncedPayment) SyncedPaymentResponse {
	return SyncedPaymentResponse{
		ID:                   payment.ID,
		StoreID:              payment.StoreID,
		ReceiptID:            payment.ReceiptID,
		Method:               payment.Method,
		AmountMinor:          payment.AmountMinor,
		Status:               string(payment.Status),
		CapturedAt:           payment.CapturedAt,
		CancelledAt:          payment.CancelledAt,
		RefundedAmountMinor:  payment.RefundedAmountMinor,
		RemainingAmountMinor: payment.RemainingAmountMinor,
		SourceEventID:        payment.SourceEventID,
		LastEventID:          payment.LastEventID,
		SyncedAt:             payment.SyncedAt,
		UpdatedAt:            payment.UpdatedAt,
	}
}

func syncedPaymentResponses(payments []domain.SyncedPayment) []SyncedPaymentResponse {
	responses := make([]SyncedPaymentResponse, 0, len(payments))
	for _, payment := range payments {
		responses = append(responses, syncedPaymentResponse(payment))
	}
	return responses
}

func syncedCashMovementResponse(movement domain.SyncedCashMovement) SyncedCashMovementResponse {
	return SyncedCashMovementResponse{
		ID:                movement.ID,
		StoreID:           movement.StoreID,
		Type:              movement.Type,
		FromContainerID:   movement.FromContainerID,
		FromContainerType: movement.FromContainerType,
		ToContainerID:     movement.ToContainerID,
		ToContainerType:   movement.ToContainerType,
		AmountMinor:       movement.AmountMinor,
		Currency:          movement.Currency,
		ActorID:           movement.ActorID,
		PostedAt:          movement.PostedAt,
		SourceEventID:     movement.SourceEventID,
		SyncedAt:          movement.SyncedAt,
	}
}

func syncedCashMovementResponses(movements []domain.SyncedCashMovement) []SyncedCashMovementResponse {
	responses := make([]SyncedCashMovementResponse, 0, len(movements))
	for _, movement := range movements {
		responses = append(responses, syncedCashMovementResponse(movement))
	}
	return responses
}

func mergeResponseSpecs(base map[string]httpapi.ResponseSpec, extra map[string]httpapi.ResponseSpec) map[string]httpapi.ResponseSpec {
	merged := make(map[string]httpapi.ResponseSpec, len(base)+len(extra))
	for key, value := range base {
		merged[key] = value
	}
	for key, value := range extra {
		merged[key] = value
	}
	return merged
}

func paginationQueryParams() []httpapi.QueryParamSpec {
	return []httpapi.QueryParamSpec{
		{Name: "limit", Description: "Maximum number of items to return", Schema: httpapi.Schema{"type": "integer", "minimum": 1, "maximum": app.MaxPageLimit}},
		{Name: "offset", Description: "Number of items to skip", Schema: httpapi.Schema{"type": "integer", "minimum": 0}},
	}
}

func reportingWindowQueryParams() []httpapi.QueryParamSpec {
	return []httpapi.QueryParamSpec{
		{Name: "since", Description: "Inclusive window start (RFC3339)", Required: true, Schema: httpapi.DateTimeSchema()},
		{Name: "until", Description: "Inclusive window end (RFC3339)", Required: true, Schema: httpapi.DateTimeSchema()},
	}
}

func storeReportingSummaryResponse(summary app.StoreReportingSummary) StoreReportingSummaryResponse {
	return StoreReportingSummaryResponse{
		StoreID:                     summary.StoreID,
		Since:                       summary.Since,
		Until:                       summary.Until,
		FiscalReceiptCount:          summary.FiscalReceiptCount,
		FiscalReceiptAmountMinor:    summary.FiscalReceiptAmountMinor,
		FiscalReturnCount:           summary.FiscalReturnCount,
		FiscalReturnAmountMinor:     summary.FiscalReturnAmountMinor,
		PaymentsCapturedAmountMinor: summary.PaymentsCapturedAmountMinor,
		PaymentsCancelledCount:      summary.PaymentsCancelledCount,
		PaymentsRefundedAmountMinor: summary.PaymentsRefundedAmountMinor,
		ReturnsSettledCount:         summary.ReturnsSettledCount,
		ReturnsSettledAmountMinor:   summary.ReturnsSettledAmountMinor,
		CashMovementsPostedCount:    summary.CashMovementsPostedCount,
		OperationalDaysClosedCount:  summary.OperationalDaysClosedCount,
	}
}

func centralReportingSummaryResponse(summary app.CentralReportingSummary) CentralReportingSummaryResponse {
	return CentralReportingSummaryResponse{
		Region:                      summary.Region,
		Since:                       summary.Since,
		Until:                       summary.Until,
		StoreCount:                  summary.StoreCount,
		FiscalReceiptCount:          summary.FiscalReceiptCount,
		FiscalReceiptAmountMinor:    summary.FiscalReceiptAmountMinor,
		FiscalReturnCount:           summary.FiscalReturnCount,
		FiscalReturnAmountMinor:     summary.FiscalReturnAmountMinor,
		PaymentsCapturedAmountMinor: summary.PaymentsCapturedAmountMinor,
		PaymentsCancelledCount:      summary.PaymentsCancelledCount,
		PaymentsRefundedAmountMinor: summary.PaymentsRefundedAmountMinor,
		ReturnsSettledCount:         summary.ReturnsSettledCount,
		ReturnsSettledAmountMinor:   summary.ReturnsSettledAmountMinor,
		CashMovementsPostedCount:    summary.CashMovementsPostedCount,
		OperationalDaysClosedCount:  summary.OperationalDaysClosedCount,
	}
}

func statusResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"region":      httpapi.StringSchema(),
		"status":      httpapi.StringSchema(),
		"storeCount":  {"type": "integer"},
		"generatedAt": httpapi.DateTimeSchema(),
	}, "region", "status", "storeCount", "generatedAt")
}

func storeResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":           httpapi.StringSchema(),
		"name":         httpapi.StringSchema(),
		"region":       httpapi.StringSchema(),
		"registeredAt": httpapi.DateTimeSchema(),
		"updatedAt":    httpapi.DateTimeSchema(),
	}, "id", "name", "region", "registeredAt", "updatedAt")
}

func storesResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"stores": httpapi.ArraySchema(storeResponseSchema()),
	}, "stores")
}

func registerStoreRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"storeId": httpapi.StringSchema(),
		"name":    httpapi.StringSchema(),
		"region":  httpapi.StringSchema(),
	}, "storeId", "name")
}

func storeAcceptedResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"store": storeResponseSchema(),
	}, "store")
}

func syncEventRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"eventId":    httpapi.StringSchema(),
		"eventType":  httpapi.StringSchema(),
		"occurredAt": httpapi.DateTimeSchema(),
		"payload":    {"type": "object"},
	}, "eventId", "eventType")
}

func acceptSyncEventsRequestSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"events": httpapi.ArraySchema(syncEventRequestSchema()),
	}, "events")
}

func syncEventsAcceptedResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"storeId":  httpapi.StringSchema(),
		"status":   httpapi.StringSchema(),
		"accepted": {"type": "integer"},
	}, "storeId", "status", "accepted")
}

func syncEventResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"eventId":       httpapi.StringSchema(),
		"eventType":     httpapi.StringSchema(),
		"sourceEventId": httpapi.StringSchema(),
		"occurredAt":    httpapi.DateTimeSchema(),
		"receivedAt":    httpapi.DateTimeSchema(),
		"payload":       {"type": "object"},
	}, "eventId", "eventType", "sourceEventId", "occurredAt", "receivedAt", "payload")
}

func paginatedSyncEventsResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"items":      httpapi.ArraySchema(syncEventResponseSchema()),
		"totalCount": {"type": "integer"},
	}, "items", "totalCount")
}

func catalogProductResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":             httpapi.StringSchema(),
		"storeId":        httpapi.StringSchema(),
		"name":           httpapi.StringSchema(),
		"barcodes":       httpapi.ArraySchema(httpapi.StringSchema()),
		"unitPriceMinor": {"type": "integer"},
		"taxCategoryId":  httpapi.StringSchema(),
		"active":         {"type": "boolean"},
		"version":        {"type": "integer"},
		"updatedAt":      httpapi.DateTimeSchema(),
	}, "id", "storeId", "name", "barcodes", "unitPriceMinor", "taxCategoryId", "active", "version", "updatedAt")
}

func catalogProductsResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"products": httpapi.ArraySchema(catalogProductResponseSchema()),
	}, "products")
}

func catalogDeltaResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"since":    httpapi.DateTimeSchema(),
		"products": httpapi.ArraySchema(catalogProductResponseSchema()),
	}, "since", "products")
}

func syncedPaymentResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":                   httpapi.StringSchema(),
		"storeId":              httpapi.StringSchema(),
		"receiptId":            httpapi.StringSchema(),
		"method":               httpapi.StringSchema(),
		"amountMinor":          {"type": "integer"},
		"status":               httpapi.StringSchema(),
		"capturedAt":           httpapi.DateTimeSchema(),
		"cancelledAt":          httpapi.DateTimeSchema(),
		"refundedAmountMinor":  {"type": "integer"},
		"remainingAmountMinor": {"type": "integer"},
		"sourceEventId":        httpapi.StringSchema(),
		"lastEventId":          httpapi.StringSchema(),
		"syncedAt":             httpapi.DateTimeSchema(),
		"updatedAt":            httpapi.DateTimeSchema(),
	}, "id", "storeId", "receiptId", "method", "amountMinor", "status", "capturedAt", "refundedAmountMinor", "remainingAmountMinor", "sourceEventId", "lastEventId", "syncedAt", "updatedAt")
}

func paginatedSyncedPaymentsResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"items":      httpapi.ArraySchema(syncedPaymentResponseSchema()),
		"totalCount": {"type": "integer"},
	}, "items", "totalCount")
}

func syncedCashMovementResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":                httpapi.StringSchema(),
		"storeId":           httpapi.StringSchema(),
		"type":              httpapi.StringSchema(),
		"fromContainerId":   httpapi.StringSchema(),
		"fromContainerType": httpapi.StringSchema(),
		"toContainerId":     httpapi.StringSchema(),
		"toContainerType":   httpapi.StringSchema(),
		"amountMinor":       {"type": "integer"},
		"currency":          httpapi.StringSchema(),
		"actorId":           httpapi.StringSchema(),
		"postedAt":          httpapi.DateTimeSchema(),
		"sourceEventId":     httpapi.StringSchema(),
		"syncedAt":          httpapi.DateTimeSchema(),
	}, "id", "storeId", "type", "fromContainerId", "fromContainerType", "toContainerId", "toContainerType", "amountMinor", "actorId", "postedAt", "sourceEventId", "syncedAt")
}

func paginatedSyncedCashMovementsResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"items":      httpapi.ArraySchema(syncedCashMovementResponseSchema()),
		"totalCount": {"type": "integer"},
	}, "items", "totalCount")
}

func syncedFiscalDocumentResponse(document domain.SyncedFiscalDocument) SyncedFiscalDocumentResponse {
	return SyncedFiscalDocumentResponse{
		ID:            document.ID,
		StoreID:       document.StoreID,
		ReceiptID:     document.ReceiptID,
		Kind:          document.Kind,
		AmountMinor:   document.AmountMinor,
		DeviceID:      document.DeviceID,
		FiscalSign:    document.FiscalSign,
		FiscalizedAt:  document.FiscalizedAt,
		ReturnID:      document.ReturnID,
		SourceEventID: document.SourceEventID,
		SyncedAt:      document.SyncedAt,
	}
}

func syncedFiscalDocumentResponses(documents []domain.SyncedFiscalDocument) []SyncedFiscalDocumentResponse {
	responses := make([]SyncedFiscalDocumentResponse, 0, len(documents))
	for _, document := range documents {
		responses = append(responses, syncedFiscalDocumentResponse(document))
	}
	return responses
}

func syncedFiscalDocumentResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":            httpapi.StringSchema(),
		"storeId":       httpapi.StringSchema(),
		"receiptId":     httpapi.StringSchema(),
		"kind":          httpapi.StringSchema(),
		"amountMinor":   {"type": "integer"},
		"deviceId":      httpapi.StringSchema(),
		"fiscalSign":    httpapi.StringSchema(),
		"fiscalizedAt":  httpapi.DateTimeSchema(),
		"returnId":      httpapi.StringSchema(),
		"sourceEventId": httpapi.StringSchema(),
		"syncedAt":      httpapi.DateTimeSchema(),
	}, "id", "storeId", "receiptId", "kind", "amountMinor", "deviceId", "fiscalSign", "fiscalizedAt", "sourceEventId", "syncedAt")
}

func paginatedSyncedFiscalDocumentsResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"items":      httpapi.ArraySchema(syncedFiscalDocumentResponseSchema()),
		"totalCount": {"type": "integer"},
	}, "items", "totalCount")
}

func syncedReturnResponse(ret domain.SyncedReturn) SyncedReturnResponse {
	return SyncedReturnResponse{
		ID:             ret.ID,
		StoreID:        ret.StoreID,
		ReceiptID:      ret.ReceiptID,
		TotalMinor:     ret.TotalMinor,
		PaymentIDs:     ret.PaymentIDs,
		CashMovementID: ret.CashMovementID,
		ActorID:        ret.ActorID,
		SettledAt:      ret.SettledAt,
		SourceEventID:  ret.SourceEventID,
		SyncedAt:       ret.SyncedAt,
	}
}

func syncedReturnResponses(returns []domain.SyncedReturn) []SyncedReturnResponse {
	responses := make([]SyncedReturnResponse, 0, len(returns))
	for _, ret := range returns {
		responses = append(responses, syncedReturnResponse(ret))
	}
	return responses
}

func syncedReturnResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":             httpapi.StringSchema(),
		"storeId":        httpapi.StringSchema(),
		"receiptId":      httpapi.StringSchema(),
		"totalMinor":     {"type": "integer"},
		"paymentIds":     httpapi.ArraySchema(httpapi.StringSchema()),
		"cashMovementId": httpapi.StringSchema(),
		"actorId":        httpapi.StringSchema(),
		"settledAt":      httpapi.DateTimeSchema(),
		"sourceEventId":  httpapi.StringSchema(),
		"syncedAt":       httpapi.DateTimeSchema(),
	}, "id", "storeId", "receiptId", "totalMinor", "paymentIds", "actorId", "settledAt", "sourceEventId", "syncedAt")
}

func paginatedSyncedReturnsResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"items":      httpapi.ArraySchema(syncedReturnResponseSchema()),
		"totalCount": {"type": "integer"},
	}, "items", "totalCount")
}

func syncedOperationalDayResponse(day domain.SyncedOperationalDay) SyncedOperationalDayResponse {
	return SyncedOperationalDayResponse{
		ID:            day.ID,
		StoreID:       day.StoreID,
		BusinessDate:  day.BusinessDate,
		ClosedByID:    day.ClosedByID,
		ClosedAt:      day.ClosedAt,
		SourceEventID: day.SourceEventID,
		SyncedAt:      day.SyncedAt,
	}
}

func syncedOperationalDayResponses(days []domain.SyncedOperationalDay) []SyncedOperationalDayResponse {
	responses := make([]SyncedOperationalDayResponse, 0, len(days))
	for _, day := range days {
		responses = append(responses, syncedOperationalDayResponse(day))
	}
	return responses
}

func syncedOperationalDayResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"id":            httpapi.StringSchema(),
		"storeId":       httpapi.StringSchema(),
		"businessDate":  httpapi.StringSchema(),
		"closedById":    httpapi.StringSchema(),
		"closedAt":      httpapi.DateTimeSchema(),
		"sourceEventId": httpapi.StringSchema(),
		"syncedAt":      httpapi.DateTimeSchema(),
	}, "id", "storeId", "businessDate", "closedById", "closedAt", "sourceEventId", "syncedAt")
}

func paginatedSyncedOperationalDaysResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"items":      httpapi.ArraySchema(syncedOperationalDayResponseSchema()),
		"totalCount": {"type": "integer"},
	}, "items", "totalCount")
}

func storeReportingSummaryResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"storeId":                     httpapi.StringSchema(),
		"since":                       httpapi.DateTimeSchema(),
		"until":                       httpapi.DateTimeSchema(),
		"fiscalReceiptCount":          {"type": "integer"},
		"fiscalReceiptAmountMinor":    {"type": "integer"},
		"fiscalReturnCount":           {"type": "integer"},
		"fiscalReturnAmountMinor":     {"type": "integer"},
		"paymentsCapturedAmountMinor": {"type": "integer"},
		"paymentsCancelledCount":      {"type": "integer"},
		"paymentsRefundedAmountMinor": {"type": "integer"},
		"returnsSettledCount":         {"type": "integer"},
		"returnsSettledAmountMinor":   {"type": "integer"},
		"cashMovementsPostedCount":    {"type": "integer"},
		"operationalDaysClosedCount":  {"type": "integer"},
	}, "storeId", "since", "until", "fiscalReceiptCount", "fiscalReceiptAmountMinor", "fiscalReturnCount", "fiscalReturnAmountMinor", "paymentsCapturedAmountMinor", "paymentsCancelledCount", "paymentsRefundedAmountMinor", "returnsSettledCount", "returnsSettledAmountMinor", "cashMovementsPostedCount", "operationalDaysClosedCount")
}

func centralReportingSummaryResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"region":                      httpapi.StringSchema(),
		"since":                       httpapi.DateTimeSchema(),
		"until":                       httpapi.DateTimeSchema(),
		"storeCount":                  {"type": "integer"},
		"fiscalReceiptCount":          {"type": "integer"},
		"fiscalReceiptAmountMinor":    {"type": "integer"},
		"fiscalReturnCount":           {"type": "integer"},
		"fiscalReturnAmountMinor":     {"type": "integer"},
		"paymentsCapturedAmountMinor": {"type": "integer"},
		"paymentsCancelledCount":      {"type": "integer"},
		"paymentsRefundedAmountMinor": {"type": "integer"},
		"returnsSettledCount":         {"type": "integer"},
		"returnsSettledAmountMinor":   {"type": "integer"},
		"cashMovementsPostedCount":    {"type": "integer"},
		"operationalDaysClosedCount":  {"type": "integer"},
	}, "since", "until", "storeCount", "fiscalReceiptCount", "fiscalReceiptAmountMinor", "fiscalReturnCount", "fiscalReturnAmountMinor", "paymentsCapturedAmountMinor", "paymentsCancelledCount", "paymentsRefundedAmountMinor", "returnsSettledCount", "returnsSettledAmountMinor", "cashMovementsPostedCount", "operationalDaysClosedCount")
}

func paginatedStoreReportingSummariesResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"items":      httpapi.ArraySchema(storeReportingSummaryResponseSchema()),
		"totalCount": {"type": "integer"},
	}, "items", "totalCount")
}

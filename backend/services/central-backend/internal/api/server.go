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
	StoreRegistry *app.StoreRegistryService
	Sync          *app.SyncService
	Catalog       *app.CatalogService
	Payments      *app.PaymentsService
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
	ID            string    `json:"id"`
	StoreID       string    `json:"storeId"`
	ReceiptID     string    `json:"receiptId"`
	Method        string    `json:"method"`
	AmountMinor   int64     `json:"amountMinor"`
	CapturedAt    time.Time `json:"capturedAt"`
	SourceEventID string    `json:"sourceEventId"`
	SyncedAt      time.Time `json:"syncedAt"`
}

type PaginatedSyncedPaymentsResponse struct {
	Items      []SyncedPaymentResponse `json:"items"`
	TotalCount int                     `json:"totalCount"`
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
	return &ServerBundle{
		Handler:  newHandler(services, opts),
		Services: services,
		Handle:   handle,
	}, nil
}

func newServices(repo infra.Repository) Services {
	return Services{
		StoreRegistry: app.NewStoreRegistryService(repo, repo),
		Sync:          app.NewSyncService(repo, repo, repo, repo, repo),
		Catalog:       app.NewCatalogService(repo, repo),
		Payments:      app.NewPaymentsService(repo, repo),
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
		Tags:        []string{"system"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Central backend status", Schema: statusResponseSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
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
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/stores",
		OperationID: "listStores",
		Summary:     "List registered stores",
		Tags:        []string{"stores"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Registered stores", Schema: storesResponseSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		stores, err := services.StoreRegistry.ListStores(r.Context())
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, StoresResponse{Stores: storeResponses(stores)})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/stores",
		OperationID:         "registerStore",
		Summary:             "Register a store",
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
			"409": {Description: "Idempotency conflict", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
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
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/stores/{storeId}/sync-events",
		OperationID:         "acceptStoreSyncEvents",
		Summary:             "Accept synchronized Store Edge events",
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
			"404": {Description: "Store not found", Schema: httpapi.ProblemSchema()},
			"409": {Description: "Idempotency conflict", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
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
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:          http.MethodGet,
		Path:            "/v1/stores/{storeId}/sync-events",
		OperationID:     "listStoreSyncEvents",
		Summary:         "List synchronized Store Edge events",
		Tags:            []string{"sync"},
		QueryParameters: paginationQueryParams(),
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Synchronized events", Schema: paginatedSyncEventsResponseSchema()},
			"400": {Description: "Invalid list query", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Store not found", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
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
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/stores/{storeId}/catalog/products",
		OperationID: "listStoreCatalogProducts",
		Summary:     "List catalog products for a store",
		Tags:        []string{"catalog"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Catalog products", Schema: catalogProductsResponseSchema()},
			"404": {Description: "Store not found", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		result, err := services.Catalog.ListProducts(r.Context(), r.PathValue("storeId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, CatalogProductsResponse{Products: catalogProductResponses(result.Products)})
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/stores/{storeId}/catalog/delta",
		OperationID: "getStoreCatalogDelta",
		Summary:     "Get catalog changes since a timestamp",
		Description: "Query parameter `since` must be an RFC3339 timestamp.",
		Tags:        []string{"catalog"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Catalog delta", Schema: catalogDeltaResponseSchema()},
			"400": {Description: "Invalid catalog query", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Store not found", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
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
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:          http.MethodGet,
		Path:            "/v1/stores/{storeId}/payments",
		OperationID:     "listStorePayments",
		Summary:         "List synchronized payments for a store",
		Tags:            []string{"sync"},
		QueryParameters: paginationQueryParams(),
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Synchronized payments", Schema: paginatedSyncedPaymentsResponseSchema()},
			"400": {Description: "Invalid list query", Schema: httpapi.ProblemSchema()},
			"404": {Description: "Store not found", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
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
	})

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:      http.MethodGet,
		Path:        "/v1/stores/{storeId}/payments/{paymentId}",
		OperationID: "getStorePayment",
		Summary:     "Get a synchronized payment",
		Tags:        []string{"sync"},
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Synchronized payment", Schema: syncedPaymentResponseSchema()},
			"404": {Description: "Store or payment not found", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		payment, err := services.Payments.GetPayment(r.Context(), r.PathValue("storeId"), r.PathValue("paymentId"))
		if err != nil {
			writeAppError(w, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, syncedPaymentResponse(payment))
	})
}

func writeAppError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, app.ErrStoreNotFound):
		httpapi.WriteProblem(w, http.StatusNotFound, "store_not_found", "Store was not found", err.Error())
	case errors.Is(err, app.ErrCatalogProductNotFound):
		httpapi.WriteProblem(w, http.StatusNotFound, "catalog_product_not_found", "Catalog product was not found", err.Error())
	case errors.Is(err, app.ErrPaymentNotFound):
		httpapi.WriteProblem(w, http.StatusNotFound, "payment_not_found", "Payment was not found", err.Error())
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
		ID:            payment.ID,
		StoreID:       payment.StoreID,
		ReceiptID:     payment.ReceiptID,
		Method:        payment.Method,
		AmountMinor:   payment.AmountMinor,
		CapturedAt:    payment.CapturedAt,
		SourceEventID: payment.SourceEventID,
		SyncedAt:      payment.SyncedAt,
	}
}

func syncedPaymentResponses(payments []domain.SyncedPayment) []SyncedPaymentResponse {
	responses := make([]SyncedPaymentResponse, 0, len(payments))
	for _, payment := range payments {
		responses = append(responses, syncedPaymentResponse(payment))
	}
	return responses
}

func paginationQueryParams() []httpapi.QueryParamSpec {
	return []httpapi.QueryParamSpec{
		{Name: "limit", Description: "Maximum number of items to return", Schema: httpapi.Schema{"type": "integer", "minimum": 1, "maximum": app.MaxPageLimit}},
		{Name: "offset", Description: "Number of items to skip", Schema: httpapi.Schema{"type": "integer", "minimum": 0}},
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
		"id":            httpapi.StringSchema(),
		"storeId":       httpapi.StringSchema(),
		"receiptId":     httpapi.StringSchema(),
		"method":        httpapi.StringSchema(),
		"amountMinor":   {"type": "integer"},
		"capturedAt":    httpapi.DateTimeSchema(),
		"sourceEventId": httpapi.StringSchema(),
		"syncedAt":      httpapi.DateTimeSchema(),
	}, "id", "storeId", "receiptId", "method", "amountMinor", "capturedAt", "sourceEventId", "syncedAt")
}

func paginatedSyncedPaymentsResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"items":      httpapi.ArraySchema(syncedPaymentResponseSchema()),
		"totalCount": {"type": "integer"},
	}, "items", "totalCount")
}

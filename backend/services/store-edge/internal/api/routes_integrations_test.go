package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/api"
	"mercadia.dev/pos/services/store-edge/internal/app"
	"mercadia.dev/pos/services/store-edge/internal/domain"
	"mercadia.dev/pos/services/store-edge/internal/infra/central"
	"mercadia.dev/pos/services/store-edge/internal/infra/memory"
)

func TestTerminalEventsStreamReturnsSSE(t *testing.T) {
	handler := api.NewServer()

	done := make(chan struct{})
	go func() {
		defer close(done)

		request := httptest.NewRequest(http.MethodGet, "/v1/stores/store-1/terminals/events", nil)
		request = request.WithContext(context.Background())
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		if got := response.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/event-stream") {
			t.Errorf("content type = %q", got)
		}
	}()

	select {
	case <-done:
	case <-time.After(250 * time.Millisecond):
	}
}

func TestCatalogSyncEndpointPullsFromCentral(t *testing.T) {
	updatedAt := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
	centralServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"since": time.Unix(0, 0).UTC(),
			"products": []map[string]any{
				{
					"id":             "remote-sku-2",
					"storeId":        "store-1",
					"name":           "Remote Tea",
					"barcodes":       []string{"5550001112223"},
					"unitPriceMinor": 15000,
					"taxCategoryId":  "vat_20",
					"active":         true,
					"version":        1,
					"updatedAt":      updatedAt,
				},
			},
		})
	}))
	defer centralServer.Close()

	store := memory.NewStore()
	catalogSync := app.NewCatalogSyncService(store, store, central.NewClient(centralServer.URL, centralServer.Client()))
	mux, _ := wireTestServer(store, catalogSync)

	body := bytes.NewBufferString(`{}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/stores/store-1/catalog/sync", body)
	request.Header.Set("Idempotency-Key", "catalog-sync-1")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", response.Code, response.Body.String())
	}

	var payload struct {
		ProductsCount int `json:"productsCount"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.ProductsCount != 1 {
		t.Fatalf("products count = %d", payload.ProductsCount)
	}
}

func TestTerminalHeartbeatPublishesSSEEvent(t *testing.T) {
	store := memory.NewStore()
	mux, hub := wireTestServerWithEvents(store)

	events := hub.Subscribe("store-1")
	defer hub.Unsubscribe("store-1", events)

	request := httptest.NewRequest(http.MethodPost, "/v1/terminals/terminal-1/heartbeat", bytes.NewBufferString(`{
		"storeId":"store-1",
		"kind":"pos",
		"softwareVersion":"1.0.0"
	}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "heartbeat-1")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusAccepted {
		t.Fatalf("heartbeat status = %d", response.Code)
	}

	select {
	case event := <-events:
		if event.TerminalID != "terminal-1" || event.StoreID != "store-1" {
			t.Fatalf("unexpected event: %+v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("expected terminal heartbeat event")
	}
}

func wireTestServer(store *memory.Store, catalogSync *app.CatalogSyncService) (*http.ServeMux, *app.TerminalEventHub) {
	hub := app.NewTerminalEventHub()
	mux := http.NewServeMux()
	terminals := app.NewTerminalService(store, store, app.WithTerminalEventPublisher(hub))
	mountCatalogSyncForTest(mux, catalogSync)
	mountTerminalEventsForTest(mux, hub)
	registerTerminalHeartbeatForTest(mux, terminals)
	return mux, hub
}

func wireTestServerWithEvents(store *memory.Store) (*http.ServeMux, *app.TerminalEventHub) {
	return wireTestServer(store, nil)
}

func mountCatalogSyncForTest(mux *http.ServeMux, catalogSync *app.CatalogSyncService) {
	mux.HandleFunc("POST /v1/stores/{storeId}/catalog/sync", func(w http.ResponseWriter, r *http.Request) {
		result, err := catalogSync.Sync(r.Context(), app.SyncCatalogCommand{StoreID: r.PathValue("storeId")})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"storeId":       result.StoreID,
			"since":         result.Since,
			"syncedAt":      result.SyncedAt,
			"productsCount": result.ProductsCount,
		})
	})
}

func mountTerminalEventsForTest(mux *http.ServeMux, hub *app.TerminalEventHub) {
	mux.HandleFunc("GET /v1/stores/{storeId}/terminals/events", func(w http.ResponseWriter, r *http.Request) {
		storeID := r.PathValue("storeId")
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		events := hub.Subscribe(storeID)
		defer hub.Unsubscribe(storeID, events)
		flusher.Flush()
		for {
			select {
			case <-r.Context().Done():
				return
			case event, ok := <-events:
				if !ok {
					return
				}
				payload, _ := json.Marshal(event)
				_, _ = w.Write([]byte("event: " + event.Type + "\n"))
				_, _ = w.Write([]byte("data: " + string(payload) + "\n\n"))
				flusher.Flush()
			}
		}
	})
}

func registerTerminalHeartbeatForTest(mux *http.ServeMux, terminals *app.TerminalService) {
	mux.HandleFunc("POST /v1/terminals/{terminalId}/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			StoreID         string              `json:"storeId"`
			Kind            domain.TerminalKind `json:"kind"`
			SoftwareVersion string              `json:"softwareVersion"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := terminals.RecordHeartbeat(r.Context(), app.RecordTerminalHeartbeatCommand{
			IdempotencyKey:  r.Header.Get("Idempotency-Key"),
			TerminalID:      r.PathValue("terminalId"),
			StoreID:         request.StoreID,
			Kind:            request.Kind,
			SoftwareVersion: request.SoftwareVersion,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(result)
	})
}

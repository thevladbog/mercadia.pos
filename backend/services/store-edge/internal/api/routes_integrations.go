package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"mercadia.dev/pos/platform/httpapi"
	"mercadia.dev/pos/services/store-edge/internal/app"
)

type CatalogSyncResponse struct {
	StoreID       string    `json:"storeId"`
	Since         time.Time `json:"since"`
	SyncedAt      time.Time `json:"syncedAt"`
	ProductsCount int       `json:"productsCount"`
}

func mountCatalogSyncRoute(mux *http.ServeMux, spec *httpapi.Spec, catalogSync *app.CatalogSyncService) {
	if catalogSync == nil {
		return
	}

	httpapi.Register(mux, spec, httpapi.Operation{
		Method:              http.MethodPost,
		Path:                "/v1/stores/{storeId}/catalog/sync",
		OperationID:         "syncStoreCatalog",
		Summary:             "Pull catalog changes from central backend",
		Description:         "Idempotent pull of catalog delta from central backend into the local store cache.",
		Tags:                []string{"catalog"},
		RequiresIdempotency: true,
		Responses: map[string]httpapi.ResponseSpec{
			"200": {Description: "Catalog sync completed", Schema: catalogSyncResponseSchema()},
			"503": {Description: "Catalog sync unavailable", Schema: httpapi.ProblemSchema()},
		},
	}, func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpapi.RequireIdempotencyKey(r); err != nil {
			httpapi.WriteProblem(w, http.StatusBadRequest, "idempotency_key_required", "Idempotency key is required", err.Error())
			return
		}

		result, err := catalogSync.Sync(r.Context(), app.SyncCatalogCommand{
			StoreID: r.PathValue("storeId"),
		})
		if err != nil {
			writeCatalogSyncError(w, err)
			return
		}

		httpapi.WriteJSON(w, http.StatusOK, CatalogSyncResponse{
			StoreID:       result.StoreID,
			Since:         result.Since,
			SyncedAt:      result.SyncedAt,
			ProductsCount: result.ProductsCount,
		})
	})
}

func writeCatalogSyncError(w http.ResponseWriter, err error) {
	switch err {
	case app.ErrInvalidCatalogSyncCommand:
		httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_catalog_sync_command", "Invalid catalog sync command", err.Error())
	case app.ErrCatalogSyncUnavailable:
		httpapi.WriteProblem(w, http.StatusServiceUnavailable, "catalog_sync_unavailable", "Catalog sync is unavailable", err.Error())
	default:
		httpapi.WriteProblem(w, http.StatusInternalServerError, "internal_error", "Internal server error", err.Error())
	}
}

func catalogSyncResponseSchema() httpapi.Schema {
	return httpapi.ObjectSchema(map[string]httpapi.Schema{
		"storeId":       httpapi.StringSchema(),
		"since":         httpapi.DateTimeSchema(),
		"syncedAt":      httpapi.DateTimeSchema(),
		"productsCount": httpapi.Schema{"type": "integer"},
	}, "storeId", "since", "syncedAt", "productsCount")
}

func mountTerminalEventsRoute(mux *http.ServeMux, hub *app.TerminalEventHub) {
	if hub == nil {
		return
	}

	mux.HandleFunc("GET /v1/stores/{storeId}/terminals/events", func(w http.ResponseWriter, r *http.Request) {
		storeID := r.PathValue("storeId")
		if storeID == "" {
			httpapi.WriteProblem(w, http.StatusBadRequest, "invalid_terminal_events_query", "Invalid terminal events query", "storeId is required")
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			httpapi.WriteProblem(w, http.StatusInternalServerError, "internal_error", "Internal server error", "streaming is not supported")
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		events := hub.Subscribe(storeID)
		defer hub.Unsubscribe(storeID, events)

		fmt.Fprintf(w, ": connected store=%s\n\n", storeID)
		flusher.Flush()

		notify := r.Context().Done()
		for {
			select {
			case <-notify:
				return
			case event, ok := <-events:
				if !ok {
					return
				}
				payload, err := json.Marshal(event)
				if err != nil {
					continue
				}
				fmt.Fprintf(w, "event: %s\n", event.Type)
				fmt.Fprintf(w, "data: %s\n\n", payload)
				flusher.Flush()
			}
		}
	})
}

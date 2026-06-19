package central_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/infra/central"
)

func TestClientCatalogDeltaSendsSyncAPIKeyWhenConfigured(t *testing.T) {
	var gotKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-Sync-Api-Key")
		if gotKey != "test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"since":    time.Unix(0, 0).UTC(),
			"products": []any{},
		})
	}))
	defer server.Close()

	client := central.NewClientWithSyncAPIKey(server.URL, "test-key", server.Client())
	products, _, err := client.CatalogDelta(context.Background(), "store-1", time.Time{})
	if err != nil {
		t.Fatalf("catalog delta: %v", err)
	}
	if gotKey != "test-key" {
		t.Fatalf("sync api key header = %q", gotKey)
	}
	if len(products) != 0 {
		t.Fatalf("products = %+v", products)
	}
}

func TestClientCatalogDeltaOmitsSyncAPIKeyWhenUnset(t *testing.T) {
	var gotKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-Sync-Api-Key")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"since":    time.Unix(0, 0).UTC(),
			"products": []any{},
		})
	}))
	defer server.Close()

	client := central.NewClient(server.URL, server.Client())
	if _, _, err := client.CatalogDelta(context.Background(), "store-1", time.Time{}); err != nil {
		t.Fatalf("catalog delta: %v", err)
	}
	if gotKey != "" {
		t.Fatalf("unexpected sync api key header = %q", gotKey)
	}
}

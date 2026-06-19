package nats_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"mercadia.dev/pos/services/central-backend/internal/app"
	"mercadia.dev/pos/services/central-backend/internal/infra/memory"
	centralnats "mercadia.dev/pos/services/central-backend/internal/infra/nats"
)

func TestConsumerProcessesPublishedMessageWhenNatsAvailable(t *testing.T) {
	natsURL := os.Getenv("MERCADIA_CENTRAL_BACKEND_NATS_URL")
	if natsURL == "" {
		t.Skip("MERCADIA_CENTRAL_BACKEND_NATS_URL is not set")
	}

	store := memory.NewStore()
	registry := app.NewStoreRegistryService(store, store)
	_, err := registry.RegisterStore(context.Background(), app.RegisterStoreCommand{
		StoreID: "store-1",
		Name:    "Main Street",
		Region:  "west",
	})
	if err != nil {
		t.Fatalf("register store: %v", err)
	}

	syncService := app.NewSyncService(store, store, store, store, store)
	consumer, err := centralnats.NewConsumer(natsURL, syncService)
	if err != nil {
		t.Fatalf("new consumer: %v", err)
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = consumer.Run(ctx)
	}()

	publisher, err := centralnats.NewTestPublisher(natsURL)
	if err != nil {
		t.Fatalf("new test publisher: %v", err)
	}
	defer publisher.Close()

	occurredAt := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
	body, err := json.Marshal(map[string]any{
		"eventId":    "integration-evt-1",
		"eventType":  "payment.captured",
		"payload":    map[string]any{"storeId": "store-1", "paymentId": "pay-1"},
		"occurredAt": occurredAt,
	})
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	if err := publisher.Publish(context.Background(), "store-1", body); err != nil {
		t.Fatalf("publish message: %v", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		exists, err := store.ExistsSyncEvent(context.Background(), "store-1", "integration-evt-1")
		if err != nil {
			t.Fatalf("check sync event: %v", err)
		}
		if exists {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("expected sync event to be persisted by consumer")
}

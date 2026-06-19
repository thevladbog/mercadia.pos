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

	syncService := app.NewSyncService(store, store, store, store, store, store, store)
	paymentsService := app.NewPaymentsService(store, store)
	cashMovementsService := app.NewCashMovementsService(store, store)
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

	capturedAt := time.Date(2026, 6, 19, 14, 30, 0, 0, time.UTC)
	postedAt := time.Date(2026, 6, 19, 15, 0, 0, 0, time.UTC)

	messages := []struct {
		eventID   string
		eventType string
		payload   map[string]any
	}{
		{
			eventID:   "integration-pay-1",
			eventType: "payment.captured",
			payload: map[string]any{
				"storeId":     "store-1",
				"paymentId":   "pay-1",
				"receiptId":   "rcpt-1",
				"method":      "card",
				"amountMinor": int64(150000),
				"capturedAt":  capturedAt,
			},
		},
		{
			eventID:   "integration-cash-1",
			eventType: "cash.movement.posted",
			payload: map[string]any{
				"storeId":           "store-1",
				"cashMovementId":    "cash-1",
				"type":              "safe_to_bank",
				"fromContainerId":   "safe-1",
				"fromContainerType": "safe",
				"toContainerId":     "bank-1",
				"toContainerType":   "bank",
				"amountMinor":       int64(200000),
				"currency":          "RUB",
				"actorId":           "senior-1",
				"postedAt":          postedAt,
			},
		},
	}

	for _, message := range messages {
		body, err := json.Marshal(map[string]any{
			"eventId":    message.eventID,
			"eventType":  message.eventType,
			"payload":    message.payload,
			"occurredAt": capturedAt,
		})
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		if err := publisher.Publish(context.Background(), "store-1", body); err != nil {
			t.Fatalf("publish message: %v", err)
		}
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		paymentExists, err := store.ExistsSyncEvent(context.Background(), "store-1", "integration-pay-1")
		if err != nil {
			t.Fatalf("check payment sync event: %v", err)
		}
		cashExists, err := store.ExistsSyncEvent(context.Background(), "store-1", "integration-cash-1")
		if err != nil {
			t.Fatalf("check cash sync event: %v", err)
		}
		if paymentExists && cashExists {
			payment, err := paymentsService.GetPayment(context.Background(), "store-1", "pay-1")
			if err != nil {
				t.Fatalf("get projected payment: %v", err)
			}
			if payment.ReceiptID != "rcpt-1" {
				t.Fatalf("projected payment = %+v", payment)
			}
			movement, err := cashMovementsService.GetCashMovement(context.Background(), "store-1", "cash-1")
			if err != nil {
				t.Fatalf("get projected cash movement: %v", err)
			}
			if movement.Type != "safe_to_bank" {
				t.Fatalf("projected cash movement = %+v", movement)
			}
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("expected sync events and projections to be persisted by consumer")
}

package app_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/app"
	"mercadia.dev/pos/services/store-edge/internal/domain"
	"mercadia.dev/pos/services/store-edge/internal/infra/memory"
)

func TestOutboxEnqueuePaymentCaptured(t *testing.T) {
	store := memory.NewStore()
	var counter int
	now := func() time.Time {
		return time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
	}
	newID := func(prefix string) string {
		counter++
		return fmt.Sprintf("%s-%d", prefix, counter)
	}
	outbox := app.NewOutboxService(store, app.WithOutboxClock(now), app.WithOutboxIDGenerator(newID))

	payment, err := domain.CreateCapturedPayment(domain.CreateCapturedPaymentInput{
		ID:          "pay-1",
		ReceiptID:   "rcpt-1",
		Method:      domain.PaymentMethodCash,
		AmountMinor: 1000,
		Now:         now(),
	})
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}
	if err := outbox.RecordPaymentCaptured(context.Background(), payment, "store-1"); err != nil {
		t.Fatalf("record payment captured: %v", err)
	}

	pending, published, err := store.CountOutboxEvents(context.Background())
	if err != nil {
		t.Fatalf("count outbox: %v", err)
	}
	if pending != 1 || published != 0 {
		t.Fatalf("pending=%d published=%d", pending, published)
	}

	events, err := store.ListPendingOutboxEvents(context.Background(), 10)
	if err != nil {
		t.Fatalf("list pending: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("pending events = %d", len(events))
	}
	if events[0].EventType != domain.OutboxEventPaymentCaptured {
		t.Fatalf("event type = %s", events[0].EventType)
	}
}

func TestMarkOutboxEventPublishedIsIdempotent(t *testing.T) {
	store := memory.NewStore()
	ctx := context.Background()
	now := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
	publishedAt := now.Add(time.Minute)

	event := domain.OutboxEvent{
		ID:            "obx-1",
		AggregateType: domain.OutboxAggregateCashMovement,
		AggregateID:   "cash-1",
		EventType:     domain.OutboxEventCashMovementPosted,
		Payload:       []byte(`{"storeId":"store-1"}`),
		CreatedAt:     now,
	}
	if err := store.SaveOutboxEvent(ctx, event); err != nil {
		t.Fatalf("save outbox event: %v", err)
	}

	updated, err := store.MarkOutboxEventPublished(ctx, event.ID, publishedAt)
	if err != nil {
		t.Fatalf("mark published: %v", err)
	}
	if !updated {
		t.Fatal("expected first mark to update row")
	}

	updated, err = store.MarkOutboxEventPublished(ctx, event.ID, publishedAt.Add(time.Hour))
	if err != nil {
		t.Fatalf("mark published again: %v", err)
	}
	if updated {
		t.Fatal("expected second mark to be idempotent no-op")
	}

	pending, published, err := store.CountOutboxEvents(ctx)
	if err != nil {
		t.Fatalf("count outbox: %v", err)
	}
	if pending != 0 || published != 1 {
		t.Fatalf("pending=%d published=%d", pending, published)
	}
}

func TestPaymentCaptureEnqueuesOutboxEvent(t *testing.T) {
	store := memory.NewStore(memory.WithProducts(testProduct()))
	checkout, payments := newTestCheckoutAndPaymentServicesWithOutbox(store)

	receiptID := openAndScanTestReceipt(t, checkout)
	if _, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-outbox",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCardMock,
		AmountMinor:    39998,
	}); err != nil {
		t.Fatalf("create payment: %v", err)
	}

	pending, _, err := store.CountOutboxEvents(context.Background())
	if err != nil {
		t.Fatalf("count outbox: %v", err)
	}
	if pending != 1 {
		t.Fatalf("pending outbox events = %d", pending)
	}
}

func newTestCheckoutAndPaymentServicesWithOutbox(store *memory.Store) (*app.CheckoutService, *app.PaymentService) {
	var counter int
	now := func() time.Time {
		return time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
	}
	newID := func(prefix string) string {
		counter++
		return fmt.Sprintf("%s-test-%d", prefix, counter)
	}
	outbox := app.NewOutboxService(store, app.WithOutboxClock(now), app.WithOutboxIDGenerator(newID))

	checkout := app.NewCheckoutService(store, store,
		app.WithProductRepository(store),
		app.WithClock(now),
		app.WithIDGenerator(newID),
	)
	payments := app.NewPaymentService(store, store, store,
		app.WithPaymentClock(now),
		app.WithPaymentIDGenerator(newID),
		app.WithPaymentOutboxRecorder(outbox),
	)
	return checkout, payments
}

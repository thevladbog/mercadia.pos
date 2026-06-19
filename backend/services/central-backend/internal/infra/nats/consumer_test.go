package nats_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"mercadia.dev/pos/services/central-backend/internal/app"
	centralnats "mercadia.dev/pos/services/central-backend/internal/infra/nats"
)

func TestStoreIDFromSubject(t *testing.T) {
	storeID, err := centralnats.StoreIDFromSubject("mercadia.store-edge.sync.store-1")
	if err != nil {
		t.Fatalf("parse subject: %v", err)
	}
	if storeID != "store-1" {
		t.Fatalf("store id = %q", storeID)
	}

	if _, err := centralnats.StoreIDFromSubject("mercadia.other.store-1"); err == nil {
		t.Fatal("expected error for unexpected subject")
	}
}

func TestDecodeSyncMessage(t *testing.T) {
	occurredAt := time.Date(2026, 6, 19, 10, 0, 0, 0, time.UTC)
	raw, err := json.Marshal(map[string]any{
		"eventId":    "evt-1",
		"eventType":  "payment.captured",
		"payload":    map[string]any{"storeId": "store-1"},
		"occurredAt": occurredAt,
	})
	if err != nil {
		t.Fatalf("marshal message: %v", err)
	}

	message, err := centralnats.DecodeSyncMessage(raw)
	if err != nil {
		t.Fatalf("decode message: %v", err)
	}
	if message.EventID != "evt-1" || message.EventType != "payment.captured" {
		t.Fatalf("unexpected message: %+v", message)
	}
	if !message.OccurredAt.Equal(occurredAt) {
		t.Fatalf("occurredAt = %v", message.OccurredAt)
	}
}

func TestIdempotencyKey(t *testing.T) {
	key := centralnats.IdempotencyKey("store-1", "evt-42")
	if key != "nats:store-1:evt-42" {
		t.Fatalf("key = %q", key)
	}
}

type mockSyncAccepter struct {
	lastCommand app.AcceptSyncEventsCommand
	err         error
}

func (m *mockSyncAccepter) AcceptEvents(_ context.Context, command app.AcceptSyncEventsCommand) (app.SyncEventsResult, error) {
	m.lastCommand = command
	if m.err != nil {
		return app.SyncEventsResult{}, m.err
	}
	return app.SyncEventsResult{
		StoreID:  command.StoreID,
		Status:   "accepted",
		Accepted: len(command.Events),
	}, nil
}

func TestConsumerHandleMessageUsesDeterministicIdempotencyKey(t *testing.T) {
	mock := &mockSyncAccepter{}

	occurredAt := time.Date(2026, 6, 19, 10, 0, 0, 0, time.UTC)
	body, err := json.Marshal(map[string]any{
		"eventId":    "evt-99",
		"eventType":  "payment.captured",
		"payload":    map[string]any{"storeId": "store-1"},
		"occurredAt": occurredAt,
	})
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	if err := centralnats.ProcessSyncMessage(context.Background(), mock, "mercadia.store-edge.sync.store-1", body); err != nil {
		t.Fatalf("handle message: %v", err)
	}
	if mock.lastCommand.IdempotencyKey != "nats:store-1:evt-99" {
		t.Fatalf("idempotency key = %q", mock.lastCommand.IdempotencyKey)
	}
	if len(mock.lastCommand.Events) != 1 || mock.lastCommand.Events[0].EventID != "evt-99" {
		t.Fatalf("unexpected command: %+v", mock.lastCommand)
	}
}

func TestConsumerHandleMessagePropagatesAcceptError(t *testing.T) {
	mock := &mockSyncAccepter{err: errors.New("accept failed")}

	body := []byte(`{"eventId":"evt-1","eventType":"payment.captured","payload":{}}`)
	if err := centralnats.ProcessSyncMessage(context.Background(), mock, "mercadia.store-edge.sync.store-1", body); err == nil {
		t.Fatal("expected accept error")
	}
}

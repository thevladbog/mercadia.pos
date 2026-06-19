package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/app"
	"mercadia.dev/pos/services/store-edge/internal/domain"
	"mercadia.dev/pos/services/store-edge/internal/infra/memory"
)

func TestRecordHeartbeatIsIdempotent(t *testing.T) {
	service := newTestTerminalService()
	command := app.RecordTerminalHeartbeatCommand{
		IdempotencyKey:  "heartbeat-1",
		TerminalID:      "pos-1",
		StoreID:         "store-1",
		Kind:            domain.TerminalKindPOS,
		SoftwareVersion: "0.1.0",
	}

	first, err := service.RecordHeartbeat(context.Background(), command)
	if err != nil {
		t.Fatalf("record first heartbeat: %v", err)
	}
	second, err := service.RecordHeartbeat(context.Background(), command)
	if err != nil {
		t.Fatalf("record second heartbeat: %v", err)
	}

	if !first.Terminal.LastSeenAt.Equal(second.Terminal.LastSeenAt) {
		t.Fatalf("expected same idempotent result, got %s and %s", first.Terminal.LastSeenAt, second.Terminal.LastSeenAt)
	}
}

func TestRecordHeartbeatRejectsReusedIdempotencyKeyForDifferentPayload(t *testing.T) {
	service := newTestTerminalService()
	command := app.RecordTerminalHeartbeatCommand{
		IdempotencyKey:  "heartbeat-1",
		TerminalID:      "pos-1",
		StoreID:         "store-1",
		Kind:            domain.TerminalKindPOS,
		SoftwareVersion: "0.1.0",
	}

	if _, err := service.RecordHeartbeat(context.Background(), command); err != nil {
		t.Fatalf("record first heartbeat: %v", err)
	}

	command.SoftwareVersion = "0.1.1"
	_, err := service.RecordHeartbeat(context.Background(), command)
	if !errors.Is(err, app.ErrIdempotencyKeyReused) {
		t.Fatalf("expected ErrIdempotencyKeyReused, got %v", err)
	}
}

func TestGetTerminalReturnsLastHeartbeat(t *testing.T) {
	service := newTestTerminalService()
	_, err := service.RecordHeartbeat(context.Background(), app.RecordTerminalHeartbeatCommand{
		IdempotencyKey:  "heartbeat-1",
		TerminalID:      "pos-1",
		StoreID:         "store-1",
		Kind:            domain.TerminalKindPOS,
		SoftwareVersion: "0.1.0",
	})
	if err != nil {
		t.Fatalf("record heartbeat: %v", err)
	}

	result, err := service.GetTerminal(context.Background(), "pos-1")
	if err != nil {
		t.Fatalf("get terminal: %v", err)
	}

	if result.Terminal.Status != domain.TerminalStatusOnline {
		t.Fatalf("terminal status = %s", result.Terminal.Status)
	}
}

func newTestTerminalService() *app.TerminalService {
	store := memory.NewStore()
	return app.NewTerminalService(store, store, app.WithTerminalClock(func() time.Time {
		return time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
	}))
}

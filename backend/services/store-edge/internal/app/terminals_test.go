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

func TestListStoreTerminalsReturnsSortedPage(t *testing.T) {
	service := newTestTerminalService()
	for _, terminalID := range []string{"pos-2", "pos-1"} {
		_, err := service.RecordHeartbeat(context.Background(), app.RecordTerminalHeartbeatCommand{
			IdempotencyKey:  "heartbeat-" + terminalID,
			TerminalID:      terminalID,
			StoreID:         "store-1",
			Kind:            domain.TerminalKindPOS,
			SoftwareVersion: "0.1.0",
		})
		if err != nil {
			t.Fatalf("record heartbeat for %s: %v", terminalID, err)
		}
	}

	result, err := service.ListStoreTerminals(context.Background(), "store-1", app.PageParams{Limit: 50, Offset: 0})
	if err != nil {
		t.Fatalf("list store terminals: %v", err)
	}
	if result.TotalCount != 2 {
		t.Fatalf("totalCount = %d", result.TotalCount)
	}
	if len(result.Items) != 2 || result.Items[0].ID != "pos-1" || result.Items[1].ID != "pos-2" {
		t.Fatalf("items = %+v", result.Items)
	}
}

func TestListStoreTerminalsDerivesOfflineStatus(t *testing.T) {
	store := memory.NewStore()
	lastSeen := time.Date(2026, 6, 18, 9, 0, 0, 0, time.UTC)
	if err := store.SaveTerminal(context.Background(), domain.Terminal{
		ID:              "pos-1",
		StoreID:         "store-1",
		Kind:            domain.TerminalKindPOS,
		Status:          domain.TerminalStatusOnline,
		SoftwareVersion: "0.1.0",
		LastSeenAt:      lastSeen,
		UpdatedAt:       lastSeen,
	}); err != nil {
		t.Fatalf("save terminal: %v", err)
	}

	service := app.NewTerminalService(store, store,
		app.WithTerminalClock(func() time.Time {
			return time.Date(2026, 6, 18, 10, 2, 0, 0, time.UTC)
		}),
		app.WithTerminalOfflineAfter(time.Minute),
	)

	result, err := service.ListStoreTerminals(context.Background(), "store-1", app.PageParams{Limit: 50, Offset: 0})
	if err != nil {
		t.Fatalf("list store terminals: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("items = %+v", result.Items)
	}
	if result.Items[0].Status != domain.TerminalStatusOffline {
		t.Fatalf("terminal status = %s", result.Items[0].Status)
	}
}

func newTestTerminalService() *app.TerminalService {
	store := memory.NewStore()
	return app.NewTerminalService(store, store, app.WithTerminalClock(func() time.Time {
		return time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
	}))
}

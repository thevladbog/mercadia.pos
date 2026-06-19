package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"mercadia.dev/pos/services/hardware-agent/internal/app"
	"mercadia.dev/pos/services/hardware-agent/internal/domain"
	"mercadia.dev/pos/services/hardware-agent/internal/infra/memory"
	"mercadia.dev/pos/services/hardware-agent/internal/infra/simulated"
)

func TestSendCommandIsIdempotent(t *testing.T) {
	service := newTestDeviceService()
	command := app.SendDeviceCommand{
		IdempotencyKey: "cmd-1",
		DeviceID:       "sim-fiscal-1",
		Type:           "get_status",
	}

	first, err := service.SendCommand(context.Background(), command)
	if err != nil {
		t.Fatalf("send first command: %v", err)
	}
	second, err := service.SendCommand(context.Background(), command)
	if err != nil {
		t.Fatalf("send second command: %v", err)
	}

	if first.Command.ID != second.Command.ID {
		t.Fatalf("expected same command id, got %s and %s", first.Command.ID, second.Command.ID)
	}
}

func TestSendCommandRejectsReusedIdempotencyKeyForDifferentPayload(t *testing.T) {
	service := newTestDeviceService()
	command := app.SendDeviceCommand{
		IdempotencyKey: "cmd-1",
		DeviceID:       "sim-fiscal-1",
		Type:           "get_status",
	}

	if _, err := service.SendCommand(context.Background(), command); err != nil {
		t.Fatalf("send first command: %v", err)
	}

	command.Type = "print_receipt"
	command.Payload = map[string]any{"totalMinor": 1500}
	_, err := service.SendCommand(context.Background(), command)
	if !errors.Is(err, app.ErrIdempotencyKeyReused) {
		t.Fatalf("expected ErrIdempotencyKeyReused, got %v", err)
	}
}

func TestSendCommandCompletesAsynchronously(t *testing.T) {
	service := newTestDeviceService()
	result, err := service.SendCommand(context.Background(), app.SendDeviceCommand{
		IdempotencyKey: "cmd-async",
		DeviceID:       "sim-fiscal-1",
		Type:           "print_receipt",
		Payload:        map[string]any{"totalMinor": 2500},
	})
	if err != nil {
		t.Fatalf("send command: %v", err)
	}
	if result.Command.Status != domain.CommandStatusAccepted {
		t.Fatalf("initial status = %s", result.Command.Status)
	}

	service.Wait()

	final, err := service.GetCommand(context.Background(), "sim-fiscal-1", result.Command.ID)
	if err != nil {
		t.Fatalf("get command: %v", err)
	}
	if final.Command.Status != domain.CommandStatusCompleted {
		t.Fatalf("final status = %s", final.Command.Status)
	}
	if final.Command.Result["fiscalSign"] == nil {
		t.Fatal("expected fiscal sign in result")
	}
}

func TestPaymentTerminalAuthorizeReturnsApproval(t *testing.T) {
	service := newTestDeviceService()
	result, err := service.SendCommand(context.Background(), app.SendDeviceCommand{
		IdempotencyKey: "pay-1",
		DeviceID:       "sim-payment-1",
		Type:           "authorize",
		Payload:        map[string]any{"amountMinor": 9900, "currency": "RUB"},
	})
	if err != nil {
		t.Fatalf("send command: %v", err)
	}

	service.Wait()

	final, err := service.GetCommand(context.Background(), "sim-payment-1", result.Command.ID)
	if err != nil {
		t.Fatalf("get command: %v", err)
	}
	if final.Command.Result["status"] != "approved" {
		t.Fatalf("status = %v", final.Command.Result["status"])
	}
	if final.Command.Result["authCode"] != "A1B2C3" {
		t.Fatalf("authCode = %v", final.Command.Result["authCode"])
	}
}

func TestGetDeviceReturnsSeededDevice(t *testing.T) {
	service := newTestDeviceService()
	result, err := service.GetDevice(context.Background(), "sim-scanner-1")
	if err != nil {
		t.Fatalf("get device: %v", err)
	}
	if result.Device.Kind != domain.DeviceKindScanner {
		t.Fatalf("kind = %s", result.Device.Kind)
	}
}

func newTestDeviceService() *app.DeviceService {
	store := memory.NewStore()
	return app.NewDeviceService(store, store, store, simulated.DefaultRegistry(),
		app.WithExecutionDelay(0),
		app.WithDeviceIDGenerator(func(prefix string) string {
			return prefix + "_test"
		}),
		app.WithDeviceClock(func() time.Time {
			return time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
		}),
	)
}

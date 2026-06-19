package memory

import (
	"context"
	"sync"
	"time"

	"mercadia.dev/pos/services/hardware-agent/internal/app"
	"mercadia.dev/pos/services/hardware-agent/internal/domain"
)

type Store struct {
	mu          sync.RWMutex
	devices     map[string]domain.Device
	commands    map[string]domain.DeviceCommand
	idempotency map[string]app.IdempotencyRecord
}

func NewStore() *Store {
	store := &Store{
		devices:     map[string]domain.Device{},
		commands:    map[string]domain.DeviceCommand{},
		idempotency: map[string]app.IdempotencyRecord{},
	}
	store.seedDevices()
	return store
}

func (s *Store) seedDevices() {
	now := time.Now().UTC()
	devices := []domain.Device{
		{ID: "sim-fiscal-1", Kind: domain.DeviceKindFiscal, Status: domain.DeviceStatusSimulated, Model: "ATOL 42F", UpdatedAt: now},
		{ID: "sim-payment-1", Kind: domain.DeviceKindPaymentTerminal, Status: domain.DeviceStatusSimulated, Model: "Ingenico iPP350", UpdatedAt: now},
		{ID: "sim-msr-1", Kind: domain.DeviceKindMSR, Status: domain.DeviceStatusSimulated, Model: "MagTek MSR605", UpdatedAt: now},
		{ID: "sim-ibutton-1", Kind: domain.DeviceKindIButton, Status: domain.DeviceStatusSimulated, Model: "DS9490R", UpdatedAt: now},
		{ID: "sim-scanner-1", Kind: domain.DeviceKindScanner, Status: domain.DeviceStatusSimulated, Model: "Honeywell 1900", UpdatedAt: now},
		{ID: "sim-printer-1", Kind: domain.DeviceKindPrinter, Status: domain.DeviceStatusSimulated, Model: "Epson TM-T88", UpdatedAt: now},
	}
	for _, device := range devices {
		s.devices[device.ID] = device
	}
}

func (s *Store) ListDevices(_ context.Context) ([]domain.Device, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	devices := make([]domain.Device, 0, len(s.devices))
	for _, device := range s.devices {
		devices = append(devices, device)
	}
	return devices, nil
}

func (s *Store) FindDevice(_ context.Context, deviceID string) (domain.Device, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	device, ok := s.devices[deviceID]
	if !ok {
		return domain.Device{}, app.ErrDeviceNotFound
	}
	return device, nil
}

func (s *Store) SaveCommand(_ context.Context, command domain.DeviceCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.commands[command.ID] = cloneCommand(command)
	return nil
}

func (s *Store) FindCommand(_ context.Context, commandID string) (domain.DeviceCommand, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	command, ok := s.commands[commandID]
	if !ok {
		return domain.DeviceCommand{}, app.ErrCommandNotFound
	}
	return cloneCommand(command), nil
}

func (s *Store) Find(ctx context.Context, operation string, key string) (app.IdempotencyRecord, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.idempotency[idempotencyMapKey(operation, key)]
	return record, ok, nil
}

func (s *Store) Save(ctx context.Context, record app.IdempotencyRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.idempotency[idempotencyMapKey(record.Operation, record.Key)] = record
	return nil
}

func idempotencyMapKey(operation string, key string) string {
	return operation + "\x00" + key
}

func cloneCommand(command domain.DeviceCommand) domain.DeviceCommand {
	if command.Payload != nil {
		payload := make(map[string]any, len(command.Payload))
		for key, value := range command.Payload {
			payload[key] = value
		}
		command.Payload = payload
	}
	if command.Result != nil {
		result := make(map[string]any, len(command.Result))
		for key, value := range command.Result {
			result[key] = value
		}
		command.Result = result
	}
	return command
}

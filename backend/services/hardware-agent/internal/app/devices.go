package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"mercadia.dev/pos/services/hardware-agent/internal/domain"
)

var (
	ErrDeviceNotFound        = errors.New("device not found")
	ErrCommandNotFound       = errors.New("command not found")
	ErrInvalidDeviceCommand  = errors.New("invalid device command")
	ErrUnsupportedCommand    = errors.New("unsupported device command")
)

type DeviceRepository interface {
	ListDevices(ctx context.Context) ([]domain.Device, error)
	FindDevice(ctx context.Context, deviceID string) (domain.Device, error)
}

type CommandRepository interface {
	SaveCommand(ctx context.Context, command domain.DeviceCommand) error
	FindCommand(ctx context.Context, commandID string) (domain.DeviceCommand, error)
}

type DeviceExecutor interface {
	Execute(ctx context.Context, device domain.Device, commandType string, payload map[string]any) (map[string]any, error)
}

type DeviceService struct {
	devices     DeviceRepository
	commands    CommandRepository
	idempotency IdempotencyStore
	executor    DeviceExecutor
	terminalID  string
	now         func() time.Time
	newID       func(prefix string) string
	delay       time.Duration
	wg          sync.WaitGroup
}

type DeviceOption func(*DeviceService)

func NewDeviceService(devices DeviceRepository, commands CommandRepository, idempotency IdempotencyStore, executor DeviceExecutor, options ...DeviceOption) *DeviceService {
	service := &DeviceService{
		devices:     devices,
		commands:    commands,
		idempotency: idempotency,
		executor:    executor,
		terminalID:  "local-terminal",
		now: func() time.Time {
			return time.Now().UTC()
		},
		newID: randomID,
		delay: 25 * time.Millisecond,
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func WithTerminalID(terminalID string) DeviceOption {
	return func(service *DeviceService) {
		service.terminalID = terminalID
	}
}

func WithDeviceClock(now func() time.Time) DeviceOption {
	return func(service *DeviceService) {
		service.now = now
	}
}

func WithDeviceIDGenerator(newID func(prefix string) string) DeviceOption {
	return func(service *DeviceService) {
		service.newID = newID
	}
}

func WithExecutionDelay(delay time.Duration) DeviceOption {
	return func(service *DeviceService) {
		service.delay = delay
	}
}

type AgentStatus struct {
	TerminalID  string
	Status      string
	DeviceCount int
	GeneratedAt time.Time
}

type SendDeviceCommand struct {
	IdempotencyKey string
	DeviceID       string
	Type           string
	Payload        map[string]any
}

type DeviceCommandResult struct {
	Command domain.DeviceCommand
}

type DeviceResult struct {
	Device domain.Device
}

func (s *DeviceService) GetAgentStatus(ctx context.Context) (AgentStatus, error) {
	devices, err := s.devices.ListDevices(ctx)
	if err != nil {
		return AgentStatus{}, err
	}
	return AgentStatus{
		TerminalID:  s.terminalID,
		Status:      "ok",
		DeviceCount: len(devices),
		GeneratedAt: s.now(),
	}, nil
}

func (s *DeviceService) ListDevices(ctx context.Context) ([]domain.Device, error) {
	return s.devices.ListDevices(ctx)
}

func (s *DeviceService) GetDevice(ctx context.Context, deviceID string) (DeviceResult, error) {
	device, err := s.devices.FindDevice(ctx, deviceID)
	if err != nil {
		return DeviceResult{}, err
	}
	return DeviceResult{Device: device}, nil
}

func (s *DeviceService) GetCommand(ctx context.Context, deviceID string, commandID string) (DeviceCommandResult, error) {
	command, err := s.commands.FindCommand(ctx, commandID)
	if err != nil {
		return DeviceCommandResult{}, err
	}
	if command.DeviceID != deviceID {
		return DeviceCommandResult{}, ErrCommandNotFound
	}
	return DeviceCommandResult{Command: command}, nil
}

func (s *DeviceService) SendCommand(ctx context.Context, command SendDeviceCommand) (DeviceCommandResult, error) {
	if command.IdempotencyKey == "" {
		return DeviceCommandResult{}, ErrIdempotencyKeyRequired
	}
	if command.DeviceID == "" || command.Type == "" {
		return DeviceCommandResult{}, ErrInvalidDeviceCommand
	}

	const operation = "devices.send_command"
	fingerprint, err := commandFingerprint(command.DeviceID, command.Type, command.Payload)
	if err != nil {
		return DeviceCommandResult{}, err
	}
	if result, found, err := s.findCommandIdempotency(ctx, operation, command.IdempotencyKey, command.DeviceID, fingerprint); err != nil || found {
		return result, err
	}

	device, err := s.devices.FindDevice(ctx, command.DeviceID)
	if err != nil {
		return DeviceCommandResult{}, err
	}

	created, err := domain.NewDeviceCommand(domain.NewDeviceCommandInput{
		ID:       s.newID("cmd"),
		DeviceID: command.DeviceID,
		Type:     command.Type,
		Payload:  command.Payload,
		Now:      s.now(),
	})
	if err != nil {
		return DeviceCommandResult{}, err
	}

	if err := s.commands.SaveCommand(ctx, created); err != nil {
		return DeviceCommandResult{}, err
	}

	result := DeviceCommandResult{Command: created}
	if err := s.idempotency.Save(ctx, IdempotencyRecord{
		Operation:   operation,
		Key:         command.IdempotencyKey,
		TargetID:    command.DeviceID,
		Fingerprint: fingerprint,
		Result:      result,
		CreatedAt:   s.now(),
	}); err != nil {
		return DeviceCommandResult{}, err
	}

	s.wg.Add(1)
	go s.runCommand(device, created.ID)

	return result, nil
}

func (s *DeviceService) Wait() {
	s.wg.Wait()
}

func (s *DeviceService) runCommand(device domain.Device, commandID string) {
	defer s.wg.Done()

	ctx := context.Background()
	command, err := s.commands.FindCommand(ctx, commandID)
	if err != nil {
		return
	}

	running := command.WithStatus(domain.CommandStatusRunning, s.now())
	_ = s.commands.SaveCommand(ctx, running)

	if s.delay > 0 {
		time.Sleep(s.delay)
	}

	result, err := s.executor.Execute(ctx, device, command.Type, command.Payload)
	now := s.now()
	if err != nil {
		failed := running.WithFailure(err.Error(), now)
		_ = s.commands.SaveCommand(ctx, failed)
		return
	}

	completed := running.WithResult(result, now)
	_ = s.commands.SaveCommand(ctx, completed)
}

func (s *DeviceService) findCommandIdempotency(ctx context.Context, operation string, key string, targetID string, fingerprint string) (DeviceCommandResult, bool, error) {
	record, found, err := s.idempotency.Find(ctx, operation, key)
	if err != nil || !found {
		return DeviceCommandResult{}, found, err
	}
	if record.TargetID != targetID || record.Fingerprint != fingerprint {
		return DeviceCommandResult{}, true, ErrIdempotencyKeyReused
	}
	result, ok := record.Result.(DeviceCommandResult)
	if !ok {
		return DeviceCommandResult{}, true, ErrIdempotencyResultMissing
	}
	command, err := s.commands.FindCommand(ctx, result.Command.ID)
	if err != nil {
		return DeviceCommandResult{}, true, err
	}
	return DeviceCommandResult{Command: command}, true, nil
}

func commandFingerprint(deviceID string, commandType string, payload map[string]any) (string, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode command payload: %w", err)
	}
	return fmt.Sprintf("%s|%s|%s", deviceID, commandType, string(encoded)), nil
}

func randomID(prefix string) string {
	var bytes [12]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		panic(fmt.Sprintf("generate id: %v", err))
	}
	return prefix + "_" + hex.EncodeToString(bytes[:])
}

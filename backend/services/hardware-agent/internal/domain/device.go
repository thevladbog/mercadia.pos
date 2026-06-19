package domain

import (
	"errors"
	"time"
)

type DeviceKind string

const (
	DeviceKindFiscal          DeviceKind = "fiscal"
	DeviceKindPaymentTerminal DeviceKind = "payment_terminal"
	DeviceKindMSR             DeviceKind = "msr"
	DeviceKindIButton         DeviceKind = "ibutton"
	DeviceKindScanner         DeviceKind = "scanner"
	DeviceKindPrinter         DeviceKind = "printer"
)

type DeviceStatus string

const (
	DeviceStatusReady     DeviceStatus = "ready"
	DeviceStatusBusy      DeviceStatus = "busy"
	DeviceStatusOffline   DeviceStatus = "offline"
	DeviceStatusSimulated DeviceStatus = "simulated"
	DeviceStatusError     DeviceStatus = "error"
)

type CommandStatus string

const (
	CommandStatusAccepted  CommandStatus = "accepted"
	CommandStatusRunning   CommandStatus = "running"
	CommandStatusCompleted CommandStatus = "completed"
	CommandStatusFailed    CommandStatus = "failed"
)

var (
	ErrInvalidDeviceInput  = errors.New("invalid device input")
	ErrInvalidCommandInput = errors.New("invalid command input")
)

type Device struct {
	ID        string
	Kind      DeviceKind
	Status    DeviceStatus
	Model     string
	UpdatedAt time.Time
}

type DeviceCommand struct {
	ID          string
	DeviceID    string
	Type        string
	Payload     map[string]any
	Status      CommandStatus
	Result      map[string]any
	Error       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
}

type NewDeviceCommandInput struct {
	ID        string
	DeviceID  string
	Type      string
	Payload   map[string]any
	Now       time.Time
}

func NewDeviceCommand(input NewDeviceCommandInput) (DeviceCommand, error) {
	if input.ID == "" || input.DeviceID == "" || input.Type == "" {
		return DeviceCommand{}, ErrInvalidCommandInput
	}
	if input.Now.IsZero() {
		input.Now = time.Now().UTC()
	}
	payload := input.Payload
	if payload == nil {
		payload = map[string]any{}
	}

	return DeviceCommand{
		ID:        input.ID,
		DeviceID:  input.DeviceID,
		Type:      input.Type,
		Payload:   payload,
		Status:    CommandStatusAccepted,
		CreatedAt: input.Now,
		UpdatedAt: input.Now,
	}, nil
}

func (command DeviceCommand) WithStatus(status CommandStatus, now time.Time) DeviceCommand {
	command.Status = status
	command.UpdatedAt = now
	if status == CommandStatusCompleted || status == CommandStatusFailed {
		completedAt := now
		command.CompletedAt = &completedAt
	}
	return command
}

func (command DeviceCommand) WithResult(result map[string]any, now time.Time) DeviceCommand {
	command.Result = result
	command.Status = CommandStatusCompleted
	command.UpdatedAt = now
	completedAt := now
	command.CompletedAt = &completedAt
	return command
}

func (command DeviceCommand) WithFailure(message string, now time.Time) DeviceCommand {
	command.Error = message
	command.Status = CommandStatusFailed
	command.UpdatedAt = now
	completedAt := now
	command.CompletedAt = &completedAt
	return command
}

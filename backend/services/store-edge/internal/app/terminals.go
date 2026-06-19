package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/domain"
)

var (
	ErrTerminalNotFound       = errors.New("terminal not found")
	ErrInvalidTerminalCommand = errors.New("invalid terminal command")
)

type TerminalRepository interface {
	SaveTerminal(ctx context.Context, terminal domain.Terminal) error
	FindTerminal(ctx context.Context, terminalID string) (domain.Terminal, error)
}

type TerminalService struct {
	terminals   TerminalRepository
	idempotency IdempotencyStore
	now         func() time.Time
}

type TerminalOption func(*TerminalService)

func NewTerminalService(terminals TerminalRepository, idempotency IdempotencyStore, options ...TerminalOption) *TerminalService {
	service := &TerminalService{
		terminals:   terminals,
		idempotency: idempotency,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func WithTerminalClock(now func() time.Time) TerminalOption {
	return func(service *TerminalService) {
		service.now = now
	}
}

type RecordTerminalHeartbeatCommand struct {
	IdempotencyKey  string
	TerminalID      string
	StoreID         string
	Kind            domain.TerminalKind
	SoftwareVersion string
}

type TerminalResult struct {
	Terminal domain.Terminal
}

func (s *TerminalService) RecordHeartbeat(ctx context.Context, command RecordTerminalHeartbeatCommand) (TerminalResult, error) {
	if command.IdempotencyKey == "" {
		return TerminalResult{}, ErrIdempotencyKeyRequired
	}
	if command.TerminalID == "" || command.StoreID == "" || command.Kind == "" {
		return TerminalResult{}, ErrInvalidTerminalCommand
	}

	const operation = "terminals.record_heartbeat"
	fingerprint := fmt.Sprintf("%s|%s|%s|%s", command.TerminalID, command.StoreID, command.Kind, command.SoftwareVersion)
	if result, found, err := s.findTerminalIdempotency(ctx, operation, command.IdempotencyKey, command.TerminalID, fingerprint); err != nil || found {
		return result, err
	}

	terminal, err := domain.RecordTerminalHeartbeat(domain.RecordTerminalHeartbeatInput{
		ID:              command.TerminalID,
		StoreID:         command.StoreID,
		Kind:            command.Kind,
		SoftwareVersion: command.SoftwareVersion,
		Now:             s.now(),
	})
	if err != nil {
		return TerminalResult{}, err
	}

	if err := s.terminals.SaveTerminal(ctx, terminal); err != nil {
		return TerminalResult{}, err
	}

	result := TerminalResult{Terminal: terminal}
	if err := s.idempotency.Save(ctx, IdempotencyRecord{
		Operation:   operation,
		Key:         command.IdempotencyKey,
		TargetID:    command.TerminalID,
		Fingerprint: fingerprint,
		Result:      result,
		CreatedAt:   s.now(),
	}); err != nil {
		return TerminalResult{}, err
	}

	return result, nil
}

func (s *TerminalService) GetTerminal(ctx context.Context, terminalID string) (TerminalResult, error) {
	terminal, err := s.terminals.FindTerminal(ctx, terminalID)
	if err != nil {
		return TerminalResult{}, err
	}
	return TerminalResult{Terminal: terminal}, nil
}

func (s *TerminalService) findTerminalIdempotency(ctx context.Context, operation string, key string, targetID string, fingerprint string) (TerminalResult, bool, error) {
	record, found, err := s.idempotency.Find(ctx, operation, key)
	if err != nil || !found {
		return TerminalResult{}, found, err
	}
	if record.TargetID != targetID || record.Fingerprint != fingerprint {
		return TerminalResult{}, true, ErrIdempotencyKeyReused
	}
	result, ok := record.Result.(TerminalResult)
	if !ok {
		return TerminalResult{}, true, ErrIdempotencyResultMissing
	}
	return result, true, nil
}

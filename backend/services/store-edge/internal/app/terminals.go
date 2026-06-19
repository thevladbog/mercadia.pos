package app

import (
	"context"
	"errors"
	"fmt"
	"sort"
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
	ListTerminalsByStore(ctx context.Context, storeID string) ([]domain.Terminal, error)
}

type TerminalEventPublisher interface {
	PublishTerminalHeartbeat(terminal domain.Terminal)
}

type TerminalService struct {
	terminals    TerminalRepository
	idempotency  IdempotencyStore
	events       TerminalEventPublisher
	now          func() time.Time
	offlineAfter time.Duration
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

func WithTerminalEventPublisher(events TerminalEventPublisher) TerminalOption {
	return func(service *TerminalService) {
		service.events = events
	}
}

func WithTerminalOfflineAfter(duration time.Duration) TerminalOption {
	return func(service *TerminalService) {
		service.offlineAfter = duration
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
	if s.events != nil {
		s.events.PublishTerminalHeartbeat(terminal)
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

func (s *TerminalService) ListStoreTerminals(ctx context.Context, storeID string, params PageParams) (PageResult[domain.Terminal], error) {
	if storeID == "" {
		return PageResult[domain.Terminal]{}, ErrInvalidTerminalCommand
	}

	terminals, err := s.terminals.ListTerminalsByStore(ctx, storeID)
	if err != nil {
		return PageResult[domain.Terminal]{}, err
	}

	sort.Slice(terminals, func(i, j int) bool {
		return terminals[i].ID < terminals[j].ID
	})

	for i := range terminals {
		terminals[i].Status = DeriveTerminalListStatus(terminals[i], s.now(), s.offlineAfter)
	}

	return PaginateSlice(terminals, params), nil
}

func DeriveTerminalListStatus(terminal domain.Terminal, now time.Time, offlineAfter time.Duration) domain.TerminalStatus {
	if offlineAfter > 0 && now.Sub(terminal.LastSeenAt) > offlineAfter {
		return domain.TerminalStatusOffline
	}
	return terminal.Status
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

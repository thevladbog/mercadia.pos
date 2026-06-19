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
	ErrCashMovementNotFound           = errors.New("cash movement not found")
	ErrCashRecountNotFound            = errors.New("cash recount not found")
	ErrInvalidCashMovementCommand     = errors.New("invalid cash movement command")
	ErrInvalidCashRecountCommand      = errors.New("invalid cash recount command")
	ErrCashRecountApprovalRequired    = errors.New("cash recount discrepancy requires approval")
	ErrCashRecountResolutionNotNeeded = errors.New("cash recount resolution is not needed")
	ErrCashRecountAlreadyResolved     = errors.New("cash recount already resolved")
	ErrSeparationOfDutiesViolation    = errors.New("actor cannot approve their own cash movement")
)

type CashRepository interface {
	SaveCashMovement(ctx context.Context, movement domain.CashMovement) error
	ListCashMovements(ctx context.Context, storeID string) ([]domain.CashMovement, error)
	SaveCashRecount(ctx context.Context, recount domain.CashRecount) error
	FindCashRecount(ctx context.Context, recountID string) (domain.CashRecount, error)
	ListCashRecounts(ctx context.Context, storeID string) ([]domain.CashRecount, error)
}

type CashService struct {
	cash        CashRepository
	idempotency IdempotencyStore
	now         func() time.Time
	newID       func(prefix string) string
}

type CashOption func(*CashService)

func NewCashService(cash CashRepository, idempotency IdempotencyStore, options ...CashOption) *CashService {
	service := &CashService{
		cash:        cash,
		idempotency: idempotency,
		now: func() time.Time {
			return time.Now().UTC()
		},
		newID: randomID,
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func WithCashClock(now func() time.Time) CashOption {
	return func(service *CashService) {
		service.now = now
	}
}

func WithCashIDGenerator(newID func(prefix string) string) CashOption {
	return func(service *CashService) {
		service.newID = newID
	}
}

type CreateCashMovementCommand struct {
	IdempotencyKey    string
	StoreID           string
	Type              domain.CashMovementType
	FromContainerID   string
	FromContainerType domain.CashContainerType
	ToContainerID     string
	ToContainerType   domain.CashContainerType
	AmountMinor       int64
	Currency          string
	Reason            string
	ActorID           string
	ApprovedByID      string
}

type CashMovementResult struct {
	Movement domain.CashMovement
}

type CreateCashRecountCommand struct {
	IdempotencyKey string
	StoreID        string
	ContainerID    string
	ContainerType  domain.CashContainerType
	Currency       string
	CountedMinor   int64
	Reason         string
	ActorID        string
	ApprovedByID   string
}

type ResolveCashRecountCommand struct {
	IdempotencyKey string
	StoreID        string
	RecountID      string
	ResolutionNote string
	ActorID        string
	ApprovedByID   string
}

type CashRecountResult struct {
	Recount domain.CashRecount
}

func (s *CashService) CreateCashMovement(ctx context.Context, command CreateCashMovementCommand) (CashMovementResult, error) {
	if command.IdempotencyKey == "" {
		return CashMovementResult{}, ErrIdempotencyKeyRequired
	}
	if command.StoreID == "" || command.Type == "" || command.FromContainerID == "" || command.FromContainerType == "" ||
		command.ToContainerID == "" || command.ToContainerType == "" || command.AmountMinor <= 0 || command.ActorID == "" {
		return CashMovementResult{}, ErrInvalidCashMovementCommand
	}
	if command.ApprovedByID != "" && command.ApprovedByID == command.ActorID {
		return CashMovementResult{}, ErrSeparationOfDutiesViolation
	}

	const operation = "cash.create_cash_movement"
	fingerprint := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%d|%s|%s|%s|%s",
		command.StoreID,
		command.Type,
		command.FromContainerID,
		command.FromContainerType,
		command.ToContainerID,
		command.ToContainerType,
		command.AmountMinor,
		command.Currency,
		command.Reason,
		command.ActorID,
		command.ApprovedByID,
	)
	if result, found, err := s.findCashIdempotency(ctx, operation, command.IdempotencyKey, command.StoreID, fingerprint); err != nil || found {
		return result, err
	}

	movement, err := domain.CreateCashMovement(domain.CreateCashMovementInput{
		ID:                s.newID("cash"),
		StoreID:           command.StoreID,
		Type:              command.Type,
		FromContainerID:   command.FromContainerID,
		FromContainerType: command.FromContainerType,
		ToContainerID:     command.ToContainerID,
		ToContainerType:   command.ToContainerType,
		AmountMinor:       command.AmountMinor,
		Currency:          command.Currency,
		Reason:            command.Reason,
		ActorID:           command.ActorID,
		ApprovedByID:      command.ApprovedByID,
		Now:               s.now(),
	})
	if err != nil {
		return CashMovementResult{}, err
	}

	if err := s.cash.SaveCashMovement(ctx, movement); err != nil {
		return CashMovementResult{}, err
	}

	result := CashMovementResult{Movement: movement}
	if err := s.idempotency.Save(ctx, IdempotencyRecord{
		Operation:   operation,
		Key:         command.IdempotencyKey,
		TargetID:    command.StoreID,
		Fingerprint: fingerprint,
		Result:      result,
		CreatedAt:   s.now(),
	}); err != nil {
		return CashMovementResult{}, err
	}

	return result, nil
}

func (s *CashService) CreateCashRecount(ctx context.Context, command CreateCashRecountCommand) (CashRecountResult, error) {
	if command.IdempotencyKey == "" {
		return CashRecountResult{}, ErrIdempotencyKeyRequired
	}
	if command.StoreID == "" || command.ContainerID == "" || command.ContainerType == "" ||
		command.CountedMinor < 0 || command.ActorID == "" {
		return CashRecountResult{}, ErrInvalidCashRecountCommand
	}
	if command.ApprovedByID != "" && command.ApprovedByID == command.ActorID {
		return CashRecountResult{}, ErrSeparationOfDutiesViolation
	}

	const operation = "cash.create_cash_recount"
	fingerprint := fmt.Sprintf("%s|%s|%s|%s|%d|%s|%s|%s",
		command.StoreID,
		command.ContainerID,
		command.ContainerType,
		command.Currency,
		command.CountedMinor,
		command.Reason,
		command.ActorID,
		command.ApprovedByID,
	)
	if result, found, err := s.findCashRecountIdempotency(ctx, operation, command.IdempotencyKey, command.StoreID, fingerprint); err != nil || found {
		return result, err
	}

	expected, err := s.expectedBalance(ctx, command.StoreID, command.ContainerID, command.ContainerType, command.Currency)
	if err != nil {
		return CashRecountResult{}, err
	}
	if expected != command.CountedMinor && command.ApprovedByID == "" {
		return CashRecountResult{}, ErrCashRecountApprovalRequired
	}

	recount, err := domain.CreateCashRecount(domain.CreateCashRecountInput{
		ID:            s.newID("crec"),
		StoreID:       command.StoreID,
		ContainerID:   command.ContainerID,
		ContainerType: command.ContainerType,
		Currency:      command.Currency,
		ExpectedMinor: expected,
		CountedMinor:  command.CountedMinor,
		Reason:        command.Reason,
		ActorID:       command.ActorID,
		ApprovedByID:  command.ApprovedByID,
		Now:           s.now(),
	})
	if err != nil {
		return CashRecountResult{}, err
	}

	if err := s.cash.SaveCashRecount(ctx, recount); err != nil {
		return CashRecountResult{}, err
	}

	result := CashRecountResult{Recount: recount}
	if err := s.idempotency.Save(ctx, IdempotencyRecord{
		Operation:   operation,
		Key:         command.IdempotencyKey,
		TargetID:    command.StoreID,
		Fingerprint: fingerprint,
		Result:      result,
		CreatedAt:   s.now(),
	}); err != nil {
		return CashRecountResult{}, err
	}

	return result, nil
}

func (s *CashService) ResolveCashRecount(ctx context.Context, command ResolveCashRecountCommand) (CashRecountResult, error) {
	if command.IdempotencyKey == "" {
		return CashRecountResult{}, ErrIdempotencyKeyRequired
	}
	if command.StoreID == "" || command.RecountID == "" || command.ResolutionNote == "" ||
		command.ActorID == "" || command.ApprovedByID == "" {
		return CashRecountResult{}, ErrInvalidCashRecountCommand
	}
	if command.ApprovedByID == command.ActorID {
		return CashRecountResult{}, ErrSeparationOfDutiesViolation
	}

	const operation = "cash.resolve_cash_recount"
	fingerprint := fmt.Sprintf("%s|%s|%s|%s|%s",
		command.StoreID,
		command.RecountID,
		command.ResolutionNote,
		command.ActorID,
		command.ApprovedByID,
	)
	if result, found, err := s.findCashRecountIdempotency(ctx, operation, command.IdempotencyKey, command.RecountID, fingerprint); err != nil || found {
		return result, err
	}

	recount, err := s.cash.FindCashRecount(ctx, command.RecountID)
	if err != nil {
		return CashRecountResult{}, err
	}
	if recount.StoreID != command.StoreID {
		return CashRecountResult{}, ErrCashRecountNotFound
	}
	if err := recount.Resolve(command.ResolutionNote, command.ActorID, s.now()); err != nil {
		if errors.Is(err, domain.ErrCashRecountResolutionNotNeeded) {
			return CashRecountResult{}, ErrCashRecountResolutionNotNeeded
		}
		if errors.Is(err, domain.ErrCashRecountAlreadyResolved) {
			return CashRecountResult{}, ErrCashRecountAlreadyResolved
		}
		return CashRecountResult{}, err
	}

	if err := s.cash.SaveCashRecount(ctx, recount); err != nil {
		return CashRecountResult{}, err
	}

	result := CashRecountResult{Recount: recount}
	if err := s.idempotency.Save(ctx, IdempotencyRecord{
		Operation:   operation,
		Key:         command.IdempotencyKey,
		TargetID:    command.RecountID,
		Fingerprint: fingerprint,
		Result:      result,
		CreatedAt:   s.now(),
	}); err != nil {
		return CashRecountResult{}, err
	}

	return result, nil
}

func (s *CashService) ListCashMovements(ctx context.Context, storeID string) ([]domain.CashMovement, error) {
	if storeID == "" {
		return nil, ErrInvalidCashMovementCommand
	}
	return s.cash.ListCashMovements(ctx, storeID)
}

func (s *CashService) ListCashRecounts(ctx context.Context, storeID string) ([]domain.CashRecount, error) {
	if storeID == "" {
		return nil, ErrInvalidCashRecountCommand
	}
	return s.cash.ListCashRecounts(ctx, storeID)
}

func (s *CashService) ListCashBalances(ctx context.Context, storeID string) ([]domain.CashBalance, error) {
	if storeID == "" {
		return nil, ErrInvalidCashMovementCommand
	}
	movements, err := s.cash.ListCashMovements(ctx, storeID)
	if err != nil {
		return nil, err
	}

	balances := map[string]domain.CashBalance{}
	for _, movement := range movements {
		if movement.Status != domain.CashMovementStatusPosted {
			continue
		}
		applyCashBalanceDelta(balances, movement.StoreID, movement.FromContainerID, movement.FromContainerType, movement.Currency, -movement.AmountMinor, movement.CreatedAt)
		applyCashBalanceDelta(balances, movement.StoreID, movement.ToContainerID, movement.ToContainerType, movement.Currency, movement.AmountMinor, movement.CreatedAt)
	}

	result := make([]domain.CashBalance, 0, len(balances))
	for _, balance := range balances {
		result = append(result, balance)
	}
	sort.Slice(result, func(i int, j int) bool {
		if result[i].ContainerType != result[j].ContainerType {
			return result[i].ContainerType < result[j].ContainerType
		}
		if result[i].ContainerID != result[j].ContainerID {
			return result[i].ContainerID < result[j].ContainerID
		}
		return result[i].Currency < result[j].Currency
	})

	return result, nil
}

func (s *CashService) expectedBalance(ctx context.Context, storeID string, containerID string, containerType domain.CashContainerType, currency string) (int64, error) {
	if currency == "" {
		currency = "RUB"
	}
	balances, err := s.ListCashBalances(ctx, storeID)
	if err != nil {
		return 0, err
	}
	for _, balance := range balances {
		if balance.ContainerID == containerID && balance.ContainerType == containerType && balance.Currency == currency {
			return balance.BalanceMinor, nil
		}
	}
	return 0, nil
}

func applyCashBalanceDelta(balances map[string]domain.CashBalance, storeID string, containerID string, containerType domain.CashContainerType, currency string, deltaMinor int64, movementAt time.Time) {
	if containerType == domain.CashContainerTypeExternal {
		return
	}
	key := fmt.Sprintf("%s|%s|%s", containerType, containerID, currency)
	balance := balances[key]
	if balance.StoreID == "" {
		balance.StoreID = storeID
		balance.ContainerID = containerID
		balance.ContainerType = containerType
		balance.Currency = currency
	}
	balance.BalanceMinor += deltaMinor
	if movementAt.After(balance.LastMovementAt) {
		balance.LastMovementAt = movementAt
	}
	balances[key] = balance
}

func (s *CashService) findCashIdempotency(ctx context.Context, operation string, key string, targetID string, fingerprint string) (CashMovementResult, bool, error) {
	record, found, err := s.idempotency.Find(ctx, operation, key)
	if err != nil || !found {
		return CashMovementResult{}, found, err
	}
	if record.TargetID != targetID || record.Fingerprint != fingerprint {
		return CashMovementResult{}, true, ErrIdempotencyKeyReused
	}
	result, ok := record.Result.(CashMovementResult)
	if !ok {
		return CashMovementResult{}, true, ErrIdempotencyResultMissing
	}
	return result, true, nil
}

func (s *CashService) findCashRecountIdempotency(ctx context.Context, operation string, key string, targetID string, fingerprint string) (CashRecountResult, bool, error) {
	record, found, err := s.idempotency.Find(ctx, operation, key)
	if err != nil || !found {
		return CashRecountResult{}, found, err
	}
	if record.TargetID != targetID || record.Fingerprint != fingerprint {
		return CashRecountResult{}, true, ErrIdempotencyKeyReused
	}
	result, ok := record.Result.(CashRecountResult)
	if !ok {
		return CashRecountResult{}, true, ErrIdempotencyResultMissing
	}
	return result, true, nil
}

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
	cash         CashRepository
	idempotency  IdempotencyStore
	outbox       OutboxRecorder
	journal      OperationJournalRecorder
	transactions TransactionRunner
	now          func() time.Time
	newID        func(prefix string) string
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

func WithCashOutboxRecorder(outbox OutboxRecorder) CashOption {
	return func(service *CashService) {
		service.outbox = outbox
	}
}

func WithCashJournal(journal OperationJournalRecorder) CashOption {
	return func(service *CashService) {
		service.journal = journal
	}
}

func WithCashTransactionRunner(runner TransactionRunner) CashOption {
	return func(service *CashService) {
		service.transactions = runner
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

type CreateBankCollectionCommand struct {
	IdempotencyKey  string
	StoreID         string
	SafeID          string
	BankContainerID string
	AmountMinor     int64
	Currency        string
	Reason          string
	ActorID         string
	ApprovedByID    string
}

type CreateBusinessExpenseCommand struct {
	IdempotencyKey string
	StoreID        string
	SafeID         string
	PayeeID        string
	AmountMinor    int64
	Currency       string
	Reason         string
	ActorID        string
	ApprovedByID   string
}

type CreateCashRecountCommand struct {
	IdempotencyKey string
	StoreID        string
	BusinessDate   string
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

	var result CashMovementResult
	if err := RunTransaction(ctx, s.transactions, func(ctx context.Context) error {
		if err := s.cash.SaveCashMovement(ctx, movement); err != nil {
			return err
		}

		if err := s.recordCashJournal(ctx, movement.StoreID, "cash.movement.created", movement.ActorID, movement.ID,
			fmt.Sprintf("%s %d from %s to %s", movement.Type, movement.AmountMinor, movement.FromContainerID, movement.ToContainerID)); err != nil {
			return err
		}

		result = CashMovementResult{Movement: movement}
		if err := s.idempotency.Save(ctx, IdempotencyRecord{
			Operation:   operation,
			Key:         command.IdempotencyKey,
			TargetID:    command.StoreID,
			Fingerprint: fingerprint,
			Result:      result,
			CreatedAt:   s.now(),
		}); err != nil {
			return err
		}
		return recordOutbox(ctx, s.outbox, func(ctx context.Context, recorder OutboxRecorder) error {
			return recorder.RecordCashMovementPosted(ctx, movement)
		})
	}); err != nil {
		return CashMovementResult{}, err
	}

	return result, nil
}

func (s *CashService) CreateBankCollection(ctx context.Context, command CreateBankCollectionCommand) (CashMovementResult, error) {
	if command.IdempotencyKey == "" {
		return CashMovementResult{}, ErrIdempotencyKeyRequired
	}
	if command.StoreID == "" || command.SafeID == "" || command.BankContainerID == "" ||
		command.AmountMinor <= 0 || command.ActorID == "" || command.ApprovedByID == "" {
		return CashMovementResult{}, ErrInvalidCashMovementCommand
	}
	if command.ActorID == command.ApprovedByID {
		return CashMovementResult{}, ErrSeparationOfDutiesViolation
	}

	reason := command.Reason
	if reason == "" {
		reason = "Bank collection from " + command.SafeID
	}

	return s.CreateCashMovement(ctx, CreateCashMovementCommand{
		IdempotencyKey:    command.IdempotencyKey,
		StoreID:           command.StoreID,
		Type:              domain.CashMovementTypeSafeToBank,
		FromContainerID:   command.SafeID,
		FromContainerType: domain.CashContainerTypeSafe,
		ToContainerID:     command.BankContainerID,
		ToContainerType:   domain.CashContainerTypeBank,
		AmountMinor:       command.AmountMinor,
		Currency:          command.Currency,
		Reason:            reason,
		ActorID:           command.ActorID,
		ApprovedByID:      command.ApprovedByID,
	})
}

func (s *CashService) CreateBusinessExpense(ctx context.Context, command CreateBusinessExpenseCommand) (CashMovementResult, error) {
	if command.IdempotencyKey == "" {
		return CashMovementResult{}, ErrIdempotencyKeyRequired
	}
	if command.StoreID == "" || command.SafeID == "" || command.PayeeID == "" ||
		command.AmountMinor <= 0 || command.Reason == "" || command.ActorID == "" || command.ApprovedByID == "" {
		return CashMovementResult{}, ErrInvalidCashMovementCommand
	}
	if command.ActorID == command.ApprovedByID {
		return CashMovementResult{}, ErrSeparationOfDutiesViolation
	}

	return s.CreateCashMovement(ctx, CreateCashMovementCommand{
		IdempotencyKey:    command.IdempotencyKey,
		StoreID:           command.StoreID,
		Type:              domain.CashMovementTypeExpense,
		FromContainerID:   command.SafeID,
		FromContainerType: domain.CashContainerTypeSafe,
		ToContainerID:     command.PayeeID,
		ToContainerType:   domain.CashContainerTypeExpense,
		AmountMinor:       command.AmountMinor,
		Currency:          command.Currency,
		Reason:            command.Reason,
		ActorID:           command.ActorID,
		ApprovedByID:      command.ApprovedByID,
	})
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
		BusinessDate:  command.BusinessDate,
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

	var result CashRecountResult
	if err := RunTransaction(ctx, s.transactions, func(ctx context.Context) error {
		if err := s.cash.SaveCashRecount(ctx, recount); err != nil {
			return err
		}

		if err := s.recordCashJournal(ctx, recount.StoreID, "cash.recount.created", recount.ActorID, recount.ID,
			fmt.Sprintf("recount %s expected=%d counted=%d", recount.ContainerID, recount.ExpectedMinor, recount.CountedMinor)); err != nil {
			return err
		}

		result = CashRecountResult{Recount: recount}
		return s.idempotency.Save(ctx, IdempotencyRecord{
			Operation:   operation,
			Key:         command.IdempotencyKey,
			TargetID:    command.StoreID,
			Fingerprint: fingerprint,
			Result:      result,
			CreatedAt:   s.now(),
		})
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

	var result CashRecountResult
	if err := RunTransaction(ctx, s.transactions, func(ctx context.Context) error {
		if err := s.cash.SaveCashRecount(ctx, recount); err != nil {
			return err
		}

		if err := s.recordCashJournal(ctx, recount.StoreID, "cash.recount.resolved", command.ActorID, recount.ID, command.ResolutionNote); err != nil {
			return err
		}

		result = CashRecountResult{Recount: recount}
		return s.idempotency.Save(ctx, IdempotencyRecord{
			Operation:   operation,
			Key:         command.IdempotencyKey,
			TargetID:    command.RecountID,
			Fingerprint: fingerprint,
			Result:      result,
			CreatedAt:   s.now(),
		})
	}); err != nil {
		return CashRecountResult{}, err
	}

	return result, nil
}

func (s *CashService) ListCashMovements(ctx context.Context, storeID string, params PageParams) (PageResult[domain.CashMovement], error) {
	if storeID == "" {
		return PageResult[domain.CashMovement]{}, ErrInvalidCashMovementCommand
	}
	movements, err := s.cash.ListCashMovements(ctx, storeID)
	if err != nil {
		return PageResult[domain.CashMovement]{}, err
	}
	return PaginateSlice(movements, params), nil
}

func (s *CashService) ListCashRecounts(ctx context.Context, storeID string, params PageParams) (PageResult[domain.CashRecount], error) {
	if storeID == "" {
		return PageResult[domain.CashRecount]{}, ErrInvalidCashRecountCommand
	}
	recounts, err := s.cash.ListCashRecounts(ctx, storeID)
	if err != nil {
		return PageResult[domain.CashRecount]{}, err
	}
	return PaginateSlice(recounts, params), nil
}

func (s *CashService) recordCashJournal(ctx context.Context, storeID string, operationType string, actorID string, referenceID string, summary string) error {
	if s.journal == nil {
		return nil
	}
	return s.journal.RecordOperation(ctx, RecordOperationCommand{
		StoreID:       storeID,
		OperationType: operationType,
		ActorID:       actorID,
		ReferenceID:   referenceID,
		Summary:       summary,
	})
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

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
	ErrChangeFundRequestNotFound         = errors.New("change fund request not found")
	ErrInvalidChangeFundRequestCommand   = errors.New("invalid change fund request command")
	ErrChangeFundRequestAlreadyFulfilled = errors.New("change fund request is already fulfilled")
)

type ChangeFundRequestRepository interface {
	SaveChangeFundRequest(ctx context.Context, request domain.ChangeFundRequest) error
	FindChangeFundRequest(ctx context.Context, requestID string) (domain.ChangeFundRequest, error)
	ListChangeFundRequestsByStore(ctx context.Context, storeID string) ([]domain.ChangeFundRequest, error)
}

type ChangeFundRequestShiftLookup interface {
	FindShift(ctx context.Context, shiftID string) (domain.Shift, error)
}

type ChangeFundRequestCashLedger interface {
	SaveCashMovement(ctx context.Context, movement domain.CashMovement) error
}

type ChangeFundRequestService struct {
	requests     ChangeFundRequestRepository
	shifts       ChangeFundRequestShiftLookup
	cash         ChangeFundRequestCashLedger
	idempotency  IdempotencyStore
	roles        ActorRoleLookup
	journal      OperationJournalRecorder
	outbox       OutboxRecorder
	transactions TransactionRunner
	now          func() time.Time
	newID        func(prefix string) string
}

type ChangeFundRequestOption func(*ChangeFundRequestService)

func NewChangeFundRequestService(requests ChangeFundRequestRepository, shifts ChangeFundRequestShiftLookup, cash ChangeFundRequestCashLedger, idempotency IdempotencyStore, roles ActorRoleLookup, options ...ChangeFundRequestOption) *ChangeFundRequestService {
	service := &ChangeFundRequestService{
		requests:    requests,
		shifts:      shifts,
		cash:        cash,
		idempotency: idempotency,
		roles:       roles,
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

func WithChangeFundRequestJournal(journal OperationJournalRecorder) ChangeFundRequestOption {
	return func(service *ChangeFundRequestService) {
		service.journal = journal
	}
}

func WithChangeFundRequestOutbox(outbox OutboxRecorder) ChangeFundRequestOption {
	return func(service *ChangeFundRequestService) {
		service.outbox = outbox
	}
}

func WithChangeFundRequestTransactionRunner(runner TransactionRunner) ChangeFundRequestOption {
	return func(service *ChangeFundRequestService) {
		service.transactions = runner
	}
}

type CreateChangeFundRequestCommand struct {
	IdempotencyKey string
	ShiftID        string
	ActorID        string
	AmountMinor    int64
	Currency       string
	Reason         string
	Breakdown      *domain.DenominationBreakdown
}

type ChangeFundRequestResult struct {
	Request domain.ChangeFundRequest
}

func (s *ChangeFundRequestService) CreateChangeFundRequest(ctx context.Context, command CreateChangeFundRequestCommand) (ChangeFundRequestResult, error) {
	if command.IdempotencyKey == "" {
		return ChangeFundRequestResult{}, ErrIdempotencyKeyRequired
	}
	if command.ShiftID == "" || command.ActorID == "" || command.AmountMinor <= 0 {
		return ChangeFundRequestResult{}, ErrInvalidChangeFundRequestCommand
	}

	const operation = "change_fund_requests.create"
	fingerprint := fmt.Sprintf("%s|%d|%s|%s|%s",
		command.ShiftID, command.AmountMinor, command.Currency, command.ActorID,
		denominationBreakdownFingerprint(command.Breakdown),
	)
	if result, found, err := s.findChangeFundRequestIdempotency(ctx, operation, command.IdempotencyKey, command.ShiftID, fingerprint); err != nil || found {
		return result, err
	}

	shift, err := s.shifts.FindShift(ctx, command.ShiftID)
	if err != nil {
		return ChangeFundRequestResult{}, err
	}
	if shift.Status != domain.ShiftStatusOpen {
		return ChangeFundRequestResult{}, ErrShiftNotOpen
	}

	request, err := domain.NewChangeFundRequest(domain.CreateChangeFundRequestInput{
		ID:          s.newID("cfr"),
		StoreID:     shift.StoreID,
		ShiftID:     shift.ID,
		ActorID:     command.ActorID,
		AmountMinor: command.AmountMinor,
		Currency:    command.Currency,
		Reason:      command.Reason,
		Breakdown:   command.Breakdown,
		Now:         s.now(),
	})
	if err != nil {
		return ChangeFundRequestResult{}, err
	}

	var result ChangeFundRequestResult
	if err := RunTransaction(ctx, s.transactions, func(ctx context.Context) error {
		if err := s.requests.SaveChangeFundRequest(ctx, request); err != nil {
			return err
		}
		if s.journal != nil {
			if err := s.journal.RecordOperation(ctx, RecordOperationCommand{
				StoreID:       request.StoreID,
				OperationType: "change_fund_request.requested",
				ActorID:       request.ActorID,
				ReferenceID:   request.ID,
				Summary:       fmt.Sprintf("change fund requested amount=%d shift=%s", request.AmountMinor, request.ShiftID),
			}); err != nil {
				return err
			}
		}

		result = ChangeFundRequestResult{Request: request}
		return s.idempotency.Save(ctx, IdempotencyRecord{
			Operation:   operation,
			Key:         command.IdempotencyKey,
			TargetID:    command.ShiftID,
			Fingerprint: fingerprint,
			Result:      result,
			CreatedAt:   s.now(),
		})
	}); err != nil {
		return ChangeFundRequestResult{}, err
	}
	return result, nil
}

type FulfillChangeFundRequestCommand struct {
	IdempotencyKey string
	RequestID      string
	ActorID        string // the fulfilling senior cashier/admin
	SafeID         string
}

func (s *ChangeFundRequestService) FulfillChangeFundRequest(ctx context.Context, command FulfillChangeFundRequestCommand) (ChangeFundRequestResult, error) {
	if command.IdempotencyKey == "" {
		return ChangeFundRequestResult{}, ErrIdempotencyKeyRequired
	}
	if command.RequestID == "" || command.ActorID == "" || command.SafeID == "" {
		return ChangeFundRequestResult{}, ErrInvalidChangeFundRequestCommand
	}
	if err := CheckActorPermission(s.roles, ctx, command.ActorID, PermissionChangeFundRequestFulfill); err != nil {
		return ChangeFundRequestResult{}, err
	}

	const operation = "change_fund_requests.fulfill"
	fingerprint := fmt.Sprintf("%s|%s|%s", command.RequestID, command.ActorID, command.SafeID)
	if result, found, err := s.findChangeFundRequestIdempotency(ctx, operation, command.IdempotencyKey, command.RequestID, fingerprint); err != nil || found {
		return result, err
	}

	request, err := s.requests.FindChangeFundRequest(ctx, command.RequestID)
	if err != nil {
		return ChangeFundRequestResult{}, err
	}
	if command.ActorID == request.ActorID {
		return ChangeFundRequestResult{}, ErrSeparationOfDutiesViolation
	}

	shift, err := s.shifts.FindShift(ctx, request.ShiftID)
	if err != nil {
		return ChangeFundRequestResult{}, err
	}
	if shift.Status != domain.ShiftStatusOpen {
		return ChangeFundRequestResult{}, ErrShiftNotOpen
	}

	movementID := s.newID("cash")
	if err := request.Fulfill(command.ActorID, command.SafeID, movementID, s.now()); err != nil {
		if errors.Is(err, domain.ErrChangeFundRequestAlreadyFulfilled) {
			return ChangeFundRequestResult{}, ErrChangeFundRequestAlreadyFulfilled
		}
		return ChangeFundRequestResult{}, ErrInvalidChangeFundRequestCommand
	}

	var result ChangeFundRequestResult
	if err := RunTransaction(ctx, s.transactions, func(ctx context.Context) error {
		movement, err := domain.CreateCashMovement(domain.CreateCashMovementInput{
			ID:                movementID,
			StoreID:           shift.StoreID,
			Type:              domain.CashMovementTypeChangeFund,
			FromContainerID:   command.SafeID,
			FromContainerType: domain.CashContainerTypeSafe,
			ToContainerID:     shift.DrawerID,
			ToContainerType:   domain.CashContainerTypeDrawer,
			AmountMinor:       request.AmountMinor,
			Currency:          request.Currency,
			Reason:            "Change fund request " + request.ID,
			Breakdown:         request.Breakdown,
			ActorID:           request.ActorID,
			ApprovedByID:      command.ActorID,
			Now:               s.now(),
		})
		if err != nil {
			return err
		}
		if err := s.cash.SaveCashMovement(ctx, movement); err != nil {
			return err
		}
		if err := s.requests.SaveChangeFundRequest(ctx, request); err != nil {
			return err
		}
		if s.journal != nil {
			if err := s.journal.RecordOperation(ctx, RecordOperationCommand{
				StoreID:       request.StoreID,
				OperationType: "change_fund_request.fulfilled",
				ActorID:       command.ActorID,
				ReferenceID:   request.ID,
				Summary:       fmt.Sprintf("change fund request %s fulfilled amount=%d safe=%s", request.ID, request.AmountMinor, command.SafeID),
			}); err != nil {
				return err
			}
		}

		result = ChangeFundRequestResult{Request: request}
		if err := s.idempotency.Save(ctx, IdempotencyRecord{
			Operation:   operation,
			Key:         command.IdempotencyKey,
			TargetID:    command.RequestID,
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
		return ChangeFundRequestResult{}, err
	}
	return result, nil
}

func (s *ChangeFundRequestService) GetChangeFundRequest(ctx context.Context, requestID string) (ChangeFundRequestResult, error) {
	if requestID == "" {
		return ChangeFundRequestResult{}, ErrInvalidChangeFundRequestCommand
	}
	request, err := s.requests.FindChangeFundRequest(ctx, requestID)
	if err != nil {
		return ChangeFundRequestResult{}, err
	}
	return ChangeFundRequestResult{Request: request}, nil
}

func (s *ChangeFundRequestService) ListChangeFundRequestsByStore(ctx context.Context, storeID string, params PageParams) (PageResult[domain.ChangeFundRequest], error) {
	if storeID == "" {
		return PageResult[domain.ChangeFundRequest]{}, ErrInvalidChangeFundRequestCommand
	}
	requests, err := s.requests.ListChangeFundRequestsByStore(ctx, storeID)
	if err != nil {
		return PageResult[domain.ChangeFundRequest]{}, err
	}
	sortChangeFundRequestsNewestFirst(requests)
	return PaginateSlice(requests, params), nil
}

func sortChangeFundRequestsNewestFirst(requests []domain.ChangeFundRequest) {
	sort.Slice(requests, func(i, j int) bool {
		if requests[i].CreatedAt.Equal(requests[j].CreatedAt) {
			return requests[i].ID > requests[j].ID
		}
		return requests[i].CreatedAt.After(requests[j].CreatedAt)
	})
}

func (s *ChangeFundRequestService) findChangeFundRequestIdempotency(ctx context.Context, operation string, key string, targetID string, fingerprint string) (ChangeFundRequestResult, bool, error) {
	record, found, err := s.idempotency.Find(ctx, operation, key)
	if err != nil || !found {
		return ChangeFundRequestResult{}, found, err
	}
	if record.TargetID != targetID || record.Fingerprint != fingerprint {
		return ChangeFundRequestResult{}, true, ErrIdempotencyKeyReused
	}
	result, ok := record.Result.(ChangeFundRequestResult)
	if !ok {
		return ChangeFundRequestResult{}, true, ErrIdempotencyResultMissing
	}
	return result, true, nil
}

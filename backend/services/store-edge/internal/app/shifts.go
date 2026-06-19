package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/domain"
)

var (
	ErrShiftNotFound               = errors.New("shift not found")
	ErrInvalidShiftCommand         = errors.New("invalid shift command")
	ErrShiftAlreadyOpenForTerminal = errors.New("shift already open for terminal")
	ErrShiftAlreadyOpenForCashier  = errors.New("shift already open for cashier")
	ErrShiftAlreadyClosed          = errors.New("shift already closed")
	ErrShiftCashCollectionRequired = errors.New("shift cash collection details are required")
	ErrShiftOpeningSafeRequired    = errors.New("shift opening safe is required when opening cash is positive")
	ErrShiftCloseBlocked           = errors.New("shift close blocked")
)

type ShiftRepository interface {
	SaveShift(ctx context.Context, shift domain.Shift) error
	FindShift(ctx context.Context, shiftID string) (domain.Shift, error)
	FindOpenShiftByTerminal(ctx context.Context, terminalID string) (domain.Shift, error)
	FindOpenShiftByCashier(ctx context.Context, cashierID string) (domain.Shift, error)
	ListOpenShiftsByStore(ctx context.Context, storeID string) ([]domain.Shift, error)
	ListShiftsByOperationalDay(ctx context.Context, operationalDayID string) ([]domain.Shift, error)
}

type ShiftReceiptRepository interface {
	ListUnresolvedReceiptsByShift(ctx context.Context, shiftID string) ([]domain.Receipt, error)
}

type ShiftService struct {
	shifts       ShiftRepository
	cash         CashRepository
	receipts     ShiftReceiptRepository
	days         OperationalDayRepository
	idempotency  IdempotencyStore
	transactions TransactionRunner
	now          func() time.Time
	newID        func(prefix string) string
}

type ShiftOption func(*ShiftService)

func NewShiftService(shifts ShiftRepository, idempotency IdempotencyStore, options ...ShiftOption) *ShiftService {
	service := &ShiftService{
		shifts:      shifts,
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

func WithShiftClock(now func() time.Time) ShiftOption {
	return func(service *ShiftService) {
		service.now = now
	}
}

func WithShiftIDGenerator(newID func(prefix string) string) ShiftOption {
	return func(service *ShiftService) {
		service.newID = newID
	}
}

func WithShiftCashLedger(cash CashRepository) ShiftOption {
	return func(service *ShiftService) {
		service.cash = cash
	}
}

func WithShiftReceiptRepository(receipts ShiftReceiptRepository) ShiftOption {
	return func(service *ShiftService) {
		service.receipts = receipts
	}
}

func WithShiftOperationalDayRepository(days OperationalDayRepository) ShiftOption {
	return func(service *ShiftService) {
		service.days = days
	}
}

func WithShiftTransactionRunner(runner TransactionRunner) ShiftOption {
	return func(service *ShiftService) {
		service.transactions = runner
	}
}

type OpenShiftCommand struct {
	IdempotencyKey   string
	StoreID          string
	TerminalID       string
	CashierID        string
	DrawerID         string
	SourceSafeID     string
	OpeningCashMinor int64
}

type CloseShiftCommand struct {
	IdempotencyKey   string
	ShiftID          string
	ClosingCashMinor int64
	SafeID           string
	ActorID          string
	ApprovedByID     string
}

type ShiftResult struct {
	Shift domain.Shift
}

func (s *ShiftService) OpenShift(ctx context.Context, command OpenShiftCommand) (ShiftResult, error) {
	if command.IdempotencyKey == "" {
		return ShiftResult{}, ErrIdempotencyKeyRequired
	}
	if command.StoreID == "" || command.TerminalID == "" || command.CashierID == "" ||
		command.DrawerID == "" || command.OpeningCashMinor < 0 {
		return ShiftResult{}, ErrInvalidShiftCommand
	}
	if command.OpeningCashMinor > 0 && command.SourceSafeID == "" {
		return ShiftResult{}, ErrShiftOpeningSafeRequired
	}

	const operation = "shifts.open_shift"
	fingerprint := fmt.Sprintf("%s|%s|%s|%s|%s|%d", command.StoreID, command.TerminalID, command.CashierID, command.DrawerID, command.SourceSafeID, command.OpeningCashMinor)
	if result, found, err := s.findShiftIdempotency(ctx, operation, command.IdempotencyKey, "", fingerprint); err != nil || found {
		return result, err
	}

	if _, err := s.shifts.FindOpenShiftByTerminal(ctx, command.TerminalID); err == nil {
		return ShiftResult{}, ErrShiftAlreadyOpenForTerminal
	} else if !errors.Is(err, ErrShiftNotFound) {
		return ShiftResult{}, err
	}
	if _, err := s.shifts.FindOpenShiftByCashier(ctx, command.CashierID); err == nil {
		return ShiftResult{}, ErrShiftAlreadyOpenForCashier
	} else if !errors.Is(err, ErrShiftNotFound) {
		return ShiftResult{}, err
	}

	var day domain.OperationalDay
	if s.days != nil {
		var err error
		day, err = s.days.FindOpenOperationalDayByStore(ctx, command.StoreID)
		if err != nil {
			if errors.Is(err, ErrOperationalDayNotFound) {
				return ShiftResult{}, ErrOpenOperationalDayRequired
			}
			return ShiftResult{}, err
		}
	}

	shift, err := domain.OpenShift(domain.OpenShiftInput{
		ID:               s.newID("shf"),
		StoreID:          command.StoreID,
		OperationalDayID: day.ID,
		BusinessDate:     day.BusinessDate,
		TerminalID:       command.TerminalID,
		CashierID:        command.CashierID,
		DrawerID:         command.DrawerID,
		OpeningCashMinor: command.OpeningCashMinor,
		Now:              s.now(),
	})
	if err != nil {
		return ShiftResult{}, err
	}

	var result ShiftResult
	if err := RunTransaction(ctx, s.transactions, func(ctx context.Context) error {
		if command.OpeningCashMinor > 0 && s.cash != nil {
			movement, err := domain.CreateCashMovement(domain.CreateCashMovementInput{
				ID:                s.newID("cash"),
				StoreID:           command.StoreID,
				Type:              domain.CashMovementTypeChangeFund,
				FromContainerID:   command.SourceSafeID,
				FromContainerType: domain.CashContainerTypeSafe,
				ToContainerID:     command.DrawerID,
				ToContainerType:   domain.CashContainerTypeDrawer,
				AmountMinor:       command.OpeningCashMinor,
				Currency:          "RUB",
				Reason:            "Opening change fund for shift " + shift.ID,
				ActorID:           command.CashierID,
				Now:               s.now(),
			})
			if err != nil {
				return err
			}
			if err := s.cash.SaveCashMovement(ctx, movement); err != nil {
				return err
			}
		}

		if err := s.shifts.SaveShift(ctx, shift); err != nil {
			return err
		}

		result = ShiftResult{Shift: shift}
		return s.idempotency.Save(ctx, IdempotencyRecord{
			Operation:   operation,
			Key:         command.IdempotencyKey,
			TargetID:    shift.ID,
			Fingerprint: fingerprint,
			Result:      result,
			CreatedAt:   s.now(),
		})
	}); err != nil {
		return ShiftResult{}, err
	}

	return result, nil
}

func (s *ShiftService) CloseShift(ctx context.Context, command CloseShiftCommand) (ShiftResult, error) {
	if command.IdempotencyKey == "" {
		return ShiftResult{}, ErrIdempotencyKeyRequired
	}
	if command.ShiftID == "" || command.ClosingCashMinor < 0 {
		return ShiftResult{}, ErrInvalidShiftCommand
	}

	const operation = "shifts.close_shift"
	fingerprint := fmt.Sprintf("%s|%d|%s|%s|%s", command.ShiftID, command.ClosingCashMinor, command.SafeID, command.ActorID, command.ApprovedByID)
	if result, found, err := s.findShiftIdempotency(ctx, operation, command.IdempotencyKey, command.ShiftID, fingerprint); err != nil || found {
		return result, err
	}

	shift, err := s.shifts.FindShift(ctx, command.ShiftID)
	if err != nil {
		return ShiftResult{}, err
	}
	if shift.Status != domain.ShiftStatusOpen {
		return ShiftResult{}, ErrShiftAlreadyClosed
	}
	if s.receipts != nil {
		unresolvedReceipts, err := s.receipts.ListUnresolvedReceiptsByShift(ctx, shift.ID)
		if err != nil {
			return ShiftResult{}, err
		}
		if len(unresolvedReceipts) > 0 {
			return ShiftResult{}, ErrShiftCloseBlocked
		}
	}
	if err := shift.Close(command.ClosingCashMinor, s.now()); err != nil {
		if errors.Is(err, domain.ErrShiftNotOpen) {
			return ShiftResult{}, ErrShiftAlreadyClosed
		}
		return ShiftResult{}, err
	}

	var result ShiftResult
	if err := RunTransaction(ctx, s.transactions, func(ctx context.Context) error {
		if command.ClosingCashMinor > 0 && s.cash != nil {
			if command.SafeID == "" || command.ActorID == "" || command.ApprovedByID == "" {
				return ErrShiftCashCollectionRequired
			}
			if command.ActorID == command.ApprovedByID {
				return ErrSeparationOfDutiesViolation
			}
			movement, err := domain.CreateCashMovement(domain.CreateCashMovementInput{
				ID:                s.newID("cash"),
				StoreID:           shift.StoreID,
				Type:              domain.CashMovementTypeDrawerToSafe,
				FromContainerID:   shift.DrawerID,
				FromContainerType: domain.CashContainerTypeDrawer,
				ToContainerID:     command.SafeID,
				ToContainerType:   domain.CashContainerTypeSafe,
				AmountMinor:       command.ClosingCashMinor,
				Currency:          "RUB",
				Reason:            "Final cashier collection for shift " + shift.ID,
				ActorID:           command.ActorID,
				ApprovedByID:      command.ApprovedByID,
				Now:               s.now(),
			})
			if err != nil {
				return err
			}
			if err := s.cash.SaveCashMovement(ctx, movement); err != nil {
				return err
			}
		}

		if err := s.shifts.SaveShift(ctx, shift); err != nil {
			return err
		}

		result = ShiftResult{Shift: shift}
		return s.idempotency.Save(ctx, IdempotencyRecord{
			Operation:   operation,
			Key:         command.IdempotencyKey,
			TargetID:    command.ShiftID,
			Fingerprint: fingerprint,
			Result:      result,
			CreatedAt:   s.now(),
		})
	}); err != nil {
		return ShiftResult{}, err
	}

	return result, nil
}

func (s *ShiftService) GetShift(ctx context.Context, shiftID string) (ShiftResult, error) {
	if shiftID == "" {
		return ShiftResult{}, ErrInvalidShiftCommand
	}
	shift, err := s.shifts.FindShift(ctx, shiftID)
	if err != nil {
		return ShiftResult{}, err
	}
	return ShiftResult{Shift: shift}, nil
}

func (s *ShiftService) ListOpenShiftsByStore(ctx context.Context, storeID string) ([]domain.Shift, error) {
	if storeID == "" {
		return nil, ErrInvalidShiftCommand
	}
	return s.shifts.ListOpenShiftsByStore(ctx, storeID)
}

func (s *ShiftService) ListShiftsByOperationalDay(ctx context.Context, operationalDayID string, params PageParams) (PageResult[domain.Shift], error) {
	if operationalDayID == "" {
		return PageResult[domain.Shift]{}, ErrInvalidShiftCommand
	}
	shifts, err := s.shifts.ListShiftsByOperationalDay(ctx, operationalDayID)
	if err != nil {
		return PageResult[domain.Shift]{}, err
	}
	return PaginateSlice(shifts, params), nil
}

func (s *ShiftService) findShiftIdempotency(ctx context.Context, operation string, key string, targetID string, fingerprint string) (ShiftResult, bool, error) {
	record, found, err := s.idempotency.Find(ctx, operation, key)
	if err != nil || !found {
		return ShiftResult{}, found, err
	}
	if targetID != "" && record.TargetID != targetID {
		return ShiftResult{}, true, ErrIdempotencyKeyReused
	}
	if record.Fingerprint != fingerprint {
		return ShiftResult{}, true, ErrIdempotencyKeyReused
	}
	result, ok := record.Result.(ShiftResult)
	if !ok {
		return ShiftResult{}, true, ErrIdempotencyResultMissing
	}
	return result, true, nil
}

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
	ErrReturnSettlementPaymentMismatch         = errors.New("return settlement payment mismatch")
	ErrReturnSettlementCumulativeTotalExceeded = errors.New("return settlement cumulative total exceeded")
)

type PaymentRefunder interface {
	RefundPayment(ctx context.Context, command RefundPaymentCommand) (PaymentResult, error)
}

type ReturnSettlementCashLedger interface {
	SaveCashMovement(ctx context.Context, movement domain.CashMovement) error
}

type ReturnSettlementShiftLookup interface {
	FindOpenShiftByCashier(ctx context.Context, cashierID string) (domain.Shift, error)
}

type ReturnSettlementOutboxRecorder interface {
	RecordReturnSettled(ctx context.Context, ret domain.Return, paymentIDs []string, storeID string, actorID string, cashMovementID string) error
}

type ReturnSettlementService struct {
	returns      ReturnRepository
	receipts     ReceiptRepository
	payments     PaymentRepository
	refunder     PaymentRefunder
	cash         ReturnSettlementCashLedger
	shifts       ReturnSettlementShiftLookup
	idempotency  IdempotencyStore
	outbox       ReturnSettlementOutboxRecorder
	journal      OperationJournalRecorder
	transactions TransactionRunner
	now          func() time.Time
	newID        func(prefix string) string
}

type ReturnSettlementOption func(*ReturnSettlementService)

func NewReturnSettlementService(
	returns ReturnRepository,
	receipts ReceiptRepository,
	payments PaymentRepository,
	refunder PaymentRefunder,
	idempotency IdempotencyStore,
	options ...ReturnSettlementOption,
) *ReturnSettlementService {
	service := &ReturnSettlementService{
		returns:     returns,
		receipts:    receipts,
		payments:    payments,
		refunder:    refunder,
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

func WithReturnSettlementOutboxRecorder(outbox ReturnSettlementOutboxRecorder) ReturnSettlementOption {
	return func(service *ReturnSettlementService) {
		service.outbox = outbox
	}
}

func WithReturnSettlementJournal(journal OperationJournalRecorder) ReturnSettlementOption {
	return func(service *ReturnSettlementService) {
		service.journal = journal
	}
}

func WithReturnSettlementCashLedger(cash ReturnSettlementCashLedger) ReturnSettlementOption {
	return func(service *ReturnSettlementService) {
		service.cash = cash
	}
}

func WithReturnSettlementShiftLookup(shifts ReturnSettlementShiftLookup) ReturnSettlementOption {
	return func(service *ReturnSettlementService) {
		service.shifts = shifts
	}
}

func WithReturnSettlementTransactionRunner(runner TransactionRunner) ReturnSettlementOption {
	return func(service *ReturnSettlementService) {
		service.transactions = runner
	}
}

type SettleReturnCommand struct {
	IdempotencyKey string
	ReturnID       string
	ActorID        string
	Reason         string
	DrawerID       string
}

type SettleReturnResult struct {
	Return   domain.Return
	Payments []domain.Payment
}

func (s *ReturnSettlementService) SettleReturn(ctx context.Context, command SettleReturnCommand) (SettleReturnResult, error) {
	if command.IdempotencyKey == "" {
		return SettleReturnResult{}, ErrIdempotencyKeyRequired
	}
	if command.ReturnID == "" {
		return SettleReturnResult{}, ErrInvalidReturnCommand
	}

	ret, err := s.returns.FindReturn(ctx, command.ReturnID)
	if err != nil {
		return SettleReturnResult{}, err
	}
	if ret.Status == domain.ReturnStatusSettled {
		return SettleReturnResult{}, ErrReturnAlreadySettled
	}
	if ret.Status != domain.ReturnStatusCompleted {
		return SettleReturnResult{}, ErrReturnSettlementNotAllowed
	}
	if ret.TotalMinor <= 0 {
		return SettleReturnResult{}, ErrInvalidReturnCommand
	}

	switch ret.Kind {
	case domain.ReturnKindWithReceipt:
		return s.settleWithReceiptReturn(ctx, command, ret)
	case domain.ReturnKindNoReceipt:
		return s.settleNoReceiptReturn(ctx, command, ret)
	default:
		return SettleReturnResult{}, ErrReturnSettlementNotAllowed
	}
}

func (s *ReturnSettlementService) settleWithReceiptReturn(ctx context.Context, command SettleReturnCommand, ret domain.Return) (SettleReturnResult, error) {
	const operation = "returns.settle"
	fingerprint := fmt.Sprintf("%s|%s|%s", command.ReturnID, command.ActorID, command.Reason)
	if result, found, err := s.findSettleIdempotency(ctx, operation, command.IdempotencyKey, command.ReturnID, fingerprint); err != nil || found {
		return result, err
	}

	receipt, err := s.receipts.FindReceipt(ctx, ret.ReceiptID)
	if err != nil {
		return SettleReturnResult{}, err
	}
	if receipt.Status != domain.ReceiptStatusFiscalized {
		return SettleReturnResult{}, ErrReceiptNotReturnable
	}
	if ret.TotalMinor > receipt.TotalMinor() {
		return SettleReturnResult{}, ErrReturnSettlementPaymentMismatch
	}

	priorSettled, err := s.priorSettledTotal(ctx, ret.ReceiptID, ret.ID)
	if err != nil {
		return SettleReturnResult{}, err
	}
	if priorSettled+ret.TotalMinor > receipt.TotalMinor() {
		return SettleReturnResult{}, ErrReturnSettlementCumulativeTotalExceeded
	}

	receiptPayments, err := s.payments.FindPaymentsByReceipt(ctx, ret.ReceiptID)
	if err != nil {
		return SettleReturnResult{}, err
	}
	refundable, totalRefundable, err := refundablePayments(receiptPayments)
	if err != nil {
		return SettleReturnResult{}, err
	}
	if totalRefundable < ret.TotalMinor {
		return SettleReturnResult{}, ErrReturnSettlementPaymentMismatch
	}

	allocations, err := allocateRefundAmounts(refundable, ret.TotalMinor, totalRefundable)
	if err != nil {
		return SettleReturnResult{}, err
	}

	var result SettleReturnResult
	if err := RunTransaction(ctx, s.transactions, func(ctx context.Context) error {
		refundedPayments := make([]domain.Payment, 0, len(allocations))
		for paymentID, amountMinor := range allocations {
			if amountMinor == 0 {
				continue
			}
			refundResult, err := s.refunder.RefundPayment(ctx, RefundPaymentCommand{
				IdempotencyKey: fmt.Sprintf("%s:%s", command.IdempotencyKey, paymentID),
				ReceiptID:      ret.ReceiptID,
				PaymentID:      paymentID,
				AmountMinor:    amountMinor,
				ActorID:        command.ActorID,
				Reason:         command.Reason,
			})
			if err != nil {
				return err
			}
			refundedPayments = append(refundedPayments, refundResult.Payment)
		}

		paymentIDs := make([]string, 0, len(refundedPayments))
		for _, payment := range refundedPayments {
			paymentIDs = append(paymentIDs, payment.ID)
		}

		var settleErr error
		result, settleErr = s.finishSettledReturn(ctx, command, ret, operation, fingerprint, paymentIDs, "", refundedPayments)
		return settleErr
	}); err != nil {
		return SettleReturnResult{}, err
	}
	return result, nil
}

func (s *ReturnSettlementService) settleNoReceiptReturn(ctx context.Context, command SettleReturnCommand, ret domain.Return) (SettleReturnResult, error) {
	if ret.ApprovedByID == "" {
		return SettleReturnResult{}, ErrReturnSettlementNotAllowed
	}
	if s.cash == nil {
		return SettleReturnResult{}, ErrCashDrawerRequired
	}

	actorID := command.ActorID
	if actorID == "" {
		actorID = ret.ActorID
	}
	if actorID == ret.ApprovedByID {
		return SettleReturnResult{}, ErrSeparationOfDutiesViolation
	}

	drawerID, err := s.resolveDrawerID(ctx, command.DrawerID, actorID)
	if err != nil {
		return SettleReturnResult{}, err
	}

	const operation = "returns.settle"
	fingerprint := fmt.Sprintf("%s|%s|%s|%s", command.ReturnID, command.ActorID, command.Reason, drawerID)
	if result, found, err := s.findSettleIdempotency(ctx, operation, command.IdempotencyKey, command.ReturnID, fingerprint); err != nil || found {
		return result, err
	}

	now := s.now()
	movement, err := domain.CreateCashMovement(domain.CreateCashMovementInput{
		ID:                s.newID("cash"),
		StoreID:           ret.StoreID,
		Type:              domain.CashMovementTypeNoReceiptReturnPayout,
		FromContainerID:   drawerID,
		FromContainerType: domain.CashContainerTypeDrawer,
		ToContainerID:     "external-customer",
		ToContainerType:   domain.CashContainerTypeExternal,
		AmountMinor:       ret.TotalMinor,
		Currency:          "RUB",
		Reason:            "No-receipt return payout " + ret.ID,
		ActorID:           actorID,
		ApprovedByID:      ret.ApprovedByID,
		Now:               now,
	})
	if err != nil {
		return SettleReturnResult{}, err
	}

	var result SettleReturnResult
	if err := RunTransaction(ctx, s.transactions, func(ctx context.Context) error {
		if err := s.cash.SaveCashMovement(ctx, movement); err != nil {
			return err
		}

		var settleErr error
		result, settleErr = s.finishSettledReturn(ctx, command, ret, operation, fingerprint, nil, movement.ID, nil)
		return settleErr
	}); err != nil {
		return SettleReturnResult{}, err
	}
	return result, nil
}

func (s *ReturnSettlementService) resolveDrawerID(ctx context.Context, drawerID string, actorID string) (string, error) {
	if drawerID != "" {
		return drawerID, nil
	}
	if s.shifts == nil {
		return "", ErrCashDrawerRequired
	}
	shift, err := s.shifts.FindOpenShiftByCashier(ctx, actorID)
	if err != nil {
		if errors.Is(err, ErrShiftNotFound) {
			return "", ErrCashDrawerRequired
		}
		return "", err
	}
	if shift.DrawerID == "" {
		return "", ErrCashDrawerRequired
	}
	return shift.DrawerID, nil
}

func (s *ReturnSettlementService) finishSettledReturn(
	ctx context.Context,
	command SettleReturnCommand,
	ret domain.Return,
	operation string,
	fingerprint string,
	paymentIDs []string,
	cashMovementID string,
	refundedPayments []domain.Payment,
) (SettleReturnResult, error) {
	now := s.now()
	if err := ret.MarkSettled(now); err != nil {
		return SettleReturnResult{}, mapReturnSettlementDomainError(err)
	}
	if err := s.returns.SaveReturn(ctx, ret); err != nil {
		return SettleReturnResult{}, err
	}

	if s.outbox != nil {
		actorID := command.ActorID
		if actorID == "" {
			actorID = ret.ActorID
		}
		if err := s.outbox.RecordReturnSettled(ctx, ret, paymentIDs, ret.StoreID, actorID, cashMovementID); err != nil {
			return SettleReturnResult{}, err
		}
	}
	if err := s.recordJournal(ctx, ret, paymentIDs, cashMovementID); err != nil {
		return SettleReturnResult{}, err
	}

	result := SettleReturnResult{
		Return:   ret,
		Payments: refundedPayments,
	}
	if err := s.idempotency.Save(ctx, IdempotencyRecord{
		Operation:   operation,
		Key:         command.IdempotencyKey,
		TargetID:    command.ReturnID,
		Fingerprint: fingerprint,
		Result:      result,
		CreatedAt:   now,
	}); err != nil {
		return SettleReturnResult{}, err
	}
	return result, nil
}

func (s *ReturnSettlementService) priorSettledTotal(ctx context.Context, receiptID string, excludeReturnID string) (int64, error) {
	returns, err := s.returns.ListReturnsByReceipt(ctx, receiptID)
	if err != nil {
		return 0, err
	}
	var total int64
	for _, ret := range returns {
		if ret.ID == excludeReturnID {
			continue
		}
		if ret.Status == domain.ReturnStatusSettled {
			total += ret.TotalMinor
		}
	}
	return total, nil
}

func refundablePayments(payments []domain.Payment) ([]domain.Payment, int64, error) {
	if len(payments) == 0 {
		return nil, 0, ErrReturnSettlementPaymentMismatch
	}

	refundable := make([]domain.Payment, 0, len(payments))
	var total int64
	for _, payment := range payments {
		remaining := payment.RefundableAmountMinor()
		if remaining <= 0 {
			continue
		}
		refundable = append(refundable, payment)
		total += remaining
	}
	if len(refundable) == 0 {
		return nil, 0, ErrReturnSettlementPaymentMismatch
	}
	return refundable, total, nil
}

func allocateRefundAmounts(payments []domain.Payment, returnTotal int64, totalRefundable int64) (map[string]int64, error) {
	if returnTotal <= 0 || returnTotal > totalRefundable {
		return nil, ErrReturnSettlementPaymentMismatch
	}

	sorted := append([]domain.Payment(nil), payments...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].CapturedAt.Equal(sorted[j].CapturedAt) {
			return sorted[i].ID < sorted[j].ID
		}
		return sorted[i].CapturedAt.Before(sorted[j].CapturedAt)
	})

	allocations := make(map[string]int64, len(sorted))
	var allocated int64
	for i, payment := range sorted {
		var amount int64
		if i == len(sorted)-1 {
			amount = returnTotal - allocated
		} else {
			amount = returnTotal * payment.RefundableAmountMinor() / totalRefundable
		}
		if amount > payment.RefundableAmountMinor() {
			amount = payment.RefundableAmountMinor()
		}
		if amount > 0 {
			allocations[payment.ID] = amount
			allocated += amount
		}
	}
	if allocated != returnTotal {
		return nil, ErrReturnSettlementPaymentMismatch
	}
	return allocations, nil
}

func mapReturnSettlementDomainError(err error) error {
	switch {
	case errors.Is(err, domain.ErrReturnAlreadySettled):
		return ErrReturnAlreadySettled
	case errors.Is(err, domain.ErrReturnSettlementNotAllowed):
		return ErrReturnSettlementNotAllowed
	default:
		return err
	}
}

func (s *ReturnSettlementService) recordJournal(ctx context.Context, ret domain.Return, paymentIDs []string, cashMovementID string) error {
	if s.journal == nil {
		return nil
	}
	summary := fmt.Sprintf("settled return %s payments=%d total=%d", ret.ID, len(paymentIDs), ret.TotalMinor)
	if cashMovementID != "" {
		summary = fmt.Sprintf("settled no-receipt return %s cashMovement=%s total=%d", ret.ID, cashMovementID, ret.TotalMinor)
	}
	return s.journal.RecordOperation(ctx, RecordOperationCommand{
		StoreID:       ret.StoreID,
		OperationType: "return.settled",
		ActorID:       ret.ActorID,
		ReferenceID:   ret.ID,
		Summary:       summary,
	})
}

func (s *ReturnSettlementService) findSettleIdempotency(ctx context.Context, operation string, key string, targetID string, fingerprint string) (SettleReturnResult, bool, error) {
	record, found, err := s.idempotency.Find(ctx, operation, key)
	if err != nil || !found {
		return SettleReturnResult{}, found, err
	}
	if record.TargetID != targetID || record.Fingerprint != fingerprint {
		return SettleReturnResult{}, true, ErrIdempotencyKeyReused
	}
	result, ok := record.Result.(SettleReturnResult)
	if !ok {
		return SettleReturnResult{}, true, ErrIdempotencyResultMissing
	}
	return result, true, nil
}

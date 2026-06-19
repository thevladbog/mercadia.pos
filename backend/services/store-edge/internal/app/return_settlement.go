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
	ErrReturnSettlementPaymentMismatch           = errors.New("return settlement payment mismatch")
	ErrReturnSettlementCumulativeTotalExceeded   = errors.New("return settlement cumulative total exceeded")
)

type PaymentRefunder interface {
	RefundPayment(ctx context.Context, command RefundPaymentCommand) (PaymentResult, error)
}

type ReturnSettlementOutboxRecorder interface {
	RecordReturnSettled(ctx context.Context, ret domain.Return, paymentIDs []string, storeID string, actorID string) error
}

type ReturnSettlementService struct {
	returns     ReturnRepository
	receipts    ReceiptRepository
	payments    PaymentRepository
	refunder    PaymentRefunder
	idempotency IdempotencyStore
	outbox      ReturnSettlementOutboxRecorder
	journal     OperationJournalRecorder
	now         func() time.Time
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

type SettleReturnCommand struct {
	IdempotencyKey string
	ReturnID       string
	ActorID        string
	Reason         string
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

	const operation = "returns.settle"
	fingerprint := fmt.Sprintf("%s|%s|%s", command.ReturnID, command.ActorID, command.Reason)
	if result, found, err := s.findSettleIdempotency(ctx, operation, command.IdempotencyKey, command.ReturnID, fingerprint); err != nil || found {
		return result, err
	}

	ret, err := s.returns.FindReturn(ctx, command.ReturnID)
	if err != nil {
		return SettleReturnResult{}, err
	}
	if ret.Kind != domain.ReturnKindWithReceipt {
		return SettleReturnResult{}, ErrReturnSettlementNotAllowed
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
			return SettleReturnResult{}, err
		}
		refundedPayments = append(refundedPayments, refundResult.Payment)
	}

	now := s.now()
	if err := ret.MarkSettled(now); err != nil {
		return SettleReturnResult{}, mapReturnSettlementDomainError(err)
	}
	if err := s.returns.SaveReturn(ctx, ret); err != nil {
		return SettleReturnResult{}, err
	}

	paymentIDs := make([]string, 0, len(refundedPayments))
	for _, payment := range refundedPayments {
		paymentIDs = append(paymentIDs, payment.ID)
	}
	if s.outbox != nil {
		actorID := command.ActorID
		if actorID == "" {
			actorID = ret.ActorID
		}
		if err := s.outbox.RecordReturnSettled(ctx, ret, paymentIDs, ret.StoreID, actorID); err != nil {
			return SettleReturnResult{}, err
		}
	}
	s.recordJournal(ctx, ret, paymentIDs)

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

func (s *ReturnSettlementService) recordJournal(ctx context.Context, ret domain.Return, paymentIDs []string) {
	if s.journal == nil {
		return
	}
	_ = s.journal.RecordOperation(ctx, RecordOperationCommand{
		StoreID:       ret.StoreID,
		OperationType: "return.settled",
		ActorID:       ret.ActorID,
		ReferenceID:   ret.ID,
		Summary:       fmt.Sprintf("settled return %s payments=%d total=%d", ret.ID, len(paymentIDs), ret.TotalMinor),
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

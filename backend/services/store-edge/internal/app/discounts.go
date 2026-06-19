package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/domain"
)

var ErrInvalidDiscountCommand = errors.New("invalid discount command")

type DiscountService struct {
	receipts    ReceiptRepository
	idempotency IdempotencyStore
	roles       ActorRoleLookup
	journal     OperationJournalRecorder
	now         func() time.Time
}

type DiscountOption func(*DiscountService)

func NewDiscountService(receipts ReceiptRepository, idempotency IdempotencyStore, roles ActorRoleLookup, options ...DiscountOption) *DiscountService {
	service := &DiscountService{
		receipts:    receipts,
		idempotency: idempotency,
		roles:       roles,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func WithDiscountJournal(journal OperationJournalRecorder) DiscountOption {
	return func(service *DiscountService) {
		service.journal = journal
	}
}

type ApplyLineDiscountCommand struct {
	IdempotencyKey string
	ReceiptID      string
	LineID         string
	AmountMinor    int64
	Reason         string
	ActorID        string
}

func (s *DiscountService) ApplyLineDiscount(ctx context.Context, command ApplyLineDiscountCommand) (ReceiptResult, error) {
	if command.IdempotencyKey == "" {
		return ReceiptResult{}, ErrIdempotencyKeyRequired
	}
	if command.ReceiptID == "" || command.LineID == "" || command.Reason == "" ||
		command.ActorID == "" || command.AmountMinor <= 0 {
		return ReceiptResult{}, ErrInvalidDiscountCommand
	}
	if err := CheckActorPermission(s.roles, ctx, command.ActorID, PermissionDiscountApply); err != nil {
		return ReceiptResult{}, err
	}

	const operation = "discount.apply_line"
	fingerprint := fmt.Sprintf("%s|%s|%d|%s|%s", command.ReceiptID, command.LineID, command.AmountMinor, command.Reason, command.ActorID)
	if result, found, err := s.findDiscountIdempotency(ctx, operation, command.IdempotencyKey, command.ReceiptID, fingerprint); err != nil || found {
		return result, err
	}

	receipt, err := s.receipts.FindReceipt(ctx, command.ReceiptID)
	if err != nil {
		return ReceiptResult{}, err
	}
	if err := receipt.ApplyLineDiscount(command.LineID, domain.ApplyLineDiscountInput{
		AmountMinor: command.AmountMinor,
		Reason:      command.Reason,
		ActorID:     command.ActorID,
		Now:         s.now(),
	}); err != nil {
		if errors.Is(err, domain.ErrReceiptClosed) {
			return ReceiptResult{}, ErrInvalidDiscountCommand
		}
		return ReceiptResult{}, ErrInvalidDiscountCommand
	}

	if err := s.receipts.SaveReceipt(ctx, receipt); err != nil {
		return ReceiptResult{}, err
	}

	if s.journal != nil {
		_ = s.journal.RecordOperation(ctx, RecordOperationCommand{
			StoreID:       receipt.StoreID,
			OperationType: "discount.applied",
			ActorID:       command.ActorID,
			ReferenceID:   receipt.ID,
			Summary:       fmt.Sprintf("line %s discount %d reason=%s", command.LineID, command.AmountMinor, command.Reason),
		})
	}

	result := ReceiptResult{Receipt: receipt}
	if err := s.idempotency.Save(ctx, IdempotencyRecord{
		Operation:   operation,
		Key:         command.IdempotencyKey,
		TargetID:    command.ReceiptID,
		Fingerprint: fingerprint,
		Result:      result,
		CreatedAt:   s.now(),
	}); err != nil {
		return ReceiptResult{}, err
	}
	return result, nil
}

func (s *DiscountService) findDiscountIdempotency(ctx context.Context, operation string, key string, targetID string, fingerprint string) (ReceiptResult, bool, error) {
	record, found, err := s.idempotency.Find(ctx, operation, key)
	if err != nil || !found {
		return ReceiptResult{}, found, err
	}
	if record.TargetID != targetID || record.Fingerprint != fingerprint {
		return ReceiptResult{}, true, ErrIdempotencyKeyReused
	}
	result, ok := record.Result.(ReceiptResult)
	if !ok {
		return ReceiptResult{}, true, ErrIdempotencyResultMissing
	}
	return result, true, nil
}

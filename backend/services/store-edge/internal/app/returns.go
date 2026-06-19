package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/domain"
)

var (
	ErrReturnNotFound           = errors.New("return not found")
	ErrInvalidReturnCommand     = errors.New("invalid return command")
	ErrReceiptNotReturnable     = errors.New("receipt is not returnable")
	ErrReturnAlreadySettled     = errors.New("return is already settled")
	ErrReturnSettlementNotAllowed = errors.New("return settlement is not allowed")
)

type ReturnRepository interface {
	SaveReturn(ctx context.Context, ret domain.Return) error
	FindReturn(ctx context.Context, returnID string) (domain.Return, error)
}

type ReturnsService struct {
	receipts    ReceiptRepository
	returns     ReturnRepository
	idempotency IdempotencyStore
	roles       ActorRoleLookup
	journal     OperationJournalRecorder
	now         func() time.Time
	newID       func(prefix string) string
}

type ReturnsOption func(*ReturnsService)

func NewReturnsService(receipts ReceiptRepository, returns ReturnRepository, idempotency IdempotencyStore, roles ActorRoleLookup, options ...ReturnsOption) *ReturnsService {
	service := &ReturnsService{
		receipts:    receipts,
		returns:     returns,
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

func WithReturnsJournal(journal OperationJournalRecorder) ReturnsOption {
	return func(service *ReturnsService) {
		service.journal = journal
	}
}

type ReturnLineCommand struct {
	LineID         string
	ProductID      string
	Name           string
	Quantity       int64
	UnitPriceMinor int64
}

type CreateReceiptReturnCommand struct {
	IdempotencyKey string
	ReceiptID      string
	Lines          []ReturnLineCommand
	Reason         string
	ActorID        string
}

type CreateNoReceiptReturnCommand struct {
	IdempotencyKey string
	StoreID        string
	Lines          []ReturnLineCommand
	Reason         string
	ActorID        string
	ApprovedByID   string
}

type ReturnResult struct {
	Return domain.Return
}

func (s *ReturnsService) CreateReceiptReturn(ctx context.Context, command CreateReceiptReturnCommand) (ReturnResult, error) {
	if command.IdempotencyKey == "" {
		return ReturnResult{}, ErrIdempotencyKeyRequired
	}
	if command.ReceiptID == "" || command.Reason == "" || command.ActorID == "" || len(command.Lines) == 0 {
		return ReturnResult{}, ErrInvalidReturnCommand
	}
	if err := CheckActorPermission(s.roles, ctx, command.ActorID, PermissionReturnsCreate); err != nil {
		return ReturnResult{}, err
	}

	const operation = "returns.create_with_receipt"
	fingerprint := fmt.Sprintf("%s|%s|%s", command.ReceiptID, command.Reason, command.ActorID)
	if result, found, err := s.findReturnIdempotency(ctx, operation, command.IdempotencyKey, command.ReceiptID, fingerprint); err != nil || found {
		return result, err
	}

	receipt, err := s.receipts.FindReceipt(ctx, command.ReceiptID)
	if err != nil {
		return ReturnResult{}, err
	}
	lineInputs := toReturnLineInputs(command.Lines, receipt)
	if err := domain.ValidateReceiptReturn(receipt, lineInputs); err != nil {
		if errors.Is(err, domain.ErrReceiptNotReturnable) {
			return ReturnResult{}, ErrReceiptNotReturnable
		}
		return ReturnResult{}, ErrInvalidReturnCommand
	}

	ret, err := domain.NewReturn(domain.CreateReturnInput{
		ID:        s.newID("ret"),
		StoreID:   receipt.StoreID,
		ReceiptID: receipt.ID,
		Kind:      domain.ReturnKindWithReceipt,
		Lines:     lineInputs,
		Reason:    command.Reason,
		ActorID:   command.ActorID,
		Now:       s.now(),
	})
	if err != nil {
		return ReturnResult{}, ErrInvalidReturnCommand
	}

	if err := s.returns.SaveReturn(ctx, ret); err != nil {
		return ReturnResult{}, err
	}
	s.recordJournal(ctx, ret)

	result := ReturnResult{Return: ret}
	if err := s.idempotency.Save(ctx, IdempotencyRecord{
		Operation:   operation,
		Key:         command.IdempotencyKey,
		TargetID:    command.ReceiptID,
		Fingerprint: fingerprint,
		Result:      result,
		CreatedAt:   s.now(),
	}); err != nil {
		return ReturnResult{}, err
	}
	return result, nil
}

func (s *ReturnsService) CreateNoReceiptReturn(ctx context.Context, command CreateNoReceiptReturnCommand) (ReturnResult, error) {
	if command.IdempotencyKey == "" {
		return ReturnResult{}, ErrIdempotencyKeyRequired
	}
	if command.StoreID == "" || command.Reason == "" || command.ActorID == "" ||
		command.ApprovedByID == "" || len(command.Lines) == 0 {
		return ReturnResult{}, ErrInvalidReturnCommand
	}
	if command.ApprovedByID == command.ActorID {
		return ReturnResult{}, ErrSeparationOfDutiesViolation
	}
	if err := CheckActorPermission(s.roles, ctx, command.ActorID, PermissionReturnsCreate); err != nil {
		return ReturnResult{}, err
	}

	const operation = "returns.create_no_receipt"
	fingerprint := fmt.Sprintf("%s|%s|%s|%s", command.StoreID, command.Reason, command.ActorID, command.ApprovedByID)
	if result, found, err := s.findReturnIdempotency(ctx, operation, command.IdempotencyKey, command.StoreID, fingerprint); err != nil || found {
		return result, err
	}

	lineInputs := toReturnLineInputs(command.Lines, domain.Receipt{})
	ret, err := domain.NewReturn(domain.CreateReturnInput{
		ID:           s.newID("ret"),
		StoreID:      command.StoreID,
		Kind:         domain.ReturnKindNoReceipt,
		Lines:        lineInputs,
		Reason:       command.Reason,
		ActorID:      command.ActorID,
		ApprovedByID: command.ApprovedByID,
		Now:          s.now(),
	})
	if err != nil {
		return ReturnResult{}, ErrInvalidReturnCommand
	}

	if err := s.returns.SaveReturn(ctx, ret); err != nil {
		return ReturnResult{}, err
	}
	s.recordJournal(ctx, ret)

	result := ReturnResult{Return: ret}
	if err := s.idempotency.Save(ctx, IdempotencyRecord{
		Operation:   operation,
		Key:         command.IdempotencyKey,
		TargetID:    command.StoreID,
		Fingerprint: fingerprint,
		Result:      result,
		CreatedAt:   s.now(),
	}); err != nil {
		return ReturnResult{}, err
	}
	return result, nil
}

func (s *ReturnsService) recordJournal(ctx context.Context, ret domain.Return) {
	if s.journal == nil {
		return
	}
	_ = s.journal.RecordOperation(ctx, RecordOperationCommand{
		StoreID:       ret.StoreID,
		OperationType: "return.completed",
		ActorID:       ret.ActorID,
		ReferenceID:   ret.ID,
		Summary:       fmt.Sprintf("%s return %s total=%d", ret.Kind, ret.Reason, ret.TotalMinor),
	})
}

func toReturnLineInputs(lines []ReturnLineCommand, receipt domain.Receipt) []domain.ReturnLineInput {
	receiptLines := map[string]domain.ReceiptLine{}
	for _, line := range receipt.Lines {
		receiptLines[line.ID] = line
	}

	result := make([]domain.ReturnLineInput, 0, len(lines))
	for _, line := range lines {
		input := domain.ReturnLineInput{
			LineID:         line.LineID,
			ProductID:      line.ProductID,
			Name:           line.Name,
			Quantity:       line.Quantity,
			UnitPriceMinor: line.UnitPriceMinor,
		}
		if receiptLine, ok := receiptLines[line.LineID]; ok {
			input.ProductID = receiptLine.ProductID
			input.Name = receiptLine.Name
			input.UnitPriceMinor = receiptLine.UnitPriceMinor
		}
		result = append(result, input)
	}
	return result
}

func (s *ReturnsService) findReturnIdempotency(ctx context.Context, operation string, key string, targetID string, fingerprint string) (ReturnResult, bool, error) {
	record, found, err := s.idempotency.Find(ctx, operation, key)
	if err != nil || !found {
		return ReturnResult{}, found, err
	}
	if record.TargetID != targetID || record.Fingerprint != fingerprint {
		return ReturnResult{}, true, ErrIdempotencyKeyReused
	}
	result, ok := record.Result.(ReturnResult)
	if !ok {
		return ReturnResult{}, true, ErrIdempotencyResultMissing
	}
	return result, true, nil
}

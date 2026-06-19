package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/domain"
)

var (
	ErrFiscalDocumentNotFound              = errors.New("fiscal document not found")
	ErrInvalidFiscalizationCommand         = errors.New("invalid fiscalization command")
	ErrReceiptNotFullyPaid                 = errors.New("receipt is not fully paid")
	ErrReceiptAlreadyFiscalized            = errors.New("receipt is already fiscalized")
	ErrReturnNotFiscalizable               = errors.New("return is not fiscalizable")
	ErrReturnNotSettled                    = errors.New("return is not settled")
	ErrReturnFiscalDocumentAlreadyExists   = errors.New("return fiscal document already exists")
)

type FiscalRepository interface {
	SaveFiscalDocument(ctx context.Context, document domain.FiscalDocument) error
	FindFiscalDocumentsByReceipt(ctx context.Context, receiptID string) ([]domain.FiscalDocument, error)
	FindFiscalDocumentByReturn(ctx context.Context, returnID string) (domain.FiscalDocument, error)
}

type FiscalReceiptPrinter interface {
	PrintReceipt(ctx context.Context, deviceID string, totalMinor int64) (string, error)
}

type FiscalizationService struct {
	receipts              ReceiptRepository
	payments              PaymentRepository
	fiscal                FiscalRepository
	returns               ReturnRepository
	idempotency           IdempotencyStore
	outbox                OutboxRecorder
	fiscalPrinter         FiscalReceiptPrinter
	hardwareAgentFallback bool
	now                   func() time.Time
	newID                 func(prefix string) string
}

type FiscalizationOption func(*FiscalizationService)

func NewFiscalizationService(receipts ReceiptRepository, payments PaymentRepository, fiscal FiscalRepository, returns ReturnRepository, idempotency IdempotencyStore, options ...FiscalizationOption) *FiscalizationService {
	service := &FiscalizationService{
		receipts:    receipts,
		payments:    payments,
		fiscal:      fiscal,
		returns:     returns,
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

func WithFiscalizationClock(now func() time.Time) FiscalizationOption {
	return func(service *FiscalizationService) {
		service.now = now
	}
}

func WithFiscalizationIDGenerator(newID func(prefix string) string) FiscalizationOption {
	return func(service *FiscalizationService) {
		service.newID = newID
	}
}

func WithFiscalizationOutboxRecorder(outbox OutboxRecorder) FiscalizationOption {
	return func(service *FiscalizationService) {
		service.outbox = outbox
	}
}

func WithFiscalReceiptPrinter(printer FiscalReceiptPrinter, fallback bool) FiscalizationOption {
	return func(service *FiscalizationService) {
		service.fiscalPrinter = printer
		service.hardwareAgentFallback = fallback
	}
}

type CreateFiscalDocumentCommand struct {
	IdempotencyKey string
	ReceiptID      string
	DeviceID       string
}

type CreateReturnFiscalDocumentCommand struct {
	IdempotencyKey string
	ReturnID       string
	DeviceID       string
}

type FiscalDocumentResult struct {
	Document domain.FiscalDocument
}

func (s *FiscalizationService) CreateFiscalDocument(ctx context.Context, command CreateFiscalDocumentCommand) (FiscalDocumentResult, error) {
	if command.IdempotencyKey == "" {
		return FiscalDocumentResult{}, ErrIdempotencyKeyRequired
	}
	if command.ReceiptID == "" || command.DeviceID == "" {
		return FiscalDocumentResult{}, ErrInvalidFiscalizationCommand
	}

	const operation = "fiscalization.create_fiscal_document"
	fingerprint := fmt.Sprintf("%s|%s", command.ReceiptID, command.DeviceID)
	if result, found, err := s.findFiscalIdempotency(ctx, operation, command.IdempotencyKey, command.ReceiptID, fingerprint); err != nil || found {
		return result, err
	}

	receipt, err := s.receipts.FindReceipt(ctx, command.ReceiptID)
	if err != nil {
		return FiscalDocumentResult{}, err
	}
	existingDocuments, err := s.fiscal.FindFiscalDocumentsByReceipt(ctx, command.ReceiptID)
	if err != nil {
		return FiscalDocumentResult{}, err
	}
	for _, document := range existingDocuments {
		if document.Kind == domain.FiscalDocumentKindReceipt {
			return FiscalDocumentResult{}, ErrReceiptAlreadyFiscalized
		}
	}
	payments, err := s.payments.FindPaymentsByReceipt(ctx, command.ReceiptID)
	if err != nil {
		return FiscalDocumentResult{}, err
	}
	if remainingAmountMinor(receipt, payments) != 0 {
		return FiscalDocumentResult{}, ErrReceiptNotFullyPaid
	}
	if receipt.Status != domain.ReceiptStatusPaid {
		return FiscalDocumentResult{}, ErrReceiptNotFullyPaid
	}

	fiscalSign := s.newID("fs")
	if s.fiscalPrinter != nil {
		printedSign, err := s.fiscalPrinter.PrintReceipt(ctx, command.DeviceID, receipt.TotalMinor())
		if err != nil {
			if !s.hardwareAgentFallback {
				return FiscalDocumentResult{}, err
			}
		} else {
			fiscalSign = printedSign
		}
	}

	document, err := domain.CreateFiscalizedDocument(domain.CreateFiscalizedDocumentInput{
		ID:          s.newID("fis"),
		ReceiptID:   command.ReceiptID,
		Kind:        domain.FiscalDocumentKindReceipt,
		AmountMinor: receipt.TotalMinor(),
		DeviceID:    command.DeviceID,
		FiscalSign:  fiscalSign,
		Now:         s.now(),
	})
	if err != nil {
		return FiscalDocumentResult{}, err
	}
	if err := s.fiscal.SaveFiscalDocument(ctx, document); err != nil {
		return FiscalDocumentResult{}, err
	}
	if err := receipt.MarkFiscalized(s.now()); err != nil {
		return FiscalDocumentResult{}, err
	}
	if err := s.receipts.SaveReceipt(ctx, receipt); err != nil {
		return FiscalDocumentResult{}, err
	}

	result := FiscalDocumentResult{Document: document}
	if err := s.idempotency.Save(ctx, IdempotencyRecord{
		Operation:   operation,
		Key:         command.IdempotencyKey,
		TargetID:    command.ReceiptID,
		Fingerprint: fingerprint,
		Result:      result,
		CreatedAt:   s.now(),
	}); err != nil {
		return FiscalDocumentResult{}, err
	}
	if err := recordOutbox(ctx, s.outbox, func(ctx context.Context, recorder OutboxRecorder) error {
		return recorder.RecordFiscalDocumentCreated(ctx, document, receipt.StoreID)
	}); err != nil {
		return FiscalDocumentResult{}, err
	}
	return result, nil
}

func (s *FiscalizationService) CreateReturnFiscalDocument(ctx context.Context, command CreateReturnFiscalDocumentCommand) (FiscalDocumentResult, error) {
	if command.IdempotencyKey == "" {
		return FiscalDocumentResult{}, ErrIdempotencyKeyRequired
	}
	if command.ReturnID == "" || command.DeviceID == "" {
		return FiscalDocumentResult{}, ErrInvalidFiscalizationCommand
	}

	const operation = "fiscalization.create_return_fiscal_document"
	fingerprint := fmt.Sprintf("%s|%s", command.ReturnID, command.DeviceID)
	if result, found, err := s.findFiscalIdempotency(ctx, operation, command.IdempotencyKey, command.ReturnID, fingerprint); err != nil || found {
		return result, err
	}

	ret, err := s.returns.FindReturn(ctx, command.ReturnID)
	if err != nil {
		return FiscalDocumentResult{}, err
	}
	if ret.Kind != domain.ReturnKindWithReceipt {
		return FiscalDocumentResult{}, ErrReturnNotFiscalizable
	}
	if ret.Status != domain.ReturnStatusSettled {
		return FiscalDocumentResult{}, ErrReturnNotSettled
	}

	receipt, err := s.receipts.FindReceipt(ctx, ret.ReceiptID)
	if err != nil {
		return FiscalDocumentResult{}, err
	}
	if receipt.Status != domain.ReceiptStatusFiscalized {
		return FiscalDocumentResult{}, ErrReceiptNotReturnable
	}

	if _, err := s.fiscal.FindFiscalDocumentByReturn(ctx, command.ReturnID); err == nil {
		return FiscalDocumentResult{}, ErrReturnFiscalDocumentAlreadyExists
	} else if !errors.Is(err, ErrFiscalDocumentNotFound) {
		return FiscalDocumentResult{}, err
	}

	receiptDocuments, err := s.fiscal.FindFiscalDocumentsByReceipt(ctx, ret.ReceiptID)
	if err != nil {
		return FiscalDocumentResult{}, err
	}
	hasSaleDocument := false
	for _, document := range receiptDocuments {
		if document.Kind == domain.FiscalDocumentKindReceipt {
			hasSaleDocument = true
			break
		}
	}
	if !hasSaleDocument {
		return FiscalDocumentResult{}, ErrReceiptNotReturnable
	}

	fiscalSign := s.newID("fs")
	if s.fiscalPrinter != nil {
		printedSign, err := s.fiscalPrinter.PrintReceipt(ctx, command.DeviceID, ret.TotalMinor)
		if err != nil {
			if !s.hardwareAgentFallback {
				return FiscalDocumentResult{}, err
			}
		} else {
			fiscalSign = printedSign
		}
	}

	document, err := domain.CreateFiscalizedDocument(domain.CreateFiscalizedDocumentInput{
		ID:          s.newID("fis"),
		ReceiptID:   ret.ReceiptID,
		ReturnID:    ret.ID,
		Kind:        domain.FiscalDocumentKindReturn,
		AmountMinor: ret.TotalMinor,
		DeviceID:    command.DeviceID,
		FiscalSign:  fiscalSign,
		Now:         s.now(),
	})
	if err != nil {
		return FiscalDocumentResult{}, err
	}
	if err := s.fiscal.SaveFiscalDocument(ctx, document); err != nil {
		return FiscalDocumentResult{}, err
	}

	result := FiscalDocumentResult{Document: document}
	if err := s.idempotency.Save(ctx, IdempotencyRecord{
		Operation:   operation,
		Key:         command.IdempotencyKey,
		TargetID:    command.ReturnID,
		Fingerprint: fingerprint,
		Result:      result,
		CreatedAt:   s.now(),
	}); err != nil {
		return FiscalDocumentResult{}, err
	}
	if err := recordOutbox(ctx, s.outbox, func(ctx context.Context, recorder OutboxRecorder) error {
		return recorder.RecordFiscalDocumentCreated(ctx, document, receipt.StoreID)
	}); err != nil {
		return FiscalDocumentResult{}, err
	}
	return result, nil
}

func (s *FiscalizationService) ListReturnFiscalDocuments(ctx context.Context, returnID string) ([]domain.FiscalDocument, error) {
	if _, err := s.returns.FindReturn(ctx, returnID); err != nil {
		return nil, err
	}
	document, err := s.fiscal.FindFiscalDocumentByReturn(ctx, returnID)
	if errors.Is(err, ErrFiscalDocumentNotFound) {
		return []domain.FiscalDocument{}, nil
	}
	if err != nil {
		return nil, err
	}
	return []domain.FiscalDocument{document}, nil
}

func (s *FiscalizationService) ListReceiptFiscalDocuments(ctx context.Context, receiptID string) ([]domain.FiscalDocument, error) {
	if _, err := s.receipts.FindReceipt(ctx, receiptID); err != nil {
		return nil, err
	}
	return s.fiscal.FindFiscalDocumentsByReceipt(ctx, receiptID)
}

func (s *FiscalizationService) findFiscalIdempotency(ctx context.Context, operation string, key string, targetID string, fingerprint string) (FiscalDocumentResult, bool, error) {
	record, found, err := s.idempotency.Find(ctx, operation, key)
	if err != nil || !found {
		return FiscalDocumentResult{}, found, err
	}
	if record.TargetID != targetID || record.Fingerprint != fingerprint {
		return FiscalDocumentResult{}, true, ErrIdempotencyKeyReused
	}
	result, ok := record.Result.(FiscalDocumentResult)
	if !ok {
		return FiscalDocumentResult{}, true, ErrIdempotencyResultMissing
	}
	return result, true, nil
}

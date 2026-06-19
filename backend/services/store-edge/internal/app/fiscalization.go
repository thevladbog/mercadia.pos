package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/domain"
)

var (
	ErrFiscalDocumentNotFound      = errors.New("fiscal document not found")
	ErrInvalidFiscalizationCommand = errors.New("invalid fiscalization command")
	ErrReceiptNotFullyPaid         = errors.New("receipt is not fully paid")
	ErrReceiptAlreadyFiscalized    = errors.New("receipt is already fiscalized")
)

type FiscalRepository interface {
	SaveFiscalDocument(ctx context.Context, document domain.FiscalDocument) error
	FindFiscalDocumentsByReceipt(ctx context.Context, receiptID string) ([]domain.FiscalDocument, error)
}

type FiscalReceiptPrinter interface {
	PrintReceipt(ctx context.Context, deviceID string, totalMinor int64) (string, error)
}

type FiscalizationService struct {
	receipts              ReceiptRepository
	payments              PaymentRepository
	fiscal                FiscalRepository
	idempotency           IdempotencyStore
	outbox                OutboxRecorder
	fiscalPrinter         FiscalReceiptPrinter
	hardwareAgentFallback bool
	now                   func() time.Time
	newID                 func(prefix string) string
}

type FiscalizationOption func(*FiscalizationService)

func NewFiscalizationService(receipts ReceiptRepository, payments PaymentRepository, fiscal FiscalRepository, idempotency IdempotencyStore, options ...FiscalizationOption) *FiscalizationService {
	service := &FiscalizationService{
		receipts:    receipts,
		payments:    payments,
		fiscal:      fiscal,
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
	if len(existingDocuments) > 0 {
		return FiscalDocumentResult{}, ErrReceiptAlreadyFiscalized
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

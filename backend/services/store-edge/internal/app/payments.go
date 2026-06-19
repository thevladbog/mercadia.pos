package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/domain"
)

var (
	ErrPaymentNotFound               = errors.New("payment not found")
	ErrInvalidPaymentCommand         = errors.New("invalid payment command")
	ErrPaymentAmountExceedsRemaining = errors.New("payment amount exceeds receipt remaining amount")
	ErrCashDrawerRequired            = errors.New("cash drawer is required for cash payment")
)

type PaymentRepository interface {
	SavePayment(ctx context.Context, payment domain.Payment) error
	FindPaymentsByReceipt(ctx context.Context, receiptID string) ([]domain.Payment, error)
}

type CardPaymentTerminal interface {
	AuthorizeAndCapture(ctx context.Context, deviceID string, amountMinor int64, currency, reference string) (string, error)
}

type PaymentService struct {
	receipts            ReceiptRepository
	payments            PaymentRepository
	cash                CashRepository
	idempotency         IdempotencyStore
	outbox              OutboxRecorder
	cardTerminal        CardPaymentTerminal
	paymentTerminalID   string
	hardwareAgentFallback bool
	now                 func() time.Time
	newID               func(prefix string) string
}

type PaymentOption func(*PaymentService)

func NewPaymentService(receipts ReceiptRepository, payments PaymentRepository, idempotency IdempotencyStore, options ...PaymentOption) *PaymentService {
	service := &PaymentService{
		receipts:    receipts,
		payments:    payments,
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

func WithPaymentClock(now func() time.Time) PaymentOption {
	return func(service *PaymentService) {
		service.now = now
	}
}

func WithPaymentIDGenerator(newID func(prefix string) string) PaymentOption {
	return func(service *PaymentService) {
		service.newID = newID
	}
}

func WithPaymentCashLedger(cash CashRepository) PaymentOption {
	return func(service *PaymentService) {
		service.cash = cash
	}
}

func WithPaymentOutboxRecorder(outbox OutboxRecorder) PaymentOption {
	return func(service *PaymentService) {
		service.outbox = outbox
	}
}

func WithCardPaymentTerminal(terminal CardPaymentTerminal, deviceID string, fallback bool) PaymentOption {
	return func(service *PaymentService) {
		service.cardTerminal = terminal
		service.paymentTerminalID = deviceID
		service.hardwareAgentFallback = fallback
	}
}

type CreatePaymentCommand struct {
	IdempotencyKey    string
	ReceiptID         string
	Method            domain.PaymentMethod
	AmountMinor       int64
	ProviderReference string
}

type PaymentResult struct {
	Payment domain.Payment
}

func (s *PaymentService) CreatePayment(ctx context.Context, command CreatePaymentCommand) (PaymentResult, error) {
	if command.IdempotencyKey == "" {
		return PaymentResult{}, ErrIdempotencyKeyRequired
	}
	if command.ReceiptID == "" || command.Method == "" || command.AmountMinor <= 0 {
		return PaymentResult{}, ErrInvalidPaymentCommand
	}

	const operation = "payments.create_payment"
	fingerprint := fmt.Sprintf("%s|%s|%d|%s", command.ReceiptID, command.Method, command.AmountMinor, command.ProviderReference)
	if result, found, err := s.findPaymentIdempotency(ctx, operation, command.IdempotencyKey, command.ReceiptID, fingerprint); err != nil || found {
		return result, err
	}

	receipt, err := s.receipts.FindReceipt(ctx, command.ReceiptID)
	if err != nil {
		return PaymentResult{}, err
	}
	existingPayments, err := s.payments.FindPaymentsByReceipt(ctx, command.ReceiptID)
	if err != nil {
		return PaymentResult{}, err
	}
	remainingBeforePayment := remainingAmountMinor(receipt, existingPayments)
	if command.AmountMinor > remainingBeforePayment {
		return PaymentResult{}, ErrPaymentAmountExceedsRemaining
	}

	providerReference := command.ProviderReference
	if command.Method == domain.PaymentMethodCardMock && s.cardTerminal != nil && s.paymentTerminalID != "" {
		reference := command.ReceiptID
		if providerReference != "" {
			reference = providerReference
		}
		terminalRef, err := s.cardTerminal.AuthorizeAndCapture(ctx, s.paymentTerminalID, command.AmountMinor, "RUB", reference)
		if err != nil {
			if !s.hardwareAgentFallback {
				return PaymentResult{}, err
			}
		} else if providerReference == "" {
			providerReference = terminalRef
		}
	}

	payment, err := domain.CreateCapturedPayment(domain.CreateCapturedPaymentInput{
		ID:                s.newID("pay"),
		ReceiptID:         command.ReceiptID,
		Method:            command.Method,
		AmountMinor:       command.AmountMinor,
		ProviderReference: providerReference,
		Now:               s.now(),
	})
	if err != nil {
		return PaymentResult{}, err
	}

	if err := s.payments.SavePayment(ctx, payment); err != nil {
		return PaymentResult{}, err
	}
	if command.Method == domain.PaymentMethodCash && s.cash != nil {
		if receipt.DrawerID == "" {
			return PaymentResult{}, ErrCashDrawerRequired
		}
		movement, err := domain.CreateCashMovement(domain.CreateCashMovementInput{
			ID:                s.newID("cash"),
			StoreID:           receipt.StoreID,
			Type:              domain.CashMovementTypeCashSale,
			FromContainerID:   "external-customer",
			FromContainerType: domain.CashContainerTypeExternal,
			ToContainerID:     receipt.DrawerID,
			ToContainerType:   domain.CashContainerTypeDrawer,
			AmountMinor:       command.AmountMinor,
			Currency:          "RUB",
			Reason:            "Cash payment for receipt " + receipt.ID,
			ActorID:           receipt.CashierID,
			Now:               s.now(),
		})
		if err != nil {
			return PaymentResult{}, err
		}
		if err := s.cash.SaveCashMovement(ctx, movement); err != nil {
			return PaymentResult{}, err
		}
	}
	if command.AmountMinor == remainingBeforePayment {
		if err := receipt.MarkPaid(s.now()); err != nil {
			return PaymentResult{}, err
		}
	} else {
		if err := receipt.MarkPaymentStarted(s.now()); err != nil {
			return PaymentResult{}, err
		}
	}
	if err := s.receipts.SaveReceipt(ctx, receipt); err != nil {
		return PaymentResult{}, err
	}

	result := PaymentResult{Payment: payment}
	if err := s.idempotency.Save(ctx, IdempotencyRecord{
		Operation:   operation,
		Key:         command.IdempotencyKey,
		TargetID:    command.ReceiptID,
		Fingerprint: fingerprint,
		Result:      result,
		CreatedAt:   s.now(),
	}); err != nil {
		return PaymentResult{}, err
	}
	if err := recordOutbox(ctx, s.outbox, func(ctx context.Context, recorder OutboxRecorder) error {
		return recorder.RecordPaymentCaptured(ctx, payment, receipt.StoreID)
	}); err != nil {
		return PaymentResult{}, err
	}

	return result, nil
}

func (s *PaymentService) ListReceiptPayments(ctx context.Context, receiptID string) ([]domain.Payment, error) {
	if _, err := s.receipts.FindReceipt(ctx, receiptID); err != nil {
		return nil, err
	}
	return s.payments.FindPaymentsByReceipt(ctx, receiptID)
}

func (s *PaymentService) findPaymentIdempotency(ctx context.Context, operation string, key string, targetID string, fingerprint string) (PaymentResult, bool, error) {
	record, found, err := s.idempotency.Find(ctx, operation, key)
	if err != nil || !found {
		return PaymentResult{}, found, err
	}
	if record.TargetID != targetID || record.Fingerprint != fingerprint {
		return PaymentResult{}, true, ErrIdempotencyKeyReused
	}
	result, ok := record.Result.(PaymentResult)
	if !ok {
		return PaymentResult{}, true, ErrIdempotencyResultMissing
	}
	return result, true, nil
}

func remainingAmountMinor(receipt domain.Receipt, payments []domain.Payment) int64 {
	paid := int64(0)
	for _, payment := range payments {
		if payment.Status == domain.PaymentStatusCaptured {
			paid += payment.AmountMinor
		}
	}
	return receipt.TotalMinor() - paid
}

package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/app"
	"mercadia.dev/pos/services/store-edge/internal/domain"
	"mercadia.dev/pos/services/store-edge/internal/infra/memory"
)

func TestCreateFiscalDocumentRequiresFullPayment(t *testing.T) {
	checkout, _, fiscalization := newTestCheckoutPaymentAndFiscalizationServices()
	receiptID := openAndScanTestReceipt(t, checkout)

	_, err := fiscalization.CreateFiscalDocument(context.Background(), app.CreateFiscalDocumentCommand{
		IdempotencyKey: "fiscal-1",
		ReceiptID:      receiptID,
		DeviceID:       "mock-atol-1",
	})
	if !errors.Is(err, app.ErrReceiptNotFullyPaid) {
		t.Fatalf("expected ErrReceiptNotFullyPaid, got %v", err)
	}
}

func TestCreateFiscalDocumentAfterFullPayment(t *testing.T) {
	checkout, payments, fiscalization := newTestCheckoutPaymentAndFiscalizationServices()
	receiptID := openAndScanTestReceipt(t, checkout)
	if _, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCash,
		AmountMinor:    39998,
	}); err != nil {
		t.Fatalf("create payment: %v", err)
	}

	result, err := fiscalization.CreateFiscalDocument(context.Background(), app.CreateFiscalDocumentCommand{
		IdempotencyKey: "fiscal-1",
		ReceiptID:      receiptID,
		DeviceID:       "mock-atol-1",
	})
	if err != nil {
		t.Fatalf("create fiscal document: %v", err)
	}

	if result.Document.Status != domain.FiscalDocumentStatusFiscalized {
		t.Fatalf("fiscal status = %s", result.Document.Status)
	}

	receipt, err := checkout.GetReceipt(context.Background(), receiptID)
	if err != nil {
		t.Fatalf("get receipt: %v", err)
	}
	if receipt.Receipt.Status != domain.ReceiptStatusFiscalized {
		t.Fatalf("receipt status = %s", receipt.Receipt.Status)
	}
}

func TestCreateFiscalDocumentIsIdempotent(t *testing.T) {
	checkout, payments, fiscalization := newTestCheckoutPaymentAndFiscalizationServices()
	receiptID := openAndScanTestReceipt(t, checkout)
	if _, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCash,
		AmountMinor:    39998,
	}); err != nil {
		t.Fatalf("create payment: %v", err)
	}
	command := app.CreateFiscalDocumentCommand{
		IdempotencyKey: "fiscal-1",
		ReceiptID:      receiptID,
		DeviceID:       "mock-atol-1",
	}

	first, err := fiscalization.CreateFiscalDocument(context.Background(), command)
	if err != nil {
		t.Fatalf("create first fiscal document: %v", err)
	}
	second, err := fiscalization.CreateFiscalDocument(context.Background(), command)
	if err != nil {
		t.Fatalf("create second fiscal document: %v", err)
	}

	if first.Document.ID != second.Document.ID {
		t.Fatalf("expected same fiscal document id, got %s and %s", first.Document.ID, second.Document.ID)
	}
}

func newTestCheckoutPaymentAndFiscalizationServices() (*app.CheckoutService, *app.PaymentService, *app.FiscalizationService) {
	store := memory.NewStore(memory.WithProducts(testProduct()))
	var counter int
	now := func() time.Time {
		return time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
	}
	newID := func(prefix string) string {
		counter++
		return fmt.Sprintf("%s-test-%d", prefix, counter)
	}

	checkout := app.NewCheckoutService(store, store,
		app.WithProductRepository(store),
		app.WithClock(now),
		app.WithIDGenerator(newID),
	)
	payments := app.NewPaymentService(store, store, store,
		app.WithPaymentClock(now),
		app.WithPaymentIDGenerator(newID),
	)
	fiscalization := app.NewFiscalizationService(store, store, store, store,
		app.WithFiscalizationClock(now),
		app.WithFiscalizationIDGenerator(newID),
	)
	return checkout, payments, fiscalization
}

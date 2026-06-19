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

func TestCreateReturnFiscalDocumentAfterSettledReturn(t *testing.T) {
	store, checkout, payments, fiscalization, returns, settlement, _ := newReturnSettlementServices(t)
	receiptID := openOperationalReceiptAndScanTestProduct(t, store, checkout)

	if _, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCash,
		AmountMinor:    39998,
	}); err != nil {
		t.Fatalf("create payment: %v", err)
	}
	if _, err := fiscalization.CreateFiscalDocument(context.Background(), app.CreateFiscalDocumentCommand{
		IdempotencyKey: "fiscal-1",
		ReceiptID:      receiptID,
		DeviceID:       "mock-atol-1",
	}); err != nil {
		t.Fatalf("create fiscal document: %v", err)
	}

	ret := createFullReceiptReturn(t, store, returns, receiptID)
	if _, err := settlement.SettleReturn(context.Background(), app.SettleReturnCommand{
		IdempotencyKey: "settle-1",
		ReturnID:       ret.ID,
	}); err != nil {
		t.Fatalf("settle return: %v", err)
	}

	result, err := fiscalization.CreateReturnFiscalDocument(context.Background(), app.CreateReturnFiscalDocumentCommand{
		IdempotencyKey: "return-fiscal-1",
		ReturnID:       ret.ID,
		DeviceID:       "mock-atol-1",
	})
	if err != nil {
		t.Fatalf("create return fiscal document: %v", err)
	}
	if result.Document.Kind != domain.FiscalDocumentKindReturn {
		t.Fatalf("document kind = %s", result.Document.Kind)
	}
	if result.Document.ReturnID != ret.ID {
		t.Fatalf("return id = %s", result.Document.ReturnID)
	}
	if result.Document.AmountMinor != ret.TotalMinor {
		t.Fatalf("amount = %d", result.Document.AmountMinor)
	}

	receipt, err := checkout.GetReceipt(context.Background(), receiptID)
	if err != nil {
		t.Fatalf("get receipt: %v", err)
	}
	if receipt.Receipt.Status != domain.ReceiptStatusFiscalized {
		t.Fatalf("receipt status = %s", receipt.Receipt.Status)
	}
}

func TestCreateReturnFiscalDocumentBlocksUnsettledReturn(t *testing.T) {
	store, checkout, payments, fiscalization, returns, _, _ := newReturnSettlementServices(t)
	receiptID := openOperationalReceiptAndScanTestProduct(t, store, checkout)
	if _, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCash,
		AmountMinor:    39998,
	}); err != nil {
		t.Fatalf("create payment: %v", err)
	}
	if _, err := fiscalization.CreateFiscalDocument(context.Background(), app.CreateFiscalDocumentCommand{
		IdempotencyKey: "fiscal-1",
		ReceiptID:      receiptID,
		DeviceID:       "mock-atol-1",
	}); err != nil {
		t.Fatalf("create fiscal document: %v", err)
	}
	ret := createFullReceiptReturn(t, store, returns, receiptID)

	_, err := fiscalization.CreateReturnFiscalDocument(context.Background(), app.CreateReturnFiscalDocumentCommand{
		IdempotencyKey: "return-fiscal-1",
		ReturnID:       ret.ID,
		DeviceID:       "mock-atol-1",
	})
	if !errors.Is(err, app.ErrReturnNotSettled) {
		t.Fatalf("expected ErrReturnNotSettled, got %v", err)
	}
}

func TestCreateReturnFiscalDocumentBlocksNoReceiptReturn(t *testing.T) {
	store := memory.NewStore(memory.WithDemoActors())
	auth := app.NewAuthService(store, store)
	returns := app.NewReturnsService(store, store, store, auth)
	fiscalization := app.NewFiscalizationService(store, store, store, store, store)

	noReceipt, err := returns.CreateNoReceiptReturn(context.Background(), app.CreateNoReceiptReturnCommand{
		IdempotencyKey: "return-no-receipt",
		StoreID:        "store-1",
		Lines:          []app.ReturnLineCommand{{ProductID: "sku-1", Name: "Milk", Quantity: 1, UnitPriceMinor: 1000}},
		Reason:         "No receipt",
		ActorID:        "senior-1",
		ApprovedByID:   "admin-1",
	})
	if err != nil {
		t.Fatalf("create no-receipt return: %v", err)
	}

	_, err = fiscalization.CreateReturnFiscalDocument(context.Background(), app.CreateReturnFiscalDocumentCommand{
		IdempotencyKey: "return-fiscal-1",
		ReturnID:       noReceipt.Return.ID,
		DeviceID:       "mock-atol-1",
	})
	if !errors.Is(err, app.ErrReturnNotFiscalizable) {
		t.Fatalf("expected ErrReturnNotFiscalizable, got %v", err)
	}
}

func TestCreateReturnFiscalDocumentIsIdempotent(t *testing.T) {
	store, checkout, payments, fiscalization, returns, settlement, _ := newReturnSettlementServices(t)
	receiptID := openOperationalReceiptAndScanTestProduct(t, store, checkout)
	if _, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCash,
		AmountMinor:    39998,
	}); err != nil {
		t.Fatalf("create payment: %v", err)
	}
	if _, err := fiscalization.CreateFiscalDocument(context.Background(), app.CreateFiscalDocumentCommand{
		IdempotencyKey: "fiscal-1",
		ReceiptID:      receiptID,
		DeviceID:       "mock-atol-1",
	}); err != nil {
		t.Fatalf("create fiscal document: %v", err)
	}
	ret := createFullReceiptReturn(t, store, returns, receiptID)
	if _, err := settlement.SettleReturn(context.Background(), app.SettleReturnCommand{
		IdempotencyKey: "settle-1",
		ReturnID:       ret.ID,
	}); err != nil {
		t.Fatalf("settle return: %v", err)
	}

	command := app.CreateReturnFiscalDocumentCommand{
		IdempotencyKey: "return-fiscal-1",
		ReturnID:       ret.ID,
		DeviceID:       "mock-atol-1",
	}
	first, err := fiscalization.CreateReturnFiscalDocument(context.Background(), command)
	if err != nil {
		t.Fatalf("create first return fiscal document: %v", err)
	}
	second, err := fiscalization.CreateReturnFiscalDocument(context.Background(), command)
	if err != nil {
		t.Fatalf("create second return fiscal document: %v", err)
	}
	if first.Document.ID != second.Document.ID {
		t.Fatalf("expected same fiscal document id, got %s and %s", first.Document.ID, second.Document.ID)
	}
}

func TestCreateReturnFiscalDocumentBlocksDuplicate(t *testing.T) {
	store, checkout, payments, fiscalization, returns, settlement, _ := newReturnSettlementServices(t)
	receiptID := openOperationalReceiptAndScanTestProduct(t, store, checkout)
	if _, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCash,
		AmountMinor:    39998,
	}); err != nil {
		t.Fatalf("create payment: %v", err)
	}
	if _, err := fiscalization.CreateFiscalDocument(context.Background(), app.CreateFiscalDocumentCommand{
		IdempotencyKey: "fiscal-1",
		ReceiptID:      receiptID,
		DeviceID:       "mock-atol-1",
	}); err != nil {
		t.Fatalf("create fiscal document: %v", err)
	}
	ret := createFullReceiptReturn(t, store, returns, receiptID)
	if _, err := settlement.SettleReturn(context.Background(), app.SettleReturnCommand{
		IdempotencyKey: "settle-1",
		ReturnID:       ret.ID,
	}); err != nil {
		t.Fatalf("settle return: %v", err)
	}

	if _, err := fiscalization.CreateReturnFiscalDocument(context.Background(), app.CreateReturnFiscalDocumentCommand{
		IdempotencyKey: "return-fiscal-1",
		ReturnID:       ret.ID,
		DeviceID:       "mock-atol-1",
	}); err != nil {
		t.Fatalf("create return fiscal document: %v", err)
	}

	_, err := fiscalization.CreateReturnFiscalDocument(context.Background(), app.CreateReturnFiscalDocumentCommand{
		IdempotencyKey: "return-fiscal-2",
		ReturnID:       ret.ID,
		DeviceID:       "mock-atol-1",
	})
	if !errors.Is(err, app.ErrReturnFiscalDocumentAlreadyExists) {
		t.Fatalf("expected ErrReturnFiscalDocumentAlreadyExists, got %v", err)
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
	fiscalization := app.NewFiscalizationService(store, store, store, store, store,
		app.WithFiscalizationClock(now),
		app.WithFiscalizationIDGenerator(newID),
	)
	return checkout, payments, fiscalization
}

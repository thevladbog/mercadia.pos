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

func TestSettleReturnRefundsCapturedCashPayment(t *testing.T) {
	store, checkout, payments, fiscalization, returns, settlement, cash := newReturnSettlementServices(t)
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

	result, err := settlement.SettleReturn(context.Background(), app.SettleReturnCommand{
		IdempotencyKey: "settle-1",
		ReturnID:       ret.ID,
		ActorID:        "senior-1",
		Reason:         "Customer return",
	})
	if err != nil {
		t.Fatalf("settle return: %v", err)
	}
	if result.Return.Status != domain.ReturnStatusSettled {
		t.Fatalf("return status = %s", result.Return.Status)
	}
	if len(result.Payments) != 1 || result.Payments[0].Status != domain.PaymentStatusRefunded {
		t.Fatalf("payments = %+v", result.Payments)
	}

	movements, err := cash.ListCashMovements(context.Background(), "store-1", app.PageParams{Limit: 50})
	if err != nil {
		t.Fatalf("list cash movements: %v", err)
	}
	var saleCount, reversalCount int
	for _, movement := range movements.Items {
		switch movement.Type {
		case domain.CashMovementTypeCashSale:
			saleCount++
		case domain.CashMovementTypeCashSaleReversal:
			reversalCount++
		}
	}
	if saleCount != 1 || reversalCount != 1 {
		t.Fatalf("cash_sale=%d cash_sale_reversal=%d", saleCount, reversalCount)
	}
}

func TestSettleReturnRefundsMixedCashAndCardPayments(t *testing.T) {
	store, checkout, payments, fiscalization, returns, settlement, _ := newReturnSettlementServices(t)
	receiptID := openOperationalReceiptAndScanTestProduct(t, store, checkout)

	if _, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCash,
		AmountMinor:    20000,
	}); err != nil {
		t.Fatalf("create cash payment: %v", err)
	}
	if _, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-2",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCardMock,
		AmountMinor:    19998,
	}); err != nil {
		t.Fatalf("create card payment: %v", err)
	}
	if _, err := fiscalization.CreateFiscalDocument(context.Background(), app.CreateFiscalDocumentCommand{
		IdempotencyKey: "fiscal-1",
		ReceiptID:      receiptID,
		DeviceID:       "mock-atol-1",
	}); err != nil {
		t.Fatalf("create fiscal document: %v", err)
	}

	ret := createFullReceiptReturn(t, store, returns, receiptID)

	result, err := settlement.SettleReturn(context.Background(), app.SettleReturnCommand{
		IdempotencyKey: "settle-1",
		ReturnID:       ret.ID,
	})
	if err != nil {
		t.Fatalf("settle return: %v", err)
	}
	if len(result.Payments) != 2 {
		t.Fatalf("payments count = %d", len(result.Payments))
	}
	for _, payment := range result.Payments {
		if payment.Status != domain.PaymentStatusRefunded {
			t.Fatalf("payment %s status = %s", payment.ID, payment.Status)
		}
	}
}

func TestSettleReturnBlocksPartialReturn(t *testing.T) {
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

	receipt, err := store.FindReceipt(context.Background(), receiptID)
	if err != nil {
		t.Fatalf("find receipt: %v", err)
	}
	if len(receipt.Lines) != 1 {
		t.Fatalf("lines count = %d", len(receipt.Lines))
	}

	partial, err := returns.CreateReceiptReturn(context.Background(), app.CreateReceiptReturnCommand{
		IdempotencyKey: "return-partial",
		ReceiptID:      receiptID,
		Lines:          []app.ReturnLineCommand{{LineID: receipt.Lines[0].ID, Quantity: 1}},
		Reason:         "Partial return",
		ActorID:        "senior-1",
	})
	if err != nil {
		t.Fatalf("create partial return: %v", err)
	}

	_, err = settlement.SettleReturn(context.Background(), app.SettleReturnCommand{
		IdempotencyKey: "settle-1",
		ReturnID:       partial.Return.ID,
	})
	if !errors.Is(err, app.ErrReturnSettlementRequiresFullReceiptReturn) {
		t.Fatalf("expected full return required, got %v", err)
	}
}

func TestSettleReturnBlocksAlreadySettled(t *testing.T) {
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
		t.Fatalf("first settle: %v", err)
	}

	_, err := settlement.SettleReturn(context.Background(), app.SettleReturnCommand{
		IdempotencyKey: "settle-2",
		ReturnID:       ret.ID,
	})
	if !errors.Is(err, app.ErrReturnAlreadySettled) {
		t.Fatalf("expected already settled, got %v", err)
	}
}

func TestSettleReturnBlocksNoReceiptReturn(t *testing.T) {
	store := memory.NewStore(memory.WithDemoActors())
	auth := app.NewAuthService(store, store)
	returns := app.NewReturnsService(store, store, store, auth)
	settlement := app.NewReturnSettlementService(store, store, store, nil, store)

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

	_, err = settlement.SettleReturn(context.Background(), app.SettleReturnCommand{
		IdempotencyKey: "settle-1",
		ReturnID:       noReceipt.Return.ID,
	})
	if !errors.Is(err, app.ErrReturnSettlementNotAllowed) {
		t.Fatalf("expected settlement not allowed, got %v", err)
	}
}

func newReturnSettlementServices(t *testing.T) (*memory.Store, *app.CheckoutService, *app.PaymentService, *app.FiscalizationService, *app.ReturnsService, *app.ReturnSettlementService, *app.CashService) {
	t.Helper()

	store := memory.NewStore(memory.WithProducts(testProduct()), memory.WithDemoActors())
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
		app.WithStoreOperations(store, store),
		app.WithClock(now),
		app.WithIDGenerator(newID),
	)
	cash := app.NewCashService(store, store)
	payments := app.NewPaymentService(store, store, store,
		app.WithPaymentCashLedger(store),
		app.WithPaymentClock(now),
		app.WithPaymentIDGenerator(newID),
	)
	fiscalization := app.NewFiscalizationService(store, store, store, store,
		app.WithFiscalizationClock(now),
		app.WithFiscalizationIDGenerator(newID),
	)
	auth := app.NewAuthService(store, store)
	returns := app.NewReturnsService(store, store, store, auth)
	settlement := app.NewReturnSettlementService(store, store, store, payments, store)

	return store, checkout, payments, fiscalization, returns, settlement, cash
}

func createFullReceiptReturn(t *testing.T, store *memory.Store, returns *app.ReturnsService, receiptID string) domain.Return {
	t.Helper()

	receipt, err := store.FindReceipt(context.Background(), receiptID)
	if err != nil {
		t.Fatalf("find receipt: %v", err)
	}
	lines := make([]app.ReturnLineCommand, 0, len(receipt.Lines))
	for _, line := range receipt.Lines {
		lines = append(lines, app.ReturnLineCommand{
			LineID:   line.ID,
			Quantity: line.Quantity,
		})
	}

	result, err := returns.CreateReceiptReturn(context.Background(), app.CreateReceiptReturnCommand{
		IdempotencyKey: "return-" + receiptID,
		ReceiptID:      receiptID,
		Lines:          lines,
		Reason:         "Customer return",
		ActorID:        "senior-1",
	})
	if err != nil {
		t.Fatalf("create receipt return: %v", err)
	}
	return result.Return
}

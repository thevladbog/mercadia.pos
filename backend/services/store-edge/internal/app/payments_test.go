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

func TestCreatePaymentCapturesAgainstReceiptRemainingAmount(t *testing.T) {
	checkout, payments := newTestCheckoutAndPaymentServices()
	receiptID := openAndScanTestReceipt(t, checkout)

	result, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCash,
		AmountMinor:    39998,
	})
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}

	if result.Payment.Status != domain.PaymentStatusCaptured {
		t.Fatalf("payment status = %s", result.Payment.Status)
	}
}

func TestCreatePartialPaymentLocksReceiptForLineChanges(t *testing.T) {
	checkout, payments := newTestCheckoutAndPaymentServices()
	receiptID := openAndScanTestReceipt(t, checkout)

	if _, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCash,
		AmountMinor:    10000,
	}); err != nil {
		t.Fatalf("create partial payment: %v", err)
	}

	_, err := checkout.ScanReceiptLine(context.Background(), app.ScanReceiptLineCommand{
		IdempotencyKey: "scan-after-payment",
		ReceiptID:      receiptID,
		Barcode:        "4600000000000",
		Quantity:       1,
	})
	if !errors.Is(err, domain.ErrReceiptClosed) {
		t.Fatalf("expected ErrReceiptClosed, got %v", err)
	}
}

func TestCreateFullPaymentMarksReceiptPaid(t *testing.T) {
	checkout, payments := newTestCheckoutAndPaymentServices()
	receiptID := openAndScanTestReceipt(t, checkout)

	if _, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCash,
		AmountMinor:    39998,
	}); err != nil {
		t.Fatalf("create payment: %v", err)
	}

	result, err := checkout.GetReceipt(context.Background(), receiptID)
	if err != nil {
		t.Fatalf("get receipt: %v", err)
	}
	if result.Receipt.Status != domain.ReceiptStatusPaid {
		t.Fatalf("receipt status = %s", result.Receipt.Status)
	}
}

func TestCreatePaymentIsIdempotent(t *testing.T) {
	checkout, payments := newTestCheckoutAndPaymentServices()
	receiptID := openAndScanTestReceipt(t, checkout)
	command := app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCash,
		AmountMinor:    39998,
	}

	first, err := payments.CreatePayment(context.Background(), command)
	if err != nil {
		t.Fatalf("create first payment: %v", err)
	}
	second, err := payments.CreatePayment(context.Background(), command)
	if err != nil {
		t.Fatalf("create second payment: %v", err)
	}

	if first.Payment.ID != second.Payment.ID {
		t.Fatalf("expected same payment id, got %s and %s", first.Payment.ID, second.Payment.ID)
	}
}

func TestCreatePaymentRejectsAmountAboveRemaining(t *testing.T) {
	checkout, payments := newTestCheckoutAndPaymentServices()
	receiptID := openAndScanTestReceipt(t, checkout)

	_, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCash,
		AmountMinor:    39999,
	})
	if !errors.Is(err, app.ErrPaymentAmountExceedsRemaining) {
		t.Fatalf("expected ErrPaymentAmountExceedsRemaining, got %v", err)
	}
}

func TestCancelPaidReceiptIsRejected(t *testing.T) {
	checkout, payments := newTestCheckoutAndPaymentServices()
	receiptID := openAndScanTestReceipt(t, checkout)

	if _, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCash,
		AmountMinor:    39998,
	}); err != nil {
		t.Fatalf("create payment: %v", err)
	}

	_, err := checkout.CancelReceipt(context.Background(), app.CancelReceiptCommand{
		IdempotencyKey: "cancel-1",
		ReceiptID:      receiptID,
		Reason:         "Customer changed mind",
		ActorID:        "cashier-1",
	})
	if !errors.Is(err, app.ErrReceiptCannotBeCancelled) {
		t.Fatalf("expected ErrReceiptCannotBeCancelled, got %v", err)
	}
}

func TestCreateCashPaymentPostsCashSaleMovement(t *testing.T) {
	store := memory.NewStore(memory.WithProducts(testProduct()))
	checkout, payments, cash := newTestCheckoutPaymentAndCashServices(store)
	receiptID := openOperationalReceiptAndScanTestProduct(t, store, checkout)

	if _, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCash,
		AmountMinor:    39998,
	}); err != nil {
		t.Fatalf("create cash payment: %v", err)
	}

	balances, err := cash.ListCashBalances(context.Background(), "store-1")
	if err != nil {
		t.Fatalf("list cash balances: %v", err)
	}
	if len(balances) != 1 {
		t.Fatalf("cash balances count = %d", len(balances))
	}
	if balances[0].ContainerID != "drawer-1" || balances[0].BalanceMinor != 39998 {
		t.Fatalf("cash balance = %+v", balances[0])
	}
}

func TestCreateCardPaymentDoesNotPostCashSaleMovement(t *testing.T) {
	store := memory.NewStore(memory.WithProducts(testProduct()))
	checkout, payments, cash := newTestCheckoutPaymentAndCashServices(store)
	receiptID := openOperationalReceiptAndScanTestProduct(t, store, checkout)

	if _, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCardMock,
		AmountMinor:    39998,
	}); err != nil {
		t.Fatalf("create card payment: %v", err)
	}

	balances, err := cash.ListCashBalances(context.Background(), "store-1")
	if err != nil {
		t.Fatalf("list cash balances: %v", err)
	}
	if len(balances) != 0 {
		t.Fatalf("cash balances count = %d", len(balances))
	}
}

type mockCardTerminal struct {
	cancelCalls int
	cancelErr   error
	refundCalls int
	refundErr   error
}

func (m *mockCardTerminal) AuthorizeAndCapture(context.Context, string, int64, string, string) (string, error) {
	return "RRN-123", nil
}

func (m *mockCardTerminal) CancelCardPayment(context.Context, string, string) error {
	m.cancelCalls++
	return m.cancelErr
}

func (m *mockCardTerminal) RefundCardPayment(context.Context, string, string, int64) error {
	m.refundCalls++
	return m.refundErr
}

func TestCancelCardPaymentReturnsReceiptToDraft(t *testing.T) {
	store := memory.NewStore(memory.WithProducts(testProduct()))
	checkout, payments := newTestCheckoutAndPaymentServicesWithStore(store)
	receiptID := openOperationalReceiptAndScanTestProduct(t, store, checkout)

	created, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCardMock,
		AmountMinor:    39998,
	})
	if err != nil {
		t.Fatalf("create card payment: %v", err)
	}

	cancelled, err := payments.CancelPayment(context.Background(), app.CancelPaymentCommand{
		IdempotencyKey: "cancel-1",
		ReceiptID:      receiptID,
		PaymentID:      created.Payment.ID,
		ActorID:        "cashier-1",
		Reason:         "Customer changed mind",
	})
	if err != nil {
		t.Fatalf("cancel payment: %v", err)
	}
	if cancelled.Payment.Status != domain.PaymentStatusCancelled {
		t.Fatalf("payment status = %s", cancelled.Payment.Status)
	}

	receipt, err := store.FindReceipt(context.Background(), receiptID)
	if err != nil {
		t.Fatalf("find receipt: %v", err)
	}
	if receipt.Status != domain.ReceiptStatusDraft {
		t.Fatalf("receipt status = %s", receipt.Status)
	}
}

func TestCancelCardPaymentCallsHardwareAgentTerminal(t *testing.T) {
	store := memory.NewStore(memory.WithProducts(testProduct()))
	terminal := &mockCardTerminal{}
	checkout, _ := newTestCheckoutAndPaymentServicesWithStore(store)
	payments := app.NewPaymentService(store, store, store,
		app.WithCardPaymentTerminal(terminal, "sim-payment-1", false),
		app.WithPaymentClock(func() time.Time {
			return time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
		}),
		app.WithPaymentIDGenerator(func(prefix string) string {
			return prefix + "-test-1"
		}),
	)
	receiptID := openOperationalReceiptAndScanTestProduct(t, store, checkout)

	created, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCardMock,
		AmountMinor:    39998,
	})
	if err != nil {
		t.Fatalf("create card payment: %v", err)
	}

	if _, err := payments.CancelPayment(context.Background(), app.CancelPaymentCommand{
		IdempotencyKey: "cancel-1",
		ReceiptID:      receiptID,
		PaymentID:      created.Payment.ID,
	}); err != nil {
		t.Fatalf("cancel payment: %v", err)
	}
	if terminal.cancelCalls != 1 {
		t.Fatalf("cancel calls = %d", terminal.cancelCalls)
	}
}

func TestCancelPaymentBlocksFiscalizedReceipt(t *testing.T) {
	store := memory.NewStore(memory.WithProducts(testProduct()))
	checkout, payments := newTestCheckoutAndPaymentServicesWithStore(store)
	receiptID := openOperationalReceiptAndScanTestProduct(t, store, checkout)

	created, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCardMock,
		AmountMinor:    39998,
	})
	if err != nil {
		t.Fatalf("create card payment: %v", err)
	}

	receipt, err := store.FindReceipt(context.Background(), receiptID)
	if err != nil {
		t.Fatalf("find receipt: %v", err)
	}
	if err := receipt.MarkFiscalized(time.Date(2026, 6, 18, 10, 1, 0, 0, time.UTC)); err != nil {
		t.Fatalf("mark fiscalized: %v", err)
	}
	if err := store.SaveReceipt(context.Background(), receipt); err != nil {
		t.Fatalf("save receipt: %v", err)
	}

	_, err = payments.CancelPayment(context.Background(), app.CancelPaymentCommand{
		IdempotencyKey: "cancel-1",
		ReceiptID:      receiptID,
		PaymentID:      created.Payment.ID,
	})
	if !errors.Is(err, app.ErrPaymentCannotBeCancelled) {
		t.Fatalf("expected ErrPaymentCannotBeCancelled, got %v", err)
	}
}

func TestCancelPaymentRequiresSameBusinessDate(t *testing.T) {
	store := memory.NewStore(memory.WithProducts(testProduct()))
	checkout := app.NewCheckoutService(store, store,
		app.WithProductRepository(store),
		app.WithStoreOperations(store, store),
		app.WithClock(func() time.Time {
			return time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
		}),
		app.WithIDGenerator(func(prefix string) string {
			return prefix + "-test-1"
		}),
	)
	payments := app.NewPaymentService(store, store, store,
		app.WithPaymentClock(func() time.Time {
			return time.Date(2026, 6, 19, 10, 0, 0, 0, time.UTC)
		}),
		app.WithPaymentIDGenerator(func(prefix string) string {
			return prefix + "-test-1"
		}),
	)
	receiptID := openOperationalReceiptAndScanTestProduct(t, store, checkout)

	created, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCardMock,
		AmountMinor:    39998,
	})
	if err != nil {
		t.Fatalf("create card payment: %v", err)
	}

	_, err = payments.CancelPayment(context.Background(), app.CancelPaymentCommand{
		IdempotencyKey: "cancel-1",
		ReceiptID:      receiptID,
		PaymentID:      created.Payment.ID,
	})
	if !errors.Is(err, app.ErrPaymentCancelSameDayRequired) {
		t.Fatalf("expected ErrPaymentCancelSameDayRequired, got %v", err)
	}
}

func TestCancelCashPaymentReturnsReceiptToDraft(t *testing.T) {
	store := memory.NewStore(memory.WithProducts(testProduct()))
	checkout, payments, cash := newTestCheckoutPaymentAndCashServices(store)
	receiptID := openOperationalReceiptAndScanTestProduct(t, store, checkout)

	balancesBefore, err := cash.ListCashBalances(context.Background(), "store-1")
	if err != nil {
		t.Fatalf("list cash balances before payment: %v", err)
	}
	drawerBefore := drawerBalanceMinor(balancesBefore, "drawer-1")

	created, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCash,
		AmountMinor:    39998,
	})
	if err != nil {
		t.Fatalf("create cash payment: %v", err)
	}

	balancesAfterPayment, err := cash.ListCashBalances(context.Background(), "store-1")
	if err != nil {
		t.Fatalf("list cash balances after payment: %v", err)
	}
	if drawerBalanceMinor(balancesAfterPayment, "drawer-1") != drawerBefore+39998 {
		t.Fatalf("drawer balance after payment = %d", drawerBalanceMinor(balancesAfterPayment, "drawer-1"))
	}

	cancelled, err := payments.CancelPayment(context.Background(), app.CancelPaymentCommand{
		IdempotencyKey: "cancel-1",
		ReceiptID:      receiptID,
		PaymentID:      created.Payment.ID,
		ActorID:        "cashier-1",
		Reason:         "Customer changed mind",
	})
	if err != nil {
		t.Fatalf("cancel cash payment: %v", err)
	}
	if cancelled.Payment.Status != domain.PaymentStatusCancelled {
		t.Fatalf("payment status = %s", cancelled.Payment.Status)
	}

	receipt, err := store.FindReceipt(context.Background(), receiptID)
	if err != nil {
		t.Fatalf("find receipt: %v", err)
	}
	if receipt.Status != domain.ReceiptStatusDraft {
		t.Fatalf("receipt status = %s", receipt.Status)
	}

	balancesAfterCancel, err := cash.ListCashBalances(context.Background(), "store-1")
	if err != nil {
		t.Fatalf("list cash balances after cancel: %v", err)
	}
	if drawerBalanceMinor(balancesAfterCancel, "drawer-1") != drawerBefore {
		t.Fatalf("drawer balance after cancel = %d want %d", drawerBalanceMinor(balancesAfterCancel, "drawer-1"), drawerBefore)
	}
}

func TestCancelCashPaymentPostsReversalMovement(t *testing.T) {
	store := memory.NewStore(memory.WithProducts(testProduct()))
	checkout, payments, cash := newTestCheckoutPaymentAndCashServices(store)
	receiptID := openOperationalReceiptAndScanTestProduct(t, store, checkout)

	created, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCash,
		AmountMinor:    39998,
	})
	if err != nil {
		t.Fatalf("create cash payment: %v", err)
	}

	if _, err := payments.CancelPayment(context.Background(), app.CancelPaymentCommand{
		IdempotencyKey: "cancel-1",
		ReceiptID:      receiptID,
		PaymentID:      created.Payment.ID,
	}); err != nil {
		t.Fatalf("cancel cash payment: %v", err)
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
			if movement.FromContainerID != "drawer-1" || movement.ToContainerID != "external-customer" {
				t.Fatalf("reversal movement containers = %+v", movement)
			}
		}
	}
	if saleCount != 1 || reversalCount != 1 {
		t.Fatalf("cash_sale=%d cash_sale_reversal=%d", saleCount, reversalCount)
	}
}

func drawerBalanceMinor(balances []domain.CashBalance, drawerID string) int64 {
	for _, balance := range balances {
		if balance.ContainerID == drawerID && balance.ContainerType == domain.CashContainerTypeDrawer {
			return balance.BalanceMinor
		}
	}
	return 0
}

func TestRefundCardPaymentKeepsFiscalizedReceipt(t *testing.T) {
	store := memory.NewStore(memory.WithProducts(testProduct()))
	checkout, payments := newTestCheckoutAndPaymentServicesWithStore(store)
	receiptID := openOperationalReceiptAndScanTestProduct(t, store, checkout)

	created, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCardMock,
		AmountMinor:    39998,
	})
	if err != nil {
		t.Fatalf("create card payment: %v", err)
	}

	receipt, err := store.FindReceipt(context.Background(), receiptID)
	if err != nil {
		t.Fatalf("find receipt: %v", err)
	}
	if err := receipt.MarkFiscalized(time.Date(2026, 6, 18, 10, 1, 0, 0, time.UTC)); err != nil {
		t.Fatalf("mark fiscalized: %v", err)
	}
	if err := store.SaveReceipt(context.Background(), receipt); err != nil {
		t.Fatalf("save receipt: %v", err)
	}

	refunded, err := payments.RefundPayment(context.Background(), app.RefundPaymentCommand{
		IdempotencyKey: "refund-1",
		ReceiptID:      receiptID,
		PaymentID:      created.Payment.ID,
		ActorID:        "cashier-1",
		Reason:         "Customer return",
	})
	if err != nil {
		t.Fatalf("refund payment: %v", err)
	}
	if refunded.Payment.Status != domain.PaymentStatusRefunded {
		t.Fatalf("payment status = %s", refunded.Payment.Status)
	}

	receipt, err = store.FindReceipt(context.Background(), receiptID)
	if err != nil {
		t.Fatalf("find receipt: %v", err)
	}
	if receipt.Status != domain.ReceiptStatusFiscalized {
		t.Fatalf("receipt status = %s", receipt.Status)
	}
}

func TestRefundPaymentBlocksSameDayPreFiscal(t *testing.T) {
	store := memory.NewStore(memory.WithProducts(testProduct()))
	checkout, payments := newTestCheckoutAndPaymentServicesWithStore(store)
	receiptID := openOperationalReceiptAndScanTestProduct(t, store, checkout)

	created, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCardMock,
		AmountMinor:    39998,
	})
	if err != nil {
		t.Fatalf("create card payment: %v", err)
	}

	_, err = payments.RefundPayment(context.Background(), app.RefundPaymentCommand{
		IdempotencyKey: "refund-1",
		ReceiptID:      receiptID,
		PaymentID:      created.Payment.ID,
	})
	if !errors.Is(err, app.ErrPaymentUseCancelInstead) {
		t.Fatalf("expected ErrPaymentUseCancelInstead, got %v", err)
	}
}

func TestRefundCardPaymentResyncsReceiptOnLaterBusinessDate(t *testing.T) {
	store := memory.NewStore(memory.WithProducts(testProduct()))
	checkout := app.NewCheckoutService(store, store,
		app.WithProductRepository(store),
		app.WithStoreOperations(store, store),
		app.WithClock(func() time.Time {
			return time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
		}),
		app.WithIDGenerator(func(prefix string) string {
			return prefix + "-test-1"
		}),
	)
	payments := app.NewPaymentService(store, store, store,
		app.WithPaymentClock(func() time.Time {
			return time.Date(2026, 6, 19, 10, 0, 0, 0, time.UTC)
		}),
		app.WithPaymentIDGenerator(func(prefix string) string {
			return prefix + "-test-1"
		}),
	)
	receiptID := openOperationalReceiptAndScanTestProduct(t, store, checkout)

	created, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCardMock,
		AmountMinor:    39998,
	})
	if err != nil {
		t.Fatalf("create card payment: %v", err)
	}

	if _, err := payments.RefundPayment(context.Background(), app.RefundPaymentCommand{
		IdempotencyKey: "refund-1",
		ReceiptID:      receiptID,
		PaymentID:      created.Payment.ID,
	}); err != nil {
		t.Fatalf("refund payment: %v", err)
	}

	receipt, err := store.FindReceipt(context.Background(), receiptID)
	if err != nil {
		t.Fatalf("find receipt: %v", err)
	}
	if receipt.Status != domain.ReceiptStatusDraft {
		t.Fatalf("receipt status = %s", receipt.Status)
	}
}

func TestRefundCardPaymentCallsHardwareAgentTerminal(t *testing.T) {
	store := memory.NewStore(memory.WithProducts(testProduct()))
	terminal := &mockCardTerminal{}
	checkout, _ := newTestCheckoutAndPaymentServicesWithStore(store)
	payments := app.NewPaymentService(store, store, store,
		app.WithCardPaymentTerminal(terminal, "sim-payment-1", false),
		app.WithPaymentClock(func() time.Time {
			return time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
		}),
		app.WithPaymentIDGenerator(func(prefix string) string {
			return prefix + "-test-1"
		}),
	)
	receiptID := openOperationalReceiptAndScanTestProduct(t, store, checkout)

	created, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCardMock,
		AmountMinor:    39998,
	})
	if err != nil {
		t.Fatalf("create card payment: %v", err)
	}

	receipt, err := store.FindReceipt(context.Background(), receiptID)
	if err != nil {
		t.Fatalf("find receipt: %v", err)
	}
	if err := receipt.MarkFiscalized(time.Date(2026, 6, 18, 10, 1, 0, 0, time.UTC)); err != nil {
		t.Fatalf("mark fiscalized: %v", err)
	}
	if err := store.SaveReceipt(context.Background(), receipt); err != nil {
		t.Fatalf("save receipt: %v", err)
	}

	if _, err := payments.RefundPayment(context.Background(), app.RefundPaymentCommand{
		IdempotencyKey: "refund-1",
		ReceiptID:      receiptID,
		PaymentID:      created.Payment.ID,
	}); err != nil {
		t.Fatalf("refund payment: %v", err)
	}
	if terminal.refundCalls != 1 {
		t.Fatalf("refund calls = %d", terminal.refundCalls)
	}
}

func TestRefundCashPaymentKeepsFiscalizedReceipt(t *testing.T) {
	store := memory.NewStore(memory.WithProducts(testProduct()))
	checkout, payments, cash := newTestCheckoutPaymentAndCashServices(store)
	receiptID := openOperationalReceiptAndScanTestProduct(t, store, checkout)

	balancesBefore, err := cash.ListCashBalances(context.Background(), "store-1")
	if err != nil {
		t.Fatalf("list cash balances before payment: %v", err)
	}
	drawerBefore := drawerBalanceMinor(balancesBefore, "drawer-1")

	created, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCash,
		AmountMinor:    39998,
	})
	if err != nil {
		t.Fatalf("create cash payment: %v", err)
	}

	balancesAfterPayment, err := cash.ListCashBalances(context.Background(), "store-1")
	if err != nil {
		t.Fatalf("list cash balances after payment: %v", err)
	}
	if drawerBalanceMinor(balancesAfterPayment, "drawer-1") != drawerBefore+39998 {
		t.Fatalf("drawer balance after payment = %d", drawerBalanceMinor(balancesAfterPayment, "drawer-1"))
	}

	receipt, err := store.FindReceipt(context.Background(), receiptID)
	if err != nil {
		t.Fatalf("find receipt: %v", err)
	}
	if err := receipt.MarkFiscalized(time.Date(2026, 6, 18, 10, 1, 0, 0, time.UTC)); err != nil {
		t.Fatalf("mark fiscalized: %v", err)
	}
	if err := store.SaveReceipt(context.Background(), receipt); err != nil {
		t.Fatalf("save receipt: %v", err)
	}

	refunded, err := payments.RefundPayment(context.Background(), app.RefundPaymentCommand{
		IdempotencyKey: "refund-1",
		ReceiptID:      receiptID,
		PaymentID:      created.Payment.ID,
		ActorID:        "cashier-1",
		Reason:         "Customer return",
	})
	if err != nil {
		t.Fatalf("refund cash payment: %v", err)
	}
	if refunded.Payment.Status != domain.PaymentStatusRefunded {
		t.Fatalf("payment status = %s", refunded.Payment.Status)
	}

	receipt, err = store.FindReceipt(context.Background(), receiptID)
	if err != nil {
		t.Fatalf("find receipt: %v", err)
	}
	if receipt.Status != domain.ReceiptStatusFiscalized {
		t.Fatalf("receipt status = %s", receipt.Status)
	}

	balancesAfterRefund, err := cash.ListCashBalances(context.Background(), "store-1")
	if err != nil {
		t.Fatalf("list cash balances after refund: %v", err)
	}
	if drawerBalanceMinor(balancesAfterRefund, "drawer-1") != drawerBefore {
		t.Fatalf("drawer balance after refund = %d want %d", drawerBalanceMinor(balancesAfterRefund, "drawer-1"), drawerBefore)
	}
}

func TestRefundCashPaymentBlocksSameDayPreFiscal(t *testing.T) {
	store := memory.NewStore(memory.WithProducts(testProduct()))
	checkout, payments, _ := newTestCheckoutPaymentAndCashServices(store)
	receiptID := openOperationalReceiptAndScanTestProduct(t, store, checkout)

	created, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCash,
		AmountMinor:    39998,
	})
	if err != nil {
		t.Fatalf("create cash payment: %v", err)
	}

	_, err = payments.RefundPayment(context.Background(), app.RefundPaymentCommand{
		IdempotencyKey: "refund-1",
		ReceiptID:      receiptID,
		PaymentID:      created.Payment.ID,
	})
	if !errors.Is(err, app.ErrPaymentUseCancelInstead) {
		t.Fatalf("expected ErrPaymentUseCancelInstead, got %v", err)
	}
}

func TestRefundCashPaymentResyncsReceiptOnLaterBusinessDate(t *testing.T) {
	store := memory.NewStore(memory.WithProducts(testProduct()))
	checkout := app.NewCheckoutService(store, store,
		app.WithProductRepository(store),
		app.WithStoreOperations(store, store),
		app.WithClock(func() time.Time {
			return time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
		}),
		app.WithIDGenerator(func(prefix string) string {
			return prefix + "-test-1"
		}),
	)
	payments := app.NewPaymentService(store, store, store,
		app.WithPaymentCashLedger(store),
		app.WithPaymentClock(func() time.Time {
			return time.Date(2026, 6, 19, 10, 0, 0, 0, time.UTC)
		}),
		app.WithPaymentIDGenerator(func(prefix string) string {
			return prefix + "-test-1"
		}),
	)
	receiptID := openOperationalReceiptAndScanTestProduct(t, store, checkout)

	created, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCash,
		AmountMinor:    39998,
	})
	if err != nil {
		t.Fatalf("create cash payment: %v", err)
	}

	if _, err := payments.RefundPayment(context.Background(), app.RefundPaymentCommand{
		IdempotencyKey: "refund-1",
		ReceiptID:      receiptID,
		PaymentID:      created.Payment.ID,
	}); err != nil {
		t.Fatalf("refund cash payment: %v", err)
	}

	receipt, err := store.FindReceipt(context.Background(), receiptID)
	if err != nil {
		t.Fatalf("find receipt: %v", err)
	}
	if receipt.Status != domain.ReceiptStatusDraft {
		t.Fatalf("receipt status = %s", receipt.Status)
	}
}

func TestRefundCashPaymentPostsReversalMovement(t *testing.T) {
	store := memory.NewStore(memory.WithProducts(testProduct()))
	checkout, payments, cash := newTestCheckoutPaymentAndCashServices(store)
	receiptID := openOperationalReceiptAndScanTestProduct(t, store, checkout)

	created, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCash,
		AmountMinor:    39998,
	})
	if err != nil {
		t.Fatalf("create cash payment: %v", err)
	}

	receipt, err := store.FindReceipt(context.Background(), receiptID)
	if err != nil {
		t.Fatalf("find receipt: %v", err)
	}
	if err := receipt.MarkFiscalized(time.Date(2026, 6, 18, 10, 1, 0, 0, time.UTC)); err != nil {
		t.Fatalf("mark fiscalized: %v", err)
	}
	if err := store.SaveReceipt(context.Background(), receipt); err != nil {
		t.Fatalf("save receipt: %v", err)
	}

	if _, err := payments.RefundPayment(context.Background(), app.RefundPaymentCommand{
		IdempotencyKey: "refund-1",
		ReceiptID:      receiptID,
		PaymentID:      created.Payment.ID,
	}); err != nil {
		t.Fatalf("refund cash payment: %v", err)
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

func TestPartialCashRefundLeavesPartiallyRefundedPayment(t *testing.T) {
	store := memory.NewStore(memory.WithProducts(testProduct()))
	checkout, payments, cash := newTestCheckoutPaymentAndCashServices(store)
	receiptID := openOperationalReceiptAndScanTestProduct(t, store, checkout)

	created, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCash,
		AmountMinor:    39998,
	})
	if err != nil {
		t.Fatalf("create cash payment: %v", err)
	}

	receipt, err := store.FindReceipt(context.Background(), receiptID)
	if err != nil {
		t.Fatalf("find receipt: %v", err)
	}
	if err := receipt.MarkFiscalized(time.Date(2026, 6, 18, 10, 1, 0, 0, time.UTC)); err != nil {
		t.Fatalf("mark fiscalized: %v", err)
	}
	if err := store.SaveReceipt(context.Background(), receipt); err != nil {
		t.Fatalf("save receipt: %v", err)
	}

	refunded, err := payments.RefundPayment(context.Background(), app.RefundPaymentCommand{
		IdempotencyKey: "refund-1",
		ReceiptID:      receiptID,
		PaymentID:      created.Payment.ID,
		AmountMinor:    19999,
	})
	if err != nil {
		t.Fatalf("partial refund: %v", err)
	}
	if refunded.Payment.Status != domain.PaymentStatusPartiallyRefunded {
		t.Fatalf("payment status = %s", refunded.Payment.Status)
	}
	if refunded.Payment.RefundedAmountMinor != 19999 {
		t.Fatalf("refunded amount = %d", refunded.Payment.RefundedAmountMinor)
	}

	movements, err := cash.ListCashMovements(context.Background(), "store-1", app.PageParams{Limit: 50})
	if err != nil {
		t.Fatalf("list cash movements: %v", err)
	}
	var reversalTotal int64
	for _, movement := range movements.Items {
		if movement.Type == domain.CashMovementTypeCashSaleReversal {
			reversalTotal += movement.AmountMinor
		}
	}
	if reversalTotal != 19999 {
		t.Fatalf("reversal total = %d", reversalTotal)
	}
}

func TestSecondPartialCashRefundCompletesPayment(t *testing.T) {
	store := memory.NewStore(memory.WithProducts(testProduct()))
	checkout, payments, _ := newTestCheckoutPaymentAndCashServices(store)
	receiptID := openOperationalReceiptAndScanTestProduct(t, store, checkout)

	created, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCash,
		AmountMinor:    39998,
	})
	if err != nil {
		t.Fatalf("create cash payment: %v", err)
	}

	receipt, err := store.FindReceipt(context.Background(), receiptID)
	if err != nil {
		t.Fatalf("find receipt: %v", err)
	}
	if err := receipt.MarkFiscalized(time.Date(2026, 6, 18, 10, 1, 0, 0, time.UTC)); err != nil {
		t.Fatalf("mark fiscalized: %v", err)
	}
	if err := store.SaveReceipt(context.Background(), receipt); err != nil {
		t.Fatalf("save receipt: %v", err)
	}

	if _, err := payments.RefundPayment(context.Background(), app.RefundPaymentCommand{
		IdempotencyKey: "refund-1",
		ReceiptID:      receiptID,
		PaymentID:      created.Payment.ID,
		AmountMinor:    19999,
	}); err != nil {
		t.Fatalf("first partial refund: %v", err)
	}

	refunded, err := payments.RefundPayment(context.Background(), app.RefundPaymentCommand{
		IdempotencyKey: "refund-2",
		ReceiptID:      receiptID,
		PaymentID:      created.Payment.ID,
		AmountMinor:    19999,
	})
	if err != nil {
		t.Fatalf("second partial refund: %v", err)
	}
	if refunded.Payment.Status != domain.PaymentStatusRefunded {
		t.Fatalf("payment status = %s", refunded.Payment.Status)
	}
	if refunded.Payment.RefundedAmountMinor != 39998 {
		t.Fatalf("refunded amount = %d", refunded.Payment.RefundedAmountMinor)
	}
}

func TestPartialCashRefundRejectsOverRefund(t *testing.T) {
	store := memory.NewStore(memory.WithProducts(testProduct()))
	checkout, payments, _ := newTestCheckoutPaymentAndCashServices(store)
	receiptID := openOperationalReceiptAndScanTestProduct(t, store, checkout)

	created, err := payments.CreatePayment(context.Background(), app.CreatePaymentCommand{
		IdempotencyKey: "payment-1",
		ReceiptID:      receiptID,
		Method:         domain.PaymentMethodCash,
		AmountMinor:    39998,
	})
	if err != nil {
		t.Fatalf("create cash payment: %v", err)
	}

	receipt, err := store.FindReceipt(context.Background(), receiptID)
	if err != nil {
		t.Fatalf("find receipt: %v", err)
	}
	if err := receipt.MarkFiscalized(time.Date(2026, 6, 18, 10, 1, 0, 0, time.UTC)); err != nil {
		t.Fatalf("mark fiscalized: %v", err)
	}
	if err := store.SaveReceipt(context.Background(), receipt); err != nil {
		t.Fatalf("save receipt: %v", err)
	}

	_, err = payments.RefundPayment(context.Background(), app.RefundPaymentCommand{
		IdempotencyKey: "refund-1",
		ReceiptID:      receiptID,
		PaymentID:      created.Payment.ID,
		AmountMinor:    40000,
	})
	if !errors.Is(err, app.ErrPaymentRefundAmountInvalid) {
		t.Fatalf("expected invalid refund amount, got %v", err)
	}
}

func newTestCheckoutAndPaymentServicesWithStore(store *memory.Store) (*app.CheckoutService, *app.PaymentService) {
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
	payments := app.NewPaymentService(store, store, store,
		app.WithPaymentClock(now),
		app.WithPaymentIDGenerator(newID),
	)
	return checkout, payments
}

func newTestCheckoutAndPaymentServices() (*app.CheckoutService, *app.PaymentService) {
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
	return checkout, payments
}

func newTestCheckoutPaymentAndCashServices(store *memory.Store) (*app.CheckoutService, *app.PaymentService, *app.CashService) {
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
	payments := app.NewPaymentService(store, store, store,
		app.WithPaymentCashLedger(store),
		app.WithPaymentClock(now),
		app.WithPaymentIDGenerator(newID),
	)
	cash := app.NewCashService(store, store,
		app.WithCashClock(now),
		app.WithCashIDGenerator(newID),
	)
	return checkout, payments, cash
}

func openOperationalReceiptAndScanTestProduct(t *testing.T, store *memory.Store, checkout *app.CheckoutService) string {
	t.Helper()

	days := app.NewOperationalDayService(store, store, store, store, store,
		app.WithOperationalDayClock(func() time.Time {
			return time.Date(2026, 6, 18, 9, 0, 0, 0, time.UTC)
		}),
		app.WithOperationalDayIDGenerator(func(prefix string) string {
			return prefix + "-test-oday"
		}),
	)
	if _, err := days.OpenOperationalDay(context.Background(), app.OpenOperationalDayCommand{
		IdempotencyKey: "oday-open-1",
		StoreID:        "store-1",
		BusinessDate:   "2026-06-18",
		OpenedByID:     "senior-1",
	}); err != nil {
		t.Fatalf("open operational day: %v", err)
	}

	shifts := app.NewShiftService(store, store,
		app.WithShiftClock(func() time.Time {
			return time.Date(2026, 6, 18, 9, 5, 0, 0, time.UTC)
		}),
		app.WithShiftIDGenerator(func(prefix string) string {
			return prefix + "-test-shift"
		}),
	)
	if _, err := shifts.OpenShift(context.Background(), app.OpenShiftCommand{
		IdempotencyKey:   "shift-open-1",
		StoreID:          "store-1",
		TerminalID:       "pos-1",
		CashierID:        "cashier-1",
		DrawerID:         "drawer-1",
		SourceSafeID:     "safe-1",
		OpeningCashMinor: 100000,
	}); err != nil {
		t.Fatalf("open shift: %v", err)
	}

	return openAndScanTestReceipt(t, checkout)
}

func openAndScanTestReceipt(t *testing.T, checkout *app.CheckoutService) string {
	t.Helper()

	opened, err := checkout.OpenReceipt(context.Background(), app.OpenReceiptCommand{
		IdempotencyKey: "open-1",
		StoreID:        "store-1",
		TerminalID:     "pos-1",
		CashierID:      "cashier-1",
	})
	if err != nil {
		t.Fatalf("open receipt: %v", err)
	}
	if _, err := checkout.ScanReceiptLine(context.Background(), app.ScanReceiptLineCommand{
		IdempotencyKey: "scan-1",
		ReceiptID:      opened.Receipt.ID,
		Barcode:        "4600000000000",
		Quantity:       2,
	}); err != nil {
		t.Fatalf("scan receipt line: %v", err)
	}
	return opened.Receipt.ID
}

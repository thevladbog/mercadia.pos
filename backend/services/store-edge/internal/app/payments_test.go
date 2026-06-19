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

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

func TestSettlePartialReturnRefundsCashPayment(t *testing.T) {
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

	ret := createPartialReceiptReturn(t, store, returns, receiptID, 1)

	result, err := settlement.SettleReturn(context.Background(), app.SettleReturnCommand{
		IdempotencyKey: "settle-1",
		ReturnID:       ret.ID,
	})
	if err != nil {
		t.Fatalf("settle partial return: %v", err)
	}
	if result.Return.Status != domain.ReturnStatusSettled {
		t.Fatalf("return status = %s", result.Return.Status)
	}
	if len(result.Payments) != 1 {
		t.Fatalf("payments count = %d", len(result.Payments))
	}
	if result.Payments[0].Status != domain.PaymentStatusPartiallyRefunded {
		t.Fatalf("payment status = %s", result.Payments[0].Status)
	}
	if result.Payments[0].RefundedAmountMinor != 19999 {
		t.Fatalf("refunded amount = %d", result.Payments[0].RefundedAmountMinor)
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

func TestSettlePartialReturnRefundsMixedPayments(t *testing.T) {
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

	ret := createPartialReceiptReturn(t, store, returns, receiptID, 1)

	result, err := settlement.SettleReturn(context.Background(), app.SettleReturnCommand{
		IdempotencyKey: "settle-1",
		ReturnID:       ret.ID,
	})
	if err != nil {
		t.Fatalf("settle partial return: %v", err)
	}
	if len(result.Payments) != 2 {
		t.Fatalf("payments count = %d", len(result.Payments))
	}

	var refundedTotal int64
	for _, payment := range result.Payments {
		if payment.Status != domain.PaymentStatusPartiallyRefunded {
			t.Fatalf("payment %s status = %s", payment.ID, payment.Status)
		}
		refundedTotal += payment.RefundedAmountMinor
	}
	if refundedTotal != 19999 {
		t.Fatalf("refunded total = %d", refundedTotal)
	}
}

func TestSettleReturnBlocksCumulativeTotalExceeded(t *testing.T) {
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

	first := createPartialReceiptReturn(t, store, returns, receiptID, 1)
	if _, err := settlement.SettleReturn(context.Background(), app.SettleReturnCommand{
		IdempotencyKey: "settle-1",
		ReturnID:       first.ID,
	}); err != nil {
		t.Fatalf("first settle: %v", err)
	}

	receipt, err := store.FindReceipt(context.Background(), receiptID)
	if err != nil {
		t.Fatalf("find receipt: %v", err)
	}
	second, err := returns.CreateReceiptReturn(context.Background(), app.CreateReceiptReturnCommand{
		IdempotencyKey: "return-second",
		ReceiptID:      receiptID,
		Lines:          []app.ReturnLineCommand{{LineID: receipt.Lines[0].ID, Quantity: 2}},
		Reason:         "Second return exceeds cumulative cap",
		ActorID:        "senior-1",
	})
	if err != nil {
		t.Fatalf("create second return: %v", err)
	}

	_, err = settlement.SettleReturn(context.Background(), app.SettleReturnCommand{
		IdempotencyKey: "settle-2",
		ReturnID:       second.Return.ID,
	})
	if !errors.Is(err, app.ErrReturnSettlementCumulativeTotalExceeded) {
		t.Fatalf("expected cumulative total exceeded, got %v", err)
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

func TestSettleNoReceiptReturnDisbursesCash(t *testing.T) {
	store := memory.NewStore(memory.WithDemoActors())
	auth := app.NewAuthService(store, store)
	returns := app.NewReturnsService(store, store, store, auth)
	cash := app.NewCashService(store, store)
	settlement := app.NewReturnSettlementService(store, store, store, nil, store,
		app.WithReturnSettlementCashLedger(store),
		app.WithReturnSettlementShiftLookup(store),
	)

	if _, err := cash.CreateCashMovement(context.Background(), app.CreateCashMovementCommand{
		IdempotencyKey:    "fund-1",
		StoreID:           "store-1",
		Type:              domain.CashMovementTypeChangeFund,
		FromContainerID:   "safe-1",
		FromContainerType: domain.CashContainerTypeSafe,
		ToContainerID:     "drawer-1",
		ToContainerType:   domain.CashContainerTypeDrawer,
		AmountMinor:       100000,
		Currency:          "RUB",
		Reason:            "Opening fund",
		ActorID:           "senior-1",
		ApprovedByID:      "admin-1",
	}); err != nil {
		t.Fatalf("fund drawer: %v", err)
	}

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

	result, err := settlement.SettleReturn(context.Background(), app.SettleReturnCommand{
		IdempotencyKey: "settle-1",
		ReturnID:       noReceipt.Return.ID,
		ActorID:        "senior-1",
		DrawerID:       "drawer-1",
	})
	if err != nil {
		t.Fatalf("settle no-receipt return: %v", err)
	}
	if result.Return.Status != domain.ReturnStatusSettled {
		t.Fatalf("return status = %s", result.Return.Status)
	}
	if len(result.Payments) != 0 {
		t.Fatalf("payments = %+v", result.Payments)
	}

	movements, err := cash.ListCashMovements(context.Background(), "store-1", app.PageParams{Limit: 50})
	if err != nil {
		t.Fatalf("list cash movements: %v", err)
	}
	var payoutCount int
	for _, movement := range movements.Items {
		if movement.Type == domain.CashMovementTypeNoReceiptReturnPayout {
			payoutCount++
			if movement.AmountMinor != 1000 || movement.ApprovedByID != "admin-1" {
				t.Fatalf("payout movement = %+v", movement)
			}
		}
	}
	if payoutCount != 1 {
		t.Fatalf("payout count = %d", payoutCount)
	}

	balances, err := cash.ListCashBalances(context.Background(), "store-1")
	if err != nil {
		t.Fatalf("list balances: %v", err)
	}
	if drawerBalanceMinor(balances, "drawer-1") != 99000 {
		t.Fatalf("drawer balance = %d", drawerBalanceMinor(balances, "drawer-1"))
	}
}

func TestSettleNoReceiptReturnBlocksApproverAsActor(t *testing.T) {
	store := memory.NewStore(memory.WithDemoActors())
	auth := app.NewAuthService(store, store)
	returns := app.NewReturnsService(store, store, store, auth)
	settlement := app.NewReturnSettlementService(store, store, store, nil, store,
		app.WithReturnSettlementCashLedger(store),
	)

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
		ActorID:        "admin-1",
		DrawerID:       "drawer-1",
	})
	if !errors.Is(err, app.ErrSeparationOfDutiesViolation) {
		t.Fatalf("expected separation of duties violation, got %v", err)
	}
}

func TestSettleNoReceiptReturnRequiresDrawer(t *testing.T) {
	store := memory.NewStore(memory.WithDemoActors())
	auth := app.NewAuthService(store, store)
	returns := app.NewReturnsService(store, store, store, auth)
	settlement := app.NewReturnSettlementService(store, store, store, nil, store,
		app.WithReturnSettlementCashLedger(store),
	)

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
		ActorID:        "senior-1",
	})
	if !errors.Is(err, app.ErrCashDrawerRequired) {
		t.Fatalf("expected cash drawer required, got %v", err)
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
	settlement := app.NewReturnSettlementService(store, store, store, payments, store,
		app.WithReturnSettlementCashLedger(store),
		app.WithReturnSettlementShiftLookup(store),
	)

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

func createPartialReceiptReturn(t *testing.T, store *memory.Store, returns *app.ReturnsService, receiptID string, quantity int64) domain.Return {
	t.Helper()

	receipt, err := store.FindReceipt(context.Background(), receiptID)
	if err != nil {
		t.Fatalf("find receipt: %v", err)
	}
	if len(receipt.Lines) != 1 {
		t.Fatalf("lines count = %d", len(receipt.Lines))
	}

	result, err := returns.CreateReceiptReturn(context.Background(), app.CreateReceiptReturnCommand{
		IdempotencyKey: fmt.Sprintf("return-partial-%s-%d", receiptID, quantity),
		ReceiptID:      receiptID,
		Lines:          []app.ReturnLineCommand{{LineID: receipt.Lines[0].ID, Quantity: quantity}},
		Reason:         "Partial return",
		ActorID:        "senior-1",
	})
	if err != nil {
		t.Fatalf("create partial return: %v", err)
	}
	return result.Return
}

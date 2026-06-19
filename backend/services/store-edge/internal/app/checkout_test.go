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

func TestOpenReceiptIsIdempotent(t *testing.T) {
	service := newTestCheckoutService()
	command := app.OpenReceiptCommand{
		IdempotencyKey: "open-1",
		StoreID:        "store-1",
		TerminalID:     "pos-1",
		CashierID:      "cashier-1",
		Channel:        "pos",
	}

	first, err := service.OpenReceipt(context.Background(), command)
	if err != nil {
		t.Fatalf("open first receipt: %v", err)
	}

	second, err := service.OpenReceipt(context.Background(), command)
	if err != nil {
		t.Fatalf("open second receipt: %v", err)
	}

	if first.Receipt.ID != second.Receipt.ID {
		t.Fatalf("expected same receipt id, got %q and %q", first.Receipt.ID, second.Receipt.ID)
	}
}

func TestAddReceiptLineUpdatesReceiptTotal(t *testing.T) {
	service := newTestCheckoutService()

	opened, err := service.OpenReceipt(context.Background(), app.OpenReceiptCommand{
		IdempotencyKey: "open-1",
		StoreID:        "store-1",
		TerminalID:     "pos-1",
		CashierID:      "cashier-1",
		Channel:        "pos",
	})
	if err != nil {
		t.Fatalf("open receipt: %v", err)
	}

	updated, err := service.AddReceiptLine(context.Background(), app.AddReceiptLineCommand{
		IdempotencyKey: "line-1",
		ReceiptID:      opened.Receipt.ID,
		ProductID:      "sku-1",
		Barcode:        "4600000000000",
		Name:           "Milk",
		Quantity:       2,
		UnitPriceMinor: 19999,
	})
	if err != nil {
		t.Fatalf("add receipt line: %v", err)
	}

	if got := updated.Receipt.TotalMinor(); got != 39998 {
		t.Fatalf("total minor = %d", got)
	}
}

func TestScanReceiptLineUsesCatalogProduct(t *testing.T) {
	store := memory.NewStore(memory.WithProducts(testProduct()))
	var counter int
	service := app.NewCheckoutService(store, store,
		app.WithProductRepository(store),
		app.WithClock(func() time.Time {
			return time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
		}),
		app.WithIDGenerator(func(prefix string) string {
			counter++
			return fmt.Sprintf("%s-test-%d", prefix, counter)
		}),
	)

	opened, err := service.OpenReceipt(context.Background(), app.OpenReceiptCommand{
		IdempotencyKey: "open-1",
		StoreID:        "store-1",
		TerminalID:     "pos-1",
		CashierID:      "cashier-1",
	})
	if err != nil {
		t.Fatalf("open receipt: %v", err)
	}

	updated, err := service.ScanReceiptLine(context.Background(), app.ScanReceiptLineCommand{
		IdempotencyKey: "scan-1",
		ReceiptID:      opened.Receipt.ID,
		Barcode:        "4600000000000",
		Quantity:       2,
	})
	if err != nil {
		t.Fatalf("scan receipt line: %v", err)
	}

	if got := updated.Receipt.TotalMinor(); got != 39998 {
		t.Fatalf("total minor = %d", got)
	}
	if got := updated.Receipt.Lines[0].ProductID; got != "sku-1" {
		t.Fatalf("product id = %s", got)
	}
}

func TestAddReceiptLineRejectsReusedIdempotencyKeyForDifferentReceipt(t *testing.T) {
	service := newTestCheckoutService()

	first, err := service.OpenReceipt(context.Background(), app.OpenReceiptCommand{
		IdempotencyKey: "open-1",
		StoreID:        "store-1",
		TerminalID:     "pos-1",
		CashierID:      "cashier-1",
	})
	if err != nil {
		t.Fatalf("open first receipt: %v", err)
	}
	second, err := service.OpenReceipt(context.Background(), app.OpenReceiptCommand{
		IdempotencyKey: "open-2",
		StoreID:        "store-1",
		TerminalID:     "pos-1",
		CashierID:      "cashier-1",
	})
	if err != nil {
		t.Fatalf("open second receipt: %v", err)
	}

	command := app.AddReceiptLineCommand{
		IdempotencyKey: "line-1",
		ReceiptID:      first.Receipt.ID,
		ProductID:      "sku-1",
		Name:           "Milk",
		Quantity:       1,
		UnitPriceMinor: 100,
	}
	if _, err := service.AddReceiptLine(context.Background(), command); err != nil {
		t.Fatalf("add first line: %v", err)
	}

	command.ReceiptID = second.Receipt.ID
	_, err = service.AddReceiptLine(context.Background(), command)
	if !errors.Is(err, app.ErrIdempotencyKeyReused) {
		t.Fatalf("expected ErrIdempotencyKeyReused, got %v", err)
	}
}

func TestOpenReceiptRejectsReusedIdempotencyKeyForDifferentCommand(t *testing.T) {
	service := newTestCheckoutService()

	command := app.OpenReceiptCommand{
		IdempotencyKey: "open-1",
		StoreID:        "store-1",
		TerminalID:     "pos-1",
		CashierID:      "cashier-1",
	}
	if _, err := service.OpenReceipt(context.Background(), command); err != nil {
		t.Fatalf("open receipt: %v", err)
	}

	command.TerminalID = "pos-2"
	_, err := service.OpenReceipt(context.Background(), command)
	if !errors.Is(err, app.ErrIdempotencyKeyReused) {
		t.Fatalf("expected ErrIdempotencyKeyReused, got %v", err)
	}
}

func TestCancelReceiptCancelsDraftReceipt(t *testing.T) {
	service := newTestCheckoutService()
	opened, err := service.OpenReceipt(context.Background(), app.OpenReceiptCommand{
		IdempotencyKey: "open-1",
		StoreID:        "store-1",
		TerminalID:     "pos-1",
		CashierID:      "cashier-1",
	})
	if err != nil {
		t.Fatalf("open receipt: %v", err)
	}

	cancelled, err := service.CancelReceipt(context.Background(), app.CancelReceiptCommand{
		IdempotencyKey: "cancel-1",
		ReceiptID:      opened.Receipt.ID,
		Reason:         "Customer changed mind",
		ActorID:        "cashier-1",
	})
	if err != nil {
		t.Fatalf("cancel receipt: %v", err)
	}
	if cancelled.Receipt.Status != domain.ReceiptStatusCancelled ||
		cancelled.Receipt.CancelReason != "Customer changed mind" ||
		cancelled.Receipt.CancelledByID != "cashier-1" {
		t.Fatalf("cancelled receipt = %+v", cancelled.Receipt)
	}
}

func TestCancelReceiptIsIdempotent(t *testing.T) {
	service := newTestCheckoutService()
	opened, err := service.OpenReceipt(context.Background(), app.OpenReceiptCommand{
		IdempotencyKey: "open-1",
		StoreID:        "store-1",
		TerminalID:     "pos-1",
		CashierID:      "cashier-1",
	})
	if err != nil {
		t.Fatalf("open receipt: %v", err)
	}
	command := app.CancelReceiptCommand{
		IdempotencyKey: "cancel-1",
		ReceiptID:      opened.Receipt.ID,
		Reason:         "Customer changed mind",
		ActorID:        "cashier-1",
	}

	first, err := service.CancelReceipt(context.Background(), command)
	if err != nil {
		t.Fatalf("cancel first receipt: %v", err)
	}
	second, err := service.CancelReceipt(context.Background(), command)
	if err != nil {
		t.Fatalf("cancel second receipt: %v", err)
	}
	if first.Receipt.ID != second.Receipt.ID || second.Receipt.Status != domain.ReceiptStatusCancelled {
		t.Fatalf("idempotent cancel result = %+v", second.Receipt)
	}
}

func TestListReceiptsByShift(t *testing.T) {
	store := memory.NewStore()
	service := app.NewCheckoutService(store, store)
	receipt, err := domain.NewReceipt(domain.NewReceiptInput{
		ID:         "receipt-1",
		StoreID:    "store-1",
		ShiftID:    "shift-1",
		TerminalID: "pos-1",
		CashierID:  "cashier-1",
		Now:        time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("create receipt: %v", err)
	}
	if err := store.SaveReceipt(context.Background(), receipt); err != nil {
		t.Fatalf("save receipt: %v", err)
	}

	receipts, err := service.ListReceiptsByShift(context.Background(), "shift-1")
	if err != nil {
		t.Fatalf("list receipts by shift: %v", err)
	}
	if len(receipts) != 1 || receipts[0].ID != "receipt-1" {
		t.Fatalf("receipts = %+v", receipts)
	}
}

func TestListReceiptsByOperationalDay(t *testing.T) {
	store := memory.NewStore()
	service := app.NewCheckoutService(store, store)
	receipt, err := domain.NewReceipt(domain.NewReceiptInput{
		ID:               "receipt-1",
		StoreID:          "store-1",
		OperationalDayID: "oday-1",
		TerminalID:       "pos-1",
		CashierID:        "cashier-1",
		Now:              time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("create receipt: %v", err)
	}
	if err := store.SaveReceipt(context.Background(), receipt); err != nil {
		t.Fatalf("save receipt: %v", err)
	}

	receipts, err := service.ListReceiptsByOperationalDay(context.Background(), "oday-1", app.PageParams{Limit: 50})
	if err != nil {
		t.Fatalf("list receipts by operational day: %v", err)
	}
	if len(receipts.Items) != 1 || receipts.Items[0].ID != "receipt-1" {
		t.Fatalf("receipts = %+v", receipts.Items)
	}
}

func TestOpenReceiptRequiresOpenOperationalDayAndShiftWhenConfigured(t *testing.T) {
	store := memory.NewStore()
	checkout := app.NewCheckoutService(store, store,
		app.WithStoreOperations(store, store),
		app.WithClock(func() time.Time {
			return time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
		}),
		app.WithIDGenerator(func(prefix string) string {
			return prefix + "-test-1"
		}),
	)

	command := app.OpenReceiptCommand{
		IdempotencyKey: "open-1",
		StoreID:        "store-1",
		TerminalID:     "pos-1",
		CashierID:      "cashier-1",
	}
	if _, err := checkout.OpenReceipt(context.Background(), command); !errors.Is(err, app.ErrOpenShiftRequired) {
		t.Fatalf("expected ErrOpenShiftRequired, got %v", err)
	}

	days := app.NewOperationalDayService(store, store, store, store, store,
		app.WithOperationalDayClock(func() time.Time {
			return time.Date(2026, 6, 18, 9, 0, 0, 0, time.UTC)
		}),
		app.WithOperationalDayIDGenerator(func(prefix string) string {
			return prefix + "-test-1"
		}),
	)
	openedDay, err := days.OpenOperationalDay(context.Background(), app.OpenOperationalDayCommand{
		IdempotencyKey: "oday-open-1",
		StoreID:        "store-1",
		BusinessDate:   "2026-06-18",
		OpenedByID:     "senior-1",
	})
	if err != nil {
		t.Fatalf("open operational day: %v", err)
	}

	shifts := app.NewShiftService(store, store,
		app.WithShiftClock(func() time.Time {
			return time.Date(2026, 6, 18, 9, 5, 0, 0, time.UTC)
		}),
		app.WithShiftIDGenerator(func(prefix string) string {
			return prefix + "-test-1"
		}),
	)
	openedShift, err := shifts.OpenShift(context.Background(), app.OpenShiftCommand{
		IdempotencyKey:   "shift-open-1",
		StoreID:          "store-1",
		TerminalID:       "pos-1",
		CashierID:        "cashier-1",
		DrawerID:         "drawer-1",
		OpeningCashMinor: 100000,
	})
	if err != nil {
		t.Fatalf("open shift: %v", err)
	}

	openedReceipt, err := checkout.OpenReceipt(context.Background(), command)
	if err != nil {
		t.Fatalf("open receipt: %v", err)
	}
	if openedReceipt.Receipt.OperationalDayID != openedDay.Day.ID ||
		openedReceipt.Receipt.BusinessDate != "2026-06-18" ||
		openedReceipt.Receipt.ShiftID != openedShift.Shift.ID ||
		openedReceipt.Receipt.DrawerID != "drawer-1" {
		t.Fatalf("receipt links = %+v", openedReceipt.Receipt)
	}
	if openedReceipt.Receipt.Status != domain.ReceiptStatusDraft {
		t.Fatalf("receipt status = %s", openedReceipt.Receipt.Status)
	}
}

func newTestCheckoutService() *app.CheckoutService {
	store := memory.NewStore()
	var counter int
	return app.NewCheckoutService(store, store,
		app.WithClock(func() time.Time {
			return time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
		}),
		app.WithIDGenerator(func(prefix string) string {
			counter++
			return fmt.Sprintf("%s-test-%d", prefix, counter)
		}),
	)
}

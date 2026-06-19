package postgres_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"mercadia.dev/pos/services/store-edge/internal/app"
	"mercadia.dev/pos/services/store-edge/internal/domain"
	"mercadia.dev/pos/services/store-edge/internal/infra/postgres"
)

func testDatabaseURL(t *testing.T) string {
	t.Helper()

	for _, key := range []string{"MERCADIA_STORE_EDGE_DATABASE_URL", "DATABASE_URL"} {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}

	t.Skip("MERCADIA_STORE_EDGE_DATABASE_URL or DATABASE_URL is not set")
	return ""
}

func newTestStore(t *testing.T) *postgres.Store {
	t.Helper()

	ctx := context.Background()
	databaseURL := testDatabaseURL(t)

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}
	t.Cleanup(pool.Close)

	if _, err := postgres.RunMigrations(ctx, pool, postgres.DefaultMigrationsDir()); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	store, err := postgres.NewStore(ctx, databaseURL)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	t.Cleanup(store.Close)

	_, err = store.Pool().Exec(ctx, `
		TRUNCATE TABLE
			outbox_events,
			operation_journal_entries,
			products,
			terminals,
			cash_recounts,
			cash_movements,
			fiscal_documents,
			payments,
			receipts,
			shifts,
			operational_days,
			idempotency_records
		RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatalf("truncate tables: %v", err)
	}

	return store
}

func TestStoreOperationalDayReceiptFlow(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	now := time.Date(2026, 6, 19, 10, 0, 0, 0, time.UTC)

	day, err := domain.OpenOperationalDay(domain.OpenOperationalDayInput{
		ID:           "day_test_1",
		StoreID:      "store_test",
		BusinessDate: "2026-06-19",
		OpenedByID:   "manager_1",
		Now:          now,
	})
	if err != nil {
		t.Fatalf("open operational day: %v", err)
	}
	if err := store.SaveOperationalDay(ctx, day); err != nil {
		t.Fatalf("save operational day: %v", err)
	}

	shift, err := domain.OpenShift(domain.OpenShiftInput{
		ID:               "shift_test_1",
		StoreID:          "store_test",
		OperationalDayID: day.ID,
		BusinessDate:     day.BusinessDate,
		TerminalID:       "terminal_1",
		CashierID:        "cashier_1",
		DrawerID:         "drawer_1",
		OpeningCashMinor: 10000,
		Now:              now,
	})
	if err != nil {
		t.Fatalf("open shift: %v", err)
	}
	if err := store.SaveShift(ctx, shift); err != nil {
		t.Fatalf("save shift: %v", err)
	}

	receipt, err := domain.NewReceipt(domain.NewReceiptInput{
		ID:               "receipt_test_1",
		StoreID:          "store_test",
		OperationalDayID: day.ID,
		BusinessDate:     day.BusinessDate,
		ShiftID:          shift.ID,
		TerminalID:       shift.TerminalID,
		CashierID:        shift.CashierID,
		DrawerID:         shift.DrawerID,
		Now:              now,
	})
	if err != nil {
		t.Fatalf("new receipt: %v", err)
	}
	if err := store.SaveReceipt(ctx, receipt); err != nil {
		t.Fatalf("save receipt: %v", err)
	}

	foundDay, err := store.FindOpenOperationalDayByStore(ctx, "store_test")
	if err != nil {
		t.Fatalf("find open operational day: %v", err)
	}
	if foundDay.ID != day.ID {
		t.Fatalf("expected day %s, got %s", day.ID, foundDay.ID)
	}

	foundShift, err := store.FindOpenShiftByTerminal(ctx, "terminal_1")
	if err != nil {
		t.Fatalf("find open shift: %v", err)
	}
	if foundShift.ID != shift.ID {
		t.Fatalf("expected shift %s, got %s", shift.ID, foundShift.ID)
	}

	foundReceipt, err := store.FindReceipt(ctx, receipt.ID)
	if err != nil {
		t.Fatalf("find receipt: %v", err)
	}
	if foundReceipt.StoreID != receipt.StoreID {
		t.Fatalf("expected store %s, got %s", receipt.StoreID, foundReceipt.StoreID)
	}

	if err := store.Save(ctx, app.IdempotencyRecord{
		Operation:   "checkout.open_receipt",
		Key:         "idem_1",
		TargetID:    receipt.ID,
		Fingerprint: "fp_1",
		Result:      app.ReceiptResult{Receipt: foundReceipt},
		CreatedAt:   now,
	}); err != nil {
		t.Fatalf("save idempotency: %v", err)
	}

	record, found, err := store.Find(ctx, "checkout.open_receipt", "idem_1")
	if err != nil || !found {
		t.Fatalf("find idempotency: found=%v err=%v", found, err)
	}
	result, ok := record.Result.(app.ReceiptResult)
	if !ok {
		t.Fatalf("expected ReceiptResult, got %T", record.Result)
	}
	if result.Receipt.ID != receipt.ID {
		t.Fatalf("expected receipt %s, got %s", receipt.ID, result.Receipt.ID)
	}
}

func TestStoreFindProductByBarcode(t *testing.T) {
	ctx := context.Background()
	databaseURL := testDatabaseURL(t)

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}
	t.Cleanup(pool.Close)

	if _, err := postgres.RunMigrations(ctx, pool, postgres.DefaultMigrationsDir()); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	product, err := domain.NewProduct(domain.Product{
		ID:             "product_test_1",
		Name:           "Integration Milk",
		Barcodes:       []string{"9990001112223"},
		UnitPriceMinor: 25000,
		TaxCategoryID:  "vat_20",
	})
	if err != nil {
		t.Fatalf("new product: %v", err)
	}

	storeWithProduct, err := postgres.NewStore(ctx, databaseURL, postgres.WithProducts(product))
	if err != nil {
		t.Fatalf("new store with products: %v", err)
	}
	defer storeWithProduct.Close()

	found, err := storeWithProduct.FindProductByBarcode(ctx, "9990001112223")
	if err != nil {
		t.Fatalf("find product by barcode: %v", err)
	}
	if found.ID != product.ID {
		t.Fatalf("expected product %s, got %s", product.ID, found.ID)
	}
}

func TestStorePing(t *testing.T) {
	store := newTestStore(t)
	if err := store.Ping(context.Background()); err != nil {
		t.Fatalf("ping postgres: %v", err)
	}
}

func TestRunRollsBackBusinessWriteWhenOutboxFails(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	now := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)

	receipt, err := domain.NewReceipt(domain.NewReceiptInput{
		ID:           "receipt_tx_rollback",
		StoreID:      "store_test",
		BusinessDate: "2026-06-19",
		CashierID:    "cashier_1",
		Now:          now,
	})
	if err != nil {
		t.Fatalf("new receipt: %v", err)
	}
	if err := store.SaveReceipt(ctx, receipt); err != nil {
		t.Fatalf("save receipt: %v", err)
	}

	payment, err := domain.CreateCapturedPayment(domain.CreateCapturedPaymentInput{
		ID:          "payment_tx_rollback",
		ReceiptID:   receipt.ID,
		Method:      domain.PaymentMethodCash,
		AmountMinor: 5000,
		Now:         now,
	})
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}

	err = store.Run(ctx, func(txCtx context.Context) error {
		if err := store.SavePayment(txCtx, payment); err != nil {
			return err
		}
		return errors.New("simulated outbox failure")
	})
	if err == nil {
		t.Fatal("expected transaction error")
	}

	if _, err := store.FindPayment(ctx, payment.ID); err == nil {
		t.Fatal("expected payment to be rolled back")
	}

	pending, published, err := store.CountOutboxEvents(ctx)
	if err != nil {
		t.Fatalf("count outbox events: %v", err)
	}
	if pending+published != 0 {
		t.Fatalf("expected no outbox rows, got pending=%d published=%d", pending, published)
	}
}

func TestPaymentCapturePersistsPaymentAndOutboxAtomically(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	now := time.Date(2026, 6, 19, 12, 30, 0, 0, time.UTC)

	receipt, err := domain.NewReceipt(domain.NewReceiptInput{
		ID:           "receipt_tx_atomic",
		StoreID:      "store_test",
		BusinessDate: "2026-06-19",
		CashierID:    "cashier_1",
		Now:          now,
	})
	if err != nil {
		t.Fatalf("new receipt: %v", err)
	}
	if err := store.SaveReceipt(ctx, receipt); err != nil {
		t.Fatalf("save receipt: %v", err)
	}

	payment, err := domain.CreateCapturedPayment(domain.CreateCapturedPaymentInput{
		ID:          "payment_tx_atomic",
		ReceiptID:   receipt.ID,
		Method:      domain.PaymentMethodCash,
		AmountMinor: 7500,
		Now:         now,
	})
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}

	outboxEvent, err := domain.NewOutboxEvent(domain.OutboxEvent{
		ID:            "outbox_tx_atomic",
		AggregateType: domain.OutboxAggregatePayment,
		AggregateID:   payment.ID,
		EventType:     domain.OutboxEventPaymentCaptured,
		CreatedAt:     now,
	})
	if err != nil {
		t.Fatalf("new outbox event: %v", err)
	}

	if err := store.Run(ctx, func(txCtx context.Context) error {
		if err := store.SavePayment(txCtx, payment); err != nil {
			return err
		}
		return store.SaveOutboxEvent(txCtx, outboxEvent)
	}); err != nil {
		t.Fatalf("run transaction: %v", err)
	}

	foundPayment, err := store.FindPayment(ctx, payment.ID)
	if err != nil {
		t.Fatalf("find payment: %v", err)
	}
	if foundPayment.AmountMinor != payment.AmountMinor {
		t.Fatalf("expected amount %d, got %d", payment.AmountMinor, foundPayment.AmountMinor)
	}

	pending, published, err := store.CountOutboxEvents(ctx)
	if err != nil {
		t.Fatalf("count outbox events: %v", err)
	}
	if pending != 1 || published != 0 {
		t.Fatalf("expected one pending outbox event, got pending=%d published=%d", pending, published)
	}
}

func TestRunRollsBackJournalWhenCallbackFails(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	now := time.Date(2026, 6, 19, 13, 0, 0, 0, time.UTC)

	entry, err := domain.NewOperationJournalEntry(domain.CreateOperationJournalEntryInput{
		ID:            "oj_tx_rollback",
		StoreID:       "store_test",
		OperationType: "test.journal.rollback",
		ActorID:       "cashier_1",
		ReferenceID:   "ref_1",
		Summary:       "simulated failure",
		Now:           now,
	})
	if err != nil {
		t.Fatalf("new journal entry: %v", err)
	}

	err = store.Run(ctx, func(txCtx context.Context) error {
		if err := store.SaveOperationJournalEntry(txCtx, entry); err != nil {
			return err
		}
		return errors.New("simulated idempotency failure")
	})
	if err == nil {
		t.Fatal("expected transaction error")
	}

	result, err := store.ListOperationJournalEntries(ctx, "store_test", app.PageParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("list journal entries: %v", err)
	}
	if result.TotalCount != 0 || len(result.Items) != 0 {
		t.Fatalf("expected no journal rows, got total=%d items=%d", result.TotalCount, len(result.Items))
	}
}

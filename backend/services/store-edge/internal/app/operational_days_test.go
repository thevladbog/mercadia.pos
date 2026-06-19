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

func TestOpenOperationalDayIsIdempotent(t *testing.T) {
	service, _ := newTestOperationalDayService()
	command := testOpenOperationalDayCommand()

	first, err := service.OpenOperationalDay(context.Background(), command)
	if err != nil {
		t.Fatalf("open first operational day: %v", err)
	}
	second, err := service.OpenOperationalDay(context.Background(), command)
	if err != nil {
		t.Fatalf("open second operational day: %v", err)
	}

	if first.Day.ID != second.Day.ID {
		t.Fatalf("expected same operational day id, got %s and %s", first.Day.ID, second.Day.ID)
	}
}

func TestOpenOperationalDayRejectsExistingOpenDay(t *testing.T) {
	service, _ := newTestOperationalDayService()
	if _, err := service.OpenOperationalDay(context.Background(), testOpenOperationalDayCommand()); err != nil {
		t.Fatalf("open first operational day: %v", err)
	}

	command := testOpenOperationalDayCommand()
	command.IdempotencyKey = "oday-open-2"
	command.BusinessDate = "2026-06-19"
	_, err := service.OpenOperationalDay(context.Background(), command)
	if !errors.Is(err, app.ErrOperationalDayAlreadyOpen) {
		t.Fatalf("expected ErrOperationalDayAlreadyOpen, got %v", err)
	}
}

func TestOperationalDayCloseReadinessIncludesOpenShiftAndNoSales(t *testing.T) {
	service, store := newTestOperationalDayService()
	shifts := app.NewShiftService(store, store,
		app.WithShiftClock(func() time.Time {
			return time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
		}),
		app.WithShiftIDGenerator(func(prefix string) string {
			return prefix + "-test-1"
		}),
	)

	day, err := service.OpenOperationalDay(context.Background(), testOpenOperationalDayCommand())
	if err != nil {
		t.Fatalf("open operational day: %v", err)
	}
	if _, err := shifts.OpenShift(context.Background(), testOpenShiftCommand()); err != nil {
		t.Fatalf("open shift: %v", err)
	}

	readiness, err := service.CheckCloseReadiness(context.Background(), day.Day.ID)
	if err != nil {
		t.Fatalf("check close readiness: %v", err)
	}
	if readiness.CanClose {
		t.Fatal("expected close readiness to be blocked")
	}
	if len(readiness.Blockers) != 2 {
		t.Fatalf("blockers count = %d", len(readiness.Blockers))
	}
}

func TestCloseOperationalDayRequiresNoSalesOverride(t *testing.T) {
	service, _ := newTestOperationalDayService()
	day, err := service.OpenOperationalDay(context.Background(), testOpenOperationalDayCommand())
	if err != nil {
		t.Fatalf("open operational day: %v", err)
	}

	_, err = service.CloseOperationalDay(context.Background(), app.CloseOperationalDayCommand{
		IdempotencyKey: "oday-close-1",
		DayID:          day.Day.ID,
		ClosedByID:     "senior-1",
	})
	if !errors.Is(err, app.ErrOperationalDayCloseBlocked) {
		t.Fatalf("expected ErrOperationalDayCloseBlocked, got %v", err)
	}

	closed, err := service.CloseOperationalDay(context.Background(), app.CloseOperationalDayCommand{
		IdempotencyKey:  "oday-close-2",
		DayID:           day.Day.ID,
		ClosedByID:      "senior-1",
		OverrideNoSales: true,
		OverrideActorID: "admin-1",
	})
	if err != nil {
		t.Fatalf("close operational day with override: %v", err)
	}
	if closed.Day.Status != "closed" || closed.Day.ClosedByID != "senior-1" {
		t.Fatalf("closed operational day = %+v", closed.Day)
	}
}

func TestOperationalDayCloseReadinessBlocksUnresolvedCashDiscrepancy(t *testing.T) {
	service, store := newTestOperationalDayService()
	var cashCounter int
	cash := app.NewCashService(store, store,
		app.WithCashClock(func() time.Time {
			return time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
		}),
		app.WithCashIDGenerator(func(prefix string) string {
			cashCounter++
			return fmt.Sprintf("%s-test-%d", prefix, cashCounter)
		}),
	)

	day, err := service.OpenOperationalDay(context.Background(), testOpenOperationalDayCommand())
	if err != nil {
		t.Fatalf("open operational day: %v", err)
	}
	recount, err := cash.CreateCashRecount(context.Background(), app.CreateCashRecountCommand{
		IdempotencyKey: "recount-1",
		StoreID:        "store-1",
		ContainerID:    "safe-1",
		ContainerType:  domain.CashContainerTypeSafe,
		Currency:       "RUB",
		CountedMinor:   100000,
		Reason:         "Safe recount",
		ActorID:        "senior-1",
		ApprovedByID:   "cashier-1",
	})
	if err != nil {
		t.Fatalf("create cash recount: %v", err)
	}

	readiness, err := service.CheckCloseReadiness(context.Background(), day.Day.ID)
	if err != nil {
		t.Fatalf("check close readiness: %v", err)
	}
	if !hasBlocker(readiness.Blockers, "unresolved_cash_recount_discrepancy") {
		t.Fatalf("expected unresolved cash discrepancy blocker, got %+v", readiness.Blockers)
	}

	if _, err := cash.ResolveCashRecount(context.Background(), app.ResolveCashRecountCommand{
		IdempotencyKey: "recount-resolve-1",
		StoreID:        "store-1",
		RecountID:      recount.Recount.ID,
		ResolutionNote: "Adjustment movement posted",
		ActorID:        "senior-1",
		ApprovedByID:   "admin-1",
	}); err != nil {
		t.Fatalf("resolve cash recount: %v", err)
	}

	readiness, err = service.CheckCloseReadiness(context.Background(), day.Day.ID)
	if err != nil {
		t.Fatalf("check close readiness after resolve: %v", err)
	}
	if hasBlocker(readiness.Blockers, "unresolved_cash_recount_discrepancy") {
		t.Fatalf("unexpected unresolved cash discrepancy blocker after resolve: %+v", readiness.Blockers)
	}
}

func TestOperationalDayCloseReadinessBlocksUnresolvedReceipt(t *testing.T) {
	service, store := newTestOperationalDayService()
	now := time.Date(2026, 6, 18, 11, 0, 0, 0, time.UTC)

	day, err := service.OpenOperationalDay(context.Background(), testOpenOperationalDayCommand())
	if err != nil {
		t.Fatalf("open operational day: %v", err)
	}
	receipt, err := domain.NewReceipt(domain.NewReceiptInput{
		ID:               "receipt-unresolved-1",
		StoreID:          "store-1",
		OperationalDayID: day.Day.ID,
		BusinessDate:     "2026-06-18",
		ShiftID:          "shift-1",
		TerminalID:       "pos-1",
		CashierID:        "cashier-1",
		DrawerID:         "drawer-1",
		Now:              now,
	})
	if err != nil {
		t.Fatalf("create receipt: %v", err)
	}
	if err := store.SaveReceipt(context.Background(), receipt); err != nil {
		t.Fatalf("save receipt: %v", err)
	}

	readiness, err := service.CheckCloseReadiness(context.Background(), day.Day.ID)
	if err != nil {
		t.Fatalf("check close readiness: %v", err)
	}
	if !hasBlocker(readiness.Blockers, "unresolved_receipt") {
		t.Fatalf("expected unresolved receipt blocker, got %+v", readiness.Blockers)
	}

	if err := receipt.Cancel(domain.CancelReceiptInput{
		Reason:  "Customer changed mind",
		ActorID: "cashier-1",
		Now:     now.Add(5 * time.Minute),
	}); err != nil {
		t.Fatalf("cancel receipt: %v", err)
	}
	if err := store.SaveReceipt(context.Background(), receipt); err != nil {
		t.Fatalf("save cancelled receipt: %v", err)
	}

	readiness, err = service.CheckCloseReadiness(context.Background(), day.Day.ID)
	if err != nil {
		t.Fatalf("check close readiness after cancel: %v", err)
	}
	if hasBlocker(readiness.Blockers, "unresolved_receipt") {
		t.Fatalf("unexpected unresolved receipt blocker after cancel: %+v", readiness.Blockers)
	}
}

func TestOperationalDaySummaryAggregatesReceiptsAndBlockers(t *testing.T) {
	service, store := newTestOperationalDayService()
	now := time.Date(2026, 6, 18, 11, 0, 0, 0, time.UTC)

	day, err := service.OpenOperationalDay(context.Background(), testOpenOperationalDayCommand())
	if err != nil {
		t.Fatalf("open operational day: %v", err)
	}
	shift, err := domain.OpenShift(domain.OpenShiftInput{
		ID:               "shift-open-1",
		StoreID:          "store-1",
		OperationalDayID: day.Day.ID,
		BusinessDate:     "2026-06-18",
		TerminalID:       "pos-1",
		CashierID:        "cashier-1",
		DrawerID:         "drawer-1",
		OpeningCashMinor: 100000,
		Now:              now,
	})
	if err != nil {
		t.Fatalf("create shift: %v", err)
	}
	if err := store.SaveShift(context.Background(), shift); err != nil {
		t.Fatalf("save shift: %v", err)
	}
	cashMovement, err := domain.CreateCashMovement(domain.CreateCashMovementInput{
		ID:                "cash-sale-1",
		StoreID:           "store-1",
		Type:              domain.CashMovementTypeCashSale,
		FromContainerID:   "external-customer",
		FromContainerType: domain.CashContainerTypeExternal,
		ToContainerID:     "drawer-1",
		ToContainerType:   domain.CashContainerTypeDrawer,
		AmountMinor:       19999,
		Currency:          "RUB",
		Reason:            "Cash sale",
		ActorID:           "cashier-1",
		Now:               now,
	})
	if err != nil {
		t.Fatalf("create cash movement: %v", err)
	}
	if err := store.SaveCashMovement(context.Background(), cashMovement); err != nil {
		t.Fatalf("save cash movement: %v", err)
	}
	recount, err := domain.CreateCashRecount(domain.CreateCashRecountInput{
		ID:            "recount-1",
		StoreID:       "store-1",
		BusinessDate:  "2026-06-18",
		ContainerID:   "drawer-1",
		ContainerType: domain.CashContainerTypeDrawer,
		Currency:      "RUB",
		ExpectedMinor: 19999,
		CountedMinor:  15000,
		Reason:        "Drawer recount",
		ActorID:       "cashier-1",
		ApprovedByID:  "senior-1",
		Now:           now,
	})
	if err != nil {
		t.Fatalf("create cash recount: %v", err)
	}
	if err := store.SaveCashRecount(context.Background(), recount); err != nil {
		t.Fatalf("save cash recount: %v", err)
	}
	draftReceipt, err := domain.NewReceipt(domain.NewReceiptInput{
		ID:               "receipt-draft-1",
		StoreID:          "store-1",
		OperationalDayID: day.Day.ID,
		BusinessDate:     "2026-06-18",
		TerminalID:       "pos-1",
		CashierID:        "cashier-1",
		Now:              now,
	})
	if err != nil {
		t.Fatalf("create draft receipt: %v", err)
	}
	fiscalizedReceipt, err := domain.NewReceipt(domain.NewReceiptInput{
		ID:               "receipt-fiscalized-1",
		StoreID:          "store-1",
		OperationalDayID: day.Day.ID,
		BusinessDate:     "2026-06-18",
		TerminalID:       "pos-2",
		CashierID:        "cashier-2",
		Now:              now,
	})
	if err != nil {
		t.Fatalf("create fiscalized receipt: %v", err)
	}
	if err := fiscalizedReceipt.AddLine(domain.AddReceiptLineInput{
		ID:             "line-1",
		ProductID:      "sku-1",
		Name:           "Milk",
		Quantity:       1,
		UnitPriceMinor: 19999,
		Now:            now,
	}); err != nil {
		t.Fatalf("add line: %v", err)
	}
	if err := fiscalizedReceipt.MarkPaid(now); err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	if err := fiscalizedReceipt.MarkFiscalized(now); err != nil {
		t.Fatalf("mark fiscalized: %v", err)
	}
	if err := store.SaveReceipt(context.Background(), draftReceipt); err != nil {
		t.Fatalf("save draft receipt: %v", err)
	}
	if err := store.SaveReceipt(context.Background(), fiscalizedReceipt); err != nil {
		t.Fatalf("save fiscalized receipt: %v", err)
	}
	payment, err := domain.CreateCapturedPayment(domain.CreateCapturedPaymentInput{
		ID:          "payment-1",
		ReceiptID:   fiscalizedReceipt.ID,
		Method:      domain.PaymentMethodCardMock,
		AmountMinor: 19999,
		Now:         now,
	})
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}
	if err := store.SavePayment(context.Background(), payment); err != nil {
		t.Fatalf("save payment: %v", err)
	}
	fiscalDocument, err := domain.CreateFiscalizedDocument(domain.CreateFiscalizedDocumentInput{
		ID:          "fiscal-1",
		ReceiptID:   fiscalizedReceipt.ID,
		Kind:        domain.FiscalDocumentKindReceipt,
		AmountMinor: 19999,
		DeviceID:    "fiscal-device-1",
		FiscalSign:  "fiscal-sign-1",
		Now:         now,
	})
	if err != nil {
		t.Fatalf("create fiscal document: %v", err)
	}
	if err := store.SaveFiscalDocument(context.Background(), fiscalDocument); err != nil {
		t.Fatalf("save fiscal document: %v", err)
	}

	summary, err := service.GetOperationalDaySummary(context.Background(), day.Day.ID)
	if err != nil {
		t.Fatalf("get operational day summary: %v", err)
	}
	if summary.CanClose {
		t.Fatal("expected summary to be blocked")
	}
	if !hasBlocker(summary.Blockers, "unresolved_receipt") {
		t.Fatalf("expected unresolved receipt blocker, got %+v", summary.Blockers)
	}
	if summary.Shifts.TotalCount != 1 || summary.Shifts.OpenCount != 1 || summary.Shifts.ClosedCount != 0 {
		t.Fatalf("shift summary = %+v", summary.Shifts)
	}
	if len(summary.Cash.Balances) != 1 ||
		summary.Cash.Balances[0].ContainerID != "drawer-1" ||
		summary.Cash.Balances[0].BalanceMinor != 19999 ||
		summary.Cash.NonZeroDrawerCount != 1 {
		t.Fatalf("cash summary = %+v", summary.Cash)
	}
	if summary.Cash.Recounts.TotalCount != 1 ||
		summary.Cash.Recounts.DiscrepancyCount != 1 ||
		summary.Cash.Recounts.OpenDiscrepancyCount != 1 {
		t.Fatalf("cash recount summary = %+v", summary.Cash.Recounts)
	}
	if summary.Receipts.TotalCount != 2 ||
		summary.Receipts.DraftCount != 1 ||
		summary.Receipts.FiscalizedCount != 1 ||
		summary.Receipts.UnresolvedCount != 1 ||
		summary.Receipts.FiscalizedSalesMinor != 19999 {
		t.Fatalf("receipt summary = %+v", summary.Receipts)
	}
	if summary.Payments.TotalCount != 1 ||
		summary.Payments.CapturedCount != 1 ||
		summary.Payments.CapturedTotalMinor != 19999 ||
		len(summary.Payments.Methods) != 1 ||
		summary.Payments.Methods[0].Method != domain.PaymentMethodCardMock ||
		summary.Payments.Methods[0].CapturedTotalMinor != 19999 {
		t.Fatalf("payment summary = %+v", summary.Payments)
	}
	if summary.Fiscal.TotalCount != 1 ||
		summary.Fiscal.FiscalizedCount != 1 ||
		summary.Fiscal.FiscalizedTotalMinor != 19999 {
		t.Fatalf("fiscal summary = %+v", summary.Fiscal)
	}
}

func TestOperationalDayCloseReadinessBlocksNonZeroDrawerBalance(t *testing.T) {
	service, store := newTestOperationalDayService()
	var cashCounter int
	cash := app.NewCashService(store, store,
		app.WithCashClock(func() time.Time {
			return time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
		}),
		app.WithCashIDGenerator(func(prefix string) string {
			cashCounter++
			return fmt.Sprintf("%s-test-%d", prefix, cashCounter)
		}),
	)

	day, err := service.OpenOperationalDay(context.Background(), testOpenOperationalDayCommand())
	if err != nil {
		t.Fatalf("open operational day: %v", err)
	}
	if _, err := cash.CreateCashMovement(context.Background(), app.CreateCashMovementCommand{
		IdempotencyKey:    "cash-sale-1",
		StoreID:           "store-1",
		Type:              domain.CashMovementTypeCashSale,
		FromContainerID:   "external-customer",
		FromContainerType: domain.CashContainerTypeExternal,
		ToContainerID:     "drawer-1",
		ToContainerType:   domain.CashContainerTypeDrawer,
		AmountMinor:       50000,
		Currency:          "RUB",
		Reason:            "Cash sale",
		ActorID:           "cashier-1",
	}); err != nil {
		t.Fatalf("create cash sale movement: %v", err)
	}

	readiness, err := service.CheckCloseReadiness(context.Background(), day.Day.ID)
	if err != nil {
		t.Fatalf("check close readiness: %v", err)
	}
	if !hasBlocker(readiness.Blockers, "nonzero_drawer_balance") {
		t.Fatalf("expected nonzero drawer blocker, got %+v", readiness.Blockers)
	}

	if _, err := cash.CreateCashMovement(context.Background(), app.CreateCashMovementCommand{
		IdempotencyKey:    "drawer-to-safe-1",
		StoreID:           "store-1",
		Type:              domain.CashMovementTypeDrawerToSafe,
		FromContainerID:   "drawer-1",
		FromContainerType: domain.CashContainerTypeDrawer,
		ToContainerID:     "safe-1",
		ToContainerType:   domain.CashContainerTypeSafe,
		AmountMinor:       50000,
		Currency:          "RUB",
		Reason:            "Final collection",
		ActorID:           "cashier-1",
		ApprovedByID:      "senior-1",
	}); err != nil {
		t.Fatalf("create final collection movement: %v", err)
	}

	readiness, err = service.CheckCloseReadiness(context.Background(), day.Day.ID)
	if err != nil {
		t.Fatalf("check close readiness after final collection: %v", err)
	}
	if hasBlocker(readiness.Blockers, "nonzero_drawer_balance") {
		t.Fatalf("unexpected nonzero drawer blocker after final collection: %+v", readiness.Blockers)
	}
}

func newTestOperationalDayService() (*app.OperationalDayService, *memory.Store) {
	store := memory.NewStore()
	var counter int
	service := app.NewOperationalDayService(store, store, store, store, store,
		app.WithOperationalDayClock(func() time.Time {
			return time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
		}),
		app.WithOperationalDayIDGenerator(func(prefix string) string {
			counter++
			return fmt.Sprintf("%s-test-%d", prefix, counter)
		}),
	)
	return service, store
}

func hasBlocker(blockers []domain.OperationalDayBlocker, code string) bool {
	for _, blocker := range blockers {
		if blocker.Code == code {
			return true
		}
	}
	return false
}

func testOpenOperationalDayCommand() app.OpenOperationalDayCommand {
	return app.OpenOperationalDayCommand{
		IdempotencyKey: "oday-open-1",
		StoreID:        "store-1",
		BusinessDate:   "2026-06-18",
		OpenedByID:     "senior-1",
	}
}

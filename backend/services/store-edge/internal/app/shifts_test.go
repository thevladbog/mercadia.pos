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

func TestOpenShiftIsIdempotent(t *testing.T) {
	service := newTestShiftService()
	command := testOpenShiftCommand()

	first, err := service.OpenShift(context.Background(), command)
	if err != nil {
		t.Fatalf("open first shift: %v", err)
	}
	second, err := service.OpenShift(context.Background(), command)
	if err != nil {
		t.Fatalf("open second shift: %v", err)
	}

	if first.Shift.ID != second.Shift.ID {
		t.Fatalf("expected same shift id, got %s and %s", first.Shift.ID, second.Shift.ID)
	}
}

func TestOpenShiftRejectsExistingTerminalShift(t *testing.T) {
	service := newTestShiftService()
	if _, err := service.OpenShift(context.Background(), testOpenShiftCommand()); err != nil {
		t.Fatalf("open first shift: %v", err)
	}

	command := testOpenShiftCommand()
	command.IdempotencyKey = "shift-open-2"
	command.CashierID = "cashier-2"
	_, err := service.OpenShift(context.Background(), command)
	if !errors.Is(err, app.ErrShiftAlreadyOpenForTerminal) {
		t.Fatalf("expected ErrShiftAlreadyOpenForTerminal, got %v", err)
	}
}

func TestOpenShiftRejectsExistingCashierShift(t *testing.T) {
	service := newTestShiftService()
	if _, err := service.OpenShift(context.Background(), testOpenShiftCommand()); err != nil {
		t.Fatalf("open first shift: %v", err)
	}

	command := testOpenShiftCommand()
	command.IdempotencyKey = "shift-open-2"
	command.TerminalID = "pos-2"
	command.DrawerID = "drawer-2"
	_, err := service.OpenShift(context.Background(), command)
	if !errors.Is(err, app.ErrShiftAlreadyOpenForCashier) {
		t.Fatalf("expected ErrShiftAlreadyOpenForCashier, got %v", err)
	}
}

func TestOpenShiftRequiresConfiguredOperationalDay(t *testing.T) {
	store := memory.NewStore()
	var counter int
	service := app.NewShiftService(store, store,
		app.WithShiftOperationalDayRepository(store),
		app.WithShiftClock(func() time.Time {
			return time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
		}),
		app.WithShiftIDGenerator(func(prefix string) string {
			counter++
			return fmt.Sprintf("%s-test-%d", prefix, counter)
		}),
	)

	_, err := service.OpenShift(context.Background(), testOpenShiftCommand())
	if !errors.Is(err, app.ErrOpenOperationalDayRequired) {
		t.Fatalf("expected ErrOpenOperationalDayRequired, got %v", err)
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

	openedShift, err := service.OpenShift(context.Background(), testOpenShiftCommand())
	if err != nil {
		t.Fatalf("open shift: %v", err)
	}
	if openedShift.Shift.OperationalDayID != openedDay.Day.ID || openedShift.Shift.BusinessDate != "2026-06-18" {
		t.Fatalf("shift operational day links = %+v", openedShift.Shift)
	}
}

func TestCloseShiftRemovesItFromOpenList(t *testing.T) {
	service := newTestShiftService()
	opened, err := service.OpenShift(context.Background(), testOpenShiftCommand())
	if err != nil {
		t.Fatalf("open shift: %v", err)
	}

	closed, err := service.CloseShift(context.Background(), app.CloseShiftCommand{
		IdempotencyKey:   "shift-close-1",
		ShiftID:          opened.Shift.ID,
		ClosingCashMinor: 125000,
	})
	if err != nil {
		t.Fatalf("close shift: %v", err)
	}
	if closed.Shift.Status != "closed" || closed.Shift.ClosingCashMinor != 125000 {
		t.Fatalf("closed shift = %+v", closed.Shift)
	}

	openShifts, err := service.ListOpenShiftsByStore(context.Background(), "store-1")
	if err != nil {
		t.Fatalf("list open shifts: %v", err)
	}
	if len(openShifts) != 0 {
		t.Fatalf("open shifts count = %d", len(openShifts))
	}
}

func TestListShiftsByOperationalDay(t *testing.T) {
	store := memory.NewStore()
	service := app.NewShiftService(store, store)
	shift, err := domain.OpenShift(domain.OpenShiftInput{
		ID:               "shift-1",
		StoreID:          "store-1",
		OperationalDayID: "oday-1",
		BusinessDate:     "2026-06-18",
		TerminalID:       "pos-1",
		CashierID:        "cashier-1",
		DrawerID:         "drawer-1",
		OpeningCashMinor: 100000,
		Now:              time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("create shift: %v", err)
	}
	if err := store.SaveShift(context.Background(), shift); err != nil {
		t.Fatalf("save shift: %v", err)
	}

	shifts, err := service.ListShiftsByOperationalDay(context.Background(), "oday-1", app.PageParams{Limit: 50})
	if err != nil {
		t.Fatalf("list shifts by operational day: %v", err)
	}
	if len(shifts.Items) != 1 || shifts.Items[0].ID != "shift-1" {
		t.Fatalf("shifts = %+v", shifts.Items)
	}
}

func TestCloseShiftBlocksUnresolvedReceipt(t *testing.T) {
	store := memory.NewStore()
	var counter int
	service := app.NewShiftService(store, store,
		app.WithShiftReceiptRepository(store),
		app.WithShiftClock(func() time.Time {
			return time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
		}),
		app.WithShiftIDGenerator(func(prefix string) string {
			counter++
			return fmt.Sprintf("%s-test-%d", prefix, counter)
		}),
	)

	opened, err := service.OpenShift(context.Background(), testOpenShiftCommand())
	if err != nil {
		t.Fatalf("open shift: %v", err)
	}
	receipt, err := domain.NewReceipt(domain.NewReceiptInput{
		ID:         "receipt-unresolved-1",
		StoreID:    "store-1",
		ShiftID:    opened.Shift.ID,
		TerminalID: "pos-1",
		CashierID:  "cashier-1",
		DrawerID:   "drawer-1",
		Now:        time.Date(2026, 6, 18, 10, 30, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("create receipt: %v", err)
	}
	if err := store.SaveReceipt(context.Background(), receipt); err != nil {
		t.Fatalf("save receipt: %v", err)
	}

	_, err = service.CloseShift(context.Background(), app.CloseShiftCommand{
		IdempotencyKey:   "shift-close-1",
		ShiftID:          opened.Shift.ID,
		ClosingCashMinor: 0,
	})
	if !errors.Is(err, app.ErrShiftCloseBlocked) {
		t.Fatalf("expected ErrShiftCloseBlocked, got %v", err)
	}

	if err := receipt.Cancel(domain.CancelReceiptInput{
		Reason:  "Customer changed mind",
		ActorID: "cashier-1",
		Now:     time.Date(2026, 6, 18, 10, 35, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("cancel receipt: %v", err)
	}
	if err := store.SaveReceipt(context.Background(), receipt); err != nil {
		t.Fatalf("save cancelled receipt: %v", err)
	}

	closed, err := service.CloseShift(context.Background(), app.CloseShiftCommand{
		IdempotencyKey:   "shift-close-2",
		ShiftID:          opened.Shift.ID,
		ClosingCashMinor: 0,
	})
	if err != nil {
		t.Fatalf("close shift after cancel: %v", err)
	}
	if closed.Shift.Status != domain.ShiftStatusClosed {
		t.Fatalf("closed shift status = %s", closed.Shift.Status)
	}
}

func TestCloseShiftWithCashPostsFinalCollectionMovement(t *testing.T) {
	store := memory.NewStore()
	var counter int
	service := app.NewShiftService(store, store,
		app.WithShiftCashLedger(store),
		app.WithShiftClock(func() time.Time {
			return time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
		}),
		app.WithShiftIDGenerator(func(prefix string) string {
			counter++
			return fmt.Sprintf("%s-test-%d", prefix, counter)
		}),
	)
	cash := app.NewCashService(store, store)

	opened, err := service.OpenShift(context.Background(), testOpenShiftCommand())
	if err != nil {
		t.Fatalf("open shift: %v", err)
	}
	if _, err := service.CloseShift(context.Background(), app.CloseShiftCommand{
		IdempotencyKey:   "shift-close-1",
		ShiftID:          opened.Shift.ID,
		ClosingCashMinor: 125000,
		SafeID:           "safe-1",
		ActorID:          "cashier-1",
		ApprovedByID:     "senior-1",
	}); err != nil {
		t.Fatalf("close shift: %v", err)
	}

	balances, err := cash.ListCashBalances(context.Background(), "store-1")
	if err != nil {
		t.Fatalf("list balances: %v", err)
	}
	byContainer := map[string]int64{}
	for _, balance := range balances {
		byContainer[balance.ContainerID] = balance.BalanceMinor
	}
	if byContainer["drawer-1"] != -25000 {
		t.Fatalf("drawer balance = %d", byContainer["drawer-1"])
	}
	if byContainer["safe-1"] != 25000 {
		t.Fatalf("safe balance = %d", byContainer["safe-1"])
	}
}

func TestOpenShiftWithOpeningCashPostsChangeFund(t *testing.T) {
	store := memory.NewStore()
	var counter int
	service := app.NewShiftService(store, store,
		app.WithShiftCashLedger(store),
		app.WithShiftClock(func() time.Time {
			return time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
		}),
		app.WithShiftIDGenerator(func(prefix string) string {
			counter++
			return fmt.Sprintf("%s-test-%d", prefix, counter)
		}),
	)
	cash := app.NewCashService(store, store)

	command := testOpenShiftCommand()
	command.SourceSafeID = "safe-1"
	if _, err := service.OpenShift(context.Background(), command); err != nil {
		t.Fatalf("open shift: %v", err)
	}

	balances, err := cash.ListCashBalances(context.Background(), "store-1")
	if err != nil {
		t.Fatalf("list balances: %v", err)
	}
	byContainer := map[string]int64{}
	for _, balance := range balances {
		byContainer[balance.ContainerID] = balance.BalanceMinor
	}
	if byContainer["drawer-1"] != 100000 {
		t.Fatalf("drawer balance = %d", byContainer["drawer-1"])
	}
	if byContainer["safe-1"] != -100000 {
		t.Fatalf("safe balance = %d", byContainer["safe-1"])
	}
}

func TestOpenShiftRejectsOpeningCashWithoutSafe(t *testing.T) {
	service := newTestShiftService()
	command := testOpenShiftCommand()
	command.SourceSafeID = ""
	_, err := service.OpenShift(context.Background(), command)
	if !errors.Is(err, app.ErrShiftOpeningSafeRequired) {
		t.Fatalf("expected ErrShiftOpeningSafeRequired, got %v", err)
	}
}

func TestOpenCloseShiftRecordsOperationJournal(t *testing.T) {
	store := memory.NewStore()
	var counter int
	journal := app.NewOperationJournalService(store)
	service := app.NewShiftService(store, store,
		app.WithShiftCashLedger(store),
		app.WithShiftJournal(journal),
		app.WithShiftClock(func() time.Time {
			return time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
		}),
		app.WithShiftIDGenerator(func(prefix string) string {
			counter++
			return fmt.Sprintf("%s-test-%d", prefix, counter)
		}),
	)

	opened, err := service.OpenShift(context.Background(), testOpenShiftCommand())
	if err != nil {
		t.Fatalf("open shift: %v", err)
	}
	if _, err := service.CloseShift(context.Background(), app.CloseShiftCommand{
		IdempotencyKey:   "shift-close-1",
		ShiftID:          opened.Shift.ID,
		ClosingCashMinor: 125000,
		SafeID:           "safe-1",
		ActorID:          "cashier-1",
		ApprovedByID:     "senior-1",
	}); err != nil {
		t.Fatalf("close shift: %v", err)
	}

	entries, err := journal.ListOperationJournal(context.Background(), "store-1", app.PageParams{Limit: 50})
	if err != nil {
		t.Fatalf("list journal: %v", err)
	}
	if entries.TotalCount < 4 {
		t.Fatalf("expected at least 4 journal entries, got %+v", entries.Items)
	}

	types := map[string]int{}
	for _, entry := range entries.Items {
		types[entry.OperationType]++
	}
	if types["shift.opened"] != 1 || types["shift.closed"] != 1 {
		t.Fatalf("shift journal types = %+v", types)
	}
	if types["cash.movement.created"] < 2 {
		t.Fatalf("expected 2 cash movement journal entries, got %+v", types)
	}
}

func TestCloseShiftWithCashRejectsSelfApproval(t *testing.T) {
	store := memory.NewStore()
	service := app.NewShiftService(store, store, app.WithShiftCashLedger(store))

	opened, err := service.OpenShift(context.Background(), testOpenShiftCommand())
	if err != nil {
		t.Fatalf("open shift: %v", err)
	}
	_, err = service.CloseShift(context.Background(), app.CloseShiftCommand{
		IdempotencyKey:   "shift-close-1",
		ShiftID:          opened.Shift.ID,
		ClosingCashMinor: 125000,
		SafeID:           "safe-1",
		ActorID:          "cashier-1",
		ApprovedByID:     "cashier-1",
	})
	if !errors.Is(err, app.ErrSeparationOfDutiesViolation) {
		t.Fatalf("expected ErrSeparationOfDutiesViolation, got %v", err)
	}
}

func newTestShiftService() *app.ShiftService {
	store := memory.NewStore()
	var counter int
	return app.NewShiftService(store, store,
		app.WithShiftClock(func() time.Time {
			return time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
		}),
		app.WithShiftIDGenerator(func(prefix string) string {
			counter++
			return fmt.Sprintf("%s-test-%d", prefix, counter)
		}),
	)
}

func testOpenShiftCommand() app.OpenShiftCommand {
	return app.OpenShiftCommand{
		IdempotencyKey:   "shift-open-1",
		StoreID:          "store-1",
		TerminalID:       "pos-1",
		CashierID:        "cashier-1",
		DrawerID:         "drawer-1",
		SourceSafeID:     "safe-1",
		OpeningCashMinor: 100000,
	}
}

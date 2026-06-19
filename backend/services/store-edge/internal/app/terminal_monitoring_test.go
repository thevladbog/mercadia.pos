package app_test

import (
	"context"
	"testing"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/app"
	"mercadia.dev/pos/services/store-edge/internal/domain"
	"mercadia.dev/pos/services/store-edge/internal/infra/memory"
)

func TestListTerminalOverviewsIncludesShiftKPIs(t *testing.T) {
	store, monitoring := newTestTerminalMonitoringService()
	now := time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)

	if err := store.SaveTerminal(context.Background(), domain.Terminal{
		ID:              "pos-1",
		StoreID:         "store-1",
		Kind:            domain.TerminalKindPOS,
		Status:          domain.TerminalStatusOnline,
		SoftwareVersion: "0.1.0",
		LastSeenAt:      now,
		UpdatedAt:       now,
	}); err != nil {
		t.Fatalf("save terminal: %v", err)
	}

	shift, err := domain.OpenShift(domain.OpenShiftInput{
		ID:               "shift-1",
		StoreID:          "store-1",
		OperationalDayID: "oday-1",
		BusinessDate:     "2026-06-18",
		TerminalID:       "pos-1",
		CashierID:        "cashier-1",
		DrawerID:         "drawer-1",
		OpeningCashMinor: 100000,
		Now:              now,
	})
	if err != nil {
		t.Fatalf("open shift: %v", err)
	}
	if err := store.SaveShift(context.Background(), shift); err != nil {
		t.Fatalf("save shift: %v", err)
	}

	fiscalized, err := domain.NewReceipt(domain.NewReceiptInput{
		ID:               "receipt-1",
		StoreID:          "store-1",
		OperationalDayID: "oday-1",
		BusinessDate:     "2026-06-18",
		ShiftID:          "shift-1",
		TerminalID:       "pos-1",
		CashierID:        "cashier-1",
		DrawerID:         "drawer-1",
		Channel:          "pos",
		Now:              now,
	})
	if err != nil {
		t.Fatalf("new receipt: %v", err)
	}
	if err := fiscalized.AddLine(domain.AddReceiptLineInput{
		ID:             "line-1",
		ProductID:      "sku-1",
		Name:           "Milk",
		Quantity:       1,
		UnitPriceMinor: 50000,
		Now:              now,
	}); err != nil {
		t.Fatalf("add line: %v", err)
	}
	if err := fiscalized.MarkPaid(now); err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	if err := fiscalized.MarkFiscalized(now); err != nil {
		t.Fatalf("mark fiscalized: %v", err)
	}
	if err := store.SaveReceipt(context.Background(), fiscalized); err != nil {
		t.Fatalf("save fiscalized receipt: %v", err)
	}

	cash := app.NewCashService(store, store)
	if _, err := cash.CreateCashMovement(context.Background(), app.CreateCashMovementCommand{
		IdempotencyKey:    "cash-1",
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
		t.Fatalf("create cash movement: %v", err)
	}

	result, err := monitoring.ListTerminalOverviews(context.Background(), "store-1", app.PageParams{Limit: 50})
	if err != nil {
		t.Fatalf("list terminal overviews: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("items = %+v", result.Items)
	}

	overview := result.Items[0]
	if overview.ShiftID != "shift-1" || overview.CashierID != "cashier-1" || overview.DrawerID != "drawer-1" {
		t.Fatalf("shift fields = %+v", overview)
	}
	if overview.ReceiptCount != 1 || overview.RevenueMinor != 50000 {
		t.Fatalf("receipt KPIs = count %d revenue %d", overview.ReceiptCount, overview.RevenueMinor)
	}
	if overview.DrawerBalanceMinor != 50000 {
		t.Fatalf("drawer balance = %d", overview.DrawerBalanceMinor)
	}
}

func TestListTerminalOverviewsSetsCurrentReceipt(t *testing.T) {
	store, monitoring := newTestTerminalMonitoringService()
	now := time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)

	if err := store.SaveTerminal(context.Background(), domain.Terminal{
		ID:         "pos-1",
		StoreID:    "store-1",
		Kind:       domain.TerminalKindPOS,
		Status:     domain.TerminalStatusOnline,
		LastSeenAt: now,
		UpdatedAt:  now,
	}); err != nil {
		t.Fatalf("save terminal: %v", err)
	}

	shift, err := domain.OpenShift(domain.OpenShiftInput{
		ID:               "shift-1",
		StoreID:          "store-1",
		OperationalDayID: "oday-1",
		BusinessDate:     "2026-06-18",
		TerminalID:       "pos-1",
		CashierID:        "cashier-1",
		DrawerID:         "drawer-1",
		OpeningCashMinor: 0,
		Now:              now,
	})
	if err != nil {
		t.Fatalf("open shift: %v", err)
	}
	if err := store.SaveShift(context.Background(), shift); err != nil {
		t.Fatalf("save shift: %v", err)
	}

	draft, err := domain.NewReceipt(domain.NewReceiptInput{
		ID:               "receipt-draft",
		StoreID:          "store-1",
		OperationalDayID: "oday-1",
		BusinessDate:     "2026-06-18",
		ShiftID:          "shift-1",
		TerminalID:       "pos-1",
		CashierID:        "cashier-1",
		DrawerID:         "drawer-1",
		Channel:          "pos",
		Now:              now,
	})
	if err != nil {
		t.Fatalf("new draft receipt: %v", err)
	}
	if err := draft.AddLine(domain.AddReceiptLineInput{
		ID:             "line-1",
		ProductID:      "sku-1",
		Name:           "Milk",
		Quantity:       2,
		UnitPriceMinor: 10000,
		Now:              now,
	}); err != nil {
		t.Fatalf("add line: %v", err)
	}
	if err := store.SaveReceipt(context.Background(), draft); err != nil {
		t.Fatalf("save draft receipt: %v", err)
	}

	result, err := monitoring.ListTerminalOverviews(context.Background(), "store-1", app.PageParams{Limit: 50})
	if err != nil {
		t.Fatalf("list terminal overviews: %v", err)
	}
	overview := result.Items[0]
	if overview.CurrentReceiptID != "receipt-draft" || overview.CurrentReceiptStatus != domain.ReceiptStatusDraft {
		t.Fatalf("current receipt = %+v", overview)
	}
	if overview.CurrentReceiptTotalMinor != 20000 {
		t.Fatalf("current receipt total = %d", overview.CurrentReceiptTotalMinor)
	}
	if !overview.AttentionNeeded {
		t.Fatal("expected attention needed for draft receipt")
	}
}

func TestListTerminalOverviewsMarksOfflineAttention(t *testing.T) {
	store := memory.NewStore()
	monitoring := app.NewTerminalMonitoringService(store, store, store, store, app.NewCashService(store, store),
		app.WithTerminalMonitoringClock(func() time.Time {
			return time.Date(2026, 6, 18, 10, 2, 0, 0, time.UTC)
		}),
		app.WithTerminalMonitoringOfflineAfter(time.Minute),
	)

	lastSeen := time.Date(2026, 6, 18, 9, 0, 0, 0, time.UTC)
	if err := store.SaveTerminal(context.Background(), domain.Terminal{
		ID:         "pos-1",
		StoreID:    "store-1",
		Kind:       domain.TerminalKindPOS,
		Status:     domain.TerminalStatusOnline,
		LastSeenAt: lastSeen,
		UpdatedAt:  lastSeen,
	}); err != nil {
		t.Fatalf("save terminal: %v", err)
	}

	result, err := monitoring.ListTerminalOverviews(context.Background(), "store-1", app.PageParams{Limit: 50})
	if err != nil {
		t.Fatalf("list terminal overviews: %v", err)
	}
	if result.Items[0].Terminal.Status != domain.TerminalStatusOffline {
		t.Fatalf("status = %s", result.Items[0].Terminal.Status)
	}
	if !result.Items[0].AttentionNeeded {
		t.Fatal("expected attention needed for offline terminal")
	}
}

func TestGetStoreMonitoringSummaryAggregatesCounts(t *testing.T) {
	store, monitoring := newTestTerminalMonitoringService()
	now := time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)

	day, err := domain.OpenOperationalDay(domain.OpenOperationalDayInput{
		ID:           "oday-1",
		StoreID:      "store-1",
		BusinessDate: "2026-06-18",
		OpenedByID:   "senior-1",
		Now:          now,
	})
	if err != nil {
		t.Fatalf("open operational day: %v", err)
	}
	if err := store.SaveOperationalDay(context.Background(), day); err != nil {
		t.Fatalf("save operational day: %v", err)
	}

	for _, spec := range []struct {
		id       string
		terminal string
		shift    bool
		offline  bool
	}{
		{id: "pos-1", terminal: "pos-1", shift: true},
		{id: "pos-2", terminal: "pos-2"},
		{id: "pos-3", terminal: "pos-3", offline: true},
	} {
		lastSeen := now
		if spec.offline {
			lastSeen = now.Add(-2 * time.Hour)
		}
		if err := store.SaveTerminal(context.Background(), domain.Terminal{
			ID:         spec.terminal,
			StoreID:    "store-1",
			Kind:       domain.TerminalKindPOS,
			Status:     domain.TerminalStatusOnline,
			LastSeenAt: lastSeen,
			UpdatedAt:  lastSeen,
		}); err != nil {
			t.Fatalf("save terminal %s: %v", spec.terminal, err)
		}
		if spec.shift {
			shift, err := domain.OpenShift(domain.OpenShiftInput{
				ID:               "shift-" + spec.terminal,
				StoreID:          "store-1",
				OperationalDayID: day.ID,
				BusinessDate:     day.BusinessDate,
				TerminalID:       spec.terminal,
				CashierID:        "cashier-1",
				DrawerID:         "drawer-" + spec.terminal,
				OpeningCashMinor: 0,
				Now:              now,
			})
			if err != nil {
				t.Fatalf("open shift: %v", err)
			}
			if err := store.SaveShift(context.Background(), shift); err != nil {
				t.Fatalf("save shift: %v", err)
			}
		}
	}

	receipt, err := domain.NewReceipt(domain.NewReceiptInput{
		ID:               "receipt-1",
		StoreID:          "store-1",
		OperationalDayID: day.ID,
		BusinessDate:     day.BusinessDate,
		ShiftID:          "shift-pos-1",
		TerminalID:       "pos-1",
		CashierID:        "cashier-1",
		DrawerID:         "drawer-pos-1",
		Channel:          "pos",
		Now:              now,
	})
	if err != nil {
		t.Fatalf("new receipt: %v", err)
	}
	if err := receipt.AddLine(domain.AddReceiptLineInput{
		ID:             "line-1",
		ProductID:      "sku-1",
		Name:           "Milk",
		Quantity:       1,
		UnitPriceMinor: 30000,
		Now:              now,
	}); err != nil {
		t.Fatalf("add line: %v", err)
	}
	if err := receipt.MarkPaid(now); err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	if err := receipt.MarkFiscalized(now); err != nil {
		t.Fatalf("mark fiscalized: %v", err)
	}
	if err := store.SaveReceipt(context.Background(), receipt); err != nil {
		t.Fatalf("save receipt: %v", err)
	}

	cash := app.NewCashService(store, store)
	if _, err := cash.CreateCashMovement(context.Background(), app.CreateCashMovementCommand{
		IdempotencyKey:    "cash-1",
		StoreID:           "store-1",
		Type:              domain.CashMovementTypeCashSale,
		FromContainerID:   "external-customer",
		FromContainerType: domain.CashContainerTypeExternal,
		ToContainerID:     "drawer-pos-1",
		ToContainerType:   domain.CashContainerTypeDrawer,
		AmountMinor:       30000,
		Currency:          "RUB",
		Reason:            "Cash sale",
		ActorID:           "cashier-1",
	}); err != nil {
		t.Fatalf("create cash movement: %v", err)
	}

	summary, err := monitoring.GetStoreMonitoringSummary(context.Background(), "store-1")
	if err != nil {
		t.Fatalf("get store monitoring summary: %v", err)
	}
	if summary.ActiveTerminalCount != 1 || summary.FreeTerminalCount != 1 || summary.OfflineTerminalCount != 1 {
		t.Fatalf("terminal counts = active %d free %d offline %d", summary.ActiveTerminalCount, summary.FreeTerminalCount, summary.OfflineTerminalCount)
	}
	if summary.ReceiptCountToday != 1 || summary.RevenueMinorToday != 30000 || summary.AverageReceiptMinor != 30000 {
		t.Fatalf("receipt summary = count %d revenue %d avg %d", summary.ReceiptCountToday, summary.RevenueMinorToday, summary.AverageReceiptMinor)
	}
	if summary.DrawerCashMinor != 30000 {
		t.Fatalf("drawer cash = %d", summary.DrawerCashMinor)
	}
}

func newTestTerminalMonitoringService() (*memory.Store, *app.TerminalMonitoringService) {
	store := memory.NewStore()
	cash := app.NewCashService(store, store)
	monitoring := app.NewTerminalMonitoringService(store, store, store, store, cash,
		app.WithTerminalMonitoringClock(func() time.Time {
			return time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
		}),
		app.WithTerminalMonitoringOfflineAfter(time.Minute),
	)
	return store, monitoring
}

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

func TestCreateCashMovementIsIdempotent(t *testing.T) {
	service := newTestCashService()
	command := testCashMovementCommand()

	first, err := service.CreateCashMovement(context.Background(), command)
	if err != nil {
		t.Fatalf("create first cash movement: %v", err)
	}
	second, err := service.CreateCashMovement(context.Background(), command)
	if err != nil {
		t.Fatalf("create second cash movement: %v", err)
	}

	if first.Movement.ID != second.Movement.ID {
		t.Fatalf("expected same cash movement id, got %s and %s", first.Movement.ID, second.Movement.ID)
	}
}

func TestCreateCashMovementRejectsSelfApproval(t *testing.T) {
	service := newTestCashService()
	command := testCashMovementCommand()
	command.ApprovedByID = command.ActorID

	_, err := service.CreateCashMovement(context.Background(), command)
	if !errors.Is(err, app.ErrSeparationOfDutiesViolation) {
		t.Fatalf("expected ErrSeparationOfDutiesViolation, got %v", err)
	}
}

func TestListCashMovements(t *testing.T) {
	service := newTestCashService()
	if _, err := service.CreateCashMovement(context.Background(), testCashMovementCommand()); err != nil {
		t.Fatalf("create cash movement: %v", err)
	}

	movements, err := service.ListCashMovements(context.Background(), "store-1", app.PageParams{Limit: 50})
	if err != nil {
		t.Fatalf("list cash movements: %v", err)
	}
	if len(movements.Items) != 1 {
		t.Fatalf("cash movements count = %d", len(movements.Items))
	}
}

func TestListCashBalancesDerivesBalancesFromPostedMovements(t *testing.T) {
	service := newTestCashService()
	if _, err := service.CreateCashMovement(context.Background(), testCashMovementCommand()); err != nil {
		t.Fatalf("create change fund movement: %v", err)
	}

	collection := testCashMovementCommand()
	collection.IdempotencyKey = "cash-2"
	collection.Type = domain.CashMovementTypeDrawerToSafe
	collection.FromContainerID = "drawer-1"
	collection.FromContainerType = domain.CashContainerTypeDrawer
	collection.ToContainerID = "safe-1"
	collection.ToContainerType = domain.CashContainerTypeSafe
	collection.AmountMinor = 200000
	collection.Reason = "Revenue collection"
	if _, err := service.CreateCashMovement(context.Background(), collection); err != nil {
		t.Fatalf("create collection movement: %v", err)
	}

	balances, err := service.ListCashBalances(context.Background(), "store-1")
	if err != nil {
		t.Fatalf("list cash balances: %v", err)
	}
	if len(balances) != 2 {
		t.Fatalf("cash balances count = %d", len(balances))
	}

	byContainer := map[string]int64{}
	for _, balance := range balances {
		byContainer[balance.ContainerID] = balance.BalanceMinor
	}
	if byContainer["drawer-1"] != 300000 {
		t.Fatalf("drawer balance = %d", byContainer["drawer-1"])
	}
	if byContainer["safe-1"] != -300000 {
		t.Fatalf("safe balance = %d", byContainer["safe-1"])
	}
}

func TestCreateCashRecountUsesDerivedExpectedBalance(t *testing.T) {
	service := newTestCashService()
	if _, err := service.CreateCashMovement(context.Background(), testCashMovementCommand()); err != nil {
		t.Fatalf("create cash movement: %v", err)
	}

	recount, err := service.CreateCashRecount(context.Background(), app.CreateCashRecountCommand{
		IdempotencyKey: "recount-1",
		StoreID:        "store-1",
		ContainerID:    "drawer-1",
		ContainerType:  domain.CashContainerTypeDrawer,
		Currency:       "RUB",
		CountedMinor:   500000,
		Reason:         "Drawer recount",
		ActorID:        "senior-1",
	})
	if err != nil {
		t.Fatalf("create cash recount: %v", err)
	}
	if recount.Recount.Status != "balanced" || recount.Recount.ExpectedMinor != 500000 || recount.Recount.DiscrepancyMinor != 0 {
		t.Fatalf("cash recount = %+v", recount.Recount)
	}
}

func TestCreateCashRecountDiscrepancyRequiresApproval(t *testing.T) {
	service := newTestCashService()

	_, err := service.CreateCashRecount(context.Background(), app.CreateCashRecountCommand{
		IdempotencyKey: "recount-1",
		StoreID:        "store-1",
		ContainerID:    "safe-1",
		ContainerType:  domain.CashContainerTypeSafe,
		Currency:       "RUB",
		CountedMinor:   100000,
		Reason:         "Safe recount",
		ActorID:        "senior-1",
	})
	if !errors.Is(err, app.ErrCashRecountApprovalRequired) {
		t.Fatalf("expected ErrCashRecountApprovalRequired, got %v", err)
	}
}

func TestCreateCashRecountRejectsSelfApproval(t *testing.T) {
	service := newTestCashService()

	_, err := service.CreateCashRecount(context.Background(), app.CreateCashRecountCommand{
		IdempotencyKey: "recount-1",
		StoreID:        "store-1",
		ContainerID:    "safe-1",
		ContainerType:  domain.CashContainerTypeSafe,
		Currency:       "RUB",
		CountedMinor:   100000,
		Reason:         "Safe recount",
		ActorID:        "senior-1",
		ApprovedByID:   "senior-1",
	})
	if !errors.Is(err, app.ErrSeparationOfDutiesViolation) {
		t.Fatalf("expected ErrSeparationOfDutiesViolation, got %v", err)
	}
}

func TestListCashRecounts(t *testing.T) {
	service := newTestCashService()
	if _, err := service.CreateCashRecount(context.Background(), app.CreateCashRecountCommand{
		IdempotencyKey: "recount-1",
		StoreID:        "store-1",
		ContainerID:    "safe-1",
		ContainerType:  domain.CashContainerTypeSafe,
		Currency:       "RUB",
		CountedMinor:   0,
		Reason:         "Safe recount",
		ActorID:        "senior-1",
	}); err != nil {
		t.Fatalf("create cash recount: %v", err)
	}

	recounts, err := service.ListCashRecounts(context.Background(), "store-1", app.PageParams{Limit: 50})
	if err != nil {
		t.Fatalf("list cash recounts: %v", err)
	}
	if len(recounts.Items) != 1 {
		t.Fatalf("cash recounts count = %d", len(recounts.Items))
	}
}

func TestResolveCashRecountMarksDiscrepancyResolved(t *testing.T) {
	service := newTestCashService()
	created, err := service.CreateCashRecount(context.Background(), app.CreateCashRecountCommand{
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
	if created.Recount.ResolutionStatus != domain.CashRecountResolutionStatusOpen {
		t.Fatalf("resolution status = %s", created.Recount.ResolutionStatus)
	}

	resolved, err := service.ResolveCashRecount(context.Background(), app.ResolveCashRecountCommand{
		IdempotencyKey: "recount-resolve-1",
		StoreID:        "store-1",
		RecountID:      created.Recount.ID,
		ResolutionNote: "Adjustment movement will be posted by cash office",
		ActorID:        "senior-1",
		ApprovedByID:   "admin-1",
	})
	if err != nil {
		t.Fatalf("resolve cash recount: %v", err)
	}
	if resolved.Recount.ResolutionStatus != domain.CashRecountResolutionStatusResolved || resolved.Recount.ResolvedByID != "senior-1" {
		t.Fatalf("resolved cash recount = %+v", resolved.Recount)
	}
}

func TestResolveCashRecountRejectsSelfApproval(t *testing.T) {
	service := newTestCashService()
	created, err := service.CreateCashRecount(context.Background(), app.CreateCashRecountCommand{
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

	_, err = service.ResolveCashRecount(context.Background(), app.ResolveCashRecountCommand{
		IdempotencyKey: "recount-resolve-1",
		StoreID:        "store-1",
		RecountID:      created.Recount.ID,
		ResolutionNote: "Adjustment movement will be posted by cash office",
		ActorID:        "senior-1",
		ApprovedByID:   "senior-1",
	})
	if !errors.Is(err, app.ErrSeparationOfDutiesViolation) {
		t.Fatalf("expected ErrSeparationOfDutiesViolation, got %v", err)
	}
}

func TestCreateBankCollectionMovesSafeToBank(t *testing.T) {
	service := newTestCashService()
	if _, err := service.CreateCashMovement(context.Background(), app.CreateCashMovementCommand{
		IdempotencyKey:    "seed-safe-1",
		StoreID:           "store-1",
		Type:              domain.CashMovementTypeCashIn,
		FromContainerID:   "external-customer",
		FromContainerType: domain.CashContainerTypeExternal,
		ToContainerID:     "safe-1",
		ToContainerType:   domain.CashContainerTypeSafe,
		AmountMinor:       500000,
		Currency:          "RUB",
		Reason:            "Seed safe balance",
		ActorID:           "senior-1",
	}); err != nil {
		t.Fatalf("seed safe: %v", err)
	}

	if _, err := service.CreateBankCollection(context.Background(), app.CreateBankCollectionCommand{
		IdempotencyKey:  "bank-collection-1",
		StoreID:         "store-1",
		SafeID:          "safe-1",
		BankContainerID: "bank-collection-1",
		AmountMinor:     200000,
		ActorID:         "senior-1",
		ApprovedByID:    "admin-1",
		Reason:          "Scheduled bank collection",
	}); err != nil {
		t.Fatalf("create bank collection: %v", err)
	}

	balances, err := service.ListCashBalances(context.Background(), "store-1")
	if err != nil {
		t.Fatalf("list balances: %v", err)
	}
	byContainer := map[string]int64{}
	for _, balance := range balances {
		byContainer[balance.ContainerID] = balance.BalanceMinor
	}
	if byContainer["safe-1"] != 300000 {
		t.Fatalf("safe balance = %d", byContainer["safe-1"])
	}
	if byContainer["bank-collection-1"] != 200000 {
		t.Fatalf("bank balance = %d", byContainer["bank-collection-1"])
	}
}

func TestCreateBusinessExpenseMovesSafeToPayee(t *testing.T) {
	service := newTestCashService()
	if _, err := service.CreateCashMovement(context.Background(), app.CreateCashMovementCommand{
		IdempotencyKey:    "seed-safe-2",
		StoreID:           "store-1",
		Type:              domain.CashMovementTypeCashIn,
		FromContainerID:   "external-customer",
		FromContainerType: domain.CashContainerTypeExternal,
		ToContainerID:     "safe-1",
		ToContainerType:   domain.CashContainerTypeSafe,
		AmountMinor:       100000,
		Currency:          "RUB",
		Reason:            "Seed safe balance",
		ActorID:           "senior-1",
	}); err != nil {
		t.Fatalf("seed safe: %v", err)
	}

	if _, err := service.CreateBusinessExpense(context.Background(), app.CreateBusinessExpenseCommand{
		IdempotencyKey: "expense-1",
		StoreID:        "store-1",
		SafeID:         "safe-1",
		PayeeID:        "vendor-supplies",
		AmountMinor:    40000,
		Reason:         "Office supplies",
		ActorID:        "senior-1",
		ApprovedByID:   "admin-1",
	}); err != nil {
		t.Fatalf("create business expense: %v", err)
	}

	balances, err := service.ListCashBalances(context.Background(), "store-1")
	if err != nil {
		t.Fatalf("list balances: %v", err)
	}
	byContainer := map[string]int64{}
	for _, balance := range balances {
		byContainer[balance.ContainerID] = balance.BalanceMinor
	}
	if byContainer["safe-1"] != 60000 {
		t.Fatalf("safe balance = %d", byContainer["safe-1"])
	}
	if byContainer["vendor-supplies"] != 40000 {
		t.Fatalf("expense balance = %d", byContainer["vendor-supplies"])
	}
}

func newTestCashService() *app.CashService {
	store := memory.NewStore()
	var counter int
	return app.NewCashService(store, store,
		app.WithCashClock(func() time.Time {
			return time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC)
		}),
		app.WithCashIDGenerator(func(prefix string) string {
			counter++
			return fmt.Sprintf("%s-test-%d", prefix, counter)
		}),
	)
}

func testCashMovementCommand() app.CreateCashMovementCommand {
	return app.CreateCashMovementCommand{
		IdempotencyKey:    "cash-1",
		StoreID:           "store-1",
		Type:              domain.CashMovementTypeChangeFund,
		FromContainerID:   "safe-1",
		FromContainerType: domain.CashContainerTypeSafe,
		ToContainerID:     "drawer-1",
		ToContainerType:   domain.CashContainerTypeDrawer,
		AmountMinor:       500000,
		Currency:          "RUB",
		Reason:            "Opening change fund",
		ActorID:           "senior-1",
		ApprovedByID:      "cashier-1",
	}
}

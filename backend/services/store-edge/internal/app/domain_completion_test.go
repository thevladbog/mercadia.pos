package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/app"
	"mercadia.dev/pos/services/store-edge/internal/domain"
	"mercadia.dev/pos/services/store-edge/internal/infra/memory"
)

func TestCreateSessionWithValidCredentials(t *testing.T) {
	store := memory.NewStore(memory.WithDemoActors())
	auth := app.NewAuthService(store, store)

	result, err := auth.CreateSession(context.Background(), app.CreateSessionCommand{
		ActorID: "senior-1",
		PIN:     "5678",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if result.Token == "" || result.ActorID != "senior-1" {
		t.Fatalf("session = %+v", result)
	}
}

func TestCreateSessionRejectsInvalidPIN(t *testing.T) {
	store := memory.NewStore(memory.WithDemoActors())
	auth := app.NewAuthService(store, store)

	_, err := auth.CreateSession(context.Background(), app.CreateSessionCommand{
		ActorID: "senior-1",
		PIN:     "0000",
	})
	if !errors.Is(err, app.ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}

func TestRBACPermissions(t *testing.T) {
	if app.HasPermission([]domain.Role{domain.RoleCashier}, app.PermissionDiscountApply) {
		t.Fatal("cashier should not apply discounts")
	}
	if !app.HasPermission([]domain.Role{domain.RoleSeniorCashier}, app.PermissionReturnsCreate) {
		t.Fatal("senior cashier should create returns")
	}
}

func TestApplyLineDiscountRequiresPermission(t *testing.T) {
	store := memory.NewStore(memory.WithDemoActors())
	auth := app.NewAuthService(store, store)
	discounts := app.NewDiscountService(store, store, auth)

	receipt, err := domain.NewReceipt(domain.NewReceiptInput{
		ID:         "receipt-1",
		StoreID:    "store-1",
		TerminalID: "pos-1",
		CashierID:  "cashier-1",
		Now:        time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create receipt: %v", err)
	}
	if err := receipt.AddLine(domain.AddReceiptLineInput{
		ID: "line-1", ProductID: "sku-1", Name: "Milk", Quantity: 1, UnitPriceMinor: 1000,
		Now: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("add line: %v", err)
	}
	if err := store.SaveReceipt(context.Background(), receipt); err != nil {
		t.Fatalf("save receipt: %v", err)
	}

	_, err = discounts.ApplyLineDiscount(context.Background(), app.ApplyLineDiscountCommand{
		IdempotencyKey: "disc-1",
		ReceiptID:      receipt.ID,
		LineID:         "line-1",
		AmountMinor:    100,
		Reason:         "Customer loyalty",
		ActorID:        "cashier-1",
	})
	if !errors.Is(err, app.ErrPermissionDenied) {
		t.Fatalf("expected permission denied, got %v", err)
	}

	updated, err := discounts.ApplyLineDiscount(context.Background(), app.ApplyLineDiscountCommand{
		IdempotencyKey: "disc-2",
		ReceiptID:      receipt.ID,
		LineID:         "line-1",
		AmountMinor:    100,
		Reason:         "Customer loyalty",
		ActorID:        "senior-1",
	})
	if err != nil {
		t.Fatalf("apply discount: %v", err)
	}
	if updated.Receipt.TotalMinor() != 900 {
		t.Fatalf("total = %d", updated.Receipt.TotalMinor())
	}
}

func TestReceiptReturnRequiresFiscalizedReceipt(t *testing.T) {
	store := memory.NewStore(memory.WithDemoActors())
	auth := app.NewAuthService(store, store)
	returns := app.NewReturnsService(store, store, store, auth)

	receipt, err := domain.NewReceipt(domain.NewReceiptInput{
		ID: "receipt-1", StoreID: "store-1", TerminalID: "pos-1", CashierID: "cashier-1",
		Now: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create receipt: %v", err)
	}
	if err := receipt.AddLine(domain.AddReceiptLineInput{
		ID: "line-1", ProductID: "sku-1", Name: "Milk", Quantity: 1, UnitPriceMinor: 1000,
		Now: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("add line: %v", err)
	}
	if err := store.SaveReceipt(context.Background(), receipt); err != nil {
		t.Fatalf("save receipt: %v", err)
	}

	_, err = returns.CreateReceiptReturn(context.Background(), app.CreateReceiptReturnCommand{
		IdempotencyKey: "ret-1",
		ReceiptID:      receipt.ID,
		Lines:          []app.ReturnLineCommand{{LineID: "line-1", Quantity: 1}},
		Reason:         "Defective",
		ActorID:        "senior-1",
	})
	if !errors.Is(err, app.ErrReceiptNotReturnable) {
		t.Fatalf("expected not returnable, got %v", err)
	}
}

func TestNoReceiptReturnRequiresApproval(t *testing.T) {
	store := memory.NewStore(memory.WithDemoActors())
	auth := app.NewAuthService(store, store)
	returns := app.NewReturnsService(store, store, store, auth)

	_, err := returns.CreateNoReceiptReturn(context.Background(), app.CreateNoReceiptReturnCommand{
		IdempotencyKey: "ret-2",
		StoreID:        "store-1",
		Lines:          []app.ReturnLineCommand{{ProductID: "sku-1", Name: "Milk", Quantity: 1, UnitPriceMinor: 1000}},
		Reason:         "No receipt",
		ActorID:        "senior-1",
	})
	if !errors.Is(err, app.ErrInvalidReturnCommand) {
		t.Fatalf("expected invalid return command, got %v", err)
	}

	result, err := returns.CreateNoReceiptReturn(context.Background(), app.CreateNoReceiptReturnCommand{
		IdempotencyKey: "ret-3",
		StoreID:        "store-1",
		Lines:          []app.ReturnLineCommand{{ProductID: "sku-1", Name: "Milk", Quantity: 1, UnitPriceMinor: 1000}},
		Reason:         "No receipt",
		ActorID:        "senior-1",
		ApprovedByID:   "admin-1",
	})
	if err != nil {
		t.Fatalf("create no-receipt return: %v", err)
	}
	if result.Return.Kind != domain.ReturnKindNoReceipt {
		t.Fatalf("return kind = %s", result.Return.Kind)
	}
}

func TestValidateDataMatrixCode(t *testing.T) {
	valid, err := domain.ValidateDataMatrixCode("0104600000000000215ABC")
	if err != nil {
		t.Fatalf("validate marking: %v", err)
	}
	if !valid.Valid {
		t.Fatal("expected valid marking code")
	}

	invalid, err := domain.ValidateDataMatrixCode("invalid")
	if err != nil {
		t.Fatalf("validate marking: %v", err)
	}
	if invalid.Valid {
		t.Fatal("expected invalid marking code")
	}
}

func TestOperationJournalRecordsCashMovement(t *testing.T) {
	store := memory.NewStore(memory.WithDemoActors())
	journal := app.NewOperationJournalService(store)
	cash := app.NewCashService(store, store, app.WithCashJournal(journal))

	if _, err := cash.CreateCashMovement(context.Background(), app.CreateCashMovementCommand{
		IdempotencyKey:    "cash-journal-1",
		StoreID:           "store-1",
		Type:              domain.CashMovementTypeChangeFund,
		FromContainerID:   "safe-1",
		FromContainerType: domain.CashContainerTypeSafe,
		ToContainerID:     "drawer-1",
		ToContainerType:   domain.CashContainerTypeDrawer,
		AmountMinor:       100000,
		Currency:          "RUB",
		ActorID:           "senior-1",
	}); err != nil {
		t.Fatalf("create cash movement: %v", err)
	}

	entries, err := journal.ListOperationJournal(context.Background(), "store-1", app.PageParams{Limit: 10})
	if err != nil {
		t.Fatalf("list journal: %v", err)
	}
	if len(entries.Items) != 1 || entries.Items[0].OperationType != "cash.movement.created" {
		t.Fatalf("journal entries = %+v", entries.Items)
	}
}

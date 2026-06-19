package domain

import (
	"errors"
	"time"
)

type CashContainerType string
type CashMovementType string
type CashMovementStatus string

const (
	CashContainerTypeDrawer   CashContainerType = "drawer"
	CashContainerTypeSafe     CashContainerType = "safe"
	CashContainerTypeBank     CashContainerType = "bank"
	CashContainerTypeExpense  CashContainerType = "expense"
	CashContainerTypeExternal CashContainerType = "external"

	CashMovementTypeChangeFund            CashMovementType = "change_fund"
	CashMovementTypeCashIn                CashMovementType = "cash_in"
	CashMovementTypeCashOut               CashMovementType = "cash_out"
	CashMovementTypeDrawerToSafe          CashMovementType = "drawer_to_safe"
	CashMovementTypeSafeToBank            CashMovementType = "safe_to_bank"
	CashMovementTypeExpense               CashMovementType = "expense"
	CashMovementTypeAdjustment            CashMovementType = "adjustment"
	CashMovementTypeCashSale              CashMovementType = "cash_sale"
	CashMovementTypeCashSaleReversal      CashMovementType = "cash_sale_reversal"
	CashMovementTypeNoReceiptReturnPayout CashMovementType = "no_receipt_return_payout"

	CashMovementStatusPosted CashMovementStatus = "posted"
)

var ErrInvalidCashMovementInput = errors.New("invalid cash movement input")

type CashMovement struct {
	ID                string
	StoreID           string
	Type              CashMovementType
	FromContainerID   string
	FromContainerType CashContainerType
	ToContainerID     string
	ToContainerType   CashContainerType
	AmountMinor       int64
	Currency          string
	Reason            string
	ActorID           string
	ApprovedByID      string
	Status            CashMovementStatus
	CreatedAt         time.Time
}

type CashBalance struct {
	StoreID        string
	ContainerID    string
	ContainerType  CashContainerType
	Currency       string
	BalanceMinor   int64
	LastMovementAt time.Time
}

type CreateCashMovementInput struct {
	ID                string
	StoreID           string
	Type              CashMovementType
	FromContainerID   string
	FromContainerType CashContainerType
	ToContainerID     string
	ToContainerType   CashContainerType
	AmountMinor       int64
	Currency          string
	Reason            string
	ActorID           string
	ApprovedByID      string
	Now               time.Time
}

func CreateCashMovement(input CreateCashMovementInput) (CashMovement, error) {
	if input.ID == "" || input.StoreID == "" || input.Type == "" || input.FromContainerID == "" ||
		input.FromContainerType == "" || input.ToContainerID == "" || input.ToContainerType == "" ||
		input.AmountMinor <= 0 || input.ActorID == "" {
		return CashMovement{}, ErrInvalidCashMovementInput
	}
	if input.Currency == "" {
		input.Currency = "RUB"
	}
	if input.Now.IsZero() {
		input.Now = time.Now().UTC()
	}

	return CashMovement{
		ID:                input.ID,
		StoreID:           input.StoreID,
		Type:              input.Type,
		FromContainerID:   input.FromContainerID,
		FromContainerType: input.FromContainerType,
		ToContainerID:     input.ToContainerID,
		ToContainerType:   input.ToContainerType,
		AmountMinor:       input.AmountMinor,
		Currency:          input.Currency,
		Reason:            input.Reason,
		ActorID:           input.ActorID,
		ApprovedByID:      input.ApprovedByID,
		Status:            CashMovementStatusPosted,
		CreatedAt:         input.Now,
	}, nil
}

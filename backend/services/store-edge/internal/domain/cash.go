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

var (
	ErrInvalidCashMovementInput      = errors.New("invalid cash movement input")
	ErrDenominationBreakdownMismatch = errors.New("denomination breakdown does not match operation total")
)

// DenominationBreakdown captures the optional bill/coin count breakdown of a
// cash operation's total. Bills maps a denomination's minor-unit value (e.g.
// 5000_00 for a 5000-currency-unit note) to how many of that note were
// counted. CoinsMinor and OtherMinor are lump sums, matching the
// DenominationGrid frontend's own "МОНЕТЫ"/"ДРУГОЕ" fields.
type DenominationBreakdown struct {
	Bills      map[int64]int
	CoinsMinor int64
	OtherMinor int64
}

// Sum returns Σ(denomination × count) + CoinsMinor + OtherMinor. This is the
// single validation primitive every breakdown-accepting constructor reuses.
func (b DenominationBreakdown) Sum() int64 {
	var total int64
	for denomination, count := range b.Bills {
		total += denomination * int64(count)
	}
	return total + b.CoinsMinor + b.OtherMinor
}

// validateDenominationBreakdown is a no-op when breakdown is nil (the
// backward-compatible "no breakdown provided" path). When non-nil, it
// rejects negative bill counts, non-positive denomination keys, negative
// lump sums, and any breakdown whose Sum() doesn't exactly equal
// expectedTotal (AmountMinor for cash movements, CountedMinor for recounts).
func validateDenominationBreakdown(breakdown *DenominationBreakdown, expectedTotal int64) error {
	if breakdown == nil {
		return nil
	}
	if breakdown.CoinsMinor < 0 || breakdown.OtherMinor < 0 {
		return ErrDenominationBreakdownMismatch
	}
	for denomination, count := range breakdown.Bills {
		if denomination <= 0 || count < 0 {
			return ErrDenominationBreakdownMismatch
		}
	}
	if breakdown.Sum() != expectedTotal {
		return ErrDenominationBreakdownMismatch
	}
	return nil
}

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
	Breakdown         *DenominationBreakdown
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
	Breakdown         *DenominationBreakdown
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
	if err := validateDenominationBreakdown(input.Breakdown, input.AmountMinor); err != nil {
		return CashMovement{}, err
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
		Breakdown:         input.Breakdown,
	}, nil
}

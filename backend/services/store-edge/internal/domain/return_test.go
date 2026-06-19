package domain

import (
	"errors"
	"testing"
)

func TestValidateReceiptReturnCumulativeBlocksPriorQuantity(t *testing.T) {
	receipt := Receipt{
		Status: ReceiptStatusFiscalized,
		Lines: []ReceiptLine{
			{ID: "line-1", Quantity: 2, UnitPriceMinor: 1000},
		},
	}
	priorReturns := []Return{
		{
			Kind:   ReturnKindWithReceipt,
			Status: ReturnStatusSettled,
			Lines:  []ReturnLine{{LineID: "line-1", Quantity: 1}},
		},
	}
	newLines := []ReturnLineInput{{LineID: "line-1", Quantity: 2}}

	err := ValidateReceiptReturnCumulative(receipt, newLines, priorReturns)
	if !errors.Is(err, ErrReturnQuantityExceeded) {
		t.Fatalf("expected ErrReturnQuantityExceeded, got %v", err)
	}
}

func TestValidateReceiptReturnCumulativeCountsCompletedReturn(t *testing.T) {
	receipt := Receipt{
		Status: ReceiptStatusFiscalized,
		Lines: []ReceiptLine{
			{ID: "line-1", Quantity: 2, UnitPriceMinor: 1000},
		},
	}
	priorReturns := []Return{
		{
			Kind:   ReturnKindWithReceipt,
			Status: ReturnStatusCompleted,
			Lines:  []ReturnLine{{LineID: "line-1", Quantity: 1}},
		},
	}
	newLines := []ReturnLineInput{{LineID: "line-1", Quantity: 2}}

	err := ValidateReceiptReturnCumulative(receipt, newLines, priorReturns)
	if !errors.Is(err, ErrReturnQuantityExceeded) {
		t.Fatalf("expected ErrReturnQuantityExceeded, got %v", err)
	}
}

func TestValidateReceiptReturnCumulativeAllowsRemainingQuantity(t *testing.T) {
	receipt := Receipt{
		Status: ReceiptStatusFiscalized,
		Lines: []ReceiptLine{
			{ID: "line-1", Quantity: 2, UnitPriceMinor: 1000},
		},
	}
	priorReturns := []Return{
		{
			Kind:   ReturnKindWithReceipt,
			Status: ReturnStatusSettled,
			Lines:  []ReturnLine{{LineID: "line-1", Quantity: 1}},
		},
	}
	newLines := []ReturnLineInput{{LineID: "line-1", Quantity: 1}}

	if err := ValidateReceiptReturnCumulative(receipt, newLines, priorReturns); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestValidateReceiptReturnCumulativeAggregatesDuplicateLineIDs(t *testing.T) {
	receipt := Receipt{
		Status: ReceiptStatusFiscalized,
		Lines: []ReceiptLine{
			{ID: "line-1", Quantity: 2, UnitPriceMinor: 1000},
		},
	}
	newLines := []ReturnLineInput{
		{LineID: "line-1", Quantity: 1},
		{LineID: "line-1", Quantity: 2},
	}

	err := ValidateReceiptReturnCumulative(receipt, newLines, nil)
	if !errors.Is(err, ErrReturnQuantityExceeded) {
		t.Fatalf("expected ErrReturnQuantityExceeded, got %v", err)
	}
}

func TestValidateReceiptReturnCumulativeIgnoresNoReceiptPriorReturns(t *testing.T) {
	receipt := Receipt{
		Status: ReceiptStatusFiscalized,
		Lines: []ReceiptLine{
			{ID: "line-1", Quantity: 2, UnitPriceMinor: 1000},
		},
	}
	priorReturns := []Return{
		{
			Kind:   ReturnKindNoReceipt,
			Status: ReturnStatusCompleted,
			Lines:  []ReturnLine{{LineID: "line-1", Quantity: 99}},
		},
	}
	newLines := []ReturnLineInput{{LineID: "line-1", Quantity: 2}}

	if err := ValidateReceiptReturnCumulative(receipt, newLines, priorReturns); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestValidateReceiptReturnCumulativeMultiLineReceipt(t *testing.T) {
	receipt := Receipt{
		Status: ReceiptStatusFiscalized,
		Lines: []ReceiptLine{
			{ID: "line-1", Quantity: 2, UnitPriceMinor: 1000},
			{ID: "line-2", Quantity: 3, UnitPriceMinor: 500},
		},
	}
	priorReturns := []Return{
		{
			Kind:   ReturnKindWithReceipt,
			Status: ReturnStatusSettled,
			Lines: []ReturnLine{
				{LineID: "line-1", Quantity: 1},
				{LineID: "line-2", Quantity: 2},
			},
		},
	}
	newLines := []ReturnLineInput{
		{LineID: "line-1", Quantity: 1},
		{LineID: "line-2", Quantity: 1},
	}

	if err := ValidateReceiptReturnCumulative(receipt, newLines, priorReturns); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestValidateReceiptReturnCumulativeDelegatesToValidateReceiptReturn(t *testing.T) {
	receipt := Receipt{
		Status: ReceiptStatusPaid,
		Lines: []ReceiptLine{
			{ID: "line-1", Quantity: 2, UnitPriceMinor: 1000},
		},
	}
	newLines := []ReturnLineInput{{LineID: "line-1", Quantity: 1}}

	err := ValidateReceiptReturnCumulative(receipt, newLines, nil)
	if !errors.Is(err, ErrReceiptNotReturnable) {
		t.Fatalf("expected ErrReceiptNotReturnable, got %v", err)
	}
}

func TestValidateReceiptReturnCumulativeEmptyPriorReturns(t *testing.T) {
	receipt := Receipt{
		Status: ReceiptStatusFiscalized,
		Lines: []ReceiptLine{
			{ID: "line-1", Quantity: 2, UnitPriceMinor: 1000},
		},
	}
	newLines := []ReturnLineInput{{LineID: "line-1", Quantity: 2, UnitPriceMinor: 1000}}

	if err := ValidateReceiptReturnCumulative(receipt, newLines, nil); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

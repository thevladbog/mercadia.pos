package domain

import (
	"errors"
	"time"
)

type ReturnKind string

const (
	ReturnKindWithReceipt ReturnKind = "with_receipt"
	ReturnKindNoReceipt   ReturnKind = "no_receipt"
)

type ReturnStatus string

const (
	ReturnStatusCompleted ReturnStatus = "completed"
	ReturnStatusSettled   ReturnStatus = "settled"
)

var (
	ErrInvalidReturnInput     = errors.New("invalid return input")
	ErrReceiptNotReturnable   = errors.New("receipt is not returnable")
	ErrReturnLineNotFound     = errors.New("return line not found")
	ErrReturnQuantityExceeded = errors.New("return quantity exceeds line quantity")
	ErrReturnAlreadySettled   = errors.New("return is already settled")
	ErrReturnSettlementNotAllowed = errors.New("return settlement is not allowed")
)

type ReturnLine struct {
	LineID         string
	ProductID      string
	Name           string
	Quantity       int64
	UnitPriceMinor int64
	TotalMinor     int64
}

type Return struct {
	ID           string
	StoreID      string
	ReceiptID    string
	Kind         ReturnKind
	Lines        []ReturnLine
	Reason       string
	ActorID      string
	ApprovedByID string
	TotalMinor   int64
	Status       ReturnStatus
	CreatedAt    time.Time
}

type ReturnLineInput struct {
	LineID         string
	ProductID      string
	Name           string
	Quantity       int64
	UnitPriceMinor int64
}

type CreateReturnInput struct {
	ID           string
	StoreID      string
	ReceiptID    string
	Kind         ReturnKind
	Lines        []ReturnLineInput
	Reason       string
	ActorID      string
	ApprovedByID string
	Now          time.Time
}

func NewReturn(input CreateReturnInput) (Return, error) {
	if input.ID == "" || input.StoreID == "" || input.Reason == "" || input.ActorID == "" {
		return Return{}, ErrInvalidReturnInput
	}
	if input.Kind == "" {
		input.Kind = ReturnKindWithReceipt
	}
	if input.Kind == ReturnKindWithReceipt && input.ReceiptID == "" {
		return Return{}, ErrInvalidReturnInput
	}
	if input.Kind == ReturnKindNoReceipt && input.ApprovedByID == "" {
		return Return{}, ErrInvalidReturnInput
	}
	if input.Kind == ReturnKindNoReceipt && input.ApprovedByID == input.ActorID {
		return Return{}, ErrInvalidReturnInput
	}
	if len(input.Lines) == 0 {
		return Return{}, ErrInvalidReturnInput
	}
	if input.Now.IsZero() {
		input.Now = time.Now().UTC()
	}

	lines := make([]ReturnLine, 0, len(input.Lines))
	var total int64
	for _, lineInput := range input.Lines {
		if lineInput.Quantity <= 0 || lineInput.UnitPriceMinor < 0 {
			return Return{}, ErrInvalidReturnInput
		}
		if input.Kind == ReturnKindWithReceipt && lineInput.LineID == "" {
			return Return{}, ErrInvalidReturnInput
		}
		if input.Kind == ReturnKindNoReceipt && (lineInput.ProductID == "" || lineInput.Name == "") {
			return Return{}, ErrInvalidReturnInput
		}
		lineTotal := lineInput.Quantity * lineInput.UnitPriceMinor
		lines = append(lines, ReturnLine{
			LineID:         lineInput.LineID,
			ProductID:      lineInput.ProductID,
			Name:           lineInput.Name,
			Quantity:       lineInput.Quantity,
			UnitPriceMinor: lineInput.UnitPriceMinor,
			TotalMinor:     lineTotal,
		})
		total += lineTotal
	}

	return Return{
		ID:           input.ID,
		StoreID:      input.StoreID,
		ReceiptID:    input.ReceiptID,
		Kind:         input.Kind,
		Lines:        lines,
		Reason:       input.Reason,
		ActorID:      input.ActorID,
		ApprovedByID: input.ApprovedByID,
		TotalMinor:   total,
		Status:       ReturnStatusCompleted,
		CreatedAt:    input.Now,
	}, nil
}

func ValidateReceiptReturn(receipt Receipt, lines []ReturnLineInput) error {
	if receipt.Status != ReceiptStatusFiscalized {
		return ErrReceiptNotReturnable
	}
	receiptLines := map[string]ReceiptLine{}
	for _, line := range receipt.Lines {
		receiptLines[line.ID] = line
	}
	for _, lineInput := range lines {
		receiptLine, ok := receiptLines[lineInput.LineID]
		if !ok {
			return ErrReturnLineNotFound
		}
		if lineInput.Quantity > receiptLine.Quantity {
			return ErrReturnQuantityExceeded
		}
	}
	return nil
}

func (r *Return) MarkSettled(now time.Time) error {
	if r.Status == ReturnStatusSettled {
		return ErrReturnAlreadySettled
	}
	if r.Status != ReturnStatusCompleted {
		return ErrReturnSettlementNotAllowed
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	r.Status = ReturnStatusSettled
	return nil
}

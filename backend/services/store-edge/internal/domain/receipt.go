package domain

import (
	"errors"
	"time"
)

type ReceiptStatus string

const (
	ReceiptStatusDraft          ReceiptStatus = "draft"
	ReceiptStatusPaymentStarted ReceiptStatus = "payment_started"
	ReceiptStatusPaid           ReceiptStatus = "paid"
	ReceiptStatusFiscalized     ReceiptStatus = "fiscalized"
	ReceiptStatusCancelled      ReceiptStatus = "cancelled"
)

var (
	ErrReceiptClosed            = errors.New("receipt is not editable")
	ErrReceiptCannotBeCancelled = errors.New("receipt cannot be cancelled")
	ErrInvalidReceiptInput      = errors.New("invalid receipt input")
)

type Receipt struct {
	ID                 string
	StoreID            string
	OperationalDayID   string
	BusinessDate       string
	ShiftID            string
	TerminalID         string
	CashierID          string
	DrawerID           string
	Channel            string
	Status             ReceiptStatus
	Lines              []ReceiptLine
	CancelReason       string
	CancelledByID      string
	CancelApprovedByID string
	CancelledAt        time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type ReceiptLine struct {
	ID             string
	ProductID      string
	Barcode        string
	Name           string
	Quantity       int64
	UnitPriceMinor int64
	TotalMinor     int64
	AddedAt        time.Time
}

type NewReceiptInput struct {
	ID               string
	StoreID          string
	OperationalDayID string
	BusinessDate     string
	ShiftID          string
	TerminalID       string
	CashierID        string
	DrawerID         string
	Channel          string
	Now              time.Time
}

type AddReceiptLineInput struct {
	ID             string
	ProductID      string
	Barcode        string
	Name           string
	Quantity       int64
	UnitPriceMinor int64
	Now            time.Time
}

type CancelReceiptInput struct {
	Reason       string
	ActorID      string
	ApprovedByID string
	Now          time.Time
}

func NewReceipt(input NewReceiptInput) (Receipt, error) {
	if input.ID == "" || input.StoreID == "" || input.TerminalID == "" || input.CashierID == "" {
		return Receipt{}, ErrInvalidReceiptInput
	}
	if input.Channel == "" {
		input.Channel = "pos"
	}
	if input.Now.IsZero() {
		input.Now = time.Now().UTC()
	}

	return Receipt{
		ID:               input.ID,
		StoreID:          input.StoreID,
		OperationalDayID: input.OperationalDayID,
		BusinessDate:     input.BusinessDate,
		ShiftID:          input.ShiftID,
		TerminalID:       input.TerminalID,
		CashierID:        input.CashierID,
		DrawerID:         input.DrawerID,
		Channel:          input.Channel,
		Status:           ReceiptStatusDraft,
		Lines:            []ReceiptLine{},
		CreatedAt:        input.Now,
		UpdatedAt:        input.Now,
	}, nil
}

func (r *Receipt) AddLine(input AddReceiptLineInput) error {
	if r.Status != ReceiptStatusDraft {
		return ErrReceiptClosed
	}
	if input.ID == "" || input.ProductID == "" || input.Name == "" || input.Quantity <= 0 || input.UnitPriceMinor < 0 {
		return ErrInvalidReceiptInput
	}
	if input.Now.IsZero() {
		input.Now = time.Now().UTC()
	}

	line := ReceiptLine{
		ID:             input.ID,
		ProductID:      input.ProductID,
		Barcode:        input.Barcode,
		Name:           input.Name,
		Quantity:       input.Quantity,
		UnitPriceMinor: input.UnitPriceMinor,
		TotalMinor:     input.Quantity * input.UnitPriceMinor,
		AddedAt:        input.Now,
	}

	r.Lines = append(r.Lines, line)
	r.UpdatedAt = input.Now
	return nil
}

func (r *Receipt) MarkPaymentStarted(now time.Time) error {
	if r.Status != ReceiptStatusDraft && r.Status != ReceiptStatusPaymentStarted {
		return ErrReceiptClosed
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	r.Status = ReceiptStatusPaymentStarted
	r.UpdatedAt = now
	return nil
}

func (r *Receipt) MarkPaid(now time.Time) error {
	if r.Status != ReceiptStatusDraft && r.Status != ReceiptStatusPaymentStarted {
		return ErrReceiptClosed
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	r.Status = ReceiptStatusPaid
	r.UpdatedAt = now
	return nil
}

func (r *Receipt) MarkFiscalized(now time.Time) error {
	if r.Status != ReceiptStatusPaid {
		return ErrReceiptClosed
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	r.Status = ReceiptStatusFiscalized
	r.UpdatedAt = now
	return nil
}

func (r *Receipt) Cancel(input CancelReceiptInput) error {
	if r.Status != ReceiptStatusDraft {
		return ErrReceiptCannotBeCancelled
	}
	if input.Reason == "" || input.ActorID == "" {
		return ErrInvalidReceiptInput
	}
	if input.Now.IsZero() {
		input.Now = time.Now().UTC()
	}

	r.Status = ReceiptStatusCancelled
	r.CancelReason = input.Reason
	r.CancelledByID = input.ActorID
	r.CancelApprovedByID = input.ApprovedByID
	r.CancelledAt = input.Now
	r.UpdatedAt = input.Now
	return nil
}

func (r Receipt) TotalMinor() int64 {
	var total int64
	for _, line := range r.Lines {
		total += line.TotalMinor
	}
	return total
}

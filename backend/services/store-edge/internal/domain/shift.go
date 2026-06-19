package domain

import (
	"errors"
	"time"
)

type ShiftStatus string

const (
	ShiftStatusOpen   ShiftStatus = "open"
	ShiftStatusClosed ShiftStatus = "closed"
)

var (
	ErrInvalidShiftInput = errors.New("invalid shift input")
	ErrShiftNotOpen      = errors.New("shift is not open")
)

type Shift struct {
	ID               string
	StoreID          string
	OperationalDayID string
	BusinessDate     string
	TerminalID       string
	CashierID        string
	DrawerID         string
	Status           ShiftStatus
	OpeningCashMinor int64
	ClosingCashMinor int64
	OpenedAt         time.Time
	ClosedAt         time.Time
	UpdatedAt        time.Time
}

type OpenShiftInput struct {
	ID               string
	StoreID          string
	OperationalDayID string
	BusinessDate     string
	TerminalID       string
	CashierID        string
	DrawerID         string
	OpeningCashMinor int64
	Now              time.Time
}

func OpenShift(input OpenShiftInput) (Shift, error) {
	if input.ID == "" || input.StoreID == "" || input.TerminalID == "" ||
		input.CashierID == "" || input.DrawerID == "" || input.OpeningCashMinor < 0 {
		return Shift{}, ErrInvalidShiftInput
	}
	if input.Now.IsZero() {
		input.Now = time.Now().UTC()
	}

	return Shift{
		ID:               input.ID,
		StoreID:          input.StoreID,
		OperationalDayID: input.OperationalDayID,
		BusinessDate:     input.BusinessDate,
		TerminalID:       input.TerminalID,
		CashierID:        input.CashierID,
		DrawerID:         input.DrawerID,
		Status:           ShiftStatusOpen,
		OpeningCashMinor: input.OpeningCashMinor,
		OpenedAt:         input.Now,
		UpdatedAt:        input.Now,
	}, nil
}

func (s *Shift) Close(closingCashMinor int64, now time.Time) error {
	if s.Status != ShiftStatusOpen {
		return ErrShiftNotOpen
	}
	if closingCashMinor < 0 {
		return ErrInvalidShiftInput
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	s.Status = ShiftStatusClosed
	s.ClosingCashMinor = closingCashMinor
	s.ClosedAt = now
	s.UpdatedAt = now
	return nil
}

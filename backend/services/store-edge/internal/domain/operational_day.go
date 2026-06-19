package domain

import (
	"errors"
	"time"
)

type OperationalDayStatus string
type OperationalDayBlockerSeverity string

const (
	OperationalDayStatusOpen   OperationalDayStatus = "open"
	OperationalDayStatusClosed OperationalDayStatus = "closed"

	OperationalDayBlockerSeverityBlocker               OperationalDayBlockerSeverity = "blocker"
	OperationalDayBlockerSeverityRequiresAdminOverride OperationalDayBlockerSeverity = "requires_admin_override"
)

var (
	ErrInvalidOperationalDayInput = errors.New("invalid operational day input")
	ErrOperationalDayNotOpen      = errors.New("operational day is not open")
)

type OperationalDay struct {
	ID           string
	StoreID      string
	BusinessDate string
	Status       OperationalDayStatus
	OpenedByID   string
	ClosedByID   string
	OpenedAt     time.Time
	ClosedAt     time.Time
	UpdatedAt    time.Time
}

type OperationalDayBlocker struct {
	Code        string
	Severity    OperationalDayBlockerSeverity
	Message     string
	ReferenceID string
}

type OpenOperationalDayInput struct {
	ID           string
	StoreID      string
	BusinessDate string
	OpenedByID   string
	Now          time.Time
}

func OpenOperationalDay(input OpenOperationalDayInput) (OperationalDay, error) {
	if input.ID == "" || input.StoreID == "" || input.BusinessDate == "" || input.OpenedByID == "" {
		return OperationalDay{}, ErrInvalidOperationalDayInput
	}
	if input.Now.IsZero() {
		input.Now = time.Now().UTC()
	}

	return OperationalDay{
		ID:           input.ID,
		StoreID:      input.StoreID,
		BusinessDate: input.BusinessDate,
		Status:       OperationalDayStatusOpen,
		OpenedByID:   input.OpenedByID,
		OpenedAt:     input.Now,
		UpdatedAt:    input.Now,
	}, nil
}

func (d *OperationalDay) Close(closedByID string, now time.Time) error {
	if d.Status != OperationalDayStatusOpen {
		return ErrOperationalDayNotOpen
	}
	if closedByID == "" {
		return ErrInvalidOperationalDayInput
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	d.Status = OperationalDayStatusClosed
	d.ClosedByID = closedByID
	d.ClosedAt = now
	d.UpdatedAt = now
	return nil
}

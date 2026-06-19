package domain

import (
	"errors"
	"time"
)

var ErrInvalidSyncedOperationalDayInput = errors.New("invalid synced operational day input")

type SyncedOperationalDay struct {
	ID            string
	StoreID       string
	BusinessDate  string
	ClosedByID    string
	ClosedAt      time.Time
	SourceEventID string
	SyncedAt      time.Time
}

func NewSyncedOperationalDay(day SyncedOperationalDay) (SyncedOperationalDay, error) {
	if day.ID == "" || day.StoreID == "" || day.BusinessDate == "" ||
		day.ClosedByID == "" || day.SourceEventID == "" {
		return SyncedOperationalDay{}, ErrInvalidSyncedOperationalDayInput
	}
	if day.ClosedAt.IsZero() {
		return SyncedOperationalDay{}, ErrInvalidSyncedOperationalDayInput
	}
	return day, nil
}

package domain

import (
	"errors"
	"time"
)

var ErrInvalidSyncedReturnInput = errors.New("invalid synced return input")

type SyncedReturn struct {
	ID             string
	StoreID        string
	ReceiptID      string
	TotalMinor     int64
	PaymentIDs     []string
	CashMovementID string
	ActorID        string
	SettledAt      time.Time
	SourceEventID  string
	SyncedAt       time.Time
}

func NewSyncedReturn(ret SyncedReturn) (SyncedReturn, error) {
	if ret.ID == "" || ret.StoreID == "" || ret.ReceiptID == "" ||
		ret.ActorID == "" || ret.TotalMinor < 0 || ret.SourceEventID == "" {
		return SyncedReturn{}, ErrInvalidSyncedReturnInput
	}
	if ret.SettledAt.IsZero() {
		return SyncedReturn{}, ErrInvalidSyncedReturnInput
	}
	if ret.PaymentIDs == nil {
		ret.PaymentIDs = []string{}
	}
	return ret, nil
}

package domain

import (
	"errors"
	"time"
)

var ErrInvalidSyncedPaymentInput = errors.New("invalid synced payment input")

type SyncedPayment struct {
	ID            string
	StoreID       string
	ReceiptID     string
	Method        string
	AmountMinor   int64
	CapturedAt    time.Time
	SourceEventID string
	SyncedAt      time.Time
}

func NewSyncedPayment(payment SyncedPayment) (SyncedPayment, error) {
	if payment.ID == "" || payment.StoreID == "" || payment.ReceiptID == "" ||
		payment.Method == "" || payment.AmountMinor < 0 || payment.SourceEventID == "" {
		return SyncedPayment{}, ErrInvalidSyncedPaymentInput
	}
	if payment.CapturedAt.IsZero() {
		return SyncedPayment{}, ErrInvalidSyncedPaymentInput
	}
	return payment, nil
}

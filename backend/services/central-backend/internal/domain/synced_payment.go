package domain

import (
	"errors"
	"time"
)

var ErrInvalidSyncedPaymentInput = errors.New("invalid synced payment input")

type SyncedPaymentStatus string

const (
	SyncedPaymentStatusCaptured          SyncedPaymentStatus = "captured"
	SyncedPaymentStatusCancelled         SyncedPaymentStatus = "cancelled"
	SyncedPaymentStatusRefunded          SyncedPaymentStatus = "refunded"
	SyncedPaymentStatusPartiallyRefunded SyncedPaymentStatus = "partially_refunded"
)

type SyncedPayment struct {
	ID                   string
	StoreID              string
	ReceiptID            string
	Method               string
	AmountMinor          int64
	Status               SyncedPaymentStatus
	CapturedAt           time.Time
	CancelledAt          *time.Time
	RefundedAmountMinor  int64
	RemainingAmountMinor int64
	SourceEventID        string
	LastEventID          string
	SyncedAt             time.Time
	UpdatedAt            time.Time
}

func NewSyncedPayment(payment SyncedPayment) (SyncedPayment, error) {
	if payment.ID == "" || payment.StoreID == "" || payment.ReceiptID == "" ||
		payment.Method == "" || payment.AmountMinor < 0 || payment.LastEventID == "" {
		return SyncedPayment{}, ErrInvalidSyncedPaymentInput
	}
	if payment.Status == "" {
		payment.Status = SyncedPaymentStatusCaptured
	}
	switch payment.Status {
	case SyncedPaymentStatusCaptured:
		if payment.CapturedAt.IsZero() {
			return SyncedPayment{}, ErrInvalidSyncedPaymentInput
		}
	case SyncedPaymentStatusCancelled:
		if payment.CancelledAt == nil || payment.CancelledAt.IsZero() {
			return SyncedPayment{}, ErrInvalidSyncedPaymentInput
		}
	case SyncedPaymentStatusRefunded, SyncedPaymentStatusPartiallyRefunded:
		if payment.RefundedAmountMinor < 0 || payment.RemainingAmountMinor < 0 {
			return SyncedPayment{}, ErrInvalidSyncedPaymentInput
		}
		if payment.UpdatedAt.IsZero() {
			return SyncedPayment{}, ErrInvalidSyncedPaymentInput
		}
	default:
		return SyncedPayment{}, ErrInvalidSyncedPaymentInput
	}
	if payment.SourceEventID == "" {
		payment.SourceEventID = payment.LastEventID
	}
	return payment, nil
}

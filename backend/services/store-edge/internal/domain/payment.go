package domain

import (
	"errors"
	"time"
)

type PaymentMethod string
type PaymentStatus string

const (
	PaymentMethodCash     PaymentMethod = "cash"
	PaymentMethodCardMock PaymentMethod = "card_mock"
	PaymentStatusCaptured PaymentStatus = "captured"
	PaymentStatusCancelled PaymentStatus = "cancelled"
)

var (
	ErrInvalidPaymentInput      = errors.New("invalid payment input")
	ErrPaymentCannotBeCancelled = errors.New("payment cannot be cancelled")
)

type Payment struct {
	ID                string
	ReceiptID         string
	Method            PaymentMethod
	Status            PaymentStatus
	AmountMinor       int64
	ProviderReference string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	CapturedAt        time.Time
}

type CreateCapturedPaymentInput struct {
	ID                string
	ReceiptID         string
	Method            PaymentMethod
	AmountMinor       int64
	ProviderReference string
	Now               time.Time
}

func CreateCapturedPayment(input CreateCapturedPaymentInput) (Payment, error) {
	if input.ID == "" || input.ReceiptID == "" || input.Method == "" || input.AmountMinor <= 0 {
		return Payment{}, ErrInvalidPaymentInput
	}
	if input.Now.IsZero() {
		input.Now = time.Now().UTC()
	}

	return Payment{
		ID:                input.ID,
		ReceiptID:         input.ReceiptID,
		Method:            input.Method,
		Status:            PaymentStatusCaptured,
		AmountMinor:       input.AmountMinor,
		ProviderReference: input.ProviderReference,
		CreatedAt:         input.Now,
		UpdatedAt:         input.Now,
		CapturedAt:        input.Now,
	}, nil
}

func (p *Payment) Cancel(now time.Time) error {
	if p.Status != PaymentStatusCaptured {
		return ErrPaymentCannotBeCancelled
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	p.Status = PaymentStatusCancelled
	p.UpdatedAt = now
	return nil
}

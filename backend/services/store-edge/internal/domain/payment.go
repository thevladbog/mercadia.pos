package domain

import (
	"errors"
	"time"
)

type PaymentMethod string
type PaymentStatus string

const (
	PaymentMethodCash              PaymentMethod = "cash"
	PaymentMethodCardMock          PaymentMethod = "card_mock"
	PaymentStatusCaptured          PaymentStatus = "captured"
	PaymentStatusPartiallyRefunded PaymentStatus = "partially_refunded"
	PaymentStatusCancelled         PaymentStatus = "cancelled"
	PaymentStatusRefunded          PaymentStatus = "refunded"
)

var (
	ErrInvalidPaymentInput         = errors.New("invalid payment input")
	ErrPaymentCannotBeCancelled    = errors.New("payment cannot be cancelled")
	ErrPaymentCannotBeRefunded     = errors.New("payment cannot be refunded")
	ErrPaymentRefundAmountInvalid  = errors.New("payment refund amount is invalid")
)

type Payment struct {
	ID                  string
	ReceiptID           string
	Method              PaymentMethod
	Status              PaymentStatus
	AmountMinor         int64
	RefundedAmountMinor int64
	ProviderReference   string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	CapturedAt          time.Time
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

func (p Payment) RefundableAmountMinor() int64 {
	switch p.Status {
	case PaymentStatusCaptured:
		return p.AmountMinor
	case PaymentStatusPartiallyRefunded:
		return p.AmountMinor - p.RefundedAmountMinor
	default:
		return 0
	}
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

func (p *Payment) Refund(now time.Time) error {
	return p.RefundAmount(p.RefundableAmountMinor(), now)
}

func (p *Payment) RefundAmount(amountMinor int64, now time.Time) error {
	if p.Status != PaymentStatusCaptured && p.Status != PaymentStatusPartiallyRefunded {
		return ErrPaymentCannotBeRefunded
	}
	if amountMinor <= 0 || amountMinor > p.RefundableAmountMinor() {
		return ErrPaymentRefundAmountInvalid
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	p.RefundedAmountMinor += amountMinor
	if p.RefundedAmountMinor >= p.AmountMinor {
		p.Status = PaymentStatusRefunded
	} else {
		p.Status = PaymentStatusPartiallyRefunded
	}
	p.UpdatedAt = now
	return nil
}

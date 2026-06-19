package domain

import (
	"errors"
	"time"
)

type FiscalDocumentStatus string
type FiscalDocumentKind string

const (
	FiscalDocumentKindReceipt      FiscalDocumentKind   = "receipt"
	FiscalDocumentKindReturn       FiscalDocumentKind   = "return"
	FiscalDocumentStatusFiscalized FiscalDocumentStatus = "fiscalized"
)

var ErrInvalidFiscalDocumentInput = errors.New("invalid fiscal document input")

type FiscalDocument struct {
	ID           string
	ReceiptID    string
	ReturnID     string
	Kind         FiscalDocumentKind
	Status       FiscalDocumentStatus
	AmountMinor  int64
	DeviceID     string
	FiscalSign   string
	FiscalizedAt time.Time
	CreatedAt    time.Time
}

type CreateFiscalizedDocumentInput struct {
	ID          string
	ReceiptID   string
	ReturnID    string
	Kind        FiscalDocumentKind
	AmountMinor int64
	DeviceID    string
	FiscalSign  string
	Now         time.Time
}

func CreateFiscalizedDocument(input CreateFiscalizedDocumentInput) (FiscalDocument, error) {
	if input.ID == "" || input.ReceiptID == "" || input.Kind == "" || input.AmountMinor <= 0 || input.DeviceID == "" || input.FiscalSign == "" {
		return FiscalDocument{}, ErrInvalidFiscalDocumentInput
	}
	switch input.Kind {
	case FiscalDocumentKindReceipt:
		if input.ReturnID != "" {
			return FiscalDocument{}, ErrInvalidFiscalDocumentInput
		}
	case FiscalDocumentKindReturn:
		if input.ReturnID == "" {
			return FiscalDocument{}, ErrInvalidFiscalDocumentInput
		}
	default:
		return FiscalDocument{}, ErrInvalidFiscalDocumentInput
	}
	if input.Now.IsZero() {
		input.Now = time.Now().UTC()
	}

	return FiscalDocument{
		ID:           input.ID,
		ReceiptID:    input.ReceiptID,
		ReturnID:     input.ReturnID,
		Kind:         input.Kind,
		Status:       FiscalDocumentStatusFiscalized,
		AmountMinor:  input.AmountMinor,
		DeviceID:     input.DeviceID,
		FiscalSign:   input.FiscalSign,
		FiscalizedAt: input.Now,
		CreatedAt:    input.Now,
	}, nil
}

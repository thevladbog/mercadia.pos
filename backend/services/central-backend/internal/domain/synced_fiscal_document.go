package domain

import (
	"errors"
	"time"
)

var ErrInvalidSyncedFiscalDocumentInput = errors.New("invalid synced fiscal document input")

type SyncedFiscalDocument struct {
	ID            string
	StoreID       string
	ReceiptID     string
	Kind          string
	AmountMinor   int64
	DeviceID      string
	FiscalSign    string
	FiscalizedAt  time.Time
	ReturnID      string
	SourceEventID string
	SyncedAt      time.Time
}

func NewSyncedFiscalDocument(document SyncedFiscalDocument) (SyncedFiscalDocument, error) {
	if document.ID == "" || document.StoreID == "" || document.ReceiptID == "" ||
		document.Kind == "" || document.DeviceID == "" || document.FiscalSign == "" ||
		document.AmountMinor < 0 || document.SourceEventID == "" {
		return SyncedFiscalDocument{}, ErrInvalidSyncedFiscalDocumentInput
	}
	if document.FiscalizedAt.IsZero() {
		return SyncedFiscalDocument{}, ErrInvalidSyncedFiscalDocumentInput
	}
	return document, nil
}

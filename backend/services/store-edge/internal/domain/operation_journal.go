package domain

import (
	"errors"
	"time"
)

var ErrInvalidOperationJournalInput = errors.New("invalid operation journal input")

type OperationJournalEntry struct {
	ID            string
	StoreID       string
	OperationType string
	ActorID       string
	ReferenceID   string
	Summary       string
	CreatedAt     time.Time
}

type CreateOperationJournalEntryInput struct {
	ID            string
	StoreID       string
	OperationType string
	ActorID       string
	ReferenceID   string
	Summary       string
	Now           time.Time
}

func NewOperationJournalEntry(input CreateOperationJournalEntryInput) (OperationJournalEntry, error) {
	if input.ID == "" || input.StoreID == "" || input.OperationType == "" || input.ActorID == "" {
		return OperationJournalEntry{}, ErrInvalidOperationJournalInput
	}
	if input.Now.IsZero() {
		input.Now = time.Now().UTC()
	}
	return OperationJournalEntry{
		ID:            input.ID,
		StoreID:       input.StoreID,
		OperationType: input.OperationType,
		ActorID:       input.ActorID,
		ReferenceID:   input.ReferenceID,
		Summary:       input.Summary,
		CreatedAt:     input.Now,
	}, nil
}

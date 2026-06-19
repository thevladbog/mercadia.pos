package app

import (
	"context"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/domain"
)

type OperationJournalRepository interface {
	SaveOperationJournalEntry(ctx context.Context, entry domain.OperationJournalEntry) error
	ListOperationJournalEntries(ctx context.Context, storeID string, params PageParams) (PageResult[domain.OperationJournalEntry], error)
}

type OperationJournalRecorder interface {
	RecordOperation(ctx context.Context, command RecordOperationCommand) error
}

type OperationJournalService struct {
	journal OperationJournalRepository
	now     func() time.Time
	newID   func(prefix string) string
}

func NewOperationJournalService(journal OperationJournalRepository) *OperationJournalService {
	return &OperationJournalService{
		journal: journal,
		now: func() time.Time {
			return time.Now().UTC()
		},
		newID: randomID,
	}
}

type RecordOperationCommand struct {
	StoreID       string
	OperationType string
	ActorID       string
	ReferenceID   string
	Summary       string
}

func (s *OperationJournalService) RecordOperation(ctx context.Context, command RecordOperationCommand) error {
	entry, err := domain.NewOperationJournalEntry(domain.CreateOperationJournalEntryInput{
		ID:            s.newID("oj"),
		StoreID:       command.StoreID,
		OperationType: command.OperationType,
		ActorID:       command.ActorID,
		ReferenceID:   command.ReferenceID,
		Summary:       command.Summary,
		Now:           s.now(),
	})
	if err != nil {
		return err
	}
	return s.journal.SaveOperationJournalEntry(ctx, entry)
}

func (s *OperationJournalService) ListOperationJournal(ctx context.Context, storeID string, params PageParams) (PageResult[domain.OperationJournalEntry], error) {
	if storeID == "" {
		return PageResult[domain.OperationJournalEntry]{}, ErrInvalidCashMovementCommand
	}
	return s.journal.ListOperationJournalEntries(ctx, storeID, params)
}

package app

import (
	"context"
	"errors"

	"mercadia.dev/pos/services/central-backend/internal/domain"
)

var (
	ErrOperationalDayNotFound     = errors.New("operational day not found")
	ErrInvalidOperationalDayQuery = errors.New("invalid operational day query")
)

type SyncedOperationalDayRepository interface {
	SaveOperationalDay(ctx context.Context, day domain.SyncedOperationalDay) error
	FindOperationalDay(ctx context.Context, storeID string, operationalDayID string) (domain.SyncedOperationalDay, error)
	ListOperationalDays(ctx context.Context, storeID string, limit, offset int) ([]domain.SyncedOperationalDay, int, error)
}

type OperationalDaysService struct {
	stores          StoreRepository
	operationalDays SyncedOperationalDayRepository
}

func NewOperationalDaysService(stores StoreRepository, operationalDays SyncedOperationalDayRepository) *OperationalDaysService {
	return &OperationalDaysService{
		stores:          stores,
		operationalDays: operationalDays,
	}
}

func (s *OperationalDaysService) ListOperationalDays(ctx context.Context, storeID string, params PageParams) (PageResult[domain.SyncedOperationalDay], error) {
	if storeID == "" {
		return PageResult[domain.SyncedOperationalDay]{}, ErrInvalidOperationalDayQuery
	}
	if _, err := s.stores.FindStore(ctx, storeID); err != nil {
		return PageResult[domain.SyncedOperationalDay]{}, err
	}
	days, total, err := s.operationalDays.ListOperationalDays(ctx, storeID, params.Limit, params.Offset)
	if err != nil {
		return PageResult[domain.SyncedOperationalDay]{}, err
	}
	return PageResult[domain.SyncedOperationalDay]{Items: days, TotalCount: total}, nil
}

func (s *OperationalDaysService) GetOperationalDay(ctx context.Context, storeID string, operationalDayID string) (domain.SyncedOperationalDay, error) {
	if storeID == "" || operationalDayID == "" {
		return domain.SyncedOperationalDay{}, ErrInvalidOperationalDayQuery
	}
	if _, err := s.stores.FindStore(ctx, storeID); err != nil {
		return domain.SyncedOperationalDay{}, err
	}
	return s.operationalDays.FindOperationalDay(ctx, storeID, operationalDayID)
}

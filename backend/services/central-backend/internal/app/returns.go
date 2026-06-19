package app

import (
	"context"
	"errors"

	"mercadia.dev/pos/services/central-backend/internal/domain"
)

var (
	ErrReturnNotFound     = errors.New("return not found")
	ErrInvalidReturnQuery = errors.New("invalid return query")
)

type SyncedReturnRepository interface {
	SaveReturn(ctx context.Context, ret domain.SyncedReturn) error
	FindReturn(ctx context.Context, storeID string, returnID string) (domain.SyncedReturn, error)
	ListReturns(ctx context.Context, storeID string, limit, offset int) ([]domain.SyncedReturn, int, error)
}

type ReturnsService struct {
	stores  StoreRepository
	returns SyncedReturnRepository
}

func NewReturnsService(stores StoreRepository, returns SyncedReturnRepository) *ReturnsService {
	return &ReturnsService{
		stores:  stores,
		returns: returns,
	}
}

func (s *ReturnsService) ListReturns(ctx context.Context, storeID string, params PageParams) (PageResult[domain.SyncedReturn], error) {
	if storeID == "" {
		return PageResult[domain.SyncedReturn]{}, ErrInvalidReturnQuery
	}
	if _, err := s.stores.FindStore(ctx, storeID); err != nil {
		return PageResult[domain.SyncedReturn]{}, err
	}
	returns, total, err := s.returns.ListReturns(ctx, storeID, params.Limit, params.Offset)
	if err != nil {
		return PageResult[domain.SyncedReturn]{}, err
	}
	return PageResult[domain.SyncedReturn]{Items: returns, TotalCount: total}, nil
}

func (s *ReturnsService) GetReturn(ctx context.Context, storeID string, returnID string) (domain.SyncedReturn, error) {
	if storeID == "" || returnID == "" {
		return domain.SyncedReturn{}, ErrInvalidReturnQuery
	}
	if _, err := s.stores.FindStore(ctx, storeID); err != nil {
		return domain.SyncedReturn{}, err
	}
	return s.returns.FindReturn(ctx, storeID, returnID)
}

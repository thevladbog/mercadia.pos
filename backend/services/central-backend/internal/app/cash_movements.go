package app

import (
	"context"
	"errors"

	"mercadia.dev/pos/services/central-backend/internal/domain"
)

var (
	ErrCashMovementNotFound     = errors.New("cash movement not found")
	ErrInvalidCashMovementQuery = errors.New("invalid cash movement query")
)

type SyncedCashMovementRepository interface {
	SaveCashMovement(ctx context.Context, movement domain.SyncedCashMovement) error
	FindCashMovement(ctx context.Context, storeID string, cashMovementID string) (domain.SyncedCashMovement, error)
	ListCashMovements(ctx context.Context, storeID string, limit, offset int) ([]domain.SyncedCashMovement, int, error)
}

type CashMovementsService struct {
	stores        StoreRepository
	cashMovements SyncedCashMovementRepository
}

func NewCashMovementsService(stores StoreRepository, cashMovements SyncedCashMovementRepository) *CashMovementsService {
	return &CashMovementsService{
		stores:        stores,
		cashMovements: cashMovements,
	}
}

func (s *CashMovementsService) ListCashMovements(ctx context.Context, storeID string, params PageParams) (PageResult[domain.SyncedCashMovement], error) {
	if storeID == "" {
		return PageResult[domain.SyncedCashMovement]{}, ErrInvalidCashMovementQuery
	}
	if _, err := s.stores.FindStore(ctx, storeID); err != nil {
		return PageResult[domain.SyncedCashMovement]{}, err
	}
	movements, total, err := s.cashMovements.ListCashMovements(ctx, storeID, params.Limit, params.Offset)
	if err != nil {
		return PageResult[domain.SyncedCashMovement]{}, err
	}
	return PageResult[domain.SyncedCashMovement]{Items: movements, TotalCount: total}, nil
}

func (s *CashMovementsService) GetCashMovement(ctx context.Context, storeID string, cashMovementID string) (domain.SyncedCashMovement, error) {
	if storeID == "" || cashMovementID == "" {
		return domain.SyncedCashMovement{}, ErrInvalidCashMovementQuery
	}
	if _, err := s.stores.FindStore(ctx, storeID); err != nil {
		return domain.SyncedCashMovement{}, err
	}
	return s.cashMovements.FindCashMovement(ctx, storeID, cashMovementID)
}

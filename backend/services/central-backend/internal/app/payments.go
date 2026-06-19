package app

import (
	"context"
	"errors"

	"mercadia.dev/pos/services/central-backend/internal/domain"
)

var (
	ErrPaymentNotFound    = errors.New("payment not found")
	ErrInvalidPaymentQuery = errors.New("invalid payment query")
)

type SyncedPaymentRepository interface {
	SavePayment(ctx context.Context, payment domain.SyncedPayment) error
	FindPayment(ctx context.Context, storeID string, paymentID string) (domain.SyncedPayment, error)
	ListPayments(ctx context.Context, storeID string, limit, offset int) ([]domain.SyncedPayment, int, error)
}

type PaymentsService struct {
	stores   StoreRepository
	payments SyncedPaymentRepository
}

func NewPaymentsService(stores StoreRepository, payments SyncedPaymentRepository) *PaymentsService {
	return &PaymentsService{
		stores:   stores,
		payments: payments,
	}
}

func (s *PaymentsService) ListPayments(ctx context.Context, storeID string, params PageParams) (PageResult[domain.SyncedPayment], error) {
	if storeID == "" {
		return PageResult[domain.SyncedPayment]{}, ErrInvalidPaymentQuery
	}
	if _, err := s.stores.FindStore(ctx, storeID); err != nil {
		return PageResult[domain.SyncedPayment]{}, err
	}
	payments, total, err := s.payments.ListPayments(ctx, storeID, params.Limit, params.Offset)
	if err != nil {
		return PageResult[domain.SyncedPayment]{}, err
	}
	return PageResult[domain.SyncedPayment]{Items: payments, TotalCount: total}, nil
}

func (s *PaymentsService) GetPayment(ctx context.Context, storeID string, paymentID string) (domain.SyncedPayment, error) {
	if storeID == "" || paymentID == "" {
		return domain.SyncedPayment{}, ErrInvalidPaymentQuery
	}
	if _, err := s.stores.FindStore(ctx, storeID); err != nil {
		return domain.SyncedPayment{}, err
	}
	return s.payments.FindPayment(ctx, storeID, paymentID)
}

package app

import (
	"context"
	"errors"
	"time"

	"mercadia.dev/pos/services/central-backend/internal/domain"
)

var (
	ErrCatalogProductNotFound = errors.New("catalog product not found")
	ErrInvalidCatalogQuery    = errors.New("invalid catalog query")
)

type CatalogProductRepository interface {
	SaveProduct(ctx context.Context, product domain.CatalogProduct) error
	FindProduct(ctx context.Context, storeID string, productID string) (domain.CatalogProduct, error)
	ListProducts(ctx context.Context, storeID string) ([]domain.CatalogProduct, error)
	ListProductsSince(ctx context.Context, storeID string, since time.Time) ([]domain.CatalogProduct, error)
}

type CatalogService struct {
	stores   StoreRepository
	products CatalogProductRepository
}

func NewCatalogService(stores StoreRepository, products CatalogProductRepository) *CatalogService {
	return &CatalogService{
		stores:   stores,
		products: products,
	}
}

type CatalogProductsResult struct {
	Products []domain.CatalogProduct
}

type CatalogDeltaResult struct {
	Since    time.Time
	Products []domain.CatalogProduct
}

func (s *CatalogService) ListProducts(ctx context.Context, storeID string) (CatalogProductsResult, error) {
	if storeID == "" {
		return CatalogProductsResult{}, ErrInvalidCatalogQuery
	}
	if _, err := s.stores.FindStore(ctx, storeID); err != nil {
		return CatalogProductsResult{}, err
	}
	products, err := s.products.ListProducts(ctx, storeID)
	if err != nil {
		return CatalogProductsResult{}, err
	}
	return CatalogProductsResult{Products: products}, nil
}

func (s *CatalogService) CatalogDelta(ctx context.Context, storeID string, since time.Time) (CatalogDeltaResult, error) {
	if storeID == "" || since.IsZero() {
		return CatalogDeltaResult{}, ErrInvalidCatalogQuery
	}
	if _, err := s.stores.FindStore(ctx, storeID); err != nil {
		return CatalogDeltaResult{}, err
	}
	products, err := s.products.ListProductsSince(ctx, storeID, since.UTC())
	if err != nil {
		return CatalogDeltaResult{}, err
	}
	return CatalogDeltaResult{
		Since:    since.UTC(),
		Products: products,
	}, nil
}

func (s *CatalogService) GetProduct(ctx context.Context, storeID string, productID string) (domain.CatalogProduct, error) {
	if storeID == "" || productID == "" {
		return domain.CatalogProduct{}, ErrInvalidCatalogQuery
	}
	if _, err := s.stores.FindStore(ctx, storeID); err != nil {
		return domain.CatalogProduct{}, err
	}
	return s.products.FindProduct(ctx, storeID, productID)
}

package app

import (
	"context"
	"errors"
	"time"

	"mercadia.dev/pos/services/store-edge/internal/domain"
	"mercadia.dev/pos/services/store-edge/internal/infra/central"
)

var (
	ErrInvalidCatalogSyncCommand = errors.New("invalid catalog sync command")
	ErrCatalogSyncUnavailable    = errors.New("catalog sync is unavailable")
)

type CatalogProductWriter interface {
	SaveProduct(ctx context.Context, product domain.Product) error
}

type CatalogSyncStateRepository interface {
	GetLastSyncedAt(ctx context.Context, storeID string) (time.Time, error)
	SaveLastSyncedAt(ctx context.Context, storeID string, syncedAt time.Time) error
}

type CatalogDeltaClient interface {
	CatalogDelta(ctx context.Context, storeID string, since time.Time) ([]central.CatalogProduct, time.Time, error)
}

type CatalogSyncService struct {
	products CatalogProductWriter
	state    CatalogSyncStateRepository
	central  CatalogDeltaClient
	now      func() time.Time
}

type CatalogSyncOption func(*CatalogSyncService)

func NewCatalogSyncService(products CatalogProductWriter, state CatalogSyncStateRepository, central CatalogDeltaClient, options ...CatalogSyncOption) *CatalogSyncService {
	service := &CatalogSyncService{
		products: products,
		state:    state,
		central:  central,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func WithCatalogSyncClock(now func() time.Time) CatalogSyncOption {
	return func(service *CatalogSyncService) {
		service.now = now
	}
}

type SyncCatalogCommand struct {
	StoreID string
}

type CatalogSyncResult struct {
	StoreID       string    `json:"storeId"`
	Since         time.Time `json:"since"`
	SyncedAt      time.Time `json:"syncedAt"`
	ProductsCount int       `json:"productsCount"`
}

func (s *CatalogSyncService) Sync(ctx context.Context, command SyncCatalogCommand) (CatalogSyncResult, error) {
	if command.StoreID == "" {
		return CatalogSyncResult{}, ErrInvalidCatalogSyncCommand
	}
	if s.central == nil {
		return CatalogSyncResult{}, ErrCatalogSyncUnavailable
	}

	since, err := s.state.GetLastSyncedAt(ctx, command.StoreID)
	if err != nil {
		return CatalogSyncResult{}, err
	}
	if since.IsZero() {
		since = time.Unix(0, 0).UTC()
	}

	products, deltaSince, err := s.central.CatalogDelta(ctx, command.StoreID, since)
	if err != nil {
		return CatalogSyncResult{}, err
	}

	syncedAt := s.now()
	maxUpdatedAt := syncedAt
	for _, product := range products {
		domainProduct, err := domain.NewProduct(domain.Product{
			ID:             product.ID,
			Name:           product.Name,
			Barcodes:       append([]string(nil), product.Barcodes...),
			UnitPriceMinor: product.UnitPriceMinor,
			TaxCategoryID:  product.TaxCategoryID,
			Active:         product.Active,
		})
		if err != nil {
			return CatalogSyncResult{}, err
		}
		if !product.Active {
			domainProduct.Active = false
		}
		if err := s.products.SaveProduct(ctx, domainProduct); err != nil {
			return CatalogSyncResult{}, err
		}
		if product.UpdatedAt.After(maxUpdatedAt) {
			maxUpdatedAt = product.UpdatedAt.UTC()
		}
	}

	if len(products) > 0 {
		syncedAt = maxUpdatedAt
	}
	if err := s.state.SaveLastSyncedAt(ctx, command.StoreID, syncedAt); err != nil {
		return CatalogSyncResult{}, err
	}

	resultSince := deltaSince
	if resultSince.IsZero() {
		resultSince = since
	}

	return CatalogSyncResult{
		StoreID:       command.StoreID,
		Since:         resultSince,
		SyncedAt:      syncedAt,
		ProductsCount: len(products),
	}, nil
}

package app

import (
	"context"
	"errors"

	"mercadia.dev/pos/services/store-edge/internal/domain"
)

var (
	ErrProductNotFound     = errors.New("product not found")
	ErrInvalidCatalogQuery = errors.New("invalid catalog query")
)

type ProductRepository interface {
	FindProductByBarcode(ctx context.Context, barcode string) (domain.Product, error)
}

type CatalogService struct {
	products ProductRepository
}

func NewCatalogService(products ProductRepository) *CatalogService {
	return &CatalogService{products: products}
}

type ProductResult struct {
	Product domain.Product
}

func (s *CatalogService) FindProductByBarcode(ctx context.Context, barcode string) (ProductResult, error) {
	if barcode == "" {
		return ProductResult{}, ErrInvalidCatalogQuery
	}
	product, err := s.products.FindProductByBarcode(ctx, barcode)
	if err != nil {
		return ProductResult{}, err
	}
	return ProductResult{Product: product}, nil
}

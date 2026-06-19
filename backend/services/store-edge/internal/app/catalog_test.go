package app_test

import (
	"context"
	"testing"

	"mercadia.dev/pos/services/store-edge/internal/app"
	"mercadia.dev/pos/services/store-edge/internal/domain"
	"mercadia.dev/pos/services/store-edge/internal/infra/memory"
)

func TestFindProductByBarcode(t *testing.T) {
	store := memory.NewStore(memory.WithProducts(testProduct()))
	service := app.NewCatalogService(store)

	result, err := service.FindProductByBarcode(context.Background(), "4600000000000")
	if err != nil {
		t.Fatalf("find product: %v", err)
	}

	if result.Product.ID != "sku-1" {
		t.Fatalf("product id = %s", result.Product.ID)
	}
}

func testProduct() domain.Product {
	product, err := domain.NewProduct(domain.Product{
		ID:             "sku-1",
		Name:           "Milk",
		Barcodes:       []string{"4600000000000"},
		UnitPriceMinor: 19999,
		TaxCategoryID:  "vat_20",
	})
	if err != nil {
		panic(err)
	}
	return product
}

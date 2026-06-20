package app_test

import (
	"context"
	"errors"
	"testing"

	"mercadia.dev/pos/services/central-backend/internal/app"
	"mercadia.dev/pos/services/central-backend/internal/domain"
	"mercadia.dev/pos/services/central-backend/internal/infra/memory"
)

func TestLayoutTemplatePublishRequiresStoreScope(t *testing.T) {
	store := memory.NewStore()
	seedCentralAdmin(t, store)
	auth := app.NewAuthService(store, store)
	layoutTemplates := app.NewLayoutTemplatesService(store, store, store)
	adminSession, err := auth.CreateSession(context.Background(), app.CreateSessionCommand{
		Email:    "admin@example.com",
		Password: "admin-pass",
	})
	if err != nil {
		t.Fatalf("create admin session: %v", err)
	}

	_, err = layoutTemplates.CreateLayoutTemplate(context.Background(), app.CreateLayoutTemplateCommand{
		TemplateID: "published-no-store",
		Name:       "Published No Store",
		Kind:       domain.LayoutTemplateKindSale,
		Status:     domain.LayoutTemplateStatusPublished,
		Grid: domain.LayoutGrid{
			Rows: 2,
			Cols: 2,
			Tiles: []domain.LayoutGridTile{
				{Label: "Coffee", ProductID: "sku-1"},
			},
		},
		Session: adminSession,
	})
	if !errors.Is(err, app.ErrLayoutTemplatePublishRequiresStore) {
		t.Fatalf("expected publish requires store, got %v", err)
	}
}

func TestLayoutTemplatePublishValidatesProducts(t *testing.T) {
	store := memory.NewStore()
	seedCentralAdmin(t, store)
	auth := app.NewAuthService(store, store)
	layoutTemplates := app.NewLayoutTemplatesService(store, store, store)
	adminSession, err := auth.CreateSession(context.Background(), app.CreateSessionCommand{
		Email:    "admin@example.com",
		Password: "admin-pass",
	})
	if err != nil {
		t.Fatalf("create admin session: %v", err)
	}

	product, err := domain.NewCatalogProduct(domain.CatalogProduct{
		ID:             "sku-1",
		StoreID:        "store-1",
		Name:           "Coffee",
		Barcodes:       []string{"4600000000001"},
		UnitPriceMinor: 19999,
		TaxCategoryID:  "vat_20",
		Active:         true,
	})
	if err != nil {
		t.Fatalf("new product: %v", err)
	}
	if err := store.SaveProduct(context.Background(), product); err != nil {
		t.Fatalf("save product: %v", err)
	}

	_, err = layoutTemplates.CreateLayoutTemplate(context.Background(), app.CreateLayoutTemplateCommand{
		TemplateID: "published-bad-product",
		Name:       "Published Bad Product",
		Kind:       domain.LayoutTemplateKindSale,
		Status:     domain.LayoutTemplateStatusPublished,
		StoreID:    "store-1",
		Grid: domain.LayoutGrid{
			Rows: 2,
			Cols: 2,
			Tiles: []domain.LayoutGridTile{
				{Label: "Missing", ProductID: "sku-missing"},
			},
		},
		Session: adminSession,
	})
	if !errors.Is(err, app.ErrLayoutTemplateInvalidProducts) {
		t.Fatalf("expected invalid products, got %v", err)
	}

	result, err := layoutTemplates.CreateLayoutTemplate(context.Background(), app.CreateLayoutTemplateCommand{
		TemplateID: "published-valid",
		Name:       "Published Valid",
		Kind:       domain.LayoutTemplateKindSale,
		Status:     domain.LayoutTemplateStatusPublished,
		StoreID:    "store-1",
		Grid: domain.LayoutGrid{
			Rows: 2,
			Cols: 2,
			Tiles: []domain.LayoutGridTile{
				{Label: "Coffee", ProductID: "sku-1"},
			},
		},
		Session: adminSession,
	})
	if err != nil {
		t.Fatalf("create published template: %v", err)
	}
	if result.Template.Status != domain.LayoutTemplateStatusPublished {
		t.Fatalf("status = %s", result.Template.Status)
	}
}

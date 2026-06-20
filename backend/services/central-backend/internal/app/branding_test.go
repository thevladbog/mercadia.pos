package app_test

import (
	"context"
	"testing"

	"mercadia.dev/pos/services/central-backend/internal/app"
	"mercadia.dev/pos/services/central-backend/internal/domain"
	"mercadia.dev/pos/services/central-backend/internal/infra/memory"
)

func TestColorSchemesAndLayoutTemplatesCRUD(t *testing.T) {
	store := memory.NewStore()
	seedCentralAdmin(t, store)

	auth := app.NewAuthService(store, store)
	colorSchemes := app.NewColorSchemesService(store)
	layoutTemplates := app.NewLayoutTemplatesService(store, store, store)

	adminSession, err := auth.CreateSession(context.Background(), app.CreateSessionCommand{
		Email:    "admin@example.com",
		Password: "admin-pass",
	})
	if err != nil {
		t.Fatalf("create admin session: %v", err)
	}

	scheme, err := colorSchemes.CreateColorScheme(context.Background(), app.CreateColorSchemeCommand{
		SchemeID:     "brand-1",
		Name:         "Default Brand",
		AccentPreset: domain.AccentPresetNeutral,
		Status:       domain.ColorSchemeStatusPublished,
		Session:      adminSession,
	})
	if err != nil {
		t.Fatalf("create color scheme: %v", err)
	}
	if scheme.ResolvedAccentColor() != "#FF6600" {
		t.Fatalf("scheme accent = %s", scheme.ResolvedAccentColor())
	}

	saleTemplate, err := layoutTemplates.CreateLayoutTemplate(context.Background(), app.CreateLayoutTemplateCommand{
		TemplateID: "sale-standard",
		Name:       "Standard Sale",
		Kind:       domain.LayoutTemplateKindSale,
		Grid: domain.LayoutGrid{
			Rows: 2,
			Cols: 2,
			Tiles: []domain.LayoutGridTile{
				{Label: "Coffee"},
				{Label: "Tea"},
			},
		},
		Session: adminSession,
	})
	if err != nil {
		t.Fatalf("create sale template: %v", err)
	}
	if saleTemplate.ResolvedAccentPreset != domain.AccentPresetSale {
		t.Fatalf("sale preset = %s", saleTemplate.ResolvedAccentPreset)
	}
	if saleTemplate.ResolvedAccentColor != "#FF6600" {
		t.Fatalf("sale accent = %s", saleTemplate.ResolvedAccentColor)
	}

	returnTemplate, err := layoutTemplates.CreateLayoutTemplate(context.Background(), app.CreateLayoutTemplateCommand{
		TemplateID: "return-standard",
		Name:       "Return Register",
		Kind:       domain.LayoutTemplateKindReturn,
		Grid: domain.LayoutGrid{
			Rows: 2,
			Cols: 2,
		},
		Session: adminSession,
	})
	if err != nil {
		t.Fatalf("create return template: %v", err)
	}
	if returnTemplate.ResolvedAccentColor != "#2563EB" {
		t.Fatalf("return accent = %s", returnTemplate.ResolvedAccentColor)
	}

	linkedTemplate, err := layoutTemplates.CreateLayoutTemplate(context.Background(), app.CreateLayoutTemplateCommand{
		TemplateID:    "sale-franchise",
		Name:          "Franchise Sale",
		Kind:          domain.LayoutTemplateKindSale,
		ColorSchemeID: scheme.ID,
		Grid:          domain.LayoutGrid{Rows: 4, Cols: 4, Tiles: []domain.LayoutGridTile{}},
		Session:       adminSession,
	})
	if err != nil {
		t.Fatalf("create linked template: %v", err)
	}
	if linkedTemplate.ResolvedAccentColor != "#FF6600" {
		t.Fatalf("linked accent = %s", linkedTemplate.ResolvedAccentColor)
	}

	listed, err := layoutTemplates.ListLayoutTemplates(context.Background(), adminSession, app.LayoutTemplateListFilter{})
	if err != nil {
		t.Fatalf("list templates: %v", err)
	}
	if len(listed) != 3 {
		t.Fatalf("listed = %+v", listed)
	}
}

func TestLayoutTemplateCustomAccentOverridesPreset(t *testing.T) {
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

	result, err := layoutTemplates.CreateLayoutTemplate(context.Background(), app.CreateLayoutTemplateCommand{
		TemplateID:   "custom-return",
		Name:         "Custom Return",
		Kind:         domain.LayoutTemplateKindReturn,
		AccentColor:  "#112233",
		AccentPreset: domain.AccentPresetReturn,
		Grid:         domain.LayoutGrid{Rows: 1, Cols: 1},
		Session:      adminSession,
	})
	if err != nil {
		t.Fatalf("create template: %v", err)
	}
	if result.ResolvedAccentColor != "#112233" {
		t.Fatalf("accent = %s", result.ResolvedAccentColor)
	}
}

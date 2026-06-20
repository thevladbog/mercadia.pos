package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"mercadia.dev/pos/services/central-backend/internal/domain"
)

type LayoutTemplateRepository interface {
	SaveLayoutTemplate(ctx context.Context, template domain.LayoutTemplate) error
	FindLayoutTemplate(ctx context.Context, templateID string) (domain.LayoutTemplate, error)
	ListLayoutTemplates(ctx context.Context, filter LayoutTemplateListFilter) ([]domain.LayoutTemplate, error)
}

type LayoutTemplateListFilter struct {
	StoreID string
	Kind    domain.LayoutTemplateKind
}

type LayoutTemplatesService struct {
	templates LayoutTemplateRepository
	schemes   ColorSchemeRepository
	products  CatalogProductRepository
	now       func() time.Time
}

func NewLayoutTemplatesService(
	templates LayoutTemplateRepository,
	schemes ColorSchemeRepository,
	products CatalogProductRepository,
) *LayoutTemplatesService {
	return &LayoutTemplatesService{
		templates: templates,
		schemes:   schemes,
		products:  products,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (s *LayoutTemplatesService) validatePublishedGrid(
	ctx context.Context,
	storeID string,
	status domain.LayoutTemplateStatus,
	grid domain.LayoutGrid,
) error {
	if status != domain.LayoutTemplateStatusPublished {
		return nil
	}
	if storeID == "" {
		return ErrLayoutTemplatePublishRequiresStore
	}
	missing := make([]string, 0)
	for _, tile := range grid.Tiles {
		if tile.ProductID == "" {
			continue
		}
		product, err := s.products.FindProduct(ctx, storeID, tile.ProductID)
		if errors.Is(err, ErrCatalogProductNotFound) {
			missing = append(missing, tile.ProductID)
			continue
		}
		if err != nil {
			return err
		}
		if !product.Active {
			missing = append(missing, tile.ProductID)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("%w: %v", ErrLayoutTemplateInvalidProducts, missing)
	}
	return nil
}

type LayoutTemplateResult struct {
	Template             domain.LayoutTemplate
	ResolvedAccentPreset domain.AccentPreset
	ResolvedAccentColor  string
}

type CreateLayoutTemplateCommand struct {
	TemplateID    string
	Name          string
	Kind          domain.LayoutTemplateKind
	AccentPreset  domain.AccentPreset
	AccentColor   string
	ColorSchemeID string
	Grid          domain.LayoutGrid
	StoreID       string
	TerminalType  string
	Status        domain.LayoutTemplateStatus
	Session       SessionResult
}

type UpdateLayoutTemplateCommand struct {
	TemplateID    string
	Name          *string
	Kind          *domain.LayoutTemplateKind
	AccentPreset  *domain.AccentPreset
	AccentColor   *string
	ColorSchemeID *string
	Grid          *domain.LayoutGrid
	StoreID       *string
	TerminalType  *string
	Status        *domain.LayoutTemplateStatus
	Session       SessionResult
}

func (s *LayoutTemplatesService) resolveWithCtx(ctx context.Context, template domain.LayoutTemplate) (LayoutTemplateResult, error) {
	var scheme *domain.ColorScheme
	if template.ColorSchemeID != "" {
		found, err := s.schemes.FindColorScheme(ctx, template.ColorSchemeID)
		if err == nil {
			scheme = &found
		}
	}
	preset, color := template.ResolvedAccent(scheme)
	return LayoutTemplateResult{
		Template:             template,
		ResolvedAccentPreset: preset,
		ResolvedAccentColor:  color,
	}, nil
}

func (s *LayoutTemplatesService) ListLayoutTemplates(ctx context.Context, session SessionResult, filter LayoutTemplateListFilter) ([]LayoutTemplateResult, error) {
	if err := CheckCentralPermission(session.Roles, PermissionReportingRead); err != nil {
		return nil, err
	}
	templates, err := s.templates.ListLayoutTemplates(ctx, filter)
	if err != nil {
		return nil, err
	}
	results := make([]LayoutTemplateResult, 0, len(templates))
	for _, template := range templates {
		resolved, err := s.resolveWithCtx(ctx, template)
		if err != nil {
			return nil, err
		}
		results = append(results, resolved)
	}
	return results, nil
}

func (s *LayoutTemplatesService) GetLayoutTemplate(ctx context.Context, templateID string, session SessionResult) (LayoutTemplateResult, error) {
	if err := CheckCentralPermission(session.Roles, PermissionReportingRead); err != nil {
		return LayoutTemplateResult{}, err
	}
	template, err := s.templates.FindLayoutTemplate(ctx, templateID)
	if err != nil {
		return LayoutTemplateResult{}, err
	}
	return s.resolveWithCtx(ctx, template)
}

func (s *LayoutTemplatesService) CreateLayoutTemplate(ctx context.Context, command CreateLayoutTemplateCommand) (LayoutTemplateResult, error) {
	if err := CheckCentralPermission(command.Session.Roles, PermissionUsersManage); err != nil {
		return LayoutTemplateResult{}, err
	}
	if command.TemplateID == "" || command.Name == "" {
		return LayoutTemplateResult{}, ErrInvalidLayoutTemplateCmd
	}
	template, err := domain.NewLayoutTemplate(domain.LayoutTemplate{
		ID:            command.TemplateID,
		Name:          command.Name,
		Kind:          command.Kind,
		AccentPreset:  command.AccentPreset,
		AccentColor:   command.AccentColor,
		ColorSchemeID: command.ColorSchemeID,
		Grid:          command.Grid,
		StoreID:       command.StoreID,
		TerminalType:  command.TerminalType,
		Status:        command.Status,
		Version:       1,
		CreatedAt:     s.now(),
		UpdatedAt:     s.now(),
	})
	if err != nil {
		return LayoutTemplateResult{}, ErrInvalidLayoutTemplateCmd
	}
	if err := s.validatePublishedGrid(ctx, template.StoreID, template.Status, template.Grid); err != nil {
		return LayoutTemplateResult{}, err
	}
	if err := s.templates.SaveLayoutTemplate(ctx, template); err != nil {
		return LayoutTemplateResult{}, err
	}
	return s.resolveWithCtx(ctx, template)
}

func (s *LayoutTemplatesService) UpdateLayoutTemplate(ctx context.Context, command UpdateLayoutTemplateCommand) (LayoutTemplateResult, error) {
	if err := CheckCentralPermission(command.Session.Roles, PermissionUsersManage); err != nil {
		return LayoutTemplateResult{}, err
	}
	if command.TemplateID == "" {
		return LayoutTemplateResult{}, ErrInvalidLayoutTemplateCmd
	}
	template, err := s.templates.FindLayoutTemplate(ctx, command.TemplateID)
	if err != nil {
		return LayoutTemplateResult{}, err
	}
	if command.Name != nil {
		template.Name = *command.Name
	}
	if command.Kind != nil {
		template.Kind = *command.Kind
	}
	if command.AccentPreset != nil {
		template.AccentPreset = *command.AccentPreset
	}
	if command.AccentColor != nil {
		template.AccentColor = *command.AccentColor
	}
	if command.ColorSchemeID != nil {
		template.ColorSchemeID = *command.ColorSchemeID
	}
	if command.Grid != nil {
		template.Grid = *command.Grid
	}
	if command.StoreID != nil {
		template.StoreID = *command.StoreID
	}
	if command.TerminalType != nil {
		template.TerminalType = *command.TerminalType
	}
	if command.Status != nil {
		template.Status = *command.Status
	}
	template.UpdatedAt = s.now()

	validated, err := domain.NewLayoutTemplate(template)
	if err != nil {
		return LayoutTemplateResult{}, ErrInvalidLayoutTemplateCmd
	}
	if err := s.validatePublishedGrid(ctx, validated.StoreID, validated.Status, validated.Grid); err != nil {
		return LayoutTemplateResult{}, err
	}
	validated.CreatedAt = template.CreatedAt
	validated.Version = template.Version
	if err := s.templates.SaveLayoutTemplate(ctx, validated); err != nil {
		return LayoutTemplateResult{}, err
	}
	return s.resolveWithCtx(ctx, validated)
}

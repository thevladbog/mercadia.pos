package app

import (
	"context"
	"errors"
	"time"

	"mercadia.dev/pos/services/central-backend/internal/domain"
)

var (
	ErrColorSchemeNotFound                = errors.New("color scheme not found")
	ErrInvalidColorSchemeCommand          = errors.New("invalid color scheme command")
	ErrLayoutTemplateNotFound             = errors.New("layout template not found")
	ErrInvalidLayoutTemplateCmd           = errors.New("invalid layout template command")
	ErrLayoutTemplatePublishRequiresStore = errors.New("layout template publish requires store scope")
	ErrLayoutTemplateInvalidProducts      = errors.New("layout template references unknown products")
)

type ColorSchemeRepository interface {
	SaveColorScheme(ctx context.Context, scheme domain.ColorScheme) error
	FindColorScheme(ctx context.Context, schemeID string) (domain.ColorScheme, error)
	ListColorSchemes(ctx context.Context) ([]domain.ColorScheme, error)
}

type ColorSchemesService struct {
	schemes ColorSchemeRepository
	now     func() time.Time
}

func NewColorSchemesService(schemes ColorSchemeRepository) *ColorSchemesService {
	return &ColorSchemesService{
		schemes: schemes,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

type CreateColorSchemeCommand struct {
	SchemeID        string
	Name            string
	LogoURL         string
	AccentPreset    domain.AccentPreset
	AccentColor     string
	BackgroundColor string
	Status          domain.ColorSchemeStatus
	StoreIDs        []string
	Session         SessionResult
}

type UpdateColorSchemeCommand struct {
	SchemeID        string
	Name            *string
	LogoURL         *string
	AccentPreset    *domain.AccentPreset
	AccentColor     *string
	BackgroundColor *string
	Status          *domain.ColorSchemeStatus
	StoreIDs        []string
	Session         SessionResult
}

func (s *ColorSchemesService) ListColorSchemes(ctx context.Context, session SessionResult) ([]domain.ColorScheme, error) {
	if err := CheckCentralPermission(session.Roles, PermissionReportingRead); err != nil {
		return nil, err
	}
	return s.schemes.ListColorSchemes(ctx)
}

func (s *ColorSchemesService) GetColorScheme(ctx context.Context, schemeID string, session SessionResult) (domain.ColorScheme, error) {
	if err := CheckCentralPermission(session.Roles, PermissionReportingRead); err != nil {
		return domain.ColorScheme{}, err
	}
	return s.schemes.FindColorScheme(ctx, schemeID)
}

func (s *ColorSchemesService) CreateColorScheme(ctx context.Context, command CreateColorSchemeCommand) (domain.ColorScheme, error) {
	if err := CheckCentralPermission(command.Session.Roles, PermissionUsersManage); err != nil {
		return domain.ColorScheme{}, err
	}
	if command.SchemeID == "" || command.Name == "" {
		return domain.ColorScheme{}, ErrInvalidColorSchemeCommand
	}
	scheme, err := domain.NewColorScheme(domain.ColorScheme{
		ID:              command.SchemeID,
		Name:            command.Name,
		LogoURL:         command.LogoURL,
		AccentPreset:    command.AccentPreset,
		AccentColor:     command.AccentColor,
		BackgroundColor: command.BackgroundColor,
		Status:          command.Status,
		StoreIDs:        append([]string(nil), command.StoreIDs...),
		CreatedAt:       s.now(),
		UpdatedAt:       s.now(),
	})
	if err != nil {
		return domain.ColorScheme{}, ErrInvalidColorSchemeCommand
	}
	if err := s.schemes.SaveColorScheme(ctx, scheme); err != nil {
		return domain.ColorScheme{}, err
	}
	return scheme, nil
}

func (s *ColorSchemesService) UpdateColorScheme(ctx context.Context, command UpdateColorSchemeCommand) (domain.ColorScheme, error) {
	if err := CheckCentralPermission(command.Session.Roles, PermissionUsersManage); err != nil {
		return domain.ColorScheme{}, err
	}
	if command.SchemeID == "" {
		return domain.ColorScheme{}, ErrInvalidColorSchemeCommand
	}
	scheme, err := s.schemes.FindColorScheme(ctx, command.SchemeID)
	if err != nil {
		return domain.ColorScheme{}, err
	}
	if command.Name != nil {
		scheme.Name = *command.Name
	}
	if command.LogoURL != nil {
		scheme.LogoURL = *command.LogoURL
	}
	if command.AccentPreset != nil {
		scheme.AccentPreset = *command.AccentPreset
	}
	if command.AccentColor != nil {
		scheme.AccentColor = *command.AccentColor
	}
	if command.BackgroundColor != nil {
		scheme.BackgroundColor = *command.BackgroundColor
	}
	if command.Status != nil {
		scheme.Status = *command.Status
	}
	if command.StoreIDs != nil {
		scheme.StoreIDs = append([]string(nil), command.StoreIDs...)
	}
	scheme.UpdatedAt = s.now()

	validated, err := domain.NewColorScheme(scheme)
	if err != nil {
		return domain.ColorScheme{}, ErrInvalidColorSchemeCommand
	}
	validated.CreatedAt = scheme.CreatedAt
	if err := s.schemes.SaveColorScheme(ctx, validated); err != nil {
		return domain.ColorScheme{}, err
	}
	return validated, nil
}

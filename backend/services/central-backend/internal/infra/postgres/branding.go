package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"

	"mercadia.dev/pos/services/central-backend/internal/app"
	"mercadia.dev/pos/services/central-backend/internal/domain"
)

func (s *Store) SaveColorScheme(ctx context.Context, scheme domain.ColorScheme) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO color_schemes (
			id, name, logo_url, accent_preset, accent_color, background_color, status, store_ids, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			logo_url = EXCLUDED.logo_url,
			accent_preset = EXCLUDED.accent_preset,
			accent_color = EXCLUDED.accent_color,
			background_color = EXCLUDED.background_color,
			status = EXCLUDED.status,
			store_ids = EXCLUDED.store_ids,
			updated_at = EXCLUDED.updated_at
	`, scheme.ID, scheme.Name, scheme.LogoURL, string(scheme.AccentPreset), scheme.AccentColor,
		scheme.BackgroundColor, string(scheme.Status), scheme.StoreIDs, scheme.CreatedAt, scheme.UpdatedAt)
	if err != nil {
		return fmt.Errorf("save color scheme: %w", err)
	}
	return nil
}

func (s *Store) FindColorScheme(ctx context.Context, schemeID string) (domain.ColorScheme, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, name, logo_url, accent_preset, accent_color, background_color, status, store_ids, created_at, updated_at
		FROM color_schemes WHERE id = $1
	`, schemeID)
	return scanColorScheme(row)
}

func (s *Store) ListColorSchemes(ctx context.Context) ([]domain.ColorScheme, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, logo_url, accent_preset, accent_color, background_color, status, store_ids, created_at, updated_at
		FROM color_schemes ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("list color schemes: %w", err)
	}
	defer rows.Close()

	schemes := make([]domain.ColorScheme, 0)
	for rows.Next() {
		scheme, err := scanColorScheme(rows)
		if err != nil {
			return nil, err
		}
		schemes = append(schemes, scheme)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list color schemes: %w", err)
	}
	return schemes, nil
}

func (s *Store) SaveLayoutTemplate(ctx context.Context, template domain.LayoutTemplate) error {
	gridJSON, err := json.Marshal(template.Grid)
	if err != nil {
		return fmt.Errorf("marshal layout grid: %w", err)
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO layout_templates (
			id, name, kind, accent_preset, accent_color, color_scheme_id, grid,
			store_id, terminal_type, status, version, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			kind = EXCLUDED.kind,
			accent_preset = EXCLUDED.accent_preset,
			accent_color = EXCLUDED.accent_color,
			color_scheme_id = EXCLUDED.color_scheme_id,
			grid = EXCLUDED.grid,
			store_id = EXCLUDED.store_id,
			terminal_type = EXCLUDED.terminal_type,
			status = EXCLUDED.status,
			version = EXCLUDED.version,
			updated_at = EXCLUDED.updated_at
	`, template.ID, template.Name, string(template.Kind), string(template.AccentPreset), template.AccentColor,
		nullIfEmpty(template.ColorSchemeID), gridJSON, nullIfEmpty(template.StoreID), template.TerminalType,
		string(template.Status), template.Version, template.CreatedAt, template.UpdatedAt)
	if err != nil {
		return fmt.Errorf("save layout template: %w", err)
	}
	return nil
}

func (s *Store) FindLayoutTemplate(ctx context.Context, templateID string) (domain.LayoutTemplate, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, name, kind, accent_preset, accent_color, color_scheme_id, grid,
			store_id, terminal_type, status, version, created_at, updated_at
		FROM layout_templates WHERE id = $1
	`, templateID)
	return scanLayoutTemplate(row)
}

func (s *Store) ListLayoutTemplates(ctx context.Context, filter app.LayoutTemplateListFilter) ([]domain.LayoutTemplate, error) {
	query := `
		SELECT id, name, kind, accent_preset, accent_color, color_scheme_id, grid,
			store_id, terminal_type, status, version, created_at, updated_at
		FROM layout_templates WHERE 1=1`
	args := []any{}
	argIndex := 1
	if filter.StoreID != "" {
		query += fmt.Sprintf(" AND store_id = $%d", argIndex)
		args = append(args, filter.StoreID)
		argIndex++
	}
	if filter.Kind != "" {
		query += fmt.Sprintf(" AND kind = $%d", argIndex)
		args = append(args, string(filter.Kind))
	}
	query += " ORDER BY name"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list layout templates: %w", err)
	}
	defer rows.Close()

	templates := make([]domain.LayoutTemplate, 0)
	for rows.Next() {
		template, err := scanLayoutTemplate(rows)
		if err != nil {
			return nil, err
		}
		templates = append(templates, template)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list layout templates: %w", err)
	}
	return templates, nil
}

type colorSchemeRow interface {
	Scan(dest ...any) error
}

func scanColorScheme(row colorSchemeRow) (domain.ColorScheme, error) {
	var scheme domain.ColorScheme
	var accentPreset, status string
	if err := row.Scan(
		&scheme.ID, &scheme.Name, &scheme.LogoURL, &accentPreset, &scheme.AccentColor,
		&scheme.BackgroundColor, &status, &scheme.StoreIDs, &scheme.CreatedAt, &scheme.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return domain.ColorScheme{}, app.ErrColorSchemeNotFound
		}
		return domain.ColorScheme{}, fmt.Errorf("scan color scheme: %w", err)
	}
	scheme.AccentPreset = domain.AccentPreset(accentPreset)
	scheme.Status = domain.ColorSchemeStatus(status)
	return scheme, nil
}

func scanLayoutTemplate(row colorSchemeRow) (domain.LayoutTemplate, error) {
	var template domain.LayoutTemplate
	var kind, accentPreset, status string
	var colorSchemeID, storeID *string
	var gridJSON []byte
	if err := row.Scan(
		&template.ID, &template.Name, &kind, &accentPreset, &template.AccentColor, &colorSchemeID, &gridJSON,
		&storeID, &template.TerminalType, &status, &template.Version, &template.CreatedAt, &template.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return domain.LayoutTemplate{}, app.ErrLayoutTemplateNotFound
		}
		return domain.LayoutTemplate{}, fmt.Errorf("scan layout template: %w", err)
	}
	template.Kind = domain.LayoutTemplateKind(kind)
	template.AccentPreset = domain.AccentPreset(accentPreset)
	template.Status = domain.LayoutTemplateStatus(status)
	if colorSchemeID != nil {
		template.ColorSchemeID = *colorSchemeID
	}
	if storeID != nil {
		template.StoreID = *storeID
	}
	if err := json.Unmarshal(gridJSON, &template.Grid); err != nil {
		return domain.LayoutTemplate{}, fmt.Errorf("decode layout grid: %w", err)
	}
	if template.Grid.Tiles == nil {
		template.Grid.Tiles = []domain.LayoutGridTile{}
	}
	return template, nil
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

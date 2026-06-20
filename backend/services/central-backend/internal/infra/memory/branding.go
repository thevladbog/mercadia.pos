package memory

import (
	"context"
	"sort"

	"mercadia.dev/pos/services/central-backend/internal/app"
	"mercadia.dev/pos/services/central-backend/internal/domain"
)

func (s *Store) SaveColorScheme(ctx context.Context, scheme domain.ColorScheme) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.colorSchemes == nil {
		s.colorSchemes = map[string]domain.ColorScheme{}
	}
	s.colorSchemes[scheme.ID] = scheme
	return nil
}

func (s *Store) FindColorScheme(ctx context.Context, schemeID string) (domain.ColorScheme, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.colorSchemes == nil {
		return domain.ColorScheme{}, app.ErrColorSchemeNotFound
	}
	scheme, ok := s.colorSchemes[schemeID]
	if !ok {
		return domain.ColorScheme{}, app.ErrColorSchemeNotFound
	}
	return scheme, nil
}

func (s *Store) ListColorSchemes(ctx context.Context) ([]domain.ColorScheme, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	schemes := make([]domain.ColorScheme, 0, len(s.colorSchemes))
	for _, scheme := range s.colorSchemes {
		schemes = append(schemes, scheme)
	}
	sort.Slice(schemes, func(i, j int) bool {
		return schemes[i].Name < schemes[j].Name
	})
	return schemes, nil
}

func (s *Store) SaveLayoutTemplate(ctx context.Context, template domain.LayoutTemplate) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.layoutTemplates == nil {
		s.layoutTemplates = map[string]domain.LayoutTemplate{}
	}
	s.layoutTemplates[template.ID] = template
	return nil
}

func (s *Store) FindLayoutTemplate(ctx context.Context, templateID string) (domain.LayoutTemplate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.layoutTemplates == nil {
		return domain.LayoutTemplate{}, app.ErrLayoutTemplateNotFound
	}
	template, ok := s.layoutTemplates[templateID]
	if !ok {
		return domain.LayoutTemplate{}, app.ErrLayoutTemplateNotFound
	}
	return template, nil
}

func (s *Store) ListLayoutTemplates(ctx context.Context, filter app.LayoutTemplateListFilter) ([]domain.LayoutTemplate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	templates := make([]domain.LayoutTemplate, 0, len(s.layoutTemplates))
	for _, template := range s.layoutTemplates {
		if filter.StoreID != "" && template.StoreID != filter.StoreID {
			continue
		}
		if filter.Kind != "" && template.Kind != filter.Kind {
			continue
		}
		templates = append(templates, template)
	}
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Name < templates[j].Name
	})
	return templates, nil
}

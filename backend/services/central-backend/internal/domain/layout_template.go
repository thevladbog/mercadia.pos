package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var ErrInvalidLayoutTemplateInput = errors.New("invalid layout template input")

type LayoutTemplateKind string

const (
	LayoutTemplateKindSale   LayoutTemplateKind = "sale"
	LayoutTemplateKindReturn LayoutTemplateKind = "return"
	LayoutTemplateKindSco    LayoutTemplateKind = "sco"
)

type LayoutTemplateStatus string

const (
	LayoutTemplateStatusDraft     LayoutTemplateStatus = "draft"
	LayoutTemplateStatusPublished LayoutTemplateStatus = "published"
)

type LayoutGridCategory struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type LayoutGridTile struct {
	Label      string `json:"label"`
	Color      string `json:"color,omitempty"`
	ProductID  string `json:"productId,omitempty"`
	Empty      bool   `json:"empty,omitempty"`
	CategoryID string `json:"categoryId,omitempty"`
	IconURL    string `json:"iconUrl,omitempty"`
}

type LayoutGrid struct {
	Rows       int                  `json:"rows"`
	Cols       int                  `json:"cols"`
	Categories []LayoutGridCategory `json:"categories,omitempty"`
	Tiles      []LayoutGridTile     `json:"tiles"`
}

type LayoutTemplate struct {
	ID            string
	Name          string
	Kind          LayoutTemplateKind
	AccentPreset  AccentPreset
	AccentColor   string
	ColorSchemeID string
	Grid          LayoutGrid
	StoreID       string
	TerminalType  string
	Status        LayoutTemplateStatus
	Version       int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewLayoutTemplate(template LayoutTemplate) (LayoutTemplate, error) {
	template.Name = strings.TrimSpace(template.Name)
	if template.ID == "" || template.Name == "" {
		return LayoutTemplate{}, ErrInvalidLayoutTemplateInput
	}
	switch template.Kind {
	case LayoutTemplateKindSale, LayoutTemplateKindReturn, LayoutTemplateKindSco:
	default:
		return LayoutTemplate{}, ErrInvalidLayoutTemplateInput
	}
	if template.AccentPreset != "" {
		if _, ok := validAccentPresets[template.AccentPreset]; !ok {
			return LayoutTemplate{}, ErrInvalidLayoutTemplateInput
		}
	}
	if template.AccentColor != "" && !hexColorPattern.MatchString(template.AccentColor) {
		return LayoutTemplate{}, ErrInvalidLayoutTemplateInput
	}
	if template.AccentColor != "" {
		template.AccentColor = strings.ToUpper(template.AccentColor)
	}
	if template.Status == "" {
		template.Status = LayoutTemplateStatusDraft
	}
	if template.Status != LayoutTemplateStatusDraft && template.Status != LayoutTemplateStatusPublished {
		return LayoutTemplate{}, ErrInvalidLayoutTemplateInput
	}
	if template.Version <= 0 {
		template.Version = 1
	}
	if template.Grid.Rows <= 0 {
		template.Grid.Rows = 4
	}
	if template.Grid.Cols <= 0 {
		template.Grid.Cols = 4
	}
	if template.Grid.Tiles == nil {
		template.Grid.Tiles = []LayoutGridTile{}
	}
	if template.Grid.Categories == nil {
		template.Grid.Categories = []LayoutGridCategory{}
	}
	now := time.Now().UTC()
	if template.CreatedAt.IsZero() {
		template.CreatedAt = now
	}
	if template.UpdatedAt.IsZero() {
		template.UpdatedAt = now
	}
	return template, nil
}

func defaultPresetForKind(kind LayoutTemplateKind) AccentPreset {
	switch kind {
	case LayoutTemplateKindReturn:
		return AccentPresetReturn
	case LayoutTemplateKindSco:
		return AccentPresetSco
	default:
		return AccentPresetSale
	}
}

func (t LayoutTemplate) ResolvedAccent(scheme *ColorScheme) (AccentPreset, string) {
	if t.AccentColor != "" {
		preset := t.AccentPreset
		if preset == "" {
			preset = defaultPresetForKind(t.Kind)
		}
		return preset, t.AccentColor
	}
	if t.AccentPreset != "" {
		return t.AccentPreset, DefaultAccentHex(t.AccentPreset)
	}
	if scheme != nil {
		preset := scheme.AccentPreset
		if preset == "" {
			preset = AccentPresetNeutral
		}
		return preset, scheme.ResolvedAccentColor()
	}
	preset := defaultPresetForKind(t.Kind)
	return preset, DefaultAccentHex(preset)
}

func ParseLayoutGrid(raw json.RawMessage) (LayoutGrid, error) {
	if len(raw) == 0 {
		return LayoutGrid{Rows: 4, Cols: 4, Tiles: []LayoutGridTile{}}, nil
	}
	var grid LayoutGrid
	if err := json.Unmarshal(raw, &grid); err != nil {
		return LayoutGrid{}, err
	}
	if grid.Tiles == nil {
		grid.Tiles = []LayoutGridTile{}
	}
	if grid.Categories == nil {
		grid.Categories = []LayoutGridCategory{}
	}
	return grid, nil
}

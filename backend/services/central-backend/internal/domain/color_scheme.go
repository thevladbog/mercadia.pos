package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidColorSchemeInput = errors.New("invalid color scheme input")
	hexColorPattern            = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
)

type ColorSchemeStatus string

const (
	ColorSchemeStatusDraft     ColorSchemeStatus = "draft"
	ColorSchemeStatusPublished ColorSchemeStatus = "published"
)

type AccentPreset string

const (
	AccentPresetSale    AccentPreset = "sale"
	AccentPresetReturn  AccentPreset = "return"
	AccentPresetSco     AccentPreset = "sco"
	AccentPresetNeutral AccentPreset = "neutral"
)

var validAccentPresets = map[AccentPreset]struct{}{
	AccentPresetSale:    {},
	AccentPresetReturn:  {},
	AccentPresetSco:     {},
	AccentPresetNeutral: {},
}

type ColorScheme struct {
	ID              string
	Name            string
	LogoURL         string
	AccentPreset    AccentPreset
	AccentColor     string
	BackgroundColor string
	Status          ColorSchemeStatus
	StoreIDs        []string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewColorScheme(scheme ColorScheme) (ColorScheme, error) {
	scheme.Name = strings.TrimSpace(scheme.Name)
	if scheme.ID == "" || scheme.Name == "" {
		return ColorScheme{}, ErrInvalidColorSchemeInput
	}
	if scheme.AccentPreset != "" {
		if _, ok := validAccentPresets[scheme.AccentPreset]; !ok {
			return ColorScheme{}, ErrInvalidColorSchemeInput
		}
	} else if scheme.AccentColor != "" {
		if !hexColorPattern.MatchString(scheme.AccentColor) {
			return ColorScheme{}, ErrInvalidColorSchemeInput
		}
		scheme.AccentColor = strings.ToUpper(scheme.AccentColor)
	} else {
		return ColorScheme{}, ErrInvalidColorSchemeInput
	}
	if scheme.BackgroundColor != "" && !hexColorPattern.MatchString(scheme.BackgroundColor) {
		return ColorScheme{}, ErrInvalidColorSchemeInput
	}
	if scheme.BackgroundColor != "" {
		scheme.BackgroundColor = strings.ToUpper(scheme.BackgroundColor)
	}
	if scheme.Status == "" {
		scheme.Status = ColorSchemeStatusDraft
	}
	if scheme.Status != ColorSchemeStatusDraft && scheme.Status != ColorSchemeStatusPublished {
		return ColorScheme{}, ErrInvalidColorSchemeInput
	}
	if scheme.StoreIDs == nil {
		scheme.StoreIDs = []string{}
	}
	now := time.Now().UTC()
	if scheme.CreatedAt.IsZero() {
		scheme.CreatedAt = now
	}
	if scheme.UpdatedAt.IsZero() {
		scheme.UpdatedAt = now
	}
	return scheme, nil
}

func DefaultAccentHex(preset AccentPreset) string {
	switch preset {
	case AccentPresetReturn:
		return "#2563EB"
	case AccentPresetSco:
		return "#F25F1C"
	case AccentPresetNeutral, AccentPresetSale:
		return "#FF6600"
	default:
		return "#FF6600"
	}
}

func (s ColorScheme) ResolvedAccentColor() string {
	if s.AccentColor != "" {
		return s.AccentColor
	}
	return DefaultAccentHex(s.AccentPreset)
}

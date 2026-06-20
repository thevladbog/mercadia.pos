export type Surface = 'admin' | 'terminal' | 'sco' | 'senior-cashier';

export type ColorMode = 'light' | 'dark';

export type AccentPreset = 'sale' | 'return' | 'sco' | 'neutral';

export type ThemeConfig = {
  surface: Surface;
  colorMode: ColorMode;
  /** Runtime hex from layout template or Color Scheme */
  accent?: string;
  /** Shorthand mapping to a built-in accent hex */
  accentPreset?: AccentPreset;
};

export type DerivedAccentTokens = {
  accent: string;
  accentHover: string;
  accentMuted: string;
  accentForeground: string;
};

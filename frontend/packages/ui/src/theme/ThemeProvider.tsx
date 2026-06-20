import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react';

import { applyTheme } from './applyTheme.js';
import type { ThemeConfig } from './types.js';

type ThemeContextValue = {
  theme: ThemeConfig;
  setTheme: (config: ThemeConfig) => void;
};

const ThemeContext = createContext<ThemeContextValue | null>(null);

const THEME_STORAGE_KEY = 'mercadia-ui-theme';

type ThemeProviderProps = {
  children: ReactNode;
  defaultTheme: ThemeConfig;
  persist?: boolean;
};

function readPersistedTheme(defaultTheme: ThemeConfig): ThemeConfig {
  if (typeof window === 'undefined') {
    return defaultTheme;
  }
  try {
    const raw = window.localStorage.getItem(THEME_STORAGE_KEY);
    if (!raw) {
      return defaultTheme;
    }
    return { ...defaultTheme, ...JSON.parse(raw) } as ThemeConfig;
  } catch {
    return defaultTheme;
  }
}

export function ThemeProvider({ children, defaultTheme, persist = false }: ThemeProviderProps) {
  const [theme, setThemeState] = useState<ThemeConfig>(() =>
    persist ? readPersistedTheme(defaultTheme) : defaultTheme,
  );

  useEffect(() => {
    applyTheme(theme);
  }, [theme]);

  useEffect(() => {
    if (!persist) {
      return;
    }
    window.localStorage.setItem(THEME_STORAGE_KEY, JSON.stringify(theme));
  }, [persist, theme]);

  const value = useMemo(
    () => ({
      theme,
      setTheme: setThemeState,
    }),
    [theme],
  );

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}

export function useTheme(): ThemeContextValue {
  const context = useContext(ThemeContext);
  if (context == null) {
    throw new Error('useTheme must be used within ThemeProvider');
  }
  return context;
}

export { applyTheme, clearTheme } from './applyTheme.js';
export { deriveAccentTokens } from './deriveAccent.js';
export { ACCENT_PRESETS, resolveAccentHex } from './presets.js';
export type {
  AccentPreset,
  ColorMode,
  DerivedAccentTokens,
  Surface,
  ThemeConfig,
} from './types.js';

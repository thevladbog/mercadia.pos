import { deriveAccentTokens } from './deriveAccent.js';
import { resolveAccentHex } from './presets.js';
import type { ThemeConfig } from './types.js';

const ACCENT_CSS_VARS = [
  '--ui-accent',
  '--ui-accent-hover',
  '--ui-accent-muted',
  '--ui-accent-foreground',
] as const;

export function applyTheme(
  config: ThemeConfig,
  root: HTMLElement = document.documentElement,
): void {
  root.dataset.surface = config.surface;
  root.dataset.colorMode = config.colorMode;

  const accentHex = resolveAccentHex({
    accent: config.accent,
    accentPreset: config.accentPreset,
  });
  const derived = deriveAccentTokens(accentHex);

  root.style.setProperty('--ui-accent', derived.accent);
  root.style.setProperty('--ui-accent-hover', derived.accentHover);
  root.style.setProperty('--ui-accent-muted', derived.accentMuted);
  root.style.setProperty('--ui-accent-foreground', derived.accentForeground);
}

export function clearTheme(root: HTMLElement = document.documentElement): void {
  delete root.dataset.surface;
  delete root.dataset.colorMode;
  for (const cssVar of ACCENT_CSS_VARS) {
    root.style.removeProperty(cssVar);
  }
}

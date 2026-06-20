import type { AccentPreset } from './types.js';

export const ACCENT_PRESETS: Record<AccentPreset, string> = {
  sale: '#FF6600',
  return: '#2563EB',
  sco: '#F25F1C',
  neutral: '#FF6600',
};

export function resolveAccentHex(options: {
  accent?: string;
  accentPreset?: AccentPreset;
}): string {
  if (options.accent != null && options.accent.length > 0) {
    return normalizeHex(options.accent);
  }
  if (options.accentPreset != null) {
    return ACCENT_PRESETS[options.accentPreset];
  }
  return ACCENT_PRESETS.neutral;
}

function normalizeHex(hex: string): string {
  const trimmed = hex.trim();
  if (/^#[0-9a-fA-F]{6}$/.test(trimmed)) {
    return trimmed.toUpperCase();
  }
  if (/^#[0-9a-fA-F]{3}$/.test(trimmed)) {
    const [, r, g, b] = trimmed;
    return `#${r}${r}${g}${g}${b}${b}`.toUpperCase();
  }
  throw new Error(`Invalid accent hex: ${hex}`);
}

import type { ColorMode, DerivedAccentTokens } from './types.js';

function parseHex(hex: string): { r: number; g: number; b: number } {
  const normalized = hex.replace('#', '');
  return {
    r: Number.parseInt(normalized.slice(0, 2), 16),
    g: Number.parseInt(normalized.slice(2, 4), 16),
    b: Number.parseInt(normalized.slice(4, 6), 16),
  };
}

function toHex(r: number, g: number, b: number): string {
  const clamp = (value: number) => Math.max(0, Math.min(255, Math.round(value)));
  return `#${clamp(r).toString(16).padStart(2, '0')}${clamp(g).toString(16).padStart(2, '0')}${clamp(b).toString(16).padStart(2, '0')}`.toUpperCase();
}

function mix(hex: string, target: { r: number; g: number; b: number }, amount: number): string {
  const source = parseHex(hex);
  return toHex(
    source.r + (target.r - source.r) * amount,
    source.g + (target.g - source.g) * amount,
    source.b + (target.b - source.b) * amount,
  );
}

function relativeLuminance(hex: string): number {
  const { r, g, b } = parseHex(hex);
  const channel = (value: number) => {
    const s = value / 255;
    return s <= 0.03928 ? s / 12.92 : ((s + 0.055) / 1.055) ** 2.4;
  };
  return 0.2126 * channel(r) + 0.7152 * channel(g) + 0.0722 * channel(b);
}

export function deriveAccentTokens(
  accent: string,
  colorMode: ColorMode = 'light',
): DerivedAccentTokens {
  const normalized = accent.startsWith('#') ? accent : `#${accent}`;
  const { r, g, b } = parseHex(normalized);
  const accentMuted =
    colorMode === 'dark'
      ? `rgb(${r} ${g} ${b} / 0.15)`
      : mix(normalized, { r: 255, g: 255, b: 255 }, 0.88);
  const accentHover =
    colorMode === 'dark'
      ? mix(normalized, { r: 255, g: 255, b: 255 }, 0.12)
      : mix(normalized, { r: 0, g: 0, b: 0 }, 0.12);

  return {
    accent: normalized.toUpperCase(),
    accentHover,
    accentMuted,
    accentForeground: relativeLuminance(normalized) > 0.55 ? '#1A1A1A' : '#FFFFFF',
  };
}

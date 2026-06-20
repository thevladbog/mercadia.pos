import type { ReactNode } from 'react';
import type { CSSProperties } from 'react';
import { useMemo } from 'react';

import { deriveAccentTokens } from '../../theme/deriveAccent.js';
import { resolveAccentHex } from '../../theme/presets.js';
import { ACCENT_PRESETS } from '../../theme/presets.js';
import type { AccentPreset, ColorMode, Surface, ThemeConfig } from '../../theme/types.js';

type ThemePreviewProps = {
  children: ReactNode;
  className?: string;
  theme: Pick<ThemeConfig, 'surface' | 'colorMode' | 'accent' | 'accentPreset'>;
};

/** Applies accent tokens locally without mutating document theme. */
export function ThemePreview({ children, className, theme }: ThemePreviewProps) {
  const style = useMemo(() => {
    const accentHex = resolveAccentHex({
      accent: theme.accent,
      accentPreset: theme.accentPreset,
    });
    const derived = deriveAccentTokens(accentHex, theme.colorMode ?? 'light');
    return {
      '--ui-accent': derived.accent,
      '--ui-accent-hover': derived.accentHover,
      '--ui-accent-muted': derived.accentMuted,
      '--ui-accent-foreground': derived.accentForeground,
    } as CSSProperties;
  }, [theme.accent, theme.accentPreset, theme.colorMode]);

  return (
    <div
      className={className}
      data-color-mode={theme.colorMode ?? 'light'}
      data-surface={theme.surface ?? 'terminal'}
      style={style}
    >
      {children}
    </div>
  );
}

export { ACCENT_PRESETS };
export type { AccentPreset, ColorMode, Surface };

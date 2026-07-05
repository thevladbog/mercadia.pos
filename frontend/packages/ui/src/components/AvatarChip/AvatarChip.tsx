import { cva, type VariantProps } from 'class-variance-authority';
import type { HTMLAttributes } from 'react';

import { cn } from '../../lib/cn.js';
import { deriveAccentTokens } from '../../theme/deriveAccent.js';

const avatarChipVariants = cva('mercadia-avatar-chip', {
  variants: {
    size: {
      sm: 'mercadia-avatar-chip--sm',
      md: 'mercadia-avatar-chip--md',
      lg: 'mercadia-avatar-chip--lg',
    },
  },
  defaultVariants: {
    size: 'md',
  },
});

/**
 * Small fixed palette used to derive a deterministic per-person avatar color
 * when the caller doesn't pass an explicit `color`. Mixes theme-aware CSS
 * custom properties (so the chip tracks the active surface/accent) with a
 * couple of fixed hex fallbacks matching colors observed in the
 * senior-cashier design export (teal, slate/blue-gray) — see Plan 018.
 */
const AVATAR_PALETTE = ['var(--ui-accent)', 'var(--ui-info)', '#0F766E', '#475569', '#7C3AED'];

/**
 * Simple deterministic string hash (djb2-style). No external dependency —
 * mirrors the "small, dependency-free pure function" style already used in
 * theme/deriveAccent.ts.
 */
function hashString(value: string): number {
  let hash = 5381;
  for (let i = 0; i < value.length; i += 1) {
    hash = (hash * 33) ^ value.charCodeAt(i);
  }
  return Math.abs(hash);
}

/** Deterministically resolve a palette color for a given initials string. */
export function resolveAvatarColor(initials: string): string {
  const index = hashString(initials) % AVATAR_PALETTE.length;
  return AVATAR_PALETTE[index];
}

/**
 * Resolve a readable foreground color for `background`. CSS custom
 * properties (e.g. `var(--ui-accent)`) can't be measured for luminance
 * without a live DOM, so those default to white — which matches every
 * avatar color observed in the design. Concrete hex colors reuse
 * `deriveAccentTokens`'s luminance/contrast math instead of reinventing it.
 */
function resolveForeground(background: string): string {
  if (!background.startsWith('#')) {
    return '#FFFFFF';
  }
  return deriveAccentTokens(background).accentForeground;
}

export type AvatarChipProps = Omit<HTMLAttributes<HTMLDivElement>, 'color'> &
  VariantProps<typeof avatarChipVariants> & {
    /** Already-computed 1-2 letter initials (this component does no name parsing). */
    initials: string;
    /** Optional explicit color override; otherwise derived deterministically from `initials`. */
    color?: string;
  };

export function AvatarChip({ initials, color, size, className, style, ...props }: AvatarChipProps) {
  const resolvedColor = color ?? resolveAvatarColor(initials);
  const foreground = resolveForeground(resolvedColor);

  return (
    <div
      className={cn(avatarChipVariants({ size }), className)}
      style={{ backgroundColor: resolvedColor, color: foreground, ...style }}
      {...props}
    >
      {initials.toUpperCase()}
    </div>
  );
}

export { avatarChipVariants };

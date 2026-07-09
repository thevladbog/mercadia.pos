import type { ReactNode } from 'react';
import { Badge } from '@mercadia/ui';

/** Mirrors `Badge`'s own variant union (`packages/ui/src/components/Badge/Badge.tsx`)
 * without importing `class-variance-authority` into this app just for the type. */
type BadgeVariant = 'default' | 'accent' | 'success' | 'warning' | 'danger' | 'info' | 'outline';

/** Icon-tile accent color. Purely visual grouping (green = money in, blue =
 * money out, red = shift close-out) mirroring design screen 02's primary
 * grid; `neutral` is used for cards that don't carry one of those meanings
 * (e.g. the promoted safe-recount card — see `DashboardPage.tsx`'s comment
 * on why it has no dedicated accent color of its own). */
export type ActionCardAccent = 'green' | 'blue' | 'red' | 'neutral';

export interface ActionCardProps {
  icon: ReactNode;
  title: string;
  /** Omitted entirely (not rendered empty) when there's no real subtitle text. */
  subtitle?: string;
  badgeText?: string;
  badgeVariant?: BadgeVariant;
  onClick: () => void;
  /** Controls size/prominence: `primary` for the 2x2 top grid, `secondary`
   * for the smaller rows below. Defaults to `secondary`. */
  variant?: 'primary' | 'secondary';
  accent?: ActionCardAccent;
}

/**
 * Single reusable action-card component backing all 3 dashboard rows (plan
 * 021) — the primary 2x2 grid, the "ОПЕРАЦИИ С СЕЙФОМ" row, and the
 * "СИСТЕМА" row all render this same component with different `variant`/
 * `accent`/content, rather than three near-duplicate components.
 */
export function ActionCard({
  icon,
  title,
  subtitle,
  badgeText,
  badgeVariant = 'danger',
  onClick,
  variant = 'secondary',
  accent = 'neutral',
}: ActionCardProps) {
  return (
    <button type="button" className={`sr-action-card sr-action-card--${variant}`} onClick={onClick}>
      {badgeText && (
        <Badge variant={badgeVariant} className="sr-action-card-badge">
          {badgeText}
        </Badge>
      )}
      <span className={`sr-action-card-icon sr-action-card-icon--${accent}`}>{icon}</span>
      <span className="sr-action-card-title">{title}</span>
      {subtitle && <span className="sr-action-card-subtitle">{subtitle}</span>}
    </button>
  );
}

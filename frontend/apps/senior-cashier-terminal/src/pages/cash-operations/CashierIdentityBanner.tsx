import type { ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { AvatarChip, Badge } from '@mercadia/ui';

import { deriveInitials, formatRoleLabel } from '@/lib/topbar-utils.js';

export interface CashierIdentityBannerProps {
  /** Reuse `pages/dashboard/icons.tsx`'s `DownArrowIcon`/`UpArrowIcon`/`BoxIcon` — never duplicate them. */
  icon: ReactNode;
  /** Same green/blue/red money-flow accent vocabulary as `ActionCardAccent` (plan 021). */
  accent: 'green' | 'blue' | 'red';
  /** Pre-translated by the caller, e.g. "Safe → Drawer" / "Drawer → Safe". */
  directionLabel: string;
  cashierId: string;
  /** `null` when no credential actor matches, or the actor has no roles —
   * the role badge is omitted entirely in that case, never guessed. */
  role: string | null;
}

/**
 * Read-only identity banner for the shift's cashier (design screens
 * 03a/03b/03c's green/blue identity card — see plan 022 Scope item 4).
 * Replaces the 2 manual `actorId`/`approvedById` free-text inputs on the 3
 * shift-based pages (Issue Change Fund, Receive Cash, Final Collection) —
 * both identities are already known once a shift is selected: the cashier
 * from `selectedShift.cashierId`, and the approver (this terminal's signed
 * -in senior cashier) from `useAuth()`'s `session.actorId`, rendered by the
 * page via the existing `TopBar` identity, not duplicated here.
 *
 * Shows `cashierId` as-is with NO display name — there is no name field on
 * `domain.Actor` (same precedent as `TopBar`/`CashierShiftRow` in every
 * prior phase), and no register/terminal number — matching plan 021's
 * established omission of a "register number" display label.
 */
export function CashierIdentityBanner({
  icon,
  accent,
  directionLabel,
  cashierId,
  role,
}: CashierIdentityBannerProps) {
  const { t } = useTranslation();
  const initials = deriveInitials(cashierId);

  return (
    <div className={`sr-cashier-identity-banner sr-cashier-identity-banner--${accent}`}>
      <span className={`sr-cashier-identity-icon sr-cashier-identity-icon--${accent}`}>{icon}</span>

      <div className="sr-cashier-identity-text">
        <span className="sr-cashier-identity-direction">{directionLabel}</span>
        <div className="sr-cashier-identity-name">
          <AvatarChip initials={initials} size="sm" />
          <span>{cashierId}</span>
          {role && <Badge variant="outline">{formatRoleLabel(role, t)}</Badge>}
        </div>
      </div>
    </div>
  );
}

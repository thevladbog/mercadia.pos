import { useTranslation } from 'react-i18next';
import { AvatarChip, Badge } from '@mercadia/ui';

import { formatMinor } from '@/lib/cash-utils.js';
import { deriveInitials, formatLoginTime, formatRoleLabel } from '@/lib/topbar-utils.js';
import type { CashierOnShift } from '@/lib/dashboard-data.js';

export interface CashierShiftRowProps {
  shift: CashierOnShift;
}

/**
 * One row of the "Кассиры на смене" panel (plan 021). Renders exactly the
 * real joined fields from `dashboard-data.ts`'s `joinCashiersOnShift` — no
 * "register number" display label (`terminalId` is shown as-is) and no
 * return-pending highlight (neither exists in the backend; see plan 021
 * "Why this matters" items 2 and 5).
 */
export function CashierShiftRow({ shift }: CashierShiftRowProps) {
  const { t } = useTranslation();
  const initials = deriveInitials(shift.cashierId);
  const openedAtMs = new Date(shift.openedAt).getTime();
  const sinceLabel = Number.isFinite(openedAtMs)
    ? t('topbar.since', { time: formatLoginTime(openedAtMs) })
    : '';

  return (
    <div className="sr-cashier-row">
      <AvatarChip initials={initials} size="sm" />

      <div className="sr-cashier-row-identity">
        <div className="sr-cashier-row-name">
          <span>{shift.cashierId}</span>
          {shift.role && <Badge variant="outline">{formatRoleLabel(shift.role, t)}</Badge>}
        </div>
        <div className="sr-cashier-row-meta muted">
          {shift.terminalId} · {sinceLabel}
        </div>
      </div>

      <div className="sr-cashier-row-figures">
        <div className="sr-cashier-row-revenue">
          {formatMinor(shift.revenueMinor)} ₽
          <span className="sr-cashier-row-figure-label muted"> {t('dashboard.revenue')}</span>
        </div>
        <div className="sr-cashier-row-drawer muted">
          {t('dashboard.cashiers.drawerAmount')} {formatMinor(shift.drawerBalanceMinor)} ₽
        </div>
      </div>
    </div>
  );
}

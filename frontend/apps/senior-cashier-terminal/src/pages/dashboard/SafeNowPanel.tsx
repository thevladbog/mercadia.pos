import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import { formatMinor } from '@/lib/cash-utils.js';
import { deriveSafeFreshnessLabel } from '@/lib/dashboard-data.js';

export interface SafeNowPanelProps {
  /** `undefined` while the balances query hasn't resolved yet (no safe
   * container found is a separate, real "no data" case — never fabricated). */
  balanceMinor: number | undefined;
  lastMovementAtIso: string | undefined;
}

/**
 * "Сейф · сейчас" panel (plan 021 "Why this matters" item 4). Shows exactly
 * ONE real stat — the safe container's balance — plus its real
 * `lastMovementAt` as a live "обновлено N назад" freshness label. There is
 * no safe limit or collection threshold anywhere in the API, so this
 * deliberately has no "К инкассации"/"ЛИМИТ" tile and no fill-bar — adding
 * either would mean fabricating a number the backend doesn't provide.
 *
 * Owns its own 1s ticking `now` state (rather than receiving `now` as a
 * prop from `DashboardPage`) so only this small panel re-renders every
 * second, not the whole dashboard tree — same self-contained-ticker pattern
 * `TopBar.tsx` already uses for its own countdown display.
 */
export function SafeNowPanel({ balanceMinor, lastMovementAtIso }: SafeNowPanelProps) {
  const { t } = useTranslation();
  const [now, setNow] = useState(() => Date.now());

  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(id);
  }, []);

  const freshnessLabel = deriveSafeFreshnessLabel(lastMovementAtIso, now, (key, params) =>
    t(key, params),
  );

  return (
    <div className="sr-panel sr-safe-now-panel">
      <div className="sr-panel-header">
        <h2 className="sr-panel-title">{t('dashboard.safeNow.title')}</h2>
        <span className="muted" style={{ fontSize: '0.85rem' }}>
          {t('dashboard.safeNow.updated', { label: freshnessLabel })}
        </span>
      </div>

      <div className="sr-safe-now-stat">
        <span className="sr-safe-now-label">{t('dashboard.safeNow.balanceLabel')}</span>
        <span className="sr-safe-now-value">
          {balanceMinor !== undefined ? `${formatMinor(balanceMinor)} ₽` : '—'}
        </span>
      </div>
    </div>
  );
}

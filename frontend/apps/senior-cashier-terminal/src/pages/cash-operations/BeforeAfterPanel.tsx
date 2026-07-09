import { useTranslation } from 'react-i18next';

import { formatMinor } from '@/lib/cash-utils.js';

import { computeAfterBalance, type BalanceDirection } from './cash-operations-data.js';

export interface ContainerBeforeAfter {
  kind: 'safe' | 'drawer';
  /** `undefined` while the balances query hasn't resolved yet — rendered as
   * "—", never fabricated as 0 (same "no data" precedent as `SafeNowPanel`). */
  beforeMinor: number | undefined;
  /** Whether this container gains ("increase") or loses ("decrease") the
   * operation's total — e.g. the safe increases and the drawer decreases
   * for Issue Change Fund, the reverse for Receive Cash. */
  direction: BalanceDirection;
}

export interface BeforeAfterPanelProps {
  /** Pre-translated by the caller — the total's meaning differs per page
   * ("Amount to issue" / "to receive" / collection / expense), same
   * already-translated convention `OperationChecklist` uses for its steps. */
  totalLabel: string;
  totalMinor: number;
  /** 1 entry for safe-only operations (Bank Collection, Business Expense),
   * 2 for the shift-based operations (safe + drawer), in the order the
   * caller wants them displayed (source container first, by convention). */
  containers: ContainerBeforeAfter[];
}

const LABEL_KEYS: Record<ContainerBeforeAfter['kind'], { before: string; after: string }> = {
  safe: { before: 'cash.beforeAfter.safeBefore', after: 'cash.beforeAfter.safeAfter' },
  drawer: { before: 'cash.beforeAfter.drawerBefore', after: 'cash.beforeAfter.drawerAfter' },
};

/**
 * Right-side "СУММА" + before/after tiles (design screens 03a/03b's right
 * column — see plan 022 Scope item 2).
 *
 * "After" values are a CLIENT-SIDE PREVIEW (`before ± total`, computed live
 * as the operator types), NOT a server-confirmed post-transaction figure —
 * a concurrent operation could change the real balance before this one is
 * submitted. Shown with a "≈" prefix and a title tooltip to make that
 * honest, mirroring the same "optimistic UI, not a guarantee" framing plan
 * 020 already established for its step-checkmarks. See
 * `computeAfterBalance`'s doc comment in `cash-operations-data.ts`.
 */
export function BeforeAfterPanel({ totalLabel, totalMinor, containers }: BeforeAfterPanelProps) {
  const { t } = useTranslation();

  return (
    <div className="sr-panel sr-before-after-panel">
      <span className="sr-before-after-total-label">{totalLabel}</span>
      <div className="sr-before-after-total-value">
        {formatMinor(totalMinor)} <span className="sr-before-after-total-currency">₽</span>
      </div>

      <div className="sr-before-after-grid">
        {containers.map((container) => {
          const labels = LABEL_KEYS[container.kind];
          const hasBefore = container.beforeMinor !== undefined;
          const afterMinor = hasBefore
            ? computeAfterBalance(container.beforeMinor as number, totalMinor, container.direction)
            : undefined;

          return (
            <div className="sr-before-after-pair" key={container.kind}>
              <div className="sr-before-after-tile">
                <span className="sr-before-after-tile-label">{t(labels.before)}</span>
                <span className="sr-before-after-tile-value">
                  {hasBefore ? `${formatMinor(container.beforeMinor as number)} ₽` : '—'}
                </span>
              </div>
              <div className="sr-before-after-tile">
                <span className="sr-before-after-tile-label">{t(labels.after)}</span>
                <span
                  className="sr-before-after-tile-value sr-before-after-tile-value--preview"
                  title={t('cash.beforeAfter.previewHint')}
                >
                  {afterMinor !== undefined ? `≈ ${formatMinor(afterMinor)} ₽` : '—'}
                </span>
              </div>
            </div>
          );
        })}
      </div>

      <p className="muted sr-before-after-hint">{t('cash.beforeAfter.previewHint')}</p>
    </div>
  );
}

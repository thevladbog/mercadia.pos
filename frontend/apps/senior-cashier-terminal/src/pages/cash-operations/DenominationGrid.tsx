import { useCallback, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { Input } from '@mercadia/ui';

import { computeDenominationTotal, formatMinor } from '@/lib/cash-utils.js';

/**
 * Local 7-entry bill list for the redesigned tile grid (design screens
 * 03a/03b/03c), deliberately SEPARATE from `cash-utils.ts`'s 11-entry
 * `RUBLE_DENOMINATIONS` — see plan 022 item 6. That list still backs
 * `DenominationInput`/`SafeRecountPage.tsx` unchanged; this one only serves
 * the new grid. Values in minor units (kopecks), ascending, matching the
 * design's left-to-right/top-to-bottom tile order exactly.
 */
const GRID_BILLS = [
  { value: 50_00, label: '50' },
  { value: 100_00, label: '100' },
  { value: 200_00, label: '200' },
  { value: 500_00, label: '500' },
  { value: 1000_00, label: '1 000' },
  { value: 2000_00, label: '2 000' },
  { value: 5000_00, label: '5 000' },
] as const;

/**
 * Fixed max visible fill-marks per tile before switching to "+N" overflow
 * text, reusing the visual pattern plan 020's `PinDots.tsx` established for
 * masked PIN entry. NOT guessed: derived by counting the actual marks in
 * the design PNGs — `03a`'s "100 ₽ · 20 купюр" tile shows 14 marks + "+6",
 * `03b`'s "1000 ₽ · 22 купюр" shows 14 marks + "+8", and `03b`'s
 * "500 ₽ · 14 купюр" shows all 14 marks with no overflow — all three data
 * points agree on a cutoff of exactly 14.
 */
const MAX_VISIBLE_MARKS = 14;

export type DenominationGridVariant = 'issue' | 'receive' | 'recount';

const TITLE_KEY_BY_VARIANT: Record<DenominationGridVariant, string> = {
  issue: 'cash.grid.titleIssue',
  receive: 'cash.grid.titleReceive',
  recount: 'cash.grid.titleRecount',
};

export interface DenominationGridProps {
  /** Switches the header copy between "to issue" / "to receive" / "to recount" — the
   * only difference between the 3 pages that use this grid, so one component
   * covers all of them instead of 3 near-duplicates. */
  variant: DenominationGridVariant;
  billValues: Record<number, string>;
  onBillValuesChange: (values: Record<number, string>) => void;
  /** Rubles-amount field for the design's consolidated "МОНЕТЫ" (coins) tile. */
  coinsMinor: number;
  onCoinsMinorChange: (value: number) => void;
  /** Rubles-amount field for the design's "ДРУГОЕ" (other) tile. */
  otherMinor: number;
  onOtherMinorChange: (value: number) => void;
}

/**
 * Redesigned tile grid for the denomination breakdown (design screens
 * 03a/03b/03c's recount grid — see plan 022 Scope item 1).
 *
 * IMPORTANT — this is still fully operator-typed input, exactly like the
 * existing `DenominationInput`. The design's "Состав определён кассиром на
 * ККМ" copy and its lock-icon / "Из ККМ · только просмотр" badges imply the
 * breakdown is locked, KKM-determined data — plan 022 item 1 confirmed that
 * has NO backend support anywhere (`CreateCashMovementBody` only ever
 * carries a flat `amountMinor`; there is no denomination field in the API).
 * This component deliberately does NOT render any locked/read-only
 * affordance — every tile is a live, editable count/amount input, same data
 * source as before, just a redesigned visual layer.
 *
 * The grand total is computed here via the EXISTING
 * `computeDenominationTotal(billValues, coinsMinor + otherMinor)` — reused
 * as-is, its signature is unchanged (plan 022 item 6) — and shown via the
 * same `cash.total` label `DenominationInput` already uses. Callers
 * (the 5 pages) independently call the same function on the same state to
 * get their own `countedMinor`, mirroring how `DenominationInput` and its
 * callers already both compute their own total from the same inputs today.
 */
export function DenominationGrid({
  variant,
  billValues,
  onBillValuesChange,
  coinsMinor,
  onCoinsMinorChange,
  otherMinor,
  onOtherMinorChange,
}: DenominationGridProps) {
  const { t } = useTranslation();

  const handleBillChange = useCallback(
    (denomValue: number, countStr: string) => {
      onBillValuesChange({ ...billValues, [denomValue]: countStr });
    },
    [billValues, onBillValuesChange],
  );

  const billsCount = useMemo(
    () => GRID_BILLS.reduce((sum, bill) => sum + (Number(billValues[bill.value]) || 0), 0),
    [billValues],
  );

  const total = useMemo(
    () => computeDenominationTotal(billValues, coinsMinor + otherMinor),
    [billValues, coinsMinor, otherMinor],
  );

  return (
    <div className="sr-panel sr-denom-grid-panel">
      <div className="sr-panel-header">
        <h2 className="sr-panel-title">{t(TITLE_KEY_BY_VARIANT[variant])}</h2>
        <span className="muted sr-denom-grid-count">
          {t('cash.grid.billsCount', { count: billsCount })}
        </span>
      </div>

      <div className="sr-denom-grid">
        {GRID_BILLS.map((bill) => {
          const countStr = billValues[bill.value] ?? '';
          const count = Number(countStr) || 0;
          const subtotal = bill.value * count;
          const visibleMarks = Math.min(Math.max(count, 0), MAX_VISIBLE_MARKS);
          const overflow = count - visibleMarks;

          return (
            <div className="sr-denom-tile" key={bill.value}>
              <span className="sr-denom-tile-label">{bill.label} ₽</span>
              <Input
                type="number"
                min={0}
                step={1}
                value={countStr}
                onChange={(e) => handleBillChange(bill.value, e.target.value)}
                placeholder="0"
                className="sr-denom-tile-input"
                aria-label={`${bill.label} ₽ — ${t('cash.grid.bills')}`}
              />
              <span className="muted sr-denom-tile-unit">{t('cash.grid.bills')}</span>

              {count > 0 ? (
                <div className="sr-denom-tile-marks" aria-hidden="true">
                  {Array.from({ length: visibleMarks }, (_, index) => (
                    <span key={index} className="sr-denom-tile-mark" />
                  ))}
                  {overflow > 0 && <span className="sr-denom-tile-overflow">+{overflow}</span>}
                </div>
              ) : (
                <div className="sr-denom-tile-marks sr-denom-tile-marks--empty">—</div>
              )}

              <span className="sr-denom-tile-subtotal">= {formatMinor(subtotal)} ₽</span>
            </div>
          );
        })}

        <div className="sr-denom-tile">
          <span className="sr-denom-tile-label">{t('cash.grid.coins')}</span>
          <Input
            type="number"
            min={0}
            step={0.01}
            value={coinsMinor > 0 ? (coinsMinor / 100).toString() : ''}
            onChange={(e) =>
              onCoinsMinorChange(Math.round(parseFloat(e.target.value || '0') * 100))
            }
            placeholder="0"
            className="sr-denom-tile-input"
            aria-label={t('cash.grid.coins')}
          />
        </div>

        <div className="sr-denom-tile">
          <span className="sr-denom-tile-label">{t('cash.grid.other')}</span>
          <Input
            type="number"
            min={0}
            step={0.01}
            value={otherMinor > 0 ? (otherMinor / 100).toString() : ''}
            onChange={(e) =>
              onOtherMinorChange(Math.round(parseFloat(e.target.value || '0') * 100))
            }
            placeholder="0"
            className="sr-denom-tile-input"
            aria-label={t('cash.grid.other')}
          />
        </div>
      </div>

      <div className="sr-denomination-total">
        {t('cash.total')}: {formatMinor(total)} ₽
      </div>
    </div>
  );
}

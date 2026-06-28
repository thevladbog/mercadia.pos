import { useCallback, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { Input } from '@mercadia/ui';

import { getDenominations, computeDenominationTotal, formatMinor } from '@/lib/cash-utils.js';

interface DenominationInputProps {
  values: Record<number, string>;
  onChange: (values: Record<number, string>) => void;
  otherAmountMinor?: number;
  onOtherAmountChange?: (value: number) => void;
}

export function DenominationInput({ values, onChange, otherAmountMinor = 0, onOtherAmountChange }: DenominationInputProps) {
  const { t } = useTranslation();
  const denominations = useMemo(() => getDenominations(), []);

  const handleChange = useCallback(
    (denomValue: number, countStr: string) => {
      onChange({ ...values, [denomValue]: countStr });
    },
    [values, onChange],
  );

  const total = useMemo(
    () => computeDenominationTotal(values, otherAmountMinor),
    [values, otherAmountMinor],
  );

  return (
    <div>
      <span className="sr-field-label">{t('cash.denominationBreakdown')}</span>
      <div className="sr-denomination-grid" style={{ marginTop: '0.5rem' }}>
        {denominations.map((denom) => (
          <div key={denom.value} className="sr-denomination-row">
            <span className="sr-denomination-label">{denom.label} ₽</span>
            <Input
              type="number"
              min={0}
              value={values[denom.value] ?? ''}
              onChange={(e) => handleChange(denom.value, e.target.value)}
              placeholder="0"
              style={{ width: 64 }}
            />
          </div>
        ))}
      </div>

      {onOtherAmountChange && (
        <div className="sr-denomination-row" style={{ marginTop: '0.5rem' }}>
          <span className="sr-denomination-label">{t('cash.otherCoins')}</span>
          <Input
            type="number"
            min={0}
            value={otherAmountMinor > 0 ? (otherAmountMinor / 100).toString() : ''}
            onChange={(e) => onOtherAmountChange(Math.round(parseFloat(e.target.value || '0') * 100))}
            placeholder="0"
            style={{ width: 100 }}
          />
        </div>
      )}

      <div className="sr-denomination-total">
        {t('cash.total')}: {formatMinor(total)} ₽
      </div>
    </div>
  );
}

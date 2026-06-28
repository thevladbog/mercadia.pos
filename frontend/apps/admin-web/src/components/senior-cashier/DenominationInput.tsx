import { useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { Input } from '@mercadia/ui';

type DenominationEntry = { value: number; label: string; count: string };

const RUSSIAN_DENOMINATIONS: DenominationEntry[] = [
  { value: 500000, label: '5 000', count: '' },
  { value: 200000, label: '2 000', count: '' },
  { value: 100000, label: '1 000', count: '' },
  { value: 50000, label: '500', count: '' },
  { value: 20000, label: '200', count: '' },
  { value: 10000, label: '100', count: '' },
  { value: 5000, label: '50', count: '' },
  { value: 1000, label: '10', count: '' },
  { value: 500, label: '5', count: '' },
  { value: 200, label: '2', count: '' },
  { value: 100, label: '1', count: '' },
];

type DenominationInputProps = {
  values: Record<number, string>;
  onChange: (values: Record<number, string>) => void;
  otherAmountMinor?: string;
  onOtherAmountChange?: (value: string) => void;
};

export function DenominationInput({
  values,
  onChange,
  otherAmountMinor,
  onOtherAmountChange,
}: DenominationInputProps) {
  const { t } = useTranslation();

  const total = RUSSIAN_DENOMINATIONS.reduce(
    (sum, d) => sum + d.value * (parseInt(values[d.value] || '0', 10) || 0),
    0,
  );

  const handleCountChange = useCallback(
    (value: number, count: string) => {
      onChange({ ...values, [value]: count });
    },
    [values, onChange],
  );

  return (
    <div className="denomination-grid">
      {RUSSIAN_DENOMINATIONS.map((d) => (
        <div key={d.value} className="denomination-row">
          <span className="denomination-label">{d.label} ₽</span>
          <Input
            min={0}
            placeholder="0"
            type="number"
            value={values[d.value] ?? ''}
            onChange={(e) => handleCountChange(d.value, e.target.value)}
          />
        </div>
      ))}
      {onOtherAmountChange !== undefined ? (
        <div className="denomination-row">
          <span className="denomination-label">{t('seniorCashier.denominations')}</span>
          <Input
            min={0}
            placeholder="0"
            type="number"
            value={otherAmountMinor ?? ''}
            onChange={(e) => onOtherAmountChange(e.target.value)}
          />
        </div>
      ) : null}
      <div className="denomination-total">
        <strong>
          {t('seniorCashier.amount')}: {(total / 100).toFixed(2)} ₽
        </strong>
      </div>
    </div>
  );
}

export function computeDenominationTotal(values: Record<number, string>): number {
  return RUSSIAN_DENOMINATIONS.reduce(
    (sum, d) => sum + d.value * (parseInt(values[d.value] || '0', 10) || 0),
    0,
  );
}

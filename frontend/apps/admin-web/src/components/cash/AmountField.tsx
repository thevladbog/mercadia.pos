import { useTranslation } from 'react-i18next';

type AmountFieldProps = {
  value: string;
  onChange: (value: string) => void;
};

export function AmountField({ value, onChange }: AmountFieldProps) {
  const { t } = useTranslation();

  return (
    <label className="field">
      <span>{t('safe.forms.amountRub')}</span>
      <input
        inputMode="decimal"
        min="0"
        required
        step="0.01"
        type="number"
        value={value}
        onChange={(event) => onChange(event.target.value)}
      />
    </label>
  );
}

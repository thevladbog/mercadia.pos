import type { ListStores200StoresItem } from '@mercadia/api-clients-central';
import { Field, Label, Select } from '@mercadia/ui';
import { useId } from 'react';
import { useTranslation } from 'react-i18next';

type StorePickerProps = {
  stores: ListStores200StoresItem[];
  value: string;
  onChange: (storeId: string) => void;
  disabled?: boolean;
  loading?: boolean;
  className?: string;
};

export function StorePicker({
  stores,
  value,
  onChange,
  disabled = false,
  loading = false,
  className = 'store-picker',
}: StorePickerProps) {
  const { t } = useTranslation();
  const selectId = useId();

  return (
    <Field className={className}>
      <Label htmlFor={selectId}>{t('common.store')}</Label>
      <Select
        disabled={disabled || loading || stores.length === 0}
        id={selectId}
        value={value}
        onChange={(event) => onChange(event.target.value)}
      >
        {stores.length === 0 ? <option value="">{t('common.noStores')}</option> : null}
        {stores.map((store) => (
          <option key={store.id} value={store.id}>
            {store.name} ({store.id})
          </option>
        ))}
      </Select>
    </Field>
  );
}

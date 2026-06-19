import type { ListStores200StoresItem } from '@mercadia/api-clients-central';
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
  className = 'field store-picker',
}: StorePickerProps) {
  const { t } = useTranslation();

  return (
    <label className={className}>
      <span>{t('common.store')}</span>
      <select
        disabled={disabled || loading || stores.length === 0}
        value={value}
        onChange={(event) => onChange(event.target.value)}
      >
        {stores.length === 0 ? <option value="">{t('common.noStores')}</option> : null}
        {stores.map((store) => (
          <option key={store.id} value={store.id}>
            {store.name} ({store.id})
          </option>
        ))}
      </select>
    </label>
  );
}

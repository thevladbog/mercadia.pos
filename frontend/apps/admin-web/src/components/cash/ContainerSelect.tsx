import type { ListCashBalances200BalancesItem } from '@mercadia/api-clients-store-edge';
import { useTranslation } from 'react-i18next';

import { containerOptionLabel } from '@/pages/cash-container-utils.js';

type ContainerSelectProps = {
  label: string;
  containers: ListCashBalances200BalancesItem[];
  value: string;
  onChange: (containerId: string) => void;
  required?: boolean;
};

export function ContainerSelect({
  label,
  containers,
  value,
  onChange,
  required = true,
}: ContainerSelectProps) {
  const { t } = useTranslation();

  return (
    <label className="field">
      <span>{label}</span>
      <select
        required={required}
        disabled={containers.length === 0}
        value={value}
        onChange={(event) => onChange(event.target.value)}
      >
        {containers.length === 0 ? <option value="">{t('safe.forms.noContainers')}</option> : null}
        {containers.map((container) => (
          <option key={container.containerId} value={container.containerId}>
            {containerOptionLabel(container, t)}
          </option>
        ))}
      </select>
    </label>
  );
}

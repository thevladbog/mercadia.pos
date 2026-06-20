import { useTranslation } from 'react-i18next';

import type { SafeBalanceRollups } from '@/pages/safe-kpi-utils.js';
import { formatMinorAmount } from '@/pages/reporting-utils.js';

type SafeKpiPanelProps = {
  rollups: SafeBalanceRollups | null;
  movementsTotal: number | null;
  recountsTotal: number | null;
  openRecountCount: number | null;
  openRecountPartial: boolean;
  isLoading: boolean;
};

export function SafeKpiPanel({
  rollups,
  movementsTotal,
  recountsTotal,
  openRecountCount,
  openRecountPartial,
  isLoading,
}: SafeKpiPanelProps) {
  const { t } = useTranslation();

  return (
    <div className="panel">
      <h3>{t('safe.kpi.title')}</h3>
      {isLoading && !rollups ? (
        <p className="muted">{t('common.loading')}</p>
      ) : rollups ? (
        <>
          <dl className="kpi-grid">
            <div>
              <dt>{t('safe.kpi.safeTotal')}</dt>
              <dd>{formatMinorAmount(rollups.safeTotalMinor)}</dd>
            </div>
            <div>
              <dt>{t('safe.kpi.drawerTotal')}</dt>
              <dd>{formatMinorAmount(rollups.drawerTotalMinor)}</dd>
            </div>
            <div>
              <dt>{t('safe.kpi.bankTotal')}</dt>
              <dd>{formatMinorAmount(rollups.bankTotalMinor)}</dd>
            </div>
            <div>
              <dt>{t('safe.kpi.nonZeroDrawers')}</dt>
              <dd>{rollups.nonZeroDrawerCount}</dd>
            </div>
            <div>
              <dt>{t('safe.kpi.containers')}</dt>
              <dd>{rollups.containerCount}</dd>
            </div>
            <div>
              <dt>{t('safe.kpi.movementsTotal')}</dt>
              <dd>{movementsTotal ?? t('common.emDash')}</dd>
            </div>
            <div>
              <dt>{t('safe.kpi.recountsTotal')}</dt>
              <dd>{recountsTotal ?? t('common.emDash')}</dd>
            </div>
            <div>
              <dt>{t('safe.kpi.openRecountDiscrepancies')}</dt>
              <dd>{openRecountCount ?? t('common.emDash')}</dd>
            </div>
          </dl>
          {openRecountPartial ? (
            <p className="muted form-hint">{t('safe.kpi.openRecountPartial')}</p>
          ) : null}
        </>
      ) : (
        <p className="muted">{t('safe.noBalances')}</p>
      )}
    </div>
  );
}

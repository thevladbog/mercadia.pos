import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

import { ActionCard } from './ActionCard.js';
import { EyeIcon, IdCardIcon, ListIcon } from './icons.js';

export interface SystemRowProps {
  /** `undefined` while `useGetStoreMonitoringSummary` hasn't resolved yet. */
  totalTerminals: number | undefined;
  alertsCount: number | undefined;
  /** `undefined` while the operation-journal derivation isn't ready yet. */
  operationsCount: number | undefined;
}

/**
 * "СИСТЕМА" row (plan 021). "Удостоверения сотрудников" replaces the
 * design's EoD ("Конец опер. дня") card — EoD close is owned by admin-web,
 * not this app, and adding an entry point here is explicitly out of scope
 * (plan 021 "Why this matters" item 6). The monitoring card's subtitle and
 * badge, and the journal card's subtitle, are real derived data (items 8
 * and 9), not fabricated — both stay blank until their source data loads.
 */
export function SystemRow({ totalTerminals, alertsCount, operationsCount }: SystemRowProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();

  const monitoringSubtitle =
    totalTerminals !== undefined && alertsCount !== undefined
      ? t('dashboard.cards.monitoringSubtitle', { nodes: totalTerminals, alerts: alertsCount })
      : undefined;
  const journalSubtitle =
    operationsCount !== undefined
      ? t('dashboard.cards.journalSubtitle', { count: operationsCount })
      : undefined;

  return (
    <div className="sr-dashboard-action-grid sr-dashboard-action-grid--secondary">
      <ActionCard
        icon={<EyeIcon />}
        title={t('dashboard.cards.monitoringTitle')}
        subtitle={monitoringSubtitle}
        badgeText={alertsCount ? String(alertsCount) : undefined}
        onClick={() => navigate('/monitoring')}
      />
      <ActionCard
        icon={<IdCardIcon />}
        title={t('dashboard.credentials')}
        subtitle={t('dashboard.cards.credentialsSubtitle')}
        onClick={() => navigate('/credentials')}
      />
      <ActionCard
        icon={<ListIcon />}
        title={t('dashboard.cards.journalTitle')}
        subtitle={journalSubtitle}
        onClick={() => navigate('/journal')}
      />
    </div>
  );
}

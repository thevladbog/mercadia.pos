import { useGetCentralStatus } from '@mercadia/api-clients-central';
import { Button } from '@mercadia/ui';
import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { useAuth } from '@/auth/useAuth.js';
import { canManageCentralUsers } from '@/auth/permissions.js';
import { formatTimestamp } from './reporting-utils.js';

type QuickLink = {
  titleKey: string;
  descKey: string;
  to: string;
};

export function CentralDashboardPage() {
  const { t } = useTranslation();
  const { roles } = useAuth();
  const statusQuery = useGetCentralStatus({
    query: { refetchOnWindowFocus: false },
  });

  const quickLinks = useMemo((): QuickLink[] => {
    const links: QuickLink[] = [
      {
        titleKey: 'dashboard.reportingTitle',
        descKey: 'dashboard.reportingDesc',
        to: '/central/reporting',
      },
      {
        titleKey: 'dashboard.storesTitle',
        descKey: 'dashboard.storesDesc',
        to: '/central/stores',
      },
      {
        titleKey: 'dashboard.syncTitle',
        descKey: 'dashboard.syncDesc',
        to: '/central/sync',
      },
      {
        titleKey: 'dashboard.catalogTitle',
        descKey: 'dashboard.catalogDesc',
        to: '/central/catalog',
      },
      {
        titleKey: 'dashboard.monitoringTitle',
        descKey: 'dashboard.monitoringDesc',
        to: '/store/monitoring',
      },
      {
        titleKey: 'dashboard.safeTitle',
        descKey: 'dashboard.safeDesc',
        to: '/store/safe',
      },
      {
        titleKey: 'dashboard.eodTitle',
        descKey: 'dashboard.eodDesc',
        to: '/store/eod',
      },
    ];
    if (canManageCentralUsers(roles)) {
      links.push(
        {
          titleKey: 'dashboard.usersTitle',
          descKey: 'dashboard.usersDesc',
          to: '/central/users',
        },
        {
          titleKey: 'dashboard.colorSchemesTitle',
          descKey: 'dashboard.colorSchemesDesc',
          to: '/central/color-schemes',
        },
        {
          titleKey: 'dashboard.layoutTemplatesTitle',
          descKey: 'dashboard.layoutTemplatesDesc',
          to: '/central/layout-templates',
        },
      );
    }
    return links;
  }, [roles]);

  const status = statusQuery.data?.status === 200 ? statusQuery.data.data : null;
  const isLoading = statusQuery.isFetching;
  const errorMessage = statusQuery.error != null ? getApiErrorMessage(statusQuery.error) : null;

  return (
    <section className="stack reporting-page">
      <div className="panel">
        <div className="panel-heading">
          <div>
            <h2>{t('dashboard.title')}</h2>
            <p className="muted">{t('dashboard.subtitle')}</p>
          </div>
          <Button
            variant="secondary"
            disabled={isLoading}
            onClick={() => void statusQuery.refetch()}
            type="button"
          >
            {isLoading ? t('common.refreshing') : t('common.refresh')}
          </Button>
        </div>
      </div>

      {errorMessage ? (
        <div className="panel error-panel">
          <p className="error">{errorMessage}</p>
        </div>
      ) : null}

      <div className="panel">
        <h3>{t('dashboard.centralStatus')}</h3>
        {statusQuery.isLoading && !status ? (
          <p className="muted">{t('dashboard.loadingStatus')}</p>
        ) : status ? (
          <dl className="kpi-grid">
            <div>
              <dt>{t('dashboard.region')}</dt>
              <dd>{status.region}</dd>
            </div>
            <div>
              <dt>{t('dashboard.status')}</dt>
              <dd>{status.status}</dd>
            </div>
            <div>
              <dt>{t('dashboard.registeredStores')}</dt>
              <dd>{status.storeCount}</dd>
            </div>
            <div>
              <dt>{t('dashboard.lastUpdated')}</dt>
              <dd>{formatTimestamp(status.generatedAt)}</dd>
            </div>
          </dl>
        ) : (
          <p className="muted">{t('dashboard.noStatus')}</p>
        )}
      </div>

      <div className="panel">
        <h3>{t('dashboard.quickLinks')}</h3>
        <div className="stack">
          {quickLinks.map((link) => (
            <div key={link.to}>
              <Link to={link.to}>{t(link.titleKey)}</Link>
              <p className="muted">{t(link.descKey)}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

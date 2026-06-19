import { useListStores } from '@mercadia/api-clients-central';
import { useGetTerminal, useListStoreMonitoringTerminals } from '@mercadia/api-clients-store-edge';
import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { useParams } from 'react-router-dom';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { monitoringExplorerHref } from './monitoring-routes.js';
import {
  MONITORING_REFRESH_INTERVAL_MS,
  terminalStatusClass,
  terminalStatusLabel,
} from './monitoring-utils.js';
import { formatMinorAmount, formatTimestamp } from './reporting-utils.js';
import { TerminalHeartbeatEventsPanel } from './TerminalHeartbeatEventsPanel.js';
import { PageBackLink } from './users-shared.js';

export function TerminalMonitoringDetailPage() {
  const { t } = useTranslation();
  const { storeId = '', terminalId = '' } = useParams();

  const queryOptions = useMemo(
    () => ({
      query: {
        enabled: storeId.length > 0 && terminalId.length > 0,
        refetchInterval: MONITORING_REFRESH_INTERVAL_MS,
      },
    }),
    [storeId, terminalId],
  );

  const terminalQueryOptions = useMemo(
    () => ({
      query: {
        enabled: terminalId.length > 0,
        refetchInterval: MONITORING_REFRESH_INTERVAL_MS,
      },
    }),
    [terminalId],
  );

  const storesQuery = useListStores();
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data.stores : [];
  const storeName = stores.find((store) => store.id === storeId)?.name;

  const terminalQuery = useGetTerminal(terminalId, terminalQueryOptions);
  const terminalsQuery = useListStoreMonitoringTerminals(storeId, undefined, queryOptions);

  const terminal = terminalQuery.data?.status === 200 ? terminalQuery.data.data : null;
  const monitoringCard = useMemo(() => {
    if (terminalsQuery.data?.status !== 200) {
      return null;
    }
    return terminalsQuery.data.data.items.find((item) => item.id === terminalId) ?? null;
  }, [terminalsQuery.data, terminalId]);

  const isLoading = terminalQuery.isFetching || terminalsQuery.isFetching;
  const errorMessage =
    terminalQuery.error != null
      ? getApiErrorMessage(terminalQuery.error)
      : terminalsQuery.error != null
        ? getApiErrorMessage(terminalsQuery.error)
        : null;

  const title = storeName ? `${storeName} (${storeId})` : storeId;
  const backHref = monitoringExplorerHref(storeId.length > 0 ? storeId : undefined);
  const statusSource = monitoringCard ?? terminal;

  function refetchAll() {
    void storesQuery.refetch();
    void terminalQuery.refetch();
    if (storeId.length > 0) {
      void terminalsQuery.refetch();
    }
  }

  return (
    <section className="stack monitoring-page">
      <PageBackLink label={t('monitoring.backToMonitoring')} to={backHref} />

      <div className="panel">
        <div className="panel-heading">
          <div>
            <h2>{t('monitoring.terminalDetailTitle')}</h2>
            <p className="muted">
              {terminalId || t('common.noData')} {t('common.emDash')}{' '}
              {title || t('reporting.unknownStore')}
            </p>
          </div>
          <button
            className="secondary"
            disabled={isLoading || terminalId.length === 0}
            onClick={refetchAll}
            type="button"
          >
            {isLoading ? t('common.refreshing') : t('common.refresh')}
          </button>
        </div>
      </div>

      {errorMessage ? (
        <div className="panel error-panel">
          <p className="error">{errorMessage}</p>
        </div>
      ) : null}

      <div className="panel">
        <h3>{t('monitoring.terminalState')}</h3>
        {terminalQuery.isLoading && !terminal ? (
          <p className="muted">{t('monitoring.loadingTerminal')}</p>
        ) : terminal ? (
          <dl className="kpi-grid">
            <div>
              <dt>ID</dt>
              <dd>{terminal.id}</dd>
            </div>
            <div>
              <dt>{t('stores.storeId')}</dt>
              <dd>{terminal.storeId}</dd>
            </div>
            <div>
              <dt>{t('monitoring.kind')}</dt>
              <dd>{terminal.kind}</dd>
            </div>
            <div>
              <dt>{t('monitoring.status')}</dt>
              <dd>
                {statusSource ? (
                  <span className={terminalStatusClass(statusSource)}>
                    {terminalStatusLabel(statusSource)}
                  </span>
                ) : (
                  terminal.status
                )}
              </dd>
            </div>
            <div>
              <dt>{t('monitoring.softwareVersion')}</dt>
              <dd>{terminal.softwareVersion ?? t('common.emDash')}</dd>
            </div>
            <div>
              <dt>{t('monitoring.lastSeen')}</dt>
              <dd>{formatTimestamp(terminal.lastSeenAt)}</dd>
            </div>
            <div>
              <dt>{t('monitoring.updated')}</dt>
              <dd>{formatTimestamp(terminal.updatedAt)}</dd>
            </div>
          </dl>
        ) : (
          <p className="muted">{t('monitoring.noTerminal')}</p>
        )}
      </div>

      <div className="panel">
        <h3>{t('monitoring.liveOperations')}</h3>
        {terminalsQuery.isLoading && !monitoringCard ? (
          <p className="muted">{t('monitoring.loadingMonitoring')}</p>
        ) : monitoringCard ? (
          <dl className="kpi-grid">
            <div>
              <dt>{t('monitoring.cashier')}</dt>
              <dd>{monitoringCard.cashierId ?? t('common.emDash')}</dd>
            </div>
            <div>
              <dt>{t('monitoring.shift')}</dt>
              <dd>{monitoringCard.shiftId ?? t('common.emDash')}</dd>
            </div>
            <div>
              <dt>{t('monitoring.drawer')}</dt>
              <dd>{monitoringCard.drawerId ?? t('common.emDash')}</dd>
            </div>
            <div>
              <dt>{t('monitoring.receiptCount')}</dt>
              <dd>{monitoringCard.receiptCount}</dd>
            </div>
            <div>
              <dt>{t('monitoring.revenue')}</dt>
              <dd>{formatMinorAmount(monitoringCard.revenueMinor)}</dd>
            </div>
            <div>
              <dt>{t('monitoring.drawerBalance')}</dt>
              <dd>{formatMinorAmount(monitoringCard.drawerBalanceMinor)}</dd>
            </div>
            <div>
              <dt>{t('monitoring.attentionNeeded')}</dt>
              <dd>{monitoringCard.attentionNeeded ? t('common.yes') : t('common.no')}</dd>
            </div>
            <div>
              <dt>{t('monitoring.currentReceipt')}</dt>
              <dd>{monitoringCard.currentReceiptId ?? t('common.emDash')}</dd>
            </div>
            <div>
              <dt>{t('monitoring.currentReceiptStatus')}</dt>
              <dd>{monitoringCard.currentReceiptStatus ?? t('common.emDash')}</dd>
            </div>
            <div>
              <dt>{t('monitoring.currentReceiptTotal')}</dt>
              <dd>
                {monitoringCard.currentReceiptTotalMinor != null
                  ? formatMinorAmount(monitoringCard.currentReceiptTotalMinor)
                  : t('common.emDash')}
              </dd>
            </div>
          </dl>
        ) : (
          <p className="muted">{t('monitoring.noLiveMonitoring')}</p>
        )}
      </div>

      <TerminalHeartbeatEventsPanel
        maxEvents={10}
        storeId={storeId}
        terminalId={terminalId}
        title={t('monitoring.recentHeartbeats')}
      />
    </section>
  );
}

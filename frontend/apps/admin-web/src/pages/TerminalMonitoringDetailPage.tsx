import { useListStores } from '@mercadia/api-clients-central';
import { useGetTerminal, useListStoreMonitoringTerminals } from '@mercadia/api-clients-store-edge';
import { useMemo } from 'react';
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
      <PageBackLink label="Back to monitoring" to={backHref} />

      <div className="panel">
        <div className="panel-heading">
          <div>
            <h2>Terminal</h2>
            <p className="muted">
              {terminalId || 'Unknown terminal'} — {title || 'Unknown store'}
            </p>
          </div>
          <button
            className="secondary"
            disabled={isLoading || terminalId.length === 0}
            onClick={refetchAll}
            type="button"
          >
            {isLoading ? 'Refreshing…' : 'Refresh'}
          </button>
        </div>
      </div>

      {errorMessage ? (
        <div className="panel error-panel">
          <p className="error">{errorMessage}</p>
        </div>
      ) : null}

      <div className="panel">
        <h3>Terminal state</h3>
        {terminalQuery.isLoading && !terminal ? (
          <p className="muted">Loading terminal…</p>
        ) : terminal ? (
          <dl className="kpi-grid">
            <div>
              <dt>ID</dt>
              <dd>{terminal.id}</dd>
            </div>
            <div>
              <dt>Store ID</dt>
              <dd>{terminal.storeId}</dd>
            </div>
            <div>
              <dt>Kind</dt>
              <dd>{terminal.kind}</dd>
            </div>
            <div>
              <dt>Status</dt>
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
              <dt>Software version</dt>
              <dd>{terminal.softwareVersion ?? '—'}</dd>
            </div>
            <div>
              <dt>Last seen</dt>
              <dd>{formatTimestamp(terminal.lastSeenAt)}</dd>
            </div>
            <div>
              <dt>Updated</dt>
              <dd>{formatTimestamp(terminal.updatedAt)}</dd>
            </div>
          </dl>
        ) : (
          <p className="muted">No terminal data.</p>
        )}
      </div>

      <div className="panel">
        <h3>Live operations</h3>
        {terminalsQuery.isLoading && !monitoringCard ? (
          <p className="muted">Loading monitoring data…</p>
        ) : monitoringCard ? (
          <dl className="kpi-grid">
            <div>
              <dt>Cashier</dt>
              <dd>{monitoringCard.cashierId ?? '—'}</dd>
            </div>
            <div>
              <dt>Shift</dt>
              <dd>{monitoringCard.shiftId ?? '—'}</dd>
            </div>
            <div>
              <dt>Drawer</dt>
              <dd>{monitoringCard.drawerId ?? '—'}</dd>
            </div>
            <div>
              <dt>Receipt count</dt>
              <dd>{monitoringCard.receiptCount}</dd>
            </div>
            <div>
              <dt>Revenue</dt>
              <dd>{formatMinorAmount(monitoringCard.revenueMinor)}</dd>
            </div>
            <div>
              <dt>Drawer balance</dt>
              <dd>{formatMinorAmount(monitoringCard.drawerBalanceMinor)}</dd>
            </div>
            <div>
              <dt>Attention needed</dt>
              <dd>{monitoringCard.attentionNeeded ? 'Yes' : 'No'}</dd>
            </div>
            <div>
              <dt>Current receipt</dt>
              <dd>{monitoringCard.currentReceiptId ?? '—'}</dd>
            </div>
            <div>
              <dt>Current receipt status</dt>
              <dd>{monitoringCard.currentReceiptStatus ?? '—'}</dd>
            </div>
            <div>
              <dt>Current receipt total</dt>
              <dd>
                {monitoringCard.currentReceiptTotalMinor != null
                  ? formatMinorAmount(monitoringCard.currentReceiptTotalMinor)
                  : '—'}
              </dd>
            </div>
          </dl>
        ) : (
          <p className="muted">No live monitoring data for this terminal.</p>
        )}
      </div>

      <TerminalHeartbeatEventsPanel
        maxEvents={10}
        storeId={storeId}
        terminalId={terminalId}
        title="Recent heartbeats"
      />
    </section>
  );
}

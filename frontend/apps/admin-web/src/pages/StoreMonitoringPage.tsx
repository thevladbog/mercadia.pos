import { useListStores } from '@mercadia/api-clients-central';
import {
  useGetStoreMonitoringSummary,
  useListStoreMonitoringTerminals,
} from '@mercadia/api-clients-store-edge';
import { useMemo, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { terminalMonitoringHref } from './monitoring-routes.js';
import {
  MONITORING_REFRESH_INTERVAL_MS,
  terminalStatusClass,
  terminalStatusLabel,
} from './monitoring-utils.js';
import { formatMinorAmount, formatTimestamp } from './reporting-utils.js';

export function StoreMonitoringPage() {
  const [searchParams] = useSearchParams();
  const initialStoreId = searchParams.get('store');

  const storesQuery = useListStores();
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data.stores : [];
  const [selectedStoreId, setSelectedStoreId] = useState<string | null>(initialStoreId);
  const activeStoreId = selectedStoreId ?? stores[0]?.id ?? '';

  const queryOptions = useMemo(
    () => ({
      query: {
        enabled: activeStoreId.length > 0,
        refetchInterval: MONITORING_REFRESH_INTERVAL_MS,
      },
    }),
    [activeStoreId],
  );

  const summaryQuery = useGetStoreMonitoringSummary(activeStoreId, queryOptions);
  const terminalsQuery = useListStoreMonitoringTerminals(activeStoreId, undefined, queryOptions);

  const summary = summaryQuery.data?.status === 200 ? summaryQuery.data.data : null;
  const terminals = terminalsQuery.data?.status === 200 ? terminalsQuery.data.data.items : null;

  const isLoading =
    storesQuery.isFetching ||
    (activeStoreId.length > 0 && (summaryQuery.isFetching || terminalsQuery.isFetching));

  const errorMessage =
    storesQuery.error != null
      ? getApiErrorMessage(storesQuery.error)
      : summaryQuery.error != null
        ? getApiErrorMessage(summaryQuery.error)
        : terminalsQuery.error != null
          ? getApiErrorMessage(terminalsQuery.error)
          : null;

  function refetchAll() {
    void storesQuery.refetch();
    if (activeStoreId.length > 0) {
      void summaryQuery.refetch();
      void terminalsQuery.refetch();
    }
  }

  return (
    <section className="stack monitoring-page">
      <div className="panel">
        <div className="panel-heading">
          <div>
            <h2>Store Monitoring</h2>
            <p className="muted">
              Live terminal status and store KPIs (auto-refresh every 5 seconds).
            </p>
          </div>
          <button className="secondary" disabled={isLoading} onClick={refetchAll} type="button">
            {isLoading ? 'Refreshing…' : 'Refresh'}
          </button>
        </div>

        <label className="field store-picker">
          <span>Store</span>
          <select
            disabled={storesQuery.isLoading || stores.length === 0}
            value={activeStoreId}
            onChange={(event) => setSelectedStoreId(event.target.value)}
          >
            {stores.length === 0 ? <option value="">No stores registered</option> : null}
            {stores.map((store) => (
              <option key={store.id} value={store.id}>
                {store.name} ({store.id})
              </option>
            ))}
          </select>
        </label>
      </div>

      {errorMessage ? (
        <div className="panel error-panel">
          <p className="error">{errorMessage}</p>
        </div>
      ) : null}

      {!activeStoreId ? (
        <div className="panel">
          <p className="muted">Select a store to view monitoring data.</p>
        </div>
      ) : (
        <>
          <div className="panel">
            <h3>Store KPIs</h3>
            {summaryQuery.isLoading && !summary ? (
              <p className="muted">Loading summary…</p>
            ) : summary ? (
              <dl className="kpi-grid">
                <div>
                  <dt>Today&apos;s revenue</dt>
                  <dd>{formatMinorAmount(summary.revenueMinorToday)}</dd>
                </div>
                <div>
                  <dt>Money in drawers</dt>
                  <dd>{formatMinorAmount(summary.drawerCashMinor)}</dd>
                </div>
                <div>
                  <dt>Active terminals</dt>
                  <dd>{summary.activeTerminalCount}</dd>
                </div>
                <div>
                  <dt>Free terminals</dt>
                  <dd>{summary.freeTerminalCount}</dd>
                </div>
                <div>
                  <dt>Attention needed</dt>
                  <dd>{summary.attentionTerminalCount}</dd>
                </div>
                <div>
                  <dt>Offline terminals</dt>
                  <dd>{summary.offlineTerminalCount}</dd>
                </div>
                <div>
                  <dt>Receipt count</dt>
                  <dd>{summary.receiptCountToday}</dd>
                </div>
                <div>
                  <dt>Average check</dt>
                  <dd>{formatMinorAmount(summary.averageReceiptMinor)}</dd>
                </div>
              </dl>
            ) : (
              <p className="muted">No summary data.</p>
            )}
          </div>

          <div className="panel">
            <div className="panel-heading">
              <h3>Terminals</h3>
              <p className="muted">
                {terminals ? `${terminals.length} terminal(s)` : 'Loading terminals…'}
              </p>
            </div>

            {terminalsQuery.isLoading && !terminals ? (
              <p className="muted">Loading terminals…</p>
            ) : terminals && terminals.length > 0 ? (
              <div className="table-wrap">
                <table>
                  <thead>
                    <tr>
                      <th>Terminal</th>
                      <th>Kind</th>
                      <th>Status</th>
                      <th>Cashier</th>
                      <th>Receipts</th>
                      <th>Revenue</th>
                      <th>Drawer</th>
                      <th>Attention</th>
                      <th>Last seen</th>
                    </tr>
                  </thead>
                  <tbody>
                    {terminals.map((terminal) => (
                      <tr key={terminal.id}>
                        <td>
                          <Link to={terminalMonitoringHref(activeStoreId, terminal.id)}>
                            {terminal.id}
                          </Link>
                        </td>
                        <td>{terminal.kind}</td>
                        <td>
                          <span className={terminalStatusClass(terminal)}>
                            {terminalStatusLabel(terminal)}
                          </span>
                        </td>
                        <td>{terminal.cashierId ?? '—'}</td>
                        <td>{terminal.receiptCount}</td>
                        <td>{formatMinorAmount(terminal.revenueMinor)}</td>
                        <td>{formatMinorAmount(terminal.drawerBalanceMinor)}</td>
                        <td>{terminal.attentionNeeded ? 'Yes' : 'No'}</td>
                        <td>{formatTimestamp(terminal.lastSeenAt)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <p className="muted">No terminals found for this store.</p>
            )}
          </div>
        </>
      )}
    </section>
  );
}

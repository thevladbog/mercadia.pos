import { useListStores } from '@mercadia/api-clients-central';
import {
  useGetStoreMonitoringSummary,
  useListStoreMonitoringTerminals,
} from '@mercadia/api-clients-store-edge';
import { useMemo, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { StorePicker } from '@/components/StorePicker.js';
import { terminalMonitoringHref } from './monitoring-routes.js';
import {
  STORE_POLL_INTERVAL_MS,
  terminalStatusClass,
  terminalStatusLabel,
} from './monitoring-utils.js';
import { formatMinorAmount, formatTimestamp } from './reporting-utils.js';
import { readStoreFromSearchParams } from './store-routes.js';
import { TerminalCardGrid } from './TerminalCardGrid.js';
import { TerminalHeartbeatEventsPanel } from './TerminalHeartbeatEventsPanel.js';

type TerminalView = 'list' | 'tiles';

function matchesTerminalSearch(
  terminal: {
    id: string;
    kind: string;
    status: string;
    cashierId?: string;
    attentionNeeded: boolean;
  },
  query: string,
): boolean {
  if (query.length === 0) {
    return true;
  }
  const haystack = [
    terminal.id,
    terminal.kind,
    terminal.status,
    terminal.cashierId ?? '',
    terminal.attentionNeeded ? 'attention' : '',
  ]
    .join(' ')
    .toLowerCase();
  return haystack.includes(query);
}

export function StoreMonitoringPage() {
  const { t } = useTranslation();
  const [searchParams] = useSearchParams();
  const initialStoreId = readStoreFromSearchParams(searchParams);

  const storesQuery = useListStores();
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data.stores : [];
  const [selectedStoreId, setSelectedStoreId] = useState<string | null>(initialStoreId);
  const activeStoreId = selectedStoreId ?? stores[0]?.id ?? '';
  const [terminalView, setTerminalView] = useState<TerminalView>('list');
  const [searchQuery, setSearchQuery] = useState('');

  const queryOptions = useMemo(
    () => ({
      query: {
        enabled: activeStoreId.length > 0,
        refetchInterval: STORE_POLL_INTERVAL_MS,
      },
    }),
    [activeStoreId],
  );

  const summaryQuery = useGetStoreMonitoringSummary(activeStoreId, queryOptions);
  const terminalsQuery = useListStoreMonitoringTerminals(activeStoreId, undefined, queryOptions);

  const summary = summaryQuery.data?.status === 200 ? summaryQuery.data.data : null;
  const terminals = terminalsQuery.data?.status === 200 ? terminalsQuery.data.data.items : null;
  const normalizedSearch = searchQuery.trim().toLowerCase();
  const filteredTerminals = useMemo(
    () =>
      terminals?.filter((terminal) => matchesTerminalSearch(terminal, normalizedSearch)) ?? null,
    [terminals, normalizedSearch],
  );

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
            <h2>{t('monitoring.title')}</h2>
            <p className="muted">{t('monitoring.subtitle')}</p>
          </div>
          <button className="secondary" disabled={isLoading} onClick={refetchAll} type="button">
            {isLoading ? t('common.refreshing') : t('common.refresh')}
          </button>
        </div>

        <StorePicker
          loading={storesQuery.isLoading}
          stores={stores}
          value={activeStoreId}
          onChange={setSelectedStoreId}
        />
      </div>

      {errorMessage ? (
        <div className="panel error-panel">
          <p className="error">{errorMessage}</p>
        </div>
      ) : null}

      {!activeStoreId ? (
        <div className="panel">
          <p className="muted">{t('monitoring.selectStore')}</p>
        </div>
      ) : (
        <>
          <div className="panel">
            <h3>{t('monitoring.storeKpis')}</h3>
            {summaryQuery.isLoading && !summary ? (
              <p className="muted">{t('monitoring.loadingSummary')}</p>
            ) : summary ? (
              <dl className="kpi-grid">
                <div>
                  <dt>{t('monitoring.revenueToday')}</dt>
                  <dd>{formatMinorAmount(summary.revenueMinorToday)}</dd>
                </div>
                <div>
                  <dt>{t('monitoring.drawerCash')}</dt>
                  <dd>{formatMinorAmount(summary.drawerCashMinor)}</dd>
                </div>
                <div>
                  <dt>{t('monitoring.activeTerminals')}</dt>
                  <dd>{summary.activeTerminalCount}</dd>
                </div>
                <div>
                  <dt>{t('monitoring.freeTerminals')}</dt>
                  <dd>{summary.freeTerminalCount}</dd>
                </div>
                <div>
                  <dt>{t('monitoring.attentionTerminals')}</dt>
                  <dd>{summary.attentionTerminalCount}</dd>
                </div>
                <div>
                  <dt>{t('monitoring.offlineTerminals')}</dt>
                  <dd>{summary.offlineTerminalCount}</dd>
                </div>
                <div>
                  <dt>{t('monitoring.receiptCount')}</dt>
                  <dd>{summary.receiptCountToday}</dd>
                </div>
                <div>
                  <dt>{t('monitoring.averageCheck')}</dt>
                  <dd>{formatMinorAmount(summary.averageReceiptMinor)}</dd>
                </div>
              </dl>
            ) : (
              <p className="muted">{t('monitoring.noSummary')}</p>
            )}
          </div>

          <div className="panel">
            <div className="panel-heading">
              <div>
                <h3>{t('monitoring.terminals')}</h3>
                <p className="muted">
                  {terminals
                    ? t('monitoring.terminalCount', { count: filteredTerminals?.length ?? 0 })
                    : t('monitoring.loadingTerminals')}
                </p>
              </div>
              <div className="view-toggle">
                <button
                  className={terminalView === 'list' ? undefined : 'secondary'}
                  onClick={() => setTerminalView('list')}
                  type="button"
                >
                  {t('monitoring.view.list')}
                </button>
                <button
                  className={terminalView === 'tiles' ? undefined : 'secondary'}
                  onClick={() => setTerminalView('tiles')}
                  type="button"
                >
                  {t('monitoring.view.tiles')}
                </button>
              </div>
            </div>

            <label className="field terminal-search">
              <span>{t('monitoring.searchHint')}</span>
              <input
                placeholder={t('monitoring.searchPlaceholder')}
                type="search"
                value={searchQuery}
                onChange={(event) => setSearchQuery(event.target.value)}
              />
            </label>

            {terminalsQuery.isLoading && !terminals ? (
              <p className="muted">{t('monitoring.loadingTerminals')}</p>
            ) : filteredTerminals && filteredTerminals.length > 0 ? (
              terminalView === 'tiles' ? (
                <TerminalCardGrid storeId={activeStoreId} terminals={filteredTerminals} />
              ) : (
                <div className="table-wrap">
                  <table>
                    <thead>
                      <tr>
                        <th>{t('monitoring.terminal')}</th>
                        <th>{t('monitoring.kind')}</th>
                        <th>{t('monitoring.status')}</th>
                        <th>{t('monitoring.cashier')}</th>
                        <th>{t('monitoring.receipts')}</th>
                        <th>{t('monitoring.revenue')}</th>
                        <th>{t('monitoring.drawer')}</th>
                        <th>{t('monitoring.attention')}</th>
                        <th>{t('monitoring.lastSeen')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {filteredTerminals.map((terminal) => (
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
                          <td>{terminal.cashierId ?? t('common.emDash')}</td>
                          <td>{terminal.receiptCount}</td>
                          <td>{formatMinorAmount(terminal.revenueMinor)}</td>
                          <td>{formatMinorAmount(terminal.drawerBalanceMinor)}</td>
                          <td>{terminal.attentionNeeded ? t('common.yes') : t('common.no')}</td>
                          <td>{formatTimestamp(terminal.lastSeenAt)}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )
            ) : (
              <p className="muted">{t('monitoring.noTerminals')}</p>
            )}
          </div>

          <TerminalHeartbeatEventsPanel storeId={activeStoreId} />
        </>
      )}
    </section>
  );
}

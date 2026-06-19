import {
  type ListStoreCashMovements200ItemsItem,
  type ListStoreFiscalDocuments200ItemsItem,
  type ListStoreOperationalDays200ItemsItem,
  type ListStorePayments200ItemsItem,
  type ListStoreReturns200ItemsItem,
  type ListStoreSyncEvents200ItemsItem,
  useListStoreCashMovements,
  useListStoreFiscalDocuments,
  useListStoreOperationalDays,
  useListStorePayments,
  useListStoreReturns,
  useListStoreSyncEvents,
  useListStores,
} from '@mercadia/api-clients-central';
import { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { TFunction } from 'i18next';
import { Link, useSearchParams } from 'react-router-dom';

import { PaginationControls } from '@/components/PaginationControls.js';
import { StorePicker } from '@/components/StorePicker.js';
import { getApiErrorMessage } from '@/auth/api-errors.js';
import { formatMinorAmount, formatTimestamp, PAGE_SIZE } from './reporting-utils.js';
import { parseSyncTab, syncEntityHref, type SyncEntityType, type SyncTab } from './sync-routes.js';

const SYNC_TABS: { id: SyncTab; labelKey: string }[] = [
  { id: 'sync-events', labelKey: 'sync.tabs.syncEvents' },
  { id: 'payments', labelKey: 'sync.tabs.payments' },
  { id: 'cash-movements', labelKey: 'sync.tabs.cashMovements' },
  { id: 'fiscal-documents', labelKey: 'sync.tabs.fiscalDocuments' },
  { id: 'returns', labelKey: 'sync.tabs.returns' },
  { id: 'operational-days', labelKey: 'sync.tabs.operationalDays' },
];

function syncTabLabel(t: TFunction, tabId: SyncTab): string {
  const tab = SYNC_TABS.find((entry) => entry.id === tabId);
  return tab ? t(tab.labelKey) : tabId;
}

export function CentralSyncExplorerPage() {
  const { t } = useTranslation();
  const [searchParams] = useSearchParams();
  const initialTab = parseSyncTab(searchParams.get('tab')) ?? 'sync-events';
  const initialStoreId = searchParams.get('store');

  const storesQuery = useListStores();
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data.stores : [];
  const [selectedStoreId, setSelectedStoreId] = useState<string | null>(initialStoreId);
  const activeStoreId = selectedStoreId ?? stores[0]?.id ?? '';
  const [activeTab, setActiveTab] = useState<SyncTab>(initialTab);
  const [offset, setOffset] = useState(0);

  const listParams = useMemo(() => ({ limit: PAGE_SIZE, offset }), [offset]);
  const queryEnabled = activeStoreId.length > 0;

  const syncEventsQuery = useListStoreSyncEvents(activeStoreId, listParams, {
    query: { enabled: queryEnabled && activeTab === 'sync-events' },
  });
  const paymentsQuery = useListStorePayments(activeStoreId, listParams, {
    query: { enabled: queryEnabled && activeTab === 'payments' },
  });
  const cashMovementsQuery = useListStoreCashMovements(activeStoreId, listParams, {
    query: { enabled: queryEnabled && activeTab === 'cash-movements' },
  });
  const fiscalDocumentsQuery = useListStoreFiscalDocuments(activeStoreId, listParams, {
    query: { enabled: queryEnabled && activeTab === 'fiscal-documents' },
  });
  const returnsQuery = useListStoreReturns(activeStoreId, listParams, {
    query: { enabled: queryEnabled && activeTab === 'returns' },
  });
  const operationalDaysQuery = useListStoreOperationalDays(activeStoreId, listParams, {
    query: { enabled: queryEnabled && activeTab === 'operational-days' },
  });

  const activeQuery = useMemo(() => {
    switch (activeTab) {
      case 'sync-events':
        return syncEventsQuery;
      case 'payments':
        return paymentsQuery;
      case 'cash-movements':
        return cashMovementsQuery;
      case 'fiscal-documents':
        return fiscalDocumentsQuery;
      case 'returns':
        return returnsQuery;
      case 'operational-days':
        return operationalDaysQuery;
    }
  }, [
    activeTab,
    syncEventsQuery,
    paymentsQuery,
    cashMovementsQuery,
    fiscalDocumentsQuery,
    returnsQuery,
    operationalDaysQuery,
  ]);

  const pageData = activeQuery.data?.status === 200 ? activeQuery.data.data : null;
  const totalCount = pageData?.totalCount ?? 0;
  const pageStart = totalCount === 0 ? 0 : offset + 1;
  const pageEnd = Math.min(offset + PAGE_SIZE, totalCount);
  const canGoPrev = offset > 0;
  const canGoNext = offset + PAGE_SIZE < totalCount;

  const isLoading = storesQuery.isFetching || activeQuery.isFetching;
  const errorMessage =
    storesQuery.error != null
      ? getApiErrorMessage(storesQuery.error)
      : activeQuery.error != null
        ? getApiErrorMessage(activeQuery.error)
        : null;

  function refetchAll() {
    void storesQuery.refetch();
    if (queryEnabled) {
      void activeQuery.refetch();
    }
  }

  return (
    <section className="stack reporting-page">
      <div className="panel">
        <div className="panel-heading">
          <div>
            <h2>{t('sync.title')}</h2>
            <p className="muted">{t('sync.subtitle')}</p>
          </div>
          <button className="secondary" disabled={isLoading} onClick={refetchAll} type="button">
            {isLoading ? t('common.refreshing') : t('common.refresh')}
          </button>
        </div>

        <StorePicker
          disabled={storesQuery.isLoading}
          loading={storesQuery.isLoading}
          stores={stores}
          value={activeStoreId}
          onChange={(storeId) => {
            setSelectedStoreId(storeId);
            setOffset(0);
          }}
        />

        <div className="filters" role="tablist" aria-label={t('sync.title')}>
          {SYNC_TABS.map((tab) => (
            <button
              key={tab.id}
              id={`sync-tab-${tab.id}`}
              className={activeTab === tab.id ? undefined : 'secondary'}
              onClick={() => {
                setActiveTab(tab.id);
                setOffset(0);
              }}
              role="tab"
              type="button"
              aria-selected={activeTab === tab.id}
              aria-controls={`sync-tab-panel-${tab.id}`}
            >
              {t(tab.labelKey)}
            </button>
          ))}
        </div>
      </div>

      {errorMessage ? (
        <div className="panel error-panel">
          <p className="error">{errorMessage}</p>
        </div>
      ) : null}

      {!activeStoreId ? (
        <div className="panel">
          <p className="muted">{t('sync.selectStore')}</p>
        </div>
      ) : (
        <div
          className="panel"
          id={`sync-tab-panel-${activeTab}`}
          role="tabpanel"
          aria-labelledby={`sync-tab-${activeTab}`}
        >
          <div className="panel-heading">
            <h3>{syncTabLabel(t, activeTab)}</h3>
            <p className="muted">
              {totalCount === 0
                ? t('common.noItems')
                : t('common.showingRange', { from: pageStart, to: pageEnd, total: totalCount })}
            </p>
          </div>

          {activeQuery.isLoading && !pageData ? (
            <p className="muted">{t('common.loading')}</p>
          ) : pageData && pageData.items.length > 0 ? (
            <>
              <div className="table-wrap">
                <table>
                  <thead>{renderTableHead(activeTab, t)}</thead>
                  <tbody>
                    {renderTableBody(
                      activeTab,
                      activeStoreId,
                      pageData.items as SyncExplorerItem[],
                    )}
                  </tbody>
                </table>
              </div>
              <PaginationControls
                canGoNext={canGoNext}
                canGoPrev={canGoPrev}
                disabled={isLoading}
                onNext={() => setOffset((current) => current + PAGE_SIZE)}
                onPrev={() => setOffset((current) => Math.max(0, current - PAGE_SIZE))}
              />
            </>
          ) : (
            <p className="muted">{t('sync.noItemsProjection')}</p>
          )}
        </div>
      )}
    </section>
  );
}

function renderTableHead(activeTab: SyncTab, t: TFunction) {
  switch (activeTab) {
    case 'sync-events':
      return (
        <tr>
          <th>Source event ID</th>
          <th>Event type</th>
          <th>Occurred</th>
          <th>{t('monitoring.eventReceived')}</th>
        </tr>
      );
    case 'payments':
      return (
        <tr>
          <th>Payment ID</th>
          <th>Method</th>
          <th>{t('safe.amount')}</th>
          <th>{t('monitoring.status')}</th>
          <th>Captured</th>
        </tr>
      );
    case 'cash-movements':
      return (
        <tr>
          <th>{t('safe.movementId')}</th>
          <th>{t('safe.type')}</th>
          <th>{t('safe.amount')}</th>
          <th>{t('safe.posted')}</th>
        </tr>
      );
    case 'fiscal-documents':
      return (
        <tr>
          <th>Document ID</th>
          <th>{t('monitoring.kind')}</th>
          <th>{t('safe.amount')}</th>
          <th>Fiscalized</th>
        </tr>
      );
    case 'returns':
      return (
        <tr>
          <th>Return ID</th>
          <th>Receipt ID</th>
          <th>Total</th>
          <th>Settled</th>
        </tr>
      );
    case 'operational-days':
      return (
        <tr>
          <th>Day ID</th>
          <th>{t('eod.businessDate')}</th>
          <th>{t('eod.closedAt')}</th>
          <th>Closed by</th>
        </tr>
      );
  }
}

type SyncExplorerItem =
  | ListStoreSyncEvents200ItemsItem
  | ListStorePayments200ItemsItem
  | ListStoreCashMovements200ItemsItem
  | ListStoreFiscalDocuments200ItemsItem
  | ListStoreReturns200ItemsItem
  | ListStoreOperationalDays200ItemsItem;

function entityIdLink(storeId: string, tab: SyncEntityType, entityId: string) {
  return <Link to={syncEntityHref(storeId, tab, entityId)}>{entityId}</Link>;
}

function renderTableBody(activeTab: SyncTab, storeId: string, items: SyncExplorerItem[]) {
  switch (activeTab) {
    case 'sync-events':
      return (items as ListStoreSyncEvents200ItemsItem[]).map((row) => (
        <tr key={row.eventId}>
          <td>{row.sourceEventId}</td>
          <td>{row.eventType}</td>
          <td>{formatTimestamp(row.occurredAt)}</td>
          <td>{formatTimestamp(row.receivedAt)}</td>
        </tr>
      ));
    case 'payments':
      return (items as ListStorePayments200ItemsItem[]).map((row) => (
        <tr key={row.id}>
          <td>{entityIdLink(storeId, 'payments', row.id)}</td>
          <td>{row.method}</td>
          <td>{formatMinorAmount(row.amountMinor)}</td>
          <td>{row.status}</td>
          <td>{formatTimestamp(row.capturedAt)}</td>
        </tr>
      ));
    case 'cash-movements':
      return (items as ListStoreCashMovements200ItemsItem[]).map((row) => (
        <tr key={row.id}>
          <td>{entityIdLink(storeId, 'cash-movements', row.id)}</td>
          <td>{row.type}</td>
          <td>{formatMinorAmount(row.amountMinor)}</td>
          <td>{formatTimestamp(row.postedAt)}</td>
        </tr>
      ));
    case 'fiscal-documents':
      return (items as ListStoreFiscalDocuments200ItemsItem[]).map((row) => (
        <tr key={row.id}>
          <td>{entityIdLink(storeId, 'fiscal-documents', row.id)}</td>
          <td>{row.kind}</td>
          <td>{formatMinorAmount(row.amountMinor)}</td>
          <td>{formatTimestamp(row.fiscalizedAt)}</td>
        </tr>
      ));
    case 'returns':
      return (items as ListStoreReturns200ItemsItem[]).map((row) => (
        <tr key={row.id}>
          <td>{entityIdLink(storeId, 'returns', row.id)}</td>
          <td>{row.receiptId}</td>
          <td>{formatMinorAmount(row.totalMinor)}</td>
          <td>{formatTimestamp(row.settledAt)}</td>
        </tr>
      ));
    case 'operational-days':
      return (items as ListStoreOperationalDays200ItemsItem[]).map((row) => (
        <tr key={row.id}>
          <td>{entityIdLink(storeId, 'operational-days', row.id)}</td>
          <td>{row.businessDate}</td>
          <td>{formatTimestamp(row.closedAt)}</td>
          <td>{row.closedById}</td>
        </tr>
      ));
  }
}

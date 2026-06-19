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

import { getApiErrorMessage } from '../auth/AuthProvider.js';
import { formatMinorAmount, formatTimestamp, PAGE_SIZE } from './reporting-utils.js';

type SyncTab =
  | 'sync-events'
  | 'payments'
  | 'cash-movements'
  | 'fiscal-documents'
  | 'returns'
  | 'operational-days';

const SYNC_TABS: { id: SyncTab; label: string }[] = [
  { id: 'sync-events', label: 'Sync events' },
  { id: 'payments', label: 'Payments' },
  { id: 'cash-movements', label: 'Cash movements' },
  { id: 'fiscal-documents', label: 'Fiscal documents' },
  { id: 'returns', label: 'Returns' },
  { id: 'operational-days', label: 'Operational days' },
];

export function CentralSyncExplorerPage() {
  const storesQuery = useListStores();
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data.stores : [];
  const [selectedStoreId, setSelectedStoreId] = useState<string | null>(null);
  const activeStoreId = selectedStoreId ?? stores[0]?.id ?? '';
  const [activeTab, setActiveTab] = useState<SyncTab>('sync-events');
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
            <h2>Central Sync Explorer</h2>
            <p className="muted">Synchronized read models projected from Store Edge events.</p>
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
            onChange={(event) => {
              setSelectedStoreId(event.target.value);
              setOffset(0);
            }}
          >
            {stores.length === 0 ? <option value="">No stores registered</option> : null}
            {stores.map((store) => (
              <option key={store.id} value={store.id}>
                {store.name} ({store.id})
              </option>
            ))}
          </select>
        </label>

        <div className="filters" role="tablist">
          {SYNC_TABS.map((tab) => (
            <button
              key={tab.id}
              className={activeTab === tab.id ? undefined : 'secondary'}
              onClick={() => {
                setActiveTab(tab.id);
                setOffset(0);
              }}
              role="tab"
              type="button"
            >
              {tab.label}
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
          <p className="muted">Select a store to view synchronized read models.</p>
        </div>
      ) : (
        <div className="panel">
          <div className="panel-heading">
            <h3>{SYNC_TABS.find((tab) => tab.id === activeTab)?.label}</h3>
            <p className="muted">
              {totalCount === 0
                ? 'No items'
                : `Showing ${pageStart}–${pageEnd} of ${totalCount}`}
            </p>
          </div>

          {activeQuery.isLoading && !pageData ? (
            <p className="muted">Loading…</p>
          ) : pageData && pageData.items.length > 0 ? (
            <>
              <div className="table-wrap">
                <table>
                  <thead>{renderTableHead(activeTab)}</thead>
                  <tbody>{renderTableBody(activeTab, pageData.items as SyncExplorerItem[])}</tbody>
                </table>
              </div>
              <div className="pagination">
                <button
                  className="secondary"
                  disabled={!canGoPrev || isLoading}
                  onClick={() => setOffset((current) => Math.max(0, current - PAGE_SIZE))}
                  type="button"
                >
                  Previous
                </button>
                <button
                  className="secondary"
                  disabled={!canGoNext || isLoading}
                  onClick={() => setOffset((current) => current + PAGE_SIZE)}
                  type="button"
                >
                  Next
                </button>
              </div>
            </>
          ) : (
            <p className="muted">No items for this projection.</p>
          )}
        </div>
      )}
    </section>
  );
}

function renderTableHead(activeTab: SyncTab) {
  switch (activeTab) {
    case 'sync-events':
      return (
        <tr>
          <th>Source event ID</th>
          <th>Event type</th>
          <th>Occurred</th>
          <th>Received</th>
        </tr>
      );
    case 'payments':
      return (
        <tr>
          <th>Payment ID</th>
          <th>Method</th>
          <th>Amount</th>
          <th>Status</th>
          <th>Captured</th>
        </tr>
      );
    case 'cash-movements':
      return (
        <tr>
          <th>Movement ID</th>
          <th>Type</th>
          <th>Amount</th>
          <th>Posted</th>
        </tr>
      );
    case 'fiscal-documents':
      return (
        <tr>
          <th>Document ID</th>
          <th>Kind</th>
          <th>Amount</th>
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
          <th>Business date</th>
          <th>Closed</th>
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

function renderTableBody(activeTab: SyncTab, items: SyncExplorerItem[]) {
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
          <td>{row.id}</td>
          <td>{row.method}</td>
          <td>{formatMinorAmount(row.amountMinor)}</td>
          <td>{row.status}</td>
          <td>{formatTimestamp(row.capturedAt)}</td>
        </tr>
      ));
    case 'cash-movements':
      return (items as ListStoreCashMovements200ItemsItem[]).map((row) => (
        <tr key={row.id}>
          <td>{row.id}</td>
          <td>{row.type}</td>
          <td>{formatMinorAmount(row.amountMinor)}</td>
          <td>{formatTimestamp(row.postedAt)}</td>
        </tr>
      ));
    case 'fiscal-documents':
      return (items as ListStoreFiscalDocuments200ItemsItem[]).map((row) => (
        <tr key={row.id}>
          <td>{row.id}</td>
          <td>{row.kind}</td>
          <td>{formatMinorAmount(row.amountMinor)}</td>
          <td>{formatTimestamp(row.fiscalizedAt)}</td>
        </tr>
      ));
    case 'returns':
      return (items as ListStoreReturns200ItemsItem[]).map((row) => (
        <tr key={row.id}>
          <td>{row.id}</td>
          <td>{row.receiptId}</td>
          <td>{formatMinorAmount(row.totalMinor)}</td>
          <td>{formatTimestamp(row.settledAt)}</td>
        </tr>
      ));
    case 'operational-days':
      return (items as ListStoreOperationalDays200ItemsItem[]).map((row) => (
        <tr key={row.id}>
          <td>{row.id}</td>
          <td>{row.businessDate}</td>
          <td>{formatTimestamp(row.closedAt)}</td>
          <td>{row.closedById}</td>
        </tr>
      ));
  }
}

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
import { Link, useLocation, useNavigate, useSearchParams } from 'react-router-dom';

import { PaginationControls } from '@/components/PaginationControls.js';
import { StorePicker } from '@/components/StorePicker.js';
import { CentralReceiptDetailModal } from '@/components/sync/CentralReceiptDetailModal.js';
import { SyncEventDetailModal } from '@/components/sync/SyncEventDetailModal.js';
import { getApiErrorMessage } from '@/auth/api-errors.js';
import { formatMinorAmount, formatTimestamp, PAGE_SIZE } from './reporting-utils.js';
import {
  parseSyncTab,
  readEventFromSearchParams,
  syncEntityHref,
  syncExplorerHref,
  type SyncEntityType,
  type SyncTab,
} from './sync-routes.js';

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

type SyncExplorerTableActions = {
  onOpenSyncEvent: (event: ListStoreSyncEvents200ItemsItem) => void;
  onOpenReceipt: (receiptId: string) => void;
};

export function CentralSyncExplorerPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams, setSearchParams] = useSearchParams();
  const eventDeepLinkId = readEventFromSearchParams(searchParams);
  const initialTab = eventDeepLinkId
    ? 'sync-events'
    : (parseSyncTab(searchParams.get('tab')) ?? 'sync-events');
  const initialStoreId = searchParams.get('store');

  const storesQuery = useListStores();
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data.stores : [];
  const [selectedStoreId, setSelectedStoreId] = useState<string | null>(initialStoreId);
  const activeStoreId = selectedStoreId ?? stores[0]?.id ?? '';
  const [activeTab, setActiveTab] = useState<SyncTab>(initialTab);
  const [offset, setOffset] = useState(0);
  const [detailSyncEvent, setDetailSyncEvent] = useState<ListStoreSyncEvents200ItemsItem | null>(
    null,
  );
  const [detailReceiptId, setDetailReceiptId] = useState<string | null>(null);
  const [dismissedDeepLinkLocationKey, setDismissedDeepLinkLocationKey] = useState<string | null>(
    null,
  );

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
  const syncEventsPage = syncEventsQuery.data?.status === 200 ? syncEventsQuery.data.data : null;
  const totalCount = pageData?.totalCount ?? 0;
  const pageStart = totalCount === 0 ? 0 : offset + 1;
  const pageEnd = Math.min(offset + PAGE_SIZE, totalCount);
  const canGoPrev = offset > 0;
  const canGoNext = offset + PAGE_SIZE < totalCount;

  const deepLinkEventOnPage = useMemo(() => {
    if (dismissedDeepLinkLocationKey === location.key || !eventDeepLinkId || !syncEventsPage) {
      return null;
    }

    return syncEventsPage.items.find((event) => event.eventId === eventDeepLinkId) ?? null;
  }, [dismissedDeepLinkLocationKey, eventDeepLinkId, location.key, syncEventsPage]);

  const activeDetailSyncEvent = detailSyncEvent ?? deepLinkEventOnPage;

  const showEventNotOnPageNotice =
    eventDeepLinkId != null &&
    dismissedDeepLinkLocationKey !== location.key &&
    activeTab === 'sync-events' &&
    syncEventsPage != null &&
    !syncEventsPage.items.some((event) => event.eventId === eventDeepLinkId);

  const isLoading = storesQuery.isFetching || activeQuery.isFetching;
  const errorMessage =
    storesQuery.error != null
      ? getApiErrorMessage(storesQuery.error)
      : activeQuery.error != null
        ? getApiErrorMessage(activeQuery.error)
        : null;

  function updateExplorerSearchParams(options: {
    tab: SyncTab;
    storeId: string;
    eventId?: string | null;
  }) {
    const params = new URLSearchParams();
    params.set('tab', options.tab);
    if (options.storeId.length > 0) {
      params.set('store', options.storeId);
    }
    if (options.eventId) {
      params.set('event', options.eventId);
    }
    setSearchParams(params, { replace: true });
  }

  function handleStoreChange(storeId: string) {
    setSelectedStoreId(storeId);
    setOffset(0);
    updateExplorerSearchParams({ tab: activeTab, storeId, eventId: null });
  }

  function handleTabChange(tab: SyncTab) {
    setActiveTab(tab);
    setOffset(0);
    updateExplorerSearchParams({ tab, storeId: activeStoreId, eventId: null });
  }

  function handleOpenSyncEvent(event: ListStoreSyncEvents200ItemsItem) {
    setDetailSyncEvent(event);
    updateExplorerSearchParams({
      tab: 'sync-events',
      storeId: activeStoreId,
      eventId: event.eventId,
    });
  }

  function handleCloseSyncEvent() {
    setDetailSyncEvent(null);
    if (eventDeepLinkId) {
      setDismissedDeepLinkLocationKey(location.key);
      void navigate(syncExplorerHref({ tab: activeTab, storeId: activeStoreId || undefined }), {
        replace: true,
      });
    }
  }

  function refetchAll() {
    void storesQuery.refetch();
    if (queryEnabled) {
      void activeQuery.refetch();
    }
  }

  const tableActions: SyncExplorerTableActions = {
    onOpenSyncEvent: handleOpenSyncEvent,
    onOpenReceipt: setDetailReceiptId,
  };

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
          onChange={handleStoreChange}
        />

        <div className="filters" role="tablist" aria-label={t('sync.title')}>
          {SYNC_TABS.map((tab) => (
            <button
              key={tab.id}
              id={`sync-tab-${tab.id}`}
              className={activeTab === tab.id ? undefined : 'secondary'}
              onClick={() => handleTabChange(tab.id)}
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

      {showEventNotOnPageNotice ? (
        <div className="panel">
          <p className="muted">{t('sync.eventDetail.notOnPage')}</p>
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
                      tableActions,
                      t,
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

      {activeDetailSyncEvent ? (
        <SyncEventDetailModal event={activeDetailSyncEvent} onClose={handleCloseSyncEvent} />
      ) : null}
      {detailReceiptId ? (
        <CentralReceiptDetailModal
          receiptId={detailReceiptId}
          storeId={activeStoreId}
          onClose={() => setDetailReceiptId(null)}
        />
      ) : null}
    </section>
  );
}

function renderTableHead(activeTab: SyncTab, t: TFunction) {
  switch (activeTab) {
    case 'sync-events':
      return (
        <tr>
          <th>{t('sync.columns.sourceEventId')}</th>
          <th>{t('sync.columns.eventType')}</th>
          <th>{t('sync.columns.occurredAt')}</th>
          <th>{t('monitoring.eventReceived')}</th>
        </tr>
      );
    case 'payments':
      return (
        <tr>
          <th>{t('sync.columns.paymentId')}</th>
          <th>{t('sync.columns.method')}</th>
          <th>{t('safe.amount')}</th>
          <th>{t('monitoring.status')}</th>
          <th>{t('sync.columns.capturedAt')}</th>
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
          <th>{t('sync.columns.documentId')}</th>
          <th>{t('monitoring.kind')}</th>
          <th>{t('safe.amount')}</th>
          <th>{t('sync.columns.fiscalizedAt')}</th>
        </tr>
      );
    case 'returns':
      return (
        <tr>
          <th>{t('sync.columns.returnId')}</th>
          <th>{t('sync.columns.receiptId')}</th>
          <th>{t('sync.columns.total')}</th>
          <th>{t('sync.columns.settledAt')}</th>
        </tr>
      );
    case 'operational-days':
      return (
        <tr>
          <th>{t('sync.columns.dayId')}</th>
          <th>{t('eod.businessDate')}</th>
          <th>{t('eod.closedAt')}</th>
          <th>{t('sync.columns.closedBy')}</th>
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

function syncEventLinkButton(
  label: string,
  event: ListStoreSyncEvents200ItemsItem,
  onOpenSyncEvent: (event: ListStoreSyncEvents200ItemsItem) => void,
  ariaLabel: string,
) {
  return (
    <button
      className="link-button"
      type="button"
      onClick={() => onOpenSyncEvent(event)}
      aria-label={ariaLabel}
    >
      {label}
    </button>
  );
}

function receiptLinkButton(
  receiptId: string,
  onOpenReceipt: (receiptId: string) => void,
  ariaLabel: string,
) {
  return (
    <button
      className="link-button"
      type="button"
      onClick={() => onOpenReceipt(receiptId)}
      aria-label={ariaLabel}
    >
      {receiptId}
    </button>
  );
}

function renderTableBody(
  activeTab: SyncTab,
  storeId: string,
  items: SyncExplorerItem[],
  actions: SyncExplorerTableActions,
  t: TFunction,
) {
  switch (activeTab) {
    case 'sync-events':
      return (items as ListStoreSyncEvents200ItemsItem[]).map((row) => (
        <tr key={row.eventId}>
          <td>
            {syncEventLinkButton(
              row.sourceEventId,
              row,
              actions.onOpenSyncEvent,
              t('sync.eventDetail.openDetails', { eventId: row.eventId }),
            )}
          </td>
          <td>
            {syncEventLinkButton(
              row.eventType,
              row,
              actions.onOpenSyncEvent,
              t('sync.eventDetail.openDetails', { eventId: row.eventId }),
            )}
          </td>
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
          <td>
            {row.receiptId.length > 0
              ? receiptLinkButton(
                  row.receiptId,
                  actions.onOpenReceipt,
                  t('monitoring.openReceiptDetails', { receiptId: row.receiptId }),
                )
              : t('common.emDash')}
          </td>
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

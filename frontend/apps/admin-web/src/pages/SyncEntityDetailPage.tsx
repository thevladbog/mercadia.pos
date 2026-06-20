import {
  useGetStoreCashMovement,
  useGetStoreFiscalDocument,
  useGetStoreOperationalDay,
  useGetStorePayment,
  useGetStoreReturn,
  useListStores,
} from '@mercadia/api-clients-central';
import { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useLocation, useParams } from 'react-router-dom';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { ReceiptDetailModal } from '@/components/eod/ReceiptDetailModal.js';
import { PageBackLink } from './users-shared.js';
import {
  fieldsFromSyncEntityRecord,
  renderSyncEntityFieldValue,
} from './sync-entity-detail-utils.js';
import {
  entityTypeFromPathname,
  SYNC_ENTITY_PARAM,
  syncExplorerHref,
  type SyncEntityType,
} from './sync-routes.js';

const SYNC_ENTITY_LABEL_KEY: Record<SyncEntityType, string> = {
  payments: 'sync.tabs.payments',
  'cash-movements': 'sync.tabs.cashMovements',
  'fiscal-documents': 'sync.tabs.fiscalDocuments',
  returns: 'sync.tabs.returns',
  'operational-days': 'sync.tabs.operationalDays',
};

function useEntityDetailQuery(
  entityType: SyncEntityType | null,
  storeId: string,
  entityId: string,
) {
  const enabled = entityType != null && storeId.length > 0 && entityId.length > 0;

  const paymentQuery = useGetStorePayment(storeId, entityId, {
    query: { enabled: enabled && entityType === 'payments' },
  });
  const cashMovementQuery = useGetStoreCashMovement(storeId, entityId, {
    query: { enabled: enabled && entityType === 'cash-movements' },
  });
  const fiscalDocumentQuery = useGetStoreFiscalDocument(storeId, entityId, {
    query: { enabled: enabled && entityType === 'fiscal-documents' },
  });
  const returnQuery = useGetStoreReturn(storeId, entityId, {
    query: { enabled: enabled && entityType === 'returns' },
  });
  const operationalDayQuery = useGetStoreOperationalDay(storeId, entityId, {
    query: { enabled: enabled && entityType === 'operational-days' },
  });

  return useMemo(() => {
    switch (entityType) {
      case 'payments':
        return paymentQuery;
      case 'cash-movements':
        return cashMovementQuery;
      case 'fiscal-documents':
        return fiscalDocumentQuery;
      case 'returns':
        return returnQuery;
      case 'operational-days':
        return operationalDayQuery;
      default:
        return paymentQuery;
    }
  }, [
    entityType,
    paymentQuery,
    cashMovementQuery,
    fiscalDocumentQuery,
    returnQuery,
    operationalDayQuery,
  ]);
}

export function SyncEntityDetailPage() {
  const { t } = useTranslation();
  const params = useParams();
  const location = useLocation();
  const entityType = entityTypeFromPathname(location.pathname);
  const entityIdParam = entityType ? SYNC_ENTITY_PARAM[entityType] : 'paymentId';
  const storeId = params.storeId ?? '';
  const entityId = params[entityIdParam] ?? '';

  const [detailReceiptId, setDetailReceiptId] = useState<string | null>(null);

  const storesQuery = useListStores();
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data.stores : [];
  const storeName = stores.find((store) => store.id === storeId)?.name;

  const detailQuery = useEntityDetailQuery(entityType, storeId, entityId);
  const detail = detailQuery.data?.status === 200 ? detailQuery.data.data : null;
  const isLoading = detailQuery.isFetching;
  const errorMessage = detailQuery.error != null ? getApiErrorMessage(detailQuery.error) : null;

  const entityLabel = entityType ? t(SYNC_ENTITY_LABEL_KEY[entityType]) : t('sync.details');
  const title = storeName ? `${storeName} (${storeId})` : storeId;
  const backHref = syncExplorerHref({
    tab: entityType ?? undefined,
    storeId: storeId.length > 0 ? storeId : undefined,
  });

  const fields = detail ? fieldsFromSyncEntityRecord(detail as Record<string, unknown>) : [];
  const fieldHandlers = useMemo(
    () => ({
      storeId,
      onOpenReceipt: setDetailReceiptId,
    }),
    [storeId],
  );

  return (
    <section className="stack reporting-page">
      <PageBackLink label={t('sync.backToSync')} to={backHref} />

      <div className="panel">
        <div className="panel-heading">
          <div>
            <h2>{entityLabel}</h2>
            <p className="muted">
              {entityId || t('common.noData')} {t('common.emDash')}{' '}
              {title || t('reporting.unknownStore')}
            </p>
          </div>
          <button
            className="secondary"
            disabled={isLoading || storeId.length === 0 || entityId.length === 0}
            onClick={() => void detailQuery.refetch()}
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
        <h3>{t('sync.details')}</h3>
        {detailQuery.isLoading && !detail ? (
          <p className="muted">{t('sync.loadingDetail')}</p>
        ) : detail ? (
          <dl className="kpi-grid">
            {fields.map((field) => (
              <div key={field.key}>
                <dt>{field.label}</dt>
                <dd>{renderSyncEntityFieldValue(field.key, field.value, fieldHandlers)}</dd>
              </div>
            ))}
          </dl>
        ) : (
          <p className="muted">{t('sync.noDetail')}</p>
        )}
      </div>

      {detailReceiptId ? (
        <ReceiptDetailModal receiptId={detailReceiptId} onClose={() => setDetailReceiptId(null)} />
      ) : null}
    </section>
  );
}

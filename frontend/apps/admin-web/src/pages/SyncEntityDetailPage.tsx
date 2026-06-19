import {
  useGetStoreCashMovement,
  useGetStoreFiscalDocument,
  useGetStoreOperationalDay,
  useGetStorePayment,
  useGetStoreReturn,
  useListStores,
} from '@mercadia/api-clients-central';
import { useMemo } from 'react';
import { useLocation, useParams } from 'react-router-dom';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { PageBackLink } from './users-shared.js';
import { formatMinorAmount, formatTimestamp } from './reporting-utils.js';
import {
  entityTypeFromPathname,
  SYNC_ENTITY_LABEL,
  SYNC_ENTITY_PARAM,
  syncExplorerHref,
  type SyncEntityType,
} from './sync-routes.js';

type DetailField = {
  label: string;
  value: string;
};

function formatFieldValue(key: string, value: string | number | string[] | undefined): string {
  if (value == null || value === '') {
    return '—';
  }
  if (Array.isArray(value)) {
    return value.length > 0 ? value.join(', ') : '—';
  }
  if (typeof value === 'number' && key.toLowerCase().includes('minor')) {
    return formatMinorAmount(value);
  }
  if (
    typeof value === 'string' &&
    (key.endsWith('At') || key === 'updatedAt' || key === 'syncedAt')
  ) {
    return formatTimestamp(value);
  }
  return String(value);
}

function fieldsFromRecord(data: Record<string, unknown>): DetailField[] {
  return Object.entries(data).map(([key, value]) => ({
    label: key.replace(/([A-Z])/g, ' $1').replace(/^./, (char) => char.toUpperCase()),
    value: formatFieldValue(key, value as string | number | string[] | undefined),
  }));
}

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
  const params = useParams();
  const location = useLocation();
  const entityType = entityTypeFromPathname(location.pathname);
  const entityIdParam = entityType ? SYNC_ENTITY_PARAM[entityType] : 'paymentId';
  const storeId = params.storeId ?? '';
  const entityId = params[entityIdParam] ?? '';

  const storesQuery = useListStores();
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data.stores : [];
  const storeName = stores.find((store) => store.id === storeId)?.name;

  const detailQuery = useEntityDetailQuery(entityType, storeId, entityId);
  const detail = detailQuery.data?.status === 200 ? detailQuery.data.data : null;
  const isLoading = detailQuery.isFetching;
  const errorMessage = detailQuery.error != null ? getApiErrorMessage(detailQuery.error) : null;

  const entityLabel = entityType ? SYNC_ENTITY_LABEL[entityType] : 'Entity';
  const title = storeName ? `${storeName} (${storeId})` : storeId;
  const backHref = syncExplorerHref({
    tab: entityType ?? undefined,
    storeId: storeId.length > 0 ? storeId : undefined,
  });

  const fields = detail ? fieldsFromRecord(detail as Record<string, unknown>) : [];

  return (
    <section className="stack reporting-page">
      <PageBackLink label="Back to sync explorer" to={backHref} />

      <div className="panel">
        <div className="panel-heading">
          <div>
            <h2>{entityLabel}</h2>
            <p className="muted">
              {entityId || 'Unknown entity'} — {title || 'Unknown store'}
            </p>
          </div>
          <button
            className="secondary"
            disabled={isLoading || storeId.length === 0 || entityId.length === 0}
            onClick={() => void detailQuery.refetch()}
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
        <h3>Details</h3>
        {detailQuery.isLoading && !detail ? (
          <p className="muted">Loading…</p>
        ) : detail ? (
          <dl className="kpi-grid">
            {fields.map((field) => (
              <div key={field.label}>
                <dt>{field.label}</dt>
                <dd>{field.value}</dd>
              </div>
            ))}
          </dl>
        ) : (
          <p className="muted">No detail data.</p>
        )}
      </div>
    </section>
  );
}

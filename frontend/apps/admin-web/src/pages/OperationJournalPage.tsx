import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useSearchParams } from 'react-router-dom';

import { PaginationControls } from '@/components/PaginationControls.js';
import { useListOperationJournal } from '@mercadia/api-clients-store-edge';
import { useListStores } from '@mercadia/api-clients-central';

import { StorePicker } from '@/components/StorePicker.js';
import { readStoreFromSearchParams } from '@/pages/store-routes.js';
import { formatTimestamp } from '@/pages/reporting-utils.js';

export function OperationJournalPage() {
  const { t } = useTranslation();
  const [searchParams] = useSearchParams();
  const initialStoreId = readStoreFromSearchParams(searchParams);
  const [selectedStoreId, setSelectedStoreId] = useState<string | null>(initialStoreId);
  const [page, setPage] = useState(0);
  const pageSize = 25;

  const storesQuery = useListStores();
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data.stores : [];
  const activeStoreId = selectedStoreId ?? stores[0]?.id ?? '';

  const journalQuery = useListOperationJournal(
    activeStoreId,
    { limit: pageSize, offset: page * pageSize },
    { query: { enabled: activeStoreId.length > 0 } },
  );

  const journal = journalQuery.data?.status === 200 ? journalQuery.data.data : null;
  const items = journal?.items ?? [];
  const totalCount = journal?.totalCount ?? 0;
  const totalPages = Math.max(1, Math.ceil(totalCount / pageSize));

  if (!activeStoreId) {
    return <div className="panel"><h1>{t('seniorCashier.operationJournal')}</h1><p className="muted">{t('common.selectStore')}</p></div>;
  }

  return (
    <div className="panel">
      <h1>{t('seniorCashier.operationJournal')}</h1>
      <StorePicker stores={stores} value={activeStoreId} onChange={setSelectedStoreId} />
      {journalQuery.isPending ? (
        <p className="muted">{t('common.loading')}</p>
      ) : items.length === 0 ? (
        <p className="muted">{t('seniorCashier.noOperations')}</p>
      ) : (
        <>
          <table className="data-table">
            <thead>
              <tr>
                <th>{t('seniorCashier.journalDate')}</th>
                <th>{t('seniorCashier.journalType')}</th>
                <th>{t('seniorCashier.journalActor')}</th>
                <th>{t('seniorCashier.journalReference')}</th>
                <th>{t('seniorCashier.journalSummary')}</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <tr key={item.id}>
                  <td>{formatTimestamp(item.createdAt)}</td>
                  <td>{item.operationType}</td>
                  <td>{item.actorId}</td>
                  <td>{item.referenceId ?? '—'}</td>
                  <td>{item.summary ?? '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
          <PaginationControls
            canGoPrev={page > 0}
            canGoNext={page < totalPages - 1}
            onPrev={() => setPage((p) => Math.max(0, p - 1))}
            onNext={() => setPage((p) => p + 1)}
          />
        </>
      )}
    </div>
  );
}

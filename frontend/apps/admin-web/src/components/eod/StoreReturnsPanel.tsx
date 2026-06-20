import { useListStoreReturns } from '@mercadia/api-clients-store-edge';
import { Button } from '@mercadia/ui';
import { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { PaginationControls } from '@/components/PaginationControls.js';
import { formatMinorAmount, formatTimestamp, PAGE_SIZE } from '@/pages/reporting-utils.js';
import { STORE_POLL_INTERVAL_MS } from '@/pages/store-polling.js';

type StoreReturnsPanelProps = {
  storeId: string;
  businessDate: string;
  onOpenReturn: (returnId: string) => void;
};

function isReturnOnBusinessDate(createdAt: string, businessDate: string): boolean {
  return createdAt.slice(0, 10) === businessDate;
}

export function StoreReturnsPanel({ storeId, businessDate, onOpenReturn }: StoreReturnsPanelProps) {
  const { t } = useTranslation();
  const [offset, setOffset] = useState(0);

  const returnsQuery = useListStoreReturns(
    storeId,
    { limit: PAGE_SIZE, offset },
    {
      query: {
        enabled: storeId.length > 0,
        refetchInterval: STORE_POLL_INTERVAL_MS,
      },
    },
  );
  const page = returnsQuery.data?.status === 200 ? returnsQuery.data.data : null;
  const errorMessage = returnsQuery.error != null ? getApiErrorMessage(returnsQuery.error) : null;
  const emDash = t('common.emDash');

  const dayReturns = useMemo(() => {
    return (page?.items ?? []).filter((item) =>
      isReturnOnBusinessDate(item.createdAt, businessDate),
    );
  }, [businessDate, page?.items]);

  return (
    <div className="panel">
      <h3>{t('eod.tabs.returns')}</h3>
      <p className="muted">{t('eod.returns.dayHint', { businessDate })}</p>
      {errorMessage ? <p className="error">{errorMessage}</p> : null}
      {returnsQuery.isLoading && !page ? (
        <p className="muted">{t('eod.returns.loading')}</p>
      ) : dayReturns.length > 0 ? (
        <>
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>{t('eod.returnDetail.returnId')}</th>
                  <th>{t('eod.returnDetail.kind')}</th>
                  <th>{t('monitoring.status')}</th>
                  <th>{t('eod.returnDetail.receiptId')}</th>
                  <th>{t('eod.returnDetail.total')}</th>
                  <th>{t('eod.created')}</th>
                </tr>
              </thead>
              <tbody>
                {dayReturns.map((item) => (
                  <tr key={item.id}>
                    <td>
                      <Button
                        variant="link"
                        size="sm"
                        onClick={() => onOpenReturn(item.id)}
                        type="button"
                      >
                        {item.id}
                      </Button>
                    </td>
                    <td>{item.kind}</td>
                    <td>{item.status}</td>
                    <td>{item.receiptId ?? emDash}</td>
                    <td>{formatMinorAmount(item.totalMinor)}</td>
                    <td>{formatTimestamp(item.createdAt)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <PaginationControls
            canGoNext={offset + PAGE_SIZE < (page?.totalCount ?? 0)}
            canGoPrev={offset > 0}
            disabled={returnsQuery.isFetching}
            onNext={() => setOffset((value) => value + PAGE_SIZE)}
            onPrev={() => setOffset((value) => Math.max(0, value - PAGE_SIZE))}
          />
        </>
      ) : (
        <p className="muted">{t('eod.returns.empty')}</p>
      )}
    </div>
  );
}

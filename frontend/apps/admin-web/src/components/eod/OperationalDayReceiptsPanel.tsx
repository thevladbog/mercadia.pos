import { useListOperationalDayReceipts } from '@mercadia/api-clients-store-edge';
import { Button } from '@mercadia/ui';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { PaginationControls } from '@/components/PaginationControls.js';
import { formatMinorAmount, formatTimestamp, PAGE_SIZE } from '@/pages/reporting-utils.js';
import { STORE_POLL_INTERVAL_MS } from '@/pages/store-polling.js';

type OperationalDayReceiptsPanelProps = {
  operationalDayId: string;
  onOpenReceipt: (receiptId: string) => void;
};

export function OperationalDayReceiptsPanel({
  operationalDayId,
  onOpenReceipt,
}: OperationalDayReceiptsPanelProps) {
  const { t } = useTranslation();
  const [offset, setOffset] = useState(0);

  const receiptsQuery = useListOperationalDayReceipts(
    operationalDayId,
    { limit: PAGE_SIZE, offset },
    {
      query: {
        enabled: operationalDayId.length > 0,
        refetchInterval: STORE_POLL_INTERVAL_MS,
      },
    },
  );
  const page = receiptsQuery.data?.status === 200 ? receiptsQuery.data.data : null;
  const errorMessage = receiptsQuery.error != null ? getApiErrorMessage(receiptsQuery.error) : null;
  const items = page?.items ?? [];

  return (
    <div className="panel">
      <h3>{t('eod.tabs.receipts')}</h3>
      {errorMessage ? <p className="error">{errorMessage}</p> : null}
      {receiptsQuery.isLoading && !page ? (
        <p className="muted">{t('eod.receipts.loading')}</p>
      ) : items.length > 0 ? (
        <>
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>{t('eod.receiptDetail.receiptId')}</th>
                  <th>{t('monitoring.status')}</th>
                  <th>{t('eod.terminalId')}</th>
                  <th>{t('monitoring.cashier')}</th>
                  <th>{t('eod.receiptDetail.total')}</th>
                  <th>{t('eod.created')}</th>
                </tr>
              </thead>
              <tbody>
                {items.map((receipt) => (
                  <tr key={receipt.id}>
                    <td>
                      <Button
                        variant="link"
                        size="sm"
                        onClick={() => onOpenReceipt(receipt.id)}
                        type="button"
                      >
                        {receipt.id}
                      </Button>
                    </td>
                    <td>{receipt.status}</td>
                    <td>{receipt.terminalId}</td>
                    <td>{receipt.cashierId}</td>
                    <td>{formatMinorAmount(receipt.totalMinor)}</td>
                    <td>{formatTimestamp(receipt.createdAt)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <PaginationControls
            canGoNext={offset + PAGE_SIZE < (page?.totalCount ?? 0)}
            canGoPrev={offset > 0}
            disabled={receiptsQuery.isFetching}
            onNext={() => setOffset((value) => value + PAGE_SIZE)}
            onPrev={() => setOffset((value) => Math.max(0, value - PAGE_SIZE))}
          />
        </>
      ) : (
        <p className="muted">{t('eod.receipts.empty')}</p>
      )}
    </div>
  );
}

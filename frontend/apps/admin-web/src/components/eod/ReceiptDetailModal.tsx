import { useGetReceipt } from '@mercadia/api-clients-store-edge';
import { useTranslation } from 'react-i18next';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { DetailDialog } from '@mercadia/ui';
import { formatMinorAmount, formatTimestamp } from '@/pages/reporting-utils.js';

type ReceiptDetailModalProps = {
  receiptId: string;
  onClose: () => void;
};

export function ReceiptDetailModal({ receiptId, onClose }: ReceiptDetailModalProps) {
  const { t } = useTranslation();
  const receiptQuery = useGetReceipt(receiptId, {
    query: { enabled: receiptId.length > 0 },
  });
  const receipt = receiptQuery.data?.status === 200 ? receiptQuery.data.data : null;
  const errorMessage = receiptQuery.error != null ? getApiErrorMessage(receiptQuery.error) : null;

  return (
    <DetailDialog
      open
      title={t('eod.receiptDetail.title')}
      cancelLabel={t('common.cancel')}
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
    >
      {receiptQuery.isLoading && !receipt ? (
        <p className="muted">{t('common.loading')}</p>
      ) : errorMessage ? (
        <p className="error">{errorMessage}</p>
      ) : receipt ? (
        <dl className="kpi-grid">
          <div>
            <dt>{t('eod.receiptDetail.receiptId')}</dt>
            <dd>{receipt.id}</dd>
          </div>
          <div>
            <dt>{t('monitoring.status')}</dt>
            <dd>{receipt.status}</dd>
          </div>
          <div>
            <dt>{t('eod.terminalId')}</dt>
            <dd>{receipt.terminalId}</dd>
          </div>
          <div>
            <dt>{t('monitoring.cashier')}</dt>
            <dd>{receipt.cashierId}</dd>
          </div>
          <div>
            <dt>{t('eod.receiptDetail.total')}</dt>
            <dd>{formatMinorAmount(receipt.totalMinor)}</dd>
          </div>
          <div>
            <dt>{t('eod.receiptDetail.lineCount')}</dt>
            <dd>{receipt.lines.length}</dd>
          </div>
          <div>
            <dt>{t('eod.created')}</dt>
            <dd>{formatTimestamp(receipt.createdAt)}</dd>
          </div>
        </dl>
      ) : (
        <p className="muted">{t('common.noData')}</p>
      )}
    </DetailDialog>
  );
}

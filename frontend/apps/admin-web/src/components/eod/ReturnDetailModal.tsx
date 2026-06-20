import { useGetReturn } from '@mercadia/api-clients-store-edge';
import { useTranslation } from 'react-i18next';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { DetailDialog } from '@mercadia/ui';
import { formatMinorAmount, formatTimestamp } from '@/pages/reporting-utils.js';

type ReturnDetailModalProps = {
  returnId: string;
  onClose: () => void;
  onOpenReceipt: (receiptId: string) => void;
};

export function ReturnDetailModal({ returnId, onClose, onOpenReceipt }: ReturnDetailModalProps) {
  const { t } = useTranslation();
  const returnQuery = useGetReturn(returnId, {
    query: { enabled: returnId.length > 0 },
  });
  const returnData = returnQuery.data?.status === 200 ? returnQuery.data.data.return : null;
  const errorMessage = returnQuery.error != null ? getApiErrorMessage(returnQuery.error) : null;
  const emDash = t('common.emDash');

  return (
    <DetailDialog
      open
      title={t('eod.returnDetail.title')}
      cancelLabel={t('common.cancel')}
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
    >
      {returnQuery.isLoading && !returnData ? (
        <p className="muted">{t('common.loading')}</p>
      ) : errorMessage ? (
        <p className="error">{errorMessage}</p>
      ) : returnData ? (
        <dl className="kpi-grid">
          <div>
            <dt>{t('eod.returnDetail.returnId')}</dt>
            <dd>{returnData.id}</dd>
          </div>
          <div>
            <dt>{t('common.store')}</dt>
            <dd>{returnData.storeId}</dd>
          </div>
          <div>
            <dt>{t('eod.returnDetail.kind')}</dt>
            <dd>{returnData.kind}</dd>
          </div>
          <div>
            <dt>{t('monitoring.status')}</dt>
            <dd>{returnData.status}</dd>
          </div>
          <div>
            <dt>{t('eod.returnDetail.receiptId')}</dt>
            <dd>
              {returnData.receiptId && returnData.receiptId.length > 0 ? (
                <button
                  className="link-button"
                  onClick={() => onOpenReceipt(returnData.receiptId!)}
                  type="button"
                >
                  {returnData.receiptId}
                </button>
              ) : (
                emDash
              )}
            </dd>
          </div>
          <div>
            <dt>{t('eod.returnDetail.total')}</dt>
            <dd>{formatMinorAmount(returnData.totalMinor)}</dd>
          </div>
          <div>
            <dt>{t('eod.returnDetail.lineCount')}</dt>
            <dd>{returnData.lines.length}</dd>
          </div>
          <div>
            <dt>{t('eod.returnDetail.reason')}</dt>
            <dd>{returnData.reason}</dd>
          </div>
          <div>
            <dt>{t('safe.actor')}</dt>
            <dd>{returnData.actorId}</dd>
          </div>
          <div>
            <dt>{t('safe.movementDetail.approvedBy')}</dt>
            <dd>
              {returnData.approvedById && returnData.approvedById.length > 0
                ? returnData.approvedById
                : emDash}
            </dd>
          </div>
          <div>
            <dt>{t('eod.created')}</dt>
            <dd>{formatTimestamp(returnData.createdAt)}</dd>
          </div>
        </dl>
      ) : (
        <p className="muted">{t('common.noData')}</p>
      )}
    </DetailDialog>
  );
}

import { useGetReturn } from '@mercadia/api-clients-store-edge';
import { useTranslation } from 'react-i18next';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { DetailDialog, Button } from '@mercadia/ui';
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
        <div className="stack">
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
                  <Button
                    variant="link"
                    size="sm"
                    onClick={() => onOpenReceipt(returnData.receiptId!)}
                    type="button"
                  >
                    {returnData.receiptId}
                  </Button>
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

          <div>
            <h4>{t('eod.returnDetail.linesSection')}</h4>
            {returnData.lines.length > 0 ? (
              <div className="table-wrap">
                <table>
                  <thead>
                    <tr>
                      <th>{t('eod.lines.product')}</th>
                      <th>{t('eod.lines.quantity')}</th>
                      <th>{t('eod.lines.unitPrice')}</th>
                      <th>{t('eod.lines.total')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {returnData.lines.map((line, index) => (
                      <tr key={line.lineId ?? `${line.productId ?? line.name}-${index}`}>
                        <td>{line.name}</td>
                        <td>{line.quantity}</td>
                        <td>{formatMinorAmount(line.unitPriceMinor)}</td>
                        <td>{formatMinorAmount(line.totalMinor)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <p className="muted">{t('eod.lines.empty')}</p>
            )}
          </div>
        </div>
      ) : (
        <p className="muted">{t('common.noData')}</p>
      )}
    </DetailDialog>
  );
}

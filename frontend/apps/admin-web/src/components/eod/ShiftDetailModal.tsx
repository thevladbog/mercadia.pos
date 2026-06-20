import { useGetShift, useListShiftReceipts } from '@mercadia/api-clients-store-edge';
import { useTranslation } from 'react-i18next';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { DetailDialog, Button } from '@mercadia/ui';
import type { EodTab } from '@/pages/eod-blocker-utils.js';
import { formatMinorAmount, formatTimestamp } from '@/pages/reporting-utils.js';

type ShiftDetailModalProps = {
  shiftId: string;
  canWrite: boolean;
  onClose: () => void;
  onEodTab: (tab: EodTab) => void;
  onOpenReceipt?: (receiptId: string) => void;
};

export function ShiftDetailModal({
  shiftId,
  canWrite,
  onClose,
  onEodTab,
  onOpenReceipt,
}: ShiftDetailModalProps) {
  const { t } = useTranslation();
  const shiftQuery = useGetShift(shiftId, {
    query: { enabled: shiftId.length > 0 },
  });
  const receiptsQuery = useListShiftReceipts(shiftId, {
    query: { enabled: shiftId.length > 0 },
  });
  const shift = shiftQuery.data?.status === 200 ? shiftQuery.data.data : null;
  const receipts = receiptsQuery.data?.status === 200 ? receiptsQuery.data.data.receipts : null;
  const errorMessage =
    shiftQuery.error != null
      ? getApiErrorMessage(shiftQuery.error)
      : receiptsQuery.error != null
        ? getApiErrorMessage(receiptsQuery.error)
        : null;

  function handleOpenShiftsTab() {
    onEodTab('open-shifts');
    onClose();
  }

  return (
    <DetailDialog
      open
      footer={
        canWrite ? (
          <Button type="button" onClick={handleOpenShiftsTab}>
            {t('eod.blockerActions.viewShift')}
          </Button>
        ) : undefined
      }
      title={t('eod.shiftDetail.title')}
      cancelLabel={t('common.cancel')}
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
    >
      {shiftQuery.isLoading && !shift ? (
        <p className="muted">{t('common.loading')}</p>
      ) : errorMessage ? (
        <p className="error">{errorMessage}</p>
      ) : shift ? (
        <div className="stack">
          <dl className="kpi-grid">
            <div>
              <dt>{t('eod.shiftId')}</dt>
              <dd>{shift.id}</dd>
            </div>
            <div>
              <dt>{t('monitoring.cashier')}</dt>
              <dd>{shift.cashierId}</dd>
            </div>
            <div>
              <dt>{t('eod.terminalId')}</dt>
              <dd>{shift.terminalId}</dd>
            </div>
            <div>
              <dt>{t('eod.shiftDetail.drawerId')}</dt>
              <dd>{shift.drawerId}</dd>
            </div>
            <div>
              <dt>{t('monitoring.status')}</dt>
              <dd>{shift.status}</dd>
            </div>
            <div>
              <dt>{t('eod.openingCash')}</dt>
              <dd>{formatMinorAmount(shift.openingCashMinor)}</dd>
            </div>
            <div>
              <dt>{t('eod.shiftDetail.closingCash')}</dt>
              <dd>{formatMinorAmount(shift.closingCashMinor)}</dd>
            </div>
            <div>
              <dt>{t('eod.opened')}</dt>
              <dd>{formatTimestamp(shift.openedAt)}</dd>
            </div>
            <div>
              <dt>{t('eod.closedAt')}</dt>
              <dd>{shift.closedAt ? formatTimestamp(shift.closedAt) : t('common.emDash')}</dd>
            </div>
          </dl>

          <div>
            <h4>{t('eod.shiftDetail.receiptsSection')}</h4>
            {receiptsQuery.isLoading && !receipts ? (
              <p className="muted">{t('common.loading')}</p>
            ) : receipts && receipts.length > 0 ? (
              <div className="table-wrap">
                <table>
                  <thead>
                    <tr>
                      <th>{t('eod.receiptDetail.receiptId')}</th>
                      <th>{t('monitoring.status')}</th>
                      <th>{t('eod.receiptDetail.total')}</th>
                      <th>{t('eod.created')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {receipts.map((receipt) => (
                      <tr key={receipt.id}>
                        <td>
                          {onOpenReceipt ? (
                            <Button
                              variant="link"
                              size="sm"
                              onClick={() => onOpenReceipt(receipt.id)}
                              type="button"
                            >
                              {receipt.id}
                            </Button>
                          ) : (
                            receipt.id
                          )}
                        </td>
                        <td>{receipt.status}</td>
                        <td>{formatMinorAmount(receipt.totalMinor)}</td>
                        <td>{formatTimestamp(receipt.createdAt)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <p className="muted">{t('eod.shiftDetail.noReceipts')}</p>
            )}
          </div>
        </div>
      ) : (
        <p className="muted">{t('common.noData')}</p>
      )}
    </DetailDialog>
  );
}

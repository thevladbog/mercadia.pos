import { useGetShift } from '@mercadia/api-clients-store-edge';
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
};

export function ShiftDetailModal({ shiftId, canWrite, onClose, onEodTab }: ShiftDetailModalProps) {
  const { t } = useTranslation();
  const shiftQuery = useGetShift(shiftId, {
    query: { enabled: shiftId.length > 0 },
  });
  const shift = shiftQuery.data?.status === 200 ? shiftQuery.data.data : null;
  const errorMessage = shiftQuery.error != null ? getApiErrorMessage(shiftQuery.error) : null;

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
      ) : (
        <p className="muted">{t('common.noData')}</p>
      )}
    </DetailDialog>
  );
}

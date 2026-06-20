import { useTranslation } from 'react-i18next';

import { formatMinorAmount } from '@/pages/reporting-utils.js';

type CurrentReceiptCellProps = {
  receiptId?: string;
  status?: string;
  totalMinor?: number;
  onOpenReceipt?: (receiptId: string) => void;
  variant: 'id' | 'status' | 'total' | 'tile';
};

export function CurrentReceiptCell({
  receiptId,
  status,
  totalMinor,
  onOpenReceipt,
  variant,
}: CurrentReceiptCellProps) {
  const { t } = useTranslation();

  if (variant === 'id') {
    if (!receiptId) {
      return <span>{t('common.emDash')}</span>;
    }
    if (!onOpenReceipt) {
      return <span>{receiptId}</span>;
    }
    return (
      <button
        className="link-button"
        type="button"
        onClick={() => onOpenReceipt(receiptId)}
        aria-label={t('monitoring.openReceiptDetails', { receiptId })}
      >
        {receiptId}
      </button>
    );
  }

  if (variant === 'status') {
    return <span>{status ?? t('common.emDash')}</span>;
  }

  if (variant === 'total') {
    return <span>{totalMinor != null ? formatMinorAmount(totalMinor) : t('common.emDash')}</span>;
  }

  return (
    <>
      <div>
        <dt>{t('monitoring.currentReceipt')}</dt>
        <dd>
          <CurrentReceiptCell receiptId={receiptId} variant="id" onOpenReceipt={onOpenReceipt} />
        </dd>
      </div>
      <div>
        <dt>{t('monitoring.currentReceiptStatus')}</dt>
        <dd>
          <CurrentReceiptCell status={status} variant="status" />
        </dd>
      </div>
      <div>
        <dt>{t('monitoring.currentReceiptTotal')}</dt>
        <dd>
          <CurrentReceiptCell totalMinor={totalMinor} variant="total" />
        </dd>
      </div>
    </>
  );
}

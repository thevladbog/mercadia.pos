import { useGetReceipt, useListReceiptPayments } from '@mercadia/api-clients-store-edge';
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
  const paymentsQuery = useListReceiptPayments(receiptId, {
    query: { enabled: receiptId.length > 0 },
  });
  const receipt = receiptQuery.data?.status === 200 ? receiptQuery.data.data : null;
  const payments = paymentsQuery.data?.status === 200 ? paymentsQuery.data.data.payments : null;
  const errorMessage =
    receiptQuery.error != null
      ? getApiErrorMessage(receiptQuery.error)
      : paymentsQuery.error != null
        ? getApiErrorMessage(paymentsQuery.error)
        : null;
  const emDash = t('common.emDash');

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
        <div className="stack">
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
              <dt>{t('eod.created')}</dt>
              <dd>{formatTimestamp(receipt.createdAt)}</dd>
            </div>
          </dl>

          <div>
            <h4>{t('eod.receiptDetail.linesSection')}</h4>
            {receipt.lines.length > 0 ? (
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
                    {receipt.lines.map((line) => (
                      <tr key={line.id}>
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

          <div>
            <h4>{t('eod.receiptDetail.paymentsSection')}</h4>
            {paymentsQuery.isLoading && !payments ? (
              <p className="muted">{t('common.loading')}</p>
            ) : payments && payments.length > 0 ? (
              <div className="table-wrap">
                <table>
                  <thead>
                    <tr>
                      <th>{t('eod.kpi.paymentMethod')}</th>
                      <th>{t('monitoring.status')}</th>
                      <th>{t('eod.lines.total')}</th>
                      <th>{t('eod.payments.capturedAt')}</th>
                      <th>{t('eod.payments.refunded')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {payments.map((payment) => (
                      <tr key={payment.id}>
                        <td>{payment.method}</td>
                        <td>{payment.status}</td>
                        <td>{formatMinorAmount(payment.amountMinor)}</td>
                        <td>{payment.capturedAt ? formatTimestamp(payment.capturedAt) : emDash}</td>
                        <td>{formatMinorAmount(payment.refundedAmountMinor)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <p className="muted">{t('eod.payments.empty')}</p>
            )}
          </div>
        </div>
      ) : (
        <p className="muted">{t('common.noData')}</p>
      )}
    </DetailDialog>
  );
}

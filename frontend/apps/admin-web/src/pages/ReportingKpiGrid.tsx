import { useTranslation } from 'react-i18next';

import { formatMinorAmount } from './reporting-utils.js';

export type ReportingKpiData = {
  cashMovementsPostedCount: number;
  fiscalReceiptAmountMinor: number;
  fiscalReceiptCount: number;
  fiscalReturnAmountMinor: number;
  fiscalReturnCount: number;
  operationalDaysClosedCount: number;
  paymentsCancelledCount: number;
  paymentsCapturedAmountMinor: number;
  paymentsRefundedAmountMinor: number;
  returnsSettledAmountMinor: number;
  returnsSettledCount: number;
  storeCount?: number;
};

type ReportingKpiGridProps = {
  data: ReportingKpiData;
};

export function ReportingKpiGrid({ data }: ReportingKpiGridProps) {
  const { t } = useTranslation();

  return (
    <dl className="kpi-grid">
      {data.storeCount != null ? (
        <div>
          <dt>{t('reporting.stores')}</dt>
          <dd>{data.storeCount}</dd>
        </div>
      ) : null}
      <div>
        <dt>{t('reporting.fiscalReceipts')}</dt>
        <dd>
          {data.fiscalReceiptCount} / {formatMinorAmount(data.fiscalReceiptAmountMinor)}
        </dd>
      </div>
      <div>
        <dt>{t('reporting.fiscalReturns')}</dt>
        <dd>
          {data.fiscalReturnCount} / {formatMinorAmount(data.fiscalReturnAmountMinor)}
        </dd>
      </div>
      <div>
        <dt>{t('reporting.paymentsCaptured')}</dt>
        <dd>{formatMinorAmount(data.paymentsCapturedAmountMinor)}</dd>
      </div>
      <div>
        <dt>{t('reporting.paymentsCancelled')}</dt>
        <dd>{data.paymentsCancelledCount}</dd>
      </div>
      <div>
        <dt>{t('reporting.paymentsRefunded')}</dt>
        <dd>{formatMinorAmount(data.paymentsRefundedAmountMinor)}</dd>
      </div>
      <div>
        <dt>{t('reporting.returnsSettled')}</dt>
        <dd>
          {data.returnsSettledCount} / {formatMinorAmount(data.returnsSettledAmountMinor)}
        </dd>
      </div>
      <div>
        <dt>{t('reporting.cashMovementsPosted')}</dt>
        <dd>{data.cashMovementsPostedCount}</dd>
      </div>
      <div>
        <dt>{t('reporting.operationalDaysClosed')}</dt>
        <dd>{data.operationalDaysClosedCount}</dd>
      </div>
    </dl>
  );
}

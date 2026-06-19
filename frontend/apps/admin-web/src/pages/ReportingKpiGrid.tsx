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
  return (
    <dl className="kpi-grid">
      {data.storeCount != null ? (
        <div>
          <dt>Stores</dt>
          <dd>{data.storeCount}</dd>
        </div>
      ) : null}
      <div>
        <dt>Fiscal receipts</dt>
        <dd>
          {data.fiscalReceiptCount} / {formatMinorAmount(data.fiscalReceiptAmountMinor)}
        </dd>
      </div>
      <div>
        <dt>Fiscal returns</dt>
        <dd>
          {data.fiscalReturnCount} / {formatMinorAmount(data.fiscalReturnAmountMinor)}
        </dd>
      </div>
      <div>
        <dt>Payments captured</dt>
        <dd>{formatMinorAmount(data.paymentsCapturedAmountMinor)}</dd>
      </div>
      <div>
        <dt>Payments cancelled</dt>
        <dd>{data.paymentsCancelledCount}</dd>
      </div>
      <div>
        <dt>Payments refunded</dt>
        <dd>{formatMinorAmount(data.paymentsRefundedAmountMinor)}</dd>
      </div>
      <div>
        <dt>Returns settled</dt>
        <dd>
          {data.returnsSettledCount} / {formatMinorAmount(data.returnsSettledAmountMinor)}
        </dd>
      </div>
      <div>
        <dt>Cash movements posted</dt>
        <dd>{data.cashMovementsPostedCount}</dd>
      </div>
      <div>
        <dt>Operational days closed</dt>
        <dd>{data.operationalDaysClosedCount}</dd>
      </div>
    </dl>
  );
}

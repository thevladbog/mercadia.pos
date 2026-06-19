import {
  useGetCentralReportingSummary,
  useListCentralStoreReportingSummaries,
} from '@mercadia/api-clients-central';
import { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';

import { PaginationControls } from '@/components/PaginationControls.js';
import { getApiErrorMessage } from '@/auth/api-errors.js';
import { ReportingKpiGrid } from './ReportingKpiGrid.js';
import {
  defaultReportingWindow,
  formatMinorAmount,
  formatTimestamp,
  fromDatetimeLocalValue,
  PAGE_SIZE,
  toDatetimeLocalValue,
} from './reporting-utils.js';

function storeReportingHref(storeId: string, since: string, until: string): string {
  const params = new URLSearchParams({ since, until });
  return `/central/reporting/stores/${encodeURIComponent(storeId)}?${params.toString()}`;
}

export function CentralReportingPage() {
  const { t } = useTranslation();
  const defaults = useMemo(() => defaultReportingWindow(), []);
  const [sinceInput, setSinceInput] = useState(toDatetimeLocalValue(defaults.since));
  const [untilInput, setUntilInput] = useState(toDatetimeLocalValue(defaults.until));
  const [regionInput, setRegionInput] = useState('');
  const [offset, setOffset] = useState(0);
  const [applied, setApplied] = useState({
    since: defaults.since,
    until: defaults.until,
    region: undefined as string | undefined,
  });

  const queryParams = useMemo(
    () => ({
      since: applied.since,
      until: applied.until,
      ...(applied.region ? { region: applied.region } : {}),
    }),
    [applied],
  );

  const summaryQuery = useGetCentralReportingSummary(queryParams);
  const storesQuery = useListCentralStoreReportingSummaries({
    ...queryParams,
    limit: PAGE_SIZE,
    offset,
  });

  const summary = summaryQuery.data?.status === 200 ? summaryQuery.data.data : null;
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data : null;
  const isLoading = summaryQuery.isFetching || storesQuery.isFetching;
  const errorMessage =
    summaryQuery.error != null
      ? getApiErrorMessage(summaryQuery.error)
      : storesQuery.error != null
        ? getApiErrorMessage(storesQuery.error)
        : null;

  function applyFilters() {
    setOffset(0);
    setApplied({
      since: fromDatetimeLocalValue(sinceInput),
      until: fromDatetimeLocalValue(untilInput),
      region: regionInput.trim() || undefined,
    });
  }

  function refetchAll() {
    void summaryQuery.refetch();
    void storesQuery.refetch();
  }

  const totalCount = stores?.totalCount ?? 0;
  const pageStart = totalCount === 0 ? 0 : offset + 1;
  const pageEnd = Math.min(offset + PAGE_SIZE, totalCount);
  const canGoPrev = offset > 0;
  const canGoNext = offset + PAGE_SIZE < totalCount;

  return (
    <section className="stack reporting-page">
      <div className="panel">
        <div className="panel-heading">
          <div>
            <h2>{t('reporting.title')}</h2>
            <p className="muted">
              {t('reporting.subtitle')} {formatTimestamp(applied.since)} {t('common.emDash')}{' '}
              {formatTimestamp(applied.until)} UTC
            </p>
          </div>
          <button className="secondary" disabled={isLoading} onClick={refetchAll} type="button">
            {isLoading ? t('common.refreshing') : t('common.refresh')}
          </button>
        </div>

        <form
          className="filters"
          onSubmit={(event) => {
            event.preventDefault();
            applyFilters();
          }}
        >
          <label className="field">
            <span>{t('reporting.sinceUtc')}</span>
            <input
              required
              type="datetime-local"
              value={sinceInput}
              onChange={(event) => setSinceInput(event.target.value)}
            />
          </label>
          <label className="field">
            <span>{t('reporting.untilUtc')}</span>
            <input
              required
              type="datetime-local"
              value={untilInput}
              onChange={(event) => setUntilInput(event.target.value)}
            />
          </label>
          <label className="field">
            <span>{t('reporting.region')}</span>
            <input
              placeholder={t('reporting.regionPlaceholder')}
              type="text"
              value={regionInput}
              onChange={(event) => setRegionInput(event.target.value)}
            />
          </label>
          <button disabled={isLoading} type="submit">
            {t('common.apply')}
          </button>
        </form>
      </div>

      {errorMessage ? (
        <div className="panel error-panel">
          <p className="error">{errorMessage}</p>
        </div>
      ) : null}

      <div className="panel">
        <h3>{t('reporting.networkKpis')}</h3>
        {summaryQuery.isLoading && !summary ? (
          <p className="muted">{t('reporting.loadingSummary')}</p>
        ) : summary ? (
          <ReportingKpiGrid data={summary} />
        ) : (
          <p className="muted">{t('reporting.noSummary')}</p>
        )}
      </div>

      <div className="panel">
        <div className="panel-heading">
          <h3>{t('reporting.storeBreakdown')}</h3>
          <p className="muted">
            {totalCount === 0
              ? t('common.noItems')
              : t('common.showingRange', { from: pageStart, to: pageEnd, total: totalCount })}
          </p>
        </div>

        {storesQuery.isLoading && !stores ? (
          <p className="muted">{t('reporting.loadingSummary')}</p>
        ) : stores && stores.items.length > 0 ? (
          <>
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>{t('common.store')}</th>
                    <th>{t('monitoring.receipts')}</th>
                    <th>{t('reporting.fiscalReceipts')}</th>
                    <th>{t('reporting.returnsSettled')}</th>
                    <th>{t('reporting.paymentsCaptured')}</th>
                    <th>{t('reporting.cashMovementsPosted')}</th>
                    <th>{t('reporting.operationalDaysClosed')}</th>
                  </tr>
                </thead>
                <tbody>
                  {stores.items.map((item) => (
                    <tr key={item.storeId}>
                      <td>
                        <Link to={storeReportingHref(item.storeId, applied.since, applied.until)}>
                          {item.storeId}
                        </Link>
                      </td>
                      <td>{item.fiscalReceiptCount}</td>
                      <td>{formatMinorAmount(item.fiscalReceiptAmountMinor)}</td>
                      <td>
                        {item.returnsSettledCount} /{' '}
                        {formatMinorAmount(item.returnsSettledAmountMinor)}
                      </td>
                      <td>{formatMinorAmount(item.paymentsCapturedAmountMinor)}</td>
                      <td>{item.cashMovementsPostedCount}</td>
                      <td>{item.operationalDaysClosedCount}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <PaginationControls
              canGoNext={canGoNext}
              canGoPrev={canGoPrev}
              disabled={isLoading}
              onNext={() => setOffset((current) => current + PAGE_SIZE)}
              onPrev={() => setOffset((current) => Math.max(0, current - PAGE_SIZE))}
            />
          </>
        ) : (
          <p className="muted">{t('reporting.noSummary')}</p>
        )}
      </div>
    </section>
  );
}

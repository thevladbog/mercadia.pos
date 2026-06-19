import {
  useGetCentralReportingSummary,
  useListCentralStoreReportingSummaries,
} from '@mercadia/api-clients-central';
import { useMemo, useState } from 'react';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import {
  defaultReportingWindow,
  formatMinorAmount,
  formatTimestamp,
  fromDatetimeLocalValue,
  PAGE_SIZE,
  toDatetimeLocalValue,
} from './reporting-utils.js';

export function CentralReportingPage() {
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
            <h2>Central Reporting</h2>
            <p className="muted">
              Cross-store aggregates for {formatTimestamp(applied.since)} –{' '}
              {formatTimestamp(applied.until)} UTC
            </p>
          </div>
          <button className="secondary" disabled={isLoading} onClick={refetchAll} type="button">
            {isLoading ? 'Refreshing…' : 'Refresh'}
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
            <span>Since (UTC)</span>
            <input
              required
              type="datetime-local"
              value={sinceInput}
              onChange={(event) => setSinceInput(event.target.value)}
            />
          </label>
          <label className="field">
            <span>Until (UTC)</span>
            <input
              required
              type="datetime-local"
              value={untilInput}
              onChange={(event) => setUntilInput(event.target.value)}
            />
          </label>
          <label className="field">
            <span>Region (optional)</span>
            <input
              placeholder="e.g. moscow"
              type="text"
              value={regionInput}
              onChange={(event) => setRegionInput(event.target.value)}
            />
          </label>
          <button disabled={isLoading} type="submit">
            Apply
          </button>
        </form>
      </div>

      {errorMessage ? (
        <div className="panel error-panel">
          <p className="error">{errorMessage}</p>
        </div>
      ) : null}

      <div className="panel">
        <h3>Network KPIs</h3>
        {summaryQuery.isLoading && !summary ? (
          <p className="muted">Loading summary…</p>
        ) : summary ? (
          <dl className="kpi-grid">
            <div>
              <dt>Stores</dt>
              <dd>{summary.storeCount}</dd>
            </div>
            <div>
              <dt>Fiscal receipts</dt>
              <dd>
                {summary.fiscalReceiptCount} / {formatMinorAmount(summary.fiscalReceiptAmountMinor)}
              </dd>
            </div>
            <div>
              <dt>Fiscal returns</dt>
              <dd>
                {summary.fiscalReturnCount} / {formatMinorAmount(summary.fiscalReturnAmountMinor)}
              </dd>
            </div>
            <div>
              <dt>Payments captured</dt>
              <dd>{formatMinorAmount(summary.paymentsCapturedAmountMinor)}</dd>
            </div>
            <div>
              <dt>Payments cancelled</dt>
              <dd>{summary.paymentsCancelledCount}</dd>
            </div>
            <div>
              <dt>Payments refunded</dt>
              <dd>{formatMinorAmount(summary.paymentsRefundedAmountMinor)}</dd>
            </div>
            <div>
              <dt>Returns settled</dt>
              <dd>
                {summary.returnsSettledCount} /{' '}
                {formatMinorAmount(summary.returnsSettledAmountMinor)}
              </dd>
            </div>
            <div>
              <dt>Cash movements posted</dt>
              <dd>{summary.cashMovementsPostedCount}</dd>
            </div>
            <div>
              <dt>Operational days closed</dt>
              <dd>{summary.operationalDaysClosedCount}</dd>
            </div>
          </dl>
        ) : (
          <p className="muted">No summary data.</p>
        )}
      </div>

      <div className="panel">
        <div className="panel-heading">
          <h3>Per-store breakdown</h3>
          <p className="muted">
            {totalCount === 0
              ? 'No stores in window'
              : `Showing ${pageStart}–${pageEnd} of ${totalCount}`}
          </p>
        </div>

        {storesQuery.isLoading && !stores ? (
          <p className="muted">Loading store rows…</p>
        ) : stores && stores.items.length > 0 ? (
          <>
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>Store</th>
                    <th>Receipts</th>
                    <th>Receipt amount</th>
                    <th>Returns settled</th>
                    <th>Payments captured</th>
                    <th>Cash movements</th>
                    <th>Days closed</th>
                  </tr>
                </thead>
                <tbody>
                  {stores.items.map((item) => (
                    <tr key={item.storeId}>
                      <td>{item.storeId}</td>
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
            <div className="pagination">
              <button
                className="secondary"
                disabled={!canGoPrev || isLoading}
                onClick={() => setOffset((current) => Math.max(0, current - PAGE_SIZE))}
                type="button"
              >
                Previous
              </button>
              <button
                className="secondary"
                disabled={!canGoNext || isLoading}
                onClick={() => setOffset((current) => current + PAGE_SIZE)}
                type="button"
              >
                Next
              </button>
            </div>
          </>
        ) : (
          <p className="muted">No store rows for the selected window.</p>
        )}
      </div>
    </section>
  );
}

import { useGetStoreReportingSummary, useListStores } from '@mercadia/api-clients-central';
import { useMemo, useState } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { PageBackLink } from './users-shared.js';
import { ReportingKpiGrid } from './ReportingKpiGrid.js';
import {
  defaultReportingWindow,
  formatTimestamp,
  fromDatetimeLocalValue,
  toDatetimeLocalValue,
} from './reporting-utils.js';

function readWindowFromSearchParams(searchParams: URLSearchParams): {
  since: string;
  until: string;
} {
  const since = searchParams.get('since');
  const until = searchParams.get('until');
  if (since && until) {
    return { since, until };
  }
  return defaultReportingWindow();
}

export function StoreReportingPage() {
  const { storeId = '' } = useParams();
  const [searchParams] = useSearchParams();
  const initialWindow = useMemo(() => readWindowFromSearchParams(searchParams), [searchParams]);

  const [sinceInput, setSinceInput] = useState(toDatetimeLocalValue(initialWindow.since));
  const [untilInput, setUntilInput] = useState(toDatetimeLocalValue(initialWindow.until));
  const [applied, setApplied] = useState(initialWindow);

  const storesQuery = useListStores();
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data.stores : [];
  const storeName = stores.find((store) => store.id === storeId)?.name;

  const summaryQuery = useGetStoreReportingSummary(
    storeId,
    { since: applied.since, until: applied.until },
    { query: { enabled: storeId.length > 0 } },
  );

  const summary = summaryQuery.data?.status === 200 ? summaryQuery.data.data : null;
  const isLoading = summaryQuery.isFetching;
  const errorMessage = summaryQuery.error != null ? getApiErrorMessage(summaryQuery.error) : null;

  function applyFilters() {
    setApplied({
      since: fromDatetimeLocalValue(sinceInput),
      until: fromDatetimeLocalValue(untilInput),
    });
  }

  const title = storeName ? `${storeName} (${storeId})` : storeId;

  return (
    <section className="stack reporting-page">
      <PageBackLink label="Back to reporting" to="/central/reporting" />

      <div className="panel">
        <div className="panel-heading">
          <div>
            <h2>Store Reporting</h2>
            <p className="muted">
              {title || 'Unknown store'} — {formatTimestamp(applied.since)} –{' '}
              {formatTimestamp(applied.until)} UTC
            </p>
          </div>
          <button
            className="secondary"
            disabled={isLoading || storeId.length === 0}
            onClick={() => void summaryQuery.refetch()}
            type="button"
          >
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
          <button disabled={isLoading || storeId.length === 0} type="submit">
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
        <h3>Store KPIs</h3>
        {summaryQuery.isLoading && !summary ? (
          <p className="muted">Loading summary…</p>
        ) : summary ? (
          <ReportingKpiGrid data={summary} />
        ) : (
          <p className="muted">No summary data.</p>
        )}
      </div>
    </section>
  );
}

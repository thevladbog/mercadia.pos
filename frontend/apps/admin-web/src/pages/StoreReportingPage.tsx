import { useGetStoreReportingSummary, useListStores } from '@mercadia/api-clients-central';
import { Button } from '@mercadia/ui';
import { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
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
  const { t } = useTranslation();
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
      <PageBackLink label={t('reporting.backToReporting')} to="/central/reporting" />

      <div className="panel">
        <div className="panel-heading">
          <div>
            <h2>{t('reporting.storeTitle')}</h2>
            <p className="muted">
              {title || t('reporting.unknownStore')} {t('common.emDash')}{' '}
              {formatTimestamp(applied.since)} {t('common.emDash')} {formatTimestamp(applied.until)}{' '}
              UTC
            </p>
          </div>
          <Button
            variant="secondary"
            disabled={isLoading || storeId.length === 0}
            onClick={() => void summaryQuery.refetch()}
            type="button"
          >
            {isLoading ? t('common.refreshing') : t('common.refresh')}
          </Button>
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
          <Button disabled={isLoading || storeId.length === 0} type="submit">
            {t('common.apply')}
          </Button>
        </form>
      </div>

      {errorMessage ? (
        <div className="panel error-panel">
          <p className="error">{errorMessage}</p>
        </div>
      ) : null}

      <div className="panel">
        <h3>{t('reporting.storeKpis')}</h3>
        {summaryQuery.isLoading && !summary ? (
          <p className="muted">{t('reporting.loadingSummary')}</p>
        ) : summary ? (
          <ReportingKpiGrid data={summary} />
        ) : (
          <p className="muted">{t('reporting.noSummary')}</p>
        )}
      </div>
    </section>
  );
}

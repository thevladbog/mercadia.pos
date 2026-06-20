import { useListStores } from '@mercadia/api-clients-central';
import { ApiError } from '@mercadia/api-clients-store-edge';
import {
  useGetCurrentOperationalDay,
  useGetOperationalDaySummary,
  useListCashBalances,
  useListOpenStoreShifts,
  useListOperationJournal,
  type ListOpenStoreShifts200ShiftsItem,
} from '@mercadia/api-clients-store-edge';
import { Button } from '@mercadia/ui';
import { useMemo, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { canWriteStoreOperations } from '@/auth/permissions.js';
import { useAuth } from '@/auth/useAuth.js';
import { CloseShiftModal } from '@/components/eod/CloseShiftModal.js';
import { BlockerActionCell, BlockerReferenceCell } from '@/components/eod/BlockerReferenceCell.js';
import { EodActionsPanel } from '@/components/eod/EodActionsPanel.js';
import { EodOpenPanel } from '@/components/eod/EodOpenPanel.js';
import { JournalReferenceCell } from '@/components/eod/JournalReferenceCell.js';
import { OperationalDayReceiptsPanel } from '@/components/eod/OperationalDayReceiptsPanel.js';
import { OperationalDayShiftsPanel } from '@/components/eod/OperationalDayShiftsPanel.js';
import { ReceiptDetailModal } from '@/components/eod/ReceiptDetailModal.js';
import { ReturnDetailModal } from '@/components/eod/ReturnDetailModal.js';
import { ShiftDetailModal } from '@/components/eod/ShiftDetailModal.js';
import { StoreReturnsPanel } from '@/components/eod/StoreReturnsPanel.js';
import { PaginationControls } from '@/components/PaginationControls.js';
import { StorePicker } from '@/components/StorePicker.js';
import { formatBlockerMessage, formatBlockerSeverity } from './eod-mutation-utils.js';
import { journalOperationTypeKey } from './eod-journal-utils.js';
import { formatMinorAmount, formatTimestamp, PAGE_SIZE } from './reporting-utils.js';
import { readStoreFromSearchParams } from './store-routes.js';
import { STORE_POLL_INTERVAL_MS } from './store-polling.js';

type EodTab = 'overview' | 'receipts' | 'shifts' | 'returns' | 'open-shifts' | 'journal';

function formatElapsed(openedAt: string): string {
  const elapsedMs = Date.now() - new Date(openedAt).getTime();
  const hours = Math.floor(elapsedMs / (1000 * 60 * 60));
  const minutes = Math.floor((elapsedMs % (1000 * 60 * 60)) / (1000 * 60));
  return `${hours}h ${minutes}m`;
}

export function StoreEodPage() {
  const { t } = useTranslation();
  const { roles } = useAuth();
  const canWrite = canWriteStoreOperations(roles);
  const [searchParams] = useSearchParams();
  const initialStoreId = readStoreFromSearchParams(searchParams);

  const storesQuery = useListStores();
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data.stores : [];
  const [selectedStoreId, setSelectedStoreId] = useState<string | null>(initialStoreId);
  const activeStoreId = selectedStoreId ?? stores[0]?.id ?? '';
  const [activeTab, setActiveTab] = useState<EodTab>('overview');
  const [journalOffset, setJournalOffset] = useState(0);
  const [closeShiftTarget, setCloseShiftTarget] = useState<ListOpenStoreShifts200ShiftsItem | null>(
    null,
  );
  const [detailShiftId, setDetailShiftId] = useState<string | null>(null);
  const [detailReceiptId, setDetailReceiptId] = useState<string | null>(null);
  const [detailReturnId, setDetailReturnId] = useState<string | null>(null);

  const pollOptions = useMemo(
    () => ({
      query: {
        enabled: activeStoreId.length > 0,
        refetchInterval: STORE_POLL_INTERVAL_MS,
      },
    }),
    [activeStoreId],
  );

  const currentDayQuery = useGetCurrentOperationalDay(activeStoreId, pollOptions);
  const operationalDay = currentDayQuery.data?.status === 200 ? currentDayQuery.data.data : null;
  const isNoOpenDay =
    currentDayQuery.error instanceof ApiError && currentDayQuery.error.status === 404;
  const operationalDayId = operationalDay?.id ?? '';

  const summaryQuery = useGetOperationalDaySummary(operationalDayId, {
    query: {
      enabled: operationalDayId.length > 0,
      refetchInterval: STORE_POLL_INTERVAL_MS,
    },
  });
  const summary = summaryQuery.data?.status === 200 ? summaryQuery.data.data : null;

  const openShiftsQuery = useListOpenStoreShifts(activeStoreId, {
    query: {
      enabled: activeStoreId.length > 0 && activeTab === 'open-shifts',
      refetchInterval: STORE_POLL_INTERVAL_MS,
    },
  });
  const openShifts = openShiftsQuery.data?.status === 200 ? openShiftsQuery.data.data.shifts : null;

  const balancesQuery = useListCashBalances(activeStoreId, {
    query: {
      enabled: activeStoreId.length > 0 && activeTab === 'open-shifts' && canWrite,
      refetchInterval: STORE_POLL_INTERVAL_MS,
    },
  });
  const shiftCloseBalances =
    balancesQuery.data?.status === 200 ? balancesQuery.data.data.balances : [];

  const journalQuery = useListOperationJournal(
    activeStoreId,
    { limit: PAGE_SIZE, offset: journalOffset },
    {
      query: {
        enabled: activeStoreId.length > 0 && activeTab === 'journal',
        refetchInterval: STORE_POLL_INTERVAL_MS,
      },
    },
  );
  const journalPage = journalQuery.data?.status === 200 ? journalQuery.data.data : null;
  const journalTotal = journalPage?.totalCount ?? 0;

  const isLoading =
    storesQuery.isFetching ||
    (activeStoreId.length > 0 &&
      (currentDayQuery.isFetching ||
        summaryQuery.isFetching ||
        (activeTab === 'open-shifts' && (openShiftsQuery.isFetching || balancesQuery.isFetching)) ||
        (activeTab === 'journal' && journalQuery.isFetching)));

  const errorMessage =
    storesQuery.error != null
      ? getApiErrorMessage(storesQuery.error)
      : currentDayQuery.error != null && !isNoOpenDay
        ? getApiErrorMessage(currentDayQuery.error)
        : summaryQuery.error != null
          ? getApiErrorMessage(summaryQuery.error)
          : openShiftsQuery.error != null
            ? getApiErrorMessage(openShiftsQuery.error)
            : balancesQuery.error != null
              ? getApiErrorMessage(balancesQuery.error)
              : journalQuery.error != null
                ? getApiErrorMessage(journalQuery.error)
                : null;

  function refetchAll() {
    void storesQuery.refetch();
    if (activeStoreId.length > 0) {
      void currentDayQuery.refetch();
      if (operationalDayId.length > 0) {
        void summaryQuery.refetch();
      }
      if (activeTab === 'open-shifts') {
        void openShiftsQuery.refetch();
        if (canWrite) {
          void balancesQuery.refetch();
        }
      }
      if (activeTab === 'journal') {
        void journalQuery.refetch();
      }
    }
  }

  function handleEodTab(tab: 'open-shifts') {
    setActiveTab(tab);
  }

  function handleOpenShift(shiftId: string) {
    setDetailShiftId(shiftId);
  }

  function handleOpenReceipt(receiptId: string) {
    setDetailReceiptId(receiptId);
  }

  function handleOpenReturn(returnId: string) {
    setDetailReturnId(returnId);
  }

  return (
    <section className="stack monitoring-page">
      <div className="panel">
        <div className="panel-heading">
          <div>
            <h2>{t('eod.title')}</h2>
            <p className="muted">{t('eod.subtitle')}</p>
          </div>
          <Button variant="secondary" disabled={isLoading} onClick={refetchAll} type="button">
            {isLoading ? t('common.refreshing') : t('common.refresh')}
          </Button>
        </div>

        <StorePicker
          loading={storesQuery.isLoading}
          stores={stores}
          value={activeStoreId}
          onChange={(storeId) => {
            setSelectedStoreId(storeId);
            setJournalOffset(0);
          }}
        />
      </div>

      {errorMessage ? (
        <div className="panel error-panel">
          <p className="error">{errorMessage}</p>
        </div>
      ) : null}

      {!activeStoreId ? (
        <div className="panel">
          <p className="muted">{t('eod.selectStore')}</p>
        </div>
      ) : currentDayQuery.isLoading && !operationalDay ? (
        <div className="panel">
          <p className="muted">{t('eod.loadingDay')}</p>
        </div>
      ) : isNoOpenDay || !operationalDay ? (
        <>
          <div className="panel">
            <p className="muted">{t('eod.noOpenDay')}</p>
          </div>
          <EodOpenPanel canWrite={canWrite} storeId={activeStoreId} />
        </>
      ) : (
        <>
          <div className="panel tab-bar">
            <Button
              variant={activeTab === 'overview' ? 'primary' : 'secondary'}
              onClick={() => setActiveTab('overview')}
              type="button"
            >
              {t('eod.tabs.overview')}
            </Button>
            <Button
              variant={activeTab === 'receipts' ? 'primary' : 'secondary'}
              onClick={() => setActiveTab('receipts')}
              type="button"
            >
              {t('eod.tabs.receipts')}
            </Button>
            <Button
              variant={activeTab === 'shifts' ? 'primary' : 'secondary'}
              onClick={() => setActiveTab('shifts')}
              type="button"
            >
              {t('eod.tabs.shifts')}
            </Button>
            <Button
              variant={activeTab === 'returns' ? 'primary' : 'secondary'}
              onClick={() => setActiveTab('returns')}
              type="button"
            >
              {t('eod.tabs.returns')}
            </Button>
            <Button
              variant={activeTab === 'open-shifts' ? 'primary' : 'secondary'}
              onClick={() => setActiveTab('open-shifts')}
              type="button"
            >
              {t('eod.tabs.openShifts')}
            </Button>
            <Button
              variant={activeTab === 'journal' ? 'primary' : 'secondary'}
              onClick={() => setActiveTab('journal')}
              type="button"
            >
              {t('eod.tabs.journal')}
            </Button>
          </div>

          {activeTab === 'overview' ? (
            <>
              <div className="panel">
                <div className="panel-heading">
                  <h3>{t('eod.tabs.overview')}</h3>
                  <span
                    className={
                      summary?.canClose
                        ? 'status-badge status-online'
                        : 'status-badge status-attention'
                    }
                  >
                    {summary?.canClose ? t('eod.canClose') : t('eod.cannotClose')}
                  </span>
                </div>

                <dl className="kpi-grid">
                  <div>
                    <dt>{t('eod.businessDate')}</dt>
                    <dd>{operationalDay.businessDate}</dd>
                  </div>
                  <div>
                    <dt>{t('eod.dayStatus')}</dt>
                    <dd>{operationalDay.status}</dd>
                  </div>
                  <div>
                    <dt>{t('eod.openedAt')}</dt>
                    <dd>{formatTimestamp(operationalDay.openedAt)}</dd>
                  </div>
                  <div>
                    <dt>{t('eod.elapsed')}</dt>
                    <dd>{formatElapsed(operationalDay.openedAt)}</dd>
                  </div>
                </dl>
              </div>

              {summary ? (
                <EodActionsPanel
                  blockers={summary.blockers}
                  canWrite={canWrite}
                  operationalDayId={operationalDayId}
                  storeId={activeStoreId}
                  onEodTab={handleEodTab}
                />
              ) : null}

              <div className="panel">
                <h3>{t('eod.blockers')}</h3>
                <p className="muted">{t('eod.blockerDisclaimer')}</p>
                {summaryQuery.isLoading && !summary ? (
                  <p className="muted">{t('common.loading')}</p>
                ) : summary && summary.blockers.length > 0 ? (
                  <div className="table-wrap">
                    <table>
                      <thead>
                        <tr>
                          <th>{t('eod.severity')}</th>
                          <th>{t('eod.code')}</th>
                          <th>{t('eod.message')}</th>
                          <th>{t('eod.reference')}</th>
                          <th>{t('eod.action')}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {summary.blockers.map((blocker) => (
                          <tr key={`${blocker.code}-${blocker.referenceId ?? blocker.message}`}>
                            <td>{formatBlockerSeverity(blocker.severity, t)}</td>
                            <td>{blocker.code}</td>
                            <td>{formatBlockerMessage(blocker, t)}</td>
                            <td>
                              <BlockerReferenceCell
                                blocker={blocker}
                                storeId={activeStoreId}
                                onEodTab={handleEodTab}
                                onOpenReceipt={handleOpenReceipt}
                                onOpenShift={handleOpenShift}
                              />
                            </td>
                            <td>
                              <BlockerActionCell
                                blocker={blocker}
                                storeId={activeStoreId}
                                onEodTab={handleEodTab}
                                onOpenReceipt={handleOpenReceipt}
                                onOpenShift={handleOpenShift}
                              />
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                ) : (
                  <p className="muted">{t('eod.noBlockers')}</p>
                )}
              </div>

              {summary ? (
                <>
                  <div className="panel">
                    <h3>{t('eod.summaryCash')}</h3>
                    <dl className="kpi-grid">
                      <div>
                        <dt>{t('eod.nonZeroDrawers')}</dt>
                        <dd>{summary.cash.nonZeroDrawerCount}</dd>
                      </div>
                      <div>
                        <dt>{t('eod.kpi.recountTotal')}</dt>
                        <dd>{summary.cash.recounts.totalCount}</dd>
                      </div>
                      <div>
                        <dt>{t('eod.kpi.openRecountDiscrepancies')}</dt>
                        <dd>{summary.cash.recounts.openDiscrepancyCount}</dd>
                      </div>
                    </dl>
                    {summary.cash.balances.length > 0 ? (
                      <div className="table-wrap">
                        <table>
                          <thead>
                            <tr>
                              <th>{t('eod.kpi.containerId')}</th>
                              <th>{t('eod.kpi.containerType')}</th>
                              <th>{t('eod.kpi.balance')}</th>
                              <th>{t('eod.kpi.lastMovement')}</th>
                            </tr>
                          </thead>
                          <tbody>
                            {summary.cash.balances.map((balance) => (
                              <tr key={balance.containerId}>
                                <td>{balance.containerId}</td>
                                <td>{balance.containerType}</td>
                                <td>{formatMinorAmount(balance.balanceMinor)}</td>
                                <td>{formatTimestamp(balance.lastMovementAt)}</td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    ) : null}
                  </div>

                  <div className="panel">
                    <h3>{t('eod.summaryReceipts')}</h3>
                    <dl className="kpi-grid">
                      <div>
                        <dt>{t('eod.totalReceipts')}</dt>
                        <dd>{summary.receipts.totalCount}</dd>
                      </div>
                      <div>
                        <dt>{t('eod.kpi.draftReceipts')}</dt>
                        <dd>{summary.receipts.draftCount}</dd>
                      </div>
                      <div>
                        <dt>{t('eod.kpi.paymentStartedReceipts')}</dt>
                        <dd>{summary.receipts.paymentStartedCount}</dd>
                      </div>
                      <div>
                        <dt>{t('eod.paidReceipts')}</dt>
                        <dd>{summary.receipts.paidCount}</dd>
                      </div>
                      <div>
                        <dt>{t('eod.kpi.unresolvedReceipts')}</dt>
                        <dd>{summary.receipts.unresolvedCount}</dd>
                      </div>
                      <div>
                        <dt>{t('eod.kpi.cancelledReceipts')}</dt>
                        <dd>{summary.receipts.cancelledCount}</dd>
                      </div>
                      <div>
                        <dt>{t('eod.fiscalizedReceipts')}</dt>
                        <dd>{summary.receipts.fiscalizedCount}</dd>
                      </div>
                      <div>
                        <dt>{t('eod.kpi.fiscalizedSales')}</dt>
                        <dd>{formatMinorAmount(summary.receipts.fiscalizedSalesMinor)}</dd>
                      </div>
                    </dl>
                  </div>

                  <div className="panel">
                    <h3>{t('eod.summaryPayments')}</h3>
                    <dl className="kpi-grid">
                      <div>
                        <dt>{t('eod.kpi.totalPayments')}</dt>
                        <dd>{summary.payments.totalCount}</dd>
                      </div>
                      <div>
                        <dt>{t('eod.capturedPayments')}</dt>
                        <dd>{summary.payments.capturedCount}</dd>
                      </div>
                      <div>
                        <dt>{t('reporting.paymentsCaptured')}</dt>
                        <dd>{formatMinorAmount(summary.payments.capturedTotalMinor)}</dd>
                      </div>
                    </dl>
                    {summary.payments.methods.length > 0 ? (
                      <div className="table-wrap">
                        <table>
                          <thead>
                            <tr>
                              <th>{t('eod.kpi.paymentMethod')}</th>
                              <th>{t('eod.kpi.paymentCount')}</th>
                              <th>{t('eod.kpi.paymentTotal')}</th>
                            </tr>
                          </thead>
                          <tbody>
                            {summary.payments.methods.map((method) => (
                              <tr key={method.method}>
                                <td>{method.method}</td>
                                <td>{method.capturedCount}</td>
                                <td>{formatMinorAmount(method.capturedTotalMinor)}</td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    ) : null}
                  </div>

                  <div className="panel">
                    <h3>{t('eod.summaryFiscal')}</h3>
                    <dl className="kpi-grid">
                      <div>
                        <dt>{t('eod.kpi.totalFiscal')}</dt>
                        <dd>{summary.fiscal.totalCount}</dd>
                      </div>
                      <div>
                        <dt>{t('eod.fiscalizedReceipts')}</dt>
                        <dd>{summary.fiscal.fiscalizedCount}</dd>
                      </div>
                      <div>
                        <dt>{t('monitoring.revenue')}</dt>
                        <dd>{formatMinorAmount(summary.fiscal.fiscalizedTotalMinor)}</dd>
                      </div>
                    </dl>
                  </div>

                  <div className="panel">
                    <h3>{t('eod.summaryShifts')}</h3>
                    <dl className="kpi-grid">
                      <div>
                        <dt>{t('eod.openShiftsCount')}</dt>
                        <dd>{summary.shifts.openCount}</dd>
                      </div>
                      <div>
                        <dt>{t('eod.closedShiftsCount')}</dt>
                        <dd>{summary.shifts.closedCount}</dd>
                      </div>
                    </dl>
                  </div>
                </>
              ) : null}
            </>
          ) : null}

          {activeTab === 'receipts' ? (
            <OperationalDayReceiptsPanel
              operationalDayId={operationalDayId}
              onOpenReceipt={handleOpenReceipt}
            />
          ) : null}

          {activeTab === 'shifts' ? (
            <OperationalDayShiftsPanel
              operationalDayId={operationalDayId}
              onOpenShift={handleOpenShift}
            />
          ) : null}

          {activeTab === 'returns' ? (
            <StoreReturnsPanel
              businessDate={operationalDay.businessDate}
              storeId={activeStoreId}
              onOpenReturn={handleOpenReturn}
            />
          ) : null}

          {activeTab === 'open-shifts' ? (
            <div className="panel">
              <h3>{t('eod.tabs.openShifts')}</h3>
              {openShiftsQuery.isLoading && !openShifts ? (
                <p className="muted">{t('eod.loadingShifts')}</p>
              ) : openShifts && openShifts.length > 0 ? (
                <div className="table-wrap">
                  <table>
                    <thead>
                      <tr>
                        <th>{t('eod.shiftId')}</th>
                        <th>{t('monitoring.cashier')}</th>
                        <th>{t('eod.terminalId')}</th>
                        <th>{t('monitoring.status')}</th>
                        <th>{t('eod.opened')}</th>
                        <th>{t('eod.openingCash')}</th>
                        {canWrite ? <th>{t('eod.actions.closeShiftColumn')}</th> : null}
                      </tr>
                    </thead>
                    <tbody>
                      {openShifts.map((shift) => (
                        <tr key={shift.id}>
                          <td>
                            <Button
                              variant="link"
                              size="sm"
                              onClick={() => handleOpenShift(shift.id)}
                              type="button"
                            >
                              {shift.id}
                            </Button>
                          </td>
                          <td>{shift.cashierId}</td>
                          <td>{shift.terminalId}</td>
                          <td>{shift.status}</td>
                          <td>{formatTimestamp(shift.openedAt)}</td>
                          <td>{formatMinorAmount(shift.openingCashMinor)}</td>
                          {canWrite ? (
                            <td>
                              <Button
                                variant="secondary"
                                size="sm"
                                onClick={() => setCloseShiftTarget(shift)}
                                type="button"
                              >
                                {t('eod.actions.closeShift')}
                              </Button>
                            </td>
                          ) : null}
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <p className="muted">{t('eod.noOpenShifts')}</p>
              )}
              {closeShiftTarget ? (
                <CloseShiftModal
                  balances={shiftCloseBalances}
                  shift={closeShiftTarget}
                  storeId={activeStoreId}
                  onClose={() => setCloseShiftTarget(null)}
                />
              ) : null}
            </div>
          ) : null}

          {activeTab === 'journal' ? (
            <div className="panel">
              <h3>{t('eod.journalTitle')}</h3>
              {journalQuery.isLoading && !journalPage ? (
                <p className="muted">{t('eod.loadingJournal')}</p>
              ) : journalPage && journalPage.items.length > 0 ? (
                <>
                  <div className="table-wrap">
                    <table>
                      <thead>
                        <tr>
                          <th>{t('eod.created')}</th>
                          <th>{t('eod.operationType')}</th>
                          <th>{t('safe.actor')}</th>
                          <th>{t('eod.summary')}</th>
                          <th>{t('eod.reference')}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {journalPage.items.map((entry) => (
                          <tr key={entry.id}>
                            <td>{formatTimestamp(entry.createdAt)}</td>
                            <td>
                              {t(journalOperationTypeKey(entry.operationType), {
                                defaultValue: entry.operationType,
                              })}
                            </td>
                            <td>{entry.actorId}</td>
                            <td>{entry.summary ?? t('common.emDash')}</td>
                            <td>
                              <JournalReferenceCell
                                entry={entry}
                                storeId={activeStoreId}
                                onOpenReceipt={handleOpenReceipt}
                                onOpenReturn={handleOpenReturn}
                                onOpenShift={handleOpenShift}
                              />
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                  <PaginationControls
                    canGoNext={journalOffset + PAGE_SIZE < journalTotal}
                    canGoPrev={journalOffset > 0}
                    disabled={journalQuery.isFetching}
                    onNext={() => setJournalOffset((value) => value + PAGE_SIZE)}
                    onPrev={() => setJournalOffset((value) => Math.max(0, value - PAGE_SIZE))}
                  />
                </>
              ) : (
                <p className="muted">{t('eod.noJournal')}</p>
              )}
            </div>
          ) : null}
        </>
      )}

      {detailShiftId ? (
        <ShiftDetailModal
          canWrite={canWrite}
          shiftId={detailShiftId}
          onClose={() => setDetailShiftId(null)}
          onEodTab={handleEodTab}
          onOpenReceipt={(receiptId) => {
            setDetailShiftId(null);
            setDetailReceiptId(receiptId);
          }}
        />
      ) : null}

      {detailReceiptId ? (
        <ReceiptDetailModal receiptId={detailReceiptId} onClose={() => setDetailReceiptId(null)} />
      ) : null}

      {detailReturnId ? (
        <ReturnDetailModal
          returnId={detailReturnId}
          onClose={() => setDetailReturnId(null)}
          onOpenReceipt={(receiptId) => {
            setDetailReturnId(null);
            setDetailReceiptId(receiptId);
          }}
        />
      ) : null}
    </section>
  );
}

import { useListStores } from '@mercadia/api-clients-central';
import {
  useListCashBalances,
  useListCashMovements,
  useListCashRecounts,
  type ListCashMovements200ItemsItem,
  type ListCashRecounts200ItemsItem,
} from '@mercadia/api-clients-store-edge';
import { Button } from '@mercadia/ui';
import { useMemo, useState } from 'react';
import { useNavigate, useSearchParams, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { useAuth } from '@/auth/useAuth.js';
import { canWriteCashOperations } from '@/auth/permissions.js';
import { CashActionsPanel } from '@/components/cash/CashActionsPanel.js';
import { CashMovementDetailModal } from '@/components/cash/CashMovementDetailModal.js';
import { CashRecountDetailModal } from '@/components/cash/CashRecountDetailModal.js';
import { SafeKpiPanel } from '@/components/cash/SafeKpiPanel.js';
import { ResolveRecountModal } from '@/components/cash/ResolveRecountModal.js';
import { PaginationControls } from '@/components/PaginationControls.js';
import { StorePicker } from '@/components/StorePicker.js';
import { formatMinorAmount, formatTimestamp, PAGE_SIZE } from './reporting-utils.js';
import { computeSafeBalanceRollups, countOpenRecountDiscrepancies } from './safe-kpi-utils.js';
import { matchesMovementSearch, matchesRecountSearch } from './safe-search-utils.js';
import {
  readMovementFromSearchParams,
  readRecountFromSearchParams,
  readStoreFromSearchParams,
  storePageHref,
} from './store-routes.js';
import { STORE_POLL_INTERVAL_MS } from './store-polling.js';

export function StoreSafePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const { roles } = useAuth();
  const [searchParams] = useSearchParams();
  const initialStoreId = readStoreFromSearchParams(searchParams);
  const recountDeepLinkId = readRecountFromSearchParams(searchParams);
  const movementDeepLinkId = readMovementFromSearchParams(searchParams);
  const canWrite = canWriteCashOperations(roles);

  const storesQuery = useListStores();
  const stores = storesQuery.data?.status === 200 ? storesQuery.data.data.stores : [];
  const [selectedStoreId, setSelectedStoreId] = useState<string | null>(initialStoreId);
  const activeStoreId = selectedStoreId ?? stores[0]?.id ?? '';

  const [movementsOffset, setMovementsOffset] = useState(0);
  const [recountsOffset, setRecountsOffset] = useState(0);
  const [resolveRecount, setResolveRecount] = useState<ListCashRecounts200ItemsItem | null>(null);
  const [detailMovement, setDetailMovement] = useState<ListCashMovements200ItemsItem | null>(null);
  const [detailRecount, setDetailRecount] = useState<ListCashRecounts200ItemsItem | null>(null);
  const [movementsSearch, setMovementsSearch] = useState('');
  const [recountsSearch, setRecountsSearch] = useState('');
  const [dismissedDeepLinkLocationKey, setDismissedDeepLinkLocationKey] = useState<string | null>(
    null,
  );

  const pollOptions = useMemo(
    () => ({
      query: {
        enabled: activeStoreId.length > 0,
        refetchInterval: STORE_POLL_INTERVAL_MS,
      },
    }),
    [activeStoreId],
  );

  const balancesQuery = useListCashBalances(activeStoreId, pollOptions);
  const movementsQuery = useListCashMovements(
    activeStoreId,
    { limit: PAGE_SIZE, offset: movementsOffset },
    pollOptions,
  );
  const recountsQuery = useListCashRecounts(
    activeStoreId,
    { limit: PAGE_SIZE, offset: recountsOffset },
    pollOptions,
  );
  const recountRollupQuery = useListCashRecounts(
    activeStoreId,
    { limit: 100, offset: 0 },
    pollOptions,
  );

  const balances = balancesQuery.data?.status === 200 ? balancesQuery.data.data.balances : null;
  const movementsPage = movementsQuery.data?.status === 200 ? movementsQuery.data.data : null;
  const recountsPage = recountsQuery.data?.status === 200 ? recountsQuery.data.data : null;
  const recountRollupPage =
    recountRollupQuery.data?.status === 200 ? recountRollupQuery.data.data : null;

  const balanceRollups = useMemo(
    () => (balances ? computeSafeBalanceRollups(balances) : null),
    [balances],
  );
  const openRecountCount = recountRollupPage
    ? countOpenRecountDiscrepancies(recountRollupPage.items)
    : null;
  const openRecountPartial = (recountRollupPage?.totalCount ?? 0) > 100;

  const movementsTotal = movementsPage?.totalCount ?? 0;
  const recountsTotal = recountsPage?.totalCount ?? 0;

  const normalizedMovementsSearch = movementsSearch.trim().toLowerCase();
  const normalizedRecountsSearch = recountsSearch.trim().toLowerCase();
  const filteredMovements = useMemo(
    () =>
      movementsPage?.items.filter((movement) =>
        matchesMovementSearch(movement, normalizedMovementsSearch),
      ) ?? null,
    [movementsPage, normalizedMovementsSearch],
  );
  const filteredRecounts = useMemo(
    () =>
      recountsPage?.items.filter((recount) =>
        matchesRecountSearch(recount, normalizedRecountsSearch),
      ) ?? null,
    [recountsPage, normalizedRecountsSearch],
  );

  const deepLinkRecountOnPage = useMemo(() => {
    if (dismissedDeepLinkLocationKey === location.key || !recountDeepLinkId || !recountsPage) {
      return null;
    }

    return recountsPage.items.find((recount) => recount.id === recountDeepLinkId) ?? null;
  }, [dismissedDeepLinkLocationKey, location.key, recountDeepLinkId, recountsPage]);

  const deepLinkMovementOnPage = useMemo(() => {
    if (dismissedDeepLinkLocationKey === location.key || !movementDeepLinkId || !movementsPage) {
      return null;
    }

    return movementsPage.items.find((movement) => movement.id === movementDeepLinkId) ?? null;
  }, [dismissedDeepLinkLocationKey, location.key, movementDeepLinkId, movementsPage]);

  const deepLinkResolveRecount = useMemo(() => {
    if (!canWrite || !deepLinkRecountOnPage || deepLinkRecountOnPage.resolutionStatus !== 'open') {
      return null;
    }

    return deepLinkRecountOnPage;
  }, [canWrite, deepLinkRecountOnPage]);

  const deepLinkDetailRecount = useMemo(() => {
    if (dismissedDeepLinkLocationKey === location.key || detailRecount || !deepLinkRecountOnPage) {
      return null;
    }

    if (canWrite && deepLinkRecountOnPage.resolutionStatus === 'open') {
      return null;
    }

    return deepLinkRecountOnPage;
  }, [canWrite, detailRecount, deepLinkRecountOnPage, dismissedDeepLinkLocationKey, location.key]);

  const deepLinkDetailMovement = useMemo(() => {
    if (
      dismissedDeepLinkLocationKey === location.key ||
      detailMovement ||
      !deepLinkMovementOnPage
    ) {
      return null;
    }

    return deepLinkMovementOnPage;
  }, [detailMovement, deepLinkMovementOnPage, dismissedDeepLinkLocationKey, location.key]);

  const showRecountNotOnPageNotice =
    recountDeepLinkId != null &&
    dismissedDeepLinkLocationKey !== location.key &&
    recountsPage != null &&
    !recountsPage.items.some((recount) => recount.id === recountDeepLinkId);

  const showMovementNotOnPageNotice =
    movementDeepLinkId != null &&
    dismissedDeepLinkLocationKey !== location.key &&
    movementsPage != null &&
    !movementsPage.items.some((movement) => movement.id === movementDeepLinkId);

  const activeResolveRecount = resolveRecount ?? deepLinkResolveRecount;
  const activeDetailRecount = detailRecount ?? deepLinkDetailRecount;
  const activeDetailMovement = detailMovement ?? deepLinkDetailMovement;

  function handleResolveRecountClose() {
    setResolveRecount(null);
    if (recountDeepLinkId) {
      setDismissedDeepLinkLocationKey(location.key);
      void navigate(storePageHref('/store/safe', activeStoreId), { replace: true });
    }
  }

  function handleDetailRecountClose() {
    setDetailRecount(null);
    if (recountDeepLinkId) {
      setDismissedDeepLinkLocationKey(location.key);
      void navigate(storePageHref('/store/safe', activeStoreId), { replace: true });
    }
  }

  function handleDetailMovementClose() {
    setDetailMovement(null);
    if (movementDeepLinkId) {
      setDismissedDeepLinkLocationKey(location.key);
      void navigate(storePageHref('/store/safe', activeStoreId), { replace: true });
    }
  }

  const isLoading =
    storesQuery.isFetching ||
    (activeStoreId.length > 0 &&
      (balancesQuery.isFetching ||
        movementsQuery.isFetching ||
        recountsQuery.isFetching ||
        recountRollupQuery.isFetching));

  const errorMessage =
    storesQuery.error != null
      ? getApiErrorMessage(storesQuery.error)
      : balancesQuery.error != null
        ? getApiErrorMessage(balancesQuery.error)
        : movementsQuery.error != null
          ? getApiErrorMessage(movementsQuery.error)
          : recountsQuery.error != null
            ? getApiErrorMessage(recountsQuery.error)
            : recountRollupQuery.error != null
              ? getApiErrorMessage(recountRollupQuery.error)
              : null;

  function refetchAll() {
    void storesQuery.refetch();
    if (activeStoreId.length > 0) {
      void balancesQuery.refetch();
      void movementsQuery.refetch();
      void recountsQuery.refetch();
      void recountRollupQuery.refetch();
    }
  }

  return (
    <section className="stack monitoring-page">
      <div className="panel">
        <div className="panel-heading">
          <div>
            <h2>{t('safe.title')}</h2>
            <p className="muted">{t('safe.subtitle')}</p>
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
            setMovementsOffset(0);
            setRecountsOffset(0);
            setMovementsSearch('');
            setRecountsSearch('');
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
          <p className="muted">{t('safe.selectStore')}</p>
        </div>
      ) : (
        <>
          <SafeKpiPanel
            isLoading={balancesQuery.isLoading && !balances}
            movementsTotal={movementsPage?.totalCount ?? null}
            openRecountCount={openRecountCount}
            openRecountPartial={openRecountPartial}
            recountsTotal={recountsPage?.totalCount ?? null}
            rollups={balanceRollups}
          />

          {balances && balances.length > 0 ? (
            <CashActionsPanel balances={balances} canWrite={canWrite} storeId={activeStoreId} />
          ) : canWrite ? (
            <div className="panel">
              <p className="muted">{t('safe.actions.noContainers')}</p>
            </div>
          ) : null}

          <div className="panel">
            <h3>{t('safe.balances')}</h3>
            {balancesQuery.isLoading && !balances ? (
              <p className="muted">{t('safe.loadingBalances')}</p>
            ) : balances && balances.length > 0 ? (
              <dl className="kpi-grid">
                {balances.map((balance) => (
                  <div
                    className={
                      balance.containerType === 'safe' ? 'safe-balance-highlight' : undefined
                    }
                    key={balance.containerId}
                  >
                    <dt>
                      {balance.containerId} ({balance.containerType})
                    </dt>
                    <dd>{formatMinorAmount(balance.balanceMinor)}</dd>
                    <dd className="muted balance-meta">
                      {t('safe.lastMovement')}: {formatTimestamp(balance.lastMovementAt)}
                    </dd>
                  </div>
                ))}
              </dl>
            ) : (
              <p className="muted">{t('safe.noBalances')}</p>
            )}
          </div>

          <div className="panel">
            <h3>{t('safe.movements.title')}</h3>
            {showMovementNotOnPageNotice ? (
              <p className="muted">{t('safe.movementDetail.notOnPage')}</p>
            ) : null}
            {movementsQuery.isLoading && !movementsPage ? (
              <p className="muted">{t('safe.loadingMovements')}</p>
            ) : movementsPage && movementsPage.items.length > 0 ? (
              <>
                <label className="field terminal-search">
                  <span>{t('safe.movements.searchHint')}</span>
                  <input
                    placeholder={t('safe.movements.searchPlaceholder')}
                    type="search"
                    value={movementsSearch}
                    onChange={(event) => setMovementsSearch(event.target.value)}
                  />
                </label>
                {filteredMovements && filteredMovements.length > 0 ? (
                  <div className="table-wrap">
                    <table>
                      <thead>
                        <tr>
                          <th>{t('safe.movementId')}</th>
                          <th>{t('safe.type')}</th>
                          <th>{t('safe.amount')}</th>
                          <th>{t('safe.from')}</th>
                          <th>{t('safe.to')}</th>
                          <th>{t('safe.actor')}</th>
                          <th>{t('safe.posted')}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {filteredMovements.map((movement) => (
                          <tr key={movement.id}>
                            <td>
                              <Button
                                variant="link"
                                size="sm"
                                onClick={() => setDetailMovement(movement)}
                                type="button"
                              >
                                {movement.id}
                              </Button>
                            </td>
                            <td>{movement.type}</td>
                            <td>{formatMinorAmount(movement.amountMinor)}</td>
                            <td>
                              {movement.fromContainerId} ({movement.fromContainerType})
                            </td>
                            <td>
                              {movement.toContainerId} ({movement.toContainerType})
                            </td>
                            <td>{movement.actorId}</td>
                            <td>{formatTimestamp(movement.createdAt)}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                ) : (
                  <p className="muted">{t('safe.movements.noSearchResults')}</p>
                )}
                <PaginationControls
                  canGoNext={movementsOffset + PAGE_SIZE < movementsTotal}
                  canGoPrev={movementsOffset > 0}
                  disabled={movementsQuery.isFetching}
                  onNext={() => setMovementsOffset((value) => value + PAGE_SIZE)}
                  onPrev={() => setMovementsOffset((value) => Math.max(0, value - PAGE_SIZE))}
                />
              </>
            ) : (
              <p className="muted">{t('safe.noMovements')}</p>
            )}
          </div>

          <div className="panel">
            <h3>{t('safe.recounts.title')}</h3>
            {showRecountNotOnPageNotice ? (
              <p className="muted">{t('safe.recountDetail.notOnPage')}</p>
            ) : null}
            {recountsQuery.isLoading && !recountsPage ? (
              <p className="muted">{t('safe.loadingRecounts')}</p>
            ) : recountsPage && recountsPage.items.length > 0 ? (
              <>
                <label className="field terminal-search">
                  <span>{t('safe.recounts.searchHint')}</span>
                  <input
                    placeholder={t('safe.recounts.searchPlaceholder')}
                    type="search"
                    value={recountsSearch}
                    onChange={(event) => setRecountsSearch(event.target.value)}
                  />
                </label>
                {filteredRecounts && filteredRecounts.length > 0 ? (
                  <div className="table-wrap">
                    <table>
                      <thead>
                        <tr>
                          <th>{t('safe.recountId')}</th>
                          <th>{t('monitoring.status')}</th>
                          <th>{t('safe.expected')}</th>
                          <th>{t('safe.counted')}</th>
                          <th>{t('safe.variance')}</th>
                          <th>{t('safe.resolution')}</th>
                          <th>{t('eod.created')}</th>
                          {canWrite ? <th>{t('safe.actions.column')}</th> : null}
                        </tr>
                      </thead>
                      <tbody>
                        {filteredRecounts.map((recount) => (
                          <tr key={recount.id}>
                            <td>
                              <Button
                                variant="link"
                                size="sm"
                                onClick={() => setDetailRecount(recount)}
                                type="button"
                              >
                                {recount.id}
                              </Button>
                            </td>
                            <td>{recount.status}</td>
                            <td>{formatMinorAmount(recount.expectedMinor)}</td>
                            <td>{formatMinorAmount(recount.countedMinor)}</td>
                            <td>{formatMinorAmount(recount.discrepancyMinor)}</td>
                            <td>{recount.resolutionStatus}</td>
                            <td>{formatTimestamp(recount.createdAt)}</td>
                            {canWrite ? (
                              <td>
                                {recount.resolutionStatus === 'open' ? (
                                  <Button
                                    variant="secondary"
                                    size="sm"
                                    onClick={() => setResolveRecount(recount)}
                                    type="button"
                                  >
                                    {t('safe.actions.resolveRecount')}
                                  </Button>
                                ) : (
                                  t('common.emDash')
                                )}
                              </td>
                            ) : null}
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                ) : (
                  <p className="muted">{t('safe.recounts.noSearchResults')}</p>
                )}
                <PaginationControls
                  canGoNext={recountsOffset + PAGE_SIZE < recountsTotal}
                  canGoPrev={recountsOffset > 0}
                  disabled={recountsQuery.isFetching}
                  onNext={() => setRecountsOffset((value) => value + PAGE_SIZE)}
                  onPrev={() => setRecountsOffset((value) => Math.max(0, value - PAGE_SIZE))}
                />
              </>
            ) : (
              <p className="muted">{t('safe.noRecounts')}</p>
            )}
          </div>

          {activeResolveRecount ? (
            <ResolveRecountModal
              recount={activeResolveRecount}
              storeId={activeStoreId}
              onClose={handleResolveRecountClose}
            />
          ) : null}

          {activeDetailMovement ? (
            <CashMovementDetailModal
              movement={activeDetailMovement}
              onClose={handleDetailMovementClose}
            />
          ) : null}

          {activeDetailRecount ? (
            <CashRecountDetailModal
              recount={activeDetailRecount}
              onClose={handleDetailRecountClose}
            />
          ) : null}
        </>
      )}
    </section>
  );
}

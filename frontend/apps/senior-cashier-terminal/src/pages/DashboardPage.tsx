import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { Badge } from '@mercadia/ui';
import {
  useGetCredentialManagement,
  useGetStoreMonitoringSummary,
  useListCashBalances,
  useListOpenStoreShifts,
  useListOperationJournal,
  useListStoreMonitoringTerminals,
} from '@mercadia/api-clients-store-edge';

import { useAuth } from '@/auth/AuthProvider.js';
import { getStoreId } from '@/api-client-config.js';
import { selectSuccessData } from '@/lib/cash-utils.js';
import { deriveLoginAt } from '@/lib/topbar-utils.js';
import {
  deriveAlertsCount,
  deriveOperationsCount,
  deriveTotalTerminalCount,
  joinCashiersOnShift,
  type AlertsSummaryForDerive,
  type CredentialActorForJoin,
  type MonitoringTerminalForJoin,
  type OpenShiftForJoin,
  type OperationJournalEntryForJoin,
  type TerminalCountsSummaryForDerive,
} from '@/lib/dashboard-data.js';
import { useTopBarActions } from '@/lib/use-topbar-actions.js';
import { TopBar } from '@/components/TopBar.js';

import { CashierShiftRow } from './dashboard/CashierShiftRow.js';
import { DashboardClock } from './dashboard/DashboardClock.js';
import './dashboard/DashboardPage.css';
import { PrimaryActionsRow } from './dashboard/PrimaryActionsRow.js';
import { SafeNowPanel } from './dashboard/SafeNowPanel.js';
import { SafeOpsRow } from './dashboard/SafeOpsRow.js';
import { SystemRow } from './dashboard/SystemRow.js';

type MonitoringSummary = AlertsSummaryForDerive & TerminalCountsSummaryForDerive;

interface CashBalance {
  balanceMinor: number;
  containerType: string;
  lastMovementAt: string;
}

// The API-enforced max (`ListOperationJournalParams`) — see
// `deriveOperationsCount`'s doc comment for why this stays an approximation.
const JOURNAL_FETCH_LIMIT = 100;

// Stable empty-array fallbacks (module scope, never reassigned) so the
// `cashiersOnShift` useMemo below doesn't see a fresh `[]` reference — and
// therefore a false "changed" dependency — on every render while a query
// hasn't resolved yet.
const EMPTY_SHIFTS: OpenShiftForJoin[] = [];
const EMPTY_TERMINALS: MonitoringTerminalForJoin[] = [];
const EMPTY_ACTORS: CredentialActorForJoin[] = [];

/**
 * Redesigned senior-cashier dashboard (plan 021, Phase 4). A thin
 * orchestrator: fetches the 6 read-only endpoints, derives everything via
 * `lib/dashboard-data.ts`, and delegates all rendering to the row/panel
 * components under `pages/dashboard/`. See plan 021's "Why this matters"
 * for why several design-screen-02 elements are absent here (no
 * return/cancel card, no EoD card, no safe limit/fill-bar, no register
 * number, no last-recount timestamp) — every omission traces to a real
 * backend gap, not a shortcut, and must not be silently re-added.
 */
export function DashboardPage() {
  const { t } = useTranslation();
  const { onHandover, onLock } = useTopBarActions();
  const { session } = useAuth();
  const storeId = useMemo(() => getStoreId(), []);

  const { data: shiftsResp } = useListOpenStoreShifts(storeId);
  const { data: terminalsResp } = useListStoreMonitoringTerminals(storeId);
  const { data: summaryResp } = useGetStoreMonitoringSummary(storeId);
  const { data: credentialsResp } = useGetCredentialManagement(storeId);
  const { data: balancesResp } = useListCashBalances(storeId);
  const { data: journalResp } = useListOperationJournal(storeId, { limit: JOURNAL_FETCH_LIMIT });

  const shifts =
    selectSuccessData<{ shifts: OpenShiftForJoin[] }>(shiftsResp)?.shifts ?? EMPTY_SHIFTS;
  const terminals =
    selectSuccessData<{ items: MonitoringTerminalForJoin[] }>(terminalsResp)?.items ??
    EMPTY_TERMINALS;
  const summary = selectSuccessData<MonitoringSummary>(summaryResp);
  const actors =
    selectSuccessData<{ actors: CredentialActorForJoin[] }>(credentialsResp)?.actors ??
    EMPTY_ACTORS;
  const balances = selectSuccessData<{ balances: CashBalance[] }>(balancesResp)?.balances ?? [];
  const journalItems =
    selectSuccessData<{ items: OperationJournalEntryForJoin[] }>(journalResp)?.items ?? [];

  const cashiersOnShift = useMemo(
    () => joinCashiersOnShift(shifts, terminals, actors),
    [shifts, terminals, actors],
  );
  const safeBalance = balances.find((balance) => balance.containerType === 'safe');

  const operationsCount =
    session && journalResp
      ? deriveOperationsCount(
          journalItems,
          session.actorId,
          new Date(deriveLoginAt(session.expiresAt)).toISOString(),
        )
      : undefined;
  const alertsCount = summaryResp ? deriveAlertsCount(summary) : undefined;
  const totalTerminals = deriveTotalTerminalCount(summary);

  return (
    <div className="sr-terminal-shell">
      <TopBar
        onHandover={onHandover}
        onLock={onLock}
        operationsCount={operationsCount}
        alertsCount={alertsCount}
      />

      <main className="sr-terminal-main">
        <div className="sr-dashboard-header">
          <div>
            <span className="sr-dashboard-eyebrow">{t('dashboard.eyebrow')}</span>
            <h1 className="sr-dashboard-heading">{t('dashboard.heading')}</h1>
          </div>
          <DashboardClock />
        </div>

        <div className="sr-dashboard-body">
          <div className="sr-dashboard-main-column">
            <PrimaryActionsRow />

            <section>
              <h2 className="sr-dashboard-section-title">{t('dashboard.sectionSafeOps')}</h2>
              <SafeOpsRow />
            </section>

            <section>
              <h2 className="sr-dashboard-section-title">{t('dashboard.sectionSystem')}</h2>
              <SystemRow
                totalTerminals={totalTerminals}
                alertsCount={alertsCount}
                operationsCount={operationsCount}
              />
            </section>
          </div>

          <div className="sr-dashboard-side-column">
            <div className="sr-panel">
              <div className="sr-panel-header">
                <h2 className="sr-panel-title">{t('dashboard.activeCashiers')}</h2>
                <Badge variant="outline">{cashiersOnShift.length}</Badge>
              </div>
              {!shiftsResp && <p className="muted">{t('common.loading')}</p>}
              {shiftsResp && cashiersOnShift.length === 0 && (
                <p className="muted">{t('dashboard.noShifts')}</p>
              )}
              {cashiersOnShift.map((shift) => (
                <CashierShiftRow key={shift.shiftId} shift={shift} />
              ))}
            </div>

            <SafeNowPanel
              balanceMinor={safeBalance?.balanceMinor}
              lastMovementAtIso={safeBalance?.lastMovementAt}
            />
          </div>
        </div>
      </main>
    </div>
  );
}

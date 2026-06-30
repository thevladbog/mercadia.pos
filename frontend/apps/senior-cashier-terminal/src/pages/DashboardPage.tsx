import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Button } from '@mercadia/ui';
import {
  useListCashBalances,
  useListOpenStoreShifts,
  useListStoreTerminals,
} from '@mercadia/api-clients-store-edge';

import { useAuth } from '@/auth/AuthProvider.js';
import { useIdleTimer } from '@/lib/use-idle-timer.js';
import { getStoreId } from '@/api-client-config.js';
import { formatMinor, selectSuccessData } from '@/lib/cash-utils.js';
import { TerminalHeader } from '@/components/TerminalHeader.js';

export function DashboardPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { logout } = useAuth();
  const storeId = useMemo(() => getStoreId(), []);

  const { remaining } = useIdleTimer();

  const { data: balancesResp } = useListCashBalances(storeId);
  const { data: shiftsResp } = useListOpenStoreShifts(storeId);
  const { data: terminalsResp } = useListStoreTerminals(storeId);

  const safeBalance = useMemo(() => {
    const balances =
      selectSuccessData<{ balances: { containerType: string; balanceMinor: number }[] }>(
        balancesResp,
      )?.balances ?? [];
    const safe = balances.find((b) => b.containerType === 'safe');
    return safe?.balanceMinor ?? 0;
  }, [balancesResp]);

  const drawerTotal = useMemo(() => {
    const balances =
      selectSuccessData<{ balances: { containerType: string; balanceMinor: number }[] }>(
        balancesResp,
      )?.balances ?? [];
    return balances
      .filter((b) => b.containerType === 'drawer')
      .reduce((sum, b) => sum + (b.balanceMinor ?? 0), 0);
  }, [balancesResp]);

  const balancesData = selectSuccessData<{
    balances: { containerType: string; balanceMinor: number; containerId: string }[];
  }>(balancesResp);
  const shiftsData = selectSuccessData<{
    shifts: { id: string; cashierId: string; drawerId: string; closingCashMinor: number }[];
  }>(shiftsResp);
  const terminalsData = selectSuccessData<{ items: { id: string }[] }>(terminalsResp);

  const activeShifts = shiftsData?.shifts?.length ?? 0;
  const activeTerminals = terminalsData?.items?.length ?? 0;

  const formatRemaining = (ms: number) => {
    const totalSec = Math.floor(ms / 1000);
    const h = Math.floor(totalSec / 3600);
    const m = Math.floor((totalSec % 3600) / 60);
    return `${h}${t('dashboard.hours')} ${m}${t('dashboard.minutes')}`;
  };

  const actions = [
    { label: t('dashboard.changeFund'), path: '/cash/change-fund', accent: true },
    { label: t('dashboard.receiveCash'), path: '/cash/receive', accent: true },
    { label: t('dashboard.finalCollection'), path: '/cash/final-collection', accent: true },
    { label: t('dashboard.safeRecount'), path: '/cash/safe-recount', accent: false },
    { label: t('dashboard.bankCollection'), path: '/cash/bank-collection', accent: false },
    { label: t('dashboard.expense'), path: '/cash/expense', accent: false },
    { label: t('dashboard.credentials'), path: '/credentials', accent: false },
    { label: t('dashboard.journal'), path: '/journal', accent: false },
    { label: t('dashboard.handover'), path: '/handover', accent: false },
  ];

  return (
    <div className="sr-terminal-shell">
      <TerminalHeader title={t('dashboard.title')} onLogout={logout} />

      <main className="sr-terminal-main">
        <div className="sr-kpi-grid">
          <div className="sr-kpi-card">
            <span className="sr-kpi-label">{t('dashboard.safeBalance')}</span>
            {balancesData && <span className="sr-kpi-value">{formatMinor(safeBalance)} ₽</span>}
          </div>
          <div className="sr-kpi-card">
            <span className="sr-kpi-label">{t('dashboard.drawerTotal')}</span>
            {balancesData && <span className="sr-kpi-value">{formatMinor(drawerTotal)} ₽</span>}
          </div>
          <div className="sr-kpi-card">
            <span className="sr-kpi-label">{t('dashboard.activeShifts')}</span>
            <span className="sr-kpi-value">{activeShifts}</span>
          </div>
          <div className="sr-kpi-card">
            <span className="sr-kpi-label">{t('dashboard.activeTerminals')}</span>
            <span className="sr-kpi-value">{activeTerminals}</span>
          </div>
        </div>

        <div className="sr-action-grid">
          {actions.map((action) => (
            <Button
              key={action.path}
              variant={action.accent ? 'primary' : 'secondary'}
              className="sr-action-btn"
              onClick={() => navigate(action.path)}
            >
              {action.label}
            </Button>
          ))}
        </div>

        <div className="sr-panel">
          <div className="sr-panel-header">
            <h2 className="sr-panel-title">{t('dashboard.activeCashiers')}</h2>
            <span className="muted" style={{ fontSize: '0.85rem' }}>
              {t('dashboard.autoLockIn')}: {formatRemaining(remaining)}
            </span>
          </div>
          {!shiftsResp && <p className="muted">{t('common.loading')}</p>}
          {shiftsResp && activeShifts === 0 && <p className="muted">{t('dashboard.noShifts')}</p>}
          {shiftsData?.shifts?.map((shift) => (
            <div
              key={shift.id}
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                padding: '0.5rem 0',
                borderBottom: '1px solid var(--ui-border)',
              }}
            >
              <div>
                <div style={{ fontWeight: 500 }}>{shift.cashierId}</div>
                <div className="muted" style={{ fontSize: '0.85rem' }}>
                  {t('dashboard.drawer')}: {shift.drawerId}
                </div>
              </div>
              <div style={{ textAlign: 'right' }}>
                <div style={{ fontWeight: 600 }}>{formatMinor(shift.closingCashMinor)} ₽</div>
                <div style={{ fontSize: '0.8rem', color: 'var(--ui-text-muted)' }}>
                  {t('dashboard.revenue')}
                </div>
              </div>
            </div>
          ))}
        </div>
      </main>
    </div>
  );
}

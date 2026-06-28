import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

import { Button } from '@mercadia/ui';
import type { ListCashBalances200BalancesItem, ListOpenStoreShifts200ShiftsItem, ListStoreTerminals200ItemsItem } from '@mercadia/api-clients-store-edge';
import { formatMinorAmount } from '@/pages/reporting-utils.js';

type Props = {
  storeId: string;
  balances: ListCashBalances200BalancesItem[] | null;
  shifts: ListOpenStoreShifts200ShiftsItem[] | null;
  terminals: ListStoreTerminals200ItemsItem[] | null;
};

export function SeniorCashierDashboard({ balances, shifts, terminals }: Props) {
  const { t } = useTranslation();
  const navigate = useNavigate();

  const safeBalance = useMemo(() => {
    if (!balances) return null;
    const safe = balances.find((b) => b.containerType === 'safe');
    return safe ? safe.balanceMinor / 100 : null;
  }, [balances]);

  const totalDrawerCash = useMemo(() => {
    if (!balances) return 0;
    return balances
      .filter((b) => b.containerType === 'drawer')
      .reduce((sum, b) => sum + b.balanceMinor, 0);
  }, [balances]);

  const posTerminals = useMemo(() => {
    if (!terminals) return [];
    return terminals.filter((t) => t.kind === 'pos');
  }, [terminals]);

  return (
    <div className="senior-dashboard">
      <div className="kpi-grid">
        <div className="kpi-card">
          <p className="kpi-label">{t('seniorCashier.safeBalance')}</p>
          <p className="kpi-value">{safeBalance != null ? `${safeBalance.toFixed(2)} ₽` : '—'}</p>
        </div>
        <div className="kpi-card">
          <p className="kpi-label">{t('seniorCashier.totalDrawerCash')}</p>
          <p className="kpi-value">{(totalDrawerCash / 100).toFixed(2)} ₽</p>
        </div>
        <div className="kpi-card">
          <p className="kpi-label">{t('seniorCashier.activeCashiers')}</p>
          <p className="kpi-value">{shifts?.length ?? 0}</p>
        </div>
        <div className="kpi-card">
          <p className="kpi-label">{t('seniorCashier.terminals')}</p>
          <p className="kpi-value">{posTerminals.length}</p>
        </div>
      </div>

      <div className="action-grid">
        <Button onClick={() => navigate('/senior-cashier/change-fund')}>
          {t('seniorCashier.changeFund')}
        </Button>
        <Button onClick={() => navigate('/senior-cashier/receive-cash')}>
          {t('seniorCashier.receiveCash')}
        </Button>
        <Button onClick={() => navigate('/senior-cashier/collection')}>
          {t('seniorCashier.collection')}
        </Button>
        <Button onClick={() => navigate('/senior-cashier/safe-recount')}>
          {t('seniorCashier.safeRecount')}
        </Button>
        <Button onClick={() => navigate('/senior-cashier/bank-collection')}>
          {t('seniorCashier.bankCollection')}
        </Button>
        <Button onClick={() => navigate('/senior-cashier/expense')}>
          {t('seniorCashier.businessExpense')}
        </Button>
        <Button variant="secondary" onClick={() => navigate('/senior-cashier/journal')}>
          {t('seniorCashier.journal')}
        </Button>
        <Button variant="secondary" onClick={() => navigate('/senior-cashier/handover')}>
          {t('seniorCashier.handover')}
        </Button>
      </div>

      {shifts && shifts.length > 0 ? (
        <div className="section">
          <h2>{t('seniorCashier.activeCashiers')}</h2>
          <div className="cashier-cards">
            {shifts.map((shift) => (
              <div key={shift.id} className="cashier-card">
                <p><strong>{shift.cashierId}</strong></p>
                <p>{t('seniorCashier.drawerAmount')}: {formatMinorAmount(shift.closingCashMinor)}</p>
              </div>
            ))}
          </div>
        </div>
      ) : null}
    </div>
  );
}

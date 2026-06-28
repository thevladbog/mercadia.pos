import { useMemo, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Button, Input, Field, Label } from '@mercadia/ui';
import {
  useListCashBalances,
  useListOpenStoreShifts,
  createCashMovement,
} from '@mercadia/api-clients-store-edge';
import { useMutation, useQueryClient } from '@tanstack/react-query';

import { useAuth } from '@/auth/AuthProvider.js';
import { getStoreId } from '@/api-client-config.js';
import {
  actorsMustDiffer,
  computeDenominationTotal,
  formatMinor,
  selectSuccessData,
} from '@/lib/cash-utils.js';
import { DenominationInput } from '@/components/DenominationInput.js';
import { CashierSelectModal } from '@/components/CashierSelectModal.js';
import { TerminalHeader } from '@/components/TerminalHeader.js';

export function IssueChangeFundPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { logout } = useAuth();
  const storeId = useMemo(() => getStoreId(), []);

  const { data: balancesResp } = useListCashBalances(storeId);
  const { data: shiftsResp } = useListOpenStoreShifts(storeId);

  const balancesData = selectSuccessData<{
    balances: { containerType: string; balanceMinor: number; containerId: string }[];
  }>(balancesResp);
  const shiftsData = selectSuccessData<{
    shifts: { id: string; cashierId: string; drawerId: string; currentBalanceMinor?: number }[];
  }>(shiftsResp);

  const [selectedShift, setSelectedShift] = useState<{
    id: string;
    cashierId?: string;
    drawerId?: string;
    currentBalanceMinor?: number;
  } | null>(null);
  const [denomValues, setDenomValues] = useState<Record<number, string>>({});
  const [otherCoins, setOtherCoins] = useState(0);
  const [actorId, setActorId] = useState('');
  const [approvedById, setApprovedById] = useState('');
  const [error, setError] = useState('');

  const countedMinor = useMemo(
    () => computeDenominationTotal(denomValues, otherCoins),
    [denomValues, otherCoins],
  );

  const safeBalance = useMemo(() => {
    const safe = balancesData?.balances.find((b) => b.containerType === 'safe');
    return safe?.balanceMinor ?? 0;
  }, [balancesData]);

  const mutation = useMutation({
    mutationFn: async () => {
      const safeContainer = balancesData?.balances.find((b) => b.containerType === 'safe');
      const drawerContainer = balancesData?.balances.find((b) => b.containerType === 'drawer');

      return createCashMovement(storeId, {
        type: 'change_fund',
        fromContainerType: 'safe',
        fromContainerId: safeContainer?.containerId ?? 'safe-1',
        toContainerType: 'drawer',
        toContainerId: selectedShift?.drawerId ?? drawerContainer?.containerId ?? 'drawer-1',
        amountMinor: countedMinor || 1,
        actorId,
        approvedById,
        reason: 'change_fund',
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['/v1/stores', storeId, 'cash-balances'] });
      navigate('/dashboard', { replace: true });
    },
    onError: (err: Error) => setError(err?.message ?? t('common.unexpectedError')),
  });

  const handleSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();
      setError('');

      if (!selectedShift) {
        setError(t('cash.selectCashier'));
        return;
      }
      if (!selectedShift.drawerId) {
        setError(t('cash.selectShift'));
        return;
      }
      if (!countedMinor || countedMinor <= 0) {
        setError(t('cash.countedAmount') + ' — должно быть больше 0');
        return;
      }
      if (!actorId || !approvedById) {
        setError(t('cash.actorSelfApproval'));
        return;
      }
      if (!actorsMustDiffer(actorId, approvedById)) {
        setError(t('cash.actorSelfApproval'));
        return;
      }

      mutation.mutate();
    },
    [selectedShift, countedMinor, actorId, approvedById, mutation, t],
  );

  return (
    <div className="sr-terminal-shell">
      <TerminalHeader
        title={t('cash.changeFundTitle')}
        onLogout={() => {
          logout();
          navigate('/login', { replace: true });
        }}
      />

      <main className="sr-terminal-main">
        <form onSubmit={handleSubmit} className="sr-form">
          <p className="muted">
            {t('cash.sourceSafe')} → {t('cash.destinationDrawer')}
          </p>
          <p className="muted">
            {t('dashboard.safeBalance')}:{' '}
            {balancesData ? formatMinor(safeBalance) : t('common.loading')} ₽
          </p>

          <CashierSelectModal
            shifts={shiftsData?.shifts ?? []}
            onSelect={setSelectedShift}
            triggerLabel={selectedShift ? selectedShift.cashierId : undefined}
          />

          <DenominationInput
            values={denomValues}
            onChange={setDenomValues}
            otherAmountMinor={otherCoins}
            onOtherAmountChange={setOtherCoins}
          />

          <p className="muted">{t('cash.confirmTwoPerson')}</p>

          <Field>
            <Label>{t('cash.actorId')}</Label>
            <Input value={actorId} onChange={(e) => setActorId(e.target.value)} required />
          </Field>

          <Field>
            <Label>{t('cash.approvedById')}</Label>
            <Input
              value={approvedById}
              onChange={(e) => setApprovedById(e.target.value)}
              required
            />
          </Field>

          {error && <p className="sr-field-error">{error}</p>}

          <div style={{ display: 'flex', gap: '0.5rem' }}>
            <Button type="button" variant="ghost" onClick={() => navigate('/dashboard')}>
              {t('common.cancel')}
            </Button>
            <Button type="submit" disabled={mutation.isPending}>
              {mutation.isPending ? t('common.submitting') : t('common.confirm')}
            </Button>
          </div>
        </form>
      </main>
    </div>
  );
}

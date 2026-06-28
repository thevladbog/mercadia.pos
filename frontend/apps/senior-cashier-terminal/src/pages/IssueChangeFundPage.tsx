import { useMemo, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Button, Input, Field, Label } from '@mercadia/ui';
import { useListCashBalances, useListOpenStoreShifts, createCashMovement } from '@mercadia/api-clients-store-edge';
import { useMutation, useQueryClient } from '@tanstack/react-query';

import { getStoreId } from '@/api-client-config.js';
import { actorsMustDiffer, formatMinor } from '@/lib/cash-utils.js';
import { DenominationInput } from '@/components/DenominationInput.js';
import { CashierSelectModal } from '@/components/CashierSelectModal.js';
import { TerminalHeader } from '@/components/TerminalHeader.js';

export function IssueChangeFundPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const storeId = useMemo(() => getStoreId(), []);

  const { data: balancesResp } = useListCashBalances(storeId);
  const { data: shiftsResp } = useListOpenStoreShifts(storeId);

  const [selectedShift, setSelectedShift] = useState<any | null>(null);
  const [denomValues, setDenomValues] = useState<Record<number, string>>({});
  const [actorId, setActorId] = useState('');
  const [approvedById, setApprovedById] = useState('');
  const [error, setError] = useState('');

  const safeBalance = useMemo(() => {
    const balances = (balancesResp?.data as any)?.balances ?? [];
    const safe = balances.find((b: any) => b.containerType === 'safe');
    return safe?.balanceMinor ?? 0;
  }, [balancesResp]);

  const mutation = useMutation({
    mutationFn: async () => {
      const fromContainers = (balancesResp?.data as any)?.balances?.filter((b: any) => b.containerType === 'safe') ?? [];
      const toContainers = (balancesResp?.data as any)?.balances?.filter((b: any) => b.containerType === 'drawer') ?? [];

      return createCashMovement(storeId, {
        type: 'change_fund',
        fromContainerType: 'safe',
        fromContainerId: fromContainers[0]?.containerId ?? 'safe-1',
        toContainerType: 'drawer',
        toContainerId: selectedShift?.drawerId ?? toContainers[0]?.containerId ?? 'drawer-1',
        amountMinor: 1,
        actorId,
        approvedById,
        reason: 'change_fund',
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['/v1/stores', storeId, 'cash-balances'] });
      navigate('/dashboard', { replace: true });
    },
    onError: (err: any) => setError(err?.message ?? t('common.unexpectedError')),
  });

  const handleSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();
      setError('');

      if (!selectedShift) {
        setError(t('cash.selectCashier'));
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
    [selectedShift, actorId, approvedById, mutation, t],
  );

  return (
    <div className="sr-terminal-shell">
      <TerminalHeader title={t('cash.changeFundTitle')} onLogout={() => navigate('/login')} />

      <main className="sr-terminal-main">
        <form onSubmit={handleSubmit} className="sr-form">
          <p className="muted">{t('cash.sourceSafe')} → {t('cash.destinationDrawer')}</p>
          <p className="muted">{t('dashboard.safeBalance')}: {formatMinor(safeBalance)} ₽</p>

          <CashierSelectModal
            shifts={(shiftsResp?.data as any)?.shifts ?? []}
            onSelect={setSelectedShift}
            triggerLabel={selectedShift ? selectedShift.cashierId : undefined}
          />

          <DenominationInput values={denomValues} onChange={setDenomValues} />

          <Field>
            <Label>{t('cash.actorId')}</Label>
            <Input value={actorId} onChange={(e) => setActorId(e.target.value)} />
          </Field>

          <Field>
            <Label>{t('cash.approvedById')}</Label>
            <Input value={approvedById} onChange={(e) => setApprovedById(e.target.value)} />
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

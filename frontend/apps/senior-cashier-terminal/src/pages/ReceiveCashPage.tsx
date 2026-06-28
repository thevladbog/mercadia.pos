import { useMemo, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Button, Input, Field, Label } from '@mercadia/ui';
import { useListOpenStoreShifts, createCashMovement } from '@mercadia/api-clients-store-edge';
import { useMutation, useQueryClient } from '@tanstack/react-query';

import { getStoreId } from '@/api-client-config.js';
import { actorsMustDiffer, computeDenominationTotal } from '@/lib/cash-utils.js';
import { DenominationInput } from '@/components/DenominationInput.js';
import { CashierSelectModal } from '@/components/CashierSelectModal.js';
import { MismatchDialog } from '@/components/MismatchDialog.js';
import { TerminalHeader } from '@/components/TerminalHeader.js';

export function ReceiveCashPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const storeId = useMemo(() => getStoreId(), []);

  const { data: shiftsResp } = useListOpenStoreShifts(storeId);

  const [selectedShift, setSelectedShift] = useState<any | null>(null);
  const [expectedInput, setExpectedInput] = useState('');
  const [denomValues, setDenomValues] = useState<Record<number, string>>({});
  const [otherCoins, setOtherCoins] = useState(0);
  const [actorId, setActorId] = useState('');
  const [approvedById, setApprovedById] = useState('');
  const [error, setError] = useState('');
  const [showMismatch, setShowMismatch] = useState(false);

  const expectedMinor = useMemo(() => Math.round(parseFloat(expectedInput || '0') * 100), [expectedInput]);
  const countedMinor = useMemo(() => computeDenominationTotal(denomValues, otherCoins), [denomValues, otherCoins]);

  const mutation = useMutation({
    mutationFn: async () => {
      const fromContainers = [{ containerId: selectedShift?.drawerId ?? 'drawer-1' }];
      const toContainers = [{ containerId: 'safe-1' }];

      return createCashMovement(storeId, {
        type: 'cash_out',
        fromContainerType: 'drawer',
        fromContainerId: fromContainers[0].containerId,
        toContainerType: 'safe',
        toContainerId: toContainers[0].containerId,
        amountMinor: countedMinor || 1,
        actorId,
        approvedById,
        reason: 'revenue_collection',
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

      if (expectedMinor > 0 && expectedMinor !== countedMinor) {
        setShowMismatch(true);
        return;
      }

      mutation.mutate();
    },
    [selectedShift, actorId, approvedById, expectedMinor, countedMinor, mutation, t],
  );

  const handleResolveMismatch = useCallback(() => {
    setShowMismatch(false);
    mutation.mutate();
  }, [mutation]);

  return (
    <div className="sr-terminal-shell">
      <TerminalHeader title={t('cash.receiveCashTitle')} onLogout={() => navigate('/login')} />

      <main className="sr-terminal-main">
        <form onSubmit={handleSubmit} className="sr-form">
          <p className="muted">{t('cash.sourceDrawer')} → {t('cash.destinationSafe')}</p>

          <CashierSelectModal
            shifts={(shiftsResp?.data as any)?.shifts ?? []}
            onSelect={setSelectedShift}
            triggerLabel={selectedShift ? selectedShift.cashierId : undefined}
          />

          <Field>
            <Label>{t('cash.expectedAmount')}</Label>
            <Input
              type="number"
              min={0}
              step={0.01}
              value={expectedInput}
              onChange={(e) => setExpectedInput(e.target.value)}
              placeholder="0.00"
            />
          </Field>

          <DenominationInput
            values={denomValues}
            onChange={setDenomValues}
            otherAmountMinor={otherCoins}
            onOtherAmountChange={setOtherCoins}
          />

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

        <MismatchDialog
          expectedMinor={expectedMinor}
          countedMinor={countedMinor}
          open={showMismatch}
          onClose={() => setShowMismatch(false)}
          onResolve={handleResolveMismatch}
        />
      </main>
    </div>
  );
}

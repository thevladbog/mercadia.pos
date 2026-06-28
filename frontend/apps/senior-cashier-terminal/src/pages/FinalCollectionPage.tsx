import { useMemo, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Button, Input, Field, Label } from '@mercadia/ui';
import { useListOpenStoreShifts, closeShift } from '@mercadia/api-clients-store-edge';
import { useMutation, useQueryClient } from '@tanstack/react-query';

import { useAuth } from '@/auth/AuthProvider.js';
import { getStoreId } from '@/api-client-config.js';
import { actorsMustDiffer, computeDenominationTotal, selectSuccessData } from '@/lib/cash-utils.js';
import { DenominationInput } from '@/components/DenominationInput.js';
import { CashierSelectModal } from '@/components/CashierSelectModal.js';
import { MismatchDialog } from '@/components/MismatchDialog.js';
import { TerminalHeader } from '@/components/TerminalHeader.js';

export function FinalCollectionPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { logout } = useAuth();
  const storeId = useMemo(() => getStoreId(), []);

  const { data: shiftsResp } = useListOpenStoreShifts(storeId);
  const shiftsData = useMemo(
    () =>
      selectSuccessData<{
        shifts: { id: string; cashierId: string; drawerId: string; closingCashMinor?: number }[];
      }>(shiftsResp),
    [shiftsResp],
  );

  const [selectedShift, setSelectedShift] = useState<{
    id: string;
    cashierId?: string;
    drawerId?: string;
    closingCashMinor?: number;
  } | null>(null);
  const [denomValues, setDenomValues] = useState<Record<number, string>>({});
  const [otherCoins, setOtherCoins] = useState(0);
  const [closingCashInput, setClosingCashInput] = useState('');
  const [safeId, setSafeId] = useState('');
  const [actorId, setActorId] = useState('');
  const [approvedById, setApprovedById] = useState('');
  const [error, setError] = useState('');
  const [showMismatch, setShowMismatch] = useState(false);

  const countedMinor = useMemo(
    () => computeDenominationTotal(denomValues, otherCoins),
    [denomValues, otherCoins],
  );
  const closingCashMinor = useMemo(
    () => Math.round(parseFloat(closingCashInput || '0') * 100),
    [closingCashInput],
  );
  const expectedMinor = selectedShift?.closingCashMinor ?? 0;

  const mutation = useMutation({
    mutationFn: async () => {
      return closeShift(selectedShift!.id, {
        closingCashMinor,
        safeId: safeId || undefined,
        actorId,
        approvedById,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['/v1/stores', storeId, 'shifts'] });
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
        setError(t('cash.selectShift'));
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

      if (countedMinor !== expectedMinor) {
        setShowMismatch(true);
        return;
      }

      mutation.mutate();
    },
    [selectedShift, actorId, approvedById, countedMinor, expectedMinor, mutation, t],
  );

  const handleResolveMismatch = useCallback(() => {
    setShowMismatch(false);
    mutation.mutate();
  }, [mutation]);

  return (
    <div className="sr-terminal-shell">
      <TerminalHeader
        title={t('cash.finalCollectionTitle')}
        onLogout={() => {
          logout();
          navigate('/login', { replace: true });
        }}
      />

      <main className="sr-terminal-main">
        <form onSubmit={handleSubmit} className="sr-form">
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

          <Field>
            <Label>{t('cash.closingCash')}</Label>
            <Input
              type="number"
              min={0}
              step={0.01}
              value={closingCashInput}
              onChange={(e) => setClosingCashInput(e.target.value)}
              placeholder="0.00"
            />
          </Field>

          <Field>
            <Label>{t('cash.sourceSafe')}</Label>
            <Input
              value={safeId}
              onChange={(e) => setSafeId(e.target.value)}
              placeholder="safe-1"
            />
          </Field>

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

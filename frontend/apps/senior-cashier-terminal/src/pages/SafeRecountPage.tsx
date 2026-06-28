import { useMemo, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Button, Input, Field, Label } from '@mercadia/ui';
import {
  useListCashBalances,
  createCashRecount,
  resolveCashRecount,
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
import { MismatchDialog } from '@/components/MismatchDialog.js';
import { TerminalHeader } from '@/components/TerminalHeader.js';

export function SafeRecountPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { logout } = useAuth();
  const storeId = useMemo(() => getStoreId(), []);

  const { data: balancesResp } = useListCashBalances(storeId);
  const balancesData = useMemo(
    () =>
      selectSuccessData<{
        balances: { containerType: string; balanceMinor: number; containerId: string }[];
      }>(balancesResp),
    [balancesResp],
  );

  const safeBalance = useMemo(() => {
    const safe = balancesData?.balances.find((b) => b.containerType === 'safe');
    return safe?.balanceMinor ?? 0;
  }, [balancesData]);

  const [denomValues, setDenomValues] = useState<Record<number, string>>({});
  const [otherCoins, setOtherCoins] = useState(0);
  const [safeId, setSafeId] = useState('safe-1');
  const [actorId, setActorId] = useState('');
  const [approvedById, setApprovedById] = useState('');
  const [recountId, setRecountId] = useState<string | null>(null);
  const [error, setError] = useState('');
  const [showMismatch, setShowMismatch] = useState(false);

  const countedMinor = useMemo(
    () => computeDenominationTotal(denomValues, otherCoins),
    [denomValues, otherCoins],
  );

  const createMutation = useMutation({
    mutationFn: async () => {
      const res = await createCashRecount(storeId, {
        containerType: 'safe',
        containerId: safeId,
        countedMinor,
        actorId,
        approvedById,
      });
      return res;
    },
    onSuccess: () => {
      setRecountId('pending');

      queryClient.invalidateQueries({ queryKey: ['/v1/stores', storeId, 'cash-recounts'] });
      queryClient.invalidateQueries({ queryKey: ['/v1/stores', storeId, 'cash-balances'] });

      if (countedMinor !== safeBalance) {
        setShowMismatch(true);
      } else {
        navigate('/dashboard', { replace: true });
      }
    },
    onError: (err: Error) => setError(err?.message ?? t('common.unexpectedError')),
  });

  const resolveMutation = useMutation({
    mutationFn: async () => {
      if (!recountId) throw new Error('No recount');
      return resolveCashRecount(storeId, recountId, {
        actorId,
        approvedById,
        resolutionNote: countedMinor === safeBalance ? 'confirmed' : 'discrepancy_recorded',
      });
    },
    onSuccess: () => {
      navigate('/dashboard', { replace: true });
    },
    onError: (err: Error) => setError(err?.message ?? t('common.unexpectedError')),
  });

  const handleSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();
      setError('');

      if (!actorId || !approvedById) {
        setError(t('cash.actorSelfApproval'));
        return;
      }
      if (!actorsMustDiffer(actorId, approvedById)) {
        setError(t('cash.actorSelfApproval'));
        return;
      }

      createMutation.mutate();
    },
    [actorId, approvedById, createMutation, t],
  );

  const handleResolveMismatch = useCallback(() => {
    setShowMismatch(false);
    resolveMutation.mutate();
  }, [resolveMutation]);

  return (
    <div className="sr-terminal-shell">
      <TerminalHeader
        title={t('cash.safeRecountTitle')}
        onLogout={() => {
          logout();
          navigate('/login', { replace: true });
        }}
      />

      <main className="sr-terminal-main">
        <form onSubmit={handleSubmit} className="sr-form">
          <p className="muted">
            {t('dashboard.safeBalance')}: {formatMinor(safeBalance)} ₽
          </p>

          <Field>
            <Label>{t('cash.sourceSafe')}</Label>
            <Input value={safeId} onChange={(e) => setSafeId(e.target.value)} />
          </Field>

          <DenominationInput
            values={denomValues}
            onChange={setDenomValues}
            otherAmountMinor={otherCoins}
            onOtherAmountChange={setOtherCoins}
          />

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
            <Button type="submit" disabled={createMutation.isPending || resolveMutation.isPending}>
              {createMutation.isPending || resolveMutation.isPending
                ? t('common.submitting')
                : t('common.confirm')}
            </Button>
          </div>
        </form>

        <MismatchDialog
          expectedMinor={safeBalance}
          countedMinor={countedMinor}
          open={showMismatch}
          onClose={() => setShowMismatch(false)}
          onResolve={handleResolveMismatch}
        />
      </main>
    </div>
  );
}

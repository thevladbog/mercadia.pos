import { useMemo, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Button, Input, Field, Label } from '@mercadia/ui';
import { createBusinessExpense } from '@mercadia/api-clients-store-edge';
import { useMutation, useQueryClient } from '@tanstack/react-query';

import { useAuth } from '@/auth/AuthProvider.js';
import { getStoreId } from '@/api-client-config.js';
import { actorsMustDiffer, computeDenominationTotal } from '@/lib/cash-utils.js';
import { DenominationInput } from '@/components/DenominationInput.js';
import { TerminalHeader } from '@/components/TerminalHeader.js';

export function BusinessExpensePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { logout } = useAuth();
  const storeId = useMemo(() => getStoreId(), []);

  const [recipient, setRecipient] = useState('');
  const [reason, setReason] = useState('');
  const [denomValues, setDenomValues] = useState<Record<number, string>>({});
  const [otherCoins, setOtherCoins] = useState(0);
  const [actorId, setActorId] = useState('');
  const [approvedById, setApprovedById] = useState('');
  const [error, setError] = useState('');

  const countedMinor = useMemo(
    () => computeDenominationTotal(denomValues, otherCoins),
    [denomValues, otherCoins],
  );

  const mutation = useMutation({
    mutationFn: async () => {
      return createBusinessExpense(storeId, {
        safeId: 'safe-1',
        payeeId: recipient,
        amountMinor: countedMinor,
        reason,
        actorId,
        approvedById,
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

      if (!recipient) {
        setError(t('cash.expenseRecipient') + ' — обязательно');
        return;
      }
      if (!reason) {
        setError(t('cash.expenseReason') + ' — обязательно');
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
    [recipient, reason, countedMinor, actorId, approvedById, mutation, t],
  );

  return (
    <div className="sr-terminal-shell">
      <TerminalHeader
        title={t('cash.expenseTitle')}
        onLogout={() => {
          logout();
          navigate('/login', { replace: true });
        }}
      />

      <main className="sr-terminal-main">
        <form onSubmit={handleSubmit} className="sr-form">
          <Field>
            <Label>{t('cash.expenseRecipient')}</Label>
            <Input value={recipient} onChange={(e) => setRecipient(e.target.value)} required />
          </Field>

          <Field>
            <Label>{t('cash.expenseReason')}</Label>
            <Input
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder={t('cash.expenseReasonPlaceholder')}
              required
            />
          </Field>

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

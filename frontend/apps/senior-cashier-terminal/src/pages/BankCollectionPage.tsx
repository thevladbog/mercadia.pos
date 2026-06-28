import { useMemo, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Button, Input, Field, Label } from '@mercadia/ui';
import { createBankCollection } from '@mercadia/api-clients-store-edge';
import { useMutation, useQueryClient } from '@tanstack/react-query';

import { getStoreId } from '@/api-client-config.js';
import { actorsMustDiffer, computeDenominationTotal } from '@/lib/cash-utils.js';
import { DenominationInput } from '@/components/DenominationInput.js';
import { TerminalHeader } from '@/components/TerminalHeader.js';

export function BankCollectionPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const storeId = useMemo(() => getStoreId(), []);

  const [collectorName, setCollectorName] = useState('');
  const [contractNumber, setContractNumber] = useState('');
  const [bagSealNumber, setBagSealNumber] = useState('');
  const [bankAccount, setBankAccount] = useState('');
  const [denomValues, setDenomValues] = useState<Record<number, string>>({});
  const [otherCoins, setOtherCoins] = useState(0);
  const [actorId, setActorId] = useState('');
  const [approvedById, setApprovedById] = useState('');
  const [error, setError] = useState('');

  const countedMinor = useMemo(() => computeDenominationTotal(denomValues, otherCoins), [denomValues, otherCoins]);

  const mutation = useMutation({
    mutationFn: async () => {
      return createBankCollection(storeId, {
        safeId: 'safe-1',
        bankContainerId: bankAccount || 'bank-1',
        amountMinor: countedMinor || 1,
        actorId,
        approvedById,
        reason: collectorName ? `Collection: ${collectorName}` : undefined,
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
    [actorId, approvedById, mutation, t],
  );

  return (
    <div className="sr-terminal-shell">
      <TerminalHeader title={t('cash.bankCollectionTitle')} onLogout={() => navigate('/login')} />

      <main className="sr-terminal-main">
        <form onSubmit={handleSubmit} className="sr-form">
          <p className="muted">{t('cash.collectorInfo')}</p>

          <Field>
            <Label>{t('cash.collectorName')}</Label>
            <Input value={collectorName} onChange={(e) => setCollectorName(e.target.value)} />
          </Field>

          <Field>
            <Label>{t('cash.collectorContract')}</Label>
            <Input value={contractNumber} onChange={(e) => setContractNumber(e.target.value)} />
          </Field>

          <Field>
            <Label>{t('cash.bagSealNumber')}</Label>
            <Input value={bagSealNumber} onChange={(e) => setBagSealNumber(e.target.value)} />
          </Field>

          <Field>
            <Label>{t('cash.bankAccount')}</Label>
            <Input value={bankAccount} onChange={(e) => setBankAccount(e.target.value)} />
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
      </main>
    </div>
  );
}

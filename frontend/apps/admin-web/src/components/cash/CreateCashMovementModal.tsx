import type { ListCashBalances200BalancesItem } from '@mercadia/api-clients-store-edge';
import { createCashMovement } from '@mercadia/api-clients-store-edge';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useMemo, useState, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { ActorFields } from '@/components/cash/ActorFields.js';
import { AmountField } from '@/components/cash/AmountField.js';
import { CashModal } from '@/components/cash/CashModal.js';
import { ContainerSelect } from '@/components/cash/ContainerSelect.js';
import { containersByType, firstContainerByType } from '@/pages/cash-container-utils.js';
import {
  actorsMustDiffer,
  createIdempotencyHeaders,
  invalidateSafeQueries,
  parseRublesToMinor,
} from '@/pages/cash-mutation-utils.js';
import { storePageHref } from '@/pages/store-routes.js';

export type CashMovementVariant = 'change_fund' | 'drawer_to_safe';

type CreateCashMovementModalProps = {
  storeId: string;
  balances: ListCashBalances200BalancesItem[];
  variant: CashMovementVariant;
  onClose: () => void;
};

export function CreateCashMovementModal({
  storeId,
  balances,
  variant,
  onClose,
}: CreateCashMovementModalProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const safeContainers = useMemo(() => containersByType(balances, 'safe'), [balances]);
  const drawerContainers = useMemo(() => containersByType(balances, 'drawer'), [balances]);

  const defaultFrom = useMemo(() => {
    if (variant === 'change_fund') {
      return firstContainerByType(balances, 'safe');
    }
    return firstContainerByType(balances, 'drawer');
  }, [balances, variant]);

  const defaultTo = useMemo(() => {
    if (variant === 'change_fund') {
      return firstContainerByType(balances, 'drawer');
    }
    return firstContainerByType(balances, 'safe');
  }, [balances, variant]);

  const [fromContainerId, setFromContainerId] = useState(defaultFrom?.containerId ?? '');
  const [toContainerId, setToContainerId] = useState(defaultTo?.containerId ?? '');
  const [amountRub, setAmountRub] = useState('');
  const [reason, setReason] = useState('');
  const [actorId, setActorId] = useState('');
  const [approvedById, setApprovedById] = useState('');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const fromContainers = variant === 'change_fund' ? safeContainers : drawerContainers;
  const toContainers = variant === 'change_fund' ? drawerContainers : safeContainers;

  const fromContainer = balances.find((b) => b.containerId === fromContainerId);
  const toContainer = balances.find((b) => b.containerId === toContainerId);

  const mutation = useMutation({
    mutationFn: async () => {
      if (!fromContainer || !toContainer) {
        throw new Error('missing containers');
      }
      const amountMinor = parseRublesToMinor(amountRub);
      if (amountMinor == null) {
        throw new Error('invalid amount');
      }
      return createCashMovement(
        storeId,
        {
          type: variant,
          fromContainerId: fromContainer.containerId,
          fromContainerType: fromContainer.containerType,
          toContainerId: toContainer.containerId,
          toContainerType: toContainer.containerType,
          amountMinor,
          actorId: actorId.trim(),
          approvedById: approvedById.trim(),
          ...(reason.trim() ? { reason: reason.trim() } : {}),
        },
        { headers: createIdempotencyHeaders() },
      );
    },
    onSuccess: async (response) => {
      if (response.status === 202) {
        await invalidateSafeQueries(queryClient, storeId);
        void navigate(storePageHref('/store/safe', storeId), {
          state: {
            notice:
              variant === 'change_fund'
                ? t('safe.notices.changeFundSuccess')
                : t('safe.notices.receiveCashSuccess'),
          },
        });
        onClose();
        return;
      }
      setErrorMessage(t('common.unexpectedError'));
    },
    onError: (error) => {
      setErrorMessage(getApiErrorMessage(error));
    },
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErrorMessage(null);

    if (!actorsMustDiffer(actorId, approvedById)) {
      setErrorMessage(t('safe.forms.validation.selfApproval'));
      return;
    }
    if (parseRublesToMinor(amountRub) == null) {
      setErrorMessage(t('safe.forms.validation.amount'));
      return;
    }
    if (!fromContainer || !toContainer) {
      setErrorMessage(t('safe.forms.validation.containers'));
      return;
    }

    mutation.mutate();
  }

  const title =
    variant === 'change_fund'
      ? t('safe.actions.issueChangeFund')
      : t('safe.actions.receiveFromCashier');

  return (
    <CashModal
      errorMessage={errorMessage}
      isSubmitting={mutation.isPending}
      submitLabel={t('safe.forms.submit')}
      title={title}
      onClose={onClose}
      onSubmit={handleSubmit}
    >
      <ContainerSelect
        containers={fromContainers}
        label={t('safe.forms.fromContainer')}
        value={fromContainerId}
        onChange={setFromContainerId}
      />
      <ContainerSelect
        containers={toContainers}
        label={t('safe.forms.toContainer')}
        value={toContainerId}
        onChange={setToContainerId}
      />
      <AmountField value={amountRub} onChange={setAmountRub} />
      <label className="field">
        <span>{t('safe.forms.reason')}</span>
        <input value={reason} onChange={(event) => setReason(event.target.value)} />
      </label>
      <ActorFields
        actorId={actorId}
        approvedById={approvedById}
        onActorIdChange={setActorId}
        onApprovedByIdChange={setApprovedById}
      />
    </CashModal>
  );
}

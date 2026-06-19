import type { ListCashRecounts200ItemsItem } from '@mercadia/api-clients-store-edge';
import { resolveCashRecount } from '@mercadia/api-clients-store-edge';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useState, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

import { getApiErrorMessage } from '@/auth/api-errors.js';
import { ActorFields } from '@/components/cash/ActorFields.js';
import { CashModal } from '@/components/cash/CashModal.js';
import {
  actorsMustDiffer,
  createIdempotencyHeaders,
  invalidateSafeQueries,
} from '@/pages/cash-mutation-utils.js';
import { formatMinorAmount } from '@/pages/reporting-utils.js';
import { storePageHref } from '@/pages/store-routes.js';

type ResolveRecountModalProps = {
  storeId: string;
  recount: ListCashRecounts200ItemsItem;
  onClose: () => void;
};

export function ResolveRecountModal({ storeId, recount, onClose }: ResolveRecountModalProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [resolutionNote, setResolutionNote] = useState('');
  const [actorId, setActorId] = useState('');
  const [approvedById, setApprovedById] = useState('');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: async () =>
      resolveCashRecount(
        storeId,
        recount.id,
        {
          resolutionNote: resolutionNote.trim(),
          actorId: actorId.trim(),
          approvedById: approvedById.trim(),
        },
        { headers: createIdempotencyHeaders() },
      ),
    onSuccess: async (response) => {
      if (response.status === 202) {
        await invalidateSafeQueries(queryClient, storeId);
        void navigate(storePageHref('/store/safe', storeId), {
          state: { notice: t('safe.notices.resolveRecountSuccess') },
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

    if (resolutionNote.trim().length === 0) {
      setErrorMessage(t('safe.forms.validation.resolutionNote'));
      return;
    }
    if (!actorsMustDiffer(actorId, approvedById)) {
      setErrorMessage(t('safe.forms.validation.selfApproval'));
      return;
    }

    mutation.mutate();
  }

  return (
    <CashModal
      errorMessage={errorMessage}
      isSubmitting={mutation.isPending}
      submitLabel={t('safe.forms.resolveSubmit')}
      title={t('safe.actions.resolveRecount')}
      onClose={onClose}
      onSubmit={handleSubmit}
    >
      <p className="muted">
        {t('safe.forms.recountSummary', {
          id: recount.id,
          variance: formatMinorAmount(recount.discrepancyMinor),
        })}
      </p>
      <label className="field">
        <span>{t('safe.forms.resolutionNote')}</span>
        <textarea
          required
          rows={3}
          value={resolutionNote}
          onChange={(event) => setResolutionNote(event.target.value)}
        />
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

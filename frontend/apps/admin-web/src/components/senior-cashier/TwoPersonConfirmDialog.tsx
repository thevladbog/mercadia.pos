import { useState, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';

import { FormDialog } from '@mercadia/ui';
import { TextField } from '@/components/FormControls.js';

type TwoPersonConfirmDialogProps = {
  actorId: string;
  title: string;
  description?: string;
  isSubmitting: boolean;
  errorMessage: string | null;
  onSubmit: (actorId: string, approvedById: string) => void;
  onClose: () => void;
};

export function TwoPersonConfirmDialog({
  actorId: initialActorId,
  title,
  description,
  isSubmitting,
  errorMessage,
  onSubmit,
  onClose,
}: TwoPersonConfirmDialogProps) {
  const { t } = useTranslation();
  const [actorId, setActorId] = useState(initialActorId);
  const [approvedById, setApprovedById] = useState('');

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!actorId.trim() || !approvedById.trim()) return;
    if (actorId.trim() === approvedById.trim()) return;
    onSubmit(actorId.trim(), approvedById.trim());
  }

  const isSelfApproval = actorId.trim() && approvedById.trim() && actorId.trim() === approvedById.trim();

  return (
    <FormDialog
      cancelLabel={t('common.cancel')}
      errorMessage={errorMessage || (isSelfApproval ? t('safe.forms.validation.selfApproval') : null)}
      isSubmitting={isSubmitting}
      submitLabel={isSubmitting ? t('common.submitting') : t('seniorCashier.confirmBySecondPerson')}
      title={title}
      onClose={onClose}
      onSubmit={handleSubmit}
    >
      {description ? <p className="muted">{description}</p> : null}
      <TextField
        label={t('seniorCashier.confirmBySenior')}
        name="actorId"
        value={actorId}
        onValueChange={setActorId}
      />
      <TextField
        label={t('seniorCashier.confirmBySecondPerson')}
        name="approvedById"
        value={approvedById}
        onValueChange={setApprovedById}
      />
    </FormDialog>
  );
}

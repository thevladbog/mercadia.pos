import { useTranslation } from 'react-i18next';

type ActorFieldsProps = {
  actorId: string;
  approvedById: string;
  requireApprover?: boolean;
  onActorIdChange: (value: string) => void;
  onApprovedByIdChange: (value: string) => void;
};

export function ActorFields({
  actorId,
  approvedById,
  requireApprover = true,
  onActorIdChange,
  onApprovedByIdChange,
}: ActorFieldsProps) {
  const { t } = useTranslation();

  return (
    <div className="form-grid">
      <label className="field">
        <span>{t('safe.forms.actorId')}</span>
        <input required value={actorId} onChange={(event) => onActorIdChange(event.target.value)} />
      </label>
      <label className="field">
        <span>{t('safe.forms.approvedById')}</span>
        <input
          required={requireApprover}
          value={approvedById}
          onChange={(event) => onApprovedByIdChange(event.target.value)}
        />
      </label>
      <p className="muted form-hint">{t('safe.forms.actorHint')}</p>
    </div>
  );
}

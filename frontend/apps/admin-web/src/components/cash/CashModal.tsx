import type { FormEvent, KeyboardEvent, ReactNode } from 'react';
import { useTranslation } from 'react-i18next';

type CashModalProps = {
  title: string;
  children: ReactNode;
  errorMessage?: string | null;
  isSubmitting: boolean;
  submitLabel: string;
  onClose: () => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
};

export function CashModal({
  title,
  children,
  errorMessage,
  isSubmitting,
  submitLabel,
  onClose,
  onSubmit,
}: CashModalProps) {
  const { t } = useTranslation();

  function handleKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    if (event.key === 'Escape') {
      onClose();
    }
  }

  return (
    <div className="modal-backdrop" role="presentation" onClick={onClose}>
      <div
        autoFocus
        className="modal-panel panel"
        role="dialog"
        aria-modal="true"
        aria-labelledby="cash-modal-title"
        tabIndex={-1}
        onClick={(event) => event.stopPropagation()}
        onKeyDown={handleKeyDown}
      >
        <h3 id="cash-modal-title">{title}</h3>
        <form className="stack" onSubmit={onSubmit}>
          {children}
          {errorMessage ? <p className="error">{errorMessage}</p> : null}
          <div className="form-actions">
            <button className="secondary" disabled={isSubmitting} onClick={onClose} type="button">
              {t('common.cancel')}
            </button>
            <button disabled={isSubmitting} type="submit">
              {isSubmitting ? t('common.submitting') : submitLabel}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

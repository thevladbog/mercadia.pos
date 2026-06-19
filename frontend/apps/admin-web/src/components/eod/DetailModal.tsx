import type { KeyboardEvent, ReactNode } from 'react';
import { useTranslation } from 'react-i18next';

type DetailModalProps = {
  title: string;
  children: ReactNode;
  footer?: ReactNode;
  onClose: () => void;
};

export function DetailModal({ title, children, footer, onClose }: DetailModalProps) {
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
        aria-labelledby="detail-modal-title"
        tabIndex={-1}
        onClick={(event) => event.stopPropagation()}
        onKeyDown={handleKeyDown}
      >
        <h3 id="detail-modal-title">{title}</h3>
        <div className="stack">{children}</div>
        <div className="form-actions">
          {footer}
          <button className="secondary" onClick={onClose} type="button">
            {t('common.cancel')}
          </button>
        </div>
      </div>
    </div>
  );
}

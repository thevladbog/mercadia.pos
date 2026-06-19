import { useTranslation } from 'react-i18next';

type PaginationControlsProps = {
  canGoPrev: boolean;
  canGoNext: boolean;
  disabled?: boolean;
  onPrev: () => void;
  onNext: () => void;
};

export function PaginationControls({
  canGoPrev,
  canGoNext,
  disabled = false,
  onPrev,
  onNext,
}: PaginationControlsProps) {
  const { t } = useTranslation();

  return (
    <div className="pagination">
      <button
        className="secondary"
        disabled={!canGoPrev || disabled}
        onClick={onPrev}
        type="button"
      >
        {t('common.previous')}
      </button>
      <button
        className="secondary"
        disabled={!canGoNext || disabled}
        onClick={onNext}
        type="button"
      >
        {t('common.next')}
      </button>
    </div>
  );
}

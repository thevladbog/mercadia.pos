import { FormDialog } from '@mercadia/ui';
import type { FormEvent, ReactNode } from 'react';
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

  return (
    <FormDialog
      cancelLabel={t('common.cancel')}
      errorMessage={errorMessage}
      isSubmitting={isSubmitting}
      submitLabel={isSubmitting ? t('common.submitting') : submitLabel}
      title={title}
      onClose={onClose}
      onSubmit={onSubmit}
    >
      {children}
    </FormDialog>
  );
}

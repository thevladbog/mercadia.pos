import { DetailDialog } from '@mercadia/ui';
import type { ReactNode } from 'react';
import { useTranslation } from 'react-i18next';

type DetailModalProps = {
  title: string;
  children: ReactNode;
  footer?: ReactNode;
  onClose: () => void;
};

export function DetailModal({ title, children, footer, onClose }: DetailModalProps) {
  const { t } = useTranslation();

  return (
    <DetailDialog
      open
      title={title}
      footer={footer}
      cancelLabel={t('common.cancel')}
      onOpenChange={(open) => {
        if (!open) {
          onClose();
        }
      }}
    >
      {children}
    </DetailDialog>
  );
}

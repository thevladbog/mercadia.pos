import { useTranslation } from 'react-i18next';

import { Button, Dialog, DialogBody, DialogContent, DialogTitle } from '@mercadia/ui';

type MismatchDialogProps = {
  expectedMinor: number;
  countedMinor: number;
  open: boolean;
  onClose: () => void;
  onResolve?: () => void;
};

export function MismatchDialog({ expectedMinor, countedMinor, open, onClose, onResolve }: MismatchDialogProps) {
  const { t } = useTranslation();
  const diffMinor = expectedMinor - countedMinor;

  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) onClose(); }}>
      <DialogContent>
        <DialogTitle>{t('seniorCashier.mismatchDetected')}</DialogTitle>
        <DialogBody>
          <p>{t('seniorCashier.mismatchDetail', {
            expected: (expectedMinor / 100).toFixed(2),
            counted: (countedMinor / 100).toFixed(2),
            diff: (diffMinor / 100).toFixed(2),
          })}</p>
          <div className="dialog-actions">
            <Button variant="secondary" onClick={onClose}>{t('common.close')}</Button>
            {onResolve ? <Button onClick={onResolve}>{t('seniorCashier.mismatchResolve')}</Button> : null}
          </div>
        </DialogBody>
      </DialogContent>
    </Dialog>
  );
}

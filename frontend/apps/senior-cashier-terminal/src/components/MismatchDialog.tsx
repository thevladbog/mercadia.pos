import { useTranslation } from 'react-i18next';
import { Button, Dialog, DialogBody, DialogContent, DialogTitle } from '@mercadia/ui';

import { formatMinor } from '@/lib/cash-utils.js';

interface MismatchDialogProps {
  expectedMinor: number;
  countedMinor: number;
  open: boolean;
  onClose: () => void;
  onResolve?: () => void;
}

export function MismatchDialog({ expectedMinor, countedMinor, open, onClose, onResolve }: MismatchDialogProps) {
  const { t } = useTranslation();
  const diff = countedMinor - expectedMinor;
  const isMatch = diff === 0;

  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) onClose(); }}>
      <DialogContent>
        <DialogTitle>{isMatch ? t('cash.mismatchCorrect') : t('cash.mismatchResolve')}</DialogTitle>
        <DialogBody>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
            <div className="sr-field-row">
              <span className="sr-field-label">{t('cash.mismatchExpected')}</span>
              <span style={{ fontSize: '1.25rem', fontWeight: 700 }}>{formatMinor(expectedMinor)} ₽</span>
            </div>
            <div className="sr-field-row">
              <span className="sr-field-label">{t('cash.mismatchCounted')}</span>
              <span style={{ fontSize: '1.25rem', fontWeight: 700 }}>{formatMinor(countedMinor)} ₽</span>
            </div>
            <div className="sr-field-row">
              <span className="sr-field-label">{t('cash.mismatchDiff')}</span>
              <span
                style={{
                  fontSize: '1.25rem',
                  fontWeight: 700,
                  color: isMatch ? 'var(--ui-success)' : 'var(--ui-danger)',
                }}
              >
                {isMatch ? '0 ₽' : `${diff > 0 ? '+' : ''}${formatMinor(diff)} ₽`}
              </span>
            </div>
          </div>

          <div style={{ display: 'flex', gap: '0.5rem', marginTop: '1rem' }}>
            <Button variant="ghost" onClick={onClose}>
              {t('common.cancel')}
            </Button>
            {onResolve && !isMatch && (
              <Button variant="primary" onClick={onResolve}>
                {t('cash.mismatchResolve')}
              </Button>
            )}
          </div>
        </DialogBody>
      </DialogContent>
    </Dialog>
  );
}

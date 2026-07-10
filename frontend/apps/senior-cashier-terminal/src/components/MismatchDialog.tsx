import type { ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { Button, Dialog, DialogBody, DialogContent, DialogTitle } from '@mercadia/ui';

import { formatMinor } from '@/lib/cash-utils.js';

interface MismatchDialogProps {
  expectedMinor: number;
  countedMinor: number;
  open: boolean;
  onClose: () => void;
  onResolve?: () => void;
  /** Disables the "resolve" button without hiding it — e.g. `SafeRecountPage`
   * (plan 028) gates it on a required discrepancy comment being non-empty. */
  resolveDisabled?: boolean;
  /** Extra content rendered between the expected/counted/diff summary and the
   * button row — e.g. `SafeRecountPage`'s (plan 028) discrepancy-comment
   * textarea. `undefined` for every other caller, so their rendered output
   * is unchanged. */
  children?: ReactNode;
}

export function MismatchDialog({
  expectedMinor,
  countedMinor,
  open,
  onClose,
  onResolve,
  resolveDisabled,
  children,
}: MismatchDialogProps) {
  const { t } = useTranslation();
  const diff = countedMinor - expectedMinor;
  const isMatch = diff === 0;

  return (
    <Dialog
      open={open}
      onOpenChange={(o) => {
        if (!o) onClose();
      }}
    >
      <DialogContent>
        <DialogTitle>{isMatch ? t('cash.mismatchCorrect') : t('cash.mismatchResolve')}</DialogTitle>
        <DialogBody>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
            <div className="sr-field-row">
              <span className="sr-field-label">{t('cash.mismatchExpected')}</span>
              <span style={{ fontSize: '1.25rem', fontWeight: 700 }}>
                {formatMinor(expectedMinor)} ₽
              </span>
            </div>
            <div className="sr-field-row">
              <span className="sr-field-label">{t('cash.mismatchCounted')}</span>
              <span style={{ fontSize: '1.25rem', fontWeight: 700 }}>
                {formatMinor(countedMinor)} ₽
              </span>
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

          {children}

          <div style={{ display: 'flex', gap: '0.5rem', marginTop: '1rem' }}>
            <Button variant="ghost" onClick={onClose}>
              {t('common.cancel')}
            </Button>
            {onResolve && !isMatch && (
              <Button variant="primary" onClick={onResolve} disabled={resolveDisabled}>
                {t('cash.mismatchResolve')}
              </Button>
            )}
          </div>
        </DialogBody>
      </DialogContent>
    </Dialog>
  );
}

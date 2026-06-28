import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button, Dialog, DialogBody, DialogClose, DialogContent, DialogTitle, DialogTrigger } from '@mercadia/ui';

import { formatMinor } from '@/lib/cash-utils.js';

interface ShiftItem {
  id: string;
  cashierId?: string;
  actorId?: string;
  currentBalanceMinor?: number;
  drawerId?: string;
}

interface CashierSelectModalProps {
  shifts: ShiftItem[];
  onSelect: (shift: ShiftItem) => void;
  triggerLabel?: string;
}

export function CashierSelectModal({ shifts, onSelect, triggerLabel }: CashierSelectModalProps) {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);

  const handleSelect = (shift: ShiftItem) => {
    onSelect(shift);
    setOpen(false);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="secondary">{triggerLabel ?? t('cash.selectCashier')}</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogTitle>{t('cash.selectCashier')}</DialogTitle>
        <DialogBody>
          {shifts.length === 0 ? (
            <p className="muted">{t('dashboard.noShifts')}</p>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              {shifts.map((shift) => (
                <div
                  key={shift.id}
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    padding: '0.75rem',
                    border: '1px solid var(--ui-border)',
                    borderRadius: 'var(--ui-radius-md)',
                    cursor: 'pointer',
                  }}
                  onClick={() => handleSelect(shift)}
                  onKeyDown={(e) => { if (e.key === 'Enter') handleSelect(shift); }}
                  tabIndex={0}
                  role="button"
                >
                  <div>
                    <div style={{ fontWeight: 500 }}>{shift.cashierId ?? shift.actorId}</div>
                    <div className="muted" style={{ fontSize: '0.85rem' }}>
                      {shift.drawerId ? `${t('dashboard.drawer')}: ${shift.drawerId}` : ''}
                    </div>
                  </div>
                  <div style={{ fontWeight: 600 }}>
                    {shift.currentBalanceMinor != null ? `${formatMinor(shift.currentBalanceMinor)} ₽` : '—'}
                  </div>
                </div>
              ))}
            </div>
          )}
          <DialogClose asChild>
            <Button variant="ghost" style={{ marginTop: '0.75rem' }}>{t('common.cancel')}</Button>
          </DialogClose>
        </DialogBody>
      </DialogContent>
    </Dialog>
  );
}

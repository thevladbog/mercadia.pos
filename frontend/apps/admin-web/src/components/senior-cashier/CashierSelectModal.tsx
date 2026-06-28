import { useState } from 'react';
import { useTranslation } from 'react-i18next';

import {
  Button,
  Dialog,
  DialogBody,
  DialogClose,
  DialogContent,
  DialogTitle,
  DialogTrigger,
} from '@mercadia/ui';
import type { ListOpenStoreShifts200ShiftsItem } from '@mercadia/api-clients-store-edge';

type CashierSelectModalProps = {
  shifts: ListOpenStoreShifts200ShiftsItem[];
  onSelect: (shift: ListOpenStoreShifts200ShiftsItem) => void;
  triggerLabel?: string;
};

export function CashierSelectModal({ shifts, onSelect, triggerLabel }: CashierSelectModalProps) {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button type="button" variant="secondary">
          {triggerLabel ?? t('seniorCashier.selectCashier')}
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogTitle>{t('seniorCashier.selectCashier')}</DialogTitle>
        <DialogBody>
          {shifts.length === 0 ? (
            <p className="muted">{t('seniorCashier.noCashiers')}</p>
          ) : (
            <ul className="cashier-select-list">
              {shifts.map((shift) => (
                <li key={shift.id}>
                  <DialogClose asChild>
                    <Button
                      type="button"
                      variant="secondary"
                      onClick={() => {
                        onSelect(shift);
                        setOpen(false);
                      }}
                    >
                      {shift.cashierId} — {t('seniorCashier.drawerAmount')}:{' '}
                      {(shift.closingCashMinor / 100).toFixed(2)} ₽
                    </Button>
                  </DialogClose>
                </li>
              ))}
            </ul>
          )}
        </DialogBody>
      </DialogContent>
    </Dialog>
  );
}

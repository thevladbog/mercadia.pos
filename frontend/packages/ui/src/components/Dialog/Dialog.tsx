import * as DialogPrimitive from '@radix-ui/react-dialog';
import type { FormEvent, ReactNode } from 'react';

import { cn } from '../../lib/cn.js';
import { Button } from '../Button/Button.js';

export const Dialog = DialogPrimitive.Root;
export const DialogTrigger = DialogPrimitive.Trigger;
export const DialogClose = DialogPrimitive.Close;

export function DialogContent({
  className,
  children,
  ...props
}: DialogPrimitive.DialogContentProps) {
  return (
    <DialogPrimitive.Portal>
      <DialogPrimitive.Overlay className="mercadia-dialog-overlay" />
      <DialogPrimitive.Content className={cn('mercadia-dialog-content', className)} {...props}>
        {children}
      </DialogPrimitive.Content>
    </DialogPrimitive.Portal>
  );
}

export function DialogTitle({ className, ...props }: DialogPrimitive.DialogTitleProps) {
  return <DialogPrimitive.Title className={cn('mercadia-dialog-title', className)} {...props} />;
}

export function DialogBody({ className, children }: { className?: string; children: ReactNode }) {
  return <div className={cn('mercadia-dialog-body', className)}>{children}</div>;
}

export function DialogFooter({ className, children }: { className?: string; children: ReactNode }) {
  return <div className={cn('mercadia-dialog-footer', className)}>{children}</div>;
}

type DetailDialogProps = {
  open: boolean;
  title: string;
  children: ReactNode;
  footer?: ReactNode;
  cancelLabel: string;
  onOpenChange: (open: boolean) => void;
};

/** Compatible with admin-web DetailModal usage pattern */
export function DetailDialog({
  open,
  title,
  children,
  footer,
  cancelLabel,
  onOpenChange,
}: DetailDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent aria-describedby={undefined}>
        <DialogTitle>{title}</DialogTitle>
        <DialogBody>{children}</DialogBody>
        <DialogFooter>
          {footer}
          <DialogClose asChild>
            <Button type="button" variant="secondary">
              {cancelLabel}
            </Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

type FormDialogProps = {
  title: string;
  children: ReactNode;
  errorMessage?: string | null;
  isSubmitting: boolean;
  submitLabel: string;
  cancelLabel: string;
  onClose: () => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
};

/** Form modal with primary submit and secondary cancel — used by admin cash/EoD flows */
export function FormDialog({
  title,
  children,
  errorMessage,
  isSubmitting,
  submitLabel,
  cancelLabel,
  onClose,
  onSubmit,
}: FormDialogProps) {
  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open) {
          onClose();
        }
      }}
    >
      <DialogContent aria-describedby={undefined}>
        <DialogTitle>{title}</DialogTitle>
        <form className="mercadia-form-dialog" onSubmit={onSubmit}>
          <DialogBody>
            {children}
            {errorMessage ? <p className="mercadia-form-dialog-error">{errorMessage}</p> : null}
          </DialogBody>
          <DialogFooter>
            <DialogClose asChild>
              <Button disabled={isSubmitting} type="button" variant="secondary">
                {cancelLabel}
              </Button>
            </DialogClose>
            <Button disabled={isSubmitting} type="submit">
              {submitLabel}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

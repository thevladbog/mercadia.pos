import * as LabelPrimitive from '@radix-ui/react-label';
import type { ComponentPropsWithoutRef, ReactNode } from 'react';

import { cn } from '../../lib/cn.js';

type LabelProps = ComponentPropsWithoutRef<typeof LabelPrimitive.Root> & {
  hint?: ReactNode;
};

export function Label({ className, children, hint, ...props }: LabelProps) {
  return (
    <LabelPrimitive.Root className={cn('mercadia-label', className)} {...props}>
      <span className="mercadia-label-text">{children}</span>
      {hint ? <span className="mercadia-label-hint">{hint}</span> : null}
    </LabelPrimitive.Root>
  );
}

export function Field({ className, children }: { className?: string; children: ReactNode }) {
  return <div className={cn('mercadia-field', className)}>{children}</div>;
}

export function Input({ className, ...props }: ComponentPropsWithoutRef<'input'>) {
  return <input className={cn('mercadia-input', className)} {...props} />;
}

export function Textarea({ className, ...props }: ComponentPropsWithoutRef<'textarea'>) {
  return <textarea className={cn('mercadia-input mercadia-textarea', className)} {...props} />;
}

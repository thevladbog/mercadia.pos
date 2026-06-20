import { cva, type VariantProps } from 'class-variance-authority';
import type { HTMLAttributes } from 'react';

import { cn } from '../../lib/cn.js';

const badgeVariants = cva('mercadia-badge', {
  variants: {
    variant: {
      default: 'mercadia-badge--default',
      accent: 'mercadia-badge--accent',
      success: 'mercadia-badge--success',
      warning: 'mercadia-badge--warning',
      danger: 'mercadia-badge--danger',
      info: 'mercadia-badge--info',
      outline: 'mercadia-badge--outline',
    },
  },
  defaultVariants: {
    variant: 'default',
  },
});

type BadgeProps = HTMLAttributes<HTMLSpanElement> & VariantProps<typeof badgeVariants>;

export function Badge({ className, variant, ...props }: BadgeProps) {
  return <span className={cn(badgeVariants({ variant }), className)} {...props} />;
}

export { badgeVariants };

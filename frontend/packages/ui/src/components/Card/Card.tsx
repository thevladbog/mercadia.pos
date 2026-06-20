import { cva, type VariantProps } from 'class-variance-authority';
import type { HTMLAttributes, ReactNode } from 'react';

import { cn } from '../../lib/cn.js';

const cardVariants = cva('mercadia-card', {
  variants: {
    variant: {
      default: '',
      dark: 'mercadia-card--dark',
      success: 'mercadia-card--success',
      warning: 'mercadia-card--warning',
      danger: 'mercadia-card--danger',
      accent: 'mercadia-card--accent',
    },
  },
  defaultVariants: {
    variant: 'default',
  },
});

type CardProps = HTMLAttributes<HTMLDivElement> & VariantProps<typeof cardVariants>;

export function Card({ className, variant, ...props }: CardProps) {
  return <div className={cn(cardVariants({ variant }), className)} {...props} />;
}

export function CardHeading({ title, subtitle }: { title: ReactNode; subtitle?: ReactNode }) {
  return (
    <div className="mercadia-card-heading">
      <h3>{title}</h3>
      {subtitle ? <p className="mercadia-card-muted">{subtitle}</p> : null}
    </div>
  );
}

export { cardVariants };

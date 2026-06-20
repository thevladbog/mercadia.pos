import { Slot } from '@radix-ui/react-slot';
import { cva, type VariantProps } from 'class-variance-authority';
import type { ButtonHTMLAttributes } from 'react';

import { cn } from '../../lib/cn.js';

const buttonVariants = cva('mercadia-button', {
  variants: {
    variant: {
      primary: 'mercadia-button--primary',
      secondary: 'mercadia-button--secondary',
      ghost: 'mercadia-button--ghost',
      link: 'mercadia-button--link',
    },
    size: {
      sm: 'mercadia-button--sm',
      md: '',
      lg: 'mercadia-button--lg',
      icon: 'mercadia-button--icon',
    },
  },
  defaultVariants: {
    variant: 'primary',
    size: 'md',
  },
});

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> &
  VariantProps<typeof buttonVariants> & {
    asChild?: boolean;
  };

export function Button({
  className,
  variant,
  size,
  asChild = false,
  type = 'button',
  ...props
}: ButtonProps) {
  const Comp = asChild ? Slot : 'button';
  return (
    <Comp className={cn(buttonVariants({ variant, size }), className)} type={type} {...props} />
  );
}

export { buttonVariants };

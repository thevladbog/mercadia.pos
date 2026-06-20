import * as TabsPrimitive from '@radix-ui/react-tabs';
import type { ReactNode } from 'react';

import { cn } from '../../lib/cn.js';

export const Tabs = TabsPrimitive.Root;

export function TabsList({ className, ...props }: TabsPrimitive.TabsListProps) {
  return <TabsPrimitive.List className={cn('mercadia-tabs-list', className)} {...props} />;
}

type TabsTriggerProps = TabsPrimitive.TabsTriggerProps & {
  count?: number;
};

export function TabsTrigger({ className, children, count, ...props }: TabsTriggerProps) {
  return (
    <TabsPrimitive.Trigger className={cn('mercadia-tabs-trigger', className)} {...props}>
      {children}
      {count != null ? <span className="mercadia-tabs-count">{count}</span> : null}
    </TabsPrimitive.Trigger>
  );
}

export function TabsContent({ className, ...props }: TabsPrimitive.TabsContentProps) {
  return <TabsPrimitive.Content className={cn('mercadia-tabs-content', className)} {...props} />;
}

type PillTabsProps = {
  value: string;
  onValueChange: (value: string) => void;
  items: { value: string; label: ReactNode; count?: number; content: ReactNode }[];
  'aria-label'?: string;
};

export function PillTabs({ value, onValueChange, items, 'aria-label': ariaLabel }: PillTabsProps) {
  return (
    <Tabs value={value} onValueChange={onValueChange}>
      <TabsList aria-label={ariaLabel}>
        {items.map((item) => (
          <TabsTrigger key={item.value} value={item.value} count={item.count}>
            {item.label}
          </TabsTrigger>
        ))}
      </TabsList>
      {items.map((item) => (
        <TabsContent key={item.value} value={item.value}>
          {item.content}
        </TabsContent>
      ))}
    </Tabs>
  );
}

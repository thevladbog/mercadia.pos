import { createContext, useContext, useMemo, type ReactNode } from 'react';

import { useIdleTimer } from '@/lib/use-idle-timer.js';

interface IdleTimerContextValue {
  /** Milliseconds remaining before the idle timer's auto-lock countdown reaches zero. */
  remaining: number;
  /** True once `remaining` has reached zero. Not yet consumed anywhere (no forced logout wired up). */
  isExpired: boolean;
}

const IdleTimerContext = createContext<IdleTimerContextValue | null>(null);

/**
 * Ticks `remaining`/`isExpired` once per second via `useIdleTimer()`. Kept as
 * its own context (separate from `AuthContext`) so that the 1s tick only
 * re-renders components that actually consume the countdown, instead of every
 * `useAuth()` consumer app-wide (see plan 019 review: idle-timer tick forcing
 * app-wide re-renders).
 */
export function IdleTimerProvider({ children }: { children: ReactNode }) {
  const { remaining, isExpired } = useIdleTimer();

  const value = useMemo(() => ({ remaining, isExpired }), [remaining, isExpired]);

  return <IdleTimerContext.Provider value={value}>{children}</IdleTimerContext.Provider>;
}

export function useIdleTimerContext(): IdleTimerContextValue {
  const ctx = useContext(IdleTimerContext);
  if (!ctx) {
    throw new Error('useIdleTimerContext must be used within IdleTimerProvider');
  }
  return ctx;
}

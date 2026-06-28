import { useEffect, useRef, useCallback, useState } from 'react';

const DEFAULT_IDLE_TIMEOUT_MS = 12 * 60 * 60 * 1000;

export function useIdleTimer(timeoutMs: number = DEFAULT_IDLE_TIMEOUT_MS) {
  const [remaining, setRemaining] = useState(timeoutMs);
  const lastActivity = useRef(Date.now());
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const reset = useCallback(() => {
    lastActivity.current = Date.now();
  }, []);

  useEffect(() => {
    const events = ['mousedown', 'keydown', 'touchstart', 'mousemove', 'scroll'] as const;
    for (const event of events) {
      window.addEventListener(event, reset, { passive: true });
    }

    intervalRef.current = setInterval(() => {
      const elapsed = Date.now() - lastActivity.current;
      const rem = Math.max(0, timeoutMs - elapsed);
      setRemaining(rem);
    }, 1000);

    return () => {
      for (const event of events) {
        window.removeEventListener(event, reset);
      }
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [timeoutMs, reset]);

  return { remaining, isExpired: remaining <= 0, reset };
}

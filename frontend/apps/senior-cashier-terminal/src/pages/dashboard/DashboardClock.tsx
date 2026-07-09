import { useEffect, useState } from 'react';

/**
 * Header "14:42 · 25.05.2026" clock (design screen 02). A separate leaf
 * component with its own ticking state so only this small span re-renders
 * every interval, not the whole dashboard tree — same self-contained-ticker
 * pattern `SafeNowPanel` and `TopBar` already use. Ticks every 30s rather
 * than every 1s since only minute-granularity is ever displayed.
 */
export function DashboardClock() {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 30_000);
    return () => clearInterval(id);
  }, []);

  const d = new Date(now);
  const hh = String(d.getHours()).padStart(2, '0');
  const mm = String(d.getMinutes()).padStart(2, '0');
  const dd = String(d.getDate()).padStart(2, '0');
  const mo = String(d.getMonth() + 1).padStart(2, '0');

  return (
    <span className="sr-dashboard-datetime">
      {hh}:{mm} · {dd}.{mo}.{d.getFullYear()}
    </span>
  );
}

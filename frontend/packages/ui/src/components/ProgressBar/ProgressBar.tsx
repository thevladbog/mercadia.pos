import type { HTMLAttributes } from 'react';

import { cn } from '../../lib/cn.js';

export type ProgressBarProps = Omit<HTMLAttributes<HTMLDivElement>, 'children'> & {
  /** Progress value, 0-100. Out-of-range values are clamped into that range. */
  value: number;
};

/**
 * Clamp `value` into the 0-100 range expected by the fill's `width`
 * percentage. Judgment call (Plan 018): clamping keeps the rendered bar
 * visually valid even if a caller passes a slightly out-of-range number
 * (e.g. a rounding artifact from a percentage calculation upstream), rather
 * than rendering a fill wider than the track or a negative width.
 */
function clampValue(value: number): number {
  if (Number.isNaN(value)) {
    return 0;
  }
  return Math.min(100, Math.max(0, value));
}

/**
 * Single accent-colored fill (see Plan 018 note): the design's EoD screen
 * uses a green-to-orange gradient fill that encodes progress, but a flat
 * `--ui-accent` fill is an accepted first-pass simplification here. Revisit
 * with the real screen as reference in the EoD-specific phase.
 */
export function ProgressBar({ value, className, style, ...props }: ProgressBarProps) {
  const clamped = clampValue(value);

  return (
    <div
      className={cn('mercadia-progress-bar', className)}
      role="progressbar"
      aria-valuenow={clamped}
      aria-valuemin={0}
      aria-valuemax={100}
      style={style}
      {...props}
    >
      <div className="mercadia-progress-bar__fill" style={{ width: `${clamped}%` }} />
    </div>
  );
}

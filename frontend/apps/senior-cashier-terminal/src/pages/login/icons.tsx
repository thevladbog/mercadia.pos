/**
 * Small inline SVG icon set for the login wizard (plan 020). No icon-library
 * dependency exists in this repo for these glyphs (person/lock/wifi/check/
 * iButton) — per the plan's "Current state" note, these stay as simple
 * colocated inline SVGs rather than pulling in a new package.
 */
import type { SVGProps } from 'react';

type IconProps = SVGProps<SVGSVGElement>;

function baseProps(props: IconProps): IconProps {
  return {
    width: 20,
    height: 20,
    viewBox: '0 0 24 24',
    fill: 'none',
    stroke: 'currentColor',
    strokeWidth: 2,
    strokeLinecap: 'round',
    strokeLinejoin: 'round',
    'aria-hidden': true,
    ...props,
  };
}

export function PersonIcon(props: IconProps) {
  return (
    <svg {...baseProps(props)}>
      <circle cx="12" cy="8" r="4" />
      <path d="M4 20c0-4.4 3.6-7 8-7s8 2.6 8 7" />
    </svg>
  );
}

export function LockIcon(props: IconProps) {
  return (
    <svg {...baseProps(props)}>
      <rect x="5" y="11" width="14" height="9" rx="2" />
      <path d="M8 11V7a4 4 0 0 1 8 0v4" />
    </svg>
  );
}

export function WifiIcon(props: IconProps) {
  return (
    <svg {...baseProps(props)}>
      <path d="M2 8.5a16 16 0 0 1 20 0" />
      <path d="M5.5 12.5a11 11 0 0 1 13 0" />
      <path d="M9 16.5a6 6 0 0 1 6 0" />
      <circle cx="12" cy="20" r="1" fill="currentColor" stroke="none" />
    </svg>
  );
}

export function CheckIcon(props: IconProps) {
  return (
    <svg {...baseProps(props)}>
      <path d="M4 12.5 9 17.5 20 6.5" />
    </svg>
  );
}

/** Circular "iButton" (1-Wire touch memory key) graphic for step 3. */
export function IButtonGraphic(props: IconProps) {
  return (
    <svg {...baseProps({ width: 96, height: 96, strokeWidth: 1.5, ...props })} viewBox="0 0 96 96">
      <circle cx="48" cy="48" r="44" opacity="0.35" />
      <circle cx="48" cy="48" r="30" fill="currentColor" opacity="0.12" />
      <circle cx="48" cy="48" r="6" fill="currentColor" stroke="none" />
    </svg>
  );
}

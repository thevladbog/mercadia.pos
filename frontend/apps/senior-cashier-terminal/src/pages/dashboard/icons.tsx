/**
 * Small inline SVG icon set for the redesigned dashboard (plan 021). Same
 * "no icon-library dependency" convention as `pages/login/icons.tsx` (plan
 * 020) — these glyphs stay as colocated inline SVGs rather than pulling in
 * a new package.
 */
import type { SVGProps } from 'react';

type IconProps = SVGProps<SVGSVGElement>;

function baseProps(props: IconProps): IconProps {
  return {
    width: 22,
    height: 22,
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

/** Change-fund card: cash moving from the safe down into a drawer. */
export function DownArrowIcon(props: IconProps) {
  return (
    <svg {...baseProps(props)}>
      <path d="M12 4v13" />
      <path d="M6 12l6 6 6-6" />
    </svg>
  );
}

/** Receive-cash card: cash moving from a drawer up into the safe. */
export function UpArrowIcon(props: IconProps) {
  return (
    <svg {...baseProps(props)}>
      <path d="M12 20V7" />
      <path d="M6 12l6-6 6 6" />
    </svg>
  );
}

/** Final-collection card: the closed cash box/collection bag. */
export function BoxIcon(props: IconProps) {
  return (
    <svg {...baseProps(props)}>
      <rect x="4" y="9" width="16" height="11" rx="1.5" />
      <path d="M4 9l2.5-5h11L20 9" />
      <path d="M10 13h4" />
    </svg>
  );
}

/** Safe-recount / monitoring card: an eye (visibility/inspection). */
export function EyeIcon(props: IconProps) {
  return (
    <svg {...baseProps(props)}>
      <path d="M2 12s3.6-7 10-7 10 7 10 7-3.6 7-10 7-10-7-10-7Z" />
      <circle cx="12" cy="12" r="3" />
    </svg>
  );
}

/** Bank-collection card: a collection truck. */
export function TruckIcon(props: IconProps) {
  return (
    <svg {...baseProps(props)}>
      <rect x="2" y="7" width="12" height="9" rx="1" />
      <path d="M14 10h4l3 3v3h-7z" />
      <circle cx="6.5" cy="18" r="1.6" />
      <circle cx="17" cy="18" r="1.6" />
    </svg>
  );
}

/** Business-expense card: a price/expense tag. */
export function TagIcon(props: IconProps) {
  return (
    <svg {...baseProps(props)}>
      <path d="M12.5 3H5a2 2 0 0 0-2 2v7.5a2 2 0 0 0 .59 1.41l8.5 8.5a2 2 0 0 0 2.82 0l6.09-6.09a2 2 0 0 0 0-2.82l-8.5-8.5A2 2 0 0 0 12.5 3Z" />
      <circle cx="8" cy="8" r="1.5" fill="currentColor" stroke="none" />
    </svg>
  );
}

/** Credentials card: a staff ID card. */
export function IdCardIcon(props: IconProps) {
  return (
    <svg {...baseProps(props)}>
      <rect x="2" y="5" width="20" height="14" rx="2" />
      <circle cx="8" cy="12" r="2" />
      <path d="M5 16.5c.6-1.6 1.8-2.5 3-2.5s2.4.9 3 2.5" />
      <path d="M14 10h6" />
      <path d="M14 14h6" />
    </svg>
  );
}

/** Journal card: a list of operations. */
export function ListIcon(props: IconProps) {
  return (
    <svg {...baseProps(props)}>
      <path d="M8 6h12" />
      <path d="M8 12h12" />
      <path d="M8 18h12" />
      <circle cx="4" cy="6" r="1" fill="currentColor" stroke="none" />
      <circle cx="4" cy="12" r="1" fill="currentColor" stroke="none" />
      <circle cx="4" cy="18" r="1" fill="currentColor" stroke="none" />
    </svg>
  );
}

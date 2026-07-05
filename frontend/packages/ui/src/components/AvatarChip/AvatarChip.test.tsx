import { renderToStaticMarkup } from 'react-dom/server';
import { describe, expect, it } from 'vitest';
import type { ReactElement } from 'react';

import { AvatarChip, resolveAvatarColor } from './AvatarChip.js';

/**
 * No DOM testing library is set up in this package (see vitest.config.ts —
 * happy-dom environment only). Render via React's SSR string renderer and
 * hand the markup to happy-dom's `document` for inspection instead of
 * pulling in a new test dependency.
 */
function renderToElement(node: ReactElement): HTMLElement {
  const html = renderToStaticMarkup(node);
  const container = document.createElement('div');
  container.innerHTML = html;
  const element = container.firstElementChild;
  if (!(element instanceof HTMLElement)) {
    throw new Error('Expected AvatarChip to render a single HTMLElement');
  }
  return element;
}

describe('AvatarChip', () => {
  it('derives the same color for the same initials (determinism)', () => {
    const first = renderToElement(<AvatarChip initials="MB" />);
    const second = renderToElement(<AvatarChip initials="MB" />);
    expect(first.style.backgroundColor).toBe(second.style.backgroundColor);
  });

  it('derives colors that vary across different initials', () => {
    const colors = new Set(['AA', 'BB', 'CC', 'DD', 'EE', 'FF'].map(resolveAvatarColor));
    expect(colors.size).toBeGreaterThan(1);
  });

  it('applies a distinct size class for sm, md, and lg', () => {
    const sm = renderToElement(<AvatarChip initials="AB" size="sm" />);
    const md = renderToElement(<AvatarChip initials="AB" size="md" />);
    const lg = renderToElement(<AvatarChip initials="AB" size="lg" />);
    expect(sm.className).toContain('mercadia-avatar-chip--sm');
    expect(md.className).toContain('mercadia-avatar-chip--md');
    expect(lg.className).toContain('mercadia-avatar-chip--lg');
    expect(sm.className).not.toBe(lg.className);
  });

  it('defaults to size md when no size prop is given', () => {
    const el = renderToElement(<AvatarChip initials="AB" />);
    expect(el.className).toContain('mercadia-avatar-chip--md');
  });

  it('lets an explicit color prop override the derived color', () => {
    const derived = renderToElement(<AvatarChip initials="ZZ" />);
    const explicit = renderToElement(<AvatarChip initials="ZZ" color="#123456" />);
    expect(explicit.style.backgroundColor).not.toBe(derived.style.backgroundColor);
    expect(explicit.style.backgroundColor).toBe('#123456');
  });

  it('renders uppercase initials as its text content', () => {
    const el = renderToElement(<AvatarChip initials="ab" />);
    expect(el.textContent).toBe('AB');
  });
});

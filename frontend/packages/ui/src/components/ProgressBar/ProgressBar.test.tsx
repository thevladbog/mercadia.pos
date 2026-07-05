import { renderToStaticMarkup } from 'react-dom/server';
import { describe, expect, it } from 'vitest';
import type { ReactElement } from 'react';

import { ProgressBar } from './ProgressBar.js';

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
    throw new Error('Expected ProgressBar to render a single HTMLElement');
  }
  return element;
}

function fillWidth(el: HTMLElement): string {
  const fill = el.querySelector<HTMLElement>('.mercadia-progress-bar__fill');
  if (!fill) {
    throw new Error('Expected a .mercadia-progress-bar__fill child');
  }
  return fill.style.width;
}

describe('ProgressBar', () => {
  it('maps value to fill width percentage', () => {
    expect(fillWidth(renderToElement(<ProgressBar value={0} />))).toBe('0%');
    expect(fillWidth(renderToElement(<ProgressBar value={45} />))).toBe('45%');
    expect(fillWidth(renderToElement(<ProgressBar value={100} />))).toBe('100%');
  });

  it('clamps out-of-range values into 0-100 (judgment call: clamping enabled)', () => {
    expect(fillWidth(renderToElement(<ProgressBar value={-20} />))).toBe('0%');
    expect(fillWidth(renderToElement(<ProgressBar value={150} />))).toBe('100%');
  });

  it('exposes the clamped value via aria-valuenow/min/max for accessibility', () => {
    const el = renderToElement(<ProgressBar value={150} />);
    expect(el.getAttribute('role')).toBe('progressbar');
    expect(el.getAttribute('aria-valuenow')).toBe('100');
    expect(el.getAttribute('aria-valuemin')).toBe('0');
    expect(el.getAttribute('aria-valuemax')).toBe('100');
  });
});

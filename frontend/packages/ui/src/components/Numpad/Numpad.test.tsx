import { renderToStaticMarkup } from 'react-dom/server';
import { describe, expect, it } from 'vitest';
import type { ReactElement } from 'react';

import { Numpad } from './Numpad.js';

/**
 * No DOM testing library is set up in this package (see vitest.config.ts —
 * happy-dom environment only). Render via React's SSR string renderer and
 * hand the markup to happy-dom's `document` for inspection, matching
 * `ProgressBar.test.tsx`'s convention.
 */
function renderToElement(node: ReactElement): HTMLElement {
  const html = renderToStaticMarkup(node);
  const container = document.createElement('div');
  container.innerHTML = html;
  const element = container.firstElementChild;
  if (!(element instanceof HTMLElement)) {
    throw new Error('Expected Numpad to render a single HTMLElement');
  }
  return element;
}

function displayText(el: HTMLElement): string {
  const display = el.querySelector<HTMLElement>('.mercadia-numpad-display');
  if (!display) {
    throw new Error('Expected a .mercadia-numpad-display child');
  }
  return display.textContent ?? '';
}

describe('Numpad', () => {
  it('defaults to unmasked (plaintext) display, matching pos-terminal amount entry', () => {
    const el = renderToElement(<Numpad value="1250" onChange={() => {}} />);
    expect(displayText(el)).toBe('1250');
  });

  it('shows "0" placeholder for an empty unmasked value', () => {
    const el = renderToElement(<Numpad value="" onChange={() => {}} />);
    expect(displayText(el)).toBe('0');
  });

  it('renders one bullet per character when mask is true', () => {
    const el = renderToElement(<Numpad value="5678" onChange={() => {}} mask />);
    expect(displayText(el)).toBe('••••');
  });

  it('shows "0" placeholder for an empty masked value (no bullets to hide)', () => {
    const el = renderToElement(<Numpad value="" onChange={() => {}} mask />);
    expect(displayText(el)).toBe('0');
  });

  it('never leaks the raw digits into the DOM when masked', () => {
    const el = renderToElement(<Numpad value="1234" onChange={() => {}} mask />);
    expect(el.innerHTML).not.toContain('1234');
  });

  it('mask defaults to false when the prop is omitted, preserving byte-for-byte behavior', () => {
    const withDefault = renderToElement(<Numpad value="42" onChange={() => {}} />);
    const withExplicitFalse = renderToElement(<Numpad value="42" onChange={() => {}} mask={false} />);
    expect(displayText(withDefault)).toBe(displayText(withExplicitFalse));
    expect(displayText(withDefault)).toBe('42');
  });
});

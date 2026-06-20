import { afterEach, describe, expect, it } from 'vitest';

import { applyTheme, clearTheme } from './applyTheme.js';

describe('applyTheme', () => {
  afterEach(() => {
    clearTheme(document.documentElement);
  });

  it('sets surface and colorMode dataset attributes', () => {
    applyTheme({ surface: 'admin', colorMode: 'light', accentPreset: 'neutral' });
    expect(document.documentElement.dataset.surface).toBe('admin');
    expect(document.documentElement.dataset.colorMode).toBe('light');
  });

  it('sets runtime accent CSS variables from preset', () => {
    applyTheme({ surface: 'terminal', colorMode: 'light', accentPreset: 'return' });
    expect(document.documentElement.style.getPropertyValue('--ui-accent').trim()).toBe('#2563EB');
    expect(document.documentElement.style.getPropertyValue('--ui-accent-hover').trim()).not.toBe(
      '',
    );
    expect(document.documentElement.style.getPropertyValue('--ui-accent-muted').trim()).not.toBe(
      '',
    );
  });
});

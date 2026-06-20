import { describe, expect, it } from 'vitest';

import { deriveAccentTokens } from './deriveAccent.js';
import { ACCENT_PRESETS, resolveAccentHex } from './presets.js';

describe('deriveAccentTokens', () => {
  it('derives hover and muted from accent hex', () => {
    const tokens = deriveAccentTokens('#2563EB');
    expect(tokens.accent).toBe('#2563EB');
    expect(tokens.accentHover).not.toBe(tokens.accent);
    expect(tokens.accentMuted).not.toBe(tokens.accent);
    expect(tokens.accentForeground).toBe('#FFFFFF');
  });

  it('uses light foreground on saturated brand orange', () => {
    const tokens = deriveAccentTokens('#FF6600');
    expect(tokens.accentForeground).toBe('#FFFFFF');
  });
});

describe('resolveAccentHex', () => {
  it('prefers explicit accent over preset', () => {
    expect(resolveAccentHex({ accent: '#112233', accentPreset: 'return' })).toBe('#112233');
  });

  it('maps accent presets', () => {
    expect(resolveAccentHex({ accentPreset: 'return' })).toBe(ACCENT_PRESETS.return);
    expect(resolveAccentHex({ accentPreset: 'sale' })).toBe(ACCENT_PRESETS.sale);
  });
});

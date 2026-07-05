import { describe, expect, it } from 'vitest';

import { parseLayoutProfile } from './layout-profile.js';

describe('parseLayoutProfile', () => {
  it('accepts the horizontal profile', () => {
    expect(parseLayoutProfile('h')).toBe('h');
  });

  it('accepts the vertical profile', () => {
    expect(parseLayoutProfile('v')).toBe('v');
  });

  it('accepts the HD profile', () => {
    expect(parseLayoutProfile('hd')).toBe('hd');
  });

  it('falls back to horizontal for an unknown value', () => {
    expect(parseLayoutProfile('ultrawide')).toBe('h');
  });

  it('falls back to horizontal for undefined', () => {
    expect(parseLayoutProfile(undefined)).toBe('h');
  });

  it('falls back to horizontal for an empty string', () => {
    expect(parseLayoutProfile('')).toBe('h');
  });
});

import { describe, expect, it } from 'vitest';

import { resolveApiBaseUrl } from './api-client-config.js';

describe('resolveApiBaseUrl', () => {
  it('returns an empty string in the browser when no env value is set (Vite proxy path)', () => {
    expect(resolveApiBaseUrl('storeEdge', { isTauri: false, envValue: undefined })).toBe('');
  });

  it('returns an empty string in the browser when the env value is an empty string', () => {
    expect(resolveApiBaseUrl('central', { isTauri: false, envValue: '' })).toBe('');
  });

  it('defaults to the store-edge localhost port in Tauri when no env value is set', () => {
    expect(resolveApiBaseUrl('storeEdge', { isTauri: true, envValue: undefined })).toBe(
      'http://127.0.0.1:8081',
    );
  });

  it('defaults to the central localhost port in Tauri when no env value is set', () => {
    expect(resolveApiBaseUrl('central', { isTauri: true, envValue: undefined })).toBe(
      'http://127.0.0.1:8082',
    );
  });

  it('defaults to the hardware-agent localhost port in Tauri when no env value is set', () => {
    expect(resolveApiBaseUrl('hardwareAgent', { isTauri: true, envValue: undefined })).toBe(
      'http://127.0.0.1:8083',
    );
  });

  it('prefers an explicit env value over the Tauri localhost default', () => {
    expect(
      resolveApiBaseUrl('storeEdge', { isTauri: true, envValue: 'https://store-edge.example.com' }),
    ).toBe('https://store-edge.example.com');
  });

  it('prefers an explicit env value in the browser too', () => {
    expect(
      resolveApiBaseUrl('hardwareAgent', { isTauri: false, envValue: 'http://localhost:9999' }),
    ).toBe('http://localhost:9999');
  });

  it('treats a whitespace-only env value as empty', () => {
    expect(resolveApiBaseUrl('central', { isTauri: true, envValue: '   ' })).toBe(
      'http://127.0.0.1:8082',
    );
  });
});

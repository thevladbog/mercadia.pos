import { describe, expect, it } from 'vitest';

import { createIdempotencyHeaders, createIdempotencyKey } from './idempotency.js';

describe('createIdempotencyKey', () => {
  it('includes the scope and action in the generated key', () => {
    const key = createIdempotencyKey('pos-terminal', 'open-receipt');
    expect(key.startsWith('pos-terminal:open-receipt:')).toBe(true);
  });

  it('produces distinct keys across separate calls', () => {
    const first = createIdempotencyKey('pos-terminal', 'scan-product');
    const second = createIdempotencyKey('pos-terminal', 'scan-product');
    expect(first).not.toBe(second);
  });
});

describe('createIdempotencyHeaders', () => {
  it('wraps the idempotency key in an Idempotency-Key header', () => {
    expect(createIdempotencyHeaders('some-key')).toEqual({ 'Idempotency-Key': 'some-key' });
  });
});

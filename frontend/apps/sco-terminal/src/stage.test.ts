import { describe, expect, it } from 'vitest';

import { isStageComingSoon, reduceScoStage } from './stage.js';

describe('reduceScoStage', () => {
  it('moves from idle to scanning on start', () => {
    expect(reduceScoStage('idle', { type: 'start' })).toBe('scanning');
  });

  it('ignores start once already scanning', () => {
    expect(reduceScoStage('scanning', { type: 'start' })).toBe('scanning');
  });

  it('moves from scanning to receipt on reviewReceipt', () => {
    expect(reduceScoStage('scanning', { type: 'reviewReceipt' })).toBe('receipt');
  });

  it('ignores reviewReceipt while idle', () => {
    expect(reduceScoStage('idle', { type: 'reviewReceipt' })).toBe('idle');
  });

  it('moves from receipt back to scanning on resumeScanning', () => {
    expect(reduceScoStage('receipt', { type: 'resumeScanning' })).toBe('scanning');
  });

  it('resets to idle on cancel from any stage', () => {
    expect(reduceScoStage('scanning', { type: 'cancel' })).toBe('idle');
    expect(reduceScoStage('receipt', { type: 'cancel' })).toBe('idle');
  });
});

describe('isStageComingSoon', () => {
  it('flags payment and done as coming soon', () => {
    expect(isStageComingSoon('payment')).toBe(true);
    expect(isStageComingSoon('done')).toBe(true);
  });

  it('does not flag idle, scanning, or receipt', () => {
    expect(isStageComingSoon('idle')).toBe(false);
    expect(isStageComingSoon('scanning')).toBe(false);
    expect(isStageComingSoon('receipt')).toBe(false);
  });
});

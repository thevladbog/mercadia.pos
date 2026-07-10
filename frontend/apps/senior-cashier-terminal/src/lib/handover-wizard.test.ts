import { describe, expect, it } from 'vitest';

import { reduceHandoverWizardStep } from './handover-wizard.js';

describe('reduceHandoverWizardStep', () => {
  it('advances pickSuccessor -> pin', () => {
    expect(reduceHandoverWizardStep('pickSuccessor', { type: 'advance' })).toBe('pin');
  });

  it('advances pin -> credential', () => {
    expect(reduceHandoverWizardStep('pin', { type: 'advance' })).toBe('credential');
  });

  it('does not advance past credential (the real login() call decides success/failure, not a 4th step)', () => {
    expect(reduceHandoverWizardStep('credential', { type: 'advance' })).toBe('credential');
  });

  it('changeSuccessor returns to pickSuccessor from pin', () => {
    expect(reduceHandoverWizardStep('pin', { type: 'changeSuccessor' })).toBe('pickSuccessor');
  });

  it('changeSuccessor returns to pickSuccessor from credential', () => {
    expect(reduceHandoverWizardStep('credential', { type: 'changeSuccessor' })).toBe(
      'pickSuccessor',
    );
  });

  it('cancel returns to pickSuccessor from any step', () => {
    expect(reduceHandoverWizardStep('pin', { type: 'cancel' })).toBe('pickSuccessor');
    expect(reduceHandoverWizardStep('credential', { type: 'cancel' })).toBe('pickSuccessor');
  });
});

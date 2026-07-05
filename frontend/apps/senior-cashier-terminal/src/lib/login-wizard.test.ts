import { describe, expect, it } from 'vitest';

import {
  LOGIN_WIZARD_STEPS,
  isLoginWizardStepConfirmed,
  loginWizardStepIndex,
  reduceLoginWizardStep,
} from './login-wizard.js';

describe('reduceLoginWizardStep', () => {
  it('advances personnelId -> pin', () => {
    expect(reduceLoginWizardStep('personnelId', { type: 'advance' })).toBe('pin');
  });

  it('advances pin -> credential', () => {
    expect(reduceLoginWizardStep('pin', { type: 'advance' })).toBe('credential');
  });

  it('does not advance past credential (the real login() call decides success/failure, not a 4th step)', () => {
    expect(reduceLoginWizardStep('credential', { type: 'advance' })).toBe('credential');
  });

  it('changeIdentity returns to personnelId from pin', () => {
    expect(reduceLoginWizardStep('pin', { type: 'changeIdentity' })).toBe('personnelId');
  });

  it('changeIdentity returns to personnelId from credential', () => {
    expect(reduceLoginWizardStep('credential', { type: 'changeIdentity' })).toBe('personnelId');
  });

  it('cancel returns to personnelId from any step', () => {
    expect(reduceLoginWizardStep('pin', { type: 'cancel' })).toBe('personnelId');
    expect(reduceLoginWizardStep('credential', { type: 'cancel' })).toBe('personnelId');
  });
});

describe('loginWizardStepIndex', () => {
  it('orders steps as personnelId(0) -> pin(1) -> credential(2)', () => {
    expect(LOGIN_WIZARD_STEPS).toEqual(['personnelId', 'pin', 'credential']);
    expect(loginWizardStepIndex('personnelId')).toBe(0);
    expect(loginWizardStepIndex('pin')).toBe(1);
    expect(loginWizardStepIndex('credential')).toBe(2);
  });
});

describe('isLoginWizardStepConfirmed', () => {
  it('marks earlier steps as confirmed relative to the current step', () => {
    expect(isLoginWizardStepConfirmed('personnelId', 'pin')).toBe(true);
    expect(isLoginWizardStepConfirmed('personnelId', 'credential')).toBe(true);
    expect(isLoginWizardStepConfirmed('pin', 'credential')).toBe(true);
  });

  it('does not mark the current or later steps as confirmed', () => {
    expect(isLoginWizardStepConfirmed('pin', 'pin')).toBe(false);
    expect(isLoginWizardStepConfirmed('credential', 'pin')).toBe(false);
    expect(isLoginWizardStepConfirmed('pin', 'personnelId')).toBe(false);
  });
});

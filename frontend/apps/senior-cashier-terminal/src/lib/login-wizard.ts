/**
 * Pure state machine backing the 3-step login wizard (plan 020), mirroring
 * how `apps/sco-terminal/src/stage.ts` keeps its sale-stage machine pure and
 * dependency-free alongside the component that consumes it.
 *
 * IMPORTANT — this models UI navigation only, not authentication. The
 * backend's `CreateSession` (`backend/services/store-edge/internal/app/auth.go`)
 * validates personnel ID + PIN + credential factor together in one atomic
 * call; there is no partial/progressive validation endpoint. Reaching
 * `'credential'` here means "the operator has typed something for steps 1
 * and 2," not "the backend accepted them." See `LoginPage.tsx` for where the
 * real `login()` call happens (once, after a successful credential read).
 */
export type LoginWizardStep = 'personnelId' | 'pin' | 'credential';

/** Steps in display order, used to derive "is this step already confirmed?". */
export const LOGIN_WIZARD_STEPS: readonly LoginWizardStep[] = ['personnelId', 'pin', 'credential'];

export type LoginWizardEvent =
  { type: 'advance' } | { type: 'changeIdentity' } | { type: 'cancel' };

/**
 * Pure transition function driving the login-wizard step machine from a UI
 * event. `changeIdentity` (step 2's "Сменить" affordance) and `cancel`
 * (step 3's "Отменить вход" affordance) both land back on `'personnelId'`;
 * `LoginPage.tsx` is responsible for deciding which in-progress fields each
 * one clears (see that file's `handleChangeIdentity`/`handleCancel`), since
 * that's stateful UI concern, not step-machine concern.
 */
export function reduceLoginWizardStep(
  step: LoginWizardStep,
  event: LoginWizardEvent,
): LoginWizardStep {
  switch (event.type) {
    case 'advance':
      if (step === 'personnelId') return 'pin';
      if (step === 'pin') return 'credential';
      return step;
    case 'changeIdentity':
    case 'cancel':
      return 'personnelId';
    default:
      return step;
  }
}

/** Zero-based index of `step` within `LOGIN_WIZARD_STEPS`. */
export function loginWizardStepIndex(step: LoginWizardStep): number {
  return LOGIN_WIZARD_STEPS.indexOf(step);
}

/**
 * True when `step` is strictly earlier than `currentStep` — the "optimistic
 * checkmark" condition for the left-panel step list (design screens
 * 01b/01c). This is a UX nicety, NOT a security signal: the backend only
 * validates all three factors together at the final `login()` call, so a
 * "confirmed" earlier step just means the operator supplied a value for it,
 * not that it was verified. See plan 020's maintenance notes.
 */
export function isLoginWizardStepConfirmed(
  step: LoginWizardStep,
  currentStep: LoginWizardStep,
): boolean {
  return loginWizardStepIndex(step) < loginWizardStepIndex(currentStep);
}

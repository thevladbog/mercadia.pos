/**
 * Pure state machine backing the shift-handoff re-authentication wizard
 * (plan 029, Phase 9), mirroring `login-wizard.ts`'s exact shape (pure,
 * dependency-free, unit tested the same way) — see that file's own doc
 * comment for the shared discipline this module follows too.
 *
 * NOT a reuse of `LoginWizardStep`: that type's first step
 * (`'personnelId'`) means free-text ID entry, whereas this flow's first step
 * means picking a real successor from `joinEligibleSuccessors`'s list — a
 * different UI interaction with a different meaning, so conflating the two
 * types would blur what each step actually represents. `'pin'`/`'credential'`
 * are otherwise identical in spirit to `login-wizard.ts`'s same-named steps:
 * this still models UI navigation only, not authentication — reaching
 * `'credential'` here means "a successor has been picked and a PIN typed,"
 * not "the backend accepted them." The real re-authentication call
 * (`useAuth().login()`, called directly so the successor's session becomes
 * the new primary session — see `ShiftHandoverPage.tsx`'s doc comment for
 * why this is the opposite of `SecondSignerAuthModal.tsx`'s deliberate
 * `createAuthSession`-and-discard approach) happens once, after step 3's
 * credential read succeeds.
 */
export type HandoverWizardStep = 'pickSuccessor' | 'pin' | 'credential';

export type HandoverWizardEvent =
  { type: 'advance' } | { type: 'changeSuccessor' } | { type: 'cancel' };

/**
 * Pure transition function driving the handover-wizard step machine from a
 * UI event. Same transition shape as `reduceLoginWizardStep`: `advance`
 * moves forward one step; `changeSuccessor` (step 2's "Сменить" affordance,
 * rendered by the reused `PinStep`'s `onChangeIdentity` prop) and `cancel`
 * (step 3's "Отменить" affordance, via `CredentialStep`'s `onCancel` prop)
 * both land back on `'pickSuccessor'` — `ShiftHandoverPage.tsx` decides which
 * in-progress fields each one clears, since that's a stateful UI concern,
 * not a step-machine concern.
 */
export function reduceHandoverWizardStep(
  step: HandoverWizardStep,
  event: HandoverWizardEvent,
): HandoverWizardStep {
  switch (event.type) {
    case 'advance':
      if (step === 'pickSuccessor') return 'pin';
      if (step === 'pin') return 'credential';
      return step;
    case 'changeSuccessor':
    case 'cancel':
      return 'pickSuccessor';
    default:
      return step;
  }
}

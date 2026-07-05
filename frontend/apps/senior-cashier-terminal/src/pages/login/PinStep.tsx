import { useTranslation } from 'react-i18next';
import { Numpad } from '@mercadia/ui';

import { IdentityMiniCard } from './IdentityMiniCard.js';
import { PinDots } from './PinDots.js';

export interface PinStepProps {
  personnelId: string;
  pin: string;
  onPinChange: (value: string) => void;
  onEnter: () => void;
  onChangeIdentity: () => void;
  maxAttempts: number;
}

/**
 * Step 2 of the login wizard (design screen 01b): masked PIN entry via
 * `Numpad`'s keypad grid + the boxed-dot `PinDots` display (see
 * `LoginPage.css`'s `.sr-login-pin-numpad` comment for why `Numpad`'s own
 * display is hidden here instead of reused).
 *
 * The design's lockout note reads "after three failed attempts," but the
 * reused `MAX_ATTEMPTS` constant (`LoginPage.tsx`) is 5, not 3 — this plan
 * does not change that business rule, so the copy interpolates the real
 * `maxAttempts` value instead of hardcoding the design's literal "three."
 */
export function PinStep({
  personnelId,
  pin,
  onPinChange,
  onEnter,
  onChangeIdentity,
  maxAttempts,
}: PinStepProps) {
  const { t } = useTranslation();

  return (
    <div className="sr-login-right-content">
      <IdentityMiniCard
        variant="confirmed"
        personnelId={personnelId}
        onChangeIdentity={onChangeIdentity}
      />

      <div className="sr-login-step-heading">
        <span className="sr-login-step-heading-label">
          {t('auth.wizard.stepOfTotal', { current: 2, total: 3 })} · {t('auth.wizard.stepPinTitle')}
        </span>
        <h2 className="sr-login-step-heading-title">{t('auth.wizard.pinPrompt')}</h2>
      </div>

      <PinDots length={pin.length} />

      <Numpad
        className="sr-login-pin-numpad"
        value={pin}
        onChange={onPinChange}
        onEnter={() => {
          if (pin.length > 0) onEnter();
        }}
        mask
        enterLabel="✓"
      />

      <p className="sr-login-security-note">
        {t('auth.wizard.lockoutNote', { count: maxAttempts })}
      </p>
    </div>
  );
}

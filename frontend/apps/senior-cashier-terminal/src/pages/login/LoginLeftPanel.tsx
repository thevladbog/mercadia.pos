import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { Badge } from '@mercadia/ui';

import {
  LOGIN_WIZARD_STEPS,
  isLoginWizardStepConfirmed,
  type LoginWizardStep,
} from '@/lib/login-wizard.js';

import { CheckIcon, LockIcon, PersonIcon, WifiIcon } from './icons.js';

const STEP_ICONS: Record<LoginWizardStep, (props: { className?: string }) => ReactElement> = {
  personnelId: PersonIcon,
  pin: LockIcon,
  credential: WifiIcon,
};

const STEP_TITLE_KEYS: Record<LoginWizardStep, string> = {
  personnelId: 'auth.wizard.stepPersonnelIdTitle',
  pin: 'auth.wizard.stepPinTitle',
  credential: 'auth.wizard.stepCredentialTitle',
};

export interface LoginLeftPanelProps {
  currentStep: LoginWizardStep;
  storeId: string;
}

/**
 * Left brand panel shared across all 3 wizard steps (design screens
 * 01a/01b/01c: identical left column, only the step list's active/done
 * markers change). Step "done" checkmarks are optimistic-UI only — see
 * `login-wizard.ts`'s `isLoginWizardStepConfirmed` doc comment.
 */
export function LoginLeftPanel({ currentStep, storeId }: LoginLeftPanelProps) {
  const { t } = useTranslation();

  return (
    <div className="sr-login-left">
      <div className="sr-login-intro">
        <div className="sr-login-brand">
          <span className="sr-login-logo" aria-hidden="true">
            M
          </span>
          <div>
            <div className="sr-login-wordmark-title">Mercadia</div>
            <div className="sr-login-wordmark-sub">{t('auth.wizard.brandSubtitle')}</div>
          </div>
        </div>

        <Badge variant="warning">{t('auth.wizard.elevatedAccess')}</Badge>

        <h1 className="sr-login-heading">{t('auth.wizard.heading')}</h1>
        <p className="sr-login-description">{t('auth.wizard.description')}</p>

        <div className="sr-login-steps">
          {LOGIN_WIZARD_STEPS.map((step, index) => {
            const Icon = STEP_ICONS[step];
            const done = isLoginWizardStepConfirmed(step, currentStep);
            const active = step === currentStep;
            return (
              <div
                key={step}
                className={`sr-login-step${active ? ' sr-login-step--active' : ''}${
                  done ? ' sr-login-step--done' : ''
                }`}
              >
                <span className="sr-login-step-marker">
                  {done ? <CheckIcon width={14} height={14} /> : index + 1}
                </span>
                <div className="sr-login-step-body">
                  <span className="sr-login-step-label">
                    {t('auth.wizard.stepLabel', { number: index + 1 })}
                  </span>
                  <span className="sr-login-step-title">{t(STEP_TITLE_KEYS[step])}</span>
                </div>
                <Icon className="sr-login-step-icon" />
              </div>
            );
          })}
        </div>
      </div>

      <div className="sr-login-footer">
        <span>{storeId}</span>
      </div>
    </div>
  );
}

import { useTranslation } from 'react-i18next';
import { AvatarChip, Badge, Button } from '@mercadia/ui';

import { deriveInitials } from '@/lib/topbar-utils.js';
import type { StaffCredentialKind } from '@/auth/ibutton.js';

/**
 * The small top-of-right-panel identity card shown on steps 2 and 3
 * (design screens 01b/01c). Both variants show only what the OPERATOR HAS
 * TYPED SO FAR, not what the backend has verified — the atomic
 * `createAuthSession` call only runs once, at the end of step 3 (see
 * `login-wizard.ts`'s doc comment). There is no per-actor role/display-name
 * to show here yet either (no name field on `domain.Actor`, same constraint
 * `topbar-utils.ts` documents), so this only ever shows the personnel ID
 * the operator entered.
 */
export type IdentityMiniCardProps =
  | {
      variant: 'confirmed';
      personnelId: string;
      onChangeIdentity: () => void;
    }
  | {
      variant: 'badges';
      personnelId: string;
      credentialKind: StaffCredentialKind;
      credentialConfirmed: boolean;
    };

export function IdentityMiniCard(props: IdentityMiniCardProps) {
  const { t } = useTranslation();
  const initials = deriveInitials(props.personnelId);

  return (
    <div className="sr-login-mini-card">
      <div className="sr-login-mini-card-identity">
        <AvatarChip initials={initials} size="md" />
        <div className="sr-login-mini-card-text">
          <div className="sr-login-mini-card-badges">
            {props.variant === 'confirmed' ? (
              <Badge variant="success">{t('auth.wizard.idConfirmed')}</Badge>
            ) : (
              <>
                <Badge variant="success">{t('auth.wizard.badgeId')}</Badge>
                <Badge variant="success">{t('auth.wizard.badgePin')}</Badge>
                <Badge variant={props.credentialConfirmed ? 'success' : 'outline'}>
                  {t(`auth.credentialKinds.${props.credentialKind}`)}
                </Badge>
              </>
            )}
          </div>
          <span className="sr-login-mini-card-name">{props.personnelId}</span>
        </div>
      </div>

      {props.variant === 'confirmed' && (
        <Button type="button" variant="secondary" size="sm" onClick={props.onChangeIdentity}>
          {t('auth.wizard.changeIdentity')}
        </Button>
      )}
    </div>
  );
}

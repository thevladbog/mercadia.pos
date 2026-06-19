import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';

import {
  CENTRAL_ROLE_ADMIN,
  CENTRAL_ROLE_OPTIONS,
  CENTRAL_ROLE_VIEWER,
} from '@/auth/permissions.js';

type CentralRoleFieldsProps = {
  roles: string[];
  onChange: (roles: string[]) => void;
};

export function CentralRoleFields({ roles, onChange }: CentralRoleFieldsProps) {
  const { t } = useTranslation();

  function roleLabel(role: string): string {
    if (role === CENTRAL_ROLE_ADMIN) {
      return t('users.roleAdmin');
    }
    if (role === CENTRAL_ROLE_VIEWER) {
      return t('users.roleViewer');
    }
    return role;
  }

  function toggleRole(role: string, checked: boolean) {
    if (checked) {
      onChange([...new Set([...roles, role])]);
      return;
    }
    onChange(roles.filter((value) => value !== role));
  }

  return (
    <fieldset className="role-fieldset">
      <legend>{t('users.roles')}</legend>
      <div className="role-options">
        {CENTRAL_ROLE_OPTIONS.map((role) => (
          <label className="checkbox-field" key={role}>
            <input
              checked={roles.includes(role)}
              type="checkbox"
              onChange={(event) => toggleRole(role, event.target.checked)}
            />
            <span>{roleLabel(role)}</span>
          </label>
        ))}
      </div>
    </fieldset>
  );
}

type PageBackLinkProps = {
  label?: string;
  to?: string;
};

export function PageBackLink({ label, to = '/central/users' }: PageBackLinkProps) {
  const { t } = useTranslation();

  return (
    <p className="page-back">
      <Link to={to}>{label ?? t('users.backToUsers')}</Link>
    </p>
  );
}

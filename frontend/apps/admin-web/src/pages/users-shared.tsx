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

function roleLabel(role: string): string {
  if (role === CENTRAL_ROLE_ADMIN) {
    return 'Central admin';
  }
  if (role === CENTRAL_ROLE_VIEWER) {
    return 'Central viewer';
  }
  return role;
}

export function CentralRoleFields({ roles, onChange }: CentralRoleFieldsProps) {
  function toggleRole(role: string, checked: boolean) {
    if (checked) {
      onChange([...new Set([...roles, role])]);
      return;
    }
    onChange(roles.filter((value) => value !== role));
  }

  return (
    <fieldset className="role-fieldset">
      <legend>Roles</legend>
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

export function PageBackLink({
  label = 'Back to users',
  to = '/central/users',
}: PageBackLinkProps) {
  return (
    <p className="page-back">
      <Link to={to}>{label}</Link>
    </p>
  );
}

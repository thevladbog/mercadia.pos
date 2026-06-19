import { Navigate, Outlet } from 'react-router-dom';

import { useAuth } from './AuthProvider.js';
import { canManageCentralUsers } from './permissions.js';

export function RequireCentralAdmin() {
  const { roles } = useAuth();

  if (!canManageCentralUsers(roles)) {
    return (
      <Navigate
        replace
        state={{ notice: 'Central admin role is required to manage users.' }}
        to="/central/reporting"
      />
    );
  }

  return <Outlet />;
}

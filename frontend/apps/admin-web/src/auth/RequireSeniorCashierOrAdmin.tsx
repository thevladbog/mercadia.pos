import { Navigate, Outlet } from 'react-router-dom';

import { useAuth } from './useAuth.js';
import { canWriteStoreOperations } from './permissions.js';

export function RequireSeniorCashierOrAdmin() {
  const { roles } = useAuth();

  if (!canWriteStoreOperations(roles)) {
    return (
      <Navigate
        replace
        state={{ notice: 'Недостаточно прав для доступа к этой странице.' }}
        to="/central/dashboard"
      />
    );
  }

  return <Outlet />;
}

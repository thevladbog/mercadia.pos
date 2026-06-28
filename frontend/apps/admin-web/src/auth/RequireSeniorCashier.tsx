import { Navigate, Outlet } from 'react-router-dom';

import { useAuth } from './useAuth.js';
import { isSeniorCashier } from './permissions.js';

export function RequireSeniorCashier() {
  const { roles } = useAuth();

  if (!isSeniorCashier(roles)) {
    return (
      <Navigate
        replace
        state={{ notice: 'Доступно только старшему кассиру.' }}
        to="/central/dashboard"
      />
    );
  }

  return <Outlet />;
}

import { Navigate, Outlet } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { useAuth } from './useAuth.js';
import { canWriteStoreOperations } from './permissions.js';

export function RequireSeniorCashierOrAdmin() {
  const { t } = useTranslation();
  const { roles } = useAuth();

  if (!canWriteStoreOperations(roles)) {
    return (
      <Navigate
        replace
        state={{ notice: t('seniorCashier.insufficientPermissions') }}
        to="/central/dashboard"
      />
    );
  }

  return <Outlet />;
}

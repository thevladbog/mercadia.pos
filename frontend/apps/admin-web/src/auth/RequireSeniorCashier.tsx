import { Navigate, Outlet } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { useAuth } from './useAuth.js';
import { isSeniorCashier } from './permissions.js';

export function RequireSeniorCashier() {
  const { t } = useTranslation();
  const { roles } = useAuth();

  if (!isSeniorCashier(roles)) {
    return (
      <Navigate
        replace
        state={{ notice: t('seniorCashier.seniorCashierOnly') }}
        to="/central/dashboard"
      />
    );
  }

  return <Outlet />;
}

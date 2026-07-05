import { useNavigate } from 'react-router-dom';

import { useAuth } from '@/auth/AuthProvider.js';

export interface TopBarActions {
  onHandover: () => void;
  onLock: () => void;
}

/**
 * Shared `<TopBar onHandover onLock>` wiring used by every page except
 * `ShiftHandoverPage` (which is the handover flow itself and intentionally
 * locks without calling `logout()`).
 */
export function useTopBarActions(): TopBarActions {
  const navigate = useNavigate();
  const { logout } = useAuth();

  return {
    onHandover: () => navigate('/handover'),
    onLock: () => {
      logout();
      navigate('/login', { replace: true });
    },
  };
}

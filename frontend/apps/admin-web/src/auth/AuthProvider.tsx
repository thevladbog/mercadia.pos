import { setSessionToken } from '@mercadia/api-clients-central';
import { useCallback, useMemo, useState, type ReactNode } from 'react';

import { AuthContext } from './auth-context.js';
import { clearPersistedAuth, persistAuth, readStoredAuth } from './auth-storage.js';
import type { AuthContextValue, AuthState } from './auth-types.js';

type AuthProviderProps = {
  children: ReactNode;
};

export function AuthProvider({ children }: AuthProviderProps) {
  const [auth, setAuth] = useState<AuthState>(() => {
    const stored = readStoredAuth();
    return {
      userId: stored.userId,
      roles: stored.roles,
      isAuthenticated: stored.userId !== null,
    };
  });

  const login = useCallback((userId: string, roles: string[], token: string) => {
    setSessionToken(token);
    persistAuth(userId, roles);
    setAuth({ userId, roles, isAuthenticated: true });
  }, []);

  const logout = useCallback(() => {
    clearPersistedAuth();
    setAuth({ userId: null, roles: [], isAuthenticated: false });
  }, []);

  const handleUnauthorized = useCallback(() => {
    logout();
  }, [logout]);

  const value = useMemo<AuthContextValue>(
    () => ({
      ...auth,
      login,
      logout,
      handleUnauthorized,
    }),
    [auth, login, logout, handleUnauthorized],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

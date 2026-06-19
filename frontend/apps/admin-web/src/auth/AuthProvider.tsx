import {
  ApiError,
  clearSessionToken,
  getSessionToken,
  setApiBaseUrl,
  setSessionToken,
} from '@mercadia/api-clients-central';
import {
  ApiError as StoreEdgeApiError,
  setApiBaseUrl as setStoreEdgeApiBaseUrl,
} from '@mercadia/api-clients-store-edge';
import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from 'react';

export type AuthState = {
  userId: string | null;
  roles: string[];
  isAuthenticated: boolean;
};

export type AuthContextValue = AuthState & {
  login: (userId: string, roles: string[], token: string) => void;
  logout: () => void;
  handleUnauthorized: () => void;
};

const AuthContext = createContext<AuthContextValue | null>(null);

const USER_ID_KEY = 'mercadia.central.userId';
const ROLES_KEY = 'mercadia.central.roles';

function readStoredAuth(): Pick<AuthState, 'userId' | 'roles'> {
  const userId = sessionStorage.getItem(USER_ID_KEY);
  const rolesRaw = sessionStorage.getItem(ROLES_KEY);

  if (!userId || !getSessionToken()) {
    return { userId: null, roles: [] };
  }

  try {
    const roles = rolesRaw ? (JSON.parse(rolesRaw) as string[]) : [];
    return { userId, roles };
  } catch {
    return { userId: null, roles: [] };
  }
}

function persistAuth(userId: string, roles: string[]): void {
  sessionStorage.setItem(USER_ID_KEY, userId);
  sessionStorage.setItem(ROLES_KEY, JSON.stringify(roles));
}

function clearPersistedAuth(): void {
  sessionStorage.removeItem(USER_ID_KEY);
  sessionStorage.removeItem(ROLES_KEY);
  clearSessionToken();
}

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

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}

export function isUnauthorizedError(error: unknown): boolean {
  return error instanceof ApiError && error.status === 401;
}

export function getApiErrorMessage(error: unknown): string {
  if (error instanceof ApiError || error instanceof StoreEdgeApiError) {
    return error.problem.detail ?? error.problem.title;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return 'Unexpected error';
}

export function configureCentralApiClient(): void {
  setApiBaseUrl(import.meta.env.VITE_CENTRAL_BACKEND_URL ?? '');
}

export function configureStoreEdgeApiClient(): void {
  setStoreEdgeApiBaseUrl(import.meta.env.VITE_STORE_EDGE_URL ?? '');
}

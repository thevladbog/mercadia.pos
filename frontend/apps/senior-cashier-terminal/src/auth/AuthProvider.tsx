import { createContext, useContext, useState, useCallback, useMemo, type ReactNode } from 'react';
import {
  ApiError,
  type CreateAuthSessionBodyCredentialFactor,
  clearSessionToken,
  createAuthSession,
  setSessionToken,
} from '@mercadia/api-clients-store-edge';

import { getStoreId } from '@/api-client-config.js';

import type { SessionResult } from './types.js';

interface AuthContextValue {
  session: SessionResult | null;
  login: (
    actorId: string,
    pin: string,
    credentialFactor: CreateAuthSessionBodyCredentialFactor,
  ) => Promise<SessionResult>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

const SESSION_KEY = 'mercadia.sr-terminal.session';

function isSessionResult(value: unknown): value is SessionResult {
  if (typeof value !== 'object' || value === null) return false;
  const candidate = value as Partial<SessionResult>;
  return (
    typeof candidate.token === 'string' &&
    typeof candidate.actorId === 'string' &&
    typeof candidate.expiresAt === 'string' &&
    Array.isArray(candidate.roles) &&
    candidate.roles.every((role) => typeof role === 'string')
  );
}

function loadSession(): SessionResult | null {
  try {
    const raw = sessionStorage.getItem(SESSION_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as unknown;
    if (!isSessionResult(parsed)) {
      clearSession();
      return null;
    }
    const expiresAt = new Date(parsed.expiresAt).getTime();
    if (!Number.isFinite(expiresAt) || expiresAt <= Date.now()) {
      clearSession();
      return null;
    }
    setSessionToken(parsed.token);
    return parsed;
  } catch {
    clearSession();
    return null;
  }
}

function saveSession(session: SessionResult): void {
  sessionStorage.setItem(SESSION_KEY, JSON.stringify(session));
  setSessionToken(session.token);
}

function clearSession(): void {
  sessionStorage.removeItem(SESSION_KEY);
  clearSessionToken();
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [session, setSession] = useState<SessionResult | null>(loadSession);

  const login = useCallback(
    async (
      actorId: string,
      pin: string,
      credentialFactor: CreateAuthSessionBodyCredentialFactor,
    ): Promise<SessionResult> => {
      try {
        const response = await createAuthSession({
          actorId,
          pin,
          storeId: getStoreId(),
          credentialFactor,
        });
        if (response.status !== 201) {
          throw new Error('Authentication failed');
        }

        const nextSession = response.data.session;
        saveSession(nextSession);
        setSession(nextSession);
        return nextSession;
      } catch (err) {
        if (err instanceof ApiError && err.status === 401) {
          throw new Error('Invalid credentials', { cause: err });
        }
        throw err;
      }
    },
    [],
  );

  const logout = useCallback(() => {
    clearSession();
    setSession(null);
  }, []);

  const value = useMemo(() => ({ session, login, logout }), [session, login, logout]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return ctx;
}

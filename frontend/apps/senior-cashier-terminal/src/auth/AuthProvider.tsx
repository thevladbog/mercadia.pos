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
  /**
   * ISO timestamp recorded locally the moment `login()` last succeeded for
   * the current session — NOT derived from `session.expiresAt` minus an
   * assumed TTL constant (plan 029's ground truth: `SessionResult`, i.e.
   * `CreateAuthSession201Session`, carries no login/issued-at field at all;
   * reverse-deriving one from a hardcoded TTL would silently break if the
   * backend's session TTL ever changes, with no compile-time signal). Mirrors
   * `login-history.ts`'s existing precedent of tracking purely-local,
   * first-party metadata alongside the real session, not sourced from the
   * backend. `null` before any session has been established on this
   * terminal/tab.
   */
  loggedInAt: string | null;
  login: (
    actorId: string,
    pin: string,
    credentialFactor: CreateAuthSessionBodyCredentialFactor,
  ) => Promise<SessionResult>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

const SESSION_KEY = 'mercadia.sr-terminal.session';
/** Companion key for `loggedInAt`, always written/cleared alongside
 * `SESSION_KEY` so the two never drift apart — see `AuthContextValue
 * .loggedInAt`'s doc comment above. */
const SESSION_META_KEY = 'mercadia.sr-terminal.session-meta';

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

/** Reads the companion `loggedInAt` timestamp. Must only be trusted when
 * `loadSession()` also returned a non-null session (see this file's initial
 * `useState` calls, which run in this exact order on first render) — a
 * missing/corrupted value here just falls back to `null`, same defensive
 * try/catch-to-null style as `loadSession` itself. */
function loadLoggedInAt(): string | null {
  try {
    return sessionStorage.getItem(SESSION_META_KEY);
  } catch {
    return null;
  }
}

function saveSession(session: SessionResult, loggedInAt: string): void {
  sessionStorage.setItem(SESSION_KEY, JSON.stringify(session));
  sessionStorage.setItem(SESSION_META_KEY, loggedInAt);
  setSessionToken(session.token);
}

function clearSession(): void {
  sessionStorage.removeItem(SESSION_KEY);
  sessionStorage.removeItem(SESSION_META_KEY);
  clearSessionToken();
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [session, setSession] = useState<SessionResult | null>(loadSession);
  // Reads AFTER `loadSession()` above in the same render pass — if that call
  // found an invalid/expired session it already cleared both storage keys,
  // so this correctly resolves to `null` in that case too.
  const [loggedInAt, setLoggedInAt] = useState<string | null>(loadLoggedInAt);

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
        const nextLoggedInAt = new Date().toISOString();
        saveSession(nextSession, nextLoggedInAt);
        setSession(nextSession);
        setLoggedInAt(nextLoggedInAt);
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
    setLoggedInAt(null);
  }, []);

  const value = useMemo(
    () => ({ session, loggedInAt, login, logout }),
    [session, loggedInAt, login, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return ctx;
}
